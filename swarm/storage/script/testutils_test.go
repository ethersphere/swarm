package script_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/storage/script"
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
func NewTestHandler(t *testing.T) (handler script.Handler, cleanup func()) {
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

	handler = script.NewHandler(&script.HandlerParams{
		ChunkStore: netStore,
	})
	localStore.Validators = append(localStore.Validators, handler)

	return handler, func() {
		//	netStore.Close()
		//	localStore.Close()
		os.RemoveAll(path)
	}
}

func JSONEquals(t *testing.T, expected, actual string) {
	//credit for the trick: turtlemonvh https://gist.github.com/turtlemonvh/e4f7404e28387fadb8ad275a99596f67
	var e interface{}
	var a interface{}

	err := json.Unmarshal([]byte(expected), &e)
	if err != nil {
		t.Fatalf("Error mashalling expected :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(actual), &a)
	if err != nil {
		t.Fatalf("Error mashalling actual :: %s", err.Error())
	}

	if !reflect.DeepEqual(e, a) {
		t.Fatalf("Error comparing JSON. Expected %s. Got %s", expected, actual)
	}
}
