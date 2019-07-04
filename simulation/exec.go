package simulation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm"
)

// ExecAdapter can manage local exec nodes
type ExecAdapter struct {
	directory string
	nodes     map[NodeID]*ExecNode
}

// ExecAdapterConfig is used to configure an ExecAdapter
type ExecAdapterConfig struct {
	// Directory stores all the nodes' data directories
	Directory string
}

// ExecNode is a node that is executed locally
type ExecNode struct {
	adapter *ExecAdapter
	config  NodeConfig
	cmd     *exec.Cmd
	status  NodeStatus
}

// NewExecAdapter creates an ExecAdapter by receiving a ExecAdapterConfig
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

// NewNode creates a new node
func (a *ExecAdapter) NewNode(config NodeConfig) (Node, error) {
	if _, ok := a.nodes[config.ID]; ok {
		return nil, fmt.Errorf("node '%s' already exists", config.ID)
	}
	status := NodeStatus{
		ID: config.ID,
	}
	node := &ExecNode{
		config:  config,
		adapter: a,
		status:  status,
	}
	a.nodes[config.ID] = node
	return node, nil
}

// Status returns the node status
func (n *ExecNode) Status() NodeStatus {
	return n.status
}

// Start starts the node
func (n *ExecNode) Start() error {
	// Check if command already exists
	if n.cmd != nil {
		return fmt.Errorf("node %s is already running", n.config.ID)
	}

	// Create command line arguments
	args := []string{"swarm"}

	// Create data directory for this node
	dir := n.dataDir()
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create node directory: %s", err)
	}

	// Configure data directory
	args = append(args, "--datadir", dir)

	// Configure IPC path
	args = append(args, "--ipcpath", n.ipcPath())

	// Automatically allocate ports
	args = append(args, "--pprofport", "0")
	args = append(args, "--bzzport", "0")
	args = append(args, "--wsport", "0")
	args = append(args, "--port", "0")

	// Append user defined arguments
	args = append(args, n.config.Args...)

	// Start command
	n.cmd = &exec.Cmd{
		Path:   "/home/rafael/go/bin/swarm",
		Args:   args,
		Dir:    dir,
		Env:    n.config.Env,
		Stdout: n.config.Stdout,
		Stderr: n.config.Stderr,
	}

	if err := n.cmd.Start(); err != nil {
		n.cmd = nil
		return fmt.Errorf("error starting node %s: %s", n.config.ID, err)
	}

	// Wait for the node to start
	var client *rpc.Client
	var err error
	defer func() {
		if err != nil {
			n.Stop()
		}
	}()
	for start := time.Now(); time.Since(start) < 10*time.Second; time.Sleep(50 * time.Millisecond) {
		client, err = rpc.Dial(n.ipcPath())
		if err == nil {
			break
		}
	}
	if client == nil {
		return fmt.Errorf("could not establish rpc connection. node %s: %v", n.config.ID, err)
	}
	defer client.Close()

	var swarminfo swarm.Info
	err = client.Call(&swarminfo, "bzz_info")
	if err != nil {
		return fmt.Errorf("could not get info via rpc call. node %s: %v", n.config.ID, err)
	}

	var p2pinfo p2p.NodeInfo
	err = client.Call(&p2pinfo, "admin_nodeInfo")
	if err != nil {
		return fmt.Errorf("could not get info via rpc call. node %s: %v", n.config.ID, err)
	}

	n.status = NodeStatus{
		ID:         n.config.ID,
		Running:    true,
		Enode:      p2pinfo.Enode,
		BzzAddr:    swarminfo.BzzKey,
		RPCListen:  n.ipcPath(),
		HTTPListen: fmt.Sprintf("http://localhost:%s", swarminfo.Port),
	}

	return nil
}

// Stop stops the node
func (n *ExecNode) Stop() error {
	if n.cmd == nil {
		return nil
	}
	defer func() {
		n.cmd = nil
		n.status.Running = false
	}()
	// Try to gracefully terminate the process
	if err := n.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return n.cmd.Process.Kill()
	}

	// Wait for the process to terminate or timeout
	waitErr := make(chan error)
	go func() {
		waitErr <- n.cmd.Wait()
	}()
	select {
	case err := <-waitErr:
		return err
	case <-time.After(20 * time.Second):
		return n.cmd.Process.Kill()
	}

}

// ipcPath returns the path to the ipc socket
func (n *ExecNode) ipcPath() string {
	ipcfile := "bzzd.ipc"
	// On windows we can have to use pipes
	if runtime.GOOS == "windows" {
		return `\\.\pipe\` + ipcfile
	}
	return fmt.Sprintf("%s/%s", n.dataDir(), ipcfile)
}

// dataDir returns the path to the data directory that the node should use
func (n *ExecNode) dataDir() string {
	return filepath.Join(n.adapter.directory, string(n.config.ID))
}
