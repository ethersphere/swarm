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
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	_, err = rand.Read(baseKey)
	if err != nil {
		t.Fatal(err)
	}

	//Mock the store
	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var testTargets = [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}
	msg := []byte("RECOVERY")
	//call Send to store trojanChunk in store
	var ch chunk.Chunk
	ch, err = Send(ctx, localStore, testTargets, "RECOVERY", msg)
	if err != nil {
		t.Fatal(err)
	}

	//verify store, that trojan chunk has been stored correctly
	var chStored chunk.Chunk
	chStored, err = localStore.Get(ctx, chunk.ModeGetRequest, ch.Address())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ch, chStored) {
		t.Fatalf("Trojan chunk not stored properly")
	}

}

//TODO: later test could be a sim test for 2 nodes, localstore + netstore
