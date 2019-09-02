// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package pss

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/pss/crypto"
	"github.com/ethersphere/swarm/pss/internal/ticker"
	"github.com/ethersphere/swarm/pss/internal/ttlset"
	"github.com/ethersphere/swarm/pss/message"
	"github.com/tilinna/clock"
)

const (
	defaultMsgTTL              = time.Second * 120
	defaultDigestCacheTTL      = time.Second * 10
	defaultSymKeyCacheCapacity = 512
	defaultMaxMsgSize          = 1024 * 1024
	defaultCleanInterval       = time.Second * 60 * 10
	defaultOutboxCapacity      = 100000
	protocolName               = "pss"
	protocolVersion            = 2
)

var (
	addressLength = len(pot.Address{})
)

var spec = &protocols.Spec{
	Name:       protocolName,
	Version:    protocolVersion,
	MaxMsgSize: defaultMaxMsgSize,
	Messages: []interface{}{
		message.Message{},
	},
}

// abstraction to enable access to p2p.protocols.Peer.Send
type senderPeer interface {
	Info() *p2p.PeerInfo
	ID() enode.ID
	Address() []byte
	Send(context.Context, interface{}) error
}

// per-key peer related information
// member `protected` prevents garbage collection of the instance
type peer struct {
	lastSeen  time.Time
	address   PssAddress
	protected bool
}

// Pss configuration parameters
type Params struct {
	MsgTTL              time.Duration
	CacheTTL            time.Duration
	privateKey          *ecdsa.PrivateKey
	SymKeyCacheCapacity int
	AllowRaw            bool // If true, enables sending and receiving messages without builtin pss encryption
}

// Sane defaults for Pss
func NewParams() *Params {
	return &Params{
		MsgTTL:              defaultMsgTTL,
		CacheTTL:            defaultDigestCacheTTL,
		SymKeyCacheCapacity: defaultSymKeyCacheCapacity,
	}
}

func (params *Params) WithPrivateKey(privatekey *ecdsa.PrivateKey) *Params {
	params.privateKey = privatekey
	return params
}

type outbox struct {
	queue   []*outboxMsg
	slots   chan int
	process chan int
	quitC   chan struct{}
	forward func(msg *message.Message) error
}

func newOutbox(capacity int, quitC chan struct{}, forward func(msg *message.Message) error) outbox {
	outbox := outbox{
		queue:   make([]*outboxMsg, capacity),
		slots:   make(chan int, capacity),
		process: make(chan int),
		quitC:   quitC,
		forward: forward,
	}
	// fill up outbox slots
	for i := 0; i < cap(outbox.slots); i++ {
		outbox.slots <- i
	}
	return outbox
}

func (o outbox) len() int {
	return cap(o.slots) - len(o.slots)
}

// enqueue a new element in the outbox if there is any slot available.
// Then send it to process. This method is blocking in the process channel!
func (o *outbox) enqueue(outboxmsg *outboxMsg) error {
	// first we try to obtain a slot in the outbox
	select {
	case slot := <-o.slots:
		o.queue[slot] = outboxmsg
		metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
		// we send this message slot to process
		select {
		case o.process <- slot:
		case <-o.quitC:
		}
		return nil
	default:
		metrics.GetOrRegisterCounter("pss.enqueue.outbox.full", nil).Inc(1)
		return errors.New("outbox full")
	}
}

func (o *outbox) processOutbox() {
	for slot := range o.process {
		go func(slot int) {
			msg := o.msg(slot)
			metrics.GetOrRegisterResettingTimer("pss.handle.outbox", nil).UpdateSince(msg.startedAt)
			if err := o.forward(msg.msg); err != nil {
				metrics.GetOrRegisterCounter("pss.forward.err", nil).Inc(1)
				// if we failed to forward, re-insert message in the queue
				log.Debug(err.Error())
				// reenqueue the message for processing
				o.reenqueue(slot)
				log.Debug("Message re-enqued", "slot", slot)
				return
			}
			// free the outbox slot
			o.free(slot)
			metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
		}(slot)
	}
}

func (o outbox) msg(slot int) *outboxMsg {
	return o.queue[slot]
}

func (o outbox) free(slot int) {
	o.slots <- slot
}

func (o outbox) reenqueue(slot int) {
	select {
	case o.process <- slot:
	case <-o.quitC:
	}

}

// Pss is the top-level struct, which takes care of message sending, receiving, decryption and encryption, message handler dispatchers
// and message forwarding. Implements node.Service
type Pss struct {
	*network.Kademlia // we can get the Kademlia address from this
	*KeyStore
	forwardCache *ttlset.TTLSet
	gcTicker     *ticker.Ticker

	privateKey *ecdsa.PrivateKey // pss can have it's own independent key
	auxAPIs    []rpc.API         // builtins (handshake, test) can add APIs

	// sending and forwarding
	peers   map[string]*protocols.Peer // keep track of all peers sitting on the pssmsg routing layer
	peersMu sync.RWMutex

	msgTTL    time.Duration
	capstring string
	outbox    outbox

	// message handling
	handlers           map[message.Topic]map[*handler]bool // topic and version based pss payload handlers. See pss.Handle()
	handlersMu         sync.RWMutex
	topicHandlerCaps   map[message.Topic]*handlerCaps // caches capabilities of each topic's handlers
	topicHandlerCapsMu sync.RWMutex

	// process
	quitC chan struct{}
}

func (p *Pss) String() string {
	return fmt.Sprintf("pss: addr %x, pubkey %v", p.BaseAddr(), hex.EncodeToString(p.Crypto.SerializePublicKey(&p.privateKey.PublicKey)))
}

// Creates a new Pss instance.
//
// In addition to params, it takes a swarm network Kademlia
// and a FileStore storage for message cache storage.
func New(k *network.Kademlia, params *Params) (*Pss, error) {
	if params.privateKey == nil {
		return nil, errors.New("missing private key for pss")
	}

	clock := clock.Realtime() //TODO: Clock should be injected by Params so it can be mocked.

	c := p2p.Cap{
		Name:    protocolName,
		Version: protocolVersion,
	}
	ps := &Pss{
		Kademlia: k,
		KeyStore: loadKeyStore(),

		privateKey: params.privateKey,
		quitC:      make(chan struct{}),

		peers:     make(map[string]*protocols.Peer),
		msgTTL:    params.MsgTTL,
		capstring: c.String(),

		handlers:         make(map[message.Topic]map[*handler]bool),
		topicHandlerCaps: make(map[message.Topic]*handlerCaps),
	}
	ps.forwardCache = ttlset.New(&ttlset.Config{
		EntryTTL: params.CacheTTL,
		Clock:    clock,
	})
	ps.gcTicker = ticker.New(&ticker.Config{
		Clock:    clock,
		Interval: params.CacheTTL,
		Callback: func() {
			ps.forwardCache.GC()
			metrics.GetOrRegisterCounter("pss.cleanfwdcache", nil).Inc(1)
		},
	})
	ps.outbox = newOutbox(defaultOutboxCapacity, ps.quitC, ps.forward)

	return ps, nil
}

/////////////////////////////////////////////////////////////////////
// SECTION: node.Service interface
/////////////////////////////////////////////////////////////////////

func (p *Pss) Start(srv *p2p.Server) error {
	go func() {
		ticker := time.NewTicker(defaultCleanInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.cleanKeys()
			case <-p.quitC:
				return
			}
		}
	}()

	// Forward outbox messages
	go p.outbox.processOutbox()

	log.Info("Started Pss")
	log.Info("Loaded EC keys", "pubkey", hex.EncodeToString(p.Crypto.SerializePublicKey(p.PublicKey())), "secp256", hex.EncodeToString(p.Crypto.CompressPublicKey(p.PublicKey())))
	return nil
}

func (p *Pss) Stop() error {
	log.Info("Pss shutting down")
	if err := p.gcTicker.Stop(); err != nil {
		return err
	}
	close(p.quitC)
	return nil
}

func (p *Pss) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    spec.Name,
			Version: spec.Version,
			Length:  spec.Length(),
			Run:     p.Run,
		},
	}
}

func (p *Pss) Run(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	pp := protocols.NewPeer(peer, rw, spec)
	p.addPeer(pp)
	defer p.removePeer(pp)
	return pp.Run(p.handle)
}

func (p *Pss) getPeer(peer *protocols.Peer) (pp *protocols.Peer, ok bool) {
	p.peersMu.RLock()
	defer p.peersMu.RUnlock()
	pp, ok = p.peers[peer.Peer.Info().ID]
	return
}

func (p *Pss) addPeer(peer *protocols.Peer) {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()
	p.peers[peer.Peer.Info().ID] = peer
}

func (p *Pss) removePeer(peer *protocols.Peer) {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()
	log.Trace("removing peer", "id", peer.Peer.Info().ID)
	delete(p.peers, peer.Peer.Info().ID)
}

func (p *Pss) APIs() []rpc.API {
	apis := []rpc.API{
		{
			Namespace: "pss",
			Version:   "1.0",
			Service:   NewAPI(p),
			Public:    true,
		},
	}
	apis = append(apis, p.auxAPIs...)
	return apis
}

// add API methods to the pss API
// must be run before node is started
func (p *Pss) addAPI(api rpc.API) {
	p.auxAPIs = append(p.auxAPIs, api)
}

// Returns the swarm Kademlia address of the pss node
func (p *Pss) BaseAddr() []byte {
	return p.Kademlia.BaseAddr()
}

// Returns the pss node's public key
func (p *Pss) PublicKey() *ecdsa.PublicKey {
	return &p.privateKey.PublicKey
}

/////////////////////////////////////////////////////////////////////
// SECTION: Message handling
/////////////////////////////////////////////////////////////////////

func (p *Pss) getTopicHandlerCaps(topic message.Topic) (hc *handlerCaps, found bool) {
	p.topicHandlerCapsMu.RLock()
	defer p.topicHandlerCapsMu.RUnlock()
	hc, found = p.topicHandlerCaps[topic]
	return
}

func (p *Pss) setTopicHandlerCaps(topic message.Topic, hc *handlerCaps) {
	p.topicHandlerCapsMu.Lock()
	defer p.topicHandlerCapsMu.Unlock()
	p.topicHandlerCaps[topic] = hc
}

// Links a handler function to a Topic
//
// All incoming messages with an envelope Topic matching the
// topic specified will be passed to the given Handler function.
//
// There may be an arbitrary number of handler functions per topic.
//
// Returns a deregister function which needs to be called to
// deregister the handler,
func (p *Pss) Register(topic *message.Topic, hndlr *handler) func() {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	handlers := p.handlers[*topic]
	if handlers == nil {
		handlers = make(map[*handler]bool)
		p.handlers[*topic] = handlers
		log.Debug("registered handler", "capabilities", hndlr.caps)
	}
	if hndlr.caps == nil {
		hndlr.caps = &handlerCaps{}
	}
	handlers[hndlr] = true

	capabilities, ok := p.getTopicHandlerCaps(*topic)
	if !ok {
		capabilities = &handlerCaps{}
		p.setTopicHandlerCaps(*topic, capabilities)
	}

	if hndlr.caps.raw {
		capabilities.raw = true
	}
	if hndlr.caps.prox {
		capabilities.prox = true
	}
	return func() { p.deregister(topic, hndlr) }
}

func (p *Pss) deregister(topic *message.Topic, hndlr *handler) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	handlers := p.handlers[*topic]
	if len(handlers) > 1 {
		delete(p.handlers, *topic)
		// topic caps might have changed now that a handler is gone
		caps := &handlerCaps{}
		for h := range handlers {
			if h.caps.raw {
				caps.raw = true
			}
			if h.caps.prox {
				caps.prox = true
			}
		}
		p.setTopicHandlerCaps(*topic, caps)
		return
	}
	delete(handlers, hndlr)
}

// Filters incoming messages for processing or forwarding.
// Check if address partially matches
// If yes, it CAN be for us, and we process it
// Only passes error to pss protocol handler if payload is not valid pssmsg
func (p *Pss) handle(ctx context.Context, msg interface{}) error {
	defer metrics.GetOrRegisterResettingTimer("pss.handle", nil).UpdateSince(time.Now())

	pssmsg, ok := msg.(*message.Message)
	if !ok {
		return fmt.Errorf("invalid message type. Expected *message.Message, got %T", msg)
	}
	log.Trace("handler", "self", label(p.Kademlia.BaseAddr()), "topic", label(pssmsg.Topic[:]))
	if int64(pssmsg.Expire) < time.Now().Unix() {
		metrics.GetOrRegisterCounter("pss.expire", nil).Inc(1)
		log.Warn("pss filtered expired message", "from", hex.EncodeToString(p.Kademlia.BaseAddr()), "to", hex.EncodeToString(pssmsg.To))
		return nil
	}
	if p.checkFwdCache(pssmsg) {
		log.Trace("pss relay block-cache match (process)", "from", hex.EncodeToString(p.Kademlia.BaseAddr()), "to", (hex.EncodeToString(pssmsg.To)))
		return nil
	}
	p.addFwdCache(pssmsg)

	psstopic := pssmsg.Topic

	// raw is simplest handler contingency to check, so check that first
	var isRaw bool
	if pssmsg.Flags.Raw {
		if capabilities, ok := p.getTopicHandlerCaps(psstopic); ok {
			if !capabilities.raw {
				log.Warn("No handler for raw message", "topic", label(psstopic[:]))
				return nil
			}
		}
		isRaw = true
	}

	// check if we can be recipient:
	// - no prox handler on message and partial address matches
	// - prox handler on message and we are in prox regardless of partial address match
	// store this result so we don't calculate again on every handler
	var isProx bool
	if capabilities, ok := p.getTopicHandlerCaps(psstopic); ok {
		isProx = capabilities.prox
	}
	isRecipient := p.isSelfPossibleRecipient(pssmsg, isProx)
	if !isRecipient {
		log.Trace("pss msg forwarding ===>", "pss", hex.EncodeToString(p.BaseAddr()), "prox", isProx)
		return p.enqueue(pssmsg)
	}

	log.Trace("pss msg processing <===", "pss", hex.EncodeToString(p.BaseAddr()), "prox", isProx, "raw", isRaw, "topic", label(pssmsg.Topic[:]))
	if err := p.process(pssmsg, isRaw, isProx); err != nil {
		qerr := p.enqueue(pssmsg)
		if qerr != nil {
			return fmt.Errorf("process fail: processerr %v, queueerr: %v", err, qerr)
		}
	}
	return nil
}

// Entry point to processing a message for which the current node can be the intended recipient.
// Attempts symmetric and asymmetric decryption with stored keys.
// Dispatches message to all handlers matching the message topic
func (p *Pss) process(pssmsg *message.Message, raw bool, prox bool) error {
	defer metrics.GetOrRegisterResettingTimer("pss.process", nil).UpdateSince(time.Now())

	var err error
	var payload []byte
	var from PssAddress
	var asymmetric bool
	var keyid string
	var keyFunc func(pssMsg *message.Message) ([]byte, string, PssAddress, error)

	psstopic := pssmsg.Topic

	if raw {
		payload = pssmsg.Payload
	} else {
		if pssmsg.Flags.Symmetric {
			keyFunc = p.processSym
		} else {
			asymmetric = true
			keyFunc = p.processAsym
		}

		payload, keyid, from, err = keyFunc(pssmsg)
		if err != nil {
			return errors.New("decryption failed")
		}
	}

	if len(pssmsg.To) < addressLength || prox {
		err = p.enqueue(pssmsg)
	}
	p.executeHandlers(psstopic, payload, from, raw, prox, asymmetric, keyid)
	return err
}

// copy all registered handlers for respective topic in order to avoid data race or deadlock
func (p *Pss) getHandlers(topic message.Topic) (ret []*handler) {
	p.handlersMu.RLock()
	defer p.handlersMu.RUnlock()
	for k := range p.handlers[topic] {
		ret = append(ret, k)
	}
	return ret
}

func (p *Pss) executeHandlers(topic message.Topic, payload []byte, from PssAddress, raw bool, prox bool, asymmetric bool, keyid string) {
	defer metrics.GetOrRegisterResettingTimer("pss.execute-handlers", nil).UpdateSince(time.Now())

	handlers := p.getHandlers(topic)
	peer := p2p.NewPeer(enode.ID{}, hex.EncodeToString(from), []p2p.Cap{})
	for _, h := range handlers {
		if !h.caps.raw && raw {
			log.Warn("norawhandler")
			continue
		}
		if !h.caps.prox && prox {
			log.Warn("noproxhandler")
			continue
		}
		err := (h.f)(payload, peer, asymmetric, keyid)
		if err != nil {
			log.Warn("Pss handler failed", "err", err)
		}
	}
}

// will return false if using partial address
func (p *Pss) isSelfRecipient(msg *message.Message) bool {
	return bytes.Equal(msg.To, p.Kademlia.BaseAddr())
}

// test match of leftmost bytes in given message to node's Kademlia address
func (p *Pss) isSelfPossibleRecipient(msg *message.Message, prox bool) bool {
	local := p.Kademlia.BaseAddr()

	// if a partial address matches we are possible recipient regardless of prox
	// if not and prox is not set, we are surely not
	if bytes.Equal(msg.To, local[:len(msg.To)]) {

		return true
	} else if !prox {
		return false
	}

	depth := p.NeighbourhoodDepth()
	po, _ := network.Pof(p.Kademlia.BaseAddr(), msg.To, 0)
	log.Trace("selfpossible", "po", po, "depth", depth)

	return depth <= po
}

/////////////////////////////////////////////////////////////////////
// SECTION: Message sending
/////////////////////////////////////////////////////////////////////

func (p *Pss) enqueue(msg *message.Message) error {
	defer metrics.GetOrRegisterResettingTimer("pss.enqueue", nil).UpdateSince(time.Now())

	outboxmsg := newOutboxMsg(msg)

	return p.outbox.enqueue(outboxmsg)
}

// Send a raw message (any encryption is responsibility of calling client)
//
// Will fail if raw messages are disallowed
func (p *Pss) SendRaw(address PssAddress, topic message.Topic, msg []byte) error {
	defer metrics.GetOrRegisterResettingTimer("pss.send.raw", nil).UpdateSince(time.Now())

	if err := validateAddress(address); err != nil {
		return err
	}

	pssMsgParams := message.Flags{
		Raw: true,
	}

	pssMsg := message.New(pssMsgParams)
	pssMsg.To = address
	pssMsg.Expire = uint32(time.Now().Add(p.msgTTL).Unix())
	pssMsg.Payload = msg
	pssMsg.Topic = topic

	p.addFwdCache(pssMsg)

	return p.enqueue(pssMsg)
}

// Send a message using symmetric encryption
//
// Fails if the key id does not match any of the stored symmetric keys
func (p *Pss) SendSym(symkeyid string, topic message.Topic, msg []byte) error {
	symkey, err := p.GetSymmetricKey(symkeyid)
	if err != nil {
		return fmt.Errorf("missing valid send symkey %s: %v", symkeyid, err)
	}
	psp, ok := p.getPeerSym(symkeyid, topic)
	if !ok {
		return fmt.Errorf("invalid topic '%s' for symkey '%s'", topic.String(), symkeyid)
	}
	return p.send(psp.address, topic, msg, false, symkey)
}

// Send a message using asymmetric encryption
//
// Fails if the key id does not match any in of the stored public keys
func (p *Pss) SendAsym(pubkeyid string, topic message.Topic, msg []byte) error {
	if _, err := p.Crypto.UnmarshalPublicKey(common.FromHex(pubkeyid)); err != nil {
		return fmt.Errorf("Cannot unmarshal pubkey: %x", pubkeyid)
	}
	psp, ok := p.getPeerPub(pubkeyid, topic)
	if !ok {
		return fmt.Errorf("invalid topic '%s' for pubkey '%s'", topic.String(), pubkeyid)
	}
	return p.send(psp.address, topic, msg, true, common.FromHex(pubkeyid))
}

// Send is payload agnostic, and will accept any byte slice as payload
// It generates an envelope for the specified recipient and topic,
// and wraps the message payload in it.
// TODO: Implement proper message padding
func (p *Pss) send(to []byte, topic message.Topic, msg []byte, asymmetric bool, key []byte) error {
	metrics.GetOrRegisterCounter("pss.send", nil).Inc(1)

	if key == nil || bytes.Equal(key, []byte{}) {
		return fmt.Errorf("Zero length key passed to pss send")
	}
	wrapParams := &crypto.WrapParams{
		Sender: p.privateKey,
	}
	if asymmetric {
		pk, err := p.Crypto.UnmarshalPublicKey(key)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal pubkey: %x", key)
		}
		wrapParams.Receiver = pk
	} else {
		wrapParams.SymmetricKey = key
	}
	// set up outgoing message container, which does encryption and envelope wrapping
	envelope, err := p.Crypto.Wrap(msg, wrapParams)
	if err != nil {
		return fmt.Errorf("failed to perform message encapsulation and encryption: %v", err)
	}
	log.Trace("pssmsg wrap done", "env", envelope, "mparams payload", hex.EncodeToString(msg), "to", hex.EncodeToString(to), "asym", asymmetric, "key", hex.EncodeToString(key))

	// prepare for devp2p transport
	pssMsgParams := message.Flags{
		Symmetric: !asymmetric,
	}
	pssMsg := message.New(pssMsgParams)
	pssMsg.To = to
	pssMsg.Expire = uint32(time.Now().Add(p.msgTTL).Unix())
	pssMsg.Payload = envelope
	pssMsg.Topic = topic

	return p.enqueue(pssMsg)
}

// sendFunc is a helper function that tries to send a message and returns true on success.
// It is set here for usage in production, and optionally overridden in tests.
var sendFunc = sendMsg

// tries to send a message, returns true if successful
func sendMsg(p *Pss, sp *network.Peer, msg *message.Message) bool {
	var isPssEnabled bool
	info := sp.Info()
	for _, capability := range info.Caps {
		if capability == p.capstring {
			isPssEnabled = true
			break
		}
	}
	if !isPssEnabled {
		log.Warn("peer doesn't have matching pss capabilities, skipping", "peer", info.Name, "caps", info.Caps, "peer", label(sp.BzzAddr.Address()))
		return false
	}

	// get the protocol peer from the forwarding peer cache
	pp, ok := p.getPeer(sp.BzzPeer.Peer)
	if !ok {
		log.Warn("peer no longer in our list, dropping message")
		return false
	}

	err := pp.Send(context.TODO(), msg)
	if err != nil {
		metrics.GetOrRegisterCounter("pss.pp.send.error", nil).Inc(1)
		log.Error(err.Error())
	}

	return err == nil
}

// Forwards a pss message to the peer(s) based on recipient address according to the algorithm
// described below. The recipient address can be of any length, and the byte slice will be matched
// to the MSB slice of the peer address of the equivalent length.
//
// If the recipient address (or partial address) is within the neighbourhood depth of the forwarding
// node, then it will be forwarded to all the nearest neighbours of the forwarding node. In case of
// partial address, it should be forwarded to all the peers matching the partial address, if there
// are any; otherwise only to one peer, closest to the recipient address. In any case, if the message
//// forwarding fails, the node should try to forward it to the next best peer, until the message is
//// successfully forwarded to at least one peer.
func (p *Pss) forward(msg *message.Message) error {
	metrics.GetOrRegisterCounter("pss.forward", nil).Inc(1)
	sent := 0 // number of successful sends
	to := make([]byte, addressLength)
	copy(to[:len(msg.To)], msg.To)
	neighbourhoodDepth := p.NeighbourhoodDepth()

	// luminosity is the opposite of darkness. the more bytes are removed from the address, the higher is darkness,
	// but the luminosity is less. here luminosity equals the number of bits given in the destination address.
	luminosityRadius := len(msg.To) * 8

	// proximity order function matching up to neighbourhoodDepth bits (po <= neighbourhoodDepth)
	pof := pot.DefaultPof(neighbourhoodDepth)

	// soft threshold for msg broadcast
	broadcastThreshold, _ := pof(to, p.BaseAddr(), 0)
	if broadcastThreshold > luminosityRadius {
		broadcastThreshold = luminosityRadius
	}

	var onlySendOnce bool // indicates if the message should only be sent to one peer with closest address

	// if measured from the recipient address as opposed to the base address (see Kademlia.EachConn
	// call below), then peers that fall in the same proximity bin as recipient address will appear
	// [at least] one bit closer, but only if these additional bits are given in the recipient address.
	if broadcastThreshold < luminosityRadius && broadcastThreshold < neighbourhoodDepth {
		broadcastThreshold++
		onlySendOnce = true
	}

	p.EachConn(to, addressLength*8, func(sp *network.Peer, po int) bool {
		if po < broadcastThreshold && sent > 0 {
			return false // stop iterating
		}
		if sendFunc(p, sp, msg) {
			sent++
			if onlySendOnce {
				return false
			}
			if po == addressLength*8 {
				// stop iterating if successfully sent to the exact recipient (perfect match of full address)
				return false
			}
		}
		return true
	})

	// cache the message
	p.addFwdCache(msg)

	if sent == 0 {
		return errors.New("unable to forward to any peers")
	} else {
		return nil
	}
}
func label(b []byte) string {
	if len(b) == 0 {
		return "-"
	}
	l := 2
	if len(b) == 1 {
		l = 1
	}
	return fmt.Sprintf("%04x", b[:l])
}

// add a message to the cache
func (p *Pss) addFwdCache(msg *message.Message) error {
	defer metrics.GetOrRegisterResettingTimer("pss.addfwdcache", nil).UpdateSince(time.Now())
	return p.forwardCache.Add(msg.Digest())
}

// check if message is in the cache
func (p *Pss) checkFwdCache(msg *message.Message) bool {
	hit := p.forwardCache.Has(msg.Digest())
	if hit {
		metrics.GetOrRegisterCounter("pss.checkfwdcache.hit", nil).Inc(1)
	} else {
		metrics.GetOrRegisterCounter("pss.checkfwdcache.miss", nil).Inc(1)
	}
	return hit
}

func validateAddress(addr PssAddress) error {
	if len(addr) > addressLength {
		return errors.New("address too long")
	}
	return nil
}
