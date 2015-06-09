package bzz

import (
	// "fmt"
	"time"

	"github.com/ethereum/go-ethereum/common/kademlia"
)

type peer struct {
	*bzzProtocol
}

// peer not necessary here
// bzz protocol could implement kademlia.Node interface with
// Addr(), LastActive() and Drop()

// Hive is the logistic manager of the swarm
// it uses a generic kademlia nodetable to find best peer list
// for any target
// this is used by the netstore to search for content in the swarm
// the bzz protocol peersMsgData exchange is relayed to Kademlia
// for db storage and filtering
// connections and disconnections are reported and relayed
// to keep the nodetable uptodate

type hive struct {
	addr kademlia.Address
	kad  *kademlia.Kademlia
	path string
	ping chan bool
	more chan bool
}

func newHive() (*hive, error) {
	kad := kademlia.New()
	kad.BucketSize = 3
	kad.MaxProx = 10
	kad.ProxBinSize = 8
	return &hive{
		kad: kad,
	}, nil
}

func (self *hive) start(baseAddr *peerAddr, hivepath string, connectPeer func(string) error) (err error) {
	self.ping = make(chan bool)
	self.path = hivepath

	self.addr = kademlia.Address(baseAddr.hash)
	self.kad.Start(self.addr)
	err = self.kad.Load(self.path)
	if err != nil {
		dpaLogger.Warnf("Warning: error reading kademlia node db (skipping): %v", err)
		err = nil
	}
	/* this loop is doing the actual table maintenance
	including bootstrapping and maintaining a healthy table
	Note: At the moment, this does not have any timer/timeout . That means if your
	peers do not reply to launch the game into movement , it will stay stuck
	add or remove a peer to wake up
	*/
	self.more = make(chan bool)
	go func() {
		clock := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-clock.C:
				select {
				case self.ping <- true:
				default:
				}
			case _, more := <-self.more:
				if !more {
					return
				}
			}
		}
	}()
	go func() {
		// whenever pinged ask kademlia about most preferred peer
		for _ = range self.ping {
			node, full := self.kad.GetNodeRecord()
			if node != nil {
				// if Url known, connect to peer
				if len(node.Url) > 0 {
					dpaLogger.Debugf("hive: attempt to connect kaddb node %v", node)
					connectPeer(node.Url)
				} else if !full {
					// a random peer is taken from the table
					peers := self.kad.GetNodes(kademlia.RandomAddress(), 1)
					if len(peers) > 0 {
						// a random address at prox bin 0 is sent for lookup
						randAddr := kademlia.RandomAddressAt(self.addr, 0)
						req := &retrieveRequestMsgData{
							Key: Key(randAddr[:]),
						}
						dpaLogger.Debugf("hive: look up random address with prox order 0 from peer %v", peers[0])
						peers[0].(peer).retrieve(req)
					}
					self.more <- true
				}
			}
			dpaLogger.Debugf("%v", self.kad)
		}
	}()
	return
}

func (self *hive) stop() error {
	// closing ping channel quits the updateloop
	close(self.ping)
	close(self.more)
	return self.kad.Stop(self.path)
}

func (self *hive) addPeer(p peer) {
	dpaLogger.Debugf("hive: add peer %v", p)
	self.kad.AddNode(p)
	// self lookup
	// dpaLogger.Debugf("hive: self lookup - \n%v\n%v\n%064x\n", self.addr, common.Hash(self.addr).Hex(), Key(common.Hash(self.addr).Bytes()[:]))
	// self lookup is encoded as nil/zero key - easier to differentiate so that
	// we do not record as request or forward it, just reply with peers
	p.retrieve(&retrieveRequestMsgData{})
	dpaLogger.Debugf("hive: self lookup sent to %v", p)
	self.ping <- true
}

func (self *hive) removePeer(p peer) {
	dpaLogger.Debugf("hive: remove peer %v", p)
	self.kad.RemoveNode(p)
	self.ping <- false
}

// Retrieve a list of live peers that are closer to target than us
func (self *hive) getPeers(target Key, max int) (peers []peer) {
	var addr kademlia.Address
	copy(addr[:], target[:])
	for _, node := range self.kad.GetNodes(addr, max) {
		peers = append(peers, node.(peer))
	}
	return
}

func newNodeRecord(addr *peerAddr) *kademlia.NodeRecord {
	return &kademlia.NodeRecord{
		Address: kademlia.Address(addr.hash),
		Active:  0,
		Url:     addr.enode,
	}
}

// called by the protocol upon receiving peerset (for target address)
// peersMsgData is converted to a slice of NodeRecords for Kademlia
// this is to store all thats needed
func (self *hive) addPeerEntries(req *peersMsgData) {
	var nrs []*kademlia.NodeRecord
	for _, p := range req.Peers {
		nrs = append(nrs, newNodeRecord(p))
	}
	self.kad.AddNodeRecords(nrs)
}
