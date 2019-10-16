package network

import (
	"sort"
	"strconv"
	"sync"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/gopubsub"
)

// KademliaBackend is the required interface of KademliaLoadBalancer.
type KademliaBackend interface {
	SubscribeToPeerChanges() (addedSub *gopubsub.Subscription, removedPeerSub *gopubsub.Subscription)
	BaseAddr() []byte
	EachBinDesc(base []byte, minProximityOrder int, consumer PeerBinConsumer)
	EachBinDescFiltered(base []byte, capKey string, minProximityOrder int, consumer PeerBinConsumer) error
	EachConn(base []byte, o int, f func(*Peer, int) bool)
}

// Creates a new KademliaLoadBalancer from a KademliaBackend
func NewKademliaLoadBalancer(kademlia KademliaBackend, useMostSimilarInit bool) *KademliaLoadBalancer {
	onPeerSub, offPeerSub := kademlia.SubscribeToPeerChanges()
	klb := &KademliaLoadBalancer{
		kademlia:         kademlia,
		resourceUseStats: newResourceLoadBalancer(),
		onPeerSub:        onPeerSub,
		offPeerSub:       offPeerSub,
		quitC:            make(chan struct{}),
	}
	if useMostSimilarInit {
		klb.initCountFunc = klb.nearestNeighbourCount
	} else {
		klb.initCountFunc = klb.leastUsedCountInBin
	}

	go klb.listenNewPeers()
	go klb.listenOffPeers()
	return klb
}

// Consumer functions

// An LBPeer represents a peer with a Use() function to signal that the peer has been used in order
// to account it for LB sorting criteria
type LBPeer struct {
	Peer *Peer
	Use  func()
}

// LBBin represents a Bin of LBPeer's
type LBBin struct {
	LBPeers        []LBPeer
	ProximityOrder int
}

// LBBinConsumer will be provided with a list of LBPeer's in LB criteria ordering (currently in least used ordering).
type LBBinConsumer func(bin LBBin) bool

// KademliaLoadBalancer struct and methods

// KademliaLoadBalancer tries to balance request to the peers in Kademlia returning the peers sorted
// by least recent used whenever several will be returned with the same po to a particular address.
// The user of KademliaLoadBalancer should signal if the returned element (LBPeer) has been used with the
// function lbPeer.Use()
type KademliaLoadBalancer struct {
	kademlia         KademliaBackend        //kademlia to obtain bins of peers
	resourceUseStats *resourceUseStats      //a resourceUseStats to count uses
	onPeerSub        *gopubsub.Subscription //a pubsub channel to be notified of new peers in kademlia
	offPeerSub       *gopubsub.Subscription //a pubsub channel to be notified of removed peers in kademlia
	quitC            chan struct{}

	initCountFunc func(peer *Peer, po int) int //Function to use for initializing a new peer count
}

// Stop unsubscribe from notifiers
func (klb *KademliaLoadBalancer) Stop() {
	klb.onPeerSub.Unsubscribe()
	klb.offPeerSub.Unsubscribe()
	close(klb.quitC)
}

// EachBinNodeAddress calls EachBinDesc with the base address of kademlia (the node address)
func (klb *KademliaLoadBalancer) EachBinNodeAddress(consumeBin LBBinConsumer) {
	klb.EachBinDesc(klb.kademlia.BaseAddr(), consumeBin)
}

// EachBinFiltered returns all bins in descending order from the perspective of base address.
// Only peers with the provided capabilities capKey are considered.
// All peers in that bin will be provided to the LBBinConsumer sorted by least used first.
func (klb *KademliaLoadBalancer) EachBinFiltered(base []byte, capKey string, consumeBin LBBinConsumer) {
	_ = klb.kademlia.EachBinDescFiltered(base, capKey, 0, func(peerBin *PeerBin) bool {
		peers := klb.peerBinToPeerList(peerBin)
		return consumeBin(LBBin{LBPeers: peers, ProximityOrder: peerBin.ProximityOrder})
	})
}

// EachBinDesc returns all bins in descending order from the perspective of base address.
// All peers in that bin will be provided to the LBBinConsumer sorted by least used first.
func (klb *KademliaLoadBalancer) EachBinDesc(base []byte, consumeBin LBBinConsumer) {
	klb.kademlia.EachBinDesc(base, 0, func(peerBin *PeerBin) bool {
		peers := klb.peerBinToPeerList(peerBin)
		return consumeBin(LBBin{LBPeers: peers, ProximityOrder: peerBin.ProximityOrder})
	})
}

func (klb *KademliaLoadBalancer) peerBinToPeerList(bin *PeerBin) []LBPeer {
	resources := make([]Resource, bin.Size)
	var i int
	bin.PeerIterator(func(entry *entry) bool {
		resources[i] = entry.conn
		i++
		return true
	})
	return klb.resourcesToLbPeers(resources)
}

func (klb *KademliaLoadBalancer) resourcesToLbPeers(resources []Resource) []LBPeer {
	sorted := klb.resourceUseStats.sortResources(resources)
	peers := klb.toLBPeers(sorted)
	return peers
}

func (klb *KademliaLoadBalancer) listenNewPeers() {
	for {
		select {
		case <-klb.quitC:
			return
		case msg, ok := <-klb.onPeerSub.ReceiveChannel():
			if !ok {
				return
			}
			signal, ok := msg.(newPeerSignal)
			if !ok {
				return
			}
			klb.addedPeer(signal.peer, signal.po)
		}
	}
}

func (klb *KademliaLoadBalancer) listenOffPeers() {
	for {
		select {
		case <-klb.quitC:
			return
		case msg := <-klb.offPeerSub.ReceiveChannel():
			peer, ok := msg.(*Peer)
			if peer != nil && ok {
				klb.removedPeer(peer)
			}
		}
	}
}

// addedPeer is called back when a new peer is added to the kademlia. Its uses will be initialized
// to the use count of the least used peer in its bin. The po of the new peer is passed to avoid having
// to calculate it again.
func (klb *KademliaLoadBalancer) addedPeer(peer *Peer, po int) {
	initCount := klb.initCountFunc(peer, 0)
	log.Debug("Adding peer", "key", peer.Label(), "initCount", initCount)
	klb.resourceUseStats.initKey(peer.Key(), initCount)
}

// leastUsedCountInBin returns the use count for the least used peer in this bin excluding the excludePeer.
func (klb *KademliaLoadBalancer) leastUsedCountInBin(excludePeer *Peer, po int) int {
	addr := klb.kademlia.BaseAddr()
	peersInSamePo := klb.getPeersForPo(addr, po)
	idx := 0
	leastUsedCount := 0
	for idx < len(peersInSamePo) {
		leastUsed := peersInSamePo[idx]
		if leastUsed.Peer.Key() != excludePeer.Key() {
			leastUsedCount = klb.resourceUseStats.getUses(leastUsed.Peer)
			log.Debug("Least used peer is", "peer", leastUsed.Peer.Label(), "leastUsedCount", leastUsedCount)
			break
		}
		idx++
	}
	return leastUsedCount
}

// nearestNeighbourCount returns the use count for the closest peer count.
func (klb *KademliaLoadBalancer) nearestNeighbourCount(newPeer *Peer, _ int) int {
	var count int
	klb.kademlia.EachConn(newPeer.Address(), 255, func(peer *Peer, po int) bool {
		if peer != newPeer {
			count = klb.resourceUseStats.getUses(peer)
			log.Debug("Nearest neighbour is", "peer", peer.Label(), "count", count)
			return false
		}
		return true
	})
	return count
}

func (klb *KademliaLoadBalancer) removedPeer(peer *Peer) {
	klb.resourceUseStats.lock.Lock()
	defer klb.resourceUseStats.lock.Lock()
	delete(klb.resourceUseStats.resourceUses, peer.Key())
}

func (klb *KademliaLoadBalancer) toLBPeers(resources []Resource) []LBPeer {
	peers := make([]LBPeer, len(resources))
	for i, res := range resources {
		peer := res.(*Peer)
		peers[i].Peer = peer
		peers[i].Use = func() {
			klb.resourceUseStats.addUse(peer)
		}
	}
	return peers
}

func (klb *KademliaLoadBalancer) getPeersForPo(base []byte, po int) []LBPeer {
	resources := make([]Resource, 0)
	klb.kademlia.EachBinDesc(base, po, func(bin *PeerBin) bool {
		if bin.ProximityOrder == po {
			return bin.PeerIterator(func(entry *entry) bool {
				resources = append(resources, entry.conn)
				return true
			})
		} else {
			return true
		}
	})
	return klb.resourcesToLbPeers(resources)
}

// Resource Use Stats

// resourceUseStats can be used to count uses of resources. A Resource is anything with a Key()
type resourceUseStats struct {
	resourceUses map[string]int
	waiting      map[string]chan struct{}
	lock         sync.RWMutex
}

type Resource interface {
	Key() string   // unique id in string format of the resource.
	Label() string // short string format of the key for debugging purposes.
}

type ResourceCount struct {
	resource Resource
	count    int
}

func newResourceLoadBalancer() *resourceUseStats {
	return &resourceUseStats{
		resourceUses: make(map[string]int),
		waiting:      make(map[string]chan struct{}),
	}
}

func (lb *resourceUseStats) sortResources(resources []Resource) []Resource {
	sorted := make([]Resource, len(resources))
	resourceCounts := lb.getAllUses(resources)
	sort.Slice(resourceCounts, func(i, j int) bool {
		return resourceCounts[i].count < resourceCounts[j].count
	})
	for i, resourceCount := range resourceCounts {
		sorted[i] = resourceCount.resource
	}
	return sorted
}

func (lbp ResourceCount) String() string {
	return lbp.resource.Key() + ":" + strconv.Itoa(lbp.count)
}

func (lb *resourceUseStats) dumpAllUses() map[string]int {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	dump := make(map[string]int)
	for k, v := range lb.resourceUses {
		dump[k] = v
	}
	return dump
}

func (lb *resourceUseStats) getAllUses(resources []Resource) []ResourceCount {
	peerUses := make([]ResourceCount, len(resources))
	for i, resource := range resources {
		peerUses[i] = ResourceCount{
			resource: resource,
			count:    lb.getUses(resource),
		}
	}
	return peerUses
}

func (lb *resourceUseStats) getUses(keyed Resource) int {
	return lb.getKeyUses(keyed.Key())
}

func (lb *resourceUseStats) getKeyUses(key string) int {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	return lb.resourceUses[key]
}

func (lb *resourceUseStats) addUse(resource Resource) int {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	log.Debug("Adding use", "key", resource.Label())
	key := resource.Key()
	lb.resourceUses[key] = lb.resourceUses[key] + 1
	return lb.resourceUses[key]
}

// waitKey blocks until some key is added to the load balancer stats.
// As peer resource initialization is asynchronous we need a way to know that the initial uses has been initialized.
func (lb *resourceUseStats) waitKey(key string) {
	lb.lock.Lock()
	if _, ok := lb.resourceUses[key]; ok {
		lb.lock.Unlock()
		return
	}
	waitChan := make(chan struct{})
	lb.waiting[key] = waitChan
	lb.lock.Unlock()
	<-waitChan
	delete(lb.waiting, key)
}

func (lb *resourceUseStats) initKey(key string, count int) {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	lb.resourceUses[key] = count
	if kChan, ok := lb.waiting[key]; ok {
		kChan <- struct{}{}
	}
}
