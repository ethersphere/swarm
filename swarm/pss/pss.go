package pss

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/pot"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

// TODO: proper padding generation for messages
const (
	digestLength               = 32 // byte length of digest used for pss cache (currently same as swarm chunk hash)
	DefaultTTL                 = 6000
	defaultSymKeyExpiry        = 1000 * 60 * 60 * 6
	defaultSymKeyRequestExpiry = 5000
	defaultWhisperWorkTime     = 3
	defaultWhisperPoW          = 0.0000000001
	defaultMaxMsgSize          = 1024 * 1024
	defaultSymKeyLength        = 32
)

const (
	SYMSTATUS_OK = iota
	SYMSTATUS_NONE
	SYMSTATUS_PENDING
	SYMSTATUS_EXPIRED
)

var (
	symKeyRequestExpiry = time.Millisecond * defaultSymKeyRequestExpiry
	symKeyExpiry        = time.Millisecond * defaultSymKeyExpiry
	addressLength       = len(pot.Address{})
)

// Toplevel pss object, takes care of message sending, receiving, decryption and encryption, message handler dispatchers and message forwarding.
//
// Implements node.Service
type Pss struct {
	network.Overlay                                             // we can get the overlayaddress from this
	fwdPool           map[discover.NodeID]*protocols.Peer       // keep track of all peers sitting on the pssmsg routing layer
	pubKeyPool        map[string]map[whisper.TopicType]*pssPeer // mapping of hex public keys to peer address and rw
	symKeyPool        map[string]map[whisper.TopicType]*pssPeer // mapping of symkeyids to peer address and rw
	handlers          map[whisper.TopicType]map[*Handler]bool   // topic and version based pss payload handlers
	fwdcache          map[pssDigest]pssCacheEntry               // checksum of unique fields from pssmsg mapped to expiry, cache to determine whether to drop msg
	cachettl          time.Duration                             // how long to keep messages in fwdcache
	lock              sync.Mutex
	dpa               *storage.DPA
	privateKey        *ecdsa.PrivateKey
	w                 *whisper.Whisper
	symKeyPairIndex   map[string]*string     // look up matching symkeys for handshake
	symKeyPairPubKey  map[string]string      // look up matching pubkey for handshake send symkey
	symKeyCache       []*string              // fast lookup of recently used symkeys; last used is on top of stack
	symKeyCacheCursor int                    // modular cursor pointing to last used, wraps on symKeyCache array
	handshakeC        map[string]chan string // adds a channel to report when a handshake succeeds
}

func (self *Pss) String() string {
	return fmt.Sprintf("pss: addr %x, pubkey %v", self.BaseAddr(), common.ToHex(crypto.FromECDSAPub(&self.privateKey.PublicKey)))
}

// Creates a new Pss instance.
//
// Needs a swarm network overlay, a DPA storage for message cache storage.
func NewPss(k network.Overlay, dpa *storage.DPA, params *PssParams) *Pss {
	return &Pss{
		Overlay:          k,
		fwdPool:          make(map[discover.NodeID]*protocols.Peer),
		symKeyPool:       make(map[string]map[whisper.TopicType]*pssPeer),
		pubKeyPool:       make(map[string]map[whisper.TopicType]*pssPeer),
		handlers:         make(map[whisper.TopicType]map[*Handler]bool),
		fwdcache:         make(map[pssDigest]pssCacheEntry),
		cachettl:         params.CacheTTL,
		dpa:              dpa,
		privateKey:       params.privateKey,
		w:                whisper.New(),
		symKeyPairIndex:  make(map[string]*string),
		symKeyPairPubKey: make(map[string]string),
		symKeyCache:      make([]*string, params.SymKeyCacheCapacity),
		handshakeC:       make(map[string]chan string),
	}
}

// Convenience accessor to the swarm overlay address of the pss node
func (self *Pss) BaseAddr() []byte {
	return self.Overlay.BaseAddr()
}

// Accessor for own public key
func (self *Pss) PublicKey() ecdsa.PublicKey {
	return self.privateKey.PublicKey
}

// For node.Service implementation. Does nothing for now, but should be included in the code for backwards compatibility.
func (self *Pss) Start(srv *p2p.Server) error {
	return nil
}

// For node.Service implementation. Does nothing for now, but should be included in the code for backwards compatibility.
func (self *Pss) Stop() error {
	return nil
}

// devp2p protocol object for the PssMsg struct.
//
// This represents the PssMsg capsule, and is the entry point for processing, receiving and sending pss messages between directly connected peers.
func (self *Pss) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		p2p.Protocol{
			Name:    pssSpec.Name,
			Version: pssSpec.Version,
			Length:  pssSpec.Length(),
			Run:     self.Run,
		},
	}
}

// Starts the PssMsg protocol
func (self *Pss) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	pp := protocols.NewPeer(p, rw, pssSpec)
	self.fwdPool[p.ID()] = pp
	return pp.Run(self.handlePssMsg)
}

// Exposes the API methods
//
// If the debug-parameter was given to the top Pss object, the TestAPI methods will also be included
func (self *Pss) APIs() []rpc.API {
	apis := []rpc.API{
		rpc.API{
			Namespace: "pss",
			Version:   "0.1",
			Service:   NewAPI(self),
			Public:    true,
		},
		rpc.API{
			Namespace: "psstest",
			Version:   "0.1",
			Service:   NewAPITest(self),
			Public:    false,
		},
	}
	return apis
}

// Links a handler function to a Topic
//
// After calling this, all incoming messages with an envelope Topic matching the Topic specified will be passed to the given Handler function.
//
// Returns a deregister function which needs to be called to deregister the handler,
func (self *Pss) Register(topic *whisper.TopicType, handler Handler) func() {
	self.lock.Lock()
	defer self.lock.Unlock()
	handlers := self.handlers[*topic]
	if handlers == nil {
		handlers = make(map[*Handler]bool)
		self.handlers[*topic] = handlers
	}
	handlers[&handler] = true
	return func() { self.deregister(topic, &handler) }
}
func (self *Pss) deregister(topic *whisper.TopicType, h *Handler) {
	self.lock.Lock()
	defer self.lock.Unlock()
	handlers := self.handlers[*topic]
	if len(handlers) == 1 {
		delete(self.handlers, *topic)
		return
	}
	delete(handlers, h)
}

// Add a Public key address mapping
// this is needed to initiate handshakes
func (self *Pss) SetPeerPublicKey(pubkey *ecdsa.PublicKey, topic whisper.TopicType, address PssAddress) {
	self.lock.Lock()
	defer self.lock.Unlock()
	psp := &pssPeer{
		address: address,
	}
	pubkeyid := common.ToHex(crypto.FromECDSAPub(pubkey))
	if _, ok := self.pubKeyPool[pubkeyid]; ok == false {
		self.pubKeyPool[pubkeyid] = make(map[whisper.TopicType]*pssPeer)
	}
	self.pubKeyPool[pubkeyid][topic] = psp
}

// Get a Public Key from id
func (self *Pss) GetPeerPublicKey(pubkeyid string, topic whisper.TopicType) (*ecdsa.PublicKey, *pssPeer) {
	pubkey := crypto.ToECDSAPub(common.FromHex(pubkeyid))
	if pubkey.X == nil {
		return nil, nil
	}
	return pubkey, self.pubKeyPool[pubkeyid][topic]
}

// Automatically generate a new symkey for a topic and address hint
func (self *Pss) generateSymmetricKey(topic whisper.TopicType, address PssAddress, expires time.Duration, addToCache bool) (string, error) {
	keyid, err := self.w.GenerateSymKey()
	if err != nil {
		return "", err
	}
	self.addSymmetricKeyToPool(keyid, topic, address, expires, addToCache)
	return keyid, nil
}

// Manually set a new symkey for a topic and address hint
//
// If addtocache is set to true, the key will be added to the collection of keys used to attempt incoming message decryption
func (self *Pss) SetSymmetricKey(key []byte, topic whisper.TopicType, address PssAddress, expires time.Duration, addtocache bool) (string, error) {
	keyid, err := self.w.AddSymKeyDirect(key)
	if err != nil {
		return "", err
	}
	self.addSymmetricKeyToPool(keyid, topic, address, expires, addtocache)
	return keyid, nil
}

// adds a symmetric key to the pss key pool, and optionally adds the key to the collection of keys used to attempt incoming message decryption
func (self *Pss) addSymmetricKeyToPool(keyid string, topic whisper.TopicType, address PssAddress, expires time.Duration, addtocache bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	var psp *pssPeer
	if expires == 0 {
		expires = symKeyExpiry
	}
	if _, ok := self.symKeyPool[keyid]; ok == false {
		self.symKeyPool[keyid] = make(map[whisper.TopicType]*pssPeer)
		psp = &pssPeer{}
		self.symKeyPool[keyid][topic] = psp
	}
	psp = self.symKeyPool[keyid][topic]
	psp.expires = time.Now().Add(expires)
	psp.address = address
	if addtocache {
		self.symKeyCacheCursor++
		self.symKeyCache[self.symKeyCacheCursor%cap(self.symKeyCache)] = &keyid
	}
}

// add a message to the cache
func (self *Pss) addFwdCache(digest pssDigest) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	var entry pssCacheEntry
	var ok bool
	if entry, ok = self.fwdcache[digest]; !ok {
		entry = pssCacheEntry{}
	}
	entry.expiresAt = time.Now().Add(self.cachettl)
	self.fwdcache[digest] = entry
	return nil
}

// check if message is in the cache
func (self *Pss) checkFwdCache(addr []byte, digest pssDigest) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	entry, ok := self.fwdcache[digest]
	if ok {
		if entry.expiresAt.After(time.Now()) {
			log.Debug(fmt.Sprintf("unexpired cache for digest %x", digest))
			return true
		} else if entry.expiresAt.IsZero() && bytes.Equal(addr, entry.receivedFrom) {
			log.Debug(fmt.Sprintf("sendermatch %x for digest %x", common.ByteLabel(addr), digest))
			return true
		}
	}
	return false
}

// DPA storage handler for message cache
func (self *Pss) storeMsg(msg *PssMsg) (pssDigest, error) {
	swg := &sync.WaitGroup{}
	wwg := &sync.WaitGroup{}
	buf := bytes.NewReader(msg.serialize())
	key, err := self.dpa.Store(buf, int64(buf.Len()), swg, wwg)
	if err != nil {
		log.Warn("Could not store in swarm", "err", err)
		return pssDigest{}, err
	}
	log.Trace("Stored msg in swarm", "key", key)
	digest := pssDigest{}
	copy(digest[:], key[:digestLength])
	return digest, nil
}

// get all registered handlers for respective topics
func (self *Pss) getHandlers(topic whisper.TopicType) map[*Handler]bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.handlers[topic]
}

// filters incoming messages for processing or forwarding
// check if address partially matches = CAN be for us = process
// terminates main protocol handler if payload is not valid pssmsg
func (self *Pss) handlePssMsg(msg interface{}) error {
	pssmsg, ok := msg.(*PssMsg)
	if ok {
		if !self.isSelfPossibleRecipient(pssmsg) {
			log.Trace("pss was for someone else :'( ... forwarding")
			return self.forward(pssmsg)
		}
		log.Trace("pss for us, yay! ... let's process!")

		return self.process(pssmsg)
	}

	return fmt.Errorf("invalid message type. Expected *PssMsg, got %T ", msg)
}

// Entry point to processing a message for which the current node can be the intended recipient.
// Attempts symmetric and asymmetric decryption with stored keys.
// Calls key processing if it's a handshake message
// Dispatches message to all handlers matching the message topic
func (self *Pss) process(pssmsg *PssMsg) error {

	// Holds a successfully decrypted message
	var recvmsg *whisper.ReceivedMessage

	// Matched address from successfully decrypted message
	var from PssAddress

	var err error

	// Set to true if the decrypted message was symmetrically encrypted
	var asymmetric bool

	// the symkeyid or hex representation of the pubkey used to encrypt the message
	var keyid string

	envelope := pssmsg.Payload

	if len(envelope.AESNonce) > 0 { // detect symkey msg according to whisperv5/envelope.go:OpenSymmetric
		// check if message can be decrypted symmetrically
		recvmsg, keyid, err = self.processSym(envelope)
		if err != nil {
			return err
		}
		from = self.symKeyPool[keyid][envelope.Topic].address
	} else {
		// check if message can be decrypted asymmetrically
		// if it cannot, we let it fall through
		asymmetric = true
		var pubkey []byte
		recvmsg, pubkey, err = self.processAsym(envelope)

		if err == nil {
			// check if message decodes to a psskeymsg
			// if no let it pass through
			// TODO: handshake initiation flood guard
			keyid = common.ToHex(pubkey)
			keymsg := &pssKeyMsg{}
			err = rlp.DecodeBytes(recvmsg.Payload, keymsg)
			if err == nil {
				return self.handleKey(pubkey, envelope, keymsg)
			}
			from = self.pubKeyPool[keyid][envelope.Topic].address
		}
	}

	// fail if the topic doesn't have a matching handler
	// note that a handshake request recipient will not have a matching handler
	// so we cannot check this prior to processing psskeymsg
	handlers := self.getHandlers(envelope.Topic)
	if len(handlers) == 0 {
		return fmt.Errorf("No registered handler for topic '%x'", envelope.Topic)
	}
	// this condition checks if we either have a successfully decrypted asym msg that's not a pssKeyMsg
	// OR if it's a successfully decrypted sym msg
	// if so we know for sure it's for this pss node
	if recvmsg != nil {
		// if we have partial address we will perform redundant forwarding
		// so noone can use traffic analysis to deduce that the messages stopped with us
		if len(pssmsg.To) < addressLength {
			go func() {
				err := self.forward(pssmsg)
				if err != nil {
					log.Warn("Redundant forward fail: %v", err)
				}
			}()
		}
		handlers := self.getHandlers(envelope.Topic)
		nid, _ := discover.HexID("0x00") // this hack is needed to satisfy the p2p method
		p := p2p.NewPeer(nid, fmt.Sprintf("%x", from), []p2p.Cap{})
		for f := range handlers {
			err := (*f)(recvmsg.Payload, p, asymmetric, keyid)
			if err != nil {
				log.Warn("Pss handler %p failed: %v", err)
			}
		}
	}

	return nil
}

// attempt to decrypt, validate and unpack sym msg
// if successful, returns the unpacked whisper ReceivedMessage struct
// encapsulating the decrypted message, and the symkeyid of the symkey
// used to decrypt the message, the latter revealing the sender
// It fails if decryption of the message fails or if the message is corrupted
func (self *Pss) processSym(envelope *whisper.Envelope) (*whisper.ReceivedMessage, string, error) {
	for i := self.symKeyCacheCursor; i > self.symKeyCacheCursor-cap(self.symKeyCache) && i > 0; i-- {
		symkeyid := self.symKeyCache[i%cap(self.symKeyCache)]
		log.Trace("attempting symmetric decrypt", "symkey", symkeyid)
		symkey, err := self.w.GetSymKey(*symkeyid)
		if err != nil {
			log.Debug("could not retrieve whisper symkey id %v: %v", symkeyid, err)
			continue
		}
		recvmsg, err := envelope.OpenSymmetric(symkey)
		if err != nil {
			log.Trace("sym decrypt failed", "symkey", symkeyid, "err", err)
			continue
		}
		if !recvmsg.Validate() {
			return nil, "", fmt.Errorf("symmetrically encrypted message has invalid signature or is corrupt")
		}
		from := self.symKeyPool[*symkeyid][envelope.Topic].address
		self.symKeyCacheCursor++
		self.symKeyCache[self.symKeyCacheCursor%cap(self.symKeyCache)] = symkeyid
		log.Debug("successfully decrypted symmetrically encrypted pss message", "symkeys tried", i, "from", from, "symkey cache insert", self.symKeyCacheCursor%cap(self.symKeyCache))
		return recvmsg, *symkeyid, nil
	}
	return nil, "", nil
}

// attempt to decrypt, validate and unpack asym msg
// if successful, returns the unpacked whisper ReceivedMessage struct
// encapsulating the decrypted message, and the byte representation of
// the pubkey used to decrypt the message, the latter revealing the sender
// fails if decryption of message fails, or if the message is corrupted
func (self *Pss) processAsym(envelope *whisper.Envelope) (*whisper.ReceivedMessage, []byte, error) {
	recvmsg, err := envelope.OpenAsymmetric(self.privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("asym default decrypt of pss msg failed: %v", "err", err)
	}
	// check signature (if signed), strip padding
	if !recvmsg.Validate() {
		return nil, nil, fmt.Errorf("invalid message")
	}
	pubkey := crypto.FromECDSAPub(recvmsg.Src)
	from := self.pubKeyPool[common.ToHex(pubkey)][envelope.Topic].address
	log.Debug("successfully decrypted asymmetrically encrypted pss message", "from", from, "pubkey", pubkey)
	return recvmsg, pubkey, nil
}

// generate and send symkey to peer using asym send (handshake)
// if a symkeyid string is passed, that symkey will be used to encrypt
// the key transfer. This will constitute a handshake response if the
// key from an unencrypted key transfer is used for the encryption.
// the method fails if symkey could not be generated, if an unknown
// symkeyid was passed, if symmetric encryption of the key transfer
// fails, or if the key transfer rlp encoding or send fails
func (self *Pss) sendKey(pubkey []byte, topic *whisper.TopicType, symkeyid string, to PssAddress) (string, error) {
	var nonce []byte
	recvkeyid, err := self.generateSymmetricKey(*topic, to, symKeyRequestExpiry, true)
	log.Trace("sending our symkey", "pubkey", pubkey, "symkey", recvkeyid)
	if err != nil {
		return "", fmt.Errorf("set receive symkey fail (addr %x pubkey %x topic %x): %v", to, pubkey, topic, err)
	}
	recvkey, err := self.w.GetSymKey(recvkeyid)
	if err != nil {
		return "", fmt.Errorf("get generated outgoing symkey fail (addr %x pubkey %x topic %x): %v", to, pubkey, topic, err)
	}
	if symkeyid != "" {
		symkey, err := self.w.GetSymKey(symkeyid)
		if err != nil {
			return "", fmt.Errorf("Invalid symkey for key transfer: %s", symkeyid)
		}
		log.Debug("before encrypt", "key", recvkey)
		recvkey, nonce, err = whisper.EncryptSymmetric(symkey, recvkey)
		log.Debug("after encrypt", "key", recvkey)
		if err != nil {
			return "", fmt.Errorf("Symkey transfer encrypt fail: %v", err)
		}
	}
	recvkeymsg := &pssKeyMsg{
		From:  self.BaseAddr(),
		Key:   recvkey,
		Nonce: nonce,
	}
	recvkeybytes, err := rlp.EncodeToBytes(recvkeymsg)
	if err != nil {
		return "", fmt.Errorf("rlp keymsg encode fail: %v", err)
	}
	// if the send fails it means this public key is not registered for this particular address AND topic
	err = self.SendAsym(common.ToHex(pubkey), *topic, recvkeybytes)
	if err != nil {
		return "", fmt.Errorf("Send symkey failed: %v", err)
	}
	return recvkeyid, nil
}

// handles an incoming keymsg
// fails if send of a keymsg response fails, or if whisper symkey store fails
func (self *Pss) handleKey(pubkey []byte, envelope *whisper.Envelope, keymsg *pssKeyMsg) error {
	if len(keymsg.Nonce) == 0 {
		// check if the key in the keymsg is symmetrically encrypted
		// if it's not, this is a handshake initiation
		log.Trace("have handshake request", "from", keymsg.From)
		// TODO: need to handle / check for expired keys also here
		sendsymkeyid, err := self.SetSymmetricKey(keymsg.Key, envelope.Topic, keymsg.From, 0, false)
		if err != nil {
			return err
		}
		// reply with an encrypted secret so that it can tell that we received its key
		// the encrypted secret will be our key encrypted with its key
		recvsymkeyid, err := self.sendKey(pubkey, &envelope.Topic, sendsymkeyid, keymsg.From)
		if err != nil {
			return err
		}
		// the key we received is paired with the key we sent
		// at this point the key combination is considered a valid
		// handshake by the responding node.
		// i.e. we do not wait for additional confirmation from the peer
		self.symKeyPairIndex[recvsymkeyid] = &sendsymkeyid
		self.symKeyPairIndex[sendsymkeyid] = &recvsymkeyid
		recvsymkey, err := self.w.GetSymKey(recvsymkeyid)
		if err != nil {
			return err
		}
		pubkeyid := common.ToHex(pubkey)
		self.symKeyPairPubKey[common.ToHex(recvsymkey)] = pubkeyid
		log.Trace("added handshake request mapping", "pubkey", pubkeyid, "symkey", common.ToHex(recvsymkey))
		log.Trace("added handshake request key", "recvkeyid", recvsymkeyid, "sendkeyid", sendsymkeyid, "pubkey", pubkeyid)
		return nil
	}

	log.Trace("have handshake response", "from", keymsg.From)
	// if not, try to decrypt the message payload with the symkeys on file
	// if it decrypts it should contain a keymsg
	// if yes, this is a handshake response
	for i := self.symKeyCacheCursor; i > self.symKeyCacheCursor-cap(self.symKeyCache) && i > 0; i-- {
		recvsymkeyid := self.symKeyCache[i%cap(self.symKeyCache)]
		recvsymkey, err := self.w.GetSymKey(*recvsymkeyid)
		if err != nil {
			continue
		}
		sendsymkey, err := whisper.DecryptSymmetric(recvsymkey, keymsg.Nonce, keymsg.Key)
		if err == nil {
			log.Trace("decrypted response", "msg", sendsymkey)
			sendsymkeyid, err := self.SetSymmetricKey(sendsymkey, envelope.Topic, keymsg.From, 0, false)
			if err != nil {
				return err
			}
			// the key we received is paired with the key we used to decrypt the message
			// at this point the key combination is considered a valid
			// handshake by the originally requesting node
			self.symKeyPool[*recvsymkeyid][envelope.Topic].expires = time.Unix(defaultSymKeyExpiry, 0)
			self.symKeyPairIndex[*recvsymkeyid] = &sendsymkeyid
			self.symKeyPairIndex[sendsymkeyid] = recvsymkeyid
			pubkeyid := common.ToHex(pubkey)
			self.symKeyPairPubKey[common.ToHex(recvsymkey)] = pubkeyid
			if _, ok := self.handshakeC[pubkeyid]; ok {
				self.alertHandshake(pubkeyid, sendsymkeyid)
			}
			log.Trace("added handshake mapping", "pubkey", pubkeyid, "symkey", common.ToHex(recvsymkey))
			log.Trace("added handshake response key", "recvkeyid", *recvsymkeyid, "sendkeyid", sendsymkeyid, "pubkey", pubkeyid)
		}
	}

	return nil
}

func (self *Pss) alertHandshake(pubkey string, symkey string) chan string {
	if _, ok := self.handshakeC[pubkey]; !ok {
		self.handshakeC[pubkey] = make(chan string)
	}
	if symkey != "" {
		self.handshakeC[pubkey] <- symkey
		close(self.handshakeC[pubkey])
		delete(self.handshakeC, pubkey)
		return nil
	}
	return self.handshakeC[pubkey]
}

// Prepares a msg for sending with symmetric encryption
//
// fails if the passed symkeyid is invalid, or if the
// symkeyid is not part of an unexpired pair of symkeys
// established by a mutual sendKey combination (pss.sendKey)
func (self *Pss) SendSym(symkeyid string, topic whisper.TopicType, msg []byte) error {
	psp := self.symKeyPool[symkeyid][topic]
	sendsymkeyid := self.symKeyPairIndex[symkeyid]
	if sendsymkeyid == nil {
		return fmt.Errorf("missing matching send symkey")
	}
	symkey, err := self.w.GetSymKey(*sendsymkeyid)
	if err != nil {
		return fmt.Errorf("missing valid send symkey %s: %v", symkeyid, err)
	}
	return self.send(psp.address, topic, msg, false, symkey)
}

// Prepares a msg for sending with asymmetric encryption
//
// Fails if the pubkey hex representation passed does not
// match any saved pubkeys
func (self *Pss) SendAsym(pubkeyid string, topic whisper.TopicType, msg []byte) error {
	pubkey, psp := self.GetPeerPublicKey(pubkeyid, topic)
	if pubkey == nil {
		return fmt.Errorf("Invalid public key id %x", pubkey)
	}
	return self.send(psp.address, topic, msg, true, common.FromHex(pubkeyid))
}

// pss send is payload agnostic, and will accept any byte slice as payload
// It generates an whisper envelope for the specified recipient and topic,
// and wraps the message payload in it.
// TODO: Implement proper message padding
func (self *Pss) send(to []byte, topic whisper.TopicType, msg []byte, asymmetric bool, key []byte) error {
	if key == nil || bytes.Equal(key, []byte{}) {
		return fmt.Errorf("Zero length key passed to pss send")
	}
	wparams := &whisper.MessageParams{
		TTL:      DefaultTTL,
		Src:      self.privateKey,
		Topic:    topic,
		WorkTime: defaultWhisperWorkTime,
		PoW:      defaultWhisperPoW,
		Payload:  msg,
		Padding:  []byte("1234567890abcdef"),
	}
	if asymmetric {
		wparams.Dst = crypto.ToECDSAPub(key)
	} else {
		wparams.KeySym = key
	}
	// set up outgoing message container, which does encryption and envelope wrapping
	woutmsg, err := whisper.NewSentMessage(wparams)
	if err != nil {
		return fmt.Errorf("failed to generate whisper message encapsulation: %v", err)
	}
	// performs encryption.
	// Does NOT perform / performs negligible PoW due to very low difficulty setting
	// after this the message is ready for sending
	envelope, err := woutmsg.Wrap(wparams)
	if err != nil {
		return fmt.Errorf("failed to perform whisper encryption: %v", err)
	}
	log.Trace("pssmsg whisper done", "env", envelope, "wparams payload", wparams.Payload, "to", to, "asym", asymmetric, "key", key)
	// prepare for devp2p transport
	pssmsg := &PssMsg{
		To:      to,
		Payload: envelope,
	}
	return self.forward(pssmsg)
}

// Forwards a pss message to the peer(s) closest to the to recipient address in the PssMsg struct
// The recipient address can be of any length, and the byte slice will be matched to the MSB slice
// of the peer address of the equivalent length.
// Handlers that are merely passing on the PssMsg to its final recipient might call this directly
func (self *Pss) forward(msg *PssMsg) error {
	to := make([]byte, addressLength)
	copy(to[:len(msg.To)], msg.To)

	// cache the message
	digest, err := self.storeMsg(msg)
	if err != nil {
		log.Warn(fmt.Sprintf("could not store message %v to cache: %v", msg, err))
	}

	// flood guard:
	// don't allow identical messages we saw shortly before
	if self.checkFwdCache(nil, digest) {
		log.Trace(fmt.Sprintf("pss relay block-cache match: FROM %x TO %x", common.ByteLabel(self.Overlay.BaseAddr()), common.ByteLabel(msg.To)))
		return nil
	}

	// send with kademlia
	// find the closest peer to the recipient and attempt to send
	sent := 0

	self.Overlay.EachConn(to, 256, func(op network.OverlayConn, po int, isproxbin bool) bool {
		sendMsg := fmt.Sprintf("MSG %x TO %x FROM %x VIA %x", digest, common.ByteLabel(to), common.ByteLabel(self.BaseAddr()), common.ByteLabel(op.Address()))
		// we need p2p.protocols.Peer.Send
		// cast and resolve
		sp, ok := op.(senderPeer)
		if !ok {
			log.Crit("Pss cannot use kademlia peer type")
			return false
		}
		pp := self.fwdPool[sp.ID()]
		if self.checkFwdCache(op.Address(), digest) {
			log.Info(fmt.Sprintf("%v: peer already forwarded to", sendMsg))
			return true
		}
		// attempt to send the message
		err := pp.Send(msg)
		if err != nil {
			log.Warn(fmt.Sprintf("%v: failed forwarding: %v", sendMsg, err))
			return true
		}
		log.Debug(fmt.Sprintf("%v: successfully forwarded", sendMsg))
		sent++
		// continue forwarding if:
		// - if the peer is end recipient but the full address has not been disclosed
		// - if the peer address matches the partial address fully
		// - if the peer is in proxbin
		if len(msg.To) < addressLength && bytes.Equal(msg.To, op.Address()[:len(msg.To)]) {
			log.Trace(fmt.Sprintf("Pss keep forwarding: Partial address + full partial match"))
			return true
		} else if isproxbin {
			log.Trace(fmt.Sprintf("%x is in proxbin, keep forwarding", common.ByteLabel(op.Address())))
			return true
		}
		// at this point we stop forwarding, and the state is as follows:
		// - the peer is end recipient and we have full address
		// - we are not in proxbin (directed routing)
		// - partial addresses don't fully match
		return false
	})

	if sent == 0 {
		return fmt.Errorf("unable to forward to any peers")
	}

	self.addFwdCache(digest)
	return nil
}

// will return false if using partial address
func (self *Pss) isSelfRecipient(msg *PssMsg) bool {
	return bytes.Equal(msg.To, self.Overlay.BaseAddr())
}

func (self *Pss) isSelfPossibleRecipient(msg *PssMsg) bool {
	local := self.Overlay.BaseAddr()
	return bytes.Equal(msg.To[:], local[:len(msg.To)])
}

// todo: maybe not enough to check that the symkey id strings are empty
func (self *Pss) SymStatus(symkeyid string, topic whisper.TopicType) int {
	if _, ok := self.symKeyPool[symkeyid]; ok == false {
		return SYMSTATUS_NONE
	}
	psp := self.symKeyPool[symkeyid][topic]
	if psp.expires.Before(time.Now()) {
		return SYMSTATUS_EXPIRED
	}
	rsymkeyid, ok := self.symKeyPairIndex[symkeyid]
	if ok == false {
		return SYMSTATUS_NONE
	}
	if symkeyid != *self.symKeyPairIndex[*rsymkeyid] {
		return SYMSTATUS_PENDING
	}
	return SYMSTATUS_OK
}

// get public key of peer from symkey peer sent message with
func (self *Pss) GetPublicKeyFromSymmetricKey(symkey []byte) string {
	return self.symKeyPairPubKey[common.ToHex(symkey)]
}
