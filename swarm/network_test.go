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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/storage"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 4, "verbosity of logs")
)

func init() {
	rand.Seed(time.Now().UnixNano())

	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

// TestSwarmNetwork runs a series of test simulations with
// static and dynamic Swarm nodes in network simulation, by
// uploading files to every node and retreiving them.
func TestSwarmNetwork(t *testing.T) {
	for _, tc := range []struct {
		name  string
		steps []testSwarmNetworkStep
	}{
		{
			name: "10_nodes",
			steps: []testSwarmNetworkStep{
				{
					nodeCount: 10,
				},
			},
		},
		// {
		// 	name: "100_nodes",
		// 	steps: []testSwarmNetworkStep{
		// 		{
		// 			nodeCount: 100,
		// 		},
		// 	},
		// },
		// This test fails sometimes.
		// {
		// 	name: "inc_node_count",
		// 	steps: []testSwarmNetworkStep{
		// 		{
		// 			nodeCount: 3,
		// 		},
		// 		{
		// 			nodeCount: 5,
		// 		},
		// 		{
		// 			nodeCount: 10,
		// 		},
		// 	},
		// },
		// {
		// 	name: "inc_node_count",
		// 	steps: []testSwarmNetworkStep{
		// 		{
		// 			nodeCount: 15,
		// 		},
		// 		{
		// 			nodeCount: 30,
		// 		},
		// 		{
		// 			nodeCount: 50,
		// 		},
		// 	},
		// },
	} {
		t.Run(tc.name, func(t *testing.T) {
			testSwarmNetwork(t, tc.steps...)
		})
	}
}

type testSwarmNetworkStep struct {
	nodeCount int
	timeout   time.Duration
}

type file struct {
	key    storage.Key
	data   string
	nodeID discover.NodeID
}

type check struct {
	key    string
	nodeID discover.NodeID
}

func testSwarmNetwork(t *testing.T, steps ...testSwarmNetworkStep) {
	dir, err := ioutil.TempDir("", "swarm-network-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	swarms := make(map[discover.NodeID]*Swarm)
	files := make([]file, 0)

	services := map[string]adapters.ServiceFunc{
		"swarm": func(ctx *adapters.ServiceContext) (node.Service, error) {
			config := api.NewConfig()
			config.PssEnabled = false

			dir, err := ioutil.TempDir(dir, "node")
			if err != nil {
				return nil, err
			}

			config.Path = dir

			privkey, err := crypto.GenerateKey()
			if err != nil {
				return nil, err
			}

			config.Init(privkey)
			s, err := NewSwarm(nil, nil, config, nil)
			if err != nil {
				return nil, err
			}
			log.Info("new swarm", "bzzKey", config.BzzKey, "baseAddr", fmt.Sprintf("%x", s.bzz.BaseAddr()))
			swarms[ctx.Config.ID] = s
			return s, nil
		},
	}

	a := adapters.NewSimAdapter(services)
	net := simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "swarm",
	})
	defer net.Shutdown()

	trigger := make(chan discover.NodeID)

	sim := simulations.NewSimulation(net)

	for i, step := range steps {
		log.Debug("test sync step", "n", i+1, "nodes", step.nodeCount)

		change := step.nodeCount - len(allNodeIDs(net))

		if change > 0 {
			_, err := addNodes(change, net, trigger)
			if err != nil {
				t.Fatal(err)
			}
		} else if change < 0 {
			err := removeNodes(-change, net)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			t.Logf("step %v: no change in nodes", i)
			continue
		}

		nodeIDs := allNodeIDs(net)
		shuffle(len(nodeIDs), func(i, j int) {
			nodeIDs[i], nodeIDs[j] = nodeIDs[j], nodeIDs[i]
		})
		for _, id := range nodeIDs {
			key, data, err := uploadFile(swarms[id])
			if err != nil {
				t.Fatal(err)
			}
			log.Trace("file uploaded", "node", id, "key", key.String())
			files = append(files, file{
				key:    key,
				data:   data,
				nodeID: id,
			})
		}

		// Currently, we need to wait before the syncing stars.
		// If checks are started before, missing chunk errors are very frequent.
		// This needs to be fixed.
		//time.Sleep(20 * time.Second)

		// nIDs := allNodeIDs(net)
		// addrs := make([][]byte, len(nIDs))
		// for i, id := range nIDs {
		// 	addrs[i] = swarms[id].bzz.BaseAddr()
		// }
		// ppmap := network.NewPeerPotMap(2, addrs)

		timeout := step.timeout
		if timeout == 0 {
			timeout = 5 * time.Minute
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		var checkStatusM sync.Map
		var nodeStatusM sync.Map
		var totalFoundCount uint64

		result := sim.Run(ctx, &simulations.Step{
			Action: func(ctx context.Context) error {
				// ticker := time.NewTicker(200 * time.Millisecond)
				// defer ticker.Stop()

				// for range ticker.C {
				// 	healthy := true
				// 	log.Debug("kademlia health check", "node count", len(nIDs), "addr count", len(addrs))
				// 	for i, id := range nIDs {
				// 		swarm := swarms[id]
				// 		//PeerPot for this node
				// 		addr := common.Bytes2Hex(swarm.bzz.BaseAddr())
				// 		pp := ppmap[addr]
				// 		//call Healthy RPC
				// 		h := swarm.bzz.Healthy(pp)
				// 		//print info
				// 		log.Debug(swarm.bzz.String())
				// 		log.Debug("kademlia", "empty bins", pp.EmptyBins, "gotNN", h.GotNN, "knowNN", h.KnowNN, "full", h.Full)
				// 		log.Debug("kademlia", "health", h.GotNN && h.KnowNN && h.Full, "addr", fmt.Sprintf("%x", swarm.bzz.BaseAddr()), "id", id, "i", i)
				// 		log.Debug("kademlia", "ill condition", !h.GotNN || !h.Full, "addr", fmt.Sprintf("%x", swarm.bzz.BaseAddr()), "id", id, "i", i)
				// 		if !h.GotNN || !h.Full {
				// 			healthy = false
				// 			// fmt.Printf("ADDRESSES \"%x\",\n", addrs[i])
				// 			// for j, a := range addrs {
				// 			// 	if i == j {
				// 			// 		continue
				// 			// 	}
				// 			// 	fmt.Printf("\"%x\",\n", a)
				// 			// }
				// 			break
				// 		}
				// 	}
				// 	if healthy {
				// 		break
				// 	}
				// }

				go func() {
					for {
						if retrieve(net, files, swarms, trigger, &checkStatusM, &nodeStatusM, &totalFoundCount) == 0 {
							return
						}
					}
				}()
				return nil
			},
			Trigger: trigger,
			Expect: &simulations.Expectation{
				Nodes: allNodeIDs(net),
				Check: func(ctx context.Context, id discover.NodeID) (bool, error) {
					return true, nil
				},
			},
		})
		if result.Error != nil {
			t.Fatal(result.Error)
		}
		log.Debug("done: test sync step", "n", i+1, "nodes", step.nodeCount)
	}
}

func allNodeIDs(net *simulations.Network) (nodes []discover.NodeID) {
	for _, n := range net.GetNodes() {
		if n.Up {
			nodes = append(nodes, n.ID())
		}
	}
	return
}

func addNodes(count int, net *simulations.Network, trigger chan discover.NodeID) (ids []discover.NodeID, err error) {
	for i := 0; i < count; i++ {
		nodeIDs := allNodeIDs(net)
		l := len(nodeIDs)
		nodeconf := adapters.RandomNodeConfig()
		node, err := net.NewNodeWithConfig(nodeconf)
		if err != nil {
			return nil, fmt.Errorf("create node: %v", err)
		}
		err = net.Start(node.ID())
		if err != nil {
			return nil, fmt.Errorf("start node: %v", err)
		}

		log.Debug("created node", "id", node.ID())

		// connect nodes in a chain
		if l > 0 {
			var otherNodeID discover.NodeID
			for i := l - 1; i >= 0; i-- {
				n := net.GetNode(nodeIDs[i])
				if n.Up {
					otherNodeID = n.ID()
					break
				}
			}
			log.Debug("connect nodes", "one", node.ID(), "other", otherNodeID)
			if err := net.Connect(node.ID(), otherNodeID); err != nil {
				return nil, err
			}
		}
		ids = append(ids, node.ID())
	}
	return ids, nil
}

func removeNodes(count int, net *simulations.Network) error {
	for i := 0; i < count; i++ {
		nodeIDs := allNodeIDs(net)
		if len(nodeIDs) == 0 {
			break
		}
		node := net.GetNode(nodeIDs[rand.Intn(len(nodeIDs))])
		if err := node.Stop(); err != nil {
			return err
		}
		log.Debug("removed node", "id", node.ID())
	}
	return nil
}

func uploadFile(swarm *Swarm) (storage.Key, string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return nil, "", err
	}
	// File data is very short, but it is ensured that its
	// uniqueness is very certain.
	data := fmt.Sprintf("test content %s %x", time.Now().Round(0), b)
	k, wait, err := swarm.api.Put(data, "text/plain", false)
	if err != nil {
		return nil, "", err
	}
	if wait != nil {
		wait()
	}
	return k, data, nil
}

func retrieve(
	net *simulations.Network,
	files []file,
	swarms map[discover.NodeID]*Swarm,
	trigger chan discover.NodeID,
	checkStatusM *sync.Map,
	nodeStatusM *sync.Map,
	totalFoundCount *uint64,
) (missing uint64) {
	shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	var totalWg sync.WaitGroup
	errc := make(chan error)

	nodeIDs := allNodeIDs(net)

	totalCheckCount := len(nodeIDs) * len(files)

	for _, id := range nodeIDs {
		if _, ok := nodeStatusM.Load(id); ok {
			continue
		}
		start := time.Now()
		var checkCount uint64
		var foundCount uint64

		totalWg.Add(1)

		var wg sync.WaitGroup

		for _, f := range files {
			swarm := swarms[id]

			checkKey := check{
				key:    f.key.String(),
				nodeID: id,
			}
			if n, ok := checkStatusM.Load(checkKey); ok && n.(int) == 0 {
				continue
			}

			checkCount++
			wg.Add(1)
			go func(f file, id discover.NodeID) {
				defer wg.Done()

				log.Debug("api get: check file", "node", id.String(), "key", f.key.String(), "total files found", atomic.LoadUint64(totalFoundCount))

				r, _, _, err := swarm.api.Get(f.key, "/")
				if err != nil {
					errc <- fmt.Errorf("api get: node %s, key %s, kademlia %s: %v", id, f.key, swarm.bzz.Hive, err)
					return
				}
				d, err := ioutil.ReadAll(r)
				if err != nil {
					errc <- fmt.Errorf("api get: read response: node %s, key %s: kademlia %s: %v", id, f.key, swarm.bzz.Hive, err)
					return
				}
				data := string(d)
				if data != f.data {
					errc <- fmt.Errorf("file contend missmatch: node %s, key %s, expected %q, got %q", id, f.key, f.data, data)
					return
				}
				checkStatusM.Store(checkKey, 0)
				atomic.AddUint64(&foundCount, 1)
				log.Info("api get: file found", "node", id.String(), "key", f.key.String(), "content", data, "files found", atomic.LoadUint64(&foundCount))
			}(f, id)
		}

		go func(id discover.NodeID) {
			defer totalWg.Done()
			wg.Wait()

			atomic.AddUint64(totalFoundCount, foundCount)

			if foundCount == checkCount {
				log.Info("all files are found for node", "id", id.String(), "duration", time.Since(start))
				nodeStatusM.Store(id, 0)
				trigger <- id
				return
			}
			log.Debug("files missing for node", "id", id.String(), "check", checkCount, "found", foundCount)
		}(id)

	}

	go func() {
		totalWg.Wait()
		close(errc)
	}()

	var errCount int
	for err := range errc {
		if err != nil {
			errCount++
		}
		log.Error(err.Error())
	}

	log.Info("check stats", "total check count", totalCheckCount, "total files found", atomic.LoadUint64(totalFoundCount), "total errors", errCount)

	return uint64(totalCheckCount) - atomic.LoadUint64(totalFoundCount)
}

// Backported from stdlib https://golang.org/src/math/rand/rand.go?s=11175:11215#L333
//
// Replace with rand.Shuffle from go 1.10 when go 1.9 support is dropped.
//
// shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(rand.Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(rand.Int31n(int32(i + 1)))
		swap(i, j)
	}
}
