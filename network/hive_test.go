// Copyright 2016 The go-ethereum Authors
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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/log"
	p2ptest "github.com/ethersphere/swarm/p2p/testing"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/state"
)

func newHiveTester(params *HiveParams, n int, store state.Store) (*bzzTester, *Hive, error) {
	// setup
	prvkey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	addr := PrivateKeyToBzzKey(prvkey)
	to := NewKademlia(addr, NewKadParams())
	pp := NewHive(params, to, store) // hive

	bt, err := newBzzBaseTester(n, prvkey, DiscoverySpec, pp.Run)
	if err != nil {
		return nil, nil, err
	}
	return bt, pp, nil
}

// TestRegisterAndConnect verifies that the protocol runs successfully
// and that the peer connection exists afterwards
func TestRegisterAndConnect(t *testing.T) {
	params := NewHiveParams()
	s, pp, err := newHiveTester(params, 1, nil)
	if err != nil {
		t.Fatal(err)
	}

	node := s.Nodes[0]
	raddr := NewBzzAddrFromEnode(node)
	pp.Register(raddr)

	// start the hive
	err = pp.Start(s.Server)
	if err != nil {
		t.Fatal(err)
	}
	defer pp.Stop()

	// both hive connect and disconect check have time delays
	// therefore we need to verify that peer is connected
	// so that we are sure that the disconnect timeout doesn't complete
	// before the hive connect method is run at least once
	timeout := time.After(time.Second)
	for {
		select {
		case <-timeout:
			t.Fatalf("expected connection")
		default:
		}
		i := 0
		pp.Kademlia.EachConn(nil, 256, func(addr *Peer, po int) bool {
			i++
			return true
		})
		if i > 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	// check that the connection actually exists
	// the timeout error means no disconnection events
	// were received within the a certain timeout
	err = s.TestDisconnected(&p2ptest.Disconnect{
		Peer:  s.Nodes[0].ID(),
		Error: nil,
	})

	if err == nil || err.Error() != "timed out waiting for peers to disconnect" {
		t.Fatalf("expected no disconnection event")
	}
}

// TestHiveStatePersistence creates a protocol simulation with n peers for a node
// After protocols complete, the node is shut down and the state is stored.
// Another simulation is created, where 0 nodes are created, but where the stored state is passed
// The test succeeds if all the peers from the stored state are known after the protocols of the
// second simulation have completed
//
// Actual connectivity is not in scope for this test, as the peers loaded from state are not known to
// the simulation; the test only verifies that the peers are known to the node
func TestHiveStatePersistence(t *testing.T) {
	dir, err := ioutil.TempDir("", "hive_test_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	const peersCount = 5

	startHive := func(t *testing.T, dir string) (h *Hive, cleanupFunc func()) {
		store, err := state.NewDBStore(dir)
		if err != nil {
			t.Fatal(err)
		}

		params := NewHiveParams()
		params.Discovery = false

		prvkey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatal(err)
		}

		h = NewHive(params, NewKademlia(PrivateKeyToBzzKey(prvkey), NewKadParams()), store)
		s := p2ptest.NewProtocolTester(prvkey, 0, func(p *p2p.Peer, rw p2p.MsgReadWriter) error { return nil })

		if err := h.Start(s.Server); err != nil {
			t.Fatal(err)
		}

		cleanupFunc = func() {
			err := h.Stop()
			if err != nil {
				t.Fatal(err)
			}

			s.Stop()
		}
		return h, cleanupFunc
	}

	h1, cleanup1 := startHive(t, dir)
	peers := make(map[string]bool)
	for i := 0; i < peersCount; i++ {
		raddr := RandomBzzAddr()
		h1.Register(raddr)
		peers[raddr.String()] = true
	}
	cleanup1()

	// start the hive and check that we know of all expected peers
	h2, cleanup2 := startHive(t, dir)
	cleanup2()

	i := 0
	h2.Kademlia.EachAddr(nil, 256, func(addr *BzzAddr, po int) bool {
		delete(peers, addr.String())
		i++
		return true
	})
	if i != peersCount {
		t.Fatalf("invalid number of entries: got %v, want %v", i, peersCount)
	}
	if len(peers) != 0 {
		t.Fatalf("%d peers left over: %v", len(peers), peers)
	}
}

// TestHiveStateConnections connect the node to some peers and then after cleanup/save in store those peers
// are retrieved and used as suggested peer initially.
func TestHiveStateConnections(t *testing.T) {
	dir, err := ioutil.TempDir("", "hive_test_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	const peersCount = 5

	nodeIdToBzzAddr := make(map[string]*BzzAddr)
	addedChan := make(chan struct{}, 5)
	startHive := func(t *testing.T, dir string) (h *Hive, cleanupFunc func()) {
		store, err := state.NewDBStore(dir)
		if err != nil {
			t.Fatal(err)
		}

		params := NewHiveParams()
		params.Discovery = false

		prvkey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatal(err)
		}

		h = NewHive(params, NewKademlia(PrivateKeyToBzzKey(prvkey), NewKadParams()), store)
		s := p2ptest.NewProtocolTester(prvkey, 0, func(p *p2p.Peer, rw p2p.MsgReadWriter) error { return nil })

		if err := h.Start(s.Server); err != nil {
			t.Fatal(err)
		}
		//Close ticker to avoid interference with initial peer suggestion
		h.ticker.Stop()
		//Overwrite addPeer so the Node is added as a peer automatically.
		// The related Overlay address is retrieved from nodeIdToBzzAddr where it has been saved before
		h.addPeer = func(node *enode.Node) {
			bzzAddr := nodeIdToBzzAddr[encodeId(node.ID())]
			if bzzAddr == nil {
				t.Fatalf("Enode [%v] not found in saved peers!", encodeId(node.ID()))
			}
			bzzPeer := newConnPeerLocal(bzzAddr.Address(), h.Kademlia)
			h.On(bzzPeer)
			addedChan <- struct{}{}
		}

		cleanupFunc = func() {
			err := h.Stop()
			if err != nil {
				t.Fatal(err)
			}

			s.Stop()
		}
		return h, cleanupFunc
	}

	h1, cleanup1 := startHive(t, dir)
	peers := make(map[string]bool)
	for i := 0; i < peersCount; i++ {
		raddr := RandomBzzAddr()
		h1.Register(raddr)
		peers[raddr.String()] = true
	}
	const initialPeers = 5
	for i := 0; i < initialPeers; i++ {
		suggestedPeer, _, _ := h1.SuggestPeer()
		if suggestedPeer != nil {
			testAddPeer(suggestedPeer, h1, nodeIdToBzzAddr)
		}

	}
	numConns := h1.conns.Size()
	connAddresses := make(map[string]string)
	h1.EachConn(h1.base, 255, func(peer *Peer, i int) bool {
		key := hexutil.Encode(peer.Address())
		connAddresses[key] = key
		return true
	})
	log.Warn("After 5 suggestions", "numConns", numConns)
	cleanup1()

	// start the hive and check that we suggest previous connected peers
	h2, _ := startHive(t, dir)
	// there should be at some point 5 conns
	connsAfterLoading := 0
	iterations := 0
	connsAfterLoading = h2.conns.Size()
	for connsAfterLoading != numConns && iterations < 5 {
		select {
		case <-addedChan:
			connsAfterLoading = h2.conns.Size()
		case <-time.After(1 * time.Second):
			iterations++
		}
		log.Trace("Iteration waiting for initial connections", "numConns", connsAfterLoading, "iterations", iterations)
	}
	if connsAfterLoading != numConns {
		t.Errorf("Expected 5 peer connecteds from previous execution but got %v", connsAfterLoading)
	}
	h2.EachConn(h2.base, 255, func(peer *Peer, i int) bool {
		key := hexutil.Encode(peer.Address())
		if connAddresses[key] != key {
			t.Errorf("Expected address %v to be in connections as it was a previous peer connected", key)
		} else {
			log.Warn("Previous peer connected again", "addr", key)
		}
		return true
	})
}

// Create a Peer with the suggested address and store the relationshsip enode -> BzzAddr for later retrieval
func testAddPeer(suggestedPeer *BzzAddr, h1 *Hive, nodeIdToBzzAddr map[string]*BzzAddr) {
	byteAddresses := suggestedPeer.Address()
	bzzPeer := newConnPeerLocal(byteAddresses, h1.Kademlia)
	nodeIdToBzzAddr[encodeId(bzzPeer.ID())] = bzzPeer.BzzAddr
	bzzPeer.kad = h1.Kademlia
	h1.On(bzzPeer)
}

func encodeId(id enode.ID) string {
	addr := id[:]
	return hexutil.Encode(addr)
}

// We create a test Peer with underlay address to localhost and using overlay address provided
func newConnPeerLocal(addr []byte, kademlia *Kademlia) *Peer {
	hash := [common.HashLength]byte{}
	copy(hash[:], addr)
	potAddress := pot.Address(hash)
	peer := newDiscPeer(potAddress)
	peer.kad = kademlia
	return peer
}
