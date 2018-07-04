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
package netsim

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/network"
)

//To test that uploading a snapshot works
func TestUploadSnapshot(t *testing.T) {
	log.Debug("Creating simulation")
	s := New(map[string]ServiceFunc{
		"bzz": func(ctx *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
			addr := network.NewAddrFromNodeID(ctx.Config.ID)
			hp := network.NewHiveParams()
			hp.Discovery = false
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			kad := network.NewKademlia(addr.Over(), network.NewKadParams())
			return network.NewBzz(config, kad, nil, nil, nil), nil, nil
		},
	}, nil)
	defer s.Close()

	nodeCount := 16
	log.Debug("Uploading snapshot")
	err := s.UploadSnapshot(fmt.Sprintf("../network/stream/testing/snapshot_%d.json", nodeCount))
	if err != nil {
		t.Fatalf("Error uploading snapshot to simulation network: %v", err)
	}

	ctx := context.Background()
	log.Debug("Starting simulation...")
	s.Run(ctx, func(ctx context.Context, sim *Simulation) error {
		log.Debug("Checking")
		nodes := sim.UpNodeIDs()
		if len(nodes) != nodeCount {
			t.Fatal("Simulation network node number doesn't match snapshot node number")
		}
		return nil
	})
	log.Debug("Done.")
}
