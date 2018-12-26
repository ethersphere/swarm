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

package simulations

import (
	"testing"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

func newTestNetwork(t *testing.T, nodeCount int) (*Network, []enode.ID) {
	adapter := adapters.NewSimAdapter(adapters.Services{
		"noopwoop": func(ctx *adapters.ServiceContext) (node.Service, error) {
			return NewNoopService(nil), nil
		},
	})

	// create network
	network := NewNetwork(adapter, &NetworkConfig{
		DefaultService: "noopwoop",
	})

	// create and start nodes
	ids := make([]enode.ID, nodeCount)
	for i := range ids {
		conf := adapters.RandomNodeConfig()
		node, err := network.NewNodeWithConfig(conf)
		if err != nil {
			t.Fatalf("error creating node: %s", err)
		}
		if err := network.Start(node.ID()); err != nil {
			t.Fatalf("error starting node: %s", err)
		}
		ids[i] = node.ID()
	}

	if len(network.Conns) > 0 {
		t.Fatal("no connections should exist after just adding nodes")
	}

	return network, ids
}

func TestConnectNodesChain(t *testing.T) {
	net, ids := newTestNetwork(t, 10)
	defer net.Shutdown()

	err := net.ConnectNodesChain(ids)
	if err != nil {
		t.Fatal(err)
	}

	verifyChain(t, net, ids)
}

func verifyChain(t *testing.T, net *Network, ids []enode.ID) {
	t.Helper()
	n := len(ids)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			c := net.GetConn(ids[i], ids[j])
			if i == j-1 {
				if c == nil {
					t.Errorf("nodes %v and %v are not connected, but they should be", i, j)
				}
			} else {
				if c != nil {
					t.Errorf("nodes %v and %v are connected, but they should not be", i, j)
				}
			}
		}
	}
}

func TestConnectNodesRing(t *testing.T) {
	net, ids := newTestNetwork(t, 10)
	defer net.Shutdown()

	err := net.ConnectNodesRing(ids)
	if err != nil {
		t.Fatal(err)
	}

	verifyRing(t, net, ids)
}

func verifyRing(t *testing.T, net *Network, ids []enode.ID) {
	t.Helper()
	n := len(ids)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			c := net.GetConn(ids[i], ids[j])
			if i == j-1 || (i == 0 && j == n-1) {
				if c == nil {
					t.Errorf("nodes %v and %v are not connected, but they should be", i, j)
				}
			} else {
				if c != nil {
					t.Errorf("nodes %v and %v are connected, but they should not be", i, j)
				}
			}
		}
	}
}
