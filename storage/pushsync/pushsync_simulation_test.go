// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package pushsync

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
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

// test syncer using pss
// the test
// * creates a simulation with connectivity loaded from a snapshot
// * for each trial, two nodes are chosen randomly, an uploader and a downloader
// * uploader uploads a number of chunks
// * wait until the uploaded chunks are synced
// * downloader downloads the chunk
// Trials are run concurrently
func TestPushSyncSimulation(t *testing.T) {
	nodeCnt := 64
	chunkCnt := 32
	trials := 32
	testSyncerWithPubSub(t, nodeCnt, chunkCnt, trials, newServiceFunc)
}

func testSyncerWithPubSub(t *testing.T, nodeCnt, chunkCnt, trials int, sf simulation.ServiceFunc) {
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"streamer": sf,
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err := sim.UploadSnapshot(ctx, fmt.Sprintf("../../network/stream/testing/snapshot_%d.json", nodeCnt))
	if err != nil {
		t.Fatal(err)
	}
	log.Info("Snapshot loaded")
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		errc := make(chan error)
		for i := 0; i < trials; i++ {
			go uploadAndDownload(ctx, sim, errc, nodeCnt, chunkCnt, i)
		}
		i := 0
		for err := range errc {
			if err != nil {
				return err
			}
			i++
			if i >= trials {
				break
			}
		}
		return nil
	})
	if result.Error != nil {
		t.Fatalf("simulation error: %v", result.Error)
	}
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

func uploadAndDownload(ctx context.Context, sim *simulation.Simulation, errc chan error, nodeCnt, chunkCnt, i int) {
	// chose 2 random nodes as uploader and downloader
	u, d := pickNodes(nodeCnt)
	// setup uploader node
	uid := sim.UpNodeIDs()[u]
	val, _ := sim.NodeItem(uid, bucketKeyPushSyncer)
	p := val.(*Pusher)
	// setup downloader node
	did := sim.UpNodeIDs()[d]
	// the created tag indicates the uploader and downloader nodes
	tagname := fmt.Sprintf("tag-%v-%v-%d", label(uid[:]), label(did[:]), i)
	log.Debug("uploading", "peer", uid, "chunks", chunkCnt, "tagname", tagname)
	tag, what, err := upload(ctx, p.store.(*localstore.DB), p.tags, tagname, chunkCnt)
	if err != nil {
		select {
		case errc <- err:
		case <-ctx.Done():
			return
		}
		return
	}

	// wait till synced
	for {
		n, total, err := tag.Status(chunk.StateSynced)
		if err == nil && n == total {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Debug("synced", "peer", uid, "chunks", chunkCnt, "tagname", tagname)
	log.Debug("downloading", "peer", did, "chunks", chunkCnt, "tagname", tagname)
	val, _ = sim.NodeItem(did, bucketKeyNetStore)
	netstore := val.(*storage.NetStore)
	select {
	case errc <- download(ctx, netstore, what):
	case <-ctx.Done():
	}
	log.Debug("downloaded", "peer", did, "chunks", chunkCnt, "tagname", tagname)
}

// newServiceFunc constructs a minimal service needed for a simulation test for Push Sync, namely:
// localstore, netstore, stream and pss
func newServiceFunc(ctx *adapters.ServiceContext, bucket *sync.Map) (node.Service, func(), error) {
	// setup localstore
	n := ctx.Config.Node()
	addr := network.NewAddr(n)
	dir, err := ioutil.TempDir("", "localstore-test")
	if err != nil {
		return nil, nil, err
	}
	lstore, err := localstore.New(dir, addr.Over(), nil)
	if err != nil {
		os.RemoveAll(dir)
		return nil, nil, err
	}

	// setup pss
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(addr.Over(), kadParams)
	bucket.Store(simulation.BucketKeyKademlia, kad)

	privKey, err := crypto.GenerateKey()
	pssp := pss.NewPssParams().WithPrivateKey(privKey)
	ps, err := pss.NewPss(kad, pssp)
	if err != nil {
		return nil, nil, err
	}
	// setup netstore
	netStore := storage.NewNetStore(lstore, enode.HexID(hexutil.Encode(kad.BaseAddr())))
	// streamer
	delivery := stream.NewDelivery(kad, netStore)
	netStore.RemoteGet = delivery.RequestFromPeers

	bucket.Store(bucketKeyNetStore, netStore)

	noSyncing := &stream.RegistryOptions{Syncing: stream.SyncingDisabled, SyncUpdateDelay: 50 * time.Millisecond}
	r := stream.NewRegistry(addr.ID(), delivery, netStore, state.NewInmemoryStore(), noSyncing, nil)

	pubSub := pss.NewPubSub(ps)

	// set up syncer
	p := NewPusher(lstore, pubSub, chunk.NewTags())
	bucket.Store(bucketKeyPushSyncer, p)

	// setup storer
	s := NewStorer(netStore, pubSub, p.PushReceipt)

	cleanup := func() {
		p.Close()
		s.Close()
		netStore.Close()
		r.Close()
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
	tag, err = tags.New(tagname, int64(n))
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < n; i++ {
		ch := storage.GenerateRandomChunk(int64(chunk.DefaultSize))
		addrs = append(addrs, ch.Address())
		store.Put(ctx, chunk.ModePutUpload, ch.WithTagID(tag.Uid))
		tag.Inc(chunk.StateStored)
	}
	return tag, addrs, nil
}

func download(ctx context.Context, store *storage.NetStore, addrs []storage.Address) error {
	errc := make(chan error)
	for _, addr := range addrs {
		go func(addr storage.Address) {
			_, err := store.Get(ctx, chunk.ModeGetRequest, storage.NewRequest(addr))
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
