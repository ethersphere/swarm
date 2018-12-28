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
	base := "00000000"

	for _, v := range []struct {
		name    string
		ons     []string
		offs    []string
		expAddr []string
	}{
		{
			name:    "example test",
			ons:     []string{"00100000", "00110000"},
			offs:    []string{"00110000"},
			expAddr: []string{},
		},
	} {
		t.Run(v.name, func(t *testing.T) {
			params := NewHiveParams()
			k := newTestKademlia(base)
			h := NewHive(params, k, nil) // hive

			for _, on := range v.ons {
				piu := newTestKadPeer(k, on, false)
				h.registerPeer(piu)
			}
			for _, off := range v.offs {
				Off(k, off)
			}

			err := testSuggestPeerWTF(h, v.expAddr)
			if err != nil {
				t.Fatalf("%v", err.Error())
			}
		})
	}
}
func testSuggestPeerWTF(h *Hive, expAddr []string) error {
	peers := h.suggestPeers()
	for i, v := range peers {
		if expAddr[i] != binStr(v.BzzAddr) {
			return fmt.Errorf("incorrect peer address suggested. expected %v, got %v", expAddr[i], binStr(v.BzzAddr))
		}
	}
	return nil
}
