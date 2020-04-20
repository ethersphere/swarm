// Copyright 2020 The Swarm Authors
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

package pss

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/chunk"
	trojan "github.com/ethersphere/swarm/pss/trojan"
)

// Pss is the top-level struct, which takes care of message sending
type Pss struct {
	localStore chunk.Store
	handlers   map[trojan.Topic]Handler
	handlersMu sync.RWMutex
}

// NewPss inits the Pss struct with the localstore
func NewPss(localStore chunk.Store) *Pss {
	return &Pss{
		localStore: localStore,
		handlers:   make(map[trojan.Topic]Handler),
	}
}

// Handler defines code to be executed upon reception of a trojan message
type Handler func(trojan.Message)

// Send constructs a padded message with topic and payload,
// wraps it in a trojan chunk such that one of the targets is a prefix of the chunk address
// stores this in localstore for push-sync to pick up and deliver
func (p *Pss) Send(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (chunk.Chunk, error) {
	metrics.GetOrRegisterCounter("trojanchunk/send", nil).Inc(1)
	//construct Trojan Chunk
	m, err := trojan.NewMessage(topic, payload)
	if err != nil {
		return nil, err
	}
	var tc chunk.Chunk
	tc, err = m.Wrap(targets)
	if err != nil {
		return nil, err
	}

	// SAVE trojanChunk to localstore, if it exists do nothing as it's already peristed
	// TODO: for second phase, use tags --> listen for response of recipient, recipient offline
	if _, err = p.localStore.Put(ctx, chunk.ModePutUpload, tc); err != nil {
		return nil, err
	}

	return tc, nil
}

// Register allows the definition of a Handler func for a specific topic on the pss struct
func (p *Pss) Register(topic trojan.Topic, hndlr Handler) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	p.handlers[topic] = hndlr
}

// Deliver allows unwrapping a chunk as a trojan message and calling its handler func based on its topic
func (p *Pss) Deliver(c chunk.Chunk) {
	m, _ := trojan.Unwrap(c)
	h := p.getHandler(m.Topic)
	if h != nil {
		h(*m)
	}
}

// getHandler returns the Handler func registered in pss for the given topic
func (p *Pss) getHandler(topic trojan.Topic) Handler {
	p.handlersMu.RLock()
	defer p.handlersMu.RUnlock()
	return p.handlers[topic]
}
