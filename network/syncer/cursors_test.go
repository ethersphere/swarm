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
	// defer cancel should come before defer simulation teardown
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

		nodeIndex := make(map[enode.ID]int)
		for i, id := range nodeIDs {
			nodeIndex[id] = i
		}
		// wait for the nodes to exchange StreamInfo messages
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < nodeCount; i++ {
			idOne := nodeIDs[i]
			idOther := nodeIDs[(i+1)%nodeCount]
			onesSyncer := sim.NodeItem(idOne, bucketKeySyncer)

			s := onesSyncer.(*SwarmSyncer)
			onesCursors := s.peers[idOther].streamCursors
			othersBins := sim.NodeItem(idOther, bucketKeyBinIndex)

			compareNodeBinsToStreams(t, onesCursors, othersBins.([]uint64))
		}
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
		for i := 1; i < nodeCount; i++ {
			idOther := nodeIDs[i]
			pivotSyncer := sim.NodeItem(idPivot, bucketKeySyncer)
			otherSyncer := sim.NodeItem(idOther, bucketKeySyncer)

			pivotCursors := pivotSyncer.(*SwarmSyncer).peers[idOther].streamCursors
			otherCursors := otherSyncer.(*SwarmSyncer).peers[idPivot].streamCursors

			othersBins := sim.NodeItem(idOther, bucketKeyBinIndex)
			pivotBins := sim.NodeItem(idPivot, bucketKeyBinIndex)

			compareNodeBinsToStreams(t, pivotCursors, othersBins.([]uint64))
			compareNodeBinsToStreams(t, otherCursors, pivotBins.([]uint64))
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
	nodeCount := 8

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
		nodeIndex := make(map[enode.ID]int)
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != 2 {
			return errors.New("not enough nodes up")
		}

		for i, id := range nodeIDs {
			nodeIndex[id] = i
		}
		// wait for the nodes to exchange StreamInfo messages
		time.Sleep(100 * time.Millisecond)
		idPivot := nodeIDs[0]
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

			idPivot = nodeIDs[0]
			for i := 1; i < j; i++ {
				idOther := nodeIDs[i]
				pivotSyncer := sim.NodeItem(idPivot, bucketKeySyncer)
				pivotCursors := pivotSyncer.(*SwarmSyncer).peers[idOther].streamCursors
				pivotKademlia := sim.NodeItem(idPivot, simulation.BucketKeyKademlia)
				pivotDepth := pivotKademlia.NeighbourhoodDepth()

				othersBins := sim.NodeItem(idOther, bucketKeyBinIndex)

				compareNodeBinsToStreams(t, pivotCursors, othersBins.([]uint64))
			}
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	/*	*/

}

// compareNodeBinsToStreams checks that the values on `onesCursors` correlate to the values in `othersBins`
// onesCursors represents the stream cursors that node A knows about node B (i.e. they shoud reflect directly in this case
// the values which node B retrieved from its local store)
// othersBins is the array of bin indexes on node B's local store as they were inserted into the store
func compareNodeBinsToStreams(t *testing.T, onesCursors map[uint]*uint, othersBins []uint64) {
	for bin, cur := range onesCursors {
		if cur == nil {
			continue
		}
		if othersBins[bin] != uint64(*cur) {
			t.Fatalf("bin indexes not equal. bin %d, got %d, want %d", bin, cur, othersBins[bin])
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

	filesize := 2000 * 4096
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
	for i := 0; i <= 7; i++ {
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
