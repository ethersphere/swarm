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

package orbit

import (
	"context"
	"errors"
	"fmt"

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

func (p *Peer) Left() {
	close(p.quit)
}

// HandleMsg is the message handler that delegates incoming messages
func (p *Peer) HandleMsg(ctx context.Context, msg interface{}) error {
	switch msg := msg.(type) {

	//case *someType:
	//return p.handleSomeMsgt(msg)

	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}
}
