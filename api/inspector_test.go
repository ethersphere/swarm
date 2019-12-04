package api

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/stream"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/state"
)

// TestInspectorPeerStreams validates that response from RPC peerStream has at
// least some data.
func TestInspectorPeerStreams(t *testing.T) {
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

	// using the same key in for underlay address as well as it is not important for test
	baseAddress := network.NewBzzAddr(baseKey, baseKey)
	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}
	netStore := storage.NewNetStore(localStore, baseAddress)

	i := NewInspector(nil, nil, netStore, stream.New(state.NewInmemoryStore(), baseAddress, stream.NewSyncProvider(netStore, network.NewKademlia(
		baseKey,
		network.NewKadParams(),
	), baseAddress, false, false)), localStore)

	server := rpc.NewServer()
	if err := server.RegisterName("inspector", i); err != nil {
		t.Fatal(err)
	}

	client := rpc.DialInProc(server)

	var peerInfo string

	err = client.Call(&peerInfo, "inspector_peerStreams")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(peerInfo, `"base":"`+baseAddress.ShortUnder()) {
		t.Error("missing base key in response")
	}
}

// TestInspectorStorageIndices validates that response from RPC storageIndices functions correctly
func TestInspectorStorageIndices(t *testing.T) {
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

	// using the same key in for underlay address as well as it is not important for test
	baseAddress := network.NewBzzAddr(baseKey, baseKey)
	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}
	netStore := storage.NewNetStore(localStore, baseAddress)

	i := NewInspector(nil, nil, netStore, stream.New(state.NewInmemoryStore(), network.NewBzzAddr(baseKey, baseKey), stream.NewSyncProvider(netStore, network.NewKademlia(
		baseKey,
		network.NewKadParams(),
	), baseAddress, false, false)), localStore)

	server := rpc.NewServer()
	if err := server.RegisterName("inspector", i); err != nil {
		t.Fatal(err)
	}

	client := rpc.DialInProc(server)

	var indiceInfo map[string]int

	err = client.Call(&indiceInfo, "inspector_storageIndices")
	if err != nil {
		t.Fatal(err)
	}
	if indiceInfo["gcSize"] != 0 {
		t.Fatalf("expected gcSize to be %d but got %d", 0, indiceInfo["gcSize"])
	}
}
