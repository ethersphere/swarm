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
	"encoding/binary"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/chunk"
)

// arbitrary address for tests
var addr = chunk.Address{
	57, 120, 209, 156, 38, 89, 19, 22, 129, 142,
	115, 215, 166, 45, 56, 9, 215, 73, 178, 153,
	36, 111, 93, 229, 222, 88, 51, 179, 181, 35,
	181, 144}

// newTrojanMessage creates an arbitrary trojan message for tests
func newTrojanMessage(t *testing.T) trojanMessage {
	// arbitrary payload
	payload := []byte("foopayload")
	payloadLength := uint16(len(payload))

	// get length as array of 8 bytes
	// TODO: better way of doing this?
	lengthBuffer := new(bytes.Buffer)
	if err := binary.Write(lengthBuffer, binary.BigEndian, payloadLength); err != nil {
		t.Fatal(err)
	}

	// set random bytes as padding
	paddingLength := 4056 - payloadLength
	padding := make([]byte, paddingLength)
	if _, err := rand.Read(padding); err != nil {
		t.Fatal(err)
	}

	return trojanMessage{
		length:  lengthBuffer.Bytes(),
		topic:   crypto.Keccak256([]byte("RECOVERY")),
		payload: payload,
		padding: padding,
	}
}

// TestNewTrojanChunk tests the creation of a trojan chunk
func TestNewTrojanChunk(t *testing.T) {
	_, err := newTrojanChunk(addr, newTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}
}

// TestSetNonce tests getting the correct nonce for a trojan chunk
func TestSetNonce(t *testing.T) {
	tc, err := newTrojanChunk(addr, newTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}
	tc.setNonce()
	// TODO: check nonce is correct for address
}

// TestTrojanDataSerialization tests that the trojanData type can be correctly serialized and deserialized
func TestTrojanDataSerialization(t *testing.T) {
	tc, err := newTrojanChunk(addr, newTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}
	tc.setNonce()
	td := tc.trojanData

	std, err := json.Marshal(&td)
	if err != nil {
		t.Fatal(err)
	}

	var dtd *trojanData
	err = json.Unmarshal(std, &dtd)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: why does this fail?: reflect.DeepEqual(td, dtd)
	if !td.equals(dtd) {
		t.Fatalf("original trojan data does not match deserialized one")
	}
}
