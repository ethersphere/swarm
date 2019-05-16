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

package stream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/network/simulation"
	"github.com/ethereum/go-ethereum/swarm/state"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/testutil"
)

const dataChunkCount = 1000

func TestSyncerSimulation(t *testing.T) {
	testSyncBetweenNodes(t, 2, dataChunkCount, true, 1)
	// This test uses much more memory when running with
	// race detector. Allow it to finish successfully by
	// reducing its scope, and still check for data races
	// with the smallest number of nodes.
	/*if !testutil.RaceEnabled {
		testSyncBetweenNodes(t, 4, dataChunkCount, true, 1)
		testSyncBetweenNodes(t, 8, dataChunkCount, true, 1)
		testSyncBetweenNodes(t, 16, dataChunkCount, true, 1)
	}*/
}

func testSyncBetweenNodes(t *testing.T, nodes, chunkCount int, skipCheck bool, po uint8) {

	sim := simulation.New(map[string]simulation.ServiceFunc{
		"streamer": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			addr := network.NewAddr(ctx.Config.Node())
			//hack to put addresses in same space
			addr.OAddr[0] = byte(0)

			netStore, delivery, clean, err := newNetStoreAndDeliveryWithBzzAddr(ctx, bucket, addr)
			if err != nil {
				return nil, nil, err
			}

			var dir string
			var store *state.DBStore
			if testutil.RaceEnabled {
				// Use on-disk DBStore to reduce memory consumption in race tests.
				dir, err = ioutil.TempDir("", "swarm-stream-")
				if err != nil {
					return nil, nil, err
				}
				store, err = state.NewDBStore(dir)
				if err != nil {
					return nil, nil, err
				}
			} else {
				store = state.NewInmemoryStore()
			}

			r := NewRegistry(addr.ID(), delivery, netStore, store, &RegistryOptions{
				Syncing:         SyncingAutoSubscribe,
				SyncUpdateDelay: 50 * time.Millisecond,
				SkipCheck:       skipCheck,
			}, nil)

			cleanup = func() {
				r.Close()
				clean()
				if dir != "" {
					os.RemoveAll(dir)
				}
			}

			return r, cleanup, nil
		},
	})
	defer sim.Close()

	// create context for simulation run
	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel should come before defer simulation teardown
	defer cancel()

	_, err := sim.AddNodesAndConnectChain(nodes)
	if err != nil {
		t.Fatal(err)
	}
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		nodeIDs := sim.UpNodeIDs()

		nodeIndex := make(map[enode.ID]int)
		for i, id := range nodeIDs {
			nodeIndex[id] = i
		}

		disconnected := watchDisconnections(ctx, sim)
		defer func() {
			if err != nil && disconnected.bool() {
				err = errors.New("disconnect events received")
			}
		}()

		// each node Subscribes to each other's swarmChunkServerStreamName
		item, ok := sim.NodeItem(nodeIDs[0], bucketKeyFileStore)
		if !ok {
			return fmt.Errorf("No filestore")
		}
		fileStore := item.(*storage.FileStore)
		size := chunkCount * chunkSize
		_, wait1, err := fileStore.Store(ctx, testutil.RandomReader(0, size), int64(size), false)
		if err != nil {
			return fmt.Errorf("fileStore.Store: %v", err)
		}
		// here we distribute chunks of a random file into stores 1...nodes
		// collect hashes in po 1 bin for each node
		_, wait2, err := fileStore.Store(ctx, testutil.RandomReader(10, size), int64(size), false)
		if err != nil {
			return fmt.Errorf("fileStore.Store: %v", err)
		}
		wait1(ctx)
		wait2(ctx)
		time.Sleep(2 * time.Second)

		log.Warn("uploader node", "enode", nodeIDs[0])
		item, ok = sim.NodeItem(nodeIDs[0], bucketKeyStore)
		if !ok {
			return fmt.Errorf("No DB")
		}
		store := item.(chunk.Store)
		until, err := store.LastPullSubscriptionBinID(po)
		if err != nil {
			return err
		}

		for idx, node := range nodeIDs {
			if nodeIDs[idx] == nodeIDs[0] {
				continue
			}

			i := nodeIndex[node]

			log.Warn("compare to", "enode", nodeIDs[idx])
			item, ok = sim.NodeItem(nodeIDs[idx], bucketKeyStore)
			if !ok {
				return fmt.Errorf("No DB")
			}
			db := item.(chunk.Store)
			shouldUntil, err := db.LastPullSubscriptionBinID(po)
			if err != nil {
				t.Fatal(err)
			}
			log.Warn("last pull subscription bin id", "shouldUntil", shouldUntil, "until", until, "po", po)
			if shouldUntil != until {
				t.Fatalf("did not get correct bin index from peer. got %d want %d", shouldUntil, until)
			}

			log.Warn("sync check", "node", node, "index", i, "bin", po)
		}
		return nil
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

//TestSameVersionID just checks that if the version is not changed,
//then streamer peers see each other
func TestSameVersionID(t *testing.T) {
	//test version ID
	v := uint(1)
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"streamer": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			addr, netStore, delivery, clean, err := newNetStoreAndDelivery(ctx, bucket)
			if err != nil {
				return nil, nil, err
			}

			r := NewRegistry(addr.ID(), delivery, netStore, state.NewInmemoryStore(), &RegistryOptions{
				Syncing: SyncingAutoSubscribe,
			}, nil)
			bucket.Store(bucketKeyRegistry, r)

			//assign to each node the same version ID
			r.spec.Version = v

			cleanup = func() {
				r.Close()
				clean()
			}

			return r, cleanup, nil
		},
	})
	defer sim.Close()

	//connect just two nodes
	log.Info("Adding nodes to simulation")
	_, err := sim.AddNodesAndConnectChain(2)
	if err != nil {
		t.Fatal(err)
	}

	log.Info("Starting simulation")
	ctx := context.Background()
	//make sure they have time to connect
	time.Sleep(200 * time.Millisecond)
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		//get the pivot node's filestore
		nodes := sim.UpNodeIDs()

		item, ok := sim.NodeItem(nodes[0], bucketKeyRegistry)
		if !ok {
			return fmt.Errorf("No filestore")
		}
		registry := item.(*Registry)

		//the peers should connect, thus getting the peer should not return nil
		if registry.getPeer(nodes[1]) == nil {
			return errors.New("Expected the peer to not be nil, but it is")
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	log.Info("Simulation ended")
}

//TestDifferentVersionID proves that if the streamer protocol version doesn't match,
//then the peers are not connected at streamer level
func TestDifferentVersionID(t *testing.T) {
	//create a variable to hold the version ID
	v := uint(0)
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"streamer": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			addr, netStore, delivery, clean, err := newNetStoreAndDelivery(ctx, bucket)
			if err != nil {
				return nil, nil, err
			}

			r := NewRegistry(addr.ID(), delivery, netStore, state.NewInmemoryStore(), &RegistryOptions{
				Syncing: SyncingAutoSubscribe,
			}, nil)
			bucket.Store(bucketKeyRegistry, r)

			//increase the version ID for each node
			v++
			r.spec.Version = v

			cleanup = func() {
				r.Close()
				clean()
			}

			return r, cleanup, nil
		},
	})
	defer sim.Close()

	//connect the nodes
	log.Info("Adding nodes to simulation")
	_, err := sim.AddNodesAndConnectChain(2)
	if err != nil {
		t.Fatal(err)
	}

	log.Info("Starting simulation")
	ctx := context.Background()
	//make sure they have time to connect
	time.Sleep(200 * time.Millisecond)
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		//get the pivot node's filestore
		nodes := sim.UpNodeIDs()

		item, ok := sim.NodeItem(nodes[0], bucketKeyRegistry)
		if !ok {
			return fmt.Errorf("No filestore")
		}
		registry := item.(*Registry)

		//getting the other peer should fail due to the different version numbers
		if registry.getPeer(nodes[1]) != nil {
			return errors.New("Expected the peer to be nil, but it is not")
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	log.Info("Simulation ended")

}

// Tests that when two nodes connect:
// 1. All subscriptions are created
// 2. All chunks are transferred from one node to another
func TestTwoNodesFullSync(t *testing.T) { //
	const chunkCount = 1000

	sim := simulation.New(map[string]simulation.ServiceFunc{
		"streamer": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			addr := network.NewAddr(ctx.Config.Node())
			//hack to put addresses in same space
			addr.OAddr[0] = byte(0) //wtf is this?

			netStore, delivery, clean, err := newNetStoreAndDeliveryWithBzzAddr(ctx, bucket, addr)
			if err != nil {
				return nil, nil, err
			}

			var dir string
			var store *state.DBStore
			if testutil.RaceEnabled {
				// Use on-disk DBStore to reduce memory consumption in race tests.
				dir, err = ioutil.TempDir("", "swarm-stream-")
				if err != nil {
					return nil, nil, err
				}
				store, err = state.NewDBStore(dir)
				if err != nil {
					return nil, nil, err
				}
			} else {
				store = state.NewInmemoryStore()
			}

			r := NewRegistry(addr.ID(), delivery, netStore, store, &RegistryOptions{
				Syncing:         SyncingAutoSubscribe,
				SyncUpdateDelay: 50 * time.Millisecond,
				SkipCheck:       true,
			}, nil)

			cleanup = func() {
				r.Close()
				clean()
				if dir != "" {
					os.RemoveAll(dir)
				}
			}

			return r, cleanup, nil
		},
	})
	defer sim.Close()

	// create context for simulation run
	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel should come before defer simulation teardown
	defer cancel()

	_, err := sim.AddNodesAndConnectChain(2)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		nodeIDs := sim.UpNodeIDs()

		nodeIndex := make(map[enode.ID]int)
		for i, id := range nodeIDs {
			nodeIndex[id] = i
		}

		disconnected := watchDisconnections(ctx, sim)
		defer func() {
			if err != nil && disconnected.bool() {
				err = errors.New("disconnect events received")
			}
		}()

		// each node Subscribes to each other's swarmChunkServerStreamName
		item, ok := sim.NodeItem(nodeIDs[0], bucketKeyFileStore)
		if !ok {
			return fmt.Errorf("No filestore")
		}
		fileStore := item.(*storage.FileStore)
		size := chunkCount * chunkSize

		_, wait1, err := fileStore.Store(ctx, testutil.RandomReader(0, size), int64(size), false)
		if err != nil {
			return fmt.Errorf("fileStore.Store: %v", err)
		}

		_, wait2, err := fileStore.Store(ctx, testutil.RandomReader(10, size), int64(size), false)
		if err != nil {
			return fmt.Errorf("fileStore.Store: %v", err)
		}

		wait1(ctx)
		wait2(ctx)
		time.Sleep(5 * time.Second)

		log.Warn("uploader node", "enode", nodeIDs[0])
		item, ok = sim.NodeItem(nodeIDs[0], bucketKeyStore)
		if !ok {
			return fmt.Errorf("No DB")
		}
		store := item.(chunk.Store)
		uploaderNodeBinIDs := make([]uint64, 17)

		log.Debug("checking pull subscription bin ids")
		for po := 0; po <= 16; po++ {
			until, err := store.LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				t.Fatal(err)
			}

			uploaderNodeBinIDs[po] = until
		}

		for idx, _ := range nodeIDs {
			if nodeIDs[idx] == nodeIDs[0] {
				continue
			}

			log.Warn("compare to", "enode", nodeIDs[idx])
			item, ok = sim.NodeItem(nodeIDs[idx], bucketKeyStore)
			if !ok {
				return fmt.Errorf("No DB")
			}
			db := item.(chunk.Store)

			time.Sleep(5 * time.Second)
			uploaderSum, otherSum := 0, 0
			for po, uploaderUntil := range uploaderNodeBinIDs {
				shouldUntil, err := db.LastPullSubscriptionBinID(uint8(po))
				if err != nil {
					t.Fatal(err)
				}
				otherSum += int(shouldUntil)
				uploaderSum += int(uploaderUntil)
			}
			if uploaderSum != otherSum {
				t.Fatalf("did not get correct bin index from peer. got %d want %d", uploaderSum, otherSum)
			}
		}
		return nil
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// connect simulation of X nodes in a star topology (min 8 nodes)
// get all chunk refs from the smoke test util
// let them sync
// iterate over all chunk refs
// for each chunk, check pos with all nodes, for the node with highest PO - check if that node has the chunk in its localstore (with a GET)
// exclusivity test (after we do the most prox inclusivity test, similar to the smoke test)
// ---
// uploader node 0
// node 1 -> 0 -> does not have chunks from 1 2 3 4 5 .. 16
//  - pick random (or any) chunk from bin 1 2 3 4 5 .. 16 from uploader node
//  - localstore get on node 1 (without caring about po of node 1 and the chunk)
//node 2 -> 1 -> does not have chunks from 0 2 3 4 5 .. 16
//node 3 -> 2 3 4 5 .. 16 -> does not have chunks from 0 1
/*

   //create rpc client
   client, err := node.Client()
   if err != nil {
       return fmt.Errorf("create node 1 rpc client fail: %v", err)
   }

   //ask it for subscriptions
   pstreams := make(map[string][]string)
   err = client.Call(&pstreams, "stream_getPeerServerSubscriptions")
   if err != nil {
       return fmt.Errorf("client call stream_getPeerSubscriptions: %v", err)
   }
*/
func TestStarNetworkSync(t *testing.T) { //
	const chunkCount = 1000

	sim := simulation.New(map[string]simulation.ServiceFunc{
		"streamer": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			addr := network.NewAddr(ctx.Config.Node())
			//hack to put addresses in same space
			//addr.OAddr[0] = byte(0) //wtf

			netStore, delivery, clean, err := newNetStoreAndDeliveryWithBzzAddr(ctx, bucket, addr)
			if err != nil {
				return nil, nil, err
			}

			var dir string
			var store *state.DBStore
			if testutil.RaceEnabled {
				// Use on-disk DBStore to reduce memory consumption in race tests.
				dir, err = ioutil.TempDir("", "swarm-stream-")
				if err != nil {
					return nil, nil, err
				}
				store, err = state.NewDBStore(dir)
				if err != nil {
					return nil, nil, err
				}
			} else {
				store = state.NewInmemoryStore()
			}

			r := NewRegistry(addr.ID(), delivery, netStore, store, &RegistryOptions{
				Syncing:         SyncingAutoSubscribe,
				SyncUpdateDelay: 50 * time.Millisecond,
				SkipCheck:       true,
			}, nil)

			cleanup = func() {
				r.Close()
				clean()
				if dir != "" {
					os.RemoveAll(dir)
				}
			}

			return r, cleanup, nil
		},
	})
	defer sim.Close()

	// create context for simulation run
	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel should come before defer simulation teardown
	defer cancel()
	const filesize = 1000 //kb
	_, err := sim.AddNodesAndConnectStar(10)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		nodeIDs := sim.UpNodeIDs()

		nodeIndex := make(map[enode.ID]int)
		for i, id := range nodeIDs {
			nodeIndex[id] = i
		}
		log.Error("node indexes", "idx", nodeIndex, "nodeIDs", nodeIDs)
		disconnected := watchDisconnections(ctx, sim)
		defer func() {
			if err != nil && disconnected.bool() {
				err = errors.New("disconnect events received")
			}
		}()

		randomBytes := testutil.RandomBytes(1010, filesize*1000)

		chunkAddrs, err := getAllRefs(randomBytes[:])
		if err != nil {
			return err
		}
		chunksProx := make([]chunkProxData, 0)
		for _, chunkAddr := range chunkAddrs {
			chunkInfo := chunkProxData{
				addr:            chunkAddr,
				uploaderNodePO:  chunk.Proximity(nodeIDs[0].Bytes(), chunkAddr),
				nodeProximities: make(map[enode.ID]int),
			}
			closestNodePO := 0
			for nodeAddr, _ := range nodeIndex {
				po := chunk.Proximity(nodeAddr.Bytes(), chunkAddr)

				chunkInfo.nodeProximities[nodeAddr] = po
				if po > closestNodePO {
					chunkInfo.closestNodePO = po
					chunkInfo.closestNode = nodeAddr
				}
				log.Error("processed chunk", "uploaderPO", chunkInfo.uploaderNodePO, "ci", chunkInfo.closestNode, "cpo", chunkInfo.closestNodePO, "cadrr", chunkInfo.addr)
			}
			chunksProx = append(chunksProx, chunkInfo)
		}

		// get the pivot node and pump some data
		item, ok := sim.NodeItem(nodeIDs[0], bucketKeyFileStore)
		if !ok {
			return fmt.Errorf("No filestore")
		}
		fileStore := item.(*storage.FileStore)
		size := chunkCount * chunkSize
		reader := bytes.NewReader(randomBytes[:])
		_, wait1, err := fileStore.Store(ctx, reader, int64(size), false)
		if err != nil {
			return fmt.Errorf("fileStore.Store: %v", err)
		}

		wait1(ctx)

		// check that chunks with a marked proximate host are where they should be
		count := 0
		goto other
		// wait to sync
		time.Sleep(30 * time.Second)
		for _, c := range chunksProx {
			// if the most proximate host is set - check that the chunk is there
			if c.closestNodePO > 0 {
				count++
				log.Warn("found chunk with proximate host set, trying to find in localstore", "po", c.closestNodePO, "closestNode", c.closestNode)
				item, ok = sim.NodeItem(c.closestNode, bucketKeyStore)
				if !ok {
					return fmt.Errorf("No DB")
				}
				store := item.(chunk.Store)

				_, err := store.Get(context.TODO(), chunk.ModeGetRequest, c.addr)
				if err != nil {
					return err
				}
			}
		}
		log.Debug("done checking stores", "checked chunks", count, "total chunks", len(chunksProx))
	other:
		// check that chunks from each po are _not_ on nodes that don't have subscriptions for these POs
		//create rpc client
		node := sim.Net.GetNode(nodeIDs[0])
		client, err := node.Client()
		if err != nil {
			return fmt.Errorf("create node 1 rpc client fail: %v", err)
		}

		//ask it for subscriptions
		pstreams := make(map[string][]string)
		err = client.Call(&pstreams, "stream_getPeerServerSubscriptions")
		if err != nil {
			return fmt.Errorf("client call stream_getPeerSubscriptions: %v", err)
		}

		//create a map of subscribed POs per node
		subMap := make(map[enode.ID][]int)
		//create a map of no-subs for a node
		noSubMap := make(map[enode.ID][]int)
		noSubMapMap := make(map[enode.ID]map[int]bool)

		for subscribedNode, streams := range pstreams {
			id := enode.HexID(subscribedNode)
			subscriptions := make([]int, 0)
			b := make([]bool, 17)
			for _, sub := range streams {
				subPO, err := ParseSyncBinKey(strings.Split(sub, "|")[1])
				if err != nil {
					return err
				}
				b[int(subPO)] = true
				found := false

				for _, v := range subscriptions {
					if v == int(subPO) {
						found = true
					}
				}
				if !found {
					subscriptions = append(subscriptions, int(subPO))
				}
			}
			noSubs := make([]int, 0)
			noMapMap := make(map[int]bool)
			for i, v := range b {
				if !v {
					noSubs = append(noSubs, i)
					noMapMap[i] = true
				}
			}
			noSubMapMap[id] = noMapMap
			noSubMap[id] = noSubs
			subMap[id] = subscriptions
		}

		// iterate over noSubMap, for each node check if it has any of the chunks it shouldn't have
		for nodeId, noSubs := range noSubMap {
			for _, c := range chunksProx {
				// if the chunk PO is equal to the sub that the node shouldnt have - check if the node has the chunk!
				if _, ok := noSubMapMap[nodeId][c.uploaderNodePO]; ok {
					count++
					item, ok = sim.NodeItem(nodeId, bucketKeyStore)
					if !ok {
						return fmt.Errorf("No DB")
					}
					store := item.(chunk.Store)

					_, err := store.Get(context.TODO(), chunk.ModeGetRequest, c.addr)
					if err == nil {
						return fmt.Errorf("got a chunk where it shouldn't be! addr %s, nodeId %s", c.addr, nodeId)
					}
				}
			}

		}
		return nil
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

type chunkProxData struct {
	addr            chunk.Address
	uploaderNodePO  int
	nodeProximities map[enode.ID]int
	closestNode     enode.ID
	closestNodePO   int
}
