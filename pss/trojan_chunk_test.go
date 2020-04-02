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
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/chunk"
)

// arbitrary address for tests
var addr = chunk.Address{
	67, 120, 209, 156, 38, 89, 15, 26, 129, 142,
	215, 214, 166, 44, 56, 9, 225, 73, 176, 153,
	56, 171, 92, 229, 242, 98, 51, 179, 180, 35,
	191, 140}

// arbitrary message for tests
func newTrojanMessage(t *testing.T) trojanMessage {
	// arbitrary payload
	payload := []byte("foopayload")
	payloadLength := int32(len(payload))

	// get length as bytes
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
		topic:   crypto.Keccak256([]byte("RECOVERY")), // TODO: will this always hash to 32 bytes?
		payload: payload,
		padding: padding,
	}
}

func TestNewTrojanChunk(t *testing.T) {
	_, err := newTrojanChunk(addr, newTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetNonce(t *testing.T) {
	tc, err := newTrojanChunk(addr, newTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}
	tc.setNonce()
}

func TestTrojanDataSerialization(t *testing.T) {
	tc, err := newTrojanChunk(addr, newTrojanMessage(t))
	if err != nil {
		t.Fatal(err)
	}
	tc.setNonce()
	td := tc.trojanData

	std, err := json.Marshal(td)
	if err != nil {
		t.Fatal(err)
	}

	var dtd *trojanData
	err = json.Unmarshal(std, &dtd)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(td, dtd) {
		t.Fatalf("original trojan data does not match deserialized one")
	}
}
