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

	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

var (
	ErrNodeNotFound = errors.New("node not found")
	ErrNoPivotNode  = errors.New("no pivot node set")
)

type Simulation struct {
	Net       *simulations.Network
	closeFunc func() error

	pivotNodeID *discover.NodeID
	mu          sync.Mutex
}

type Options struct {
	ServiceFunc adapters.ServiceFunc
}

func NewSimulation(o Options) (s *Simulation, err error) {
	a := adapters.NewSimAdapter(map[string]adapters.ServiceFunc{
		"service": o.ServiceFunc,
	})
	net := simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "service",
	})
	closeFunc := func() error {
		net.Shutdown()
		return nil
	}
	s = &Simulation{
		Net:       net,
		closeFunc: closeFunc,
	}
	return s, nil
}

func (s *Simulation) Close() error {
	return s.closeFunc()
}
