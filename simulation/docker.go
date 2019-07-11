package simulation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm"
	"github.com/ethersphere/swarm/log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
)

const (
	dockerP2PPort       = 30399
	dockerWebsocketPort = 8546
	dockerHTTPPort      = 8500
	dockerPProfPort     = 6060
)

type DockerAdapter struct {
	client *client.Client
	image  string
	nodes  map[NodeID]*DockerNode
}

type DockerAdapterConfig struct {
	// BuildContext can be used to build a docker image
	// from a Dockerfile and a context directory
	BuildContext DockerBuildContext
	// DockerImage points to an existing docker image
	// e.g. ethersphere/swarm:latest
	DockerImage string
	// DaemonAddr is the docker daemon address
	DaemonAddr string
}

// DockerBuildContext defines the build context to build
// local docker images
type DockerBuildContext struct {
	// Dockefile is the path to the dockerfile
	Dockerfile string
	// Directory is the directory that will be used
	// in the context of a docker build
	Directory string
	// Tag is used to tag the image
	Tag string
}

type DockerNode struct {
	config  NodeConfig
	adapter *DockerAdapter
	status  NodeStatus
	ipAddr  string
}

func DefaultDockerAdapterConfig() DockerAdapterConfig {
	return DockerAdapterConfig{
		DaemonAddr: client.DefaultDockerHost,
	}
}

func DefaultDockerBuildContext() DockerBuildContext {
	return DockerBuildContext{
		Dockerfile: "Dockerfile",
		Directory:  ".",
	}
}

// NewDockerAdapter creates an ExecAdapter by receiving a DockerAdapterConfig
func NewDockerAdapter(config DockerAdapterConfig) (*DockerAdapter, error) {
	if config.BuildContext.Dockerfile != "" && config.DockerImage != "" {
		return nil, fmt.Errorf("only one can be defined: BuildContext (%v) or DockerImage(%s)",
			config.BuildContext, config.DockerImage)
	}

	if config.BuildContext.Dockerfile == "" && config.DockerImage == "" {
		return nil, errors.New("required: BuildContext or ExecutablePath")
	}

	// Create docker client
	cli, err := client.NewClientWithOpts(
		client.WithHost(config.DaemonAddr),
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		return nil, fmt.Errorf("could not create the docker client: %v", err)
	}

	// Figure out which docker image should be used
	image := config.DockerImage

	// Build docker image
	if config.BuildContext.Dockerfile != "" {
		var err error
		image, err = buildImage(config.BuildContext)
		if err != nil {
			return nil, fmt.Errorf("could not build the docker image: %v", err)
		}
	}

	return &DockerAdapter{
		image:  image,
		client: cli,
		nodes:  make(map[NodeID]*DockerNode),
	}, nil
}

// NewNode creates a new node
func (a *DockerAdapter) NewNode(config NodeConfig) (Node, error) {
	if _, ok := a.nodes[config.ID]; ok {
		return nil, fmt.Errorf("node '%s' already exists", config.ID)
	}
	status := NodeStatus{
		ID: config.ID,
	}
	node := &DockerNode{
		config:  config,
		adapter: a,
		status:  status,
	}
	a.nodes[config.ID] = node
	return node, nil
}

// Status returns the node status
func (n *DockerNode) Status() NodeStatus {
	return n.status
}

// Start starts the node
func (n *DockerNode) Start() error {
	var err error
	defer func() {
		if err != nil {
			log.Error("Stopping node due to errors", "err", err)
			if err := n.Stop(); err != nil {
				log.Error("Failed stopping node", "err", err)
			}
		}
	}()

	// Define arguments
	args := []string{}

	// Append user defined arguments
	args = append(args, n.config.Args...)

	// Append network ports arguments
	args = append(args, "--pprofport", strconv.Itoa(dockerPProfPort))
	args = append(args, "--bzzport", strconv.Itoa(dockerHTTPPort))
	args = append(args, "--ws")
	// TODO: Can we get the APIs from somewhere instead of hardcoding them here?
	args = append(args, "--wsapi", "admin,net,debug,bzz,accounting,hive")
	args = append(args, "--wsport", strconv.Itoa(dockerWebsocketPort))
	args = append(args, "--wsaddr", "0.0.0.0")
	args = append(args, "--wsorigins", "*")
	args = append(args, "--port", strconv.Itoa(dockerP2PPort))
	args = append(args, "--natif", "eth0")

	// Start the node via a container
	ctx := context.Background()
	dockercli := n.adapter.client

	resp, err := dockercli.ContainerCreate(ctx, &container.Config{
		Image: n.adapter.image,
		Cmd:   args,
		Env:   n.config.Env,
	}, &container.HostConfig{}, nil, n.containerName())
	if err != nil {
		return fmt.Errorf("failed to create container %s: %v", n.containerName(), err)
	}

	if err := dockercli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %v", n.containerName(), err)
	}

	// Get container logs

	go func() {
		// Stderr
		stderr, err := dockercli.ContainerLogs(context.Background(), n.containerName(), types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: false,
			Follow:     true,
		})
		if err != nil && err != io.EOF {
			log.Error("Error getting stderr container logs", "err", err)
		}
		defer stderr.Close()
		if _, err := io.Copy(n.config.Stderr, stderr); err != nil && err != io.EOF {
			log.Error("Error writing stderr container logs", "err", err)
		}
	}()
	go func() {
		// Stdout
		stdout, err := dockercli.ContainerLogs(context.Background(), n.containerName(), types.ContainerLogsOptions{
			ShowStderr: false,
			ShowStdout: true,
			Follow:     true,
		})
		if err != nil && err != io.EOF {
			log.Error("Error getting stdout container logs", "err", err)
		}
		defer stdout.Close()
		if _, err := io.Copy(n.config.Stdout, stdout); err != nil && err != io.EOF {
			log.Error("Error writing stdout container logs", "err", err)
		}
	}()

	// Get container info
	cinfo := types.ContainerJSON{}

	for start := time.Now(); time.Since(start) < 10*time.Second; time.Sleep(50 * time.Millisecond) {
		cinfo, err = dockercli.ContainerInspect(ctx, n.containerName())
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("could not get container info: %v", err)
	}

	// Get the container IP addr
	n.ipAddr = cinfo.NetworkSettings.IPAddress

	// Wait for the node to start
	var client *rpc.Client
	wsAddr := fmt.Sprintf("ws://%s:%d", n.ipAddr, dockerWebsocketPort)
	for start := time.Now(); time.Since(start) < 30*time.Second; time.Sleep(50 * time.Millisecond) {
		client, err = rpc.Dial(wsAddr)
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
		ID:          n.config.ID,
		Running:     true,
		Enode:       strings.Replace(p2pinfo.Enode, "127.0.0.1", n.ipAddr, 1),
		BzzAddr:     swarminfo.BzzKey,
		RPCListen:   wsAddr,
		HTTPListen:  fmt.Sprintf("http://%s:%d", n.ipAddr, dockerHTTPPort),
		PprofListen: fmt.Sprintf("http://%s:%d", n.ipAddr, dockerPProfPort),
	}

	return nil
}

// Stop stops the node
func (n *DockerNode) Stop() error {
	cli := n.adapter.client

	var stopTimeout = 30 * time.Second
	err := cli.ContainerStop(context.Background(), n.containerName(), &stopTimeout)
	if err != nil {
		return fmt.Errorf("failed to stop container %s : %v", n.containerName(), err)
	}

	err = cli.ContainerRemove(context.Background(), n.containerName(), types.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove container %s : %v", n.containerName(), err)
	}
	n.status.Running = false
	return nil
}

func (n *DockerNode) containerName() string {
	return fmt.Sprintf("sim-docker-%s", n.config.ID)
}

// buildImage builds a docker image and returns the image identifier (tag).
func buildImage(buildContext DockerBuildContext) (string, error) {
	// Connect to docker daemon
	c, err := client.NewClientWithOpts(
		client.WithHost(client.DefaultDockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return "", fmt.Errorf("could not create docker client: %v", err)
	}
	defer c.Close()
	// Use directory for build context
	ctx, err := archive.TarWithOptions(buildContext.Directory, &archive.TarOptions{})
	if err != nil {
		return "", err
	}

	// Default image tag
	imageTag := "sim-docker:latest"

	// Use a tag if one is defined
	if buildContext.Tag != "" {
		imageTag = buildContext.Tag
	}

	// Build image
	opts := types.ImageBuildOptions{
		SuppressOutput: false,
		PullParent:     true,
		Tags:           []string{imageTag},
		Dockerfile:     buildContext.Dockerfile,
	}

	buildResp, err := c.ImageBuild(context.Background(), ctx, opts)
	if err != nil {
		return "", fmt.Errorf("build error: %v", err)
	}

	// Parse build output
	d := json.NewDecoder(buildResp.Body)
	var event *jsonmessage.JSONMessage
	for {
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		log.Info("Docker build", "msg", event.Stream)
		if event.Error != nil {
			log.Error("Docker build error", "err", event.Error.Message)
			return "", fmt.Errorf("failed to build docker image: %v", event.Error)
		}
	}
	return imageTag, nil
}
