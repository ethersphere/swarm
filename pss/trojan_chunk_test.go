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

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/pss/message"
)

func TestFindMessageNonce(T *testing.T) {
	findMessageNonce(chunk.Address{}, newTrojanHeaders())
}

func TestTrojanMessageSerialization(t *testing.T) {
	addr := chunk.Address{
		67, 120, 209, 156, 38, 89, 15, 26, 129, 142,
		215, 214, 166, 44, 56, 9, 225, 73, 176, 153,
		56, 171, 92, 229, 242, 98, 51, 179, 180, 35,
		191, 140}
	msg := message.Message{
		To:      []byte{},
		Flags:   message.Flags{},
		Expire:  0,
		Topic:   message.NewTopic([]byte("footopic")),
		Payload: []byte("foopayload"),
	}
	tm, err := newTrojanMessage(addr, msg)
	if err != nil {
		t.Fatal(err)
	}

	stm, err := json.Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}

	var dtm *trojanMessage
	err = json.Unmarshal(stm, &dtm)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tm, dtm) {
		t.Fatalf("original trojan message does not match deserialized one")
	}
}
