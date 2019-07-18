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

package newstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/simulations"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

var (
	bucketKeyFileStore = simulation.BucketKey("filestore")
	bucketKeyBinIndex  = simulation.BucketKey("bin-indexes")
	bucketKeyStream    = simulation.BucketKey("stream")

	simContextTimeout = 20 * time.Second
)

// TestNodesExchangeCorrectBinIndexes tests that two nodes exchange the correct cursors for all streams
// it tests that all streams are exchanged
func TestNodesExchangeCorrectBinIndexes(t *testing.T) {
	nodeCount := 2

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000),
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), simContextTimeout)
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
		sim.NodeItem(idOne, bucketKeyStream).(*SlipStream).getPeer(idOther).mtx.Lock()
		onesCursors := sim.NodeItem(idOne, bucketKeyStream).(*SlipStream).getPeer(idOther).getCursorsCopy()
		sim.NodeItem(idOne, bucketKeyStream).(*SlipStream).getPeer(idOther).mtx.Unlock()

		sim.NodeItem(idOther, bucketKeyStream).(*SlipStream).getPeer(idOne).mtx.Lock()
		othersCursors := sim.NodeItem(idOther, bucketKeyStream).(*SlipStream).getPeer(idOne).getCursorsCopy()
		sim.NodeItem(idOther, bucketKeyStream).(*SlipStream).getPeer(idOne).mtx.Unlock()

		onesBins := sim.NodeItem(idOne, bucketKeyBinIndex).([]uint64)
		othersBins := sim.NodeItem(idOther, bucketKeyBinIndex).([]uint64)

		if err := compareNodeBinsToStreams(t, onesCursors, othersBins); err != nil {
			return err
		}
		if err := compareNodeBinsToStreams(t, othersCursors, onesBins); err != nil {
			return err
		}

		// check that the stream fetchers were created on each node
		//checkHistoricalStreams(t, onesCursors, onesHistoricalFetchers)
		//checkHistoricalStreams(t, othersCursors, othersHistoricalFetchers)

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

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000),
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), simContextTimeout)
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
			peerRecord := sim.NodeItem(idPivot, bucketKeyStream).(*SlipStream).getPeer(idOther)

			// these are the cursors that the pivot node holds for the other peer
			pivotCursors := peerRecord.getCursorsCopy()
			otherSyncer := sim.NodeItem(idOther, bucketKeyStream).(*SlipStream).getPeer(idPivot)
			otherCursors := otherSyncer.getCursorsCopy()
			otherKademlia := sim.NodeItem(idOther, simulation.BucketKeyKademlia).(*network.Kademlia)

			othersBins := sim.NodeItem(idOther, bucketKeyBinIndex).([]uint64)

			po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
			depth := pivotKademlia.NeighbourhoodDepth()
			log.Debug("i", "i", i, "po", po, "d", depth, "idOther", idOther, "peerRecord", peerRecord, "pivotCursors", pivotCursors)

			// if the peer is outside the depth - the pivot node should not request any streams
			if po >= depth {
				if err := compareNodeBinsToStreams(t, pivotCursors, othersBins); err != nil {
					return err
				}
			}

			if err := compareNodeBinsToStreams(t, otherCursors, pivotBins); err != nil {
				return err
			}
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

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000),
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), simContextTimeout)
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
		pivotSyncer := sim.NodeItem(idPivot, bucketKeyStream)
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
				pivotCursors := pivotSyncer.(*SlipStream).getPeer(idOther).getCursorsCopy()

				// check that the pivot node is interested just in bins >= depth
				if po >= depth {
					othersBins := sim.NodeItem(idOther, bucketKeyBinIndex).([]uint64)
					if err := compareNodeBinsToStreamsWithDepth(t, pivotCursors, othersBins, pivotDepth); err != nil {
						return err
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

var reestablishCursorsSnapshotFilename = "testdata/reestablish-cursors-snapshot.json"

// TestNodesRemovesCursors creates a pivot network of 2 nodes where the pivot's depth = 0.
// test sequence:
// - select another node with po >= depth (of the pivot's kademlia)
// - add other nodes to the pivot until the depth goes above that peer's po (depth > peerPo)
// - asserts that the pivot does not maintain any cursors of the node that moved out of depth
// - start removing nodes from the simulation until that peer is again within depth
// - check that the cursors are being re-established
func xTestNodeRemovesAndReestablishCursors(t *testing.T) {
	if *update {
		generateReestablishCursorsSnapshot(t, 2)
	}

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000),
	})

	// load the snapshot
	if err := sim.UploadSnapshot(context.Background(), reestablishCursorsSnapshotFilename); err != nil {
		t.Fatal(err)
	}
	// load additional test specific data from the snapshot
	d, err := ioutil.ReadFile(reestablishCursorsSnapshotFilename)
	if err != nil {
		t.Fatal(err)
	}
	var s reestablishCursorsState
	if err := json.Unmarshal(d, &s); err != nil {
		t.Fatal(err)
	}

	pivotEnode := s.PivotEnode
	pivotKademlia := sim.NodeItem(pivotEnode, simulation.BucketKeyKademlia).(*network.Kademlia)
	lookupEnode := s.LookupEnode
	lookupPO := s.PO
	nodeCount := len(sim.UpNodeIDs())

	log.Debug("tracking enode", "enode", lookupEnode, "po", lookupPO)

	// expecting some cursors
	waitForCursors(t, sim, pivotEnode, lookupEnode, true)

	//append nodes to simulation until the node po moves out of the depth, then assert no subs from pivot to that node
	for i := float64(1); pivotKademlia.NeighbourhoodDepth() <= lookupPO; i++ {
		// calculate the number of nodes to add:
		// - logarithmically increase by the number of iterations
		//   - ensure that the logarithm is greater then 0 by starting the iteration from 1, not 0
		//   - ensure that the logarithm is greater then 0 by adding 1
		// - multiply by the difference between target and current depth
		//   - ensure that is greater then 0 by adding 1
		//   - multiply it by empirical constant 4
		newNodeCount := int(math.Logb(i)+1) * ((lookupPO-pivotKademlia.NeighbourhoodDepth())*4 + 1)
		id, err := sim.AddNodes(newNodeCount)
		if err != nil {
			t.Fatal(err)
		}
		err = sim.Net.ConnectNodesStar(id, pivotEnode)
		if err != nil {
			t.Fatal(err)
		}
		nodeCount += newNodeCount
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodeCount {
			t.Fatalf("got %v up nodes, want %v", len(nodeIDs), nodeCount)
		}
		log.Debug("added new nodes to reach depth", "new nodes", newNodeCount, "current depth", pivotKademlia.NeighbourhoodDepth(), "target depth", lookupPO)
	}

	log.Debug("added nodes to sim, node moved out of depth", "depth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO, "lookupEnode", lookupEnode)

	// no cursors should exist at this point
	waitForCursors(t, sim, pivotEnode, lookupEnode, false)

	var removed int
	// remove nodes from the simulation until the peer moves again into depth
	log.Error("pulling the plug on some nodes to make the depth go up again", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO, "lookupEnode", lookupEnode)
	for pivotKademlia.NeighbourhoodDepth() > lookupPO {
		_, err := sim.StopRandomNode(pivotEnode, lookupEnode)
		if err != nil {
			t.Fatal(err)
		}
		removed++
		log.Debug("removed 1 node", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO)
	}
	log.Debug("done removing nodes", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO, "removed", removed)

	// expecting some new cursors
	waitForCursors(t, sim, pivotEnode, lookupEnode, true)
}

// compareNodeBinsToStreams checks that the values on `onesCursors` correlate to the values in `othersBins`
// onesCursors represents the stream cursors that node A knows about node B (i.e. they shoud reflect directly in this case
// the values which node B retrieved from its local store)
// othersBins is the array of bin indexes on node B's local store as they were inserted into the store
func compareNodeBinsToStreams(t *testing.T, onesCursors map[string]uint64, othersBins []uint64) (err error) {
	if len(onesCursors) == 0 {
		return errors.New("no cursors")
	}
	if len(othersBins) == 0 {
		return errors.New("no bins")
	}

	for nameKey, cur := range onesCursors {
		id, err := strconv.Atoi(parseID(nameKey).Key)
		if err != nil {
			return err
		}
		if othersBins[id] != uint64(cur) {
			return fmt.Errorf("bin indexes not equal. bin %d, got %d, want %d", id, cur, othersBins[id])
		}
	}
	return nil
}

func parseID(str string) ID {
	v := strings.Split(str, "|")
	if len(v) != 2 {
		panic("too short")
	}
	return NewID(v[0], v[1])
}

func compareNodeBinsToStreamsWithDepth(t *testing.T, onesCursors map[string]uint64, othersBins []uint64, depth uint) (err error) {
	log.Debug("compareNodeBinsToStreamsWithDepth", "cursors", onesCursors, "othersBins", othersBins, "depth", depth)
	if len(onesCursors) == 0 || len(othersBins) == 0 {
		return errors.New("no cursors")
	}
	// inclusive test
	for nameKey, cur := range onesCursors {
		bin, err := strconv.Atoi(parseID(nameKey).Key)
		if err != nil {
			return err
		}
		if uint(bin) < depth {
			return fmt.Errorf("cursor at bin %d should not exist. depth %d", bin, depth)
		}
		if othersBins[bin] != uint64(cur) {
			return fmt.Errorf("bin indexes not equal. bin %d, got %d, want %d", bin, cur, othersBins[bin])
		}
	}

	// exclusive test
	for i := 0; i < int(depth); i++ {
		// should not have anything shallower than depth
		id := NewID("SYNC", fmt.Sprintf("%d", i))
		if _, ok := onesCursors[id.String()]; ok {
			return fmt.Errorf("oneCursors contains id %s, but it should not", id)
		}
	}
	return nil
}

func newBzzSyncWithLocalstoreDataInsertion(numChunks int) func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	return func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
		n := ctx.Config.Node()
		addr := network.NewAddr(n)

		localStore, localStoreCleanup, err := newTestLocalStore(n.ID(), addr, nil)
		if err != nil {
			return nil, nil, err
		}

		kad := network.NewKademlia(addr.Over(), network.NewKadParams())
		netStore := storage.NewNetStore(localStore, n.ID())
		lnetStore := storage.NewLNetStore(netStore)
		fileStore := storage.NewFileStore(lnetStore, storage.NewFileStoreParams(), chunk.NewTags())
		if numChunks > 0 {
			filesize := numChunks * 4096
			cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, wait, err := fileStore.Store(cctx, testutil.RandomReader(int(time.Now().Unix()), filesize), int64(filesize), false)
			if err != nil {
				return nil, nil, err
			}
			if err := wait(cctx); err != nil {
				return nil, nil, err
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

		var store *state.DBStore
		// Use on-disk DBStore to reduce memory consumption in race tests.
		dir, err := ioutil.TempDir(tmpDir, "statestore-")
		if err != nil {
			return nil, nil, err
		}
		store, err = state.NewDBStore(dir)
		if err != nil {
			return nil, nil, err
		}

		sp := NewSyncProvider(netStore, kad)
		o := NewSlipStream(store, sp)
		bucket.Store(bucketKeyBinIndex, binIndexes)
		bucket.Store(bucketKeyFileStore, fileStore)
		bucket.Store(simulation.BucketKeyKademlia, kad)
		bucket.Store(bucketKeyStream, o)

		cleanup = func() {
			localStore.Close()
			localStoreCleanup()
			store.Close()
			os.RemoveAll(dir)
		}

		return o, cleanup, nil
	}
}

// data appended to reestablish cursors snapshot
type reestablishCursorsState struct {
	PivotEnode  enode.ID `json:"pivotEnode"`
	LookupEnode enode.ID `json:"lookupEnode"`
	PO          int      `json:"po"`
}

// function that generates a simulation and saves its snapshot for
// TestNodeRemovesAndReestablishCursors test.
func generateReestablishCursorsSnapshot(t *testing.T, tagetPO int) {
	sim, pivotEnode, lookupEnode := setupReestablishCursorsSimulation(t, tagetPO)
	defer sim.Close()

	waitForCursors(t, sim, pivotEnode, lookupEnode, true)

	s, err := sim.Net.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	d, err := json.Marshal(struct {
		*simulations.Snapshot
		reestablishCursorsState
	}{
		Snapshot: s,
		reestablishCursorsState: reestablishCursorsState{
			PivotEnode:  pivotEnode,
			LookupEnode: lookupEnode,
			PO:          tagetPO,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("save snapshot file")

	err = ioutil.WriteFile(reestablishCursorsSnapshotFilename, d, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

// function that generates a simulation that can be used in TestNodeRemovesAndReestablishCursors
// test with a provided target po which is a depth to reach in the test by adding new nodes.
func setupReestablishCursorsSimulation(t *testing.T, tagetPO int) (sim *simulation.Simulation, pivotEnode, lookupEnode enode.ID) {
	// initial node count
	nodeCount := 5

	sim = simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000),
	})

	nodeIDs, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	pivotEnode = nodeIDs[0]
	log.Debug("simulation pivot node", "id", pivotEnode)
	pivotKademlia := sim.NodeItem(pivotEnode, simulation.BucketKeyKademlia).(*network.Kademlia)

	// make sure that we get a node with po <= depth
	for i := 1; i < nodeCount; i++ {
		log.Debug("looking for a peer", "i", i, "nodecount", nodeCount)
		otherKademlia := sim.NodeItem(nodeIDs[i], simulation.BucketKeyKademlia).(*network.Kademlia)
		po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
		depth := pivotKademlia.NeighbourhoodDepth()
		if po > depth {
			if po != tagetPO {
				log.Debug("wrong depth to reach, generating new simulation", "depth", po)
				return setupReestablishCursorsSimulation(t, tagetPO)
			}
			lookupEnode = nodeIDs[i]
			return
		}
		// append a node to the simulation
		id, err := sim.AddNode()
		if err != nil {
			t.Fatal(err)
		}
		log.Debug("added node to simulation, connecting to pivot", "id", id, "pivot", pivotEnode)
		if err = sim.Net.Connect(id, pivotEnode); err != nil {
			t.Fatal(err)
		}
		nodeCount++
		// wait for node to be set
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("node with po<=depth not found")
	return
}

// waitForCursors checks if the pivot node has some cursors or not
// by periodically checking for them.
func waitForCursors(t *testing.T, sim *simulation.Simulation, pivotEnode, lookupEnode enode.ID, wantSome bool) {
	t.Helper()

	var got int
	for i := 0; i < 1000; i++ { // 10s total wait
		time.Sleep(10 * time.Millisecond)
		s, ok := sim.NodeItem(pivotEnode, bucketKeyStream).(*SlipStream)
		if !ok {
			continue
		}
		p := s.getPeer(lookupEnode)
		if p == nil {
			continue
		}
		got = len(p.getCursorsCopy())
		if got != 0 == wantSome {
			return
		}
	}
	if wantSome {
		t.Fatalf("got %v cursors, but want some", got)
	} else {
		t.Fatalf("got %v cursors, but want none", got)
	}
}
