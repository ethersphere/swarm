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

package retrieve

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/mock"
	"github.com/ethersphere/swarm/testutil"
)

var (
	loglevel           = flag.Int("loglevel", 5, "verbosity of logs")
	bucketKeyFileStore = simulation.BucketKey("filestore")
	bucketKeyBinIndex  = simulation.BucketKey("bin-indexes")
	bucketKeySyncer    = simulation.BucketKey("syncer")
)

func init() {
	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

// TestChunkDelivery brings up two nodes, stores a few chunks on the first node, then tries to retrieve them through the second node
func TestChunkDelivery(t *testing.T) {
	nodeCount := 2
	chunkCount := 10

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-retrieve": newBzzRetrieveWithLocalstore,
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

		item := sim.NodeItem(sim.UpNodeIDs()[0], bucketKeyFileStore)

		log.Debug("subscriptions on all bins exist between the two nodes, proceeding to check bin indexes")
		log.Debug("uploader node", "enode", nodeIDs[0])
		item = sim.NodeItem(nodeIDs[0], bucketKeyFileStore)
		store := item.(chunk.Store)

		//put some data into just the first node
		filesize := chunkCount * 4096
		cctx := context.Background()
		_, wait, err := item.(*storage.FileStore).Store(cctx, testutil.RandomReader(0, filesize), int64(filesize), false)
		if err != nil {
			return err
		}
		if err := wait(cctx); err != nil {
			return err
		}

		id, err := sim.AddNodes(1)
		if err != nil {
			return err
		}
		err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
		if err != nil {
			return err
		}
		nodeIDs = sim.UpNodeIDs()
		syncingNodeId := nodeIDs[1]

		uploaderNodeBinIDs := make([]uint64, 17)

		log.Debug("checking pull subscription bin ids")
		var uploaderSum uint64
		for po := 0; po <= 16; po++ {
			until, err := store.LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			log.Debug("uploader node got bin index", "bin", po, "binIndex", until)

			uploaderNodeBinIDs[po] = until
			uploaderSum += until
		}

		// check that the sum of bin indexes is equal
		log.Debug("compare to", "enode", syncingNodeId)
		//waitChunks(t, sim.NodeItem(syncingNodeId, bucketKeyFileStore).(chunk.Store), uploaderSum, 10*time.Second)

		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

func newBzzRetrieveWithLocalstore(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
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

	var store *state.DBStore
	// Use on-disk DBStore to reduce memory consumption in race tests.
	dir, err := ioutil.TempDir("", "statestore-")
	if err != nil {
		return nil, nil, err
	}
	store, err = state.NewDBStore(dir)
	if err != nil {
		return nil, nil, err
	}

	r := NewRetrieval(enode.ID{}, kad, netStore)
	bucket.Store(bucketKeyFileStore, fileStore)
	bucket.Store(simulation.BucketKeyKademlia, kad)

	cleanup = func() {
		localStore.Close()
		localStoreCleanup()
		store.Close()
		os.RemoveAll(dir)
	}

	return r, cleanup, nil
}

func newTestLocalStore(id enode.ID, addr *network.BzzAddr, globalStore mock.GlobalStorer) (localStore *localstore.DB, cleanup func(), err error) {
	dir, err := ioutil.TempDir("", "localstore-")
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() {
		os.RemoveAll(dir)
	}

	var mockStore *mock.NodeStore
	if globalStore != nil {
		mockStore = globalStore.NewNodeStore(common.BytesToAddress(id.Bytes()))
	}

	localStore, err = localstore.New(dir, addr.Over(), &localstore.Options{
		MockStore: mockStore,
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return localStore, cleanup, nil
}

func getAllRefs(testData []byte) (storage.AddressCollection, error) {
	datadir, err := ioutil.TempDir("", "chunk-debug")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir: %v", err)
	}
	defer os.RemoveAll(datadir)
	fileStore, cleanup, err := storage.NewLocalFileStore(datadir, make([]byte, 32), chunk.NewTags())
	if err != nil {
		return nil, err
	}
	defer cleanup()

	reader := bytes.NewReader(testData)
	return fileStore.GetAllReferences(context.Background(), reader, false)
}
