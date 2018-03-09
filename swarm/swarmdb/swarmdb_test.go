package swarmdb

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/swarmdb"
)

const serviceName = "swarmdb"
const testMinProxBinSize = 1

var services = adapters.Services{
	serviceName: newService,
}

var (
	nodeCount    = flag.Int("nodes", 1, "number of nodes to create (default 10)")
	initCount    = flag.Int("conns", 1, "number of originally connected peers	 (default 1)")
	snapshotFile = flag.String("snapshot", "", "create snapshot")
	loglevel     = flag.Int("loglevel", 3, "verbosity of logs")
)

func init() {
	flag.Parse()
	// register the discovery service which will run as a devp2p
	// protocol when using the exec adapter
	adapters.RegisterServices(services)

	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

func TestSwarmDB(t *testing.T) {

}

func TestDiscoverySimulationSimAdapter(t *testing.T) {
	testDiscoverySimulationSimAdapter(t, *nodeCount, *initCount)
}

func testDiscoverySimulationSimAdapter(t *testing.T, nodes, conns int) {
	testDiscoverySimulation(t, nodes, adapters.NewSimAdapter(services))
}

func testDiscoverySimulation(t *testing.T, nodes int, adapter adapters.NodeAdapter) {
	startedAt := time.Now()
	result, err := discoverySimulation(nodes, adapter)
	if err != nil {
		t.Fatalf("Setting up simulation failed: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("Simulation failed: %s", result.Error)
	}
	t.Logf("Simulation with %d nodes passed in %s", nodes, result.FinishedAt.Sub(result.StartedAt))
	finishedAt := time.Now()
	t.Logf("Setup: %s, shutdown: %s", result.StartedAt.Sub(startedAt), finishedAt.Sub(result.FinishedAt))
}

func discoverySimulation(nodes int, adapter adapters.NodeAdapter) (*simulations.StepResult, error) {
	// create network
	net := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: serviceName,
	})
	defer net.Shutdown()
	trigger := make(chan discover.NodeID) // Create Trigger Channel, Action and a function
	ids := make([]discover.NodeID, nodes)
	for i := 0; i < nodes; i++ {
		conf := adapters.RandomNodeConfig()
		node, err := net.NewNodeWithConfig(conf)
		if err != nil {
			return nil, fmt.Errorf("error starting node: %s", err)
		}
		if err := net.Start(node.ID()); err != nil {
			return nil, fmt.Errorf("error starting node %s: %s", node.ID().TerminalString(), err)
		}
		ids[i] = node.ID()
	}

	// run a simulation which connects the 10 nodes in a ring and waits
	// for full peer discovery
	var addrs [][]byte
	id := ids[0]
	action := func(ctx context.Context) error {
		go func() {
			trigger <- id
		}()
		return nil
	}
	log.Debug(fmt.Sprintf("nodes: %v", len(addrs)))

	check := func(ctx context.Context, id discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		node := net.GetNode(id)
		if node == nil {
			return false, fmt.Errorf("unknown node: %s", id)
		}
		client, err := node.Client()
		if err != nil {
			return false, fmt.Errorf("error getting node client: %s", err)
		}
		//cr_db := swarmdb.CreateDatabse()
		config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
		swarmdb.NewKeyManager(config)
		u := config.GetSWARMDBUser()

		//if err := client.Call(nil /* first return value */, "swarmdb_createDatabase", u *SWARMDBUser, owner string, database string, encrypted int); err != nil {
		if err := client.Call(nil, "swarmdb_createDatabase", u, "owner.eth", "database", 1); err != nil {
			return false, fmt.Errorf("Error Creating SwarmDB Database: %s", err)
		}
		return true, nil
	}

	// 64 nodes ~ 1min
	// 128 nodes ~
	timeout := 300 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result := simulations.NewSimulation(net).Run(ctx, &simulations.Step{
		Action:  action,
		Trigger: trigger,
		Expect: &simulations.Expectation{
			Nodes: ids,
			Check: check,
		},
	})
	if result.Error != nil {
		return result, nil
	}
	return result, nil
}

func newService(ctx *adapters.ServiceContext) (node.Service, error) {
	config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	swdb, err := swarmdb.NewSwarmDB(config)
	return swdb, err

}
