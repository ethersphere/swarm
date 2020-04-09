package pss

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage/localstore"
)

func TestTrojanChunkRetrieval(t *testing.T) {
	ctx := context.TODO()

	localStore, err := newMockLocalStore()
	if err != nil {
		t.Fatal(err)
	}

	testTargets := [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}
	payload := []byte("RECOVERY CHUNK")
	//call Send to store trojanChunk in store
	var ch chunk.Chunk
	if ch, err = Send(ctx, localStore, testTargets, "RECOVERY", payload); err != nil {
		t.Fatal(err)
	}

	//verify store, that trojan chunk has been stored correctly
	var storedChunk chunk.Chunk
	if storedChunk, err = localStore.Get(ctx, chunk.ModeGetRequest, ch.Address()); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ch, storedChunk) {
		t.Fatalf("Trojan chunk not stored properly")
	}

}

func newMockLocalStore() (*localstore.DB, error) {
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	if _, err = rand.Read(baseKey); err != nil {
		return nil, err
	}

	//Mock the store
	return localstore.New(dir, baseKey, &localstore.Options{})
}

//TODO: later test could be a simulation test for 2 nodes, localstore + netstore
