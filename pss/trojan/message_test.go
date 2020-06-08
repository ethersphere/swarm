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
	"encoding/binary"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	chunktesting "github.com/ethersphere/swarm/chunk/testing"
)

// arbitrary targets for tests
var t1 = Target([]byte{57, 120})
var t2 = Target([]byte{209, 156})
var t3 = Target([]byte{156, 38})
var t4 = Target([]byte{89, 19})
var t5 = Target([]byte{22, 129})
var testTargets = Targets([]Target{t1, t2, t3, t4, t5})

// arbitrary topic for tests
var testTopic = NewTopic("foo")

// newTestMessage creates an arbitrary Message for tests
func newTestMessage(t *testing.T) Message {
	payload := []byte("foopayload")
	m, err := NewMessage(testTopic, payload)
	if err != nil {
		t.Fatal(err)
	}

	return m
}

// TestNewMessage tests the correct and incorrect creation of a Message struct
func TestNewMessage(t *testing.T) {
	smallPayload := make([]byte, 32)
	m, err := NewMessage(testTopic, smallPayload)
	if err != nil {
		t.Fatal(err)
	}

	// verify topic
	if m.Topic != testTopic {
		t.Fatalf("expected message topic to be %v but is %v instead", testTopic, m.Topic)
	}

	maxPayload := make([]byte, MaxPayloadSize)
	if _, err := NewMessage(testTopic, maxPayload); err != nil {
		t.Fatal(err)
	}

	// the creation should fail if the payload is too big
	invalidPayload := make([]byte, MaxPayloadSize+1)
	if _, err := NewMessage(testTopic, invalidPayload); err != ErrPayloadTooBig {
		t.Fatalf("expected error when creating message of invalid payload size to be %q, but got %v", ErrPayloadTooBig, err)
	}
}

// TestWrap tests the creation of a chunk from a list of targets
// its address length and span should be correct
// its resulting address should have a prefix which matches one of the given targets
// its resulting data should have a hash that matches its address exactly
func TestWrap(t *testing.T) {
	m := newTestMessage(t)
	c, err := m.Wrap(testTargets)
	if err != nil {
		t.Fatal(err)
	}

	addr := c.Address()
	addrLen := len(addr)
	if addrLen != chunk.AddressLength {
		t.Fatalf("chunk has an unexpected address length of %d rather than %d", addrLen, chunk.AddressLength)
	}

	addrPrefix := addr[:len(testTargets[0])]
	if !contains(testTargets, addrPrefix) {
		t.Fatal("chunk address prefix does not match any of the targets")
	}

	data := c.Data()
	dataSize := len(data)
	expectedSize := 8 + chunk.DefaultSize // span + payload
	if dataSize != expectedSize {
		t.Fatalf("chunk data has an unexpected size of %d rather than %d", dataSize, expectedSize)
	}

	span := binary.LittleEndian.Uint64(data[:8])
	remDataLen := len(data[8:])
	if int(span) != remDataLen {
		t.Fatalf("chunk span set to %d, but rest of chunk data is of size %d", span, remDataLen)
	}

	dataHash, err := hash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(addr, dataHash) {
		t.Fatal("chunk address does not match its data hash")
	}
}

// TestWrapFail tests that the creation of a chunk fails when given targets are invalid
func TestWrapFail(t *testing.T) {
	m := newTestMessage(t)

	emptyTargets := Targets([]Target{})
	if _, err := m.Wrap(emptyTargets); err != ErrEmptyTargets {
		t.Fatalf("expected error when creating chunk for empty targets to be %q, but got %v", ErrEmptyTargets, err)
	}

	t1 := Target([]byte{34})
	t2 := Target([]byte{25, 120})
	t3 := Target([]byte{180, 18, 255})
	varLenTargets := Targets([]Target{t1, t2, t3})
	if _, err := m.Wrap(varLenTargets); err != ErrVarLenTargets {
		t.Fatalf("expected error when creating chunk for variable-length targets to be %q, but got %v", ErrVarLenTargets, err)
	}
}

// TestPadBytes tests that different types of byte slices are correctly padded with leading 0s
// all slices are interpreted as big-endian
func TestPadBytes(t *testing.T) {
	s := make([]byte, 32)

	// empty slice should be unchanged
	p := padBytes(s)
	if !bytes.Equal(p, s) {
		t.Fatalf("expected byte padding to result in %x, but is %x", s, p)
	}

	// slice of length 3
	s = []byte{255, 128, 64}
	p = padBytes(s)
	e := append(make([]byte, 29), s...) // 29 zeros plus the 3 original bytes
	if !bytes.Equal(p, e) {
		t.Fatalf("expected byte padding to result in %x, but is %x", e, p)
	}

	// simulate toChunk behavior
	s = []byte{255, 255, 255}
	i := new(big.Int).SetBytes(s) // byte slice to big.Int
	i.Add(i, big.NewInt(1))       // add 1 to slice as big.Int
	s = i.Bytes()                 // []byte{1, 0, 0, 0}

	p = padBytes(s)
	e = append(make([]byte, 28), s...) // 28 zeros plus the 4 original bytes
	if !bytes.Equal(p, e) {
		t.Fatalf("expected byte padding to result in %x, but is %x", e, p)
	}
}

// TestUnwrap tests the correct unwrapping of a trojan chunk to obtain a message
func TestUnwrap(t *testing.T) {
	m := newTestMessage(t)
	c, err := m.Wrap(testTargets)
	if err != nil {
		t.Fatal(err)
	}

	um, err := Unwrap(c)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(m, *um) {
		t.Fatalf("original message does not match unwrapped one")
	}
}

// TestIsPotential tests if chunk are correctly interpreted as potentially trojan
func TestIsPotential(t *testing.T) {
	c := chunktesting.GenerateTestRandomChunk()

	// invalid type
	c.WithType(chunk.Unknown)
	if IsPotential(c) {
		t.Fatal("non content-address chunk marked as potential trojan")
	}

	// correct type, but invalid length
	c.WithType(chunk.ContentAddressed)
	length := len(c.Data()) + 1
	lengthBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBuf, uint16(length))
	// put into bytes #41 and #42
	copy(c.Data()[40:42], lengthBuf[:])
	if IsPotential(c) {
		t.Fatal("chunk with invalid trojan message length marked as potential trojan")
	}
}

// TestMessageSerialization tests that the Message type can be correctly serialized and deserialized
func TestMessageSerialization(t *testing.T) {
	m := newTestMessage(t)

	sm, err := m.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	dsm := new(Message)
	err = dsm.UnmarshalBinary(sm)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(m, *dsm) {
		t.Fatalf("original message does not match deserialized one")
	}
}
