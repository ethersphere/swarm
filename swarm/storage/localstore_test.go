package storage

import (
	"io/ioutil"
	"os"
	"testing"
)

var (
	hashfunc = MakeHashFunc(DefaultHash)
)

// tests that the content address validator correctly checks the data
// tests that resource update chunks are passed through content address validator
// the test checking the resouce update validator internal correctness is found in resource_test.go
func TestValidator(t *testing.T) {
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
	copy(badChunk.SData, goodChunk.SData)

	PutChunks(store, goodChunk, badChunk)
	if err := goodChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on good content address chunk in spite of no validation, but got: %s", err)
	}
	if err := badChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on bad content address chunk in spite of no validation, but got: %s", err)
	}

	// add content address validator and check puts
	// bad should fail, good should pass
	store.Validators = append(store.Validators, NewContentAddressValidator(hashfunc))
	chunks = GenerateRandomChunks(DefaultChunkSize, 2)
	goodChunk = chunks[0]
	badChunk = chunks[1]
	copy(badChunk.SData, goodChunk.SData)

	PutChunks(store, goodChunk, badChunk)
	if err := goodChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on good content address chunk with content address validator only, but got: %s", err)
	}
	if err := badChunk.GetErrored(); err == nil {
		t.Fatal("expected error on bad content address chunk with content address validator only, but got nil")
	}

	// append a validator that always denies
	// bad should fail, good should pass,
	var negV boolTestValidator
	store.Validators = append(store.Validators, negV)

	chunks = GenerateRandomChunks(DefaultChunkSize, 2)
	goodChunk = chunks[0]
	badChunk = chunks[1]
	copy(badChunk.SData, goodChunk.SData)

	PutChunks(store, goodChunk, badChunk)
	if err := goodChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on good content address chunk with content address validator only, but got: %s", err)
	}
	if err := badChunk.GetErrored(); err == nil {
		t.Fatal("expected error on bad content address chunk with content address validator only, but got nil")
	}

	// append a validator that always approves
	// all shall pass
	var posV boolTestValidator = true
	store.Validators = append(store.Validators, posV)

	chunks = GenerateRandomChunks(DefaultChunkSize, 2)
	goodChunk = chunks[0]
	badChunk = chunks[1]
	copy(badChunk.SData, goodChunk.SData)

	PutChunks(store, goodChunk, badChunk)
	if err := goodChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on good content address chunk with content address validator only, but got: %s", err)
	}
	if err := badChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on bad content address chunk with content address validator only, but got: %s", err)
	}
}

type boolTestValidator bool

func (self boolTestValidator) Validate(addr Address, data []byte) bool {
	return bool(self)
}
