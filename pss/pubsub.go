// Copyright 2019 The go-ethereum Authors
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

package pss

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/pss/message"
)

// PubSub implements the pushsync.PubSub interface using pss
type PubSub struct {
	pss *Pss
}

// NewPubSub creates a new PubSub
func NewPubSub(p *Pss) *PubSub {
	return &PubSub{
		pss: p,
	}
}

// BaseAddr returns Kademlia base address
func (p *PubSub) BaseAddr() []byte {
	return p.pss.BaseAddr()
}

func isPssPeer(bp *network.BzzPeer) bool {
	return bp.HasCap(protocolName)
}

// IsClosestTo returns true is self is the closest known node to addr
// as uniquely defined by the MSB XOR distance
// among pss capable peers
func (p *PubSub) IsClosestTo(addr []byte) bool {
	return p.pss.IsClosestTo(addr, isPssPeer)
}

// Register registers a handler
func (p *PubSub) Register(topic string, prox bool, handler func(msg []byte, p *p2p.Peer) error) func() {
	f := func(msg []byte, peer *p2p.Peer, _ bool, _ string) error {
		return handler(msg, peer)
	}
	h := NewHandler(f).WithRaw()
	if prox {
		h = h.WithProxBin()
	}
	pt := message.NewTopic([]byte(topic))
	return p.pss.Register(&pt, h)
}

// Send sends a message using pss SendRaw
func (p *PubSub) Send(to []byte, topic string, msg []byte) error {
	pt := message.NewTopic([]byte(topic))
	return p.pss.SendRaw(PssAddress(to), pt, msg)
}
