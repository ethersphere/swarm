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
	"reflect"
	"testing"

	"github.com/ethersphere/swarm/chunk"
)

// arbitrary targets for tests
var testTargets = [][]byte{
	[]byte{57, 120},
	[]byte{209, 156},
	[]byte{156, 38},
	[]byte{89, 19},
	[]byte{22, 129}}

// arbitrary topic for tests
var testTopic = NewTopic("foo")

// newTestMessage creates an arbitrary message for tests
func newTestMessage(t *testing.T) message {
	payload := []byte("foopayload")
	m, err := newMessage(testTopic, payload)
	if err != nil {
		t.Fatal(err)
	}

	return m
}

// TestNewMessage tests the creation of a message
func TestNewMessage(t *testing.T) {
	smallPayload := make([]byte, 32)
	if _, err := newMessage(testTopic, smallPayload); err != nil {
		t.Fatal(err)
	}

	maxPayload := make([]byte, MaxPayloadSize)
	if _, err := newMessage(testTopic, maxPayload); err != nil {
		t.Fatal(err)
	}

	invalidPayload := make([]byte, MaxPayloadSize+1)
	if _, err := newMessage(testTopic, invalidPayload); err != errPayloadTooBig {
		t.Fatalf("expected error when creating trojan message of invalid size to be %q, but got %v", errPayloadTooBig, err)
	}
}

// TestNewTrojanChunk tests the creation of a trojan chunk
// its fields as a regular chunk should be correct
// its resulting address should have a prefix which matches one of the given targets
// its resulting payload should have a hash that matches its address exactly
func TestNewTrojanChunk(t *testing.T) {
	tc, err := newTrojanChunk(testTargets, newTestMessage(t))
	if err != nil {
		t.Fatal(err)
	}

	addr := tc.Address()
	addrLen := len(addr)

	if addrLen != chunk.AddressLength {
		t.Fatalf("trojan chunk payload has an unexpected address len of %d rather than %d", addrLen, chunk.AddressLength)
	}

	addrPrefix := addr[:len(testTargets[0])]

	if !contains(testTargets, addrPrefix) {
		t.Fatal("trojan chunk address prefix does not match any of the targets")
	}

	payload := tc.Data()
	payloadSize := len(payload)
	expectedSize := chunk.DefaultSize + 8 // payload + span

	if payloadSize != expectedSize {
		t.Fatalf("trojan chunk payload has an unexpected size of %d rather than %d", payloadSize, expectedSize)
	}

	span := binary.BigEndian.Uint64(payload[:8])
	remPayloadLen := len(payload[8:])

	if int(span) != remPayloadLen {
		t.Fatalf("trojan chunk span set to %d, but the rest of the chunk payload is of size %d", span, remPayloadLen)
	}

	payloadHash, err := hash(payload)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(addr, payloadHash) {
		t.Fatal("trojan chunk address does not match its payload hash")
	}
}

// TestNewTrojanChunk tests the creation of a trojan chunk fails when given targets are invalid
func TestNewTrojanChunkFailure(t *testing.T) {
	emptyTargets := [][]byte{}
	if _, err := newTrojanChunk(emptyTargets, newTestMessage(t)); err != errEmptyTargets {
		t.Fatalf("expected error when creating trojan chunk for empty targets to be %q, but got %v", errEmptyTargets, err)
	}

	varLenTargets := [][]byte{
		[]byte{34},
		[]byte{25, 120},
		[]byte{180, 18, 255},
	}
	if _, err := newTrojanChunk(varLenTargets, newTestMessage(t)); err != errVarLenTargets {
		t.Fatalf("expected error when creating trojan chunk for empty targets to be %q, but got %v", errVarLenTargets, err)
	}
}

// TestMessageSerialization tests that the message type can be correctly serialized and deserialized
func TestMessageSerialization(t *testing.T) {
	m := newTestMessage(t)

	sm, err := m.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	dsm := new(message)
	err = dsm.UnmarshalBinary(sm)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(m, *dsm) {
		t.Fatalf("original trojan message does not match deserialized one")
	}
}
