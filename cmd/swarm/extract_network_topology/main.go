package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

// TODO: set correct node ids
// TODO: set correct node services
// echo '{"jsonrpc":"2.0","method":"admin_nodeInfo","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/<<namespace>>/pods/http:<<deploymentName>>-<<index>>:8546/proxy/ --origin localhost

// TODO: init correctly Snap.Conns
// echo '{"jsonrpc":"2.0","method":"admin_peers","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/gluk256/pods/http:swarm-3:8546/proxy/ --origin localhost | jq ".[]" | tail -n+3 | jq ".[] | .enode"
// echo '{"jsonrpc":"2.0","method":"admin_peers","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/<<namespace>>/pods/http:<<deploymentName>>-<<index>>:8546/proxy/ --origin localhost
// echo '{"jsonrpc":"2.0","method":"admin_peers","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/gluk256/pods/http:swarm-3:8546/proxy/ --origin localhost | jq ".[]" | tail -n+3 | jq ".[] | .id"

var (
	nodes          int
	namespace      string
	deploymentName string
)

func init() {
	flag.IntVar(&nodes, "nodes", 3, "number of nodes in the deployment")
	flag.StringVar(&namespace, "namespace", "staging", "kubernetes namespace of the deployment")
	flag.StringVar(&deploymentName, "deploymentName", "swarm-private", "deployment name")
}

func main() {
	flag.Parse()

	privateKeys := []string{}
	names := []string{}

	// get private keys
	for i := 0; i < nodes; i++ {
		cmd := exec.Command("kubectl", "exec", "-n", namespace, "-ti", fmt.Sprintf("%s-%d", deploymentName, i), "--", "cat", "/root/.ethereum/swarm/nodekey")
		res, err := cmd.Output()
		if err != nil {
			panic(err)
		}

		names = append(names, fmt.Sprintf("%s-%d", deploymentName, i))
		privateKeys = append(privateKeys, string(res))
	}

	snap := simulations.Snapshot{}
	// generate snapshot
	for i := 0; i < nodes; i++ {
		prvkey, err := crypto.HexToECDSA(privateKeys[i])
		if err != nil {
			panic(err)
		}

		n := simulations.Node{
			Config: &adapters.NodeConfig{
				Name:            names[i],
				PrivateKey:      prvkey,
				EnableMsgEvents: true,
			},
			Up: true,
		}

		snap.Nodes = append(snap.Nodes, simulations.NodeSnapshot{Node: n})
	}

	js, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(js))
}
