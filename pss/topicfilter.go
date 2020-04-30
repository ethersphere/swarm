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
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/pss/message"
)

// TopicFilter implements the pushsync.TopicFilter interface using pss
type TopicFilter struct {
	pss        *Pss
	messageTTL time.Duration // expire duration of a topic filter message. Depends on the use case.
}

// NewTopicFilter creates a new TopicFilter
func NewTopicFilter(p *Pss, messageTTL time.Duration) *TopicFilter {
	return &TopicFilter{
		pss:        p,
		messageTTL: messageTTL,
	}
}

// BaseAddr returns Kademlia base address
func (p *TopicFilter) BaseAddr() []byte {
	return p.pss.BaseAddr()
}

func isPssPeer(bp *network.BzzPeer) bool {
	return bp.HasCap(protocolName)
}

// IsClosestTo returns true is self is the closest known node to addr
// as uniquely defined by the MSB XOR distance
// among pss capable peers
func (p *TopicFilter) IsClosestTo(addr []byte) bool {
	return p.pss.IsClosestTo(addr, isPssPeer)
}

// Register registers a handler
func (p *TopicFilter) Register(topic string, prox bool, handler func(msg []byte, p *p2p.Peer) error) func() {
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
func (p *TopicFilter) Send(to []byte, topic string, msg []byte) error {
	defer metrics.GetOrRegisterResettingTimer("pss/topicfilter/send", nil).UpdateSince(time.Now())
	pt := message.NewTopic([]byte(topic))
	return p.pss.SendRaw(PssAddress(to), pt, msg, p.messageTTL)
}
