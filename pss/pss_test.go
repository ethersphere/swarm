package pss

import (
	"crypto/rand"
	"testing"

	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/storage/localstore"
)

func TestTrojanChunkRetrieval(t *testing.T) {
	baseKey := make([]byte, 32)
	_, err = rand.Read(baseKey)
	if err != nil {
		t.Fatal(err)
	}

	// using the same key in for underlay address as well as it is not important for test
	baseAddress := network.NewBzzAddr(baseKey, baseKey)
	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var testTargets = [][]byte{
		[]byte{57, 120},
		[]byte{209, 156},
		[]byte{156, 38},
		[]byte{89, 19},
		[]byte{22, 129}}

	Send(localStore, testTargets, "RECOVERY", "RECOVERY")
	// //netStore := storage.NewNetStore(localStore, baseAddress)
	// //localStore.Get()

	// //Mock the store
	// //call send
	// //verify store, that trojan chunk has been stored correctly
	// slice := make([]byte, n)
	// if randomData {
	// 	rand.Seed(time.Now().UnixNano())
	// 	rand.Read(slice)
	// }
	// ctx := context.Background()
	// dataPut := string(slice)
	// tag, err := swarm.api.Tags.Create("test-local-store-and-retrieve", 0, false)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// ctx = sctx.SetTag(ctx, tag.Uid)
	// k, wait, err := swarm.api.Store(ctx, strings.NewReader(dataPut), int64(len(dataPut)), false)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if wait != nil {
	// 	err = wait(ctx)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }

	// r, _ := swarm.api.Retrieve(context.TODO(), k)

	// d, err := ioutil.ReadAll(r)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// dataGet := string(d)

	// if len(dataPut) != len(dataGet) {
	// 	t.Fatalf("data not matched: length expected %v, got %v", len(dataPut), len(dataGet))
	// } else {
	// 	if dataPut != dataGet {
	// 		t.Fatal("data not matched")
	// 	}
	// }
}

//later test could be a sim test for 2 nodes
