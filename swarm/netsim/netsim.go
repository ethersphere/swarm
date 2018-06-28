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
	buckets     map[discover.NodeID]*Bucket
	mu          sync.RWMutex
}

type Options struct {
	ServiceFunc func(ctx *adapters.ServiceContext, bucket *Bucket) (node.Service, error)
}

func NewSimulation(o Options) (s *Simulation, err error) {
	s = &Simulation{
		buckets: make(map[discover.NodeID]*Bucket),
	}
	a := adapters.NewSimAdapter(map[string]adapters.ServiceFunc{
		"service": func(ctx *adapters.ServiceContext) (node.Service, error) {
			b := newBucket()
			n, err := o.ServiceFunc(ctx, b)
			if err != nil {
				return nil, err
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

func (s *Simulation) Close() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var wg sync.WaitGroup
	for _, c := range s.buckets {
		c := c
		wg.Add(1)
		go func() {
			defer wg.Done()

			c.mu.RLock()
			defer c.mu.RUnlock()

			for _, s := range c.values {
				if closer, ok := s.(interface{ Close() }); ok {
					closer.Close()
				}
				if closer, ok := s.(interface{ Close() error }); ok {
					closer.Close()
				}
			}
		}()
	}
	wg.Wait()
	s.Net.Shutdown()
	return nil
}
