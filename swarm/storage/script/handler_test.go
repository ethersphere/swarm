// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package script_test

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/storage/script"
	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

func TestHandler(t *testing.T) {

	handler, cleanup := script.NewTestHandler(t)
	defer cleanup()

	// build a simple script. The key expects a number in the signature that added to 3 equals 5.
	sb := vm.NewScriptBuilder()
	sb.AddOp(vm.OP_3, vm.OP_ADD, vm.OP_5, vm.OP_EQUAL)
	scriptKey, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	// Build a sig script that provides the right answer
	sb = vm.NewScriptBuilder()
	sb.AddOp(vm.OP_2)
	scriptSig, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("Dad crédito a las obras y no a las palabras")

	chunk, err := script.NewChunk(scriptKey, scriptSig, payload)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := handler.Put(ctx, chunk); err != nil {
		t.Fatal(err)
	}

	retrievedChunk, err := handler.Get(ctx, chunk.Address())
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(retrievedChunk.Data(), chunk.Data()) {
		t.Fatalf("Expected retrieved chunk to contain the same data. Expected: %v, got %v", chunk.Data(), retrievedChunk.Data())
	}

	// Test JSON marshaller / unmarshaller
	expectedJSON := `{"address":"0x3b5ef6b1e92dfcaa84c47ea169aaf92db4b80611f6818b3296ad66f8826d56e6","scriptKey":{"binary":"0x53935587","script":"3 ADD 5 EQUAL"},"scriptSig":{"binary":"0x52","script":"2"},"data":"0x446164206372c3a96469746f2061206c6173206f627261732079206e6f2061206c61732070616c6162726173"}`
	jsonBytes, err := json.Marshal(chunk)
	JSONEquals(t, expectedJSON, string(jsonBytes))

	err = json.Unmarshal(jsonBytes, retrievedChunk)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(retrievedChunk, chunk) {
		t.Fatal("retrieved chunk from JSON does not match")
	}

	// Test address inference
	noAddressJSON := `{"scriptKey":{"binary":"0x53935587","script":"OP_3 OP_ADD OP_5 OP_EQUAL"},"scriptSig":{"binary":"0x52","script":"OP_2"},"data":"0x446164206372c3a96469746f2061206c6173206f627261732079206e6f2061206c61732070616c6162726173"}`

	retrieved := new(script.Chunk)
	err = json.Unmarshal([]byte(noAddressJSON), retrieved)
	if err != nil {
		t.Fatal(err)
	}
	jsonBytes, err = json.Marshal(retrieved)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(retrieved.Address(), chunk.Address()) {
		t.Fatal("Expected address to match")
	}

}
func TestChunkSizeCheck(t *testing.T) {
	_, err := script.NewChunk(make([]byte, 2000), make([]byte, 2000), make([]byte, 2000))
	if err != script.ErrChunkTooBig {
		t.Fatalf("Expected NewChunk to fail with ErrChunkTooBig, got %s", err)
	}
	c := new(script.Chunk)
	if err = c.UnmarshalBinary(make([]byte, chunk.DefaultSize+100)); err != script.ErrChunkTooBig {
		t.Fatalf("Expected UnmarshalBinary to fail with ErrChunkTooBig, got %s", err)
	}
}
