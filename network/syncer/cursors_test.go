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

package syncer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

var (
	bucketKeyFileStore = simulation.BucketKey("filestore")
	bucketKeyBinIndex  = simulation.BucketKey("bin-indexes")
	bucketKeySyncer    = simulation.BucketKey("syncer")
)

// TestNodesExchangeCorrectBinIndexes tests that two nodes exchange the correct cursors for all streams
// it tests that all streams are exchanged
func TestNodesExchangeCorrectBinIndexes(t *testing.T) {
	nodeCount := 2

	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion,
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodeCount {
			return errors.New("not enough nodes up")
		}

		// wait for the nodes to exchange StreamInfo messages
		time.Sleep(100 * time.Millisecond)
		idOne := nodeIDs[0]
		idOther := nodeIDs[1]
		onesCursors := sim.NodeItem(idOne, bucketKeySyncer).(*SwarmSyncer).peers[idOther].streamCursors
		othersCursors := sim.NodeItem(idOther, bucketKeySyncer).(*SwarmSyncer).peers[idOne].streamCursors

		onesHistoricalFetchers := sim.NodeItem(idOne, bucketKeySyncer).(*SwarmSyncer).peers[idOther].historicalStreams
		othersHistoricalFetchers := sim.NodeItem(idOther, bucketKeySyncer).(*SwarmSyncer).peers[idOne].historicalStreams

		onesBins := sim.NodeItem(idOne, bucketKeyBinIndex).([]uint64)
		othersBins := sim.NodeItem(idOther, bucketKeyBinIndex).([]uint64)

		compareNodeBinsToStreams(t, onesCursors, othersBins)
		compareNodeBinsToStreams(t, othersCursors, onesBins)

		// check that the stream fetchers were created on each node
		checkHistoricalStreamStates(t, onesCursors, onesHistoricalFetchers)
		checkHistoricalStreamStates(t, othersCursors, othersHistoricalFetchers)

		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestNodesExchangeCorrectBinIndexesInPivot creates a pivot network of 8 nodes, in which the pivot node
// has depth > 0, puts data into every node's localstore and checks that the pivot node exchanges
// with each other node the correct indexes
func TestNodesExchangeCorrectBinIndexesInPivot(t *testing.T) {
	nodeCount := 8

	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion,
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodeCount {
			return errors.New("not enough nodes up")
		}

		// wait for the nodes to exchange StreamInfo messages
		time.Sleep(100 * time.Millisecond)
		idPivot := nodeIDs[0]
		pivotBins := sim.NodeItem(idPivot, bucketKeyBinIndex).([]uint64)
		pivotKademlia := sim.NodeItem(idPivot, simulation.BucketKeyKademlia).(*network.Kademlia)

		for i := 1; i < nodeCount; i++ {
			idOther := nodeIDs[i]
			pivotPeers := sim.NodeItem(idPivot, bucketKeySyncer).(*SwarmSyncer).peers
			peerRecord := sim.NodeItem(idPivot, bucketKeySyncer).(*SwarmSyncer).peers[idOther]
			pivotCursors := sim.NodeItem(idPivot, bucketKeySyncer).(*SwarmSyncer).peers[idOther].streamCursors
			otherSyncer := sim.NodeItem(idOther, bucketKeySyncer)
			otherCursors := otherSyncer.(*SwarmSyncer).peers[idPivot].streamCursors
			otherKademlia := sim.NodeItem(idOther, simulation.BucketKeyKademlia).(*network.Kademlia)

			othersBins := sim.NodeItem(idOther, bucketKeyBinIndex).([]uint64)

			po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
			depth := pivotKademlia.NeighbourhoodDepth()
			log.Debug("i", "i", i, "po", po, "d", depth, "idOther", idOther, "peerRecord", peerRecord, "pivotCursors", pivotCursors, "peers", pivotPeers)

			// if the peer is outside the depth - the pivot node should not request any streams
			if po >= depth {
				compareNodeBinsToStreams(t, pivotCursors, othersBins)
			}

			compareNodeBinsToStreams(t, otherCursors, pivotBins)
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestNodesCorrectBinsDynamic adds nodes to a star toplogy, connecting new nodes to the pivot node
// after each connection is made, the cursors on the pivot are checked, to reflect the bins that we are
// currently still interested in. this makes sure that correct bins are of interest
// when nodes enter the kademlia of the pivot node
func TestNodesCorrectBinsDynamic(t *testing.T) {
	nodeCount := 10

	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion,
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(2)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != 2 {
			return errors.New("not enough nodes up")
		}

		// wait for the nodes to exchange StreamInfo messages
		time.Sleep(100 * time.Millisecond)
		idPivot := nodeIDs[0]
		pivotSyncer := sim.NodeItem(idPivot, bucketKeySyncer)
		pivotKademlia := sim.NodeItem(idPivot, simulation.BucketKeyKademlia).(*network.Kademlia)
		pivotDepth := uint(pivotKademlia.NeighbourhoodDepth())

		for j := 2; j <= nodeCount; j++ {
			// append a node to the simulation
			id, err := sim.AddNodes(1)
			if err != nil {
				return err
			}
			err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
			if err != nil {
				return err
			}
			nodeIDs := sim.UpNodeIDs()
			if len(nodeIDs) != j+1 {
				return fmt.Errorf("not enough nodes up. got %d, want %d", len(nodeIDs), j)
			}
			time.Sleep(50 * time.Millisecond)
			idPivot = nodeIDs[0]
			for i := 1; i < j; i++ {
				idOther := nodeIDs[i]
				otherKademlia := sim.NodeItem(idOther, simulation.BucketKeyKademlia).(*network.Kademlia)
				po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
				depth := pivotKademlia.NeighbourhoodDepth()
				pivotCursors := pivotSyncer.(*SwarmSyncer).peers[idOther].streamCursors

				// check that the pivot node is interested just in bins >= depth
				if po >= depth {
					othersBins := sim.NodeItem(idOther, bucketKeyBinIndex).([]uint64)
					compareNodeBinsToStreamsWithDepth(t, pivotCursors, othersBins, pivotDepth)
				}
			}
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestNodesRemovesCursors creates a pivot network of 2 nodes where the pivot's depth = 0.
// the test then selects another node with po=0 to the pivot, and starts adding other nodes to the pivot until the depth goes above 0
// the test then asserts that the pivot does not maintain any cursors of the node that moved out of depth
func TestNodeRemovesCursors(t *testing.T) {
	nodeCount := 2

	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion,
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodeCount {
			return errors.New("not enough nodes up")
		}

		// wait for the nodes to exchange StreamInfo messages
		time.Sleep(100 * time.Millisecond)
		idPivot := nodeIDs[0]
		pivotKademlia := sim.NodeItem(idPivot, simulation.BucketKeyKademlia).(*network.Kademlia)
		// make sure that we get an otherID with po <= depth
		found := false
		foundId := 0
		foundPo := 0
		for i := 1; i < nodeCount; i++ {
			log.Debug("looking for a peer", "i", i, "nodecount", nodeCount)
			idOther := nodeIDs[i]
			otherKademlia := sim.NodeItem(idOther, simulation.BucketKeyKademlia).(*network.Kademlia)
			po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
			depth := pivotKademlia.NeighbourhoodDepth()
			if po <= depth {
				foundId = i
				foundPo = po
				found = true
				break
			}

			// append a node to the simulation
			id, err := sim.AddNodes(1)
			if err != nil {
				return err
			}
			err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
			if err != nil {
				return err
			}
			nodeCount += 1
			nodeIDs = sim.UpNodeIDs()
			if len(nodeIDs) != nodeCount {
				return fmt.Errorf("not enough nodes up. got %d, want %d", len(nodeIDs), nodeCount)
			}
		}
		time.Sleep(500 * time.Millisecond)
		if !found {
			panic("did not find a node with po<=depth")
		} else {
			pivotCursors := sim.NodeItem(nodeIDs[0], bucketKeySyncer).(*SwarmSyncer).peers[nodeIDs[foundId]].streamCursors
			if len(pivotCursors) == 0 {
				panic("pivotCursors for node should not be empty")
			}
		}

		//append nodes to simulation until the node po moves out of the depth, then assert no subs from pivot to that node
		for pivotKademlia.NeighbourhoodDepth() <= foundPo {
			id, err := sim.AddNodes(1)
			if err != nil {
				return err
			}
			err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
			if err != nil {
				return err
			}
			nodeCount += 1
			nodeIDs = sim.UpNodeIDs()
			if len(nodeIDs) != nodeCount {
				return fmt.Errorf("not enough nodes up. got %d, want %d", len(nodeIDs), nodeCount)
			}
		}

		log.Debug("added nodes to sim, node moved out of depth", "depth", pivotKademlia.NeighbourhoodDepth(), "peerPo", foundPo, "foundId", foundId, "nodeIDs", nodeIDs)

		pivotCursors := sim.NodeItem(nodeIDs[0], bucketKeySyncer).(*SwarmSyncer).peers[nodeIDs[foundId]].streamCursors
		if len(pivotCursors) > 0 {
			panic("pivotCursors for node should be empty")
		}

		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// compareNodeBinsToStreams checks that the values on `onesCursors` correlate to the values in `othersBins`
// onesCursors represents the stream cursors that node A knows about node B (i.e. they shoud reflect directly in this case
// the values which node B retrieved from its local store)
// othersBins is the array of bin indexes on node B's local store as they were inserted into the store
func compareNodeBinsToStreams(t *testing.T, onesCursors map[uint]uint64, othersBins []uint64) {
	if len(onesCursors) == 0 {
		panic("no cursors")
	}
	if len(othersBins) == 0 {
		panic("no bins")
	}

	for bin, cur := range onesCursors {
		if othersBins[bin] != uint64(cur) {
			t.Fatalf("bin indexes not equal. bin %d, got %d, want %d", bin, cur, othersBins[bin])
		}
	}
}

func compareNodeBinsToStreamsWithDepth(t *testing.T, onesCursors map[uint]uint64, othersBins []uint64, depth uint) {
	log.Debug("compareNodeBinsToStreamsWithDepth", "cursors", onesCursors, "othersBins", othersBins, "depth", depth)
	if len(onesCursors) == 0 || len(othersBins) == 0 {
		panic("no cursors")
	}
	// inclusive test
	for bin, cur := range onesCursors {
		if bin < depth {
			panic(fmt.Errorf("cursor at bin %d should not exist. depth %d", bin, depth))
		}
		if othersBins[bin] != uint64(cur) {
			panic(fmt.Errorf("bin indexes not equal. bin %d, got %d, want %d", bin, cur, othersBins[bin]))
		}
	}

	// exclusive test
	for i := 0; i < int(depth); i++ {
		// should not have anything shallower than depth
		if _, ok := onesCursors[uint(i)]; ok {
			panic("should be nil")
		}
	}
}

func checkHistoricalStreamStates(t *testing.T, onesCursors map[uint]uint64, onesStreams map[uint]*syncStreamFetch) {
	for k, v := range onesCursors {
		if v > 0 {
			// there should be a matching stream state
			if _, ok := onesStreams[k]; !ok {
				t.Fatalf("stream for bin id %d should exist", k)
			}
		} else {
			// index is zero -> no historical stream for this bin. check that it doesn't exist
			if _, ok := onesStreams[k]; ok {
				t.Fatalf("stream for bin id %d should not exist", k)
			}
		}
	}
}

func newBzzSyncWithLocalstoreDataInsertion(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	n := ctx.Config.Node()
	addr := network.NewAddr(n)

	localStore, localStoreCleanup, err := newTestLocalStore(n.ID(), addr, nil)
	if err != nil {
		return nil, nil, err
	}

	kad := network.NewKademlia(addr.Over(), network.NewKadParams())
	netStore := storage.NewNetStore(localStore, enode.ID{})
	lnetStore := storage.NewLNetStore(netStore)
	fileStore := storage.NewFileStore(lnetStore, storage.NewFileStoreParams(), chunk.NewTags())

	filesize := 1000 * 4096
	cctx := context.Background()
	_, wait, err := fileStore.Store(cctx, testutil.RandomReader(0, filesize), int64(filesize), false)
	if err != nil {
		return nil, nil, err
	}
	if err := wait(cctx); err != nil {
		return nil, nil, err
	}

	// verify bins just upto 8 (given random distribution and 1000 chunks
	// bin index `i` cardinality for `n` chunks is assumed to be n/(2^i+1)
	for i := 0; i <= 5; i++ {
		if binIndex, err := netStore.LastPullSubscriptionBinID(uint8(i)); binIndex == 0 || err != nil {
			return nil, nil, fmt.Errorf("error querying bin indexes. bin %d, index %d, err %v", i, binIndex, err)
		}
	}

	binIndexes := make([]uint64, 17)
	for i := 0; i <= 16; i++ {
		binIndex, err := netStore.LastPullSubscriptionBinID(uint8(i))
		if err != nil {
			return nil, nil, err
		}
		binIndexes[i] = binIndex
	}
	o := NewSwarmSyncer(enode.ID{}, nil, kad, netStore)
	bucket.Store(bucketKeyBinIndex, binIndexes)
	bucket.Store(bucketKeyFileStore, fileStore)
	bucket.Store(simulation.BucketKeyKademlia, kad)
	bucket.Store(bucketKeySyncer, o)

	cleanup = func() {
		localStore.Close()
		localStoreCleanup()
	}

	return o, cleanup, nil
}
