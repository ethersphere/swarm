package swarm

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
)

//func TestKeys(t *testing.T) {

//a, _ := crypto.GenerateKey()
//b, _ := crypto.GenerateKey()
//c, _ := crypto.GenerateKey()

//pka := crypto.FromECDSA(a)
////pkaa := hex.EncodeToString(pka)

//pkb := crypto.FromECDSA(b)
////pkbb := hex.EncodeToString(pkb)

//pkc := crypto.FromECDSA(c)
////pkcc := hex.EncodeToString(pkc)

////t.Fatal(string(pka))
//}

func TestGlobalPinning(t *testing.T) {
	//bzzAddrs := make([]
	//a, _ := crypto.GenerateKey()
	//b, _ := crypto.GenerateKey()
	//c, _ := crypto.GenerateKey()

	//pka := crypto.FromECDSA(a)
	////pkaa := hex.EncodeToString(pka)

	//pkb := crypto.FromECDSA(b)
	////pkbb := hex.EncodeToString(pkb)

	//pkc := crypto.FromECDSA(c)
	//pkcc := hex.EncodeToString(pkc)

	//missingChunkAddress := pot.RandomAddress()

	nodes := make([]string, 3)
	pks := make([]*ecdsa.PrivateKey, 3)
	i := 0

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"swarm": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			config := api.NewConfig()
			dir, err := ioutil.TempDir("", "swarm-network-test-node")
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
			config.SyncEnabled = false
			config.PushSyncEnabled = true
			config.DisableAutoConnect = true

			pks[i] = privkey
			nodes[i] = config.BzzKey
			if i == 0 {
				config.GlobalPinner = true
			} else {
				config.RecoveryPublisher = nodes[0]
			}

			swarm, err := NewSwarm(config, nil)
			if err != nil {
				return nil, cleanup, err
			}
			bucket.Store(simulation.BucketKeyKademlia, swarm.bzz.Hive.Kademlia)
			bucket.Store(bucketKeyInspector, swarm.inspector)
			log.Info("new swarm", "bzzKey", config.BzzKey, "baseAddr", fmt.Sprintf("%x", swarm.bzz.BaseAddr()))
			i++
			return swarm, cleanup, nil
		},
	})
	defer sim.Close()

	ids, err := sim.AddNodes(3)
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(nodes)

	// connect the nodes in a certain way

	// generate a chunk address to retrieve
	chunkToRetrieve := storage.GenerateRandomChunk(1234)

	//a0 := common.FromHex(nodes[0]) // global pinner
	a1 := common.FromHex(nodes[1])
	a2 := common.FromHex(nodes[2])
	//p0 := chunk.Proximity(ch.Address(), a0)
	p1 := chunk.Proximity(chunkToRetrieve.Address(), a1)
	p2 := chunk.Proximity(chunkToRetrieve.Address(), a2)

	c := 0

	if p1 > p2 {
		// chunk is closer to node 1 than to node 2
		// 1 -> 2 -> 0
		// connect node 3 to node 1
		sim.Net.Connect(ids[0], ids[2])
		sim.Net.Connect(ids[2], ids[1])
		c = 1
	} else {
		// chunk is either equal distance from 1-2 or is closer to 2
		// 2 -> 1 -> 0
		sim.Net.Connect(ids[0], ids[1])
		sim.Net.Connect(ids[2], ids[1])
		c = 2
	}

	s := sim.Service("swarm", ids[c]).(*Swarm)
	req := storage.NewRequest(chunkToRetrieve.Address())
	_, err = s.netStore.Get(context.Background(), chunk.ModeGetRequest, req)

	// check that err is ErrNoSuitablePeer

	// create this feed chunk and store it even on the requester node
	signer := feed.NewGenericSigner(pks[0])
	topic, _ := feed.NewTopic("RECOVERY", nil)
	fd := feed.Feed{
		Topic: topic,
		User:  signer.Address(),
	}
	feedRequest := feed.NewFirstRequest(fd.Topic)
	pinnerAddr := nodes[0]
	pinnerPrefix := pinnerAddr[:2]
	target, err := hex.DecodeString(pinnerPrefix)

	t1 := trojan.Target(target)
	targets := trojan.Targets([]trojan.Target{t1})
	feedData, err := json.Marshal(targets)
	if err != nil {
		t.Fatal(err)
	}

	feedRequest.SetData(feedData)
	if err := feedRequest.Sign(signer); err != nil {
		t.Fatal(err)
	}

	_, err = s.api.FeedsUpdate(context.Background(), feedRequest)
	if err != nil {
		t.Fatal(err)
	}

	// retrieve again
	chunkk, err := s.netStore.Get(context.Background(), chunk.ModeGetRequest, req)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(chunkk.Data(), chunkToRetrieve.Address()) {
		t.Fatal("doesnt match")
	}

	//c -> b -> a (a is global pinner)
	/*
		c tries to retrieve and fails
		triggers the recovery process
		query the feed address of the recovery publisher (a)
		target is the first byte of overlay address of node a
		feed gives list of targets
		we mine the target, and then we get a chunk that we push into the push index
		pushsync picks up the chunk, sends it to b which forards to c
		c tries to retrieve the chunk
		finds it
		QED

	*/
}
