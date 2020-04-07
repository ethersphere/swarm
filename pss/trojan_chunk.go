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
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
)

// MessageTopic is an alias for a 32 byte fixed-size array which contains an encoding of a message topic
type MessageTopic [32]byte

type trojanMessage struct {
	length  [2]byte // big-endian encoding of message length
	topic   MessageTopic
	payload []byte
	padding []byte
}

const trojanPayloadMaxSize = 4064 // in bytes

// newMessageTopic creates a new MessageTopic variable with the given input string
// the input string is taken as a byte slice and hashed
func newMessageTopic(topic string) MessageTopic {
	// TODO: is it ok to use this instead of `crypto.Keccak256`?
	return MessageTopic(crypto.Keccak256Hash([]byte(topic)))
}

// newTrojanMessage creates a new trojanMessage variable with the given topic and message payload
func newTrojanMessage(topic MessageTopic, payload []byte) (trojanMessage, error) {
	if len(payload) > 4064 {
		return trojanMessage{}, fmt.Errorf("trojan message payload cannot be greater than %d bytes", trojanPayloadMaxSize)
	}

	// get length as array of 2 bytes
	payloadLength := uint16(len(payload))
	lengthBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBuffer, payloadLength)

	// set random bytes as padding
	paddingLength := trojanPayloadMaxSize - payloadLength
	padding := make([]byte, paddingLength)
	if _, err := rand.Read(padding); err != nil {
		return trojanMessage{}, err
	}

	// create new trojan message var and set fields
	tm := new(trojanMessage)
	copy(tm.length[:], lengthBuffer[:])
	tm.payload = payload
	tm.padding = padding

	return *tm, nil
}

// newTrojanChunk creates a new trojan chunk for the given address and trojan message
// TODO: discuss if instead of receiving a trojan message, we should receive a byte slice as payload
func newTrojanChunk(address chunk.Address, message trojanMessage) (chunk.Chunk, error) {
	// create span
	span := newTrojanChunkSpan()

	// find nonce for trojan chunk
	nonce, err := message.findNonce(span, address)
	if err != nil {
		return nil, err
	}

	// serialize trojan message struct
	m, err := message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// serialize trojan chunk fields to be used as payload for chunk
	chunkData, err := serializeTrojanChunk(span, nonce, m)
	if err != nil {
		return nil, err
	}

	return chunk.NewChunk(address, chunkData), nil
}

// newTrojanChunkSpan creates a pre-set 8-byte span for a trojan chunk
func newTrojanChunkSpan() []byte {
	span := make([]byte, 8)
	binary.BigEndian.PutUint64(span, 4096) // TODO: should this be little-endian?
	return span
}

// serializeTrojanChunk appends the span, nonce and payload of a trojan message and returns the result
// this can be used:
// - to form the payload for a regular chunk
// - to be used as the input for trojan chunk hash calculation
func serializeTrojanChunk(span, nonce, payload []byte) ([]byte, error) {
	h := append(span, nonce...)
	s := append(h, payload...)

	return s, nil
}

// findNonce determines the nonce so that when the given trojan chunk fields are hashed, the result will fall in the neighbourhood of the given address
// this is done by iterating the BMT hash of the serialization of a trojan chunk until the desired nonce is found
func (tm *trojanMessage) findNonce(span []byte, addr chunk.Address) ([]byte, error) {
	emptyNonce := []byte{}

	// init BMT hash function
	hashFunc := storage.MakeHashFunc(storage.BMTHash)()

	// start out with random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return emptyNonce, err
	}

	// serialize trojan message
	m, err := tm.MarshalBinary()
	if err != nil {
		return emptyNonce, err
	}

	// hash trojan chunk fields with different nonces until a desired one is found
	hashWithinNeighbourhood := false // TODO: this could be correct on the 1st try
	// TODO: prevent infinite loop
	for hashWithinNeighbourhood != true {
		s, _ := serializeTrojanChunk(span, nonce, m) // err always nil here
		if _, err := hashFunc.Write(s); err != nil {
			return emptyNonce, err
		}
		hash := hashFunc.Sum(nil)

		// TODO: what is the correct way to check if hash is in the same neighbourhood as trojan chunk address?
		_ = chunk.Proximity(addr, hash)

		// TODO: replace placeholder condition
		if true {
			// if nonce found, stop loop
			hashWithinNeighbourhood = true
		} else {
			// else, add 1 to nonce and try again
			// TODO: test loop-around
			nonceInt := new(big.Int).SetBytes(nonce)
			nonce = nonceInt.Add(nonceInt, big.NewInt(1)).Bytes()
		}
	}

	return nonce, nil
}

// MarshalBinary serializes a trojanMessage struct
func (tm *trojanMessage) MarshalBinary() (data []byte, err error) {
	m := append(tm.length[:], tm.topic[:]...)
	m = append(m, tm.payload...)
	m = append(m, tm.padding...)
	return m, nil
}

// UnmarshalBinary deserializes a trojanMesage struct
func (tm *trojanMessage) UnmarshalBinary(data []byte) (err error) {
	copy(tm.length[:], data[:2])  // first 2 bytes are length
	copy(tm.topic[:], data[2:34]) // following 32 bytes are topic

	// rest of the bytes are payload and padding
	length := binary.BigEndian.Uint16(tm.length[:])
	payloadEnd := 34 + length
	tm.payload = data[34:payloadEnd]
	tm.padding = data[payloadEnd:]
	return nil
}
