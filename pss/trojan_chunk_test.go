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
	"encoding/json"
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
var msg = trojanMessage{
	length:  []byte{}, // TODO: how to set this value?
	topic:   crypto.Keccak256([]byte("RECOVERY")),
	payload: []byte("foopayload"),
	padding: []byte{}, // TODO: how to set this value?
}

func TestNewTrojanChunk(t *testing.T) {
	_, err := newTrojanChunk(addr, msg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetNonce(t *testing.T) {
	tc, err := newTrojanChunk(addr, msg)
	if err != nil {
		t.Fatal(err)
	}
	tc.setNonce()
}

func TestTrojanDataSerialization(t *testing.T) {
	tc, err := newTrojanChunk(addr, msg)
	if err != nil {
		t.Fatal(err)
	}
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
