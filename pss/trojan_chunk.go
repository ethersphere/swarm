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
	"bytes"
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
	if len(payload) > trojanPayloadMaxSize {
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

// newTrojanChunk creates a new trojan chunk for the given targets and trojan message
// TODO: discuss if instead of receiving a trojan message, we should receive a byte slice as payload
func newTrojanChunk(targets [][]byte, message trojanMessage) (chunk.Chunk, error) {
	// create span
	span := newTrojanChunkSpan()

	// find nonce for trojan chunk
	nonce, target, err := message.findNonce(span, targets)
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

	return chunk.NewChunk(target, chunkData), nil
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
// the matching hash is also returned as the target
func (tm *trojanMessage) findNonce(span []byte, targets [][]byte) (nonce, target []byte, err error) {
	emptyNonce, emptyTarget := []byte{}, []byte{}
	// init BMT hash function
	hashFunc := storage.MakeHashFunc(storage.BMTHash)()

	// start out with random nonce
	nonce = make([]byte, 32)
	if _, randErr := rand.Read(nonce); randErr != nil {
		return emptyNonce, emptyTarget, randErr
	}

	// serialize trojan message
	m, marshalErr := tm.MarshalBinary()
	if marshalErr != nil {
		return emptyNonce, emptyTarget, marshalErr
	}

	// hash trojan chunk fields with different nonces until an acceptable one is found
	// TODO: prevent infinite loop
	for {
		hash, hashErr := hashTrojanChunk(span, nonce, m, hashFunc)
		if hashErr != nil {
			return emptyNonce, emptyTarget, hashErr
		}

		if hashInTargets(hash, targets) {
			// if nonce found, stop loop and return matching hash as target
			return nonce, target, nil
		}
		// else, add 1 to nonce and try again
		// TODO: test loop-around
		nonceInt := new(big.Int).SetBytes(nonce)
		nonce = nonceInt.Add(nonceInt, big.NewInt(1)).Bytes()
	}
}

func hashTrojanChunk(span, nonce, payload []byte, hashFunc storage.SwarmHash) ([]byte, error) {
	s, _ := serializeTrojanChunk(span, nonce, payload) // err always nil here
	if _, err := hashFunc.Write(s); err != nil {
		return []byte{}, err
	}
	return hashFunc.Sum(nil), nil
}

func hashInTargets(hash []byte, targets [][]byte) bool {
	for i := range targets {
		if bytes.Equal(hash, targets[i]) {
			return true
		}
	}
	return false
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
