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

package network

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pot"
)

func init() {
	h := log.LvlFilterHandler(log.LvlWarn, log.StreamHandler(os.Stderr, log.TerminalFormat(true)))
	log.Root().SetHandler(h)
}

func testKadPeerAddr(s string) *BzzAddr {
	a := pot.NewAddressFromString(s)
	return NewBzzAddr(a, a)
}

func newTestKademliaParams() *KadParams {
	params := NewKadParams()
	params.MinBinSize = 2
	params.NeighbourhoodSize = 2
	return params
}

type testKademlia struct {
	*Kademlia
	t *testing.T
}

func newTestKademlia(t *testing.T, b string) *testKademlia {
	base := pot.NewAddressFromString(b)
	return &testKademlia{
		Kademlia: NewKademlia(base, newTestKademliaParams()),
		t:        t,
	}
}

func (tk *testKademlia) newTestKadPeer(s string) *Peer {
	return NewPeer(&BzzPeer{BzzAddr: testKadPeerAddr(s)}, tk.Kademlia)
}

func (tk *testKademlia) On(ons ...string) {
	for _, s := range ons {
		tk.Kademlia.On(tk.newTestKadPeer(s))
	}
}

func (tk *testKademlia) Off(offs ...string) {
	for _, s := range offs {
		tk.Kademlia.Off(tk.newTestKadPeer(s))
	}
}

func (tk *testKademlia) Register(regs ...string) {
	var as []*BzzAddr
	for _, s := range regs {
		as = append(as, testKadPeerAddr(s))
	}
	err := tk.Kademlia.Register(as...)
	if err != nil {
		panic(err.Error())
	}
}

// tests the validity of neighborhood depth calculations
//
// in particular, it tests that if there are one or more consecutive
// empty bins above the farthest "nearest neighbor-peer" then
// the depth should be set at the farthest of those empty bins
//
// TODO: Make test adapt to change in NeighbourhoodSize
func TestNeighbourhoodDepth(t *testing.T) {
	baseAddressBytes := RandomBzzAddr().OAddr
	kad := NewKademlia(baseAddressBytes, NewKadParams())

	baseAddress := pot.NewAddressFromBytes(baseAddressBytes)

	// generate the peers
	var peers []*Peer
	for i := 0; i < 7; i++ {
		addr := pot.RandomAddressAt(baseAddress, i)
		peers = append(peers, newTestDiscoveryPeer(addr, kad))
	}
	var sevenPeers []*Peer
	for i := 0; i < 2; i++ {
		addr := pot.RandomAddressAt(baseAddress, 7)
		sevenPeers = append(sevenPeers, newTestDiscoveryPeer(addr, kad))
	}

	testNum := 0
	// first try with empty kademlia
	depth := kad.NeighbourhoodDepth()
	if depth != 0 {
		t.Fatalf("%d expected depth 0, was %d", testNum, depth)
	}
	testNum++

	// add one peer on 7
	kad.On(sevenPeers[0])
	depth = kad.NeighbourhoodDepth()
	if depth != 0 {
		t.Fatalf("%d expected depth 0, was %d", testNum, depth)
	}
	testNum++

	// add a second on 7
	kad.On(sevenPeers[1])
	depth = kad.NeighbourhoodDepth()
	if depth != 0 {
		t.Fatalf("%d expected depth 0, was %d", testNum, depth)
	}
	testNum++

	// add from 0 to 6
	for i, p := range peers {
		kad.On(p)
		depth = kad.NeighbourhoodDepth()
		if depth != i+1 {
			t.Fatalf("%d.%d expected depth %d, was %d", i+1, testNum, i, depth)
		}
	}
	testNum++

	kad.Off(sevenPeers[1])
	depth = kad.NeighbourhoodDepth()
	if depth != 6 {
		t.Fatalf("%d expected depth 6, was %d", testNum, depth)
	}
	testNum++

	kad.Off(peers[4])
	depth = kad.NeighbourhoodDepth()
	if depth != 4 {
		t.Fatalf("%d expected depth 4, was %d", testNum, depth)
	}
	testNum++

	kad.Off(peers[3])
	depth = kad.NeighbourhoodDepth()
	if depth != 3 {
		t.Fatalf("%d expected depth 3, was %d", testNum, depth)
	}
	testNum++
}

// TestHighMinBinSize tests that the saturation function also works
// if MinBinSize is > 2, the connection count is < k.MinBinSize
// and there are more peers available than connected
func TestHighMinBinSize(t *testing.T) {
	// a function to test for different MinBinSize values
	testKad := func(minBinSize int) {
		// create a test kademlia
		tk := newTestKademlia(t, "11111111")
		// set its MinBinSize to desired value
		tk.KadParams.MinBinSize = minBinSize

		// add a couple of peers (so we have NN and depth)
		tk.On("00000000") // bin 0
		tk.On("11100000") // bin 3
		tk.On("11110000") // bin 4

		first := "10000000" // add a first peer at bin 1
		tk.Register(first)  // register it
		// we now have one registered peer at bin 1;
		// iterate and connect one peer at each iteration;
		// should be unhealthy until at minBinSize - 1
		// we connect the unconnected but registered peer
		for i := 1; i < minBinSize; i++ {
			peer := fmt.Sprintf("1000%b", 8|i)
			tk.On(peer)
			if i == minBinSize-1 {
				tk.On(first)
				tk.checkHealth(true)
				return
			}
			tk.checkHealth(false)
		}
	}
	// test MinBinSizes of 3 to 5
	testMinBinSizes := []int{3, 4, 5}
	for _, k := range testMinBinSizes {
		testKad(k)
	}
}

// TestHealthStrict tests the simplest definition of health
// Which means whether we are connected to all neighbors we know of
func TestHealthStrict(t *testing.T) {

	// base address is all zeros
	// no peers
	// unhealthy (and lonely)
	tk := newTestKademlia(t, "11111111")
	tk.checkHealth(false)

	// know one peer but not connected
	// unhealthy
	tk.Register("11100000")
	tk.checkHealth(false)

	// know one peer and connected
	// unhealthy: not saturated
	tk.On("11100000")
	tk.checkHealth(true)

	// know two peers, only one connected
	// unhealthy
	tk.Register("11111100")
	tk.checkHealth(false)

	// know two peers and connected to both
	// healthy
	tk.On("11111100")
	tk.checkHealth(true)

	// know three peers, connected to the two deepest
	// healthy
	tk.Register("00000000")
	tk.checkHealth(false)

	// know three peers, connected to all three
	// healthy
	tk.On("00000000")
	tk.checkHealth(true)

	// add fourth peer deeper than current depth
	// unhealthy
	tk.Register("11110000")
	tk.checkHealth(false)

	// connected to three deepest peers
	// healthy
	tk.On("11110000")
	tk.checkHealth(true)

	// add additional peer in same bin as deepest peer
	// unhealthy
	tk.Register("11111101")
	tk.checkHealth(false)

	// four deepest of five peers connected
	// healthy
	tk.On("11111101")
	tk.checkHealth(true)

	// add additional peer in bin 0
	// unhealthy: unsaturated bin 0, 2 known but 1 connected
	tk.Register("00000001")
	tk.checkHealth(false)

	// Connect second in bin 0
	// healthy
	tk.On("00000001")
	tk.checkHealth(true)

	// add peer in bin 1
	// unhealthy, as it is known but not connected
	tk.Register("10000000")
	tk.checkHealth(false)

	// connect  peer in bin 1
	// depth change, is now 1
	// healthy, 1 peer in bin 1 known and connected
	tk.On("10000000")
	tk.checkHealth(true)

	// add second peer in bin 1
	// unhealthy, as it is known but not connected
	tk.Register("10000001")
	tk.checkHealth(false)

	// connect second peer in bin 1
	// healthy,
	tk.On("10000001")
	tk.checkHealth(true)

	// connect third peer in bin 1
	// healthy,
	tk.On("10000011")
	tk.checkHealth(true)

	// add peer in bin 2
	// unhealthy, no depth change
	tk.Register("11000000")
	tk.checkHealth(false)

	// connect peer in bin 2
	// depth change - as we already have peers in bin 3 and 4,
	// we have contiguous bins, no bin < po 5 is empty -> depth 5
	// healthy, every bin < depth has the max available peers,
	// even if they are < MinBinSize
	tk.On("11000000")
	tk.checkHealth(true)

	// add peer in bin 2
	// unhealthy, peer bin is below depth 5 but
	// has more available peers (2) than connected ones (1)
	// --> unsaturated
	tk.Register("11000011")
	tk.checkHealth(false)
}

func (tk *testKademlia) checkHealth(expectHealthy bool) {
	tk.t.Helper()
	kid := common.Bytes2Hex(tk.BaseAddr())
	addrs := [][]byte{tk.BaseAddr()}
	tk.EachAddr(nil, 255, func(addr *BzzAddr, po int) bool {
		addrs = append(addrs, addr.Address())
		return true
	})

	pp := NewPeerPotMap(tk.NeighbourhoodSize, addrs)
	healthParams := tk.GetHealthInfo(pp[kid])

	// definition of health, all conditions but be true:
	// - we at least know one peer
	// - we know all neighbors
	// - we are connected to all known neighbors
	health := healthParams.Healthy()
	if expectHealthy != health {
		tk.t.Fatalf("expected kademlia health %v, is %v\n%v", expectHealthy, health, tk.String())
	}
}

func (tk *testKademlia) checkSuggestPeer(expAddr string, expDepth int, expChanged bool) {
	tk.t.Helper()
	addr, depth, changed := tk.SuggestPeer()
	log.Trace("suggestPeer return", "addr", addr, "depth", depth, "changed", changed)
	if binStr(addr) != expAddr {
		tk.t.Fatalf("incorrect peer address suggested. expected %v, got %v", expAddr, binStr(addr))
	}
	if depth != expDepth {
		tk.t.Fatalf("incorrect saturation depth suggested. expected %v, got %v", expDepth, depth)
	}
	if changed != expChanged {
		tk.t.Fatalf("expected depth change = %v, got %v", expChanged, changed)
	}
}

func binStr(a *BzzAddr) string {
	if a == nil {
		return "<nil>"
	}
	return pot.ToBin(a.Address())[:8]
}

func TestSuggestPeerFindPeers(t *testing.T) {
	tk := newTestKademlia(t, "00000000")
	tk.On("00100000")
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.On("00010000")
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.On("10000000", "10000001")
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.On("01000000")
	tk.Off("10000001")
	tk.checkSuggestPeer("10000001", 0, true)

	tk.On("00100001")
	tk.Off("01000000")
	tk.checkSuggestPeer("01000000", 0, false)

	// second time disconnected peer not callable
	// with reasonably set Interval
	tk.checkSuggestPeer("<nil>", 0, false)

	// on and off again, peer callable again
	tk.On("01000000")
	tk.Off("01000000")
	tk.checkSuggestPeer("01000000", 0, false)

	tk.On("01000000", "10000001")
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.Register("00010001")
	tk.checkSuggestPeer("00010001", 0, false)

	tk.On("00010001")
	tk.Off("01000000")
	tk.checkSuggestPeer("01000000", 0, false)

	tk.On("01000000")
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.Register("01000001")
	tk.checkSuggestPeer("01000001", 0, false)

	tk.On("01000001")
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.Register("10000010", "01000010", "00100010")
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.Register("00010010")
	tk.checkSuggestPeer("00010010", 0, false)

	tk.Off("00100001")
	tk.checkSuggestPeer("00100010", 2, true)

	tk.Off("01000001")
	tk.checkSuggestPeer("01000010", 1, true)

	tk.checkSuggestPeer("01000001", 0, false)
	tk.checkSuggestPeer("00100001", 0, false)
	tk.checkSuggestPeer("<nil>", 0, false)

	tk.On("01000001", "00100001")
	tk.Register("10000100", "01000100", "00100100")
	tk.Register("00000100", "00000101", "00000110")
	tk.Register("00000010", "00000011", "00000001")

	tk.checkSuggestPeer("00000110", 0, false)
	tk.checkSuggestPeer("00000101", 0, false)
	tk.checkSuggestPeer("00000100", 0, false)
	tk.checkSuggestPeer("00000011", 0, false)
	tk.checkSuggestPeer("00000010", 0, false)
	tk.checkSuggestPeer("00000001", 0, false)
	tk.checkSuggestPeer("<nil>", 0, false)

}

// a node should stay in the address book if it's removed from the kademlia
func TestOffEffectingAddressBookNormalNode(t *testing.T) {
	tk := newTestKademlia(t, "00000000")
	// peer added to kademlia
	tk.On("01000000")
	// peer should be in the address book
	if tk.addrs.Size() != 1 {
		t.Fatal("known peer addresses should contain 1 entry")
	}
	// peer should be among live connections
	if tk.conns.Size() != 1 {
		t.Fatal("live peers should contain 1 entry")
	}
	// remove peer from kademlia
	tk.Off("01000000")
	// peer should be in the address book
	if tk.addrs.Size() != 1 {
		t.Fatal("known peer addresses should contain 1 entry")
	}
	// peer should not be among live connections
	if tk.conns.Size() != 0 {
		t.Fatal("live peers should contain 0 entry")
	}
}

func TestSuggestPeerRetries(t *testing.T) {
	tk := newTestKademlia(t, "00000000")
	tk.RetryInterval = int64(300 * time.Millisecond) // cycle
	tk.MaxRetries = 50
	tk.RetryExponent = 2
	sleep := func(n int) {
		ts := tk.RetryInterval
		for i := 1; i < n; i++ {
			ts *= int64(tk.RetryExponent)
		}
		time.Sleep(time.Duration(ts))
	}

	tk.Register("01000000")
	tk.On("00000001", "00000010")
	tk.checkSuggestPeer("01000000", 0, false)

	tk.checkSuggestPeer("<nil>", 0, false)

	sleep(1)
	tk.checkSuggestPeer("01000000", 0, false)

	tk.checkSuggestPeer("<nil>", 0, false)

	sleep(1)
	tk.checkSuggestPeer("01000000", 0, false)

	tk.checkSuggestPeer("<nil>", 0, false)

	sleep(2)
	tk.checkSuggestPeer("01000000", 0, false)

	tk.checkSuggestPeer("<nil>", 0, false)

	sleep(2)
	tk.checkSuggestPeer("<nil>", 0, false)
}

func TestKademliaHiveString(t *testing.T) {
	tk := newTestKademlia(t, "00000000")
	tk.On("01000000", "00100000")
	tk.Register("10000000", "10000001")
	tk.MaxProxDisplay = 8
	h := tk.String()
	expH := "\n=========================================================================\nMon Feb 27 12:10:28 UTC 2017 KΛÐΞMLIΛ hive: queen's address: 0000000000000000000000000000000000000000000000000000000000000000\npopulation: 2 (4), NeighbourhoodSize: 2, MinBinSize: 2, MaxBinSize: 4\n============ DEPTH: 0 ==========================================\n000  0                              |  2 8100 (0) 8000 (0)\n001  1 4000                         |  1 4000 (0)\n002  1 2000                         |  1 2000 (0)\n003  0                              |  0\n004  0                              |  0\n005  0                              |  0\n006  0                              |  0\n007  0                              |  0\n========================================================================="
	if expH[104:] != h[104:] {
		t.Fatalf("incorrect hive output. expected %v, got %v", expH, h)
	}
}

func newTestDiscoveryPeer(addr pot.Address, kad *Kademlia) *Peer {
	rw := &p2p.MsgPipeRW{}
	p := p2p.NewPeer(enode.ID{}, "foo", []p2p.Cap{})
	pp := protocols.NewPeer(p, rw, &protocols.Spec{})
	bp := &BzzPeer{
		Peer:    pp,
		BzzAddr: NewBzzAddr(addr.Bytes(), []byte(fmt.Sprintf("%x", addr[:]))),
	}
	return NewPeer(bp, kad)
}

// TestKademlia_SubscribeToNeighbourhoodDepthChange checks if correct
// signaling over SubscribeToNeighbourhoodDepthChange channels are made
// when neighbourhood depth is changed.
func TestKademlia_SubscribeToNeighbourhoodDepthChange(t *testing.T) {

	testSignal := func(t *testing.T, k *testKademlia, prevDepth int, c <-chan struct{}) (newDepth int) {
		t.Helper()

		select {
		case _, ok := <-c:
			if !ok {
				t.Error("closed signal channel")
			}
			newDepth = k.NeighbourhoodDepth()
			if prevDepth == newDepth {
				t.Error("depth not changed")
			}
			return newDepth
		case <-time.After(2 * time.Second):
			t.Error("timeout")
		}
		return newDepth
	}

	t.Run("single subscription", func(t *testing.T) {
		k := newTestKademlia(t, "00000000")

		c, u := k.SubscribeToNeighbourhoodDepthChange()
		defer u()

		depth := k.NeighbourhoodDepth()

		k.On("11111101", "01000000", "10000000", "00000010")

		testSignal(t, k, depth, c)
	})

	t.Run("multiple subscriptions", func(t *testing.T) {
		k := newTestKademlia(t, "00000000")

		c1, u1 := k.SubscribeToNeighbourhoodDepthChange()
		defer u1()

		c2, u2 := k.SubscribeToNeighbourhoodDepthChange()
		defer u2()

		depth := k.NeighbourhoodDepth()

		k.On("11111101", "01000000", "10000000", "00000010")

		testSignal(t, k, depth, c1)

		testSignal(t, k, depth, c2)
	})

	t.Run("multiple changes", func(t *testing.T) {
		k := newTestKademlia(t, "00000000")

		c, u := k.SubscribeToNeighbourhoodDepthChange()
		defer u()

		depth := k.NeighbourhoodDepth()

		k.On("11111101", "01000000", "10000000", "00000010")

		depth = testSignal(t, k, depth, c)

		k.On("11111101", "01000010", "10000010", "00000110")

		testSignal(t, k, depth, c)
	})

	t.Run("no depth change", func(t *testing.T) {
		k := newTestKademlia(t, "00000000")

		c, u := k.SubscribeToNeighbourhoodDepthChange()
		defer u()

		// does not trigger the depth change
		k.On("11111101")

		select {
		case _, ok := <-c:
			if !ok {
				t.Error("closed signal channel")
			}
			t.Error("signal received")
		case <-time.After(1 * time.Second):
			// all fine
		}
	})

	t.Run("no new peers", func(t *testing.T) {
		k := newTestKademlia(t, "00000000")

		changeC, unsubscribe := k.SubscribeToNeighbourhoodDepthChange()
		defer unsubscribe()

		select {
		case _, ok := <-changeC:
			if !ok {
				t.Error("closed signal channel")
			}
			t.Error("signal received")
		case <-time.After(1 * time.Second):
			// all fine
		}
	})
}

// TestCapabilitiesIndex checks that capability indices contains only the peers that have the filters' capability bits set
func TestCapabilityIndex(t *testing.T) {
	kp := NewKadParams()
	addr := RandomBzzAddr()
	base := addr.OAddr
	k := NewKademlia(base, kp)

	// "more" matches "more" only
	testMoreCapability := capability.NewCapability(42, 3)
	testMoreCapability.Set(0)
	testMoreCapability.Set(2)
	k.RegisterCapabilityIndex("more", *testMoreCapability)

	// "less" matches "more" and "less"
	testLessCapability := capability.NewCapability(42, 3)
	testLessCapability.Set(2)
	k.RegisterCapabilityIndex("less", *testLessCapability)

	// "none" matches neither "more" nor "less"
	testNoneCapability := capability.NewCapability(42, 3)
	testNoneCapability.Set(1)
	k.RegisterCapabilityIndex("none", *testNoneCapability)

	// "other" is a different capability array
	testOtherCapability := capability.NewCapability(666, 3)
	testOtherCapability.Set(0)
	testOtherCapability.Set(2)
	k.RegisterCapabilityIndex("other", *testOtherCapability)

	moreAddr := RandomBzzAddr()
	moreAddr.Capabilities.Add(testMoreCapability)

	lessAddr := RandomBzzAddr()
	lessAddr.Capabilities.Add(testLessCapability)

	otherAddr := RandomBzzAddr()
	otherAddr.Capabilities.Add(testOtherCapability)

	// all includes two different capability arrays
	allAddr := RandomBzzAddr()
	allAddr.Capabilities.Add(testOtherCapability)
	allAddr.Capabilities.Add(testMoreCapability)

	// proceed to check the matches, first for the "known peers" pot
	// first adding the different peer configurations to the kademlia
	k.Register(moreAddr, lessAddr, otherAddr, allAddr)

	// Call without filter should still return all peers
	var c int
	k.EachAddr(base, 255, func(_ *BzzAddr, _ int) bool {
		c++
		return true
	})
	if c != 4 {
		t.Fatalf("EachAddr expected 3 peers, got %d", c)
	}

	// Matches "more", "all"
	c = 0
	k.EachAddrFiltered(base, "more", 255, func(a *BzzAddr, _ int) bool {
		c++
		cp := a.Capabilities.Get(42)
		if !cp.Match(testMoreCapability) {
			t.Fatalf("EachAddrFiltered 'more' capability mismatch, expected %v, got %v", testMoreCapability, cp)
		}
		return true
	})
	if c != 2 {
		t.Fatalf("EachAddrFiltered 'full' expected 2 peer, got %d", c)
	}

	// Matches "more", "less", "all"
	c = 0
	k.EachAddrFiltered(base, "less", 255, func(a *BzzAddr, _ int) bool {
		c++
		return true
	})
	if c != 3 {
		t.Fatalf("EachAddrFiltered 'less' expected 2 peers, got %d", c)
	}

	// No matches
	c = 0
	k.EachAddrFiltered(base, "none", 255, func(a *BzzAddr, _ int) bool {
		c++
		return true
	})
	if c != 0 {
		t.Fatalf("EachAddrFiltered 'none' expected 0 peers, got %d", c)
	}

	// Matches "other", "all"
	// Also checks that "all" has both capabilities
	c = 0
	k.EachAddrFiltered(base, "other", 255, func(a *BzzAddr, _ int) bool {
		c++
		cp := a.Capabilities.Get(666)
		if !cp.Match(testOtherCapability) {
			t.Fatalf("EachAddrFiltered 'other' capability mismatch, expected %v, got %v", testOtherCapability, cp)
		}
		cp = a.Capabilities.Get(42)
		if cp != nil {
			c++
		}
		return true
	})
	if c != 3 {
		t.Fatalf("EachAddrFiltered 'other' expected 3 capability matches, got %d", c)
	}

	// Now check the connection pot index
	// We add the "all" and "less" peers
	allBzzPeer := &BzzPeer{
		BzzAddr: allAddr,
	}
	allPeer := NewPeer(allBzzPeer, k)
	k.On(allPeer)

	lessBzzPeer := &BzzPeer{
		BzzAddr: lessAddr,
	}
	lessPeer := NewPeer(lessBzzPeer, k)
	k.On(lessPeer)

	// Call without filter should return the single connected peer
	c = 0
	k.EachConn(base, 255, func(_ *Peer, _ int) bool {
		c++
		return true
	})
	if c != 2 {
		t.Fatalf("EachConn expected 2 peers, got %d", c)
	}

	// Check that the "all" peer exists in the indices for both capability arrays
	// first the "other" index ...
	c = 0
	k.EachConnFiltered(base, "other", 255, func(p *Peer, _ int) bool {
		c++
		cp := p.Capabilities.Get(666)
		if !cp.Match(testOtherCapability) {
			t.Fatalf("EachConnFiltered 'other' missing capability %v", testOtherCapability)
		}
		cp = p.Capabilities.Get(42)
		if !cp.Match(testMoreCapability) {
			t.Fatalf("EachConnFiltered 'other' missing capability %v", testMoreCapability)
		}
		return true
	})
	if c != 1 {
		t.Fatalf("EachConnFiltered 'other' expected 1 peer, got %d", c)
	}

	// ...then the "more" index
	c = 0
	k.EachConnFiltered(base, "more", 255, func(p *Peer, _ int) bool {
		c++
		cp := p.Capabilities.Get(666)
		if !cp.Match(testOtherCapability) {
			t.Fatalf("EachConnFiltered 'more' missing capability %v", testOtherCapability)
		}
		cp = p.Capabilities.Get(42)
		if !cp.Match(testMoreCapability) {
			t.Fatalf("EachConnFiltered 'more' missing capability %v", testMoreCapability)
		}
		return true
	})
	if c != 1 {
		t.Fatalf("EachConnFiltered 'more' expected 1 peer, got %d", c)
	}

	// Disconnect the "all" peer
	k.Off(allPeer)

	// Check that the "all" is now removed from connections
	c = 0
	k.EachConnFiltered(base, "other", 255, func(_ *Peer, _ int) bool {
		c++
		return true
	})
	if c != 0 {
		t.Fatalf("EachConnFiltered 'other' expected 0 peers, got %d", c)
	}

	// Check that there is still an "other" peer among known peers
	// (the two matched peers will be "all" and "other")
	c = 0
	k.EachAddrFiltered(base, "other", 255, func(_ *BzzAddr, _ int) bool {
		c++
		return true
	})
	if c != 2 {
		t.Fatalf("EachAddrFiltered 'other' expected 2 peers, got %d", c)
	}

	// Check that the "less" peer is still registered as connected
	c = 0
	k.EachConnFiltered(base, "less", 255, func(p *Peer, _ int) bool {
		c++
		cp := p.Capabilities.Get(42)
		if !cp.Match(testLessCapability) {
			t.Fatalf("EachConnFiltered 'less' missing capability %v", testLessCapability)
		}
		return true
	})
	if c != 1 {
		t.Fatalf("EachConnFiltered 'less' expected 1 peer, got %d", c)
	}

	// Remove "less" from both connection and known peers (pruning) list
	// TODO replace with the "prune" method when one is implemented
	k.removeFromCapabilityIndex(lessPeer, false)

	// Check that the "less" peer is no longer registered as connected
	c = 0
	k.EachConnFiltered(base, "less", 255, func(p *Peer, _ int) bool {
		c++
		return true
	})
	if c != 0 {
		t.Fatalf("EachConnFiltered 'less' expected 0 peers, got %d", c)
	}

	// check that the "less" peer is not known anymore
	// (the two matched peers will be "all" and "more")
	c = 0
	k.EachAddrFiltered(base, "less", 255, func(p *BzzAddr, _ int) bool {
		c++
		cp := p.Capabilities.Get(42)
		if !cp.Match(testMoreCapability) {
			t.Fatalf("EachConnFiltered 'less' should now return only capability 'more': %v", testMoreCapability)
		}
		return true
	})
	if c != 2 {
		t.Fatalf("EachAddrFiltered 'less' expected 2 peer, got %d", c)
	}

	// Remove "all" from known peers list (pruning only)
	// TODO replace with the "prune" method when one is implemented
	k.removeFromCapabilityIndex(allPeer, false)

	// check that the "less" peer is not known anymore
	// (the two matched peers will be "all" and "more")
	c = 0
	k.EachAddrFiltered(base, "more", 255, func(p *BzzAddr, _ int) bool {
		c++
		cp := p.Capabilities.Get(666)
		if cp != nil {
			t.Fatalf("EachAddrFiltered 'more' should not contain a peer with capability %v", testOtherCapability)
		}
		return true
	})
	if c != 1 {
		t.Fatalf("EachAddrFiltered 'more' expected 1 peer, got %d", c)
	}
}
