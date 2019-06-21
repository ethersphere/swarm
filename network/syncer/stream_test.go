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

func TestNodesExchangeCorrectBinIndexes(t *testing.T) {
	nodeCount := 2

	// create a standard sim
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"bzz-sync": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
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
				t.Fatal(err)
			}
			if err := wait(cctx); err != nil {
				t.Fatal(err)
			}

			// verify bins just upto 8 (given random distribution and 1000 chunks
			// bin index `i` cardinality for `n` chunks is assumed to be n/(2^i+1)
			for i := 0; i <= 7; i++ {
				if binIndex, err := netStore.LastPullSubscriptionBinID(uint8(i)); binIndex == 0 || err != nil {
					t.Fatalf("error querying bin indexes. bin %d, index %d, err %v", i, binIndex, err)
				}
			}

			binIndexes := make([]uint64, 17)
			for i := 0; i <= 16; i++ {
				binIndex, err := netStore.LastPullSubscriptionBinID(uint8(i))
				if err != nil {
					t.Fatal(err)
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
		},
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
		for i := 0; i < 2; i++ {
			idOne := nodeIDs[i]
			idOther := nodeIDs[(i+1)%2]
			onesSyncer, ok := sim.NodeItem(idOne, bucketKeySyncer)
			if !ok {
				t.Fatal("cant find item")
			}

			s := onesSyncer.(*SwarmSyncer)
			onesCursors := s.peers[idOther].streamCursors
			othersBins, ok := sim.NodeItem(idOther, bucketKeyBinIndex)
			if !ok {
				t.Fatal("cant find item")
			}

			compareNodeBinsToStreams(t, onesCursors, othersBins.([]uint64))
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
