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

// messageTopic is an alias for a 32 byte fixed-size array which contains an encoding of a message topic
type messageTopic [32]byte

type trojanMessage struct {
	length  [2]byte // big-endian encoding of message length
	topic   messageTopic
	payload []byte
	padding []byte
}

const trojanPayloadMaxSize = 4064                              // in bytes
var trojanHashingFunc = storage.MakeHashFunc(storage.SHA3Hash) // TODO: make this work with storage.BMTHash

// newMessageTopic creates a new messageTopic variable with the given input string
// the input string is taken as a byte slice and hashed
func newMessageTopic(topic string) messageTopic {
	// TODO: is it ok to use this instead of `crypto.Keccak256`?
	return messageTopic(crypto.Keccak256Hash([]byte(topic)))
}

// newTrojanMessage creates a new trojanMessage variable with the given topic and message payload
// it finds a length and nonce for the trojanMessage according to the given input and maximum payload size
func newTrojanMessage(topic messageTopic, payload []byte) (trojanMessage, error) {
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
	if err := checkTargets(targets); err != nil {
		return nil, err
	}

	// create span
	span := newTrojanChunkSpan()

	// iterate trojan chunk fields to find coherent address and payload for chunk
	addr, payload, err := iterateTrojanChunk(targets, span, message)
	if err != nil {
		return nil, err
	}

	return chunk.NewChunk(addr, payload), nil
}

// checkTargets verifies that the list of given targets is non empty and with elements of matching size
func checkTargets(targets [][]byte) error {
	if len(targets) == 0 {
		return fmt.Errorf("target list cannot be empty")
	}
	validLength := len(targets[0]) // take first element as allowed length
	for i := 1; i < len(targets); i++ {
		if len(targets[i]) != validLength {
			return fmt.Errorf("target list cannot have targets of different length")
		}
	}
	return nil
}

// newTrojanChunkSpan creates a pre-set 8-byte span for a trojan chunk
func newTrojanChunkSpan() []byte {
	span := make([]byte, 8)
	binary.BigEndian.PutUint64(span, chunk.DefaultSize) // TODO: should this be little-endian?
	return span
}

// iterateTrojanChunk finds a nonce so that when the given trojan chunk fields are hashed, the result will fall in the neighbourhood of one of the given targets
// this is done by iterating the BMT hash of the serialization of the trojan chunk fields until the desired nonce is found
// the function returns the matching hash to be used as an address for a regular chunk, plus its payload
// the payload is the serialization of the trojan chunk fields which correctly hash into the matching address
func iterateTrojanChunk(targets [][]byte, span []byte, message trojanMessage) (addr, payload []byte, err error) {
	// start out with random nonce
	nonce := make([]byte, 32)
	if _, errRand := rand.Read(nonce); err != nil {
		return nil, nil, errRand
	}
	nonceInt := new(big.Int).SetBytes(nonce)
	targetsLength := len(targets[0])

	// serialize trojan message
	m, marshalErr := message.MarshalBinary() // TODO: this should be encrypted
	if marshalErr != nil {
		return nil, nil, marshalErr
	}

	// hash trojan chunk fields with different nonces until an acceptable one is found
	// TODO: prevent infinite loop
	for {
		s := append(append(span, nonce...), m...) // err always nil here
		hash, hashErr := hashTrojanChunk(s)
		if hashErr != nil {
			return nil, nil, hashErr
		}

		// take as much of the hash as the targets are long
		if hashPrefixInTargets(hash[:targetsLength], targets) {
			// if nonce found, stop loop and return matching hash as address
			return hash, s, nil
		}
		// else, add 1 to nonce and try again
		// TODO: test loop-around
		nonceInt.Add(nonceInt, big.NewInt(1))
		nonce = nonceInt.Bytes()
	}
}

// hashTrojanChunk hashes the serialization of trojan chunk fields with the trojan hashing func
func hashTrojanChunk(s []byte) ([]byte, error) {
	hasher := trojanHashingFunc()
	hasher.Reset()
	if _, err := hasher.Write(s); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

// hashPrefixInTargets returns whether the given hash prefix appears in the targets given as a collection of byte slices
func hashPrefixInTargets(hashPrefix []byte, targets [][]byte) bool {
	for i := range targets {
		if bytes.Equal(hashPrefix, targets[i]) {
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
