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
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

var (
	ErrNodeNotFound = errors.New("node not found")
)

// ConnectNodesChain connects all nodes in a chain topology.
// If ids argument is nil, all nodes that are up will be connected.
func (net *Network) ConnectNodesChain(ids []enode.ID) (err error) {
	if ids == nil {
		ids = net.getUpNodeIDs()
	}
	l := len(ids)
	for i := 0; i < l-1; i++ {
		if err := net.connect(ids[i], ids[i+1]); err != nil {
			return err
		}
	}
	return nil
}

// ConnectNodesRing connects all nodes in a ring topology.
// If ids argument is nil, all nodes that are up will be connected.
func (net *Network) ConnectNodesRing(ids []enode.ID) (err error) {
	if ids == nil {
		ids = net.getUpNodeIDs()
	}
	l := len(ids)
	if l < 2 {
		return nil
	}
	if err := net.ConnectNodesChain(ids); err != nil {
		return err
	}
	return net.connect(ids[l-1], ids[0])
}

// connect connects two nodes but ignores already connected error.
func (net *Network) connect(oneID, otherID enode.ID) error {
	return ignoreAlreadyConnectedErr(net.Connect(oneID, otherID))
}

func ignoreAlreadyConnectedErr(err error) error {
	if err == nil || strings.Contains(err.Error(), "already connected") {
		return nil
	}
	return err
}
