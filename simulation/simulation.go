package simulation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/sync/errgroup"
)

type NodeMap struct {
	sync.RWMutex
	internal map[NodeID]Node
}

func NewNodeMap() *NodeMap {
	return &NodeMap{
		internal: make(map[NodeID]Node),
	}
}

func (nm *NodeMap) Load(key NodeID) (value Node, ok bool) {
	nm.RLock()
	result, ok := nm.internal[key]
	nm.RUnlock()
	return result, ok
}

func (nm *NodeMap) LoadAll() []Node {
	nm.RLock()
	v := []Node{}
	for _, node := range nm.internal {
		v = append(v, node)
	}
	nm.RUnlock()
	return v
}

func (nm *NodeMap) Store(key NodeID, value Node) {
	nm.Lock()
	nm.internal[key] = value
	nm.Unlock()
}

type Simulation struct {
	adapter Adapter
	nodes   *NodeMap
}

// NewSimulation creates a new simulation given an adapter
func NewSimulation(adapter Adapter) *Simulation {
	sim := &Simulation{
		adapter: adapter,
		nodes:   NewNodeMap(),
	}
	return sim
}

// NewSimulationFromSnapshot creates aimulation given an adapter and a snapshot
func NewSimulationFromSnapshot(adapter *Adapter, snapshot *NetworkSnapshot) (*Simulation, error) {
	return nil, errors.New("not implemented")
}

// Get returns a node by ID
func (s *Simulation) Get(id NodeID) (Node, error) {
	node, ok := s.nodes.Load(id)
	if !ok {
		return nil, fmt.Errorf("a node with id %s already exists", id)
	}
	return node, nil
}

// GetAll returns all nodes
func (s *Simulation) GetAll() []Node {
	return s.nodes.LoadAll()
}

// Init initializes a node with the NodeConfig
func (s *Simulation) Init(config NodeConfig) error {
	if _, ok := s.nodes.Load(config.ID); ok {
		return fmt.Errorf("a node with id %s already exists", config.ID)
	}

	node, err := s.adapter.NewNode(config)
	if err != nil {
		return fmt.Errorf("failed to create node: %v", err)
	}
	s.nodes.Store(config.ID, node)
	return nil
}

// Start starts a given node
func (s *Simulation) Start(id NodeID) error {
	node, ok := s.nodes.Load(id)
	if !ok {
		return fmt.Errorf("a node with id %s does not exists", id)
	}

	if err := node.Start(); err != nil {
		return fmt.Errorf("could not start node: %v", err)
	}
	return nil
}

// Stop stops a node by ID
func (s *Simulation) Stop(id NodeID) error {
	node, ok := s.nodes.Load(id)
	if !ok {
		return fmt.Errorf("a node with id %s does not exists", id)
	}

	if err := node.Stop(); err != nil {
		return fmt.Errorf("could not stop node: %v", err)
	}
	return nil
}

// StartAll starts all nodes
func (s *Simulation) StartAll() error {
	g, _ := errgroup.WithContext(context.Background())
	for _, node := range s.nodes.LoadAll() {
		g.Go(node.Start)
	}
	return g.Wait()
}

// StopAll stops all nodes
func (s *Simulation) StopAll() error {
	g, _ := errgroup.WithContext(context.Background())
	for _, node := range s.nodes.LoadAll() {
		g.Go(node.Stop)
	}
	return g.Wait()
}

// RPCClient returns an RPC Client for a given node
func (s *Simulation) RPCClient(id NodeID) (*rpc.Client, error) {
	node, ok := s.nodes.Load(id)
	if !ok {
		return nil, fmt.Errorf("a node with id %s does not exists", id)
	}

	info := node.Info()

	var client *rpc.Client
	var err error
	for start := time.Now(); time.Since(start) < 10*time.Second; time.Sleep(50 * time.Millisecond) {
		client, err = rpc.Dial(info.RPCListen)
		if err == nil {
			break
		}
	}
	if client == nil {
		return nil, fmt.Errorf("could not establish rpc connection: %v", err)
	}

	return client, nil
}

// HTTPBaseAddr returns the address for the HTTP API
func (s *Simulation) HTTPBaseAddr(id NodeID) (string, error) {
	node, ok := s.nodes.Load(id)
	if !ok {
		return "", fmt.Errorf("a node with id %s does not exists", id)
	}
	info := node.Info()
	return info.HTTPListen, nil
}

// Snapshot returns a snapshot of the simulation
func (s *Simulation) Snapshot() (NetworkSnapshot, error) {
	snap := NetworkSnapshot{}
	return snap, errors.New("not implemented")
}
