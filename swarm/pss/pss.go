package pss

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
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
	defaultDigestCacheTTL       = time.Second
	defaultSymKeyCacheCapacity  = 512
	digestLength                = 32 // byte length of digest used for pss cache (currently same as swarm chunk hash)
	defaultSymKeyBufferCapacity = 6
	defaultSymKeySendLimit      = 1024
	defaultSymKeyRequestExpiry  = 5000
	defaultWhisperWorkTime      = 3
	defaultWhisperPoW           = 0.0000000001
	defaultMaxMsgSize           = 1024 * 1024
	defaultSymKeyLength         = 32
	defaultCleanInterval        = 1000 * 60 * 10
)

var (
	addressLength = len(pot.Address{})
)

// used for exchange of symkeys in handshake exchange
//
// using this struct a node can one or both of:
// - send new keys to a peer to be used for sending messages to the node
// - request new keys from a peer to be used for sending messages to the peer
//
// Sent keys are concatenated in the Keys member.
// The keys have equal length, specified by the KeyLength member.
// The Limit-member specifies for how many messages the keys are valid.
// If the peer sends more than this number of messages using the same key,
// delivery of the message will silently fail.
//
// Amount of new keys requested from peer is specified in the RequestCont member
//
// The address hint for the peer is stored in the From field.
type pssKeyMsg struct {
	From         []byte
	Limit        uint16
	Keys         []byte
	KeyLength    uint8
	RequestCount uint8
}

type pssCacheEntry struct {
	expiresAt    time.Time
	receivedFrom []byte
}

// abstraction to enable access to p2p.protocols.Peer.Send
type senderPeer interface {
	ID() discover.NodeID
	Address() []byte
	Send(interface{}) error
}

// per-key peer related information
// sendCount stores how many messages this key has been used for
// sendLimit stores how many messages this key is valid for
type pssPeer struct {
	sendLimit uint16
	sendCount uint16
	address   PssAddress
}

// Pss configuration parameters
type PssParams struct {
	CacheTTL             time.Duration
	privateKey           *ecdsa.PrivateKey
	SymKeyRequestExpiry  time.Duration
	SymKeyCacheCapacity  int
	SymKeyBufferCapacity int
	SymKeySendLimit      uint16
}

// Sane defaults for Pss
func NewPssParams(privatekey *ecdsa.PrivateKey) *PssParams {
	return &PssParams{
		CacheTTL:             defaultDigestCacheTTL,
		privateKey:           privatekey,
		SymKeyCacheCapacity:  defaultSymKeyCacheCapacity,
		SymKeyRequestExpiry:  time.Millisecond * defaultSymKeyRequestExpiry,
		SymKeySendLimit:      defaultSymKeySendLimit,
		SymKeyBufferCapacity: defaultSymKeyBufferCapacity,
	}
}

// Toplevel pss object, takes care of message sending, receiving, decryption and encryption, message handler dispatchers and message forwarding.
//
// Implements node.Service
type Pss struct {
	network.Overlay                                                // we can get the overlayaddress from this
	fwdPool              map[discover.NodeID]*protocols.Peer       // keep track of all peers sitting on the pssmsg routing layer
	pubKeyPool           map[string]map[whisper.TopicType]*pssPeer // mapping of hex public keys to peer address. We use string because we need unified interface to pass key to registered handlers
	symKeyPool           map[string]*pssPeer                       // mapping of symkeyids to peer address
	handlers             map[whisper.TopicType]map[*Handler]bool   // topic and version based pss payload handlers
	fwdcache             map[pssDigest]pssCacheEntry               // checksum of unique fields from pssmsg mapped to expiry, cache to determine whether to drop msg
	cachettl             time.Duration                             // how long to keep messages in fwdcache
	lock                 sync.Mutex
	dpa                  *storage.DPA
	privateKey           *ecdsa.PrivateKey
	w                    *whisper.Whisper
	symKeyPubKeyIndex    map[string]string // look up matching pubkey for handshake send symkey
	pubKeySymKeyIndex    map[string]map[whisper.TopicType][]*string
	pubKeyIndex          map[string]*ecdsa.PublicKey
	symKeyCache          []*string                // fast lookup of recently used symkeys; last used is on top of stack
	symKeyCacheCursor    int                      // modular cursor pointing to last used, wraps on symKeyCache array
	handshakeC           map[string]chan []string // adds a channel to report when a handshake succeeds
	symKeyCacheCapacity  int                      // max amount of symkeys to keep.
	symKeyRequestExpiry  time.Duration            // max wait time to receive a response to a handshake symkey request
	symKeySendLimit      uint16                   // amount of messages a symkey is valid for
	symKeyBufferCapacity int                      // amount of hanshake-negotiated outgoing symkeys kept simultaneously
	quitC                chan struct{}
}

func (self *Pss) String() string {
	return fmt.Sprintf("pss: addr %x, pubkey %v", self.BaseAddr(), common.ToHex(crypto.FromECDSAPub(&self.privateKey.PublicKey)))
}

// Creates a new Pss instance.
//
// Needs a swarm network overlay, a DPA storage for message cache storage.
func NewPss(k network.Overlay, dpa *storage.DPA, params *PssParams) *Pss {
	return &Pss{
		Overlay:              k,
		fwdPool:              make(map[discover.NodeID]*protocols.Peer),
		symKeyPool:           make(map[string]*pssPeer),
		pubKeyPool:           make(map[string]map[whisper.TopicType]*pssPeer),
		handlers:             make(map[whisper.TopicType]map[*Handler]bool),
		fwdcache:             make(map[pssDigest]pssCacheEntry),
		cachettl:             params.CacheTTL,
		dpa:                  dpa,
		privateKey:           params.privateKey,
		w:                    whisper.New(),
		symKeyPubKeyIndex:    make(map[string]string),
		pubKeySymKeyIndex:    make(map[string]map[whisper.TopicType][]*string),
		pubKeyIndex:          make(map[string]*ecdsa.PublicKey),
		symKeyCache:          make([]*string, params.SymKeyCacheCapacity),
		handshakeC:           make(map[string]chan []string),
		symKeyCacheCapacity:  params.SymKeyCacheCapacity,
		symKeyRequestExpiry:  params.SymKeyRequestExpiry,
		symKeySendLimit:      params.SymKeySendLimit,
		symKeyBufferCapacity: params.SymKeyBufferCapacity,
		quitC:                make(chan struct{}),
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
	go func() {
		tickC := time.Tick(defaultCleanInterval)
		select {
		case <-tickC:
			self.cleanSymmetricKeys()
		case <-self.quitC:
			log.Info("pss shutting down")
		}
	}()
	return nil
}

// For node.Service implementation. Does nothing for now, but should be included in the code for backwards compatibility.
func (self *Pss) Stop() error {
	close(self.quitC)
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
			Version:   "0.2",
			Service:   NewAPI(self),
			Public:    true,
		},
		rpc.API{
			Namespace: "psstest",
			Version:   "0.2",
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
	self.pubKeyIndex[pubkeyid] = pubkey
	if _, ok := self.pubKeyPool[pubkeyid]; ok == false {
		self.pubKeyPool[pubkeyid] = make(map[whisper.TopicType]*pssPeer)
	}
	self.pubKeyPool[pubkeyid][topic] = psp
}

// Automatically generate a new symkey for a topic and address hint
func (self *Pss) generateSymmetricKey(address PssAddress, sendlimit uint16, addToCache bool) (string, error) {
	keyid, err := self.w.GenerateSymKey()
	if err != nil {
		return "", err
	}
	self.addSymmetricKeyToPool(keyid, address, sendlimit, addToCache)
	return keyid, nil
}

// Manually set a new symkey for a topic and address hint
//
// If addtocache is set to true, the key will be added to the collection of keys used to attempt incoming message decryption
func (self *Pss) SetSymmetricKey(key []byte, address PssAddress, sendlimit uint16, addtocache bool) (string, error) {
	keyid, err := self.w.AddSymKeyDirect(key)
	if err != nil {
		return "", err
	}
	self.addSymmetricKeyToPool(keyid, address, sendlimit, addtocache)
	return keyid, nil
}

// adds a symmetric key to the pss key pool, and optionally adds the key to the collection of keys used to attempt incoming message decryption
func (self *Pss) addSymmetricKeyToPool(keyid string, address PssAddress, sendlimit uint16, addtocache bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if sendlimit == 0 {
		sendlimit = self.symKeySendLimit
	}
	if _, ok := self.symKeyPool[keyid]; ok == false {
		self.symKeyPool[keyid] = &pssPeer{}
	}
	psp := self.symKeyPool[keyid]
	psp.sendLimit = sendlimit
	psp.address = address
	if addtocache {
		self.symKeyCacheCursor++
		self.symKeyCache[self.symKeyCacheCursor%cap(self.symKeyCache)] = &keyid
	}
	log.Trace("added symkey", "symkeyid", keyid, "address", address, "sendlimit", sendlimit, "cache", addtocache)
}

// returns all symkeys that are active for respective public keys after handshake exchange
func (self *Pss) getSymmetricKeyBuffer(pubkeyid string, topic *whisper.TopicType) (symkeyids []string, remaining []uint16) {
	if _, ok := self.pubKeySymKeyIndex[pubkeyid]; !ok {
		return
	}
	for _, symkeyid := range self.pubKeySymKeyIndex[pubkeyid][*topic] {
		capacity, _ := self.GetSymmetricKeyCapacity(*symkeyid)
		if capacity == 0 {
			continue
		}
		symkeyids = append(symkeyids, *symkeyid)
		remaining = append(remaining, capacity)
	}
	return
}

// Resolves whisper symkey id to symkey bytes
//
// Expired symkeys will not be returned
func (self *Pss) GetSymmetricKey(symkeyid string) ([]byte, error) {
	capacity, err := self.GetSymmetricKeyCapacity(symkeyid)
	if err != nil {
		return nil, err
	}
	if capacity == 0 {
		return nil, errors.New("expired")
	}
	symkey, err := self.w.GetSymKey(symkeyid)
	if err != nil {
		return nil, err
	}
	return symkey, nil
}

// checks if symkey is valid for more messages.
// if not, the symkey will be instantly garbage collected.
func (self *Pss) GetSymmetricKeyCapacity(symkeyid string) (uint16, error) {
	if _, ok := self.symKeyPool[symkeyid]; !ok {
		return 0, errors.New(fmt.Sprintf("Invalid symkeyid %s", symkeyid))
	}
	capacity := self.symKeyPool[symkeyid].sendLimit - self.symKeyPool[symkeyid].sendCount
	if capacity == 0 {
		delete(self.symKeyPool, symkeyid)
	}
	return capacity, nil
}

// symkey garbage collection
func (self *Pss) cleanSymmetricKeys() {
	var expiredkeyids []string
	for keyid, psp := range self.symKeyPool {
		var match bool
		if psp.sendLimit <= psp.sendCount {
			log.Trace("cleanup expired symkey", "id", keyid)
			expiredkeyids = append(expiredkeyids, keyid)
			continue
		} else {
			for _, cacheid := range self.symKeyCache {
				if *cacheid == keyid {
					match = true
					log.Trace("cleanup cache match", "id", keyid)
				}
			}
		}
		if match == false {
			expiredkeyids = append(expiredkeyids, keyid)
		}
	}
	for _, keyid := range expiredkeyids {
		self.lock.Lock()
		delete(self.symKeyPubKeyIndex, keyid)
		self.lock.Unlock()
		for _, topicmap := range self.pubKeySymKeyIndex {
			for _, indexkeys := range topicmap {
				for i, indexkeyid := range indexkeys {
					if *indexkeyid == keyid {
						self.lock.Lock()
						indexkeys[i] = indexkeys[len(indexkeys)-1]
						indexkeys = indexkeys[:len(indexkeys)-1]
						self.lock.Unlock()
					}
				}
			}
		}
		self.lock.Lock()
		delete(self.symKeyPool, keyid)
		self.lock.Unlock()
		log.Debug("symkey deleted", "symkey", keyid)
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
		if recvmsg != nil {
			from = self.symKeyPool[keyid].address
		}
	} else {
		// check if message can be decrypted asymmetrically
		// if it cannot, we let it fall through
		asymmetric = true
		recvmsg, keyid, err = self.processAsym(envelope)

		if err == nil && recvmsg != nil {
			// check if message decodes to a psskeymsg
			// if no let it pass through
			// TODO: handshake initiation flood guard
			keymsg := &pssKeyMsg{}
			err = rlp.DecodeBytes(recvmsg.Payload, keymsg)
			if err == nil {
				err := self.handleKey(keyid, envelope, keymsg)
				if err != nil {
					log.Error("handlekey fail", "error", err)
				}
				return err
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
		from := self.symKeyPool[*symkeyid].address
		self.symKeyCacheCursor++
		self.symKeyCache[self.symKeyCacheCursor%cap(self.symKeyCache)] = symkeyid
		log.Debug("successfully decrypted symmetrically encrypted pss message", "symkeys tried", i, "from", common.ToHex(from), "symkey cache insert", self.symKeyCacheCursor%cap(self.symKeyCache))
		return recvmsg, *symkeyid, nil
	}
	return nil, "", nil
}

// attempt to decrypt, validate and unpack asym msg
// if successful, returns the unpacked whisper ReceivedMessage struct
// encapsulating the decrypted message, and the byte representation of
// the pubkey used to decrypt the message, the latter revealing the sender
// fails if decryption of message fails, or if the message is corrupted
func (self *Pss) processAsym(envelope *whisper.Envelope) (*whisper.ReceivedMessage, string, error) {
	recvmsg, err := envelope.OpenAsymmetric(self.privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("asym default decrypt of pss msg failed: %v", "err", err)
	}
	// check signature (if signed), strip padding
	if !recvmsg.Validate() {
		return nil, "", fmt.Errorf("invalid message")
	}
	pubkeyid := common.ToHex(crypto.FromECDSAPub(recvmsg.Src))
	from := self.pubKeyPool[pubkeyid][envelope.Topic].address
	log.Debug("successfully decrypted asymmetrically encrypted pss message", "from", from, "pubkeyid", pubkeyid)
	return recvmsg, pubkeyid, nil
}

// send and request symkeys from peer through asym send of a keymsg
// will generate and send "keycount" new keys, valid for "msglimit" messages
// if symkey buffer for the passed pubkey is not full, it will request the difference amount of new keys from the peer.
// returns empty (does not fail) if buffer is full and "keycount" is zero
// fails if symkey could not be generated, or if rlp encode or send of the keymsg fails
func (self *Pss) sendKey(pubkeyid string, topic *whisper.TopicType, keycount uint8, msglimit uint16, to PssAddress) ([]string, error) {
	//var nonce []byte
	recvkeys := make([]byte, keycount*defaultSymKeyLength)
	recvkeyids := make([]string, keycount)

	// check if buffer is not full
	_, counts := self.getSymmetricKeyBuffer(pubkeyid, topic)
	requestcount := uint8(self.symKeyBufferCapacity - len(counts))

	// return if there's nothing to be accomplished
	if requestcount == 0 && len(counts) == 0 && keycount == 0 {
		return []string{}, nil
	}

	// generate new keys to send
	for i := 0; i < len(recvkeyids); i++ {
		var err error
		recvkeyids[i], err = self.generateSymmetricKey(to, msglimit, true)
		if err != nil {
			return []string{}, fmt.Errorf("set receive symkey fail (addr %x pubkey %x topic %x): %v", to, pubkeyid, topic, err)
		}
		recvkey, err := self.w.GetSymKey(recvkeyids[i])
		if err != nil {
			return []string{}, fmt.Errorf("get generated outgoing symkey fail (addr %x pubkey %x topic %x): %v", to, pubkeyid, topic, err)
		}
		offset := i * defaultSymKeyLength
		copy(recvkeys[offset:offset+defaultSymKeyLength], recvkey)
		self.symKeyPubKeyIndex[recvkeyids[i]] = pubkeyid
	}

	// encode and send the message
	recvkeymsg := &pssKeyMsg{
		From:         self.BaseAddr(),
		Keys:         recvkeys,
		KeyLength:    defaultSymKeyLength,
		RequestCount: requestcount,
		Limit:        self.symKeySendLimit,
	}
	log.Trace("sending our symkeys", "pubkey", pubkeyid, "symkeys", recvkeyids, "limit", self.symKeySendLimit, "requestcount", requestcount, "keycount", len(recvkeys))
	recvkeybytes, err := rlp.EncodeToBytes(recvkeymsg)
	if err != nil {
		return []string{}, fmt.Errorf("rlp keymsg encode fail: %v", err)
	}
	// if the send fails it means this public key is not registered for this particular address AND topic
	err = self.SendAsym(pubkeyid, *topic, recvkeybytes)
	if err != nil {
		return []string{}, fmt.Errorf("Send symkey failed: %v", err)
	}
	return recvkeyids, nil
}

// handles an incoming keymsg
// processes and adds new keys included in the keymsg
// calls sendKey with the RequestCount in the keymsg as the "keycount" param
// fails if send of a keymsg response fails, or if whisper symkey store fails
// TODO: terminate adding or drop msg if amount of keys (or request) is over an upper bound
func (self *Pss) handleKey(pubkeyid string, envelope *whisper.Envelope, keymsg *pssKeyMsg) error {
	// new keys from peer
	if len(keymsg.Keys) > 0 {
		if _, ok := self.pubKeySymKeyIndex[pubkeyid]; !ok {
			self.pubKeySymKeyIndex[pubkeyid] = make(map[whisper.TopicType][]*string)
		}
		log.Trace("keys from peer", "from", keymsg.From, "count", uint8(len(keymsg.Keys))/keymsg.KeyLength)

		var sendsymkeyids []string
		for i := 0; i < len(keymsg.Keys); i += int(keymsg.KeyLength) {
			sendsymkey := make([]byte, keymsg.KeyLength)
			copy(sendsymkey, keymsg.Keys[i:i+int(keymsg.KeyLength)])
			sendsymkeyid, err := self.SetSymmetricKey(sendsymkey, keymsg.From, keymsg.Limit, false)
			if err != nil {
				return err
			}
			self.pubKeySymKeyIndex[pubkeyid][envelope.Topic] = append(self.pubKeySymKeyIndex[pubkeyid][envelope.Topic], &sendsymkeyid)
			sendsymkeyids = append(sendsymkeyids, sendsymkeyid)
		}
		if len(sendsymkeyids) > 0 {
			self.alertHandshake(pubkeyid, sendsymkeyids)
		}
	}

	// peer request for keys
	if keymsg.RequestCount > 0 {
		log.Trace("keys to peer", "from", keymsg.From, "count", keymsg.RequestCount)
		// we don't need to remember the key ids here
		//_, err := self.sendKey(pubkeyid, &envelope.Topic, keymsg.RequestCount, self.symKeySendLimit, sendsymkeyid, keymsg.From)
		_, err := self.sendKey(pubkeyid, &envelope.Topic, keymsg.RequestCount, self.symKeySendLimit, keymsg.From)
		if err != nil {
			return err
		}
	}

	return nil
}

// used to enable blocked key requests to peer
// if passed without symkey a new keyid array channel is created with pubkeyid as key
// if passed with symkeyids, and the channel on the pubkeyid is active, symkeyids are passed to channel
func (self *Pss) alertHandshake(pubkeyid string, symkeys []string) chan []string {
	if len(symkeys) > 0 {
		if _, ok := self.handshakeC[pubkeyid]; ok {
			self.handshakeC[pubkeyid] <- symkeys
			close(self.handshakeC[pubkeyid])
			delete(self.handshakeC, pubkeyid)
		}
		return nil
	} else {
		if _, ok := self.handshakeC[pubkeyid]; !ok {
			self.handshakeC[pubkeyid] = make(chan []string)
		}
	}
	return self.handshakeC[pubkeyid]
}

// Prepares a msg for sending with symmetric encryption
//
// fails if the passed symkeyid is invalid, or if the symkey has expired
func (self *Pss) SendSym(symkeyid string, topic whisper.TopicType, msg []byte) error {
	symkey, err := self.GetSymmetricKey(symkeyid)
	if err != nil {
		return errors.New(fmt.Sprintf("missing valid send symkey %s: %v", symkeyid, err))
	}
	psp := self.symKeyPool[symkeyid]
	err = self.send(psp.address, topic, msg, false, symkey)
	if err == nil {
		self.symKeyPool[symkeyid].sendCount++
	}
	return err
}

// Prepares a msg for sending with asymmetric encryption
//
// Fails if the pubkey hex representation passed does not match any saved pubkeys
func (self *Pss) SendAsym(pubkeyid string, topic whisper.TopicType, msg []byte) error {
	pubkey := self.pubKeyIndex[pubkeyid]
	if pubkey == nil {
		return fmt.Errorf("Invalid public key id %x", pubkey)
	}
	psp := self.pubKeyPool[pubkeyid][topic]
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
		log.Trace(fmt.Sprintf("%v: successfully forwarded", sendMsg))
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
