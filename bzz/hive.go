package bzz

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

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

type hive struct {
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

func newHive(addr common.Hash, id discover.NodeID, params *HiveParams) (*hive, error) {
	kad := kademlia.New(kademlia.Address(addr), params.KadParams)
	return &hive{
		callInterval: params.CallInterval,
		id:           id,
		kad:          kad,
		addr:         kad.Addr(),
		path:         params.KadDbPath,
	}, nil
}

func (self *hive) start(listenAddr func() string, connectPeer func(string) error) (err error) {
	self.ping = make(chan bool)
	self.more = make(chan bool)
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
						Key: Key(randAddr[:]),
					}
					glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: call any bee in area %x messenger bee %v", randAddr[:4], peers[0])
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
			glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: queen's address: %x, population: %d (%d)\n%v", self.addr[:4], self.kad.Count(), self.kad.DBCount(), self.kad)
		}
	}()
	return
}

func (self *hive) pinger() {
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

func (self *hive) stop() error {
	// closing ping channel quits the updateloop
	close(self.ping)
	if self.more != nil {
		close(self.more)
		self.more = nil
	}
	return self.kad.Save(self.path, saveSync)
}

// called at the end of a successful protocol handshake
func (self *hive) addPeer(p *peer) {
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
func (self *hive) removePeer(p *peer) {
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
func (self *hive) getPeers(target Key, max int) (peers []*peer) {
	var addr kademlia.Address
	copy(addr[:], target[:])
	for _, node := range self.kad.FindClosest(addr, max) {
		peers = append(peers, node.(*peer))
	}
	return
}

// disconnects all the peers
func (self *hive) dropAll() {
	glog.V(logger.Detail).Infof("[BZZ] KΛÐΞMLIΛ hive: dropping all bees")
	for _, node := range self.kad.FindClosest(kademlia.Address{}, 0) {
		node.Drop()
	}
}

//
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
func (self *hive) addPeerEntries(req *peersMsgData) {
	var nrs []*kademlia.NodeRecord
	for _, p := range req.Peers {
		nrs = append(nrs, newNodeRecord(p))
	}
	self.kad.Add(nrs)
}

// peer wraps the protocol instance so that it implements kademlia.Node interface
type peer struct {
	*bzzProtocol
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
		glog.V(logger.Debug).Infof("meta for node record %v: %v -> %v", record, record.Meta, state)
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
