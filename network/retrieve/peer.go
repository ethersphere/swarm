// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package retrieve

import (
	"errors"

	"github.com/ethersphere/swarm/network"
)

// ErrMaxPeerServers will be returned if peer server limit is reached.
// It will be sent in the SubscribeErrorMsg.
var ErrMaxPeerServers = errors.New("max peer servers")

// Peer is the Peer extension for the streaming protocol
type Peer struct {
	*network.BzzPeer
	quit chan struct{}
}

// NewPeer is the constructor for Peer
func NewPeer(peer *network.BzzPeer) *Peer {
	p := &Peer{
		BzzPeer: peer,
		quit:    make(chan struct{}),
	}
	return p
}
