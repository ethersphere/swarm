// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.
package network

import (
	"encoding/binary"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/pot"
)

// TestAddedNodes checks that when adding a node it is assigned the correct number of uses.
// This number of uses will be the least number of uses of a peer in its bin.
func TestAddedNodes(t *testing.T) {
	kademlia := newTestKademlia(t, "11110000")
	first := newTestKadPeer("010101010")
	kademlia.Kademlia.On(first)
	second := newTestKadPeer("010101011")
	kademlia.Kademlia.On(second)
	klb := NewKademliaLoadBalancer(kademlia, false)

	defer klb.Stop()
	firstUses := klb.resourceUseStats.GetUses(first)
	if firstUses != 0 {
		t.Errorf("Expected 0 uses for new peer at start")
	}
	peersFor0 := klb.getPeersForPo(kademlia.base, 0)
	peersFor0[0].AddUseCount()
	// Now new peers still should have 0 uses
	third := newTestKadPeer("011101011")
	kademlia.Kademlia.On(third)
	klb.resourceUseStats.WaitKey(third.Key())
	thirdUses := klb.resourceUseStats.GetUses(third)
	if thirdUses != 0 {
		t.Errorf("Expected 0 uses for new peer because minimum in bin is 0. Instead %v", thirdUses)
	}
	peersFor0 = klb.getPeersForPo(kademlia.base, 0)
	peersFor0[0].AddUseCount()
	peersFor0[1].AddUseCount() //Now all peers should have 1 use
	//New peers should start with 1 use
	fourth := newTestKadPeer("011100011")
	kademlia.Kademlia.On(fourth)
	klb.resourceUseStats.WaitKey(fourth.Key())
	fourthUses := klb.resourceUseStats.GetUses(fourth)
	if fourthUses != 1 {
		t.Errorf("Expected 1 use for new peer because minimum in bin should be 1. Instead %v", fourthUses)
	}
}

// TestAddedNodesNearestNeighbour checks that when adding a node it is assigned the correct number of uses.
// This number of uses will be the most similar peer uses.
func TestAddedNodesNearestNeighbour(t *testing.T) {
	kademlia := newTestKademlia(t, "11110000")
	first := newTestKadPeer("01010101")
	kademlia.Kademlia.On(first)
	second := newTestKadPeer("01110101")
	kademlia.Kademlia.On(second)
	klb := NewKademliaLoadBalancer(kademlia, true)

	defer klb.Stop()
	firstUses := klb.resourceUseStats.GetUses(first)
	if firstUses != 0 {
		t.Errorf("Expected 0 uses for new peer at start")
	}
	peersFor0 := klb.getPeersForPo(kademlia.base, 0)
	peersFor0[0].AddUseCount()
	// Now third peer should have the same uses as second
	third := newTestKadPeer("01110111") // most similar peer is second 01110101
	kademlia.Kademlia.On(third)
	klb.resourceUseStats.WaitKey(third.Key())
	secondUses := klb.resourceUseStats.GetUses(second)
	thirdUses := klb.resourceUseStats.GetUses(third)
	if thirdUses != secondUses {
		t.Errorf("Expected %v uses for new peer because is most similar to second. Instead %v", secondUses, thirdUses)
	}
	//Now we use third peer twice
	peersFor0 = klb.getPeersForPo(kademlia.base, 0)
	for _, lbPeer := range peersFor0 {
		if lbPeer.Peer.Key() == third.key {
			lbPeer.AddUseCount()
			lbPeer.AddUseCount()
		}
	}

	fourth := newTestKadPeer("01110110") // most similar peer is third 01110111
	kademlia.Kademlia.On(fourth)
	klb.resourceUseStats.WaitKey(fourth.Key())
	//We expect fourth to be initialized with third peer use count
	fourthUses := klb.resourceUseStats.GetUses(fourth)
	thirdUses = klb.resourceUseStats.GetUses(third)
	if fourthUses != thirdUses {
		t.Errorf("Expected %v use for new peer because most similar is peer 3. Instead %v", thirdUses, fourthUses)
	}

}

var testCount = 0

// TestEachBinBaseUses tests that EachBinDesc returns first the least used peer in its bin
// We will create 3 bins with two peers each. We will call EachBinDesc 6 times twice with an address
// on each bin, so at the end all peers should have 1 use (because the address in each bin is equidistant to
// the peers in that bin).
// Then we will use an address in a bin that is nearer to one of the peers and we will check that that peer is always
// returned first.
func TestEachBinBaseUses(t *testing.T) {
	myCount := testCount
	testCount++
	tk := newTestKademlia(t, "11111111")
	klb := NewKademliaLoadBalancer(tk, false)
	tk.On("01010101") //Peer 1 dec 85 hex 55
	tk.On("01010100") // 2 dec 84 hex 54
	tk.On("10010100") // 3 dec 148 hex 94
	tk.On("10010001") // 4 dec 145 hex 91
	tk.On("11010100") // 5 dec 212 hex d4
	tk.On("11010101") // 6 dec 213 hex d5

	//Waiting for all peers to be registered
	resources := klb.resourceUseStats.Len()
	for resources != 6 {
		time.Sleep(10 * time.Millisecond)
		resources = klb.resourceUseStats.Len()
	}

	pivotAddressBin0 := pot.NewAddressFromString("00000000") // Two nearest peers (1,2) hex 00
	pivotAddressBin1 := pot.NewAddressFromString("10000000") // Two nearest peers (3,4) hex 80
	pivotAddressBin2 := pot.NewAddressFromString("11000000") // Two nearest peers (5,6) hex c0
	countUse := func(bin LBBin) bool {
		peerLogLines := make([]string, 0)
		for idx, lbPeer := range bin.LBPeers {
			currentUses := klb.resourceUseStats.GetUses(lbPeer.Peer)
			peerLogLine := "Peer " + peerToBitString(lbPeer.Peer) + " " + string(idx) + " currentUses " + strconv.FormatInt(int64(currentUses), 10)
			peerLogLines = append(peerLogLines, peerLogLine)
		}

		log.Debug("peers for address in bin", "peers", peerLogLines, "po", bin.ProximityOrder, "count", myCount)
		chosen := bin.LBPeers[0]
		log.Debug("Chosen peer is", "chosen", chosen.Peer.Label(), "uses", klb.resourceUseStats.GetUses(chosen.Peer), "count", myCount)
		chosen.AddUseCount()
		return false
	}
	// Use peer 1 and 2
	klb.EachBinDesc(pivotAddressBin0, countUse)
	klb.EachBinDesc(pivotAddressBin0, countUse)

	peer1Uses := klb.resourceUseStats.GetKeyUses(bitStringToHex("01010101"))
	if peer1Uses != 1 {
		t.Errorf("expected %v uses of %v but got %v", 1, "01010101", peer1Uses)
	}
	peer2Uses := klb.resourceUseStats.GetKeyUses(bitStringToHex("01010100"))
	if peer2Uses != 1 {
		t.Errorf("expected %v uses of %v but got %v", 1, "01010100", peer2Uses)
	}

	// Use peers 3 and 4
	klb.EachBinDesc(pivotAddressBin1, countUse)
	klb.EachBinDesc(pivotAddressBin1, countUse)

	peer3Uses := klb.resourceUseStats.GetKeyUses(bitStringToHex("10010100"))
	if peer3Uses != 1 {
		t.Errorf("expected %v uses of %v but got %v", 1, "10010100", peer3Uses)
	}
	peer4Uses := klb.resourceUseStats.GetKeyUses(bitStringToHex("10010001"))
	if peer4Uses != 1 {
		t.Errorf("expected %v uses of %v but got %v", 1, "10010001", peer4Uses)
	}

	// Use peers 5 and 6
	klb.EachBinDesc(pivotAddressBin2, countUse)
	klb.EachBinDesc(pivotAddressBin2, countUse)

	peer5Uses := klb.resourceUseStats.GetKeyUses(bitStringToHex("11010100"))
	if peer5Uses != 1 {
		t.Errorf("expected %v uses of %v but got %v", 1, "11010100", peer5Uses)
	}
	peer6Uses := klb.resourceUseStats.GetKeyUses(bitStringToHex("11010101"))
	if peer6Uses != 1 {
		t.Errorf("expected %v uses of %v but got %v", 1, "11010101", peer6Uses)
	}

	//Now a message that is nearer 10010001 than 10010100 in its bin. It will be taken always regardless of uses
	pivotAddressBin3 := pot.NewAddressFromString("10010011") // Nearer 4 hex 93

	//Both calls to 4
	klb.EachBinDesc(pivotAddressBin3, countUse)
	klb.EachBinDesc(pivotAddressBin3, countUse)

	count := klb.resourceUseStats.GetKeyUses(bitStringToHex("10010001"))
	if count != 3 {
		t.Errorf("Expected 3 uses of 10010001 but got %v", count)
	}
}

func expectUses(actualUses int, expected int, peer string, t *testing.T) {
	if actualUses != expected {
		t.Errorf("expected %v uses of %v but got %v", expected, peer, actualUses)
	}
}

// TestEachBinFiltered checks that when load balancing peers, only those with the provided capabilities are chosen.
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
	useStats := klb.resourceUseStats
	useStats.WaitKey(capPeer.Key())
	tk.On("01010101") // bin 0 dec 85 hex 55
	useStats.WaitKey(bitStringToHex("01010101"))
	tk.On("01010100") // bin 0 dec 84 hex 54
	useStats.WaitKey(bitStringToHex("01010100"))
	tk.On("10010100") // bin 1 dec 148
	useStats.WaitKey(bitStringToHex("10010100"))
	tk.On("10010001") // bin 1 dec 145
	useStats.WaitKey(bitStringToHex("10010001"))
	tk.On("11010100") // bin 2 dec 212
	useStats.WaitKey(bitStringToHex("11010100"))
	tk.On("11010101") // bin 2 dec 213
	useStats.WaitKey(bitStringToHex("11010101"))
	stats := make(map[string]int)
	countUse := func(bin LBBin) bool {
		peer := bin.LBPeers[0].Peer
		bin.LBPeers[0].AddUseCount()
		key := peerToBitString(peer)
		stats[key] = stats[key] + 1
		return false
	}

	pivotAddressBin1 := pot.NewAddressFromString("10000000") // Two nearest peers (1,2)
	// Instead of selecting peers 10010100 or 10010001, capPeer is always chosen (10100000)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)

	count := useStats.GetUses(capPeer)
	if count != 3 || stats["10100000"] != 3 {
		t.Errorf("Expected 3 uses of capability peer but got %v/%v", count, stats["10100000"])
	}

	secondCapPeer := tk.newTestKadPeerWithCapabilities("10100001", caps[capKey])
	tk.Kademlia.On(secondCapPeer)
	useStats.WaitKey(secondCapPeer.Key())
	secondCountStart := useStats.GetUses(secondCapPeer)
	count = useStats.GetUses(capPeer)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	klb.EachBinFiltered(pivotAddressBin1, capKey, countUse)
	secondCount := useStats.GetUses(secondCapPeer)
	if secondCount-secondCountStart != 2 {
		t.Errorf("Expected 2 uses of second capability peer but got %v", secondCount-secondCountStart)
	}

}

// TestResourceUseStats checks that on and off messages are delivered in order
func TestResourceUseStats(t *testing.T) {

	testResourceUseStats := func(t *testing.T, delay time.Duration) {
		k := NewKademlia(make([]byte, 32), NewKadParams())
		lb := NewKademliaLoadBalancer(k, false)
		for i := uint64(0); i < 10; i++ {
			a := make([]byte, 8)
			binary.BigEndian.PutUint64(a, i)
			p := NewPeer(&BzzPeer{BzzAddr: NewBzzAddr(a, nil)}, nil)
			k.On(p)
			if delay > 0 {
				time.Sleep(delay)
			}
			k.Off(p)
			if delay > 0 {
				time.Sleep(delay)
			}
		}

		// we need to sleep to allow all messages to be received by lb
		count := 0
		retries := 0
		for count == 0 && retries < 15 {
			time.Sleep(10 * time.Millisecond)
			count = lb.resourceUseStats.Len()
			retries++
		}
		if count > 0 {
			t.Errorf("got resourceUseStats %v, want 0, uses: %v", count, lb.resourceUseStats.DumpAllUses())
		}
		lb.Stop()
	}

	t.Run("no delay", func(t *testing.T) {
		testResourceUseStats(t, 0)
	})
	t.Run("1ms delay", func(t *testing.T) {
		testResourceUseStats(t, time.Millisecond)
	})
}

func newTestKadPeer(s string) *Peer {
	return NewPeer(&BzzPeer{BzzAddr: testKadPeerAddr(s)}, nil)
}

// Debug functions

// bitStringToHex converts an address in bit format (11001100) to hex format. BitString format is used to create test
// peers, hex format is used in the load balancer stats.
func bitStringToHex(binary string) string {
	var byteSlice = make([]byte, 32)
	i, _ := strconv.ParseInt(binary, 2, 0)
	byteSlice[0] = byte(i)
	return hexutil.Encode(byteSlice)
}

// converts the peer address to bit string format
func peerToBitString(peer *Peer) string {
	return byteToBitString(peer.Address()[0])
}

func byteToBitString(b byte) string {
	binary := strconv.FormatUint(uint64(b), 2)
	if len(binary) < 8 {
		for i := 8 - len(binary); i > 0; i-- {
			binary = "0" + binary
		}
	}
	return binary
}
