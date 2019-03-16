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

package simulation

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/network"
)

func TestWaitTillHealthy(t *testing.T) {
	t.Skip("WaitTillHealthy depends on discovery, which relies on a reliable SuggestPeer, which is not reliable")

	sim := New(map[string]ServiceFunc{
		"bzz": func(ctx *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
			addr := network.NewAddr(ctx.Config.Node())
			hp := network.NewHiveParams()
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			kad := network.NewKademlia(addr.Over(), network.NewKadParams())
			// store kademlia in node's bucket under BucketKeyKademlia
			// so that it can be found by WaitTillHealthy method.
			b.Store(BucketKeyKademlia, kad)
			return network.NewBzz(config, kad, nil, nil, nil), nil, nil
		},
	})
	defer sim.Close()

	_, err := sim.AddNodesAndConnectRing(10)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	ill, err := sim.WaitTillHealthy(ctx)
	if err != nil {
		for id, kad := range ill {
			t.Log("Node", id)
			t.Log(kad.String())
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestWaitTillSnapshotRecreated tests that we indeed have a network
// configuration specified in the snapshot file, after we wait for it.
//
// First we create a first simulation
// Run it as nodes connected in a ring
// Wait until the network is healthy
// Then we create a snapshot
// With this snapshot we create a new simulation
// Call WaitTillSnapshotRecreated() function and wait until it returns
// Iterate the nodes and check if all the connections are successfully recreated
func TestWaitTillSnapshotRecreated(t *testing.T) {
	var err error
	sim := New(createSimServiceMap(true))
	_, err = sim.AddNodesAndConnectRing(16)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err = sim.WaitTillHealthy(ctx)
	if err != nil {
		t.Fatal(err)
	}

	originalConnections := sim.getActualConnections()
	snap, err := sim.Net.Snapshot()
	sim.Close()
	if err != nil {
		t.Fatal(err)
	}

	controlSim := New(createSimServiceMap(false))
	defer controlSim.Close()
	err = controlSim.Net.Load(snap)
	if err != nil {
		t.Fatal(err)
	}
	err = controlSim.WaitTillSnapshotRecreated(ctx, *snap)
	if err != nil {
		t.Fatal(err)
	}
	controlConnections := controlSim.getActualConnections()

	for _, c := range originalConnections {
		if !exist(controlConnections, c) {
			t.Fatal("connection was not recreated")
		}
	}
}

// exist returns true if val is found in arr
func exist(arr []uint64, val uint64) bool {
	for _, c := range arr {
		if c == val {
			return true
		}
	}
	return false
}

func TestRemoveDuplicatesAndSingletons(t *testing.T) {
	singletons := []uint64{
		0x3c127c6f6cb026b0,
		0x0f45190d72e71fc5,
		0xb0184c02449e0bb6,
		0xa85c7b84239c54d3,
		0xe3b0c44298fc1c14,
		0x9afbf4c8996fb924,
		0x27ae41e4649b934c,
		0xa495991b7852b855,
	}

	doubles := []uint64{
		0x1b879f878de7fc7a,
		0xc6791470521bdab4,
		0xdd34b0ee39bbccc6,
		0x4d904fbf0f31da10,
		0x6403c2560432c8f8,
		0x18954e33cf3ad847,
		0x90db00e98dc7a8a6,
		0x92886b0dfcc1809b,
	}

	var arr []uint64
	arr = append(arr, doubles...)
	arr = append(arr, singletons...)
	arr = append(arr, doubles...)
	arr = removeDuplicatesAndSingletons(arr)

	for _, i := range singletons {
		if exist(arr, i) {
			t.Fatalf("singleton not removed: %d", i)
		}
	}

	for _, i := range doubles {
		if !exist(arr, i) {
			t.Fatalf("wrong value removed: %d", i)
		}
	}

	for j := 0; j < len(doubles); j++ {
		v := doubles[j] + singletons[j]
		if exist(arr, v) {
			t.Fatalf("non-existing value found, index: %d", j)
		}
	}
}

func TestIsAllDeployed(t *testing.T) {
	a := []uint64{
		0x3c127c6f6cb026b0,
		0x0f45190d72e71fc5,
		0xb0184c02449e0bb6,
		0xa85c7b84239c54d3,
		0xe3b0c44298fc1c14,
		0x9afbf4c8996fb924,
		0x27ae41e4649b934c,
		0xa495991b7852b855,
	}

	b := []uint64{
		0x1b879f878de7fc7a,
		0xc6791470521bdab4,
		0xdd34b0ee39bbccc6,
		0x4d904fbf0f31da10,
		0x6403c2560432c8f8,
		0x18954e33cf3ad847,
		0x90db00e98dc7a8a6,
		0x92886b0dfcc1809b,
	}

	var c []uint64
	c = append(c, a...)
	c = append(c, b...)

	if !isAllDeployed(a, c) {
		t.Fatal("isAllDeployed failed")
	}

	if !isAllDeployed(b, c) {
		t.Fatal("isAllDeployed failed")
	}

	if isAllDeployed(c, a) {
		t.Fatal("isAllDeployed failed: false positive")
	}

	if isAllDeployed(c, b) {
		t.Fatal("isAllDeployed failed: false positive")
	}

	c = c[2:]

	if isAllDeployed(a, c) {
		t.Fatal("isAllDeployed failed: false positive")
	}

	if !isAllDeployed(b, c) {
		t.Fatal("isAllDeployed failed")
	}
}
