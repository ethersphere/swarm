package swarm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/storage"
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

func TestSync(t *testing.T) {
	var (
		nodeCount             = 10
		iterationCount        = 5
		fileSize       int64  = 60 * 4096
		dbCapacity     uint64 = 100
	)

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bootnode": newServiceFunc(true, dbCapacity),
		"swarm":    newServiceFunc(false, dbCapacity),
	})
	defer sim.Close()

	bootnode, err := sim.AddNode(simulation.AddNodeWithService("bootnode"))
	if err != nil {
		t.Fatal(err)
	}

	nodes, err := sim.AddNodes(nodeCount, simulation.AddNodeWithService("swarm"))
	if err != nil {
		t.Fatal(err)
	}

	if err := sim.Net.ConnectNodesStar(nodes, bootnode); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	errd := false
	for i := 1; i <= iterationCount && !errd; i++ {
		nodeIndex := 0
		fmt.Println("sync simulation start", "iteration", i, "uploadingNode", nodeIndex)

		startUpload := time.Now()
		addr, checksum := uploadRandomFile(t, sim.MustNodeItem(nodes[nodeIndex], bucketKeyAPI).(*api.API), fileSize)
		fmt.Println("sync simulation upload", "iteration", i, "upload", time.Since(startUpload), "checksum", checksum)

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
		fmt.Println("sync simulation syncing", "iteration", i, "syncing", time.Since(startSyncing)-api.InspectorIsPullSyncingTolerance)

		retrievalStart := time.Now()

		for ni, n := range nodes {
			if false {
				checkFile(t, sim.MustNodeItem(n, bucketKeyAPI).(*api.API), addr, checksum)
			}
			insp := sim.MustNodeItem(n, bucketKeyInspector).(*api.Inspector)
			si, err := insp.StorageIndices()
			if err != nil {
				t.Fatal(err)
			}

			sijson, err := json.MarshalIndent(si, "", "    ")
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("sync simulation storage indexes", "iteration", i, "node", ni, "indexes", string(sijson))

			if uint64(si["data"]) > dbCapacity {
				errd = true
				//sijson, err := json.MarshalIndent(si, "", "    ")
				//if err != nil {
				//t.Fatal(err)
				//}
				//fmt.Println("sync simulation storage indexes", "iteration", i, "node", ni, "indexes", string(sijson))
				k := sim.MustNodeItem(n, simulation.BucketKeyKademlia).(*network.Kademlia)
				fmt.Println("sync simulation kademlia", "iteration", i, "node", ni, "kademlia", k.String())
				x, err := insp.PeerStreams()
				if err != nil {
					t.Fatal(err)
				}
				xjson, err := json.MarshalIndent(x, "", "    ")
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("sync simulation subscriptions", "iteration", i, "node", ni, "subscriptions", string(xjson))

				binIDs := make(map[uint8]uint64)
				n := sim.MustNodeItem(n, bucketKeyLocalStore).(chunk.Store)
				for i := uint8(0); i <= chunk.MaxPO; i++ {
					binIDs[i], err = n.LastPullSubscriptionBinID(i)
					if err != nil {
						t.Fatal(err)
					}
				}
				binIDsjson, err := json.MarshalIndent(binIDs, "", "    ")
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("sync simulation binids", "iteration", i, "node", ni, "binids", string(binIDsjson))
			}
		}
		fmt.Println("sync simulation retrieval", "iteration", i, "retrieval", time.Since(retrievalStart))
		fmt.Println("sync simulation done", "iteration", i, "duration", time.Since(startUpload)-api.InspectorIsPullSyncingTolerance)
	}
}

func newServiceFunc(bootnode bool, dbCapacity uint64) func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	return func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
		config := api.NewConfig()

		config.BootnodeMode = bootnode
		config.DbCapacity = dbCapacity
		config.PushSyncEnabled = false
		config.SyncEnabled = true

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

		sw, err := NewSwarm(config, nil)
		if err != nil {
			return nil, cleanup, err
		}
		bucket.Store(simulation.BucketKeyKademlia, sw.bzz.Hive.Kademlia)
		bucket.Store(bucketKeyLocalStore, sw.netStore.Store)
		bucket.Store(bucketKeyAPI, sw.api)
		bucket.Store(bucketKeyInspector, sw.inspector)
		fmt.Println("new swarm", "bzzKey", config.BzzKey, "baseAddr", fmt.Sprintf("%x", sw.bzz.BaseAddr()))
		return sw, cleanup, nil
	}
}

func uploadRandomFile(t *testing.T, a *api.API, length int64) (chunk.Address, string) {
	t.Helper()

	ctx := context.Background()

	hasher := md5.New()

	key, wait, err := a.Store(
		ctx,
		io.TeeReader(io.LimitReader(random, length), hasher),
		length,
		false,
	)
	if err != nil {
		t.Fatalf("store file: %v", err)
	}

	if err := wait(ctx); err != nil {
		t.Fatalf("wait for file to be stored: %v", err)
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

func checkFile(t *testing.T, a *api.API, addr chunk.Address, checksum string) {
	t.Helper()

	r, _ := a.Retrieve(context.Background(), addr)

	hasher := md5.New()

	n, err := io.Copy(hasher, r)
	if err != nil {
		t.Fatal(err)
	}

	got := hex.EncodeToString(hasher.Sum(nil))

	if got != checksum {
		t.Fatalf("got file checksum %s (length %v), want %s", got, n, checksum)
	}
}
