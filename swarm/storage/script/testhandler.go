package script

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

type mockNetFetcher struct{}

func (m *mockNetFetcher) Request(ctx context.Context, hopCount uint8) {
}
func (m *mockNetFetcher) Offer(ctx context.Context, source *enode.ID) {
}

func newFakeNetFetcher(context.Context, storage.Address, *sync.Map) storage.NetFetcher {
	return &mockNetFetcher{}
}

// NewTestHandler creates Handler object to be used for testing purposes.
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
	netStore, err := storage.NewNetStore(localStore, nil)
	if err != nil {
		t.Fatal(err)
	}
	netStore.NewNetFetcherFunc = newFakeNetFetcher

	handler = NewHandler(&HandlerParams{
		ChunkStore: netStore,
	})
	localStore.Validators = append(localStore.Validators, handler)

	return handler, func() {
		//	netStore.Close()
		//	localStore.Close()
		os.RemoveAll(path)
	}
}
