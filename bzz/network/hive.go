package network

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/bzz/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/kademlia"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

// Hive is the logistic manager of the swarm
// it uses a generic kademlia nodetable to find best peer list
// for any target
// this is used by the netstore to search for content in the swarm
// the bzz protocol peersMsgData exchange is relayed to Kademlia
// for db storage and filtering
// connections and disconnections are reported and relayed
// to keep the nodetable uptodate

type Hive struct {
	listenAddr   func() string
	callInterval uint
	id           discover.NodeID
	addr         kademlia.Address
	kad          *kademlia.Kademlia
	path         string
	ping         chan bool
	more         chan bool
}

const (
	callInterval = 1000000000
	bucketSize   = 3
	maxProx      = 10
	proxBinSize  = 8
)

type HiveParams struct {
	CallInterval uint
	KadDbPath    string
	*kademlia.KadParams
}

func NewHiveParams(path string) *HiveParams {
	kad := kademlia.NewKadParams()
	kad.BucketSize = bucketSize
	kad.MaxProx = maxProx
	kad.ProxBinSize = proxBinSize

	return &HiveParams{
		CallInterval: callInterval,
		KadDbPath:    filepath.Join(path, "bzz-peers.json"),
		KadParams:    kad,
	}
}

func NewHive(addr common.Hash, params *HiveParams) *Hive {
	kad := kademlia.New(kademlia.Address(addr), params.KadParams)
	return &Hive{
		callInterval: params.CallInterval,
		kad:          kad,
		addr:         kad.Addr(),
		path:         params.KadDbPath,
	}
}

// public accessor to the hive base address
func (self *Hive) Addr() kademlia.Address {
	return self.addr
}

// Start receives network info only at startup
// listedAddr is a function to retrieve listening address to advertise to peers
// connectPeer is a function to connect to a peer based on its NodeID or enode URL
// there are called on the p2p.Server which runs on the node
func (self *Hive) Start(id discover.NodeID, listenAddr func() string, connectPeer func(string) error) (err error) {
	self.ping = make(chan bool)
	self.more = make(chan bool)
	self.id = id
	self.listenAddr = listenAddr
	err = self.kad.Load(self.path, nil)
	if err != nil {
		glog.V(logger.Warn).Infof("[BZZ] KΛÐΞMLIΛ Warning: error reading kaddb '%s' (skipping): %v", self.path, err)
		err = nil
	}
	// this loop is doing bootstrapping and maintains a healthy table
	go self.pinger()
	go func() {
		// whenever pinged ask kademlia about most preferred peer
		for _ = range self.ping {
			node, proxLimit := self.kad.FindBest()
			if node != nil && len(node.Url) > 0 {
				glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: call for bee %v", node.Url)
				// enode or any lower level connection address is unnecessary in future
				// discovery table is used to look it up.
				connectPeer(node.Url)
			} else if proxLimit > -1 {
				// a random peer is taken from the table
				peers := self.kad.FindClosest(kademlia.RandomAddressAt(self.addr, rand.Intn(self.kad.MaxProx)), 1)
				if len(peers) > 0 {
					// a random address at prox bin 0 is sent for lookup
					randAddr := kademlia.RandomAddressAt(self.addr, proxLimit)
					req := &retrieveRequestMsgData{
						Key: storage.Key(randAddr[:]),
					}
					glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: call any bee in area %v messenger bee %v", randAddr, peers[0])
					peers[0].(*peer).retrieve(req)
				}
				if self.more == nil {
					glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: buzz buzz need more bees")
					self.more = make(chan bool)
					go self.pinger()
				}
				self.more <- true
				glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: buzz kept alive")
			} else {
				if self.more != nil {
					close(self.more)
					self.more = nil
				}
			}
			glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: queen's address: %v, population: %d (%d)\n%v", self.addr, self.kad.Count(), self.kad.DBCount(), self.kad)
		}
	}()
	return
}

// pinger is awake until Kademlia Table is saturated
// it restarts if the table becomes non-full again due to disconnections
func (self *Hive) pinger() {
	clock := time.NewTicker(time.Duration(self.callInterval))
	for {
		select {
		case <-clock.C:
			if self.kad.DBCount() > 0 {
				select {
				case self.ping <- true:
				default:
				}
			}
		case _, more := <-self.more:
			if !more {
				return
			}
		}
	}
}

func (self *Hive) Stop() error {
	// closing ping channel quits the updateloop
	close(self.ping)
	self.ping = nil
	if self.more != nil {
		close(self.more)
		self.more = nil
	}
	return self.kad.Save(self.path, saveSync)
}

// called at the end of a successful protocol handshake
func (self *Hive) addPeer(p *peer) {
	glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: hi new bee %v", p)
	self.kad.On(p, loadSync)
	// self lookup (can be encoded as nil/zero key since peers addr known) + no id ()
	// the most common way of saying hi in bzz is initiation of gossip
	// let me know about anyone new from my hood , here is the storageradius
	// to send the 6 byte self lookup
	// we do not record as request or forward it, just reply with peers
	p.retrieve(&retrieveRequestMsgData{})
	glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: 'whatsup wheresdaparty' sent to %v", p)
	if self.ping != nil {
		self.ping <- true
	}
}

// called after peer disconnected
func (self *Hive) removePeer(p *peer) {
	glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: bee %v gone offline", p)
	self.kad.Off(p, saveSync)
	if self.ping != nil {
		self.ping <- false
	}
	if self.kad.Count() == 0 {
		glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: empty, all bees gone", p)
	}
}

// Retrieve a list of live peers that are closer to target than us
func (self *Hive) getPeers(target storage.Key, max int) (peers []*peer) {
	var addr kademlia.Address
	copy(addr[:], target[:])
	for _, node := range self.kad.FindClosest(addr, max) {
		peers = append(peers, node.(*peer))
	}
	return
}

// disconnects all the peers
func (self *Hive) DropAll() {
	glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: dropping all bees")
	for _, node := range self.kad.FindClosest(kademlia.Address{}, 0) {
		node.Drop()
	}
}

// contructor for kademlia.NodeRecord based on peer address alone
// TODO: should go away and only addr passed to kademlia
func newNodeRecord(addr *peerAddr) *kademlia.NodeRecord {
	now := kademlia.Time(time.Now())
	return &kademlia.NodeRecord{
		Addr:  addr.Addr,
		Url:   addr.String(),
		Seen:  now,
		After: now,
	}
}

// called by the protocol when receiving peerset (for target address)
// peersMsgData is converted to a slice of NodeRecords for Kademlia
// this is to store all thats needed
func (self *Hive) HandlePeersMsg(req *peersMsgData, from *peer) {
	var nrs []*kademlia.NodeRecord
	for _, p := range req.Peers {
		nrs = append(nrs, newNodeRecord(p))
	}
	self.kad.Add(nrs)
}

// peer wraps the protocol instance to represent a connected peer
// it implements kademlia.Node interface
type peer struct {
	*bzz // protocol instance running on peer connection
}

// protocol instance implements kademlia.Node interface (embedded peer)
func (self *peer) Addr() kademlia.Address {
	return self.remoteAddr.Addr
}

func (self *peer) Url() string {
	return self.remoteAddr.String()
}

// TODO take into account traffic
func (self *peer) LastActive() time.Time {
	return time.Now()
}

// reads the serialised form of sync state persisted as the 'Meta' attribute
// and sets the decoded syncState on the online node
func loadSync(record *kademlia.NodeRecord, node kademlia.Node) error {
	if p, ok := node.(*peer); ok {
		if record.Meta == nil {
			glog.V(logger.Detail).Infof("no meta for node record %v", record)
			return nil
		}
		state, err := decodeSync(record.Meta)
		glog.V(logger.Detail).Infof("meta for node record %v: %s -> %v", record, string(*(record.Meta)), state)
		p.syncState = state
		return err
	}
	return fmt.Errorf("invalid type")
}

// callback when saving a sync state
func saveSync(record *kademlia.NodeRecord, node kademlia.Node) {
	if p, ok := node.(*peer); ok {
		meta, err := encodeSync(p.syncState)
		if err != nil {
			glog.V(logger.Warn).Infof("error saving sync state for %v: %v", node, err)
			return
		}
		record.Meta = meta
	}
}

// the immediate response to a retrieve request,
// sends relevant peer data given by the kademlia hive to the requester
// TODO: remember peers sent for duration of the session, only new peers sent
func (self *Hive) peers(req *retrieveRequestMsgData) {
	// FIXME: should check req.MaxPeers but then should not default to zero or make sure we set it when sending retrieveRequests
	// we might need chunk.req to cache relevant peers response,
	// hive change would expire it
	if req != nil && req.MaxPeers >= 0 {
		var addrs []*peerAddr
		if req.timeout == nil || time.Now().Before(*(req.timeout)) {
			key := req.Key
			// self lookup from remote peer
			if storage.IsZeroKey(key) {
				addr := req.from.Addr()
				key = storage.Key(addr[:])
				req.Key = nil
			}
			// get peer addresses from hive
			for _, peer := range self.getPeers(key, int(req.MaxPeers)) {
				addrs = append(addrs, peer.remoteAddr)
			}
			glog.V(logger.Detail).Infof("[BZZ] Hive sending %d peer addresses to %v. req.Id: %v, req.Key: %x", len(addrs), req.from, req.Id, req.Key.Log())

			peersData := &peersMsgData{
				Peers: addrs,
				Key:   req.Key,
				Id:    req.Id,
			}
			peersData.setTimeout(req.timeout)
			req.from.peers(peersData)
		}
	}
}
