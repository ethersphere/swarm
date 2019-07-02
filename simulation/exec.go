package simulation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ExecAdapter struct {
	directory string
	nodes     map[NodeID]*ExecNode
}

type ExecAdapterConfig struct {
	Directory string
}

type ExecNode struct {
	config NodeConfig
}

func NewExecAdapter(config ExecAdapterConfig) (*ExecAdapter, error) {
	if _, err := os.Stat(config.Directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("'%s' directory does not exist", config.Directory)
	}
	a := &ExecAdapter{
		directory: config.Directory,
		nodes:     make(map[NodeID]*ExecNode),
	}
	return a, nil
}

func (a *ExecAdapter) NewNode(config NodeConfig) (Node, error) {
	if _, ok := a.nodes[config.ID]; ok {
		return nil, fmt.Errorf("node '%s' already exists", config.ID)
	}

	dir := filepath.Join(a.directory, string(config.ID))
	if err := os.Mkdir(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create node directory: %s", err)
	}

	node := &ExecNode{
		config: config,
	}

	a.nodes[config.ID] = node

	return node, nil
}

// Status returns the node status
func (n *ExecNode) Status() NodeStatus {
	status := NodeStatus{
		ID: n.config.ID,
	}
	// TODO: fill the rest
	return status
}

// Start starts the node
func (n *ExecNode) Start() error {
	return errors.New("not implemented")
}

// Stop stops the node
func (n *ExecNode) Stop() error {
	return errors.New("not implemented")
}
