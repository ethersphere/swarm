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
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

var (
	ErrNodeNotFound = errors.New("node not found")
	ErrNoPivotNode  = errors.New("no pivot node set")
)

type Simulation struct {
	Net *simulations.Network

	pivotNodeID *discover.NodeID
	buckets     map[discover.NodeID]*sync.Map
	mu          sync.RWMutex
}

var BucketKeyCleanup BucketKey = "cleanup"

type Options struct {
	ServiceFunc func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error)
}

func NewSimulation(o Options) (s *Simulation, err error) {
	s = &Simulation{
		buckets: make(map[discover.NodeID]*sync.Map),
	}
	a := adapters.NewSimAdapter(map[string]adapters.ServiceFunc{
		"service": func(ctx *adapters.ServiceContext) (node.Service, error) {
			b := new(sync.Map)
			n, cleanup, err := o.ServiceFunc(ctx, b)
			if err != nil {
				return nil, err
			}
			if cleanup != nil {
				b.Store(BucketKeyCleanup, cleanup)
			}
			s.mu.Lock()
			defer s.mu.Unlock()
			s.buckets[ctx.Config.ID] = b
			return n, nil
		},
	})
	s.Net = simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "service",
	})
	return s, nil
}

var maxParallelCleanups = 10

func (s *Simulation) Close() error {
	sem := make(chan struct{}, maxParallelCleanups)
	var wg sync.WaitGroup
	for _, v := range s.ServicesItems(BucketKeyCleanup) {
		cleanup, ok := v.(func())
		if !ok {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			cleanup()
		}()
	}
	wg.Wait()
	s.Net.Shutdown()
	return nil
}
