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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

var (
	bucketKeyFileStore = simulation.BucketKey("filestore")
	bucketKeyBinIndex  = simulation.BucketKey("bin-indexes")
	bucketKeySyncer    = simulation.BucketKey("syncer")

	simContextTimeout = 20 * time.Second
)

// TestNodesExchangeCorrectBinIndexes tests that two nodes exchange the correct cursors for all streams
// it tests that all streams are exchanged
func TestNodesExchangeCorrectBinIndexes(t *testing.T) {
	nodeCount := 2

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000, StreamGetCursors),
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
		sim.NodeItem(idOne, bucketKeySyncer).(*SlipStream).getPeer(idOther).mtx.Lock()
		onesCursors := sim.NodeItem(idOne, bucketKeySyncer).(*SlipStream).getPeer(idOther).getCursorsCopy()
		sim.NodeItem(idOne, bucketKeySyncer).(*SlipStream).getPeer(idOther).mtx.Unlock()

		sim.NodeItem(idOther, bucketKeySyncer).(*SlipStream).getPeer(idOne).mtx.Lock()
		othersCursors := sim.NodeItem(idOther, bucketKeySyncer).(*SlipStream).getPeer(idOne).getCursorsCopy()
		sim.NodeItem(idOther, bucketKeySyncer).(*SlipStream).getPeer(idOne).mtx.Unlock()

		//onesHistoricalFetchers := sim.NodeItem(idOne, bucketKeySyncer).(*SlipStream).getPeer(idOther).historicalStreams
		//othersHistoricalFetchers := sim.NodeItem(idOther, bucketKeySyncer).(*SlipStream).getPeer(idOne).historicalStreams

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

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000, StreamGetCursors),
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
			peerRecord := sim.NodeItem(idPivot, bucketKeySyncer).(*SlipStream).getPeer(idOther)

			// these are the cursors that the pivot node holds for the other peer
			pivotCursors := peerRecord.getCursorsCopy()
			otherSyncer := sim.NodeItem(idOther, bucketKeySyncer).(*SlipStream).getPeer(idPivot)
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
				//checkHistoricalStreams(t, pivotCursors, pivotHistoricalFetchers)
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

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000, StreamGetCursors),
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

// TestNodesRemovesCursors creates a pivot network of 2 nodes where the pivot's depth = 0.
// test sequence:
// - select another node with po >= depth (of the pivot's kademlia)
// - add other nodes to the pivot until the depth goes above that peer's po (depth > peerPo)
// - asserts that the pivot does not maintain any cursors of the node that moved out of depth
// - start removing nodes from the simulation until that peer is again within depth
// - check that the cursors are being re-established
func TestNodeRemovesAndReestablishCursors(t *testing.T) {
	nodeCount := 5

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(1000, StreamGetCursors),
	})
	defer sim.Close()

	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), simContextTimeout)
	defer cancel()

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodeCount {
			return errors.New("not enough nodes up")
		}

		// wait for the nodes to exchange StreamInfo messages
		time.Sleep(100 * time.Millisecond)
		idPivot := nodeIDs[0]
		log.Debug("simulation pivot node", "id", idPivot)
		pivotKademlia := sim.NodeItem(idPivot, simulation.BucketKeyKademlia).(*network.Kademlia)
		// make sure that we get an otherID with po <= depth
		found := false
		foundId := 0
		foundPo := 0
		var foundEnode enode.ID
		//pivotPeerLen = len(sim.NodeItem(idPivot, bucketKeySyncer).(*SwarmSyncer).peers)
		for i := 1; i < nodeCount; i++ {
			log.Debug("looking for a peer", "i", i, "nodecount", nodeCount)
			idOther := nodeIDs[i]
			otherKademlia := sim.NodeItem(idOther, simulation.BucketKeyKademlia).(*network.Kademlia)
			po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
			depth := pivotKademlia.NeighbourhoodDepth()
			if po >= depth {
				foundId = i
				foundPo = po
				found = true
				foundEnode = nodeIDs[i]
				// check that we established some streams for this peer
				//pivotCursors := sim.NodeItem(idPivot, bucketKeySyncer).(*SlipStream).getPeer(idOther).getCursorsCopy()
				//pivotHistoricalFetchers := sim.NodeItem(idPivot, bucketKeySyncer).(*SlipStream).getPeer(idOther).historicalStreams

				//checkHistoricalStreams(t, pivotCursors, pivotHistoricalFetchers)
				break
			}

			// append a node to the simulation
			id, err := sim.AddNodes(1)
			if err != nil {
				return err
			}
			log.Debug("added node to simulation, connecting to pivot", "id", id, "pivot", idPivot)
			err = sim.Net.ConnectNodesStar(id, idPivot)
			if err != nil {
				return err
			}
			nodeCount++
			nodeIDs = sim.UpNodeIDs()
			if len(nodeIDs) != nodeCount {
				return fmt.Errorf("not enough nodes up. got %d, want %d", len(nodeIDs), nodeCount)
			}

			// allow the new node to exchange the stream info messages
			time.Sleep(200 * time.Millisecond)
		}

		if !found {
			panic("did not find a node with po<=depth")
		} else {
			log.Debug("tracking enode", "enode", foundEnode)
			cursorsCount := sim.NodeItem(nodeIDs[0], bucketKeySyncer).(*SlipStream).getPeer(nodeIDs[foundId]).cursorsCount()
			if cursorsCount == 0 {
				return errors.New("pivotCursors for node should not be empty")
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
			nodeCount++
			nodeIDs = sim.UpNodeIDs()
			if len(nodeIDs) != nodeCount {
				return fmt.Errorf("not enough nodes up. got %d, want %d", len(nodeIDs), nodeCount)
			}
		}

		log.Debug("added nodes to sim, node moved out of depth", "depth", pivotKademlia.NeighbourhoodDepth(), "peerPo", foundPo, "foundId", foundId, "nodeIDs", nodeIDs)

		pivotCursors := sim.NodeItem(nodeIDs[0], bucketKeySyncer).(*SlipStream).getPeer(nodeIDs[foundId]).getCursorsCopy()
		if len(pivotCursors) != 0 {
			panic("pivotCursors for node should be empty")
		}
		//pvotHistoricalFetchers := sim.NodeItem(idPivot, bucketKeySyncer).(*SlipStream).getPeer(nodeIDs)foundId]].historicalStreams
		//if len(pivotHistoricalFetchers) != 0 {
		//log.Error("pivot fetcher length>0", "len", len(pivotHistoricalFetchers))
		//panic("pivot historical fetchers for node should be empty")
		//}
		removed := 0
		// remove nodes from the simulation until the peer moves again into depth
		log.Error("pulling the plug on some nodes to make the depth go up again", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", foundPo, "peerIndex", foundId)
		for pivotKademlia.NeighbourhoodDepth() > foundPo {
			_, err := sim.StopRandomNode(nodeIDs[0], foundEnode)
			if err != nil {
				panic(err)
			}
			removed++
			time.Sleep(100 * time.Millisecond)
			log.Error("removed 1 node", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", foundPo)

			nodeIDs = sim.UpNodeIDs()
		}
		log.Error("done removing nodes", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", foundPo, "removed", removed)

		// wait for cursors msg again
		time.Sleep(100 * time.Millisecond)
		peer := sim.NodeItem(idPivot, bucketKeySyncer).(*SlipStream).getPeer(foundEnode)
		if peer.cursorsCount() == 0 {
			return errors.New("pivotCursors for foundEnode should no longer be empty")
		}
		//pivotHistoricalFetchers = sim.NodeItem(idPivot, bucketKeySyncer).(*SlipStream).getPeer(foundEnode).historicalStreams
		//if len(pivotHistoricalFetchers) == 0 {
		//log.Error("pivot fetcher length == 0", "len", len(pivotHistoricalFetchers))
		//panic("pivot historical fetchers for node should not be empty")
		//}
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

//func checkHistoricalStreams(t *testing.T, onesCursors map[uint]uint64, onesStreams map[uint]*syncStreamFetch) {
//if len(onesCursors) == 0 {
//}
//if len(onesStreams) == 0 {
//t.Fatal("zero length cursors")
//}

//for k, v := range onesCursors {
//if v > 0 {
//// there should be a matching stream state
//if _, ok := onesStreams[k]; !ok {
//t.Fatalf("stream for bin id %d should exist", k)
//}
//} else {
//// index is zero -> no historical stream for this bin. check that it doesn't exist
//if _, ok := onesStreams[k]; ok {
//t.Fatalf("stream for bin id %d should not exist", k)
//}
//}
//}
//}

func newBzzSyncWithLocalstoreDataInsertion(numChunks int, autostartBehavior StreamInitBehavior) func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	return func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
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
		if numChunks > 0 {
			filesize := numChunks * 4096
			cctx := context.Background()
			_, wait, err := fileStore.Store(cctx, testutil.RandomReader(0, filesize), int64(filesize), false)
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
		dir, err := ioutil.TempDir("", "swarm-stream-")
		if err != nil {
			return nil, nil, err
		}
		store, err = state.NewDBStore(dir)
		if err != nil {
			return nil, nil, err
		}

		sp := NewSyncProvider(netStore, kad, autostartBehavior)
		o := NewSlipStream(store, kad, sp)
		bucket.Store(bucketKeyBinIndex, binIndexes)
		bucket.Store(bucketKeyFileStore, fileStore)
		bucket.Store(simulation.BucketKeyKademlia, kad)
		bucket.Store(bucketKeySyncer, o)

		cleanup = func() {
			localStore.Close()
			localStoreCleanup()
			store.Close()
			os.RemoveAll(dir)
		}

		return o, cleanup, nil
	}
}
