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
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/network/simulation"
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

	// create a standard sim
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion,
	})
	defer sim.Close()

	// create context for simulation run
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	// defer cancel should come before defer simulation teardown
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	//run the simulation
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != 2 {
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
			idOther := nodeIDs[(i+1)%2]
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

	// create a standard sim
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion,
	})
	defer sim.Close()

	// create context for simulation run
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	// defer cancel should come before defer simulation teardown
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	//run the simulation
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

	// create a standard sim
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion,
	})
	defer sim.Close()

	// create context for simulation run
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	// defer cancel should come before defer simulation teardown
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(2)
	if err != nil {
		t.Fatal(err)
	}

	//run the simulation
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

			for i, id := range nodeIDs {
				nodeIndex[id] = i
			}

			idPivot = nodeIDs[0]
			for i := 1; i < j; i++ {
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
