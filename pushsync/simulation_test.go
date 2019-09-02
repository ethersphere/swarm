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

package pushsync

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/network/stream"
	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
)

var (
	bucketKeyPushSyncer = simulation.BucketKey("pushsyncer")
	bucketKeyNetStore   = simulation.BucketKey("netstore")
)

var (
	nodeCntFlag   = flag.Int("nodes", 16, "number of nodes in simulation")
	chunkCntFlag  = flag.Int("chunks", 16, "number of chunks per upload in simulation")
	testCasesFlag = flag.Int("cases", 16, "number of concurrent upload-download cases to test in simulation")
)

// test syncer using pss
// the test
// * creates a simulation with connectivity loaded from a snapshot
// * for each test case, two nodes are chosen randomly, an uploader and a downloader
// * uploader uploads a number of chunks
// * wait until the uploaded chunks are synced
// * downloader downloads the chunk
// Testcases are run concurrently
func TestPushsyncSimulation(t *testing.T) {
	nodeCnt := *nodeCntFlag
	chunkCnt := *chunkCntFlag
	testcases := *testCasesFlag

	err := testSyncerWithPubSub(nodeCnt, chunkCnt, testcases, newServiceFunc)
	if err != nil {
		t.Fatal(err)
	}
}

func testSyncerWithPubSub(nodeCnt, chunkCnt, testcases int, sf simulation.ServiceFunc) error {
	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"pushsync": sf,
	})
	defer sim.Close()

	ctx := context.Background()
	snapCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err := sim.UploadSnapshot(snapCtx, filepath.Join("../network/stream/testing", fmt.Sprintf("snapshot_%d.json", nodeCnt)))
	if err != nil {
		return fmt.Errorf("error while loading snapshot: %v", err)
	}

	start := time.Now()
	log.Info("Snapshot loaded. Simulation starting", "at", start)
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		errc := make(chan error)
		for j := 0; j < testcases; j++ {
			j := j
			go func() {
				err := uploadAndDownload(ctx, sim, nodeCnt, chunkCnt, j)
				select {
				case errc <- err:
				case <-ctx.Done():
				}
			}()
		}
		i := 0
		for err := range errc {
			if err != nil {
				return err
			}
			i++
			if i >= testcases {
				return nil
			}
		}
		return nil
	})

	if result.Error != nil {
		return fmt.Errorf("simulation error: %v", result.Error)
	}
	log.Error("simulation", "duration", time.Since(start))
	return nil
}

// pickNodes selects 2 distinct
func pickNodes(n int) (i, j int) {
	i = rand.Intn(n)
	j = rand.Intn(n - 1)
	if j >= i {
		j++
	}
	return
}

func uploadAndDownload(ctx context.Context, sim *simulation.Simulation, nodeCnt, chunkCnt, i int) error {
	// chose 2 random nodes as uploader and downloader
	u, d := pickNodes(nodeCnt)
	// setup uploader node
	uid := sim.UpNodeIDs()[u]
	p := sim.MustNodeItem(uid, bucketKeyPushSyncer).(*Pusher)
	// setup downloader node
	did := sim.UpNodeIDs()[d]
	// the created tag indicates the uploader and downloader nodes
	tagname := fmt.Sprintf("tag-%v-%v-%d", label(uid[:]), label(did[:]), i)
	log.Debug("uploading", "peer", uid, "chunks", chunkCnt, "tagname", tagname)
	tag, ref, err := upload(ctx, p.store.(*localstore.DB), p.tags, tagname, chunkCnt)
	if err != nil {
		return err
	}
	log.Debug("uploaded", "peer", uid, "chunks", chunkCnt, "tagname", tagname)

	// wait till pushsync is done
	syncTimeout := 30 * time.Second
	sctx, cancel := context.WithTimeout(ctx, syncTimeout)
	err = tag.WaitTillDone(sctx, chunk.StateSynced)
	if err != nil {
		log.Debug("tag", "tag", tag)
		cancel()
		return fmt.Errorf("error waiting syncing: %v", err)
	}
	cancel()

	log.Debug("downloading", "peer", did, "chunks", chunkCnt, "tagname", tagname)
	netstore := sim.MustNodeItem(did, bucketKeyNetStore).(*storage.NetStore)
	err = download(ctx, netstore, ref)
	log.Debug("downloaded", "peer", did, "chunks", chunkCnt, "tagname", tagname, "err", err)
	return err
}

// newServiceFunc constructs a minimal service needed for a simulation test for Push Sync, namely:
// localstore, netstore, stream and pss
func newServiceFunc(ctx *adapters.ServiceContext, bucket *sync.Map) (node.Service, func(), error) {
	// setup localstore
	n := ctx.Config.Node()
	addr := network.NewAddr(n)
	dir, err := ioutil.TempDir("", "pushsync-test")
	if err != nil {
		return nil, nil, err
	}
	lstore, err := localstore.New(dir, addr.Over(), nil)
	if err != nil {
		os.RemoveAll(dir)
		return nil, nil, err
	}
	// setup netstore
	netStore := storage.NewNetStore(lstore, n.ID())

	// setup pss
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(addr.Over(), kadParams)
	bucket.Store(simulation.BucketKeyKademlia, kad)

	privKey, err := crypto.GenerateKey()
	pssp := pss.NewParams().WithPrivateKey(privKey)
	ps, err := pss.New(kad, pssp)
	if err != nil {
		return nil, nil, err
	}

	// // streamer for retrieval
	delivery := stream.NewDelivery(kad, netStore)
	netStore.RemoteGet = delivery.RequestFromPeers

	bucket.Store(bucketKeyNetStore, netStore)

	// set up syncer
	noSyncing := &stream.RegistryOptions{Syncing: stream.SyncingDisabled, SyncUpdateDelay: 50 * time.Millisecond}
	r := stream.NewRegistry(addr.ID(), delivery, netStore, state.NewInmemoryStore(), noSyncing, nil)

	pubSub := pss.NewPubSub(ps)
	// setup pusher
	p := NewPusher(lstore, pubSub, chunk.NewTags())
	bucket.Store(bucketKeyPushSyncer, p)

	// setup storer
	s := NewStorer(netStore, pubSub)

	cleanup := func() {
		p.Close()
		s.Close()
		r.Close()
		netStore.Close()
		os.RemoveAll(dir)
	}

	return &StreamerAndPss{r, ps}, cleanup, nil
}

// implements the node.Service interface
type StreamerAndPss struct {
	*stream.Registry
	pss *pss.Pss
}

func (s *StreamerAndPss) Protocols() []p2p.Protocol {
	return append(s.Registry.Protocols(), s.pss.Protocols()...)
}

func (s *StreamerAndPss) Start(srv *p2p.Server) error {
	err := s.Registry.Start(srv)
	if err != nil {
		return err
	}
	return s.pss.Start(srv)
}

func (s *StreamerAndPss) Stop() error {
	err := s.Registry.Stop()
	if err != nil {
		return err
	}
	return s.pss.Stop()
}

func upload(ctx context.Context, store Store, tags *chunk.Tags, tagname string, n int) (tag *chunk.Tag, addrs []storage.Address, err error) {
	tag, err = tags.Create(ctx, tagname, int64(n))
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < n; i++ {
		ch := storage.GenerateRandomChunk(int64(chunk.DefaultSize))
		addrs = append(addrs, ch.Address())
		_, err := store.Put(ctx, chunk.ModePutUpload, ch.WithTagID(tag.Uid))
		if err != nil {
			return nil, nil, err
		}
		tag.Inc(chunk.StateStored)
	}
	return tag, addrs, nil
}

func download(ctx context.Context, store *storage.NetStore, addrs []storage.Address) error {
	errc := make(chan error)
	for _, addr := range addrs {
		go func(addr storage.Address) {
			_, err := store.Get(ctx, chunk.ModeGetRequest, storage.NewRequest(addr))
			log.Debug("Get", "addr", hex.EncodeToString(addr[:]), "err", err)
			select {
			case errc <- err:
			case <-ctx.Done():
			}
		}(addr)
	}
	i := 0
	for err := range errc {
		if err != nil {
			return err
		}
		i++
		if i == len(addrs) {
			break
		}
	}
	return nil
}
