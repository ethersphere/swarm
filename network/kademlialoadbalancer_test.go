package network

import (
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/pot"
)

// TestAddedNodes checks that when adding a node it is assigned the correct number of uses.
// This number of uses will be the least number of uses of a peer in its bin
func TestAddedNodes(t *testing.T) {
	kademlia := newTestKademliaBackend("11110000")
	first := newTestKadPeer("010101010")
	kademlia.addPeer(first, 0)
	second := newTestKadPeer("010101011")
	kademlia.addPeer(second, 0)
	klb := NewKademliaLoadBalancer(kademlia, false)

	defer klb.Stop()
	firstUses := klb.resourceUseStats.getUses(first)
	if firstUses != 0 {
		t.Errorf("Expected 0 uses for new peer at start")
	}
	peersFor0 := klb.getPeersForPo(kademlia.baseAddr, 0)
	peersFor0[0].Use()
	// Now new peers still should have 0 uses
	third := newTestKadPeer("011101011")
	kademlia.addPeer(third, 0)
	klb.resourceUseStats.waitKey(third.Key())
	thirdUses := klb.resourceUseStats.getUses(third)
	if thirdUses != 0 {
		t.Errorf("Expected 0 uses for new peer because minimum in bin is 0. Instead %v", thirdUses)
	}
	peersFor0 = klb.getPeersForPo(kademlia.baseAddr, 0)
	peersFor0[0].Use()
	peersFor0[1].Use() //Now all peers should have 1 use
	//New peers should start with 1 use
	fourth := newTestKadPeer("011100011")
	kademlia.addPeer(fourth, 0)
	klb.resourceUseStats.waitKey(fourth.Key())
	fourthUses := klb.resourceUseStats.getUses(fourth)
	if fourthUses != 1 {
		t.Errorf("Expected 1 use for new peer because minimum in bin should be 1. Instead %v", fourthUses)
	}
}

// TestAddedNodesMostSimilar checks that when adding a node it is assigned the correct number of uses.
// This number of uses will be the most similar peer uses.
func TestAddedNodesMostSimilar(t *testing.T) {
	kademlia := newTestKademliaBackend("11110000")
	first := newTestKadPeer("01010101")
	kademlia.addPeer(first, 0)
	second := newTestKadPeer("01110101")
	kademlia.addPeer(second, 0)
	klb := NewKademliaLoadBalancer(kademlia, true)

	defer klb.Stop()
	firstUses := klb.resourceUseStats.getUses(first)
	if firstUses != 0 {
		t.Errorf("Expected 0 uses for new peer at start")
	}
	peersFor0 := klb.getPeersForPo(kademlia.baseAddr, 0)
	peersFor0[0].Use()
	// Now third peer should have the same uses as second
	third := newTestKadPeer("01110111") // most similar peer is second 01110101
	kademlia.addPeer(third, 0)
	klb.resourceUseStats.waitKey(third.Key())
	secondUses := klb.resourceUseStats.getUses(second)
	thirdUses := klb.resourceUseStats.getUses(third)
	if thirdUses != secondUses {
		t.Errorf("Expected %v uses for new peer because is most similar to second. Instead %v", secondUses, thirdUses)
	}
	peersFor0 = klb.getPeersForPo(kademlia.baseAddr, 0)
	peersFor0[0].Use()
	peersFor0[1].Use()
	//New peers should start with 1 use
	fourth := newTestKadPeer("01110110") // most similar peer is third 01110111
	kademlia.addPeer(fourth, 0)
	klb.resourceUseStats.waitKey(fourth.Key())
	fourthUses := klb.resourceUseStats.getUses(fourth)
	thirdUses = klb.resourceUseStats.getUses(third)
	if fourthUses != thirdUses {
		t.Errorf("Expected %v use for new peer because most similiar is peer 3. Instead %v", thirdUses, fourthUses)
	}

}

// TestEachBinBaseUses tests that EachBin returns first the least used peer in its bin
// We will create 3 bins with two peers each. We will call EachBin 6 times twice with an address
// on each bin, so at the end all peers should have 1 use (because the address in each bin is equidistant to
// the peers in that bin).
// Then wi will use an address in a bin that is nearer one of the peers and we will check that that peer is always
// returned first
func TestEachBinBaseUses(t *testing.T) {
	tk := newTestKademlia(t, "11111111")
	klb := NewKademliaLoadBalancer(tk, false)
	tk.On("01010101") //Peer 1 dec 85
	tk.On("01010100") // 2 dec 84
	tk.On("10010100") // 3 dec 148
	tk.On("10010001") // 4 dec 145
	tk.On("11010100") // 5 dec 212
	tk.On("11010101") // 6 dec 213

	pivotAddressBin0 := pot.NewAddressFromString("00000000") // Two nearest peers (1,2)
	pivotAddressBin1 := pot.NewAddressFromString("10000000") // Two nearest peers (3,4)
	pivotAddressBin2 := pot.NewAddressFromString("11000000") // Two nearest peers (5,6)
	countUse := func(bin LBBin) bool {
		bin.LBPeers[0].Use()
		return false
	}
	// Use peer 1 and 2
	klb.EachBin(pivotAddressBin0, countUse)
	klb.EachBin(pivotAddressBin0, countUse)

	// Use peers 3 and 4
	klb.EachBin(pivotAddressBin1, countUse)
	klb.EachBin(pivotAddressBin1, countUse)

	// Use peers 5 and 6
	klb.EachBin(pivotAddressBin2, countUse)
	klb.EachBin(pivotAddressBin2, countUse)

	resourceUses := klb.resourceUseStats.dumpAllUses()
	if len(resourceUses) != 6 {
		t.Errorf("Expected all 6 peers to be used but got %v", len(resourceUses))
	}
	for key, uses := range resourceUses {
		if uses != 1 {
			bytes, _ := hexutil.Decode(key)
			binaryKey := byteToBinary(bytes[0]) + byteToBinary(bytes[1])
			t.Errorf("Expected only 1 use of %v but got %v", binaryKey, uses)
		}
	}

	//Now a message that is nearer 10010001 than 10010100 in its bin. It will be taken always regardless of uses
	pivotAddressBin3 := pot.NewAddressFromString("10010011") // Nearer 4

	//Both calls to 4
	klb.EachBin(pivotAddressBin3, countUse)
	klb.EachBin(pivotAddressBin3, countUse)

	count := klb.resourceUseStats.getKeyUses(stringBinaryToHex("10010001"))
	if count != 3 {
		t.Errorf("Expected 3 uses of 10010001 but got %v", count)
	}
}

func TestEachBinFiltered(t *testing.T) {
	tk := newTestKademlia(t, "11111111")
	klb := NewKademliaLoadBalancer(tk, false)
	caps := make(map[string]*capability.Capability)

	capKey := "42:101"
	caps[capKey] = capability.NewCapability(42, 3)
	_ = caps[capKey].Set(0)
	_ = caps[capKey].Set(2)
	_ = tk.RegisterCapabilityIndex(capKey, *caps[capKey])

	capPeer := tk.newTestKadPeerWithCapabilities("10100000", caps[capKey])
	tk.Kademlia.On(capPeer)

	tk.On("01010101") // bin 0 dec 85
	tk.On("01010100") // bin 0 dec 84
	tk.On("10010100") // bin 1 dec 148
	tk.On("10010001") // bin 1 dec 145
	tk.On("11010100") // bin 2 dec 212
	tk.On("11010101") // bin 2 dec 213
	stats := make(map[string]int)
	countUse := func(bin LBBin) bool {
		peer := bin.LBPeers[0].Peer
		bin.LBPeers[0].Use()
		key := peerToBinaryId(peer)
		stats[key] = stats[key] + 1
		return false
	}

	pivotAddressBin1 := pot.NewAddressFromString("10000000") // Two nearest peers (1,2)
	// Instead of selecting peers 10010100 or 10010001, capPeer is always chosen (10100000)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)

	useStats := klb.resourceUseStats
	count := useStats.getUses(capPeer)
	if count != 3 || stats["10100000"] != 3 {
		t.Errorf("Expected 3 uses of capability peer but got %v/%v", count, stats["10100000"])
	}

	secondCapPeer := tk.newTestKadPeerWithCapabilities("10100000", caps[capKey])
	tk.Kademlia.On(secondCapPeer)
	useStats.waitKey(secondCapPeer.Key())
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	secondCount := useStats.getUses(secondCapPeer)
	if secondCount == 0 {
		t.Errorf("Expected some use of second capability peer but got %v", secondCount)
	}

}

type testKademliaBackend struct {
	baseAddr       []byte
	addedChannel   chan newPeerSignal
	removedChannel chan *Peer
	bins           map[int][]*Peer
	subscribed     bool
	maxPo          int
}

func (tkb *testKademliaBackend) EachConn(base []byte, maxPo int, consume func(*Peer, int) bool) {
	po, _ := Pof(base, tkb.baseAddr, 0)
	bin := tkb.bins[po]
	peersInBin := make([]*Peer, len(bin))
	copy(peersInBin, bin)
	sort.Slice(peersInBin, func(i, j int) bool {
		peerIPo, _ := Pof(base, peersInBin[i], 0)
		peerJPo, _ := Pof(base, peersInBin[j], 0)
		return peerIPo > peerJPo
	})
	for _, peer := range peersInBin {
		if !consume(peer, po) {
			return
		}
	}
	for i := po + 1; po < maxPo ; po++ {
		bin = tkb.bins[i]
		for _, peer := range bin {
			if !consume(peer, po) {
				return
			}
		}
	}
	for i := po - 1; po >= 0 ; po-- {
		bin = tkb.bins[i]
		for _, peer := range bin {
			if !consume(peer, i) {
				return
			}
		}
	}

}

func newTestKademliaBackend(address string) *testKademliaBackend {
	return &testKademliaBackend{
		baseAddr:       pot.NewAddressFromString(address),
		addedChannel:   make(chan newPeerSignal, 1),
		removedChannel: make(chan *Peer, 1),
		bins:           make(map[int][]*Peer),
	}
}

func (tkb testKademliaBackend) BaseAddr() []byte {
	return tkb.baseAddr
}

func (tkb *testKademliaBackend) SubscribeToPeerChanges() (addedC <-chan newPeerSignal, removedC <-chan *Peer, unsubscribe func()) {
	unsubscribe = func() {
		tkb.subscribed = false
		close(tkb.addedChannel)
		close(tkb.removedChannel)
	}
	tkb.subscribed = true
	return tkb.addedChannel, tkb.removedChannel, unsubscribe
}

func (tkb testKademliaBackend) EachBinDescFiltered(base []byte, capKey string, minProximityOrder int, consumer PeerBinConsumer) error {
	tkb.EachBinDesc(base, minProximityOrder, consumer)
	return nil
}

func (tkb testKademliaBackend) EachBinDesc(_ []byte, minProximityOrder int, consumer PeerBinConsumer) {
	type poPeers struct {
		po    int
		peers []*Peer
	}
	var poPeersList []poPeers
	for po, peers := range tkb.bins {
		poPeersList = append(poPeersList, poPeers{po: po, peers: peers})
	}
	sort.Slice(poPeersList, func(i, j int) bool {
		return poPeersList[i].po > poPeersList[j].po
	})
	for _, aPoPeers := range poPeersList {
		peers := aPoPeers.peers
		po := aPoPeers.po
		if peers != nil && po >= minProximityOrder {
			bin := &PeerBin{
				ProximityOrder: po,
				Size:           len(peers),
				PeerIterator: func(consumePeer PeerConsumer) bool {
					for _, peer := range peers {
						if !consumePeer(&entry{conn: peer}) {
							return false
						}
					}
					return true
				},
			}
			if !consumer(bin) {
				return
			}
		}
	}
}

func (tkb *testKademliaBackend) addPeer(peer *Peer, po int) {
	if tkb.bins[po] == nil {
		if po > tkb.maxPo {
			tkb.maxPo = po
		}
		tkb.bins[po] = make([]*Peer, 0)
	}
	tkb.bins[po] = append(tkb.bins[po], peer)
	if tkb.subscribed {
		tkb.addedChannel <- newPeerSignal{
			peer: peer,
			po:   po,
		}
	}
	time.Sleep(100 * time.Millisecond)
}

func (tkb *testKademliaBackend) removePeer(peer *Peer) {
	for po, bin := range tkb.bins {
		for i, aPeer := range bin {
			if aPeer == peer {
				tkb.bins[po] = append(bin[:i], bin[i+1:]...)
				if len(tkb.bins[po]) == 0  && tkb.maxPo >= po{
					tkb.updateMaxPo()
				}
				break
			}
		}
	}
	if tkb.subscribed {
		tkb.removedChannel <- peer
	}
}

func (tkb *testKademliaBackend) updateMaxPo() {
	tkb.maxPo = 0
	for k := range tkb.bins {
		if k > tkb.maxPo {
			tkb.maxPo = k
		}
	}
}

func newTestKadPeer(s string) *Peer {
	return NewPeer(&BzzPeer{BzzAddr: testKadPeerAddr(s)}, nil)
}
