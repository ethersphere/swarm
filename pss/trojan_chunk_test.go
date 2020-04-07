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
	"reflect"
	"testing"

	"github.com/ethersphere/swarm/chunk"
)

// arbitrary address for tests
var testAddr = chunk.Address{
	57, 120, 209, 156, 38, 89, 19, 22, 129, 142,
	115, 215, 166, 45, 56, 9, 215, 73, 178, 153,
	36, 111, 93, 229, 222, 88, 51, 179, 181, 35,
	181, 144}

// newTestTrojanMessage creates an arbitrary trojan message for tests
func newTestTrojanMessage(t *testing.T) trojanMessage {
	payload := []byte("foopayload")
	tm, err := newTrojanMessage(newMessageTopic("RECOVERY"), payload)
	if err != nil {
		t.Fatal(err)
	}

	return tm
}

// TestNewTrojanChunk tests the creation of a trojan chunk
func TestNewTrojanChunk(t *testing.T) {
	_, err := newTrojanChunk(testAddr, newTestTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}
}

// TestFindNonce tests getting the correct nonce for a trojan chunk
func TestFindNonce(t *testing.T) {
	tm := newTestTrojanMessage(t)
	span := newTrojanChunkSpan()

	tm.findNonce(span, testAddr)
	// TODO: check nonce is correct for address
}

// TestTrojanMessageSerialization tests that the trojanMessage type can be correctly serialized and deserialized
func TestTrojanMessageSerialization(t *testing.T) {
	tm := newTestTrojanMessage(t)

	stm, err := tm.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	dtm := new(trojanMessage)
	err = dtm.UnmarshalBinary(stm)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tm, *dtm) {
		t.Fatalf("original trojan message does not match deserialized one")
	}
}
