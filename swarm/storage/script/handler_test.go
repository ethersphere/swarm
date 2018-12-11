package script_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/script"

	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

func TestHandler(t *testing.T) {

	handler, cleanup := NewTestHandler(t)
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

}
