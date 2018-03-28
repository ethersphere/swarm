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
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	colorable "github.com/mattn/go-colorable"
	"golang.org/x/sync/errgroup"
)

var (
	loglevel = flag.Int("loglevel", 4, "verbosity of logs")
)

func init() {
	rand.Seed(time.Now().UnixNano())

	//log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

func TestSync(t *testing.T) {
	testSync(t,
		// testSyncStep{
		// 	nodeCount: 1,
		// },
		// testSyncStep{
		// 	nodeCount: 3,
		// },
		testSyncStep{
			nodeCount: 10,
		},
		// testSyncStep{
		// 	nodeCount: 8,
		// },
	)
}

type testSyncStep struct {
	nodeCount int
	timeout   time.Duration
}

func testSync(t *testing.T, steps ...testSyncStep) {
	dir, err := ioutil.TempDir("", "swarm-network-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	type file struct {
		key    storage.Key
		data   string
		nodeID discover.NodeID
	}

	swarms := make(map[discover.NodeID]*Swarm)
	files := make([]file, 0)

	services := map[string]adapters.ServiceFunc{
		"swarm": func(ctx *adapters.ServiceContext) (node.Service, error) {
			config := api.NewConfig()

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

	type fileOnNode struct {
		key    string
		nodeID discover.NodeID
	}

	//fileStatus := make(map[fileOnNode]int)
	//foundCount := 0

	var fileStatusM sync.Map

	var foundCount uint64

	check := func(ctx context.Context, id discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		//ok := true

		rand.Shuffle(len(files), func(i, j int) {
			files[i], files[j] = files[j], files[i]
		})

		var g errgroup.Group

		//for _, id := range allNodeIDs(net) {
		for _, f := range files {
			swarm := swarms[id]

			ffKey := fileOnNode{
				key:    f.key.String(),
				nodeID: id,
			}
			if n, ok := fileStatusM.Load(ffKey); ok && n.(int) == 0 {
				continue
			}
			// if n := fileStatus[ffKey]; n == -1 {
			// 	continue
			// }
			f := f
			g.Go(func() error {
				r, _, _, err := swarm.api.Get(f.key, "/")
				if err != nil {
					log.Warn("api get", "node", id.String(), "key", f.key.String(), "err", err)
					return fmt.Errorf("api get: node %s, key %s: %v", id, f.key, err)
				}
				d, err := ioutil.ReadAll(r)
				if err != nil {
					log.Warn("api get: read response", "node", id.String(), "key", f.key.String(), "err", err)
					return fmt.Errorf("api get: read response: node %s, key %s: %v", id, f.key, err)
				}
				data := string(d)
				if data != f.data {
					return fmt.Errorf("file contend missmatch: node %s, key %s, expected %q, got %q", id, f.key, f.data, data)
				}
				fileStatusM.Store(ffKey, 0)
				atomic.AddUint64(&foundCount, 1)
				log.Info("api get: file found", "node", id.String(), "key", f.key.String(), "content", data, "total files found", atomic.LoadUint64(&foundCount))
				return nil
			})
			// fileStatus[ffKey]++
			// log.Warn("api get", "node", id.String(), "key", f.key.String(), "err", err, "tries", fileStatus[ffKey])

			// foundCount++
			// log.Info("api get: file found", "node", id.String(), "key", f.key.String(), "content", data, "total files found", foundCount)
			// fileStatus[ffKey] = -1

			// ffKey := fileOnNode{
			// 	key:    f.key.String(),
			// 	nodeID: id,
			// }
			// if n := fileStatus[ffKey]; n == -1 {
			// 	continue
			// }
			// r, _, _, err := swarm.api.Get(f.key, "/")
			// if err != nil {
			// 	fileStatus[ffKey]++
			// 	log.Warn("api get", "node", id.String(), "key", f.key.String(), "err", err, "tries", fileStatus[ffKey])
			// 	ok = false
			// 	continue
			// 	//return false, nil
			// }
			// d, err := ioutil.ReadAll(r)
			// if err != nil {
			// 	fileStatus[ffKey]++
			// 	log.Warn("api get: read response", "node", id.String(), "key", f.key.String(), "err", err, "tries", fileStatus[ffKey])
			// 	ok = false
			// 	continue
			// 	//return false, nil
			// }
			// data := string(d)
			// if data != f.data {
			// 	return false, fmt.Errorf("file contend missmatch: node %s, key %s, expected %q, got %q", id, f.key, f.data, data)
			// }
			// foundCount++
			// log.Info("api get: file found", "node", id.String(), "key", f.key.String(), "content", data, "total files found", foundCount)
			// fileStatus[ffKey] = -1
		}
		//}

		if err := g.Wait(); err != nil {
			log.Error("check", "total files found", atomic.LoadUint64(&foundCount), "err", err)
			return false, nil
		}

		return true, nil

		//return ok, nil
	}

	// timingTicker := time.NewTicker(time.Second * 1)
	// defer timingTicker.Stop()
	// go func() {
	// 	for range timingTicker.C {
	// 		for _, id := range allNodeIDs(net) {
	// 			trigger <- id
	// 		}
	// 	}
	// }()

	sim := simulations.NewSimulation(net)

	for i, step := range steps {
		log.Debug("test sync step", "n", i+1, "nodes", step.nodeCount)
		nodeIDs := allNodeIDs(net)

		change := step.nodeCount - len(nodeIDs)

		if change > 0 {
			ids, err := addNodes(change, net, trigger)
			if err != nil {
				t.Fatal(err)
			}
			nodeIDs = append(nodeIDs, ids...)

			// TODO: shuffle the nodeIDs
			for _, id := range nodeIDs {
				key, data, err := uploadFile(swarms[id])
				if err != nil {
					t.Fatal(err)
				}
				log.Debug("file uploaded", "node", id, "key", key.String())
				files = append(files, file{
					key:    key,
					data:   data,
					nodeID: id,
				})
			}
			//time.Sleep(20 * time.Second)
		} else if change < 0 {
			err := removeNodes(-change, net)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			t.Logf("step %v: no change in nodes", i)
			continue
		}
		//startNodes := removeNodeIDs(allNodeIDs(net), ids...)
		// if len(startNodes) > 0 {
		// 	otherNode := startNodes[rand.Intn(len(startNodes))]
		// 	ids = append(ids, otherNode)
		// }

		//time.Sleep(2 * time.Second)

		timeout := step.timeout
		if timeout == 0 {
			timeout = 10 * time.Minute
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		//only needed for healthy call when debugging
		nIDs := allNodeIDs(net)
		addrs := make([][]byte, len(nIDs))
		for i, id := range nIDs {
			addrs[i] = swarms[id].bzz.BaseAddr()
		}
		ppmap := network.NewPeerPot(2, nIDs, addrs)

		result := sim.Run(ctx, &simulations.Step{
			Action: func(ctx context.Context) error {
				ticker := time.NewTicker(200 * time.Millisecond)
				defer ticker.Stop()

				for range ticker.C {
					healthy := true
					for _, id := range nIDs {
						swarm := swarms[id]
						//PeerPot for this node
						pp := ppmap[id]
						//call Healthy RPC
						h := swarm.bzz.Healthy(pp)
						//print info
						log.Debug(swarm.bzz.String())
						log.Debug("kademlia", "health", h.GotNN && h.KnowNN && h.Full)
						if !h.GotNN || !h.Full {
							healthy = false
							break
						}
					}
					if healthy {
						break
					}
				}
				return nil
			},
			Trigger: trigger,
			Expect: &simulations.Expectation{
				Nodes: allNodeIDs(net),
				Check: check,
			},
		})
		if result.Error != nil {
			t.Fatal(result.Error)
		}
		log.Debug("done: test sync step", "n", i+1, "nodes", step.nodeCount)
	}

	// var found, missing int

	// for k, n := range fileStatus {
	// 	if n == -1 {
	// 		log.Debug("file found", "key", k.key, "node", k.nodeID)
	// 		found++
	// 	} else {
	// 		log.Debug("file missing", "key", k.key, "node", k.nodeID, "tries", n)
	// 		missing++
	// 	}
	// }

	// log.Debug("files sum", "found", found, "missing", missing)
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
		//nodes := net.GetNodes()
		nodeconf := adapters.RandomNodeConfig()
		node, err := net.NewNodeWithConfig(nodeconf)
		if err != nil {
			return nil, fmt.Errorf("create node: %v", err)
		}
		err = net.Start(node.ID())
		if err != nil {
			return nil, fmt.Errorf("start node: %v", err)
		}
		if err := triggerChecks(trigger, net, node.ID()); err != nil {
			return nil, fmt.Errorf("error triggering checks for node %s: %s", node.ID().TerminalString(), err)
		}

		// nodes := net.GetNodes()
		// // remove the current node from list
		// for i := range nodes {
		// 	// TODO: verify that nodes that are not up are removed!
		// 	if nodes[i] == node || !nodes[i].Up {
		// 		nodes = append(nodes[:i], nodes[i+1:]...)
		// 	}
		// }

		log.Debug("created node", "id", node.ID())
		nodeIDs := allNodeIDs(net)
		l := len(nodeIDs)
		if l > 1 {
			log.Debug("connect nodes", "one", nodeIDs[l-1], "other", nodeIDs[l-2])
			if err := net.Connect(nodeIDs[l-1], nodeIDs[l-2]); err != nil {
				return nil, err
			}
		}

		//conns := 1
		// Chain connection
		//wg := sync.WaitGroup{}
		// for i := 0; i < l-1; i++ {
		// 	net.Connect(nodeIDs[i], nodeIDs[i+1])

		// collect the overlay addresses, to
		// for j := 0; conns < 1; j++ {
		// 	var k int
		// 	if j == 0 {
		// 		k = i - 1
		// 	} else {
		// 		k = rand.Intn(len(nodeIDs))
		// 	}
		// 	if i > 0 {
		// 		wg.Add(1)
		// 		go func(i, k int) {
		// 			defer wg.Done()
		// 			net.Connect(nodeIDs[i], nodeIDs[k])
		// 		}(i, k)
		// 	}
		// }
		//}
		//wg.Wait()

		// Full connectivity
		// for _, n := range nodes {
		// 	if n.Up {
		// 		if err := net.Connect(node.ID(), n.ID()); err != nil {
		// 			return nil, err
		// 		}
		// 	}
		// }

		// Random connection to a single node
		// if len(nodes) >= 1 {
		// 	other := nodes[rand.Intn(len(nodes))]
		// 	if err := net.Connect(node.ID(), other.ID()); err != nil {
		// 		return nil, err
		// 	}
		// }
		ids = append(ids, node.ID())
	}
	return ids, nil
}

func removeNodes(count int, net *simulations.Network) error {
	for i := 0; i < count; i++ {
		nodes := net.GetNodes()
		if len(nodes) == 0 {
			break
		}
		node := nodes[rand.Intn(len(nodes))]

		if err := node.Stop(); err != nil {
			return err
		}
		for _, n := range nodes {
			if n != node {
				net.Disconnect(n.ID(), node.ID())
			}
		}
		log.Debug("removed node", "id", node.ID())
	}
	return nil
}

func uploadFile(swarm *Swarm) (storage.Key, string, error) {
	data := "test file content " + time.Now().String()
	k, wait, err := swarm.api.Put(data, "text/plain")
	if err != nil {
		return nil, "", err
	}
	if wait != nil {
		wait()
	}
	return k, data, nil
}

func triggerChecks(trigger chan discover.NodeID, net *simulations.Network, id discover.NodeID) error {
	node := net.GetNode(id)
	if node == nil {
		return fmt.Errorf("unknown node: %s", id)
	}
	client, err := node.Client()
	if err != nil {
		return err
	}
	events := make(chan *p2p.PeerEvent)
	sub, err := client.Subscribe(context.Background(), "admin", events, "peerEvents")
	if err != nil {
		return fmt.Errorf("error getting peer events for node %v: %s", id, err)
	}

	go func() {
		defer sub.Unsubscribe()

		tick := time.NewTicker(1000 * time.Millisecond)
		defer tick.Stop()

		for {
			select {
			case e := <-events:
				log.Trace("check: event received", "node", id.String(), "event", e)
				trigger <- id
			case <-tick.C:
				log.Trace("check: tick", "node", id.String())
				trigger <- id
			case err := <-sub.Err():
				if err != nil {
					log.Error(fmt.Sprintf("error getting peer events for node %v", id), "err", err)
				}
				return
			}
		}
	}()
	return nil
}

// func removeNodeIDs(ids []discover.NodeID, n ...discover.NodeID) (r []discover.NodeID) {
// 	for _, i := range ids {
// 		var found bool
// 		for _, j := range n {
// 			if i != j {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			r = append(r, i)
// 		}
// 	}
// 	return
// }
