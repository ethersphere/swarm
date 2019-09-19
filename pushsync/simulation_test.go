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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/retrieval"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
	"golang.org/x/sync/errgroup"
)

var (
	bucketKeyPushSyncer = simulation.BucketKey("pushsyncer")
	bucketKeyNetStore   = simulation.BucketKey("netstore")
)

var (
	nodeCntFlag   = flag.Int("nodes", 4, "number of nodes in simulation")
	chunkCntFlag  = flag.Int("chunks", 4, "number of chunks per upload in simulation")
	testCasesFlag = flag.Int("cases", 4, "number of concurrent upload-download cases to test in simulation")
)

// test syncer using pss
// the test
// * creates a simulation with connectivity loaded from a snapshot
// * for each test case, two nodes are chosen randomly, an uploader and a downloader
// * uploader uploads a number of chunks
// * wait until the uploaded chunks are push-synced
// * downloader has one shot to download all the chunks
// Testcases are run concurrently
func TestPushsyncSimulation(t *testing.T) {
	nodeCnt := *nodeCntFlag
	chunkCnt := *chunkCntFlag
	testcases := *testCasesFlag

	err := testPushsyncSimulation(nodeCnt, chunkCnt, testcases, newServiceFunc)
	if err != nil {
		t.Fatal(err)
	}
}

func testPushsyncSimulation(nodeCnt, chunkCnt, testcases int, sf simulation.ServiceFunc) error {
	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"pushsync": sf,
	})
	defer sim.Close()

	ctx := context.Background()
	snapCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err := sim.UploadSnapshot(snapCtx, filepath.Join("../network/stream/testdata", fmt.Sprintf("snapshot_%d.json", nodeCnt)))
	if err != nil {
		return fmt.Errorf("error while loading snapshot: %v", err)
	}

	start := time.Now()
	log.Error("Snapshot loaded. Simulation starting", "at", start)
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		var errg errgroup.Group
		for j := 0; j < testcases; j++ {
			j := j
			errg.Go(func() error {
				return uploadAndDownload(ctx, sim, nodeCnt, chunkCnt, j)
			})
		}
		if err := errg.Wait(); err != nil {
			return err
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
	defer cancel()
	err = tag.WaitTillDone(sctx, chunk.StateSynced)
	if err != nil {
		log.Debug("tag", "tag", tag)
		return fmt.Errorf("error waiting syncing: %v", err)
	}

	log.Debug("downloading", "peer", did, "chunks", chunkCnt, "tagname", tagname)
	netstore := sim.MustNodeItem(did, bucketKeyNetStore).(*storage.NetStore)
	err = download(ctx, netstore, ref)
	log.Debug("downloaded", "peer", did, "chunks", chunkCnt, "tagname", tagname, "err", err)
	return err
}

// newServiceFunc constructs a minimal service needed for a simulation test for Push Sync, namely:
// localstore, netstore, retrieval and pss. Bzz service is required on the same node.
func newServiceFunc(ctx *adapters.ServiceContext, bucket *sync.Map) (node.Service, func(), error) {
	// setup localstore
	n := ctx.Config.Node()
	addr := network.NewBzzAddrFromEnode(n)
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
	netStore := storage.NewNetStore(lstore, addr.Over(), n.ID())

	// setup pss
	k, _ := bucket.LoadOrStore(simulation.BucketKeyKademlia, network.NewKademlia(addr.Over(), network.NewKadParams()))
	kad := k.(*network.Kademlia)

	privKey, err := crypto.GenerateKey()
	pssp := pss.NewParams().WithPrivateKey(privKey)
	ps, err := pss.New(kad, pssp)
	if err != nil {
		return nil, nil, err
	}

	bucket.Store(bucketKeyNetStore, netStore)

	r := retrieval.New(kad, netStore, kad.BaseAddr(), nil)
	netStore.RemoteGet = r.RequestFromPeers

	pubSub := pss.NewPubSub(ps)
	// setup pusher
	p := NewPusher(lstore, pubSub, chunk.NewTags())
	bucket.Store(bucketKeyPushSyncer, p)

	// setup storer
	s := NewStorer(netStore, pubSub)

	cleanup := func() {
		p.Close()
		s.Close()
		netStore.Close()
		os.RemoveAll(dir)
	}

	return &RetrievalAndPss{r, ps}, cleanup, nil
}

// implements the node.Service interface
type RetrievalAndPss struct {
	retrieval *retrieval.Retrieval
	pss       *pss.Pss
}

func (s *RetrievalAndPss) APIs() []rpc.API {
	return nil
}

func (s *RetrievalAndPss) Protocols() []p2p.Protocol {
	return append(s.retrieval.Protocols(), s.pss.Protocols()...)
}

func (s *RetrievalAndPss) Start(srv *p2p.Server) error {
	err := s.retrieval.Start(srv)
	if err != nil {
		return err
	}
	return s.pss.Start(srv)
}

func (s *RetrievalAndPss) Stop() error {
	err := s.retrieval.Stop()
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
	var g errgroup.Group
	for _, addr := range addrs {
		addr := addr
		g.Go(func() error {
			_, err := store.Get(ctx, chunk.ModeGetRequest, storage.NewRequest(addr))
			log.Debug("Get", "addr", hex.EncodeToString(addr[:]), "err", err)
			return err
		})
	}
	return g.Wait()
}
