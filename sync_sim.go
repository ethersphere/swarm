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

// This file is meant to be run with go run, not build into binary, so it is excluded with the build flag.

// +build ignore

package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm"
	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/storage"
)

var (
	cli = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	nodeCount           = cli.Int("nodes", 4, "number of nodes")
	iterations          = cli.Int("iterations", 100, "number upload and retrieve iterations to perform")
	fileSize            = cli.Int("file-size", 50*1024*1024, "upload file size in bytes")
	dbCapacity          = cli.Int("database-capacity", 5000000, "nodes database capacity")
	randomUploadingNode = cli.Bool("random-uploading-node", true, "pick a random node to upload file in every iteration")
	randomRetrievalNode = cli.Bool("random-retrieval-node", true, "pick a single random node to retrieve a file in every iteration")
	verbosity           = cli.Int("verbosity", 2, "verbosity of logs")
	help                = cli.Bool("help", false, "Show program usage.")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	random = rand.New(rand.NewSource(time.Now().UnixNano()))

	bucketKeyLocalStore = "localstore"
	bucketKeyAPI        = "api"
	bucketKeyInspector  = "inspector"
)

// This function runs a long running simulation to measure syncing performance and correctness.
// It can be configured with cli flags defined above.
func main() {
	if err := cli.Parse(os.Args[1:]); err != nil {
		cli.Usage()
		return
	}

	if *help {
		cli.Usage()
		return
	}

	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*verbosity), log.StreamHandler(os.Stdout, log.TerminalFormat(false))))

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bootnode": newServiceFunc(true),
		"swarm":    newServiceFunc(false),
	})
	defer sim.Close()

	bootnode, err := sim.AddNode(simulation.AddNodeWithService("bootnode"))
	if err != nil {
		fatal(err)
	}

	nodes, err := sim.AddNodes(*nodeCount, simulation.AddNodeWithService("swarm"))
	if err != nil {
		fatal(err)
	}

	if err := sim.Net.ConnectNodesStar(nodes, bootnode); err != nil {
		fatal(err)
	}

	for i := 1; i <= *iterations; i++ {
		nodeIndex := 0
		if *randomUploadingNode {
			nodeIndex = random.Intn(len(nodes))
		}
		log.Info("sync simulation start", "iteration", i, "uploadingNode", nodeIndex)

		startUpload := time.Now()
		addr, checksum := uploadRandomFile(sim.MustNodeItem(nodes[nodeIndex], bucketKeyAPI).(*api.API), int64(*fileSize))
		log.Info("sync simulation upload", "iteration", i, "upload", time.Since(startUpload), "checksum", checksum)

		startSyncing := time.Now()

		time.Sleep(1 * time.Second)

		for syncing := true; syncing; {
			time.Sleep(100 * time.Millisecond)
			syncing = false
			for _, n := range nodes {
				if sim.MustNodeItem(n, bucketKeyInspector).(*api.Inspector).IsPullSyncing() {
					syncing = true
				}
			}
		}
		log.Info("sync simulation syncing", "iteration", i, "syncing", time.Since(startSyncing)-api.InspectorIsPullSyncingTolerance)

		retrievalStart := time.Now()

		if *randomRetrievalNode {
			i := nodeIndex
			for i == nodeIndex {
				i = random.Intn(len(nodes))
			}
			checkFile(sim.MustNodeItem(nodes[i], bucketKeyAPI).(*api.API), addr, checksum)
		} else {
			for _, n := range nodes {
				checkFile(sim.MustNodeItem(n, bucketKeyAPI).(*api.API), addr, checksum)
			}
		}

		log.Info("sync simulation retrieval", "iteration", i, "retrieval", time.Since(retrievalStart))
		log.Info("sync simulation done", "iteration", i, "duration", time.Since(startUpload)-api.InspectorIsPullSyncingTolerance)
	}
}

func newServiceFunc(bootnode bool) func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	return func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
		config := api.NewConfig()

		config.BootnodeMode = bootnode
		config.DbCapacity = uint64(*dbCapacity)

		dir, err := ioutil.TempDir("", "swarm-sync-test-node")
		if err != nil {
			return nil, nil, err
		}
		cleanup = func() {
			err := os.RemoveAll(dir)
			if err != nil {
				log.Error("cleaning up swarm temp dir", "err", err)
			}
		}

		config.Path = dir

		privkey, err := crypto.GenerateKey()
		if err != nil {
			return nil, cleanup, err
		}
		nodekey, err := crypto.GenerateKey()
		if err != nil {
			return nil, cleanup, err
		}

		config.Init(privkey, nodekey)
		config.Port = ""

		sw, err := swarm.NewSwarm(config, nil)
		if err != nil {
			return nil, cleanup, err
		}
		bucket.Store(simulation.BucketKeyKademlia, sw.Bzz().Hive.Kademlia)
		bucket.Store(bucketKeyLocalStore, sw.NetStore().Store)
		bucket.Store(bucketKeyAPI, sw.API())
		bucket.Store(bucketKeyInspector, api.NewInspector(sw.API(), sw.Bzz().Hive, sw.NetStore(), sw.NewStreamer()))
		log.Info("new swarm", "bzzKey", config.BzzKey, "baseAddr", fmt.Sprintf("%x", sw.Bzz().BaseAddr()))
		return sw, cleanup, nil
	}
}

func uploadRandomFile(a *api.API, length int64) (chunk.Address, string) {
	ctx := context.Background()

	hasher := md5.New()

	key, wait, err := a.Store(
		ctx,
		io.TeeReader(io.LimitReader(random, length), hasher),
		length,
		false,
	)
	if err != nil {
		fatalf("store file: %v", err)
	}

	if err := wait(ctx); err != nil {
		fatalf("wait for file to be stored: %v", err)
	}

	return key, hex.EncodeToString(hasher.Sum(nil))
}

func storeFile(ctx context.Context, a *api.API, r io.Reader, length int64, contentType string, toEncrypt bool) (k storage.Address, wait func(context.Context) error, err error) {
	key, wait, err := a.Store(ctx, r, length, toEncrypt)
	if err != nil {
		return nil, nil, err
	}
	return key, wait, nil
}

func checkFile(a *api.API, addr chunk.Address, checksum string) {
	r, _ := a.Retrieve(context.Background(), addr)

	hasher := md5.New()

	n, err := io.Copy(hasher, r)
	if err != nil {
		fatal(err)
	}

	got := hex.EncodeToString(hasher.Sum(nil))

	if got != checksum {
		fatalf("got file checksum %s (length %v), want %s", got, n, checksum)
	}
}

func fatal(err error) {
	log.Error(err.Error())
	os.Exit(1)
}

func fatalf(s string, a ...interface{}) {
	log.Error(fmt.Sprintf(s, a...))
	os.Exit(1)
}
