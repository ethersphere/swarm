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
	Enode   string
	BzzAddr string

	RPCListen   string // RPC listener address. Should be a valid ipc or websocket path
	HTTPListen  string // HTTP listener address: e.g. http://localhost:8500
	PprofListen string // PProf listener address: e.g http://localhost:6060
}

type NetworkSnapshot struct {
	Nodes []NodeSnapshot
}

type NodeSnapshot struct {
	Config NodeConfig
}
