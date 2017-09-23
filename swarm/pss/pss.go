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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

// TODO: proper padding generation for messages
const (
	defaultDigestCacheTTL      = time.Second
	defaultSymKeyCacheCapacity = 512
	digestLength               = 32 // byte length of digest used for pss cache (currently same as swarm chunk hash)
	defaultWhisperWorkTime     = 3
	defaultWhisperPoW          = 0.0000000001
	defaultMaxMsgSize          = 1024 * 1024
	defaultCleanInterval       = 1000 * 60 * 10
)

var (
	addressLength = len(pot.Address{})
)

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
	lastSeen  time.Time
	address   *PssAddress
	protected bool
}

// Pss configuration parameters
type PssParams struct {
	CacheTTL            time.Duration
	privateKey          *ecdsa.PrivateKey
	SymKeyCacheCapacity int
}

// Sane defaults for Pss
func NewPssParams(privatekey *ecdsa.PrivateKey) *PssParams {
	return &PssParams{
		CacheTTL:            defaultDigestCacheTTL,
		privateKey:          privatekey,
		SymKeyCacheCapacity: defaultSymKeyCacheCapacity,
	}
}

// Toplevel pss object, takes care of message sending, receiving, decryption and encryption, message handler dispatchers and message forwarding.
//
// Implements node.Service
type Pss struct {
	network.Overlay // we can get the overlayaddress from this
	privateKey      *ecdsa.PrivateKey
	dpa             *storage.DPA
	w               *whisper.Whisper
	auxAPIs         []rpc.API
	lock            sync.Mutex
	quitC           chan struct{}

	// forwarding
	fwdPool  map[discover.NodeID]*protocols.Peer // keep track of all peers sitting on the pssmsg routing layer
	fwdCache map[pssDigest]pssCacheEntry         // checksum of unique fields from pssmsg mapped to expiry, cache to determine whether to drop msg
	cacheTTL time.Duration                       // how long to keep messages in fwdCache

	// keys and peers
	pubKeyPool          map[string]map[whisper.TopicType]*pssPeer // mapping of hex public keys to peer address. We use string because we need unified interface to pass key to registered handlers
	symKeyPool          map[string]map[whisper.TopicType]*pssPeer // mapping of symkeyids to peer address
	symKeyCache         []*string                                 // fast lookup of recently used symkeys; last used is on top of stack
	symKeyCacheCursor   int                                       // modular cursor pointing to last used, wraps on symKeyCache array
	symKeyCacheCapacity int                                       // max amount of symkeys to keep.

	// message handling
	handlers map[whisper.TopicType]map[*Handler]bool // topic and version based pss payload handlers
}

func (self *Pss) String() string {
	return fmt.Sprintf("pss: addr %x, pubkey %v", self.BaseAddr(), common.ToHex(crypto.FromECDSAPub(&self.privateKey.PublicKey)))
}

// Creates a new Pss instance.
//
// Needs a swarm network overlay, a DPA storage for message cache storage.
func NewPss(k network.Overlay, dpa *storage.DPA, params *PssParams) *Pss {
	return &Pss{
		Overlay:    k,
		privateKey: params.privateKey,
		dpa:        dpa,
		w:          whisper.New(),
		quitC:      make(chan struct{}),

		fwdPool:  make(map[discover.NodeID]*protocols.Peer),
		fwdCache: make(map[pssDigest]pssCacheEntry),
		cacheTTL: params.CacheTTL,

		pubKeyPool:          make(map[string]map[whisper.TopicType]*pssPeer),
		symKeyPool:          make(map[string]map[whisper.TopicType]*pssPeer),
		symKeyCache:         make([]*string, params.SymKeyCacheCapacity),
		symKeyCacheCapacity: params.SymKeyCacheCapacity,

		handlers: make(map[whisper.TopicType]map[*Handler]bool),
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
			self.clean()
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

func (self *Pss) addAPI(api rpc.API) {
	self.auxAPIs = append(self.auxAPIs, api)
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
	}
	for _, auxapi := range self.auxAPIs {
		apis = append(apis, auxapi)
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
func (self *Pss) SetPeerPublicKey(pubkey *ecdsa.PublicKey, topic whisper.TopicType, address *PssAddress) {
	self.lock.Lock()
	defer self.lock.Unlock()
	pubkeyid := common.ToHex(crypto.FromECDSAPub(pubkey))
	psp := &pssPeer{
		address: address,
	}
	if _, ok := self.pubKeyPool[pubkeyid]; ok == false {
		self.pubKeyPool[pubkeyid] = make(map[whisper.TopicType]*pssPeer)
	}
	self.pubKeyPool[pubkeyid][topic] = psp
	log.Trace("added pubkey", "pubkeyid", pubkeyid, "topic", topic, "address", address)
}

// Automatically generate a new symkey for a topic and address hint
func (self *Pss) generateSymmetricKey(topic whisper.TopicType, address *PssAddress, sendlimit uint16, addToCache bool) (string, error) {
	keyid, err := self.w.GenerateSymKey()
	if err != nil {
		return "", err
	}
	self.addSymmetricKeyToPool(keyid, topic, address, sendlimit, addToCache)
	return keyid, nil
}

// Manually set a new symkey for a topic and address hint
//
// If addtocache is set to true, the key will be added to the collection of keys used to attempt incoming message decryption
func (self *Pss) SetSymmetricKey(key []byte, topic whisper.TopicType, address *PssAddress, sendlimit uint16, addtocache bool) (string, error) {
	keyid, err := self.w.AddSymKeyDirect(key)
	if err != nil {
		return "", err
	}
	self.addSymmetricKeyToPool(keyid, topic, address, sendlimit, addtocache)
	return keyid, nil
}

// adds a symmetric key to the pss key pool, and optionally adds the key to the collection of keys used to attempt incoming message decryption
func (self *Pss) addSymmetricKeyToPool(keyid string, topic whisper.TopicType, address *PssAddress, sendlimit uint16, addtocache bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	psp := &pssPeer{
		address: address,
	}
	if _, ok := self.symKeyPool[keyid]; !ok {
		self.symKeyPool[keyid] = make(map[whisper.TopicType]*pssPeer)
	}
	self.symKeyPool[keyid][topic] = psp
	if addtocache {
		self.symKeyCacheCursor++
		self.symKeyCache[self.symKeyCacheCursor%cap(self.symKeyCache)] = &keyid
	}
	log.Trace("added symkey", "symkeyid", keyid, "topic", topic, "address", address, "cache", addtocache)
}

// Resolves whisper symkey id to symkey bytes
//
// Expired symkeys will not be returned
func (self *Pss) GetSymmetricKey(symkeyid string) ([]byte, error) {
	symkey, err := self.w.GetSymKey(symkeyid)
	if err != nil {
		return nil, err
	}
	return symkey, nil
}

// symkey garbage collection
func (self *Pss) clean() {
	for keyid, peertopics := range self.symKeyPool {
		var expiredtopics []whisper.TopicType
		for topic, psp := range peertopics {
			log.Trace("check topic", "topic", topic, "id", keyid)
			var match bool
			if psp.protected {
				continue
			}

			for i := self.symKeyCacheCursor; i > self.symKeyCacheCursor-cap(self.symKeyCache) && i > 0; i-- {
				cacheid := self.symKeyCache[i%cap(self.symKeyCache)]
				log.Trace("check cache", "idx", i, "id", *cacheid)
				if *cacheid == keyid {
					match = true
				}
			}
			if match == false {
				expiredtopics = append(expiredtopics, topic)
			}
		}
		for _, topic := range expiredtopics {
			delete(self.symKeyPool[keyid], topic)
			log.Trace("symkey cleanup deletion", "symkeyid", keyid, "topic", topic, "val", self.symKeyPool[keyid])
		}
	}
}

// add a message to the cache
func (self *Pss) addFwdCache(digest pssDigest) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	var entry pssCacheEntry
	var ok bool
	if entry, ok = self.fwdCache[digest]; !ok {
		entry = pssCacheEntry{}
	}
	entry.expiresAt = time.Now().Add(self.cacheTTL)
	self.fwdCache[digest] = entry
	return nil
}

// check if message is in the cache
func (self *Pss) checkFwdCache(addr []byte, digest pssDigest) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	entry, ok := self.fwdCache[digest]
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
	var err error
	var recvmsg *whisper.ReceivedMessage
	var from *PssAddress
	var asymmetric bool
	var keyid string
	var keyFunc func(envelope *whisper.Envelope) (*whisper.ReceivedMessage, string, *PssAddress, error)

	envelope := pssmsg.Payload

	handlers := self.getHandlers(envelope.Topic)
	if len(handlers) == 0 {
		return fmt.Errorf("No registered handler for topic '%x'", envelope.Topic)
	}

	if len(envelope.AESNonce) > 0 { // detect symkey msg according to whisperv5/envelope.go:OpenSymmetric
		keyFunc = self.processSym
	} else {
		asymmetric = true
		keyFunc = self.processAsym
	}
	recvmsg, keyid, from, err = keyFunc(envelope)
	if err != nil {
		log.Trace("decrypt message fail", "err", err, "asym", asymmetric)
	}

	if recvmsg != nil {
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
				log.Warn("Pss handler %p failed: %v", f, err)
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
func (self *Pss) processSym(envelope *whisper.Envelope) (*whisper.ReceivedMessage, string, *PssAddress, error) {
	for i := self.symKeyCacheCursor; i > self.symKeyCacheCursor-cap(self.symKeyCache) && i > 0; i-- {
		symkeyid := self.symKeyCache[i%cap(self.symKeyCache)]
		symkey, err := self.w.GetSymKey(*symkeyid)
		if err != nil {
			continue
		}
		recvmsg, err := envelope.OpenSymmetric(symkey)
		if err != nil {
			continue
		}
		if !recvmsg.Validate() {
			return nil, "", nil, errors.New(fmt.Sprintf("symmetrically encrypted message has invalid signature or is corrupt"))
		}
		from := self.symKeyPool[*symkeyid][envelope.Topic].address
		self.symKeyCacheCursor++
		self.symKeyCache[self.symKeyCacheCursor%cap(self.symKeyCache)] = symkeyid
		return recvmsg, *symkeyid, from, nil
	}
	return nil, "", nil, nil
}

// attempt to decrypt, validate and unpack asym msg
// if successful, returns the unpacked whisper ReceivedMessage struct
// encapsulating the decrypted message, and the byte representation of
// the pubkey used to decrypt the message, the latter revealing the sender
// fails if decryption of message fails, or if the message is corrupted
func (self *Pss) processAsym(envelope *whisper.Envelope) (*whisper.ReceivedMessage, string, *PssAddress, error) {
	recvmsg, err := envelope.OpenAsymmetric(self.privateKey)
	if err != nil {
		return nil, "", nil, errors.New(fmt.Sprintf("asym default decrypt of pss msg failed: %v", "err", err))
	}
	// check signature (if signed), strip padding
	if !recvmsg.Validate() {
		return nil, "", nil, errors.New("invalid message")
	}
	pubkeyid := common.ToHex(crypto.FromECDSAPub(recvmsg.Src))
	from := self.pubKeyPool[pubkeyid][envelope.Topic].address
	return recvmsg, pubkeyid, from, nil
}

// Prepares a msg for sending with symmetric encryption
//
// fails if the passed symkeyid is invalid, or if the symkey has expired
func (self *Pss) SendSym(symkeyid string, topic whisper.TopicType, msg []byte) error {
	symkey, err := self.GetSymmetricKey(symkeyid)
	if err != nil {
		return errors.New(fmt.Sprintf("missing valid send symkey %s: %v", symkeyid, err))
	}
	psp := self.symKeyPool[symkeyid][topic]
	err = self.send(*psp.address, topic, msg, false, symkey)
	return err
}

// Prepares a msg for sending with asymmetric encryption
//
// Fails if the pubkey hex representation passed does not match any saved pubkeys
func (self *Pss) SendAsym(pubkeyid string, topic whisper.TopicType, msg []byte) error {
	//pubkey := self.pubKeyIndex[pubkeyid]
	pubkey := crypto.ToECDSAPub(common.FromHex(pubkeyid))
	if pubkey == nil {
		return fmt.Errorf("Invalid public key id %x", pubkey)
	}
	psp := self.pubKeyPool[pubkeyid][topic]
	return self.send(*psp.address, topic, msg, true, common.FromHex(pubkeyid))
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
