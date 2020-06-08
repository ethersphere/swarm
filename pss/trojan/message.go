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

package trojan

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
)

// Topic is an alias for a 32 byte fixed-size array which contains an encoding of a message topic
type Topic [32]byte

// Message represents a trojan message, which is a message that will be hidden within a chunk payload as part of its data
type Message struct {
	length  [2]byte // big-endian encoding of Message payload length
	Topic   Topic
	Payload []byte // contains the chunk address to be repaired
	padding []byte
}

// Target is an alias for an address which can be mined to construct a trojan message
type Target []byte

// Targets is an alias for a collection of targets
type Targets []Target

// MaxPayloadSize is the maximum allowed payload size for the Message type, in bytes
const MaxPayloadSize = 4030

var hashFunc = storage.MakeHashFunc(storage.BMTHash)

// ErrPayloadTooBig is returned when a given payload for a Message type is longer than the maximum amount allowed
var ErrPayloadTooBig = fmt.Errorf("message payload size cannot be greater than %d bytes", MaxPayloadSize)

// ErrEmptyTargets is returned when the given target list for a trojan chunk is empty
var ErrEmptyTargets = errors.New("target list cannot be empty")

// ErrVarLenTargets is returned when the given target list for a trojan chunk has addresses of different lengths
var ErrVarLenTargets = errors.New("target list cannot have targets of different length")

// NewTopic creates a new Topic variable with the given input string
// the input string is taken as a byte slice and hashed
func NewTopic(topic string) Topic {
	return Topic(crypto.Keccak256Hash([]byte(topic)))
}

// NewMessage creates a new Message variable with the given topic and payload
// it finds a length and nonce for the message according to the given input and maximum payload size
func NewMessage(topic Topic, payload []byte) (Message, error) {
	if len(payload) > MaxPayloadSize {
		return Message{}, ErrPayloadTooBig
	}

	// get length as array of 2 bytes
	payloadSize := uint16(len(payload))
	lengthBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBuf, payloadSize)

	// set random bytes as padding
	paddingLen := MaxPayloadSize - payloadSize
	padding := make([]byte, paddingLen)
	if _, err := rand.Read(padding); err != nil {
		return Message{}, err
	}

	// create new Message var and set fields
	m := new(Message)
	copy(m.length[:], lengthBuf[:])
	m.Topic = topic
	m.Payload = payload
	m.padding = padding

	return *m, nil
}

// Wrap creates a new trojan chunk for the given targets and Message
// a trojan chunk is a content-addressed chunk made up of span, a nonce, and a payload which contains the Message
// the chunk address will have one of the targets as its prefix and thus will be forwarded to the neighbourhood of the recipient overlay address the target is derived from
func (m *Message) Wrap(targets Targets) (chunk.Chunk, error) {
	if err := checkTargets(targets); err != nil {
		return nil, err
	}

	span := make([]byte, 8)
	// 4064 bytes for trojan message + 32 bytes for nonce = 4096 bytes as payload for resulting chunk
	binary.LittleEndian.PutUint64(span, chunk.DefaultSize)

	return m.toChunk(targets, span)
}

// Unwrap creates a new trojan message from the given chunk payload
// this function assumes the chunk has been validated as a content-addressed chunk
// it will return the resulting message if the unwrapping is successful, and an error otherwise
func Unwrap(c chunk.Chunk) (*Message, error) {
	d := c.Data()

	// unmarshal chunk payload into message
	m := new(Message)
	// first 40 bytes are span + nonce
	err := m.UnmarshalBinary(d[40:])

	return m, err
}

// Valid returns true if the given chunk is a potential trojan
func Valid(c chunk.Chunk) bool {
	if c.Type() != chunk.ContentAddressed {
		return false
	}
	// check for minimum chunk data length
	// span (8) + nonce (32) + length (2) + topic (32) = 74
	return len(c.Data()) >= 74
}

// checkTargets verifies that the list of given targets is non empty and with elements of matching size
func checkTargets(targets Targets) error {
	if len(targets) == 0 {
		return ErrEmptyTargets
	}
	validLen := len(targets[0]) // take first element as allowed length
	for i := 1; i < len(targets); i++ {
		if len(targets[i]) != validLen {
			return ErrVarLenTargets
		}
	}
	return nil
}

// toChunk finds a nonce so that when the given trojan chunk fields are hashed, the result will fall in the neighbourhood of one of the given targets
// this is done by iteratively enumerating different nonces until the BMT hash of the serialization of the trojan chunk fields results in a chunk address that has one of the targets as its prefix
// the function returns a new chunk, with the found matching hash to be used as its address,
// and its data set to the serialization of the trojan chunk fields which correctly hash into the matching address
func (m *Message) toChunk(targets Targets, span []byte) (chunk.Chunk, error) {
	// start out with random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	nonceInt := new(big.Int).SetBytes(nonce)
	targetsLen := len(targets[0])

	// serialize message
	b, err := m.MarshalBinary() // TODO: this should be encrypted
	if err != nil {
		return nil, err
	}

	// hash chunk fields with different nonces until an acceptable one is found
	// TODO: prevent infinite loop
	for {
		s := append(append(span, nonce...), b...) // serialize chunk fields
		hash, err := hash(s)
		if err != nil {
			return nil, err
		}

		// take as much of the hash as the targets are long
		if contains(targets, hash[:targetsLen]) {
			// if nonce found, stop loop and return chunk
			return chunk.NewChunk(hash, s), nil
		}
		// else, add 1 to nonce and try again
		nonceInt.Add(nonceInt, big.NewInt(1))
		// loop around in case of overflow
		if nonceInt.BitLen() > 256 {
			nonceInt = big.NewInt(0)
		}
		nonce = padBytes(nonceInt.Bytes()) // pad in case Bytes call is not 32 bytes long
	}
}

// hash hashes the given serialization of chunk fields with the hashing func
func hash(s []byte) ([]byte, error) {
	hasher := hashFunc()
	hasher.SetSpanBytes(s[:8])
	if _, err := hasher.Write(s[8:]); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

// contains returns whether the given collection contains the given elem
func contains(col Targets, elem []byte) bool {
	for i := range col {
		if bytes.Equal(elem, col[i]) {
			return true
		}
	}
	return false
}

// padBytes adds 0s to the given byte slice as left padding,
// returning this as a new byte slice with a length of exactly 32
// given param is assumed to be at most 32 bytes long
func padBytes(b []byte) []byte {
	l := len(b)
	if l == 32 {
		return b
	}
	bb := make([]byte, 32)
	copy(bb[32-l:], b)
	return bb
}

// MarshalBinary serializes a message struct
func (m *Message) MarshalBinary() (data []byte, err error) {
	data = append(m.length[:], m.Topic[:]...)
	data = append(data, m.Payload...)
	data = append(data, m.padding...)
	return data, nil
}

// UnmarshalBinary deserializes a message struct
func (m *Message) UnmarshalBinary(data []byte) (err error) {
	copy(m.length[:], data[:2])  // first 2 bytes are length
	copy(m.Topic[:], data[2:34]) // following 32 bytes are topic

	// rest of the bytes are payload and padding
	length := binary.BigEndian.Uint16(m.length[:])
	payloadEnd := 34 + length
	m.Payload = data[34:payloadEnd]
	m.padding = data[payloadEnd:]
	return nil
}
