package storage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/contracts/ens"
)

var (
	hashfunc = MakeHashFunc(DefaultHash)
)

// tests that the content address validator correctly checks the data
// tests that resource update chunks are passed through content address validator
// the test checking the resouce update validator internal correctness is found in resource_test.go
func TestValidator(t *testing.T) {
	t.Skip("just for now")
	// set up localstore
	datadir, err := ioutil.TempDir("", "storage-testvalidator")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(datadir)

	params := NewDefaultLocalStoreParams()
	params.Init(datadir)
	store, err := NewLocalStore(params, nil)
	if err != nil {
		t.Fatal(err)
	}

	// check puts with no validators, both succeed
	chunks := GenerateRandomChunks(259, 2)
	goodChunk := chunks[0]
	badChunk := chunks[1]
	copy(badChunk.Data(), goodChunk.Data())

	err = mputChunks(store, goodChunk)
	if err != nil {
		t.Fatalf("expected no error on good content address chunk in spite of no validation, but got: %s", err)
	}
	err = mputChunks(store, badChunk)
	if err != nil {
		t.Fatalf("expected no error on bad content address chunk in spite of no validation, but got: %s", err)
	}

	// add content address validator and check puts
	// bad should fail, good should pass
	store.Validators = append(store.Validators, NewContentAddressValidator(hashfunc))
	chunks = GenerateRandomChunks(DefaultChunkSize, 2)
	goodChunk = chunks[0]
	badChunk = chunks[1]
	copy(badChunk.Data(), goodChunk.Data())

	err = mputChunks(store, goodChunk, badChunk)
	if err != nil {
		t.Fatalf("expected no error on good content address chunk with content address validator only, but got: %s", err)
	}
	err = mputChunks(store, badChunk)
	if err == nil {
		t.Fatal("expected error on bad content address chunk with content address validator only, but got nil")
	}

	// append resource validator to validators and check puts
	// bad should fail, good should pass, resource should pass
	rhParams := &ResourceHandlerParams{}
	rh, err := NewResourceHandler(rhParams)
	if err != nil {
		t.Fatal(err)
	}
	store.Validators = append(store.Validators, rh)

	goodChunk = GenerateRandomChunk(DefaultChunkSize)
	key := rh.resourceHash(42, 1, ens.EnsNode("xyzzy.eth"))
	data := []byte("bar")
	uglyChunk := newUpdateChunk(key, nil, 42, 1, "xyzzy.eth", data, len(data))

	err = mputChunks(store, goodChunk)
	if err != nil {
		t.Fatalf("expected no error on good content address chunk with both validators, but got: %s", err)
	}
	err = mputChunks(store, badChunk)
	if err != nil {
		t.Fatal("expected error on bad chunk address with both validators, but got nil")
	}
	err = mputChunks(store, uglyChunk)
	if err != nil {
		t.Fatalf("expected no error on resource update chunk with both validators, but got: %s", err)
	}

	// (redundant check)
	// use only resource validator, and check puts
	// bad should fail, good should fail, resource should pass
	store.Validators[0] = store.Validators[1]
	store.Validators = store.Validators[:1]

	goodChunk = GenerateRandomChunk(DefaultChunkSize)
	key = rh.resourceHash(42, 2, ens.EnsNode("xyzzy.eth"))
	data = []byte("baz")
	uglyChunk = newUpdateChunk(key, nil, 42, 2, "xyzzy.eth", data, len(data))

	err = mputChunks(store, goodChunk)
	if err != nil {
		t.Fatal("expected error on good content address chunk with resource validator only, but got nil")
	}
	err = mputChunks(store, badChunk)
	if err != nil {
		t.Fatal("expected error on bad content address chunk with resource validator only, but got nil")
	}
	err = mputChunks(store, uglyChunk)
	if err != nil {
		t.Fatalf("expected no error on resource update chunk with resource validator only, but got: %s", err)
	}
}
