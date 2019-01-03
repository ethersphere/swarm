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
	"fmt"
	"io/ioutil"
	golog "log"
	"os"
	"testing"
	"time"

	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
	"github.com/ethereum/go-ethereum/swarm/state"
)

func newHiveTester(t *testing.T, params *HiveParams, n int, store state.Store) (*bzzTester, *Hive) {
	// setup
	addr := RandomAddr() // tested peers peer address
	to := NewKademlia(addr.OAddr, NewKadParams())
	pp := NewHive(params, to, store) // hive

	return newBzzBaseTester(t, n, addr, DiscoverySpec, pp.Run), pp
}

func TestRegisterAndConnect(t *testing.T) {
	params := NewHiveParams()
	s, pp := newHiveTester(t, params, 1, nil)

	node := s.Nodes[0]
	raddr := NewAddr(node)
	pp.Register(raddr)

	// start the hive and wait for the connection
	err := pp.Start(s.Server)
	if err != nil {
		t.Fatal(err)
	}
	defer pp.Stop()
	// retrieve and broadcast
	err = s.TestDisconnected(&p2ptest.Disconnect{
		Peer:  s.Nodes[0].ID(),
		Error: nil,
	})

	if err == nil || err.Error() != "timed out waiting for peers to disconnect" {
		t.Fatalf("expected peer to connect")
	}
}

func TestHiveStatePersistance(t *testing.T) {
	golog.SetOutput(os.Stdout)

	dir, err := ioutil.TempDir("", "hive_test_store")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	store, err := state.NewDBStore(dir) //start the hive with an empty dbstore
	if err != nil {
		t.Fatal(err)
	}

	params := NewHiveParams()
	s, pp := newHiveTester(t, params, 5, store)

	peers := make(map[string]bool)
	for _, node := range s.Nodes {
		raddr := NewAddr(node)
		pp.Register(raddr)
		peers[raddr.String()] = true
	}

	// start the hive and wait for the connection
	err = pp.Start(s.Server)
	if err != nil {
		t.Fatal(err)
	}
	pp.Stop()
	store.Close()

	persistedStore, err := state.NewDBStore(dir) //start the hive with an empty dbstore
	if err != nil {
		t.Fatal(err)
	}

	s1, pp := newHiveTester(t, params, 1, persistedStore)

	//start the hive and wait for the connection

	pp.Start(s1.Server)
	i := 0
	pp.Kademlia.EachAddr(nil, 256, func(addr *BzzAddr, po int, nn bool) bool {
		delete(peers, addr.String())
		i++
		return true
	})
	if i != 5 {
		t.Errorf("invalid number of entries: got %v, want %v", i, 5)
	}
	if len(peers) != 0 {
		t.Fatalf("%d peers left over: %v", len(peers), peers)
	}
}

// TestSuggestPeer tests for the different moving parts of the kademlia SuggestPeer method
// We have to take several things in accoutn when testing this functionality:
//	1. Proximity order of the peers (duh)
//	2. Callability of peers (i.e. we tried to connect to a peer recently but couldn't so we have to make sure
//		 the exponential backoff kicks in (we won't test for its correctness here though)
//	3.
func TestSuggestPeerWTF(t *testing.T) {
	base := "00000001"
	hiveRetryInterval := int64(10 * time.Millisecond)

	for _, v := range []struct {
		name     string
		ons      []string
		offs     []string
		expAddr  []string
		expDepth int
		skip     bool
		wait     time.Duration
	}{
		{
			name:     "no peers to suggest (all ON)",
			ons:      []string{"00000010", "00010000", "00000011", "00000111"},
			offs:     []string{},
			expAddr:  []string{},
			expDepth: 0,
			skip:     true,
		},
		{
			name:     "no peers (too early to try)",
			ons:      []string{"00100000", "00110000", "00111000", "00011011", "00010101"},
			offs:     []string{"00110000", "00111000"},
			expAddr:  []string{},
			expDepth: 0,
			wait:     2 * time.Millisecond, //retry interval is set to 10ms, try before
			skip:     true,
		},
		{
			name:     "suggest deeper then shallower",
			ons:      []string{"00100000", "00110000", "00111000", "00011011", "00010101"},
			offs:     []string{"00110000", "00111000"},
			expAddr:  []string{"00111000", "00110000"},
			expDepth: 0,
			wait:     20 * time.Millisecond,
			skip:     true,
		},
		{
			name:     "shallow bin ON (depth=1, bin 1 empty), suggest a peer with po > depth",
			ons:      []string{"10000000", "00100000", "00100001", "00100011"},
			offs:     []string{"00100011"},
			expAddr:  []string{"00100011"},
			expDepth: 1,
			wait:     20 * time.Millisecond,
			skip:     true,
		},
		{
			name:     "shallow bin ON (depth=3, bin 2 peer known but not connected), suggest a peer with po < depth",
			ons:      []string{"10000000", "01000000", "00100000", "00100001", "00001111", "00001011"},
			offs:     []string{"00100001"},
			expAddr:  []string{"00100001"},
			expDepth: 3,
			wait:     20 * time.Millisecond,
		},
	} {
		if v.skip {
			continue
		}
		t.Run(v.name, func(t *testing.T) {
			params := NewHiveParams()
			params.RetryInterval = hiveRetryInterval
			k := newTestKademlia(base)
			h := NewHive(params, k, nil) // hive

			for _, on := range v.ons {
				piu := newTestKadPeer(k, on, false)
				h.registerPeer(piu)
			}
			for _, off := range v.offs {
				Off(k, off)
			}

			if h.Kademlia.NeighbourhoodDepth() != v.expDepth {
				t.Fatalf("wrong neighbourhood depth. got: %d, want: %d", k.NeighbourhoodDepth(), v.expDepth)
			}

			if v.wait > 0 {
				time.Sleep(v.wait)
			}

			err := testSuggestPeerWTF(t, h, v.expAddr)
			if err != nil {
				t.Fatalf("%v", err.Error())
			}
		})
	}
}
func testSuggestPeerWTF(t *testing.T, h *Hive, expAddr []string) error {
	peers := h.suggestPeers()
	if len(peers) != len(expAddr) {
		t.Fatalf("expected %d suggested peers but got %d instead", len(expAddr), len(peers))
	}
	for i, v := range peers {
		if expAddr[i] != binStr(v.BzzAddr) {
			return fmt.Errorf("incorrect peer address suggested. expected %v, got %v", expAddr[i], binStr(v.BzzAddr))
		}
	}
	return nil
}
