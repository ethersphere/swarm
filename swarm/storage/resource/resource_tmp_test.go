package resource

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var (
	hashfunc = storage.MakeHashFunc(storage.DefaultHash)
)

// tests that the content address validator correctly checks the data
// tests that resource update chunks are passed through content address validator
// the test checking the resouce update validator internal correctness is found in resource_test.go
func TestValidator(t *testing.T) {

	// set up localstore
	datadir, err := ioutil.TempDir("", "storage-testresourcevalidator")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(datadir)

	params := storage.NewDefaultLocalStoreParams()
	params.Init(datadir)
	store, err := storage.NewLocalStore(params, nil)
	if err != nil {
		t.Fatal(err)
	}

	// add content address validator and resource validator to validators and check puts
	// bad should fail, good should pass
	store.Validators = append(store.Validators, storage.NewContentAddressValidator(hashfunc))
	rhParams := &ResourceHandlerParams{}
	rh, err := NewResourceHandler(rhParams)
	if err != nil {
		t.Fatal(err)
	}
	store.Validators = append(store.Validators, rh)

	chunks := storage.GenerateRandomChunks(storage.DefaultChunkSize, 2)
	goodChunk := chunks[0]
	badChunk := chunks[1]
	badChunk.SData = goodChunk.SData
	key := rh.resourceHash(42, 1, ens.EnsNode("xyzzy.eth"))
	data := []byte("bar")
	uglyChunk := newUpdateChunk(key, nil, 42, 1, "xyzzy.eth", data, len(data))

	storage.PutChunks(store, goodChunk, badChunk, uglyChunk)
	if err := goodChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on good content address chunk with both validators, but got: %s", err)
	}
	if err := badChunk.GetErrored(); err == nil {
		t.Fatal("expected error on bad chunk address with both validators, but got nil")
	}
	if err := uglyChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on resource update chunk with both validators, but got: %s", err)
	}

	// (redundant check)
	// use only resource validator, and check puts
	// bad should fail, good should fail, resource should pass
	store.Validators[0] = store.Validators[1]
	store.Validators = store.Validators[:1]

	chunks = storage.GenerateRandomChunks(storage.DefaultChunkSize, 2)
	goodChunk = chunks[0]
	badChunk = chunks[1]
	badChunk.SData = goodChunk.SData

	key = rh.resourceHash(42, 2, ens.EnsNode("xyzzy.eth"))
	data = []byte("baz")
	uglyChunk = newUpdateChunk(key, nil, 42, 2, "xyzzy.eth", data, len(data))

	storage.PutChunks(store, goodChunk, badChunk, uglyChunk)
	if goodChunk.GetErrored() == nil {
		t.Fatal("expected error on good content address chunk with resource validator only, but got nil")
	}
	if badChunk.GetErrored() == nil {
		t.Fatal("expected error on bad content address chunk with resource validator only, but got nil")
	}
	if err := uglyChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on resource update chunk with resource validator only, but got: %s", err)
	}
}
