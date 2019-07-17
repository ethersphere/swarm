package snapshot

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/simulation"
)

func TestExecSnapshotFromFile(t *testing.T) {
	snap, err := simulation.LoadSnapshotFromFile("exec.json")
	if err != nil {
		t.Fatal(err)
	}

	dir := snap.Adapter.Config.(simulation.ExecAdapterConfig).BaseDataDirectory
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	sim, err := simulation.NewSimulationFromSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}

	nodes := sim.GetAll()
	if len(nodes) != len(snap.Nodes) {
		t.Fatalf("Got %d . Expected %d nodes", len(nodes), len(snap.Nodes))
	}

	err = sim.StartAll()
	if err != nil {
		t.Fatal(err)
	}
	err = sim.StopAll()
	if err != nil {
		t.Fatal(err)
	}

}
func TestExecSnapshot(t *testing.T) {
	execPath := "../../../build/bin/swarm"

	if _, err := os.Stat(execPath); err != nil {
		if os.IsNotExist(err) {
			t.Skip("swarm binary not found. build it before running the test")
		}
	}

	tmpdir, err := ioutil.TempDir("", "test-sim-exec")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	adapter, err := simulation.NewExecAdapter(simulation.ExecAdapterConfig{
		ExecutablePath:    execPath,
		BaseDataDirectory: tmpdir,
	})
	if err != nil {
		t.Fatalf("could not create exec adapter: %v", err)
	}

	sim := simulation.NewSimulation(adapter)

	count := 5
	// Create some nodes
	for i := 0; i < count; i++ {
		id := simulation.NodeID(fmt.Sprintf("snap-node%d", i))
		// Generate keys
		bzzkey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("could not generate key: %v", err)
		}
		bzzkeyhex := hex.EncodeToString(crypto.FromECDSA(bzzkey))

		nodekey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("could not generate key: %v", err)
		}
		nodekeyhex := hex.EncodeToString(crypto.FromECDSA(nodekey))

		// Set CLI args
		args := []string{
			"--bootnodes", "",
			"--bzzkeyhex", bzzkeyhex,
			"--nodekeyhex", nodekeyhex,
			"--bzznetworkid", "499",
		}

		cfg := simulation.NodeConfig{
			ID:     id,
			Args:   args,
			Stdout: ioutil.Discard,
			Stderr: ioutil.Discard,
		}

		if err := sim.Init(cfg); err != nil {
			t.Fatalf("failed to create node %s: %v", cfg.ID, err)
		}
	}

	snap, err := sim.Snapshot()
	if err != nil {
		t.Fatalf("could not create snapshot: %v", err)
	}

	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("could not marshal json: %v", err)
	}
	fmt.Println(string(b))

	// Create new simulation from snapshot
	sim2, err := simulation.NewSimulationFromSnapshot(snap)

	if !reflect.DeepEqual(sim, sim2) {
		t.Fatal("simulations are not equal")
	}

}
