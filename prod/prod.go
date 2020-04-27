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

package prod

import (
	"context"
	"sync"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/pss/trojan"
)

// Sender defines code to be executed upon trigger of lost TODO: chunk????
type Sender func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error)

// Prod todo: define comment
type Prod struct {
	senders   map[*chunk.Address]Sender
	sendersMu sync.RWMutex
}

// NewProd inits the Prod struct
func NewProd() *Prod {
	return &Prod{
		senders: make(map[*chunk.Address]Sender),
	}
}

// Recover invokes underlying pss.Send as the first step of global pinning
// Does it return the chunk or the monitor?
func (p *Prod) Recover(ctx context.Context, chunkAddress chunk.Address) (chunk.Chunk, error) {
	var err error
	h := p.getSender(chunkAddress)

	// pss := NewPss(localStore, tags)

	//are the targets injected in the send or is it when it's called
	targets, err := p.getTargets(chunkAddress)
	if err != nil {
		return nil, err
	}
	payload := chunkAddress
	topic := trojan.NewTopic("RECOVERY")

	if h != nil {
		if _, err = h(ctx, targets, topic, payload); err != nil {
			return nil, err
		}
	}

	// should it wait for the chunk and timeout if not
	// using the monitor of pss until this is received
	return nil, nil
}

func (p *Prod) getTargets(chunkAddress chunk.Address) ([][]byte, error) {
	//this should get the feed and return correct target of pinners
	return [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}, nil
}

// Register should be called internally
// for every chunk that is globally pinned thru the feed

// register
func (p *Prod) register(chunkAddress chunk.Address, sender Sender) {
	p.sendersMu.Lock()
	defer p.sendersMu.Unlock()
	p.senders[&chunkAddress] = sender
}

// getSender returns the Sender func registered in prod for the given chunk address
func (p *Prod) getSender(chunkAddress chunk.Address) Sender {
	p.sendersMu.RLock()
	defer p.sendersMu.RUnlock()
	return p.senders[&chunkAddress]
}
