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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/mock"
	"github.com/ethersphere/swarm/testutil"
)

var (
	loglevel = flag.Int("loglevel", 2, "verbosity of logs")
)

func init() {
	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}
func newTestLocalStore(id enode.ID, addr *network.BzzAddr, globalStore mock.GlobalStorer) (localStore *localstore.DB, cleanup func(), err error) {
	dir, err := ioutil.TempDir("", "swarm-stream-")
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
