package simulation

import "errors"

type DockerAdapter struct {
	directory string
}

type DockerAdapterConfig struct {
	DaemonAddr string
}

type DockerNode struct {
	config NodeConfig
}

func DefaultDockerAdapterConfig() DockerAdapterConfig {
	return DockerAdapterConfig{
		DaemonAddr: "unix:///var/run/docker.sock",
	}
}

func NewDockerAdapter(config DockerAdapterConfig) (*DockerAdapter, error) {
	return nil, errors.New("not implemented")
}

func (a *DockerAdapter) NewNode(config NodeConfig) (*Node, error) {
	return nil, errors.New("not implemented")
}

// Status returns the node status
func (n *DockerNode) Status() NodeStatus {
	return NodeStatus{}
}

// Start starts the node
func (n *DockerNode) Start() error {
	return errors.New("not implemented")
}

// Stop stops the node
func (n *DockerNode) Stop() error {
	return errors.New("not implemented")
}
