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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
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
	"github.com/ethersphere/swarm/network/retrieval"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/mock"
)

var (
	loglevel = flag.Int("loglevel", 4, "verbosity of logs")
	update   = flag.Bool("update", false, "Update golden files in testdata directory")
)

func init() {
	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

var (
	serviceNameStream          = "bzz-stream"
	bucketKeyFileStore         = "filestore"
	bucketKeyLocalStore        = "localstore"
	bucketKeyInitialBinIndexes = "bin-indexes"

	simContextTimeout = 90 * time.Second
)

func nodeRegistry(sim *simulation.Simulation, id enode.ID) (s *Registry) {
	return sim.Service(serviceNameStream, id).(*Registry)
}

func nodeFileStore(sim *simulation.Simulation, id enode.ID) (s *storage.FileStore) {
	return sim.MustNodeItem(id, bucketKeyFileStore).(*storage.FileStore)
}

func nodeInitialBinIndexes(sim *simulation.Simulation, id enode.ID) (s []uint64) {
	return sim.MustNodeItem(id, bucketKeyInitialBinIndexes).([]uint64)
}

func nodeKademlia(sim *simulation.Simulation, id enode.ID) (k *network.Kademlia) {
	return sim.MustNodeItem(id, simulation.BucketKeyKademlia).(*network.Kademlia)
}

func nodeBinIndexes(t *testing.T, store interface {
	LastPullSubscriptionBinID(bin uint8) (id uint64, err error)
}) []uint64 {
	t.Helper()

	binIndexes := make([]uint64, 17)
	for i := 0; i <= 16; i++ {
		binIndex, err := store.LastPullSubscriptionBinID(uint8(i))
		if err != nil {
			t.Fatal(err)
		}
		binIndexes[i] = binIndex
	}
	return binIndexes
}

type SyncSimServiceOptions struct {
	InitialChunkCount     uint64
	SyncOnlyWithinDepth   bool
	StreamConstructorFunc func(state.Store, []byte, ...StreamProvider) node.Service
}

func newSyncSimServiceFunc(o *SyncSimServiceOptions) func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	if o == nil {
		o = new(SyncSimServiceOptions)
	}
	if o.StreamConstructorFunc == nil {
		o.StreamConstructorFunc = func(s state.Store, b []byte, p ...StreamProvider) node.Service {
			return New(s, b, p...)
		}
	}
	return func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
		n := ctx.Config.Node()
		addr := network.NewAddr(n)

		localStore, localStoreCleanup, err := newTestLocalStore(n.ID(), addr, nil)
		if err != nil {
			return nil, nil, err
		}

		var kad *network.Kademlia

		// check if another kademlia already exists and load it if necessary - we dont want two independent copies of it
		if kv, ok := bucket.Load(simulation.BucketKeyKademlia); ok {
			kad = kv.(*network.Kademlia)
		} else {
			kad = network.NewKademlia(addr.Over(), network.NewKadParams())
			bucket.Store(simulation.BucketKeyKademlia, kad)
		}

		netStore := storage.NewNetStore(localStore, kad.BaseAddr(), n.ID())
		lnetStore := storage.NewLNetStore(netStore)
		fileStore := storage.NewFileStore(lnetStore, storage.NewFileStoreParams(), chunk.NewTags())
		bucket.Store(bucketKeyFileStore, fileStore)
		bucket.Store(bucketKeyLocalStore, localStore)

		ret := retrieval.New(kad, netStore, kad.BaseAddr())
		netStore.RemoteGet = ret.RequestFromPeers

		if o.InitialChunkCount > 0 {
			_, err := uploadChunks(context.Background(), localStore, o.InitialChunkCount)
			if err != nil {
				return nil, nil, err
			}
			binIndexes := make([]uint64, 17)
			for i := uint8(0); i <= 16; i++ {
				binIndex, err := localStore.LastPullSubscriptionBinID(i)
				if err != nil {
					return nil, nil, err
				}
				binIndexes[i] = binIndex
			}
			bucket.Store(bucketKeyInitialBinIndexes, binIndexes)
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

		sp := NewSyncProvider(netStore, kad, true, o.SyncOnlyWithinDepth)
		ss := o.StreamConstructorFunc(store, addr.Over(), sp)

		cleanup = func() {
			//ss.Stop() // wait for handlers to finish before closing localstore
			localStore.Close()
			localStoreCleanup()
			store.Close()
			os.RemoveAll(dir)
		}

		return ss, cleanup, nil
	}
}

func newTestLocalStore(id enode.ID, addr *network.BzzAddr, globalStore mock.GlobalStorer) (localStore *localstore.DB, cleanup func(), err error) {
	dir, err := ioutil.TempDir(tmpDir, "localstore-")
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

func parseID(str string) ID {
	v := strings.Split(str, "|")
	if len(v) != 2 {
		panic("too short")
	}
	return NewID(v[0], v[1])
}

func uploadChunks(ctx context.Context, store chunk.Store, count uint64) (chunks []chunk.Address, err error) {
	for i := uint64(0); i < count; i++ {
		c := storage.GenerateRandomChunk(4096)
		exists, err := store.Put(ctx, chunk.ModePutUpload, c)
		if err != nil {
			return nil, err
		}
		if exists[0] {
			return nil, errors.New("generated already existing chunk")
		}
		chunks = append(chunks, c.Address())
	}
	return chunks, nil
}

func mustUploadChunks(ctx context.Context, t testing.TB, store chunk.Store, count uint64) (chunks []chunk.Address) {
	t.Helper()

	chunks, err := uploadChunks(ctx, store, count)
	if err != nil {
		t.Fatal(err)
	}
	return chunks
}

// Test run global tmp dir. Please, use it as the first argument
// to ioutil.TempDir function calls in this package tests.
var tmpDir string

func TestMain(m *testing.M) {
	// Remove the sync init delay in tests.
	defer func(b time.Duration) { SyncInitBackoff = b }(SyncInitBackoff)
	SyncInitBackoff = 0

	// Tests in this package generate a lot of temporary directories
	// that may not be removed if tests are interrupted with SIGINT.
	// This function constructs a single top-level directory to be used
	// to store all data from a test execution. It removes the
	// tmpDir with defer, or by catching keyboard interrupt signal,
	// so that all data will be removed even on forced termination.

	var err error
	tmpDir, err = ioutil.TempDir("", "swarm-stream-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Stop(c)

	go func() {
		first := true
		for range c {
			fmt.Fprintln(os.Stderr, "signal: interrupt")
			if first {
				fmt.Fprintln(os.Stderr, "removing swarm stream tmp directory", tmpDir)
				os.RemoveAll(tmpDir)
				os.Exit(1)
			}
		}
	}()
	os.Exit(m.Run())
}

// syncPauser implements pauser interface used only in tests.
type syncPauser struct {
	c   *sync.Cond
	cMu sync.Mutex
	mu  sync.RWMutex
}

func (p *syncPauser) pause() {
	p.mu.Lock()
	if p.c == nil {
		p.c = sync.NewCond(&p.cMu)
	}
	p.mu.Unlock()
}

func (p *syncPauser) resume() {
	p.c.L.Lock()
	p.c.Broadcast()
	p.c.L.Unlock()
	p.mu.Lock()
	p.c = nil
	p.mu.Unlock()
}

func (p *syncPauser) wait() {
	p.mu.RLock()
	if p.c != nil {
		p.c.L.Lock()
		p.c.Wait()
		p.c.L.Unlock()
	}
	p.mu.RUnlock()
}
