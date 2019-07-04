package simulation

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestExecAdapter(t *testing.T) {

	tmpdir, err := ioutil.TempDir("", "test-adapter-exec")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	adapter, err := NewExecAdapter(ExecAdapterConfig{
		Directory: tmpdir,
	})
	if err != nil {
		t.Fatalf("could not create exec adapter: %v", err)
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("could not generate key: %v", err)
	}
	hexKey := hex.EncodeToString(crypto.FromECDSA(key))

	args := []string{
		"--bootnodes", "",
		"--bzzkeyhex", hexKey,
		"--bzznetworkid", "49",
	}
	nodeconfig := NodeConfig{
		ID:     "node1",
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	node, err := adapter.NewNode(nodeconfig)
	if err != nil {
		t.Fatal(err)
	}
	status := node.Status()
	if status.ID != "node1" {
		t.Fatal("node id is different")
	}

	_, err = adapter.NewNode(nodeconfig)
	if err == nil {
		t.Fatal("a node with the same id was registered")
	}

	err = node.Start()
	if err != nil {
		t.Fatalf("node did not start: %v", err)
	}

	err = node.Stop()
	if err != nil {
		t.Fatalf("node didn't stop: %v", err)
	}

	err = node.Start()
	if err != nil {
		t.Fatalf("node didn't start again: %v", err)
	}

	err = node.Stop()
	if err != nil {
		t.Fatalf("node didn't stop: %v", err)
	}

}
