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

type trojanHeaders struct {
	span           []byte
	nonce          []byte
	decryptionHint []byte
}

type trojanMessage struct {
	trojanHeaders
	pssMsgCyphertext message.Message
}

// creates a new trojan message structure
// determines the nonce so that when the message is hashed, it falls in the neighbourhood of the given address
func newTrojanMessage(address chunk.Address, pssMessage message.Message) *trojanMessage {
	// set initial headers
	trojanHeaders := newTrojanHeaders(pssMessage)
	// find nonce for headers and address
	findMessageNonce(address, trojanHeaders)
	// cypher pss message, plain for now
	pssMsgCyphertext := pssMessage
	return &trojanMessage{
		trojanHeaders:    trojanHeaders,
		pssMsgCyphertext: pssMsgCyphertext,
	}
}

func newTrojanHeaders(pssMessage message.Message) *trojanHeaders {
	// create span, empty for now
	span := make([]byte, 8)
	// create initial nonce
	nonce := make([]byte, 32)
	// create decryption hint, empty for now
	decryptionHint := make([]byte, 32)

	return &trojanHeaders{
		span:           span,
		nonce:          nonce,
		decryptionHint: decryptionHint,
	}
}

var emptyChunk = chunk.NewChunk([]byte{}, []byte{})

// creates a new addressed chunk structure with the given trojan message content serialized as its data
func newTrojanChunk(address chunk.Address, message trojanMessage) (chunk.Chunk, error) {
	chunkData, err := json.Marshal(message) // what is the correct way of serializing a trojan message?
	if err != nil {
		return emptyChunk, err
	}
	return chunk.NewChunk(address, chunkData), nil
}
