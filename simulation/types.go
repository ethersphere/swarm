package simulation

import "io"

type Node interface {
	Status() NodeStatus
	// Start starts the node
	Start() error
	// Stop stops the node
	Stop() error
}

type Adapter interface {
	// NewNode creates a new node based on the NodeConfig
	NewNode(config NodeConfig) (Node, error)
	// InfluxAddr() string
	// JaegerAddr() string
}

type NodeID string

type NodeConfig struct {
	// Arbitrary string used to identify a node
	ID NodeID
	// Command line arguments
	Args []string
	// Environment variables
	Env []string
	// Stdout and Stderr specify the nodes' standard output and error
	Stdout io.Writer
	Stderr io.Writer
}

// All the fields of NodeStatus should probably just be Getter functions
type NodeStatus struct {
	ID      NodeID
	Running bool // True if the node is running
	Enode   []byte
	BzzAddr []byte

	RPCAddr   string // RPC addr. Should ideally be a websocket address for remote RPC calls: e.g. ws://localhost:8501
	HTTPAddr  string // HTTP addr: e.g. http://localhost:8500
	PprofAddr string // pprof,metrics, etc ?
}

type NetworkSnapshot struct {
	Nodes []NodeSnapshot
}

type NodeSnapshot struct {
	Config NodeConfig
}
