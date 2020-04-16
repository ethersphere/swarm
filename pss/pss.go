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

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/chunk"
	trojan "github.com/ethersphere/swarm/pss/trojan"
)

// Pss is the top-level struct, which takes care of message sending
type Pss struct {
	localStore chunk.Store
}

// State exposes underlying chunk states
type State struct {
	stateStored chunk.State // chunk stored locally
	stateSent   chunk.State // chunk sent to neighbourhood
	stateSynced chunk.State // proof is received; chunk removed from sync db; chunk is available everywhere
}

// NewPss inits the Pss struct with the localstore
func NewPss(localStore chunk.Store) *Pss {
	return &Pss{
		localStore: localStore,
	}
}

// Send constructs a padded message with topic and payload,
// wraps it in a trojan chunk such that one of the targets is a prefix of the chunk address
// stores this in localstore for push-sync to pick up and deliver
func (p *Pss) Send(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (chunk.Chunk, *chunk.Tag, error) {
	metrics.GetOrRegisterCounter("trojanchunk/send", nil).Inc(1)
	//construct Trojan Chunk
	m, err := trojan.NewMessage(topic, payload)
	if err != nil {
		return nil, nil, err
	}
	var tc chunk.Chunk
	tc, err = m.Wrap(targets)
	if err != nil {
		return nil, nil, err
	}

	tag := chunk.NewTag(1, "pss-chunks-tag", 0, false)
	tc.WithTagID(tag.Uid)

	// SAVE trojanChunk to localstore, if it exists do nothing as it's already peristed
	if _, err = p.localStore.Put(ctx, chunk.ModePutUpload, tc); err != nil {
		return nil, nil, err
	}
	//p.tag.Inc(chunk.StateStored)

	return tc, tag, nil
}
