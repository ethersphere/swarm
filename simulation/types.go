package simulation

import (
	"io"
)

type Node interface {
	Info() NodeInfo
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

// TODO: All the fields of NodeInfo should probably just be Getter functions
// TODO: Mabye have a field `interfaces map[NodeInterface]string` to manage the connection strings for each interface?
//       Instead of having the RPCListen, HTTPListen, PprofListen strings

type NodeInfo struct {
	ID      NodeID
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
