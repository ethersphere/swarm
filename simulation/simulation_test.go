package simulation

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm"
)

func TestAdapters(t *testing.T) {

	nodeCount := 10

	// Test exec adapter
	t.Run("exec", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "test-sim-exec")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpdir)
		adapter, err := NewExecAdapter(ExecAdapterConfig{
			// TODO: fix this
			ExecutablePath:    "/home/rafael/go/bin/swarm",
			BaseDataDirectory: tmpdir,
		})
		if err != nil {
			t.Fatalf("could not create exec adapter: %v", err)
		}
		startSimulation(t, adapter, nodeCount)
	})

	// Test docker adapter
	t.Run("docker", func(t *testing.T) {
		config := DefaultDockerAdapterConfig()
		config.DockerImage = "sim-docker-test:latest"
		/*config.BuildContext.Dockerfile = "Dockerfile"
		config.BuildContext.Directory = "../"
		config.BuildContext.Tag = "sim-docker-test:latest"*/

		adapter, err := NewDockerAdapter(config)
		if err != nil {
			t.Fatalf("could not create docker adapter: %v", err)
		}
		startSimulation(t, adapter, nodeCount)
	})

	// Test kubernetes adapter
	t.Run("kubernetes", func(t *testing.T) {
		config := DefaultKubernetesAdapterConfig()
		config.Namespace = "simulation-test"
		config.DockerImage = "skylenet/swarm-test:why"
		/*config.BuildContext.Dockerfile = "Dockerfile"
		config.BuildContext.Directory = "../"
		config.BuildContext.Tag = "swarm-test:why"
		config.BuildContext.Registry = "skylenet"
		config.BuildContext.Username = "skylenet"
		config.BuildContext.Password = "xxxxxxxx"*/

		adapter, err := NewKubernetesAdapter(config)
		if err != nil {
			t.Fatalf("could not create kubernetes adapter: %v", err)
		}
		startSimulation(t, adapter, nodeCount)
	})

}

func startSimulation(t *testing.T, adapter Adapter, count int) {
	nodeIDs := make([]NodeID, count)
	sim := NewSimulation(adapter)

	// Create nodes
	for i := 0; i < count; i++ {
		nodeIDs[i] = NodeID(fmt.Sprintf("node%d", i))
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

		cfg := NodeConfig{
			ID:     nodeIDs[i],
			Args:   args,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := sim.Create(cfg); err != nil {
			t.Fatalf("failed to create node %s: %v", cfg.ID, err)
		}

	}

	// Start nodes
	for _, id := range nodeIDs {
		err := sim.Start(id)
		if err != nil {
			t.Errorf("failed to start node %s: %v", id, err)
		}
	}

	defer func() {
		// Stop nodes
		for _, id := range nodeIDs {
			err := sim.Stop(id)
			if err != nil {
				t.Errorf("failed to stop node %s: %v", id, err)
			}
		}
	}()

	// Test some RPC calls
	nodes := make([]Node, count)
	rpcClients := make([]*rpc.Client, count)

	for idx, id := range nodeIDs {

		node, err := sim.Get(id)
		nodes[idx] = node
		if err != nil {
			t.Fatalf("could not get node %s: %v", id, err)
		}
		client, err := sim.RPCClient(id)
		if err != nil {
			t.Fatalf("failed to get an rpc client for node %s: %v", id, err)
			continue
		}
		defer client.Close()
		rpcClients[idx] = client

		var swarminfo swarm.Info
		err = client.Call(&swarminfo, "bzz_info")
		if err != nil {
			t.Errorf("failed getting bzz info: %v", err)
		}
	}

	// Add all nodes to each other
	for i := 0; i < count; i++ {
		for j := 0; j < count; j++ {
			if i == j {
				continue
			}
			err := rpcClients[i].Call(nil, "admin_addPeer", nodes[j].Status().Enode)
			if err != nil {
				t.Errorf("could not add peer %s: %v", nodes[j].Status().ID, err)
			}
		}

	}

	var hive string
	err := rpcClients[0].Call(&hive, "bzz_hive")
	if err != nil {
		t.Errorf("could not get hive info: %v", err)
	}

	fmt.Println(hive)
}
