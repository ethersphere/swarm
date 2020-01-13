// Copyright 2018 The go-ethereum Authors
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

package swarm

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/sctx"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

var (
	printKademlia      = flag.Bool("print-kademlia", false, "prints kademlia tables before test step starts")
	waitKademlia       = flag.Bool("waitkademlia", false, "wait for healthy kademlia before checking files availability")
	bucketKeyInspector = "inspector"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	testutil.Init()
}

// TestSwarmNetwork runs a series of test simulations with
// static and dynamic Swarm nodes in network simulation, by
// uploading files to every node and retrieving them.
func TestSwarmNetwork(t *testing.T) {
	var tests = []testSwarmNetworkCase{
		{
			name: "10_nodes",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 10,
				},
			},
			options: &testSwarmNetworkOptions{
				Timeout: 45 * time.Second,
			},
		},
		{
			name: "dec_inc_node_count",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 3,
				},
				{
					nodeCount: 1,
				},
				{
					nodeCount: 5,
				},
			},
			options: &testSwarmNetworkOptions{
				Timeout: 90 * time.Second,
			},
		},
	}

	if *testutil.Longrunning {
		tests = append(tests, longRunningCases()...)
	} else if testutil.RaceEnabled {
		tests = shortCaseForRace()

	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testSwarmNetwork(t, tc.options, tc.steps...)
		})
	}
}

type testSwarmNetworkCase struct {
	name    string
	steps   []testSwarmNetworkStep
	options *testSwarmNetworkOptions
}

// testSwarmNetworkStep is the configuration
// for the state of the simulation network.
type testSwarmNetworkStep struct {
	// number of swarm nodes that must be in the Up state
	nodeCount int
}

// testSwarmNetworkOptions contains optional parameters for running
// testSwarmNetwork.
type testSwarmNetworkOptions struct {
	Timeout time.Duration
}

func longRunningCases() []testSwarmNetworkCase {
	return []testSwarmNetworkCase{
		{
			name: "50_nodes",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 50,
				},
			},
			options: &testSwarmNetworkOptions{
				Timeout: 3 * time.Minute,
			},
		},
		{
			name: "inc_node_count",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 2,
				},
				{
					nodeCount: 5,
				},
				{
					nodeCount: 10,
				},
			},
			options: &testSwarmNetworkOptions{
				Timeout: 90 * time.Second,
			},
		},
		{
			name: "dec_node_count",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 10,
				},
				{
					nodeCount: 6,
				},
				{
					nodeCount: 3,
				},
			},
			options: &testSwarmNetworkOptions{
				Timeout: 90 * time.Second,
			},
		},
		{
			name: "inc_dec_node_count",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 3,
				},
				{
					nodeCount: 5,
				},
				{
					nodeCount: 25,
				},
				{
					nodeCount: 10,
				},
				{
					nodeCount: 4,
				},
			},
			options: &testSwarmNetworkOptions{
				Timeout: 5 * time.Minute,
			},
		},
	}
}

func shortCaseForRace() []testSwarmNetworkCase {
	// As for now, Travis with -race can only run 8 nodes
	return []testSwarmNetworkCase{
		{
			name: "8_nodes",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 8,
				},
			},
			options: &testSwarmNetworkOptions{
				Timeout: 1 * time.Minute,
			},
		},
	}
}

// file represents the file uploaded on a particular node.
type file struct {
	addr   storage.Address
	data   string
	nodeID enode.ID
}

// check represents a reference to a file that is retrieved
// from a particular node.
type check struct {
	key    string
	nodeID enode.ID
}

// testSwarmNetwork is a helper function used for testing different
// static and dynamic Swarm network simulations.
// It is responsible for:
//  - Setting up a Swarm network simulation, and updates the number of nodes within the network on every step according to steps.
//  - Uploading a unique file to every node on every step.
//  - May wait for Kademlia on every node to be healthy.
//  - Checking if a file is retrievable from all nodes.
func testSwarmNetwork(t *testing.T, o *testSwarmNetworkOptions, steps ...testSwarmNetworkStep) {
	t.Helper()

	if o == nil {
		o = new(testSwarmNetworkOptions)
	}

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

			swarm, err := NewSwarm(config, nil)
			if err != nil {
				return nil, cleanup, err
			}
			bucket.Store(simulation.BucketKeyKademlia, swarm.bzz.Hive.Kademlia)
			bucket.Store(bucketKeyInspector, swarm.inspector)
			log.Info("new swarm", "bzzKey", config.BzzKey, "baseAddr", fmt.Sprintf("%x", swarm.bzz.BaseAddr()))
			return swarm, cleanup, nil
		},
	})
	defer sim.Close()

	ctx := context.Background()
	if o.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.Timeout)
		defer cancel()
	}

	files := make([]file, 0)

	for i, step := range steps {
		log.Debug("test sync step", "n", i+1, "nodes", step.nodeCount)

		change := step.nodeCount - len(sim.UpNodeIDs())

		if change > 0 {
			_, err := sim.AddNodesAndConnectChain(change)
			if err != nil {
				t.Fatal(err)
			}
		} else if change < 0 {
			_, err := sim.StopRandomNodes(-change)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			t.Logf("step %v: no change in nodes", i)
			continue
		}

		for {
			ids := sim.UpNodeIDs()
			if len(ids) == step.nodeCount {
				break
			}
			log.Info("test run waiting for node count to normalise. sleeping 1 second", "count", len(ids), "want", step.nodeCount)
			time.Sleep(1 * time.Second)
		}

		result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
			nodeIDs := sim.UpNodeIDs()
			rand.Shuffle(len(nodeIDs), func(i, j int) {
				nodeIDs[i], nodeIDs[j] = nodeIDs[j], nodeIDs[i]
			})

			if *waitKademlia {
				if _, err := sim.WaitTillHealthy(ctx); err != nil {
					return err
				}
			}

			if *printKademlia {
				for _, id := range nodeIDs {
					swarm := sim.Service("swarm", id).(*Swarm)
					log.Debug("node kademlias", "node", id.String())
					fmt.Println(swarm.bzz.Hive.String())
				}
			}

			for _, id := range nodeIDs {
				key, data, err := uploadFile(sim.Service("swarm", id).(*Swarm))
				if err != nil {
					return err
				}
				log.Trace("file uploaded", "node", id, "key", key.String())
				files = append(files, file{
					addr:   key,
					data:   data,
					nodeID: id,
				})
			}

			for syncing := true; syncing; {
				syncing = false
				time.Sleep(1 * time.Second)

				for _, id := range nodeIDs {
					if sim.MustNodeItem(id, bucketKeyInspector).(*api.Inspector).IsPullSyncing() {
						syncing = true
						break
					}
				}
			}

			for {
				// File retrieval check is repeated until all uploaded files are retrieved from all nodes
				// or until the timeout is reached.
				if missing := retrieveF(t, sim, files); missing == 0 {
					log.Debug("test step finished with no files missing")
					return nil
				} else {
					t.Logf("retry retrieve. missing %d", missing)
					log.Error("retrying retrieve", "missing", missing, "files", len(files))
					time.Sleep(1 * time.Second)
				}
			}
		})

		if result.Error != nil {
			t.Fatal(result.Error)
		}
		log.Debug("done: test sync step", "n", i+1, "nodes", step.nodeCount)
	}
}

// uploadFile, uploads a short file to the swarm instance
// using the api.Put method.
func uploadFile(swarm *Swarm) (storage.Address, string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return nil, "", err
	}
	// File data is very short, but it is ensured that its
	// uniqueness is very certain.
	data := fmt.Sprintf("test content %s %x", time.Now().Round(0), b)
	ctx := context.TODO()
	k, wait, err := putString(ctx, swarm.api, data, "text/plain", false)
	if err != nil {
		return nil, "", err
	}
	if wait != nil {
		err = wait(ctx)
	}
	return k, data, err
}

// retrieveF is the function that is used for checking the availability of
// uploaded files in testSwarmNetwork test helper function.
func retrieveF(
	t *testing.T,
	sim *simulation.Simulation,
	files []file,
) (missing uint64) {
	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	nodeIDs := sim.UpNodeIDs()

	for _, id := range nodeIDs {
		missing := 0

		swarm := sim.Service("swarm", id).(*Swarm)

		for _, f := range files {
			log.Debug("api get: check file", "node", id.String(), "key", f.addr.String())

			r, _, _, _, err := swarm.api.Get(context.TODO(), api.NOOPDecrypt, f.addr, "/")
			if err != nil {
				t.Logf("api get - node cannot get key: node %s, key %s, kademlia %s: %v", id, f.addr, swarm.bzz.Hive, err)
				missing++
				continue
			}
			d, err := ioutil.ReadAll(r)
			if err != nil {
				t.Logf("api get - error read response: node %s, key %s: kademlia %s: %v", id, f.addr, swarm.bzz.Hive, err)
				missing++
				continue
			}
			data := string(d)
			if data != f.data {
				missing++
				t.Logf("api get - file content missmatch: node %s, key %s, expected %q, got %q", id, f.addr, f.data, data)
				continue
			}
		}
	}

	return missing
}

// putString provides singleton manifest creation on top of api.API
func putString(ctx context.Context, a *api.API, content string, contentType string, toEncrypt bool) (k storage.Address, wait func(context.Context) error, err error) {
	r := strings.NewReader(content)
	tag, err := a.Tags.Create("unnamed-tag", 0, false)

	log.Trace("created new tag", "uid", tag.Uid)

	ctx = sctx.SetTag(ctx, tag.Uid)
	key, waitContent, err := a.Store(ctx, r, int64(len(content)), toEncrypt)
	if err != nil {
		return nil, nil, err
	}
	manifest := fmt.Sprintf(`{"entries":[{"hash":"%v","contentType":"%s"}]}`, key, contentType)
	r = strings.NewReader(manifest)
	key, waitManifest, err := a.Store(ctx, r, int64(len(manifest)), toEncrypt)
	if err != nil {
		return nil, nil, err
	}
	tag.DoneSplit(key)
	return key, func(ctx context.Context) error {
		err := waitContent(ctx)
		if err != nil {
			return err
		}
		return waitManifest(ctx)
	}, nil
}
