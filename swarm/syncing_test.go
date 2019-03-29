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

package swarm

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/network/simulation"
)

// this test is just for measuring syncing time, it will run indefinitely
func TestSyncing(t *testing.T) {
	const (
		nodeCount      = 20
		fileSize       = 8 * 1024 * 1024
		maxStableCount = 10
		checkDelay     = 100 * time.Millisecond
		// a syncing is considered done when all nodes have the same number of chunks
		// for maxStableCount with check delays of checkDelay
	)
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"swarm": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			config := api.NewConfig()

			dir, err := ioutil.TempDir("", "swarm-syncing-")
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
			log.Info("new swarm", "bzzKey", config.BzzKey, "baseAddr", fmt.Sprintf("%x", swarm.bzz.BaseAddr()))
			return swarm, cleanup, nil
		},
	})
	defer sim.Close()

	ids, err := sim.AddNodesAndConnectChain(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(",upload,sync,min,max")
	uploadCount := 0
	for {
		counts := make(map[enode.ID]uint64)
		for _, id := range ids {
			uploadStart := time.Now()
			key, _, err := uploadFile(sim.Service("swarm", id).(*Swarm), fileSize)
			if err != nil {
				t.Fatal(err)
			}
			uploadDuration := time.Since(uploadStart).Seconds()
			uploadCount++
			log.Trace("file uploaded", "node", id, "key", key.String())
			start := time.Now()
			var min, max uint64
			var syncDuration float64
			for stableCount := 0; stableCount < maxStableCount; {
				stable := true
				for _, id := range ids {
					s := sim.Service("swarm", id).(*Swarm)
					count := s.Size()

					if count != counts[id] {
						stable = false
					}
					if count == 0 {
						stable = false
					}
					counts[id] = count
					if count > max {
						max = count
					}
					if count < min || min == 0 {
						min = count
					}

					c := s.SyncClientDeliveryCount()
					if c < 0 {
						panic("negative sync client delivery count")
					}
					if c != 0 {
						stable = false
					}
				}
				if stable {
					if stableCount == 0 {
						syncDuration = time.Since(start).Seconds()
					}
					stableCount++
				} else {
					stableCount = 0
				}
				time.Sleep(checkDelay)
			}
			fmt.Printf("%v,%.2f,%.2f,%v,%v\n", uploadCount, uploadDuration, syncDuration, min, max)
		}
	}
}
