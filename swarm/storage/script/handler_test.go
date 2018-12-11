package script_test

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/script"

	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

func TestHandler(t *testing.T) {

	handler, cleanup := script.NewTestHandler(t)
	defer cleanup()

	sb := vm.NewScriptBuilder()
	sb.AddOp(vm.OP_3, vm.OP_ADD, vm.OP_5, vm.OP_EQUAL)
	scriptKey, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

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
