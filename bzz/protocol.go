package bzz

/*
BZZ implements the bzz wire protocol of swarm
the protocol instance is launched on each peer by the network layer if the
BZZ protocol handler is registered on the p2p server.

The protocol takes care of actually communicating the bzz protocol
* encoding and decoding requests for storage and retrieval
* handling the protocol handshake
* dispaching to netstore for handling the DHT logic
* registering peers in the KΛÐΞMLIΛ table via the hive logistic manager
* handling sync protocol messages via the syncer
* talks the SWAP payent protocol (swap accounting is done within netStore)
*/

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/chequebook"
	"github.com/ethereum/go-ethereum/common/kademlia"
	"github.com/ethereum/go-ethereum/common/swap"
	"github.com/ethereum/go-ethereum/errs"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

const (
	Version            = 0
	ProtocolLength     = uint64(8)
	ProtocolMaxMsgSize = 10 * 1024 * 1024
	NetworkId          = 322
)

// bzz protocol message codes
const (
	statusMsg          = iota // 0x01
	storeRequestMsg           // 0x02
	retrieveRequestMsg        // 0x03
	peersMsg                  // 0x04
	syncRequestMsg            // 0x05
	deliveryRequestMsg        // 0x06
	unsyncedKeysMsg           // 0x07
	paymentMsg                // 0x08
)

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrVersionMismatch
	ErrNetworkIdMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSwap
	ErrSync
)

var errorToString = map[int]string{
	ErrMsgTooLarge:       "Message too long",
	ErrDecode:            "Invalid message",
	ErrInvalidMsgCode:    "Invalid message code",
	ErrVersionMismatch:   "Protocol version mismatch",
	ErrNetworkIdMismatch: "NetworkId mismatch",
	ErrNoStatusMsg:       "No status message",
	ErrExtraStatusMsg:    "Extra status message",
	ErrSwap:              "SWAP error",
	ErrSync:              "Sync error",
}

// bzzProtocol represents the swarm wire protocol
// an instance is running on each peer
type bzzProtocol struct {
	selfID     discover.NodeID
	netStore   *netStore
	peer       *p2p.Peer
	remoteAddr *peerAddr
	key        Key
	rw         p2p.MsgReadWriter
	errors     *errs.Errors

	swap        *swap.Swap
	swapParams  *swapParams
	swapEnabled bool
	syncer      *syncer
	syncParams  *SyncParams
	syncState   *syncState
	syncEnabled bool
}

/*
 Handshake

 [0x01, Version: B_8, ID: B, Addr: [NodeID: B_64, IP: B_4 or B_6, Port: P], NetworkID; B_8, Caps: [[cap1: B_3, capVersion1: P], [cap2: B_3, capVersion2: P], ...]]

* Version: 8 byte integer version of the protocol
* ID: arbitrary byte sequence client identifier human readable
* Addr: the address advertised by the node, format similar to DEVp2p wire protocol
* Swap: info for the swarm accounting protocol
* NetworkID: 8 byte integer network identifier
* Caps: swarm-specific capabilities, format identical to devp2p
* SyncState: syncronisation state (db iterator key and address space etc) persisted about the peer

*/
type statusMsgData struct {
	Version   uint64
	ID        string
	Addr      *peerAddr
	Swap      *swapProfile
	NetworkId uint64
}

func (self *statusMsgData) String() string {
	return fmt.Sprintf("Status: Version: %v, ID: %v, Addr: %v, Swap: %v, NetworkId: %v", self.Version, self.ID, self.Addr, self.Swap, self.NetworkId)
}

/*
 store requests are forwarded to the peers in their kademlia proximity bin
 if they are distant
 if they are within our storage radius or have any incentive to store it
 then attach your nodeID to the metadata
 if the storage request is sufficiently close (within our proxLimit, i. e., the
 last row of the routing table)
*/
type storeRequestMsgData struct {
	Key   Key    // hash of datasize | data
	SData []byte // the actual chunk Data
	// optional
	Id             uint64     // request ID. if delivery, the ID is retrieve request ID
	requestTimeout *time.Time // expiry for forwarding - [not serialised][not currently used]
	storageTimeout *time.Time // expiry of content - [not serialised][not currently used]
	peer           *peer      // [not serialised] protocol registers the requester
}

func (self storeRequestMsgData) String() string {
	var from string
	if self.peer == nil {
		from = "self"
	} else {
		from = self.peer.Addr().String()
	}
	return fmt.Sprintf("From: %v, Key: %v; ID: %v, requestTimeout: %v, storageTimeout: %v, SData %x", from, self.Key, self.Id, self.requestTimeout, self.storageTimeout, self.SData[:10])
}

/*
Retrieve request

Timeout in milliseconds. Note that zero timeout retrieval requests do not request forwarding, but prompt for a peers message response. therefore they serve also
as messages to retrieve peers.

MaxSize specifies the maximum size that the peer will accept. This is useful in
particular if we allow storage and delivery of multichunk payload representing
the entire or partial subtree unfolding from the requested root key.
So when only interested in limited part of a stream (infinite trees) or only
testing chunk availability etc etc, we can indicate it by limiting the size here.

Request ID can be newly generated or kept from the request originator.
If request ID Is missing or zero, the request is handled as a lookup only
prompting a peers response but not launching a search. Lookup requests are meant
to be used to bootstrap kademlia tables.

In the special case that the key is the zero value as well, the remote peer's
address is assumed (the message is to be handled as a self lookup request).
The response is a PeersMsg with the peers in the kademlia proximity bin
corresponding to the address.
*/

type retrieveRequestMsgData struct {
	Key Key
	// optional
	Id       uint64     // request id, request is a lookup if missing or zero
	MaxSize  uint64     // maximum size of delivery accepted
	MaxPeers uint64     // maximum number of peers returned
	Timeout  uint64     // the longest time we are expecting a response
	timeout  *time.Time // [not serialised]
	peer     *peer      // [not serialised] protocol registers the requester
}

func (self retrieveRequestMsgData) String() string {
	var from string
	if self.peer == nil {
		from = "ourselves"
	} else {
		from = self.peer.Addr().String()
	}
	var target []byte
	if len(self.Key) > 3 {
		target = self.Key[:4]
	}
	return fmt.Sprintf("From: %v, Key: %x; ID: %v, MaxSize: %v, MaxPeers: %d", from, target, self.Id, self.MaxSize, self.MaxPeers)
}

// lookups are encoded by missing request ID
func (self retrieveRequestMsgData) isLookup() bool {
	return self.Id == 0
}

func isZeroKey(key Key) bool {
	return len(key) == 0 || bytes.Equal(key, zeroKey)
}

// sets timeout fields
func (self retrieveRequestMsgData) setTimeout(t *time.Time) {
	self.timeout = t
	if t != nil {
		self.Timeout = uint64(t.UnixNano())
	} else {
		self.Timeout = 0
	}
}

func (self retrieveRequestMsgData) getTimeout() (t *time.Time) {
	if self.Timeout > 0 && self.timeout == nil {
		timeout := time.Unix(int64(self.Timeout), 0)
		t = &timeout
		self.timeout = t
	}
	return
}

// peerAddr is sent in StatusMsg as part of the handshake
type peerAddr struct {
	IP   net.IP
	Port uint16
	ID   []byte // the 64 byte NodeID (ECDSA Public Key)
	Addr kademlia.Address
}

// peerAddr pretty prints as enode
func (self peerAddr) String() string {
	return fmt.Sprintf("enode://%x@%v:%d", self.ID, self.IP, self.Port)
}

/*
peers Msg is one response to retrieval; it is always encouraged after a retrieval
request to respond with a list of peers in the same kademlia proximity bin.
The encoding of a peer is identical to that in the devp2p base protocol peers
messages: [IP, Port, NodeID]
note that a node's DPA address is not the NodeID but the hash of the NodeID.

Timeout serves to indicate whether the responder is forwarding the query within
the timeout or not.

NodeID serves as the owner of payment contracts and signer of proofs of transfer.

The Key is the target (if response to a retrieval request) or missing (zero value)
peers address (hash of NodeID) if retrieval request was a self lookup.

Peers message is requested by retrieval requests with a missing or zero value request ID
*/
type peersMsgData struct {
	Peers   []*peerAddr //
	Timeout uint64      //
	timeout *time.Time  // indicate whether responder is expected to deliver content
	Key     Key         // present if a response to a retrieval request
	Id      uint64      // present if a response to a retrieval request

	peer *peer
}

// peers msg pretty printer
func (self peersMsgData) String() string {
	var from string
	if self.peer == nil {
		from = "ourselves"
	} else {
		from = self.peer.Addr().String()
	}
	var target []byte
	if len(self.Key) > 3 {
		target = self.Key[:4]
	}
	return fmt.Sprintf("From: %v, Key: %x; ID: %v, Peers: %v", from, target, self.Id, self.Peers)
}

func (self peersMsgData) setTimeout(t *time.Time) {
	self.timeout = t
	if t != nil {
		self.Timeout = uint64(t.UnixNano())
	} else {
		self.Timeout = 0
	}
}

func (self peersMsgData) getTimeout() (t *time.Time) {
	if self.Timeout > 0 && self.timeout == nil {
		timeout := time.Unix(int64(self.Timeout), 0)
		t = &timeout
		self.timeout = t
	}
	return
}

/*
syncRequest

is sent after the handshake to initiate syncing
the syncState of the remote node is persisted in kaddb and set on the
peer/protocol instance when the node is registered by hive as online{
*/

type syncRequestMsgData struct {
	SyncState *syncState `rlp:"nil"`
}

func (self *syncRequestMsgData) String() string {
	return fmt.Sprintf("%v", self.SyncState)
}

/*
deliveryRequest

is sent once a batch of sync keys is filtered. The ones not found are
sent as a list of syncReuest (hash, priority) in the Deliver field.
When the source receives the sync request it continues to iterate
and fetch at most N items as yet unsynced.
At the same time responds with deliveries of the items.
*/
type deliveryRequestMsgData struct {
	Deliver []*syncRequest
}

func (self *deliveryRequestMsgData) String() string {
	return fmt.Sprintf("sync request for new chunks\ndelivery request for %v chunks", len(self.Deliver))
}

/*
unsyncedKeys

is sent first after the handshake if SyncState iterator brings up hundreds, thousands?
and subsequently sent as a response to deliveryRequestMsgData.

Syncing is the iterative process of exchanging unsyncedKeys and deliveryRequestMsgs
both ways.

State contains the sync state sent by the source. When the source receives the
sync state it continues to iterate and fetch at most N items as yet unsynced.
At the same time responds with deliveries of the items.
*/
type unsyncedKeysMsgData struct {
	Unsynced []*syncRequest
	State    syncState
}

func (self *unsyncedKeysMsgData) String() string {
	return fmt.Sprintf("sync: keys of %d new chunks (upto %v) => synced: %v", len(self.Unsynced), self.State)
}

/*
payment

is sent when the swap balance is tilted in favour of the remote peer
and in absolute units exceeds the PayAt parameter in the remote peer's profile
*/

type paymentMsgData struct {
	Units   uint               // units actually paid for (checked against amount by swap)
	Promise *chequebook.Cheque // payment with cheque
}

func (self *paymentMsgData) String() string {
	return fmt.Sprintf("payment for %d units: %v", self.Units, self.Promise)
}

/*
main entrypoint, wrappers starting a server that will run the bzz protocol
use this constructor to attach the protocol ("class") to server caps
the Dev p2p layer then runs the protocol instance on each peer
*/
func BzzProtocol(netstore *netStore, sp *swapParams, sy *SyncParams) (p2p.Protocol, error) {

	return p2p.Protocol{
		Name:    "bzz",
		Version: Version,
		Length:  ProtocolLength,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			return runBzzProtocol(netstore, sp, sy, p, rw)
		},
	}, nil
}

// the main loop that handles incoming messages
// note RemovePeer in the post-disconnect hook
func runBzzProtocol(netstore *netStore, sp *swapParams, sy *SyncParams, p *p2p.Peer, rw p2p.MsgReadWriter) (err error) {

	self := &bzzProtocol{
		netStore: netstore,
		rw:       rw,
		peer:     p,
		errors: &errs.Errors{
			Package: "BZZ",
			Errors:  errorToString,
		},
		swapParams:  sp,
		syncParams:  sy,
		swapEnabled: true,
		syncEnabled: true,
	}

	err = self.handleStatus()
	if err != nil {
		return err
	}
	defer func() {
		// if the handler loop exits, the peer is disconnecting
		// deregister the peer in the hive
		self.netStore.hive.removePeer(&peer{bzzProtocol: self})
		if self.syncer != nil {
			self.syncer.stop() // quits request db and delivery loops, save requests
		}
		if self.swap != nil {
			self.swap.Stop() // quits chequebox autocash etc
		}
	}()

	for {
		err = self.handle()
		if err != nil {
			return
		}
	}
	return
}

// may need to implement protocol drop only? don't want to kick off the peer
// if they are useful for other protocols
func (self *bzzProtocol) Drop() {
	self.peer.Disconnect(p2p.DiscSubprotocolError)
}

// main loop that handles all incoming messages
func (self *bzzProtocol) handle() error {
	msg, err := self.rw.ReadMsg()
	glog.V(logger.Debug).Infof("[BZZ] Incoming MSG: %v", msg)
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return self.protoError(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	// make sure that the payload has been fully consumed
	defer msg.Discard()

	switch msg.Code {

	case statusMsg:
		// no extra status message allowed. The one needed already handled by
		// handleStatus
		glog.V(logger.Debug).Infof("[BZZ] Status message: %v", msg)
		return self.protoError(ErrExtraStatusMsg, "")

	case storeRequestMsg:
		// store requests are dispatched to netStore
		var req storeRequestMsgData
		if err := msg.Decode(&req); err != nil {
			return self.protoError(ErrDecode, "msg %v: %v", msg, err)
		}
		req.peer = &peer{bzzProtocol: self}
		glog.V(logger.Debug).Infof("[BZZ] incoming store request: %s", req.String())
		// swap accounting is done within netStore
		self.netStore.addStoreRequest(&req)

	case retrieveRequestMsg:
		// retrieve Requests are dispatched to netStore
		var req retrieveRequestMsgData
		if err := msg.Decode(&req); err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}
		if req.Key == nil {
			return self.protoError(ErrDecode, "protocol handler: req.Key == nil || req.Timeout == nil")
		}
		req.peer = &peer{bzzProtocol: self}
		glog.V(logger.Debug).Infof("[BZZ] incoming retrieve request: %v", req)
		// swap accounting is done within netStore
		self.netStore.addRetrieveRequest(&req)

	case peersMsg:
		// response to lookups and immediate response to retrieve requests
		// dispatches new peer data to the hive that adds them to KADDB
		var req peersMsgData
		if err := msg.Decode(&req); err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}
		req.peer = &peer{bzzProtocol: self}
		glog.V(logger.Debug).Infof("[BZZ] incoming peer addresses: %v", req)
		self.netStore.hive.addPeerEntries(&req)

	case syncRequestMsg:
		var req syncRequestMsgData
		if err := msg.Decode(&req); err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}
		glog.V(logger.Debug).Infof("[BZZ] sync request received: %v", req)
		self.sync(req.SyncState)

	case unsyncedKeysMsg:
		// coming from parent node offering
		var req unsyncedKeysMsgData
		if err := msg.Decode(&req); err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}
		err := self.syncer.handleUnsyncedKeysMsg(req.Unsynced)
		if err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}
		// set peers state to persist
		self.syncState = &req.State

	case deliveryRequestMsg:
		// response to syncKeysMsg hashes filtered not existing in db
		// also relays the last synced state to the source
		var req deliveryRequestMsgData
		if err := msg.Decode(&req); err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}
		err := self.syncer.handleDeliveryRequestMsg(req.Deliver)
		if err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}

	case paymentMsg:
		// swap protocol message for payment, Units paid for, Cheque paid with
		var req paymentMsgData
		if err := msg.Decode(&req); err != nil {
			return self.protoError(ErrDecode, "->msg %v: %v", msg, err)
		}
		glog.V(logger.Debug).Infof("[BZZ] incoming payment: %s", req.String())
		self.swap.Receive(int(req.Units), req.Promise)

	default:
		// no other message is allowed
		return self.protoError(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (self *bzzProtocol) handleStatus() (err error) {

	handshake := &statusMsgData{
		Version:   uint64(Version),
		ID:        "honey",
		Addr:      self.selfAddr(),
		NetworkId: uint64(NetworkId),
		Swap: &swapProfile{
			Profile:    self.swapParams.Profile,
			payProfile: self.swapParams.payProfile,
		},
	}

	err = p2p.Send(self.rw, statusMsg, handshake)
	if err != nil {
		self.protoError(ErrNoStatusMsg, err.Error())
	}

	// read and handle remote status
	var msg p2p.Msg
	msg, err = self.rw.ReadMsg()
	if err != nil {
		return err
	}

	if msg.Code != statusMsg {
		self.protoError(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, statusMsg)
	}

	if msg.Size > ProtocolMaxMsgSize {
		return self.protoError(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}

	var status statusMsgData
	if err := msg.Decode(&status); err != nil {
		return self.protoError(ErrDecode, "msg %v: %v", msg, err)
	}

	if status.NetworkId != NetworkId {
		return self.protoError(ErrNetworkIdMismatch, "%d (!= %d)", status.NetworkId, NetworkId)
	}

	if Version != status.Version {
		return self.protoError(ErrVersionMismatch, "%d (!= %d)", status.Version, Version)
	}

	self.remoteAddr = self.peerAddr(status.Addr)
	glog.V(logger.Detail).Infof("[BZZ] self: advertised IP: %v, peer advertised: %v, local address: %v\npeer: advertised IP: %v, remote address: %v\n", self.selfAddr(), self.remoteAddr, self.peer.LocalAddr(), status.Addr.IP, self.peer.RemoteAddr())

	if self.swapEnabled {
		// set remote profile for accounting
		self.swap, err = newSwap(self.swapParams, status.Swap, self)
		if err != nil {
			return self.protoError(ErrSwap, "%v", err)
		}
	}

	glog.V(logger.Info).Infof("[BZZ] Peer %08x is [bzz] capable (%d/%d)\n", self.remoteAddr.Addr[:4], status.Version, status.NetworkId)
	self.netStore.hive.addPeer(&peer{bzzProtocol: self})

	// hive sets syncstate so sync should start after node added
	if self.syncEnabled {
		self.syncRequest()
	}
	return nil
}

func (self *bzzProtocol) sync(state *syncState) error {
	// syncer setup
	if self.syncer != nil {
		return self.protoError(ErrSync, "sync request can only be sent once")
	}
	// keyIterator func
	kitf := func(s syncState) keyIterator {
		it, err := self.netStore.localStore.dbStore.newSyncIterator(s.DbSyncState)
		if err != nil {
			return nil
		}
		return keyIterator(it)
	}
	counter := self.netStore.localStore.dbStore.Counter

	remoteaddr := self.remoteAddr.Addr
	start, stop := self.netStore.hive.kad.KeyRange(remoteaddr)
	if state == nil {
		state = newSyncState(start, stop, counter())
		glog.V(logger.Warn).Infof("[BZZ] peer %v provided no sync state, setting up full sync: %v\n", remoteaddr, state)
	}
	var err error
	self.syncer, err = newSyncer(
		self.netStore.requestDb, Key(remoteaddr[:]),
		counter, kitf, self.netStore.localStore.Get,
		self.unsyncedKeys, self.deliveryRequest, self.store,
		self.syncParams, *state,
	)
	if err != nil {
		return self.protoError(ErrSync, "%v", err)
	}
	return nil
}

func (self *bzzProtocol) String() string {
	return self.remoteAddr.String()
}

// repair reported address if IP missing
func (self *bzzProtocol) peerAddr(base *peerAddr) *peerAddr {
	if base.IP.IsUnspecified() {
		host, _, _ := net.SplitHostPort(self.peer.RemoteAddr().String())
		base.IP = net.ParseIP(host)
	}
	return base
}

// returns self advertised node connection info (listening address w enodes)
// IP will get repaired on the other end if missing
// or resolved via ID by discovery at dialout
func (self *bzzProtocol) selfAddr() *peerAddr {
	id := self.netStore.hive.id
	host, port, _ := net.SplitHostPort(self.netStore.hive.listenAddr())
	intport, _ := strconv.Atoi(port)
	addr := &peerAddr{
		Addr: self.netStore.hive.addr,
		ID:   id[:],
		IP:   net.ParseIP(host),
		Port: uint16(intport),
	}
	return addr
}

// outgoing messages
// send retrieveRequestMsg
func (self *bzzProtocol) retrieve(req *retrieveRequestMsgData) error {
	glog.V(logger.Debug).Infof("[BZZ] sending retrieve request: %v", req)
	return self.send(retrieveRequestMsg, req)
}

// send storeRequestMsg
func (self *bzzProtocol) store(req *storeRequestMsgData) error {
	glog.V(logger.Debug).Infof("[BZZ] sending store request: %v", req)
	return self.send(storeRequestMsg, req)
}

func (self *bzzProtocol) syncRequest() error {
	req := &syncRequestMsgData{
		SyncState: self.syncState,
	}
	return self.send(syncRequestMsg, req)
}

// queue storeRequestMsg in request db
func (self *bzzProtocol) deliveryRequest(reqs []*syncRequest) error {
	req := &deliveryRequestMsgData{
		Deliver: reqs,
	}
	return self.send(deliveryRequestMsg, req)
}

// batch of syncRequests to send off
func (self *bzzProtocol) unsyncedKeys(reqs []*syncRequest, state syncState) error {
	req := &unsyncedKeysMsgData{
		Unsynced: reqs,
		State:    state,
	}
	return self.send(unsyncedKeysMsg, req)
}

// send paymentMsg
func (self *bzzProtocol) Pay(units int, promise swap.Promise) {
	req := &paymentMsgData{uint(units), promise.(*chequebook.Cheque)}
	self.payment(req)
}

// send paymentMsg
func (self *bzzProtocol) payment(req *paymentMsgData) error {
	glog.V(logger.Debug).Infof("[BZZ] sending payment: %v", req)
	return self.send(paymentMsg, req)
}

// sends peersMsg
func (self *bzzProtocol) peers(req *peersMsgData) error {
	glog.V(logger.Debug).Infof("[BZZ] sending peers: %v", req)
	return self.send(peersMsg, req)
}

func (self *bzzProtocol) protoError(code int, format string, params ...interface{}) (err *errs.Error) {
	err = self.errors.New(code, format, params...)
	err.Log(glog.V(logger.Info))
	return
}

func (self *bzzProtocol) protoErrorDisconnect(err *errs.Error) {
	err.Log(glog.V(logger.Info))
	if err.Fatal() {
		self.peer.Disconnect(p2p.DiscSubprotocolError)
	}
}

func (self *bzzProtocol) send(msg uint64, data interface{}) error {
	err := p2p.Send(self.rw, msg, data)
	if err != nil {
		self.Drop()
	}
	return err
}
