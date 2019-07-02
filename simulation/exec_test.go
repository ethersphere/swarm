package simulation

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestExecAdapter(t *testing.T) {

	tmpdir, err := ioutil.TempDir("", "test-exec")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	adapter, err := NewExecAdapter(ExecAdapterConfig{
		Directory: tmpdir,
	})
	if err != nil {
		t.Fatal(err)
	}

	nodeconfig := NodeConfig{
		ID: "node1",
	}
	node, err := adapter.NewNode(nodeconfig)
	if err != nil {
		t.Fatal(err)
	}
	status := node.Status()
	if status.ID != "node1" {
		t.Error("node id is different")
	}

	_, err = adapter.NewNode(nodeconfig)
	if err == nil {
		t.Error("a node with the same id was registered")
	}

}
