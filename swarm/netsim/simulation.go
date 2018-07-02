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
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

// Common errors that are returned by functions in this package.
var (
	ErrNodeNotFound = errors.New("node not found")
	ErrNoPivotNode  = errors.New("no pivot node set")
)

// Simulation provides methods on network, nodes and services
// to manage them.
type Simulation struct {
	Net *simulations.Network

	serviceNames []string
	cleanupFuncs []func()
	buckets      map[discover.NodeID]*sync.Map
	pivotNodeID  *discover.NodeID
	shutdownWG   sync.WaitGroup
	mu           sync.RWMutex
}

// ServiceFunc is used in NewSimulation to declare new service constructor.
// The first argument provides ServiceContext from the adapters package
// giving for example the access to NodeID. Second argument is the sync.Map
// where all "global" state related to the service should be kept.
// All cleanups needed for constructed service and any other constructed
// objects should ne provided in a single returned cleanup function.
type ServiceFunc func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error)

// NewSimulation creates a new Simulation instance with new
// simulations.Network initialized with provided services.
func NewSimulation(services map[string]ServiceFunc) (s *Simulation) {
	s = &Simulation{
		buckets: make(map[discover.NodeID]*sync.Map),
	}

	adapterServices := make(map[string]adapters.ServiceFunc, len(services))
	for name, serviceFunc := range services {
		s.serviceNames = append(s.serviceNames, name)
		adapterServices[name] = func(ctx *adapters.ServiceContext) (node.Service, error) {
			b := new(sync.Map)
			service, cleanup, err := serviceFunc(ctx, b)
			if err != nil {
				return nil, err
			}
			s.mu.Lock()
			defer s.mu.Unlock()
			if cleanup != nil {
				s.cleanupFuncs = append(s.cleanupFuncs, cleanup)
			}
			s.buckets[ctx.Config.ID] = b
			return service, nil
		}
	}

	s.Net = simulations.NewNetwork(
		adapters.NewSimAdapter(adapterServices),
		&simulations.NetworkConfig{ID: "0"},
	)
	return s
}

// RunFunc is the function that will be called
// on Simulation.Run method call.
type RunFunc func(context.Context, *Simulation) error

// Result is the returned value of Simulation.Run method.
type Result struct {
	Duration time.Duration
	Error    error
}

// Run calls the RunFunc function while taking care of
// cancelation provided through the Context.
func (s *Simulation) Run(ctx context.Context, f RunFunc) (r Result) {
	start := time.Now()
	errc := make(chan error)
	quit := make(chan struct{})
	defer close(quit)
	go func() {
		select {
		case errc <- f(ctx, s):
		case <-quit:
		}
	}()
	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errc:
	}
	return Result{
		Duration: time.Since(start),
		Error:    err,
	}
}

// Maximal number of parallel calls to cleanup functions on
// Simulation.Close.
var maxParallelCleanups = 10

// Close calls all cleanup functions that are returned by
// ServiceFunc, waits for all of them to finish and other
// functions that explicitly block shutdownWG
// (like Simulation.PeerEvents) and shuts down the network
// at the end. It is used to clean all resources from the
// simulation.
func (s *Simulation) Close() {
	sem := make(chan struct{}, maxParallelCleanups)
	s.mu.RLock()
	cleanupFuncs := make([]func(), len(s.cleanupFuncs))
	for i, f := range s.cleanupFuncs {
		if f != nil {
			cleanupFuncs[i] = f
		}
	}
	s.mu.RUnlock()
	for _, cleanup := range cleanupFuncs {
		s.shutdownWG.Add(1)
		sem <- struct{}{}
		go func(cleanup func()) {
			defer s.shutdownWG.Done()
			defer func() { <-sem }()

			cleanup()
		}(cleanup)
	}
	s.shutdownWG.Wait()
	s.Net.Shutdown()
}
