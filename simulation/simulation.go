package simulation

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
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

func (nm *NodeMap) Store(key NodeID, value Node) {
	nm.Lock()
	nm.internal[key] = value
	nm.Unlock()
}

type Simulation struct {
	adapter Adapter
	nodes   *NodeMap
}

func NewSimulation(adapter Adapter) *Simulation {
	sim := &Simulation{
		adapter: adapter,
		nodes:   NewNodeMap(),
	}
	return sim
}
func NewSimulationFromSnapshot(adapter *Adapter, snapshot *NetworkSnapshot) (*Simulation, error) {
	return nil, errors.New("not implemented")
}

func (s *Simulation) Get(id NodeID) (Node, error) {
	node, ok := s.nodes.Load(id)
	if !ok {
		return nil, fmt.Errorf("a node with id %s already exists", id)
	}
	return node, nil
}

func (s *Simulation) GetAll() ([]Node, error) {
	return nil, errors.New("not implemented")
}

// Create creates a given node with the NodeConfig
func (s *Simulation) Create(config NodeConfig) error {
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

// Stop stops a given node
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
func (s *Simulation) StartAll() error {
	return errors.New("not implemented")
}
func (s *Simulation) StopAll() error {
	return errors.New("not implemented")
}

func (s *Simulation) RPCClient(id NodeID) (*rpc.Client, error) {
	node, ok := s.nodes.Load(id)
	if !ok {
		return nil, fmt.Errorf("a node with id %s does not exists", id)
	}

	status := node.Status()

	if !status.Running {
		return nil, fmt.Errorf("node %s is not running", id)
	}

	var client *rpc.Client
	var err error
	for start := time.Now(); time.Since(start) < 10*time.Second; time.Sleep(50 * time.Millisecond) {
		client, err = rpc.Dial(status.RPCListen)
		if err == nil {
			break
		}
	}
	if client == nil {
		return nil, fmt.Errorf("could not establish rpc connection: %v", err)
	}

	return client, nil
}

func (s *Simulation) HTTPBaseAddr(id NodeID) (string, error) {
	node, ok := s.nodes.Load(id)
	if !ok {
		return "", fmt.Errorf("a node with id %s does not exists", id)
	}

	status := node.Status()

	if !status.Running {
		return "", fmt.Errorf("node %s is not running", id)
	}

	return status.HTTPListen, nil
}

func (s *Simulation) Snapshot() (NetworkSnapshot, error) {
	snap := NetworkSnapshot{}
	return snap, errors.New("not implemented")
}
