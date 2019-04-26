package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

// echo '{"jsonrpc":"2.0","method":"admin_nodeInfo","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/<<namespace>>/pods/http:<<deploymentName>>-<<index>>:8546/proxy/ --origin localhost
// echo '{"jsonrpc":"2.0","method":"admin_peers","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/gluk256/pods/http:swarm-3:8546/proxy/ --origin localhost | jq ".[]" | tail -n+3 | jq ".[] | .enode"
// echo '{"jsonrpc":"2.0","method":"admin_peers","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/<<namespace>>/pods/http:<<deploymentName>>-<<index>>:8546/proxy/ --origin localhost
// echo '{"jsonrpc":"2.0","method":"admin_peers","id":1}' | websocat ws://localhost:8001/api/v1/namespaces/gluk256/pods/http:swarm-3:8546/proxy/ --origin localhost | jq ".[]" | tail -n+3 | jq ".[] | .id"

var (
	nodes          int
	namespace      string
	deploymentName string
)

func init() {
	flag.IntVar(&nodes, "nodes", 0, "number of nodes in the deployment")
	flag.StringVar(&namespace, "namespace", "staging", "kubernetes namespace of the deployment")
	flag.StringVar(&deploymentName, "deploymentName", "swarm-private", "deployment name")
}

func getClient(wsHost string) *rpc.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rpcClient, err := rpc.DialContext(ctx, wsHost)
	if err != nil {
		panic(err)
	}

	return rpcClient
}

// getBzzAddrFromHost returns the bzzAddr for a given host
func getBzzAddrFromHost(client *rpc.Client) (string, error) {
	var hive string

	err := client.Call(&hive, "bzz_hive")
	if err != nil {
		return "", err
	}

	// we make an ugly assumption about the output format of the hive.String() method
	// ideally we should replace this with an API call that returns the bzz addr for a given host,
	// but this also works for now (provided we don't change the hive.String() method, which we haven't in some time
	return strings.Split(strings.Split(hive, "\n")[3], " ")[10], nil
}

func getNodeInfoId(client *rpc.Client) (string, error) {
	var nodeInfo p2p.NodeInfo

	err := client.Call(&nodeInfo, "admin_nodeInfo")
	if err != nil {
		return "", err
	}

	// we make an ugly assumption about the output format of the hive.String() method
	// ideally we should replace this with an API call that returns the bzz addr for a given host,
	// but this also works for now (provided we don't change the hive.String() method, which we haven't in some time
	return nodeInfo.ID, nil
}

func main() {
	flag.Parse()

	fmt.Println("host                    ;       bzz addr                                                                ;       id")
	var wg sync.WaitGroup
	wg.Add(nodes)
	for i := 0; i < nodes; i++ {
		i := i
		go func() {
			cl := getClient(fmt.Sprintf("ws://localhost:8001/api/v1/namespaces/%s/pods/http:%s-%d:8546/proxy/", namespace, deploymentName, i))
			defer cl.Close()
			defer wg.Done()

			bzzAddr, err := getBzzAddrFromHost(cl)
			if err != nil {
				panic(err)
			}

			id, err := getNodeInfoId(cl)
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s-%2d\t;\t%s\t;\t%s\n", deploymentName, i, bzzAddr, id)
		}()
	}

	wg.Wait()
}
