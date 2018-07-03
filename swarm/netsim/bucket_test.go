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

func TestServiceBucket(t *testing.T) {
	testKey := BucketKey("Key")
	testValue := "Value"

	sim := NewSimulation(map[string]ServiceFunc{
		"noop": func(_ *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
			b.Store(testKey, testValue)
			return newNoopService(), nil, nil
		},
	}, nil)
	defer sim.Close()

	id, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	v, ok := sim.NodeItem(id, testKey)
	if !ok {
		t.Fatal("bucket item not found")
	}
	s, ok := v.(string)
	if !ok {
		t.Fatal("bucket item value is not string")
	}
	if s != testValue {
		t.Fatalf("expected %q, got %q", testValue, s)
	}
}
