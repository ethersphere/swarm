package api

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/newstream"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"

	"github.com/ethereum/go-ethereum/p2p/enode"
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

	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}
	netStore := storage.NewNetStore(localStore, baseKey, enode.ID{})

	i := NewInspector(nil, nil, netStore, newstream.New(state.NewInmemoryStore(), baseKey, newstream.NewSyncProvider(netStore, network.NewKademlia(
		baseKey,
		network.NewKadParams(),
	), false, false)))

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

	// if want := hex.EncodeToString(baseKey)[:16]; peerInfo.Base != want {
	// 	t.Fatalf("got base key %q, want %q", peerInfo.Base, want)
	// }

	if !strings.Contains(peerInfo, `"base":"`+hex.EncodeToString(baseKey)[:16]+`"`) {
		t.Error("missing base key in response")
	}

	t.Log(peerInfo)
}
