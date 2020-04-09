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

package pss

import (
	"context"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/chunk"
	trojan "github.com/ethersphere/swarm/pss/trojan"
)

// Pss is the top-level struct, which takes care of message sending, receiving, decryption and encryption, message handler dispatchers
// and message forwarding. Implements node.Service
type Pss struct {
	localStore chunk.Store
}

// NewPss inits the Pss struct with the localstore
func NewPss(localStore chunk.Store) *Pss {
	return &Pss{
		localStore: localStore,
	}
}

// Send a message without encryption
// Generate a trojan chunk envelope and is stored in localstore for desired targets to mine this chunk and retrieve message
func (p *Pss) Send(ctx context.Context, targets [][]byte, topic string, payload []byte) (chunk.Chunk, error) {
	metrics.GetOrRegisterCounter("trojanchunk/send", nil).Inc(1)
	//construct Trojan Chunk
	t := trojan.NewTopic(topic)
	m, err := trojan.NewMessage(t, payload)
	if err != nil {
		return nil, err
	}
	var tc chunk.Chunk
	tc, err = trojan.Wrap(targets, m)
	if err != nil {
		return nil, err
	}

	//SAVE trojanChunk to localstore, if it exists do nothing as it's already peristed
	//TODO: for second phase, use tags --> listen for response of recipient, recipient offline
	if _, err = p.localStore.Put(ctx, chunk.ModePutUpload, tc); err != nil {
		return nil, err
	}

	return tc, nil
}
