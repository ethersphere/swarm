package discovery

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/simulation"
	colorable "github.com/mattn/go-colorable"
)

var (
	nodes    = flag.Int("nodes", 20, "number of nodes to create")
	loglevel = flag.Int("loglevel", 3, "verbosity of logs")
	rawlog   = flag.Bool("rawlog", false, "remove terminal formatting from logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(!*rawlog))))
}

func TestDiscovery(t *testing.T) {

	nodeCount := *nodes

	// Test exec adapter
	t.Run("exec", func(t *testing.T) {
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
		startSimulation(t, adapter, nodeCount)
	})

	// Test docker adapter
	t.Run("docker", func(t *testing.T) {
		config := simulation.DefaultDockerAdapterConfig()
		if !simulation.IsDockerAvailable(config.DaemonAddr) {
			t.Skip("docker is not available, skipping test")
		}
		config.DockerImage = "ethersphere/swarm:edge"
		adapter, err := simulation.NewDockerAdapter(config)
		if err != nil {
			t.Fatalf("could not create docker adapter: %v", err)
		}
		startSimulation(t, adapter, nodeCount)
	})

	// Test kubernetes adapter
	t.Run("kubernetes", func(t *testing.T) {
		config := simulation.DefaultKubernetesAdapterConfig()
		config.Namespace = "simulation-test"
		config.DockerImage = "ethersphere/swarm:edge"
		adapter, err := simulation.NewKubernetesAdapter(config)
		if err != nil {
			t.Fatalf("could not create kubernetes adapter: %v", err)
		}
		startSimulation(t, adapter, nodeCount)
	})
}

func startSimulation(t *testing.T, adapter simulation.Adapter, count int) {
	nodeIDs := make([]simulation.NodeID, count)
	sim := simulation.NewSimulation(adapter)

	// Create nodes
	for i := 0; i < count; i++ {
		nodeIDs[i] = simulation.NodeID(fmt.Sprintf("node%d", i))
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
			ID:     nodeIDs[i],
			Args:   args,
			Stdout: ioutil.Discard,
			Stderr: ioutil.Discard,
		}

		if err := sim.Init(cfg); err != nil {
			t.Fatalf("failed to create node %s: %v", cfg.ID, err)
		}
	}

	// Start nodes
	now := time.Now()

	log.Info("Starting nodes...", "count", count)

	err := sim.StartAll()

	log.Info("Started nodes", "time", fmt.Sprintf("%fs", time.Since(now).Seconds()))

	if err != nil {
		sim.StopAll()
		t.Fatalf("failed to start nodes: %v", err)
	}
	defer func() {
		err := sim.StopAll()
		if err != nil {
			t.Fatalf("could not stop all nodes: %v", err)
		}
	}()

	// Generate RPC clients

	var clients struct {
		RPC []*rpc.Client
		mu  sync.Mutex
	}
	clients.RPC = make([]*rpc.Client, count)

	nodes := sim.GetAll()
	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for idx, node := range nodes {
		go func(node simulation.Node, idx int) {
			defer wg.Done()
			id := node.Info().ID
			log.Info("getting rpc client", "node", id)
			client, err := sim.RPCClient(id)
			if err != nil {
				t.Errorf("failed to get an rpc client for node %s: %v", id, err)
			}
			clients.mu.Lock()
			clients.RPC[idx] = client
			clients.mu.Unlock()
		}(node, idx)
	}
	wg.Wait()

	log.Info("Adding peers...")
	for i := 0; i < count-1; i++ {
		go func(idx int) {
			err := clients.RPC[idx].Call(nil, "admin_addPeer", nodes[idx+1].Info().Enode)
			if err != nil {
				t.Errorf("could not add peer %s: %v", nodes[idx+1].Info().ID, err)
			}
		}(i)
	}

	// Wait for healthy kademlia on all peers
	addrs := [][]byte{}

	for _, node := range nodes {
		byteaddr, err := hexutil.Decode(node.Info().BzzAddr)
		if err != nil {
			t.Fatalf("failed to decode hex")
		}
		addrs = append(addrs, byteaddr)
	}

	ppmap := network.NewPeerPotMap(network.NewKadParams().NeighbourhoodSize, addrs)

	log.Info("Waiting for healthy kademlia...")

	for i := 0; i < count; {
		healthy := &network.Health{}
		if err := clients.RPC[i].Call(&healthy, "hive_getHealthInfo", ppmap[nodes[i].Info().BzzAddr[2:]]); err != nil {
			t.Errorf("failed to call hive_getHealthInfo")
		}
		if healthy.Healthy() {
			i++
		} else {
			log.Info("Node isn't healthy, checking again all nodes...", "node", nodes[i].Info().ID)
			time.Sleep(500 * time.Millisecond)
			i = 0 // Start checking all nodes again
		}
	}

	// Check hive output
	var hive string
	err = clients.RPC[0].Call(&hive, "bzz_hive")
	if err != nil {
		t.Errorf("could not get hive info: %v", err)
	}
	fmt.Println(hive)

	var knownPeers []string
	err = clients.RPC[0].Call(&knownPeers, "bzz_listKnown")
	if err != nil {
		t.Errorf("could not get known peers")
	}

	if len(knownPeers) != count-1 {
		t.Errorf("wrong known peer count. Should be %d, was %d", count-1, len(knownPeers))
	}
}
