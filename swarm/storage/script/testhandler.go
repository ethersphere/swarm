package script

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage"
)

// NewTestHandler creates a mock Handler object to be used for testing purposes.
func NewTestHandler(t *testing.T) (handler Handler, cleanup func()) {
	path, err := ioutil.TempDir("", "bzzscript-test")
	if err != nil {
		t.Fatal(err)
	}

	localstoreparams := storage.NewDefaultLocalStoreParams()
	localstoreparams.Init(path)
	localStore, err := storage.NewLocalStore(localstoreparams, nil)
	if err != nil {
		t.Fatalf("localstore create fail, path %s: %v", path, err)
	}

	handler = NewHandler(&HandlerParams{
		ChunkStore: localStore,
	})
	localStore.Validators = append(localStore.Validators, handler)

	return handler, func() {
		os.RemoveAll(path)
	}
}
