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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

var (
	ErrNodeNotFound    = errors.New("node not found")
	ErrNoPivotNode     = errors.New("no pivot node set")
	DefaultHttpSimPort = "8888"
)

type Simulation struct {
	Net *simulations.Network

	pivotNodeID *discover.NodeID
	buckets     map[discover.NodeID]*sync.Map
	shutdownWG  sync.WaitGroup
	mu          sync.RWMutex
	httpSrv     *http.Server
}

var (
	BucketKeyCleanup BucketKey = "cleanup"
)

type Options struct {
	ServiceFunc func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error)
	WithHTTP    bool
	HttpSimPort string
}

func NewSimulation(o Options) (s *Simulation) {
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
	if o.WithHTTP {
		if o.HttpSimPort == "" {
			o.HttpSimPort = DefaultHttpSimPort
		}
		log.Info(fmt.Sprintf("starting simulation server on 0.0.0.0:%d...", o.HttpSimPort))
		s.httpSrv = &http.Server{
			Addr:    fmt.Sprintf(":%s", o.HttpSimPort),
			Handler: simulations.NewServer(s.Net),
		}
		//start the HTTP server
		go s.httpSrv.ListenAndServe()
		log.Info("Waiting for frontend to be ready...(send POST /runsim to HTTP server)")
		<-s.Net.RunC
		log.Info("Received signal from frontend - starting simulation run.")
	}
	return s
}

type RunFunc func(context.Context, *Simulation) error

type Result struct {
	Duration time.Duration
	Error    error
}

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

var maxParallelCleanups = 10

func (s *Simulation) Close() {
	sem := make(chan struct{}, maxParallelCleanups)
	for _, v := range s.ServicesItems(BucketKeyCleanup) {
		cleanup, ok := v.(func())
		if !ok {
			continue
		}
		s.shutdownWG.Add(1)
		sem <- struct{}{}
		go func() {
			defer s.shutdownWG.Done()
			defer func() { <-sem }()

			cleanup()
		}()
	}
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := s.httpSrv.Shutdown(ctx)
		if err != nil {
			log.Error("Error shutting down HTTP server!", "err", err)
		}
	}
	s.shutdownWG.Wait()
	s.Net.Shutdown()
	return
}
