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
	"reflect"
	"testing"
)

// arbitrary targets for tests
var testTargets = [][]byte{
	[]byte{57, 120},
	[]byte{209, 156},
	[]byte{156, 38},
	[]byte{89, 19},
	[]byte{22, 129}}

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
	tc, err := newTrojanChunk(testTargets, newTestTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}

	addr := tc.Address()
	addrPrefix := addr[:len(testTargets[0])]

	if !hashPrefixInTargets(addrPrefix, testTargets) {
		t.Fatal("trojan chunk address prefix does not match any of the targets")
	}

	payload := tc.Data()
	payloadHash, err := hashTrojanChunk(payload)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(addr, payloadHash) {
		t.Fatal("trojan chunk address does not match its payload hash")
	}
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
