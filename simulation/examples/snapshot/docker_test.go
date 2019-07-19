package snapshot

import (
	"fmt"
	"testing"

	"github.com/ethersphere/swarm/simulation"
)

func TestDockerSnapshotFromFile(t *testing.T) {
	snap, err := simulation.LoadSnapshotFromFile("docker.json")
	if err != nil {
		t.Fatal(err)
	}

	sim, err := simulation.NewSimulationFromSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}

	nodes := sim.GetAll()
	if len(nodes) != len(snap.Nodes) {
		t.Fatalf("Got %d . Expected %d nodes", len(nodes), len(snap.Nodes))
	}

	// Check hive output on the first node
	node, err := sim.Get(simulation.NodeID("test-0"))
	if err != nil {
		t.Error(err)
	}

	client, err := sim.RPCClient(node.Info().ID)
	if err != nil {
		t.Errorf("Failed to get rpc client: %v", err)
	}

	var hive string
	err = client.Call(&hive, "bzz_hive")
	if err != nil {
		t.Errorf("could not get hive info: %v", err)
	}

	fmt.Println(hive)

	err = sim.StopAll()
	if err != nil {
		t.Error(err)
	}

}
