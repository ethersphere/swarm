package snapshot

import (
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

	err = sim.StartAll()
	if err != nil {
		t.Error(err)
	}

	err = sim.StopAll()
	if err != nil {
		t.Error(err)
	}

}
