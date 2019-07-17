package goclient

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/simulation"
	"golang.org/x/sync/errgroup"
)

type GoClientSimulation struct {
	*simulation.Simulation
}

func NewGoClientSimulation(adapter simulation.Adapter) *GoClientSimulation {
	sim := simulation.NewSimulation(adapter)
	return &GoClientSimulation{sim}
}

func (s *GoClientSimulation) AddBootnode(id simulation.NodeID, args []string) (simulation.Node, error) {
	a := []string{
		"--bootnode-mode",
		"--bootnodes", "",
	}
	a = append(a, args...)
	return s.AddNode(id, a)
}

func (s *GoClientSimulation) AddNode(id simulation.NodeID, args []string) (simulation.Node, error) {
	bzzkey, err := randomHexKey()
	if err != nil {
		return nil, err
	}
	nodekey, err := randomHexKey()
	if err != nil {
		return nil, err
	}
	a := []string{
		"--bzzkeyhex", bzzkey,
		"--nodekeyhex", nodekey,
	}
	a = append(a, args...)
	cfg := simulation.NodeConfig{
		ID:     id,
		Args:   a,
		Stdout: ioutil.Discard,
		Stderr: ioutil.Discard,
	}
	err = s.Init(cfg)
	if err != nil {
		return nil, err
	}

	err = s.Start(id)
	if err != nil {
		return nil, err
	}
	node, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (s *GoClientSimulation) AddNodes(idPrefix string, count int, args []string) ([]simulation.Node, error) {
	g, _ := errgroup.WithContext(context.Background())

	for i := 0; i < count; i++ {
		id := simulation.NodeID(fmt.Sprintf("%s%d", idPrefix, i))
		g.Go(func() error {
			node, err := s.AddNode(id, args)
			if err != nil {
				log.Warn("Failed to add node", "id", id, "err", err.Error())
			} else {
				log.Info("Added node", "id", id, "enode", node.Info().Enode)
			}
			return err
		})
	}
	err := g.Wait()
	if err != nil {
		return nil, err
	}

	nodes := make([]simulation.Node, count)
	for i := 0; i < count; i++ {
		id := simulation.NodeID(fmt.Sprintf("%s%d", idPrefix, i))
		nodes[i], err = s.Get(id)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (s *GoClientSimulation) CreateClusterWithBootnode(idPrefix string, count int, args []string) ([]simulation.Node, error) {
	bootnode, err := s.AddBootnode(simulation.NodeID(fmt.Sprintf("%s-bootnode", idPrefix)), args)
	if err != nil {
		return nil, err
	}

	nodeArgs := []string{
		"--bootnodes", bootnode.Info().Enode,
	}
	nodeArgs = append(nodeArgs, args...)

	n, err := s.AddNodes(idPrefix, count, nodeArgs)
	if err != nil {
		return nil, err
	}
	nodes := []simulation.Node{bootnode}
	nodes = append(nodes, n...)
	return nodes, nil
}

func (s *GoClientSimulation) WaitForHealthyNetwork() error {
	nodes := s.GetAll()

	// Generate RPC clients
	var clients struct {
		RPC []*rpc.Client
		mu  sync.Mutex
	}
	clients.RPC = make([]*rpc.Client, len(nodes))

	g, _ := errgroup.WithContext(context.Background())

	for idx, node := range nodes {
		node := node
		idx := idx
		g.Go(func() error {
			id := node.Info().ID
			client, err := s.RPCClient(id)
			if err != nil {
				return err
			}
			clients.mu.Lock()
			clients.RPC[idx] = client
			clients.mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	for _, c := range clients.RPC {
		defer c.Close()
	}

	// Generate addresses for PotMap
	addrs := [][]byte{}
	for _, node := range nodes {
		byteaddr, err := hexutil.Decode(node.Info().BzzAddr)
		if err != nil {
			return err
		}
		addrs = append(addrs, byteaddr)
	}

	ppmap := network.NewPeerPotMap(network.NewKadParams().NeighbourhoodSize, addrs)

	log.Info("Waiting for healthy kademlia...")

	for i := 0; i < len(nodes); {
		healthy := &network.Health{}
		if err := clients.RPC[i].Call(&healthy, "hive_getHealthInfo", ppmap[nodes[i].Info().BzzAddr[2:]]); err != nil {
			return err
		}
		if healthy.Healthy() {
			i++
		} else {
			log.Info("Node isn't healthy yet, checking again all nodes...", "id", nodes[i].Info().ID)
			time.Sleep(500 * time.Millisecond)
			i = 0 // Start checking all nodes again
		}
	}
	log.Info("Healthy kademlia on all nodes")
	return nil
}

func randomHexKey() (string, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", err
	}
	keyhex := hex.EncodeToString(crypto.FromECDSA(key))
	return keyhex, nil
}
