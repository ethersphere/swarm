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
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

func TestService(t *testing.T) {
	sim := NewSimulation(map[string]ServiceFunc{
		"noop": func(_ *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
			return newNoopService(), nil, nil
		},
	}, nil)
	defer sim.Close()

	id, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	_, ok := sim.Service("noop", id).(*noopService)
	if !ok {
		t.Fatalf("service is not of %T type", &noopService{})
	}

	_, ok = sim.RandomService("noop").(*noopService)
	if !ok {
		t.Fatalf("service is not of %T type", &noopService{})
	}

	_, ok = sim.Services("noop")[0].(*noopService)
	if !ok {
		t.Fatalf("service is not of %T type", &noopService{})
	}
}
