// Copyright 2020 The go-ethereum Authors
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
	"encoding/json"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/pss/message"
)

type pssEnvelope struct {
	// headers ? missing
	message []byte
}

type trojanMessage struct {
	span             [8]byte
	nonce            [32]byte
	decryptionHint   [32]byte
	pssMsgCyphertext message.Message
}

var emptyChunk = chunk.NewChunk([]byte{}, []byte{})

func newTrojanChunk(address chunk.Address, message trojanMessage) (chunk.Chunk, error) {
	chunkData, err := json.Marshal(message) // what is the correct way of serializing a trojan message?
	if err != nil {
		return emptyChunk, err
	}
	return chunk.NewChunk(address, chunkData), nil
}
