package discovery

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/state"

	"github.com/pborman/uuid"
)

const (
	bootNodeCount             = 3
	defaultHealthCheckRetires = 10
	defaultHealthCheckDelay   = time.Millisecond * 250
)

var (
	mu              sync.Mutex
	bootNodes       []*discover.NodeID
	events          chan *p2p.PeerEvent
	ids             []discover.NodeID
	dynamicServices adapters.Services
)

// Test to verify that restarted nodes will reach healthy state
//
// First, it brings up and connects bootnodes, and connects each of the further nodes
// to a random bootnode
//
// if network is healthy, it proceeds to randomly and asynchronously stop and start nodes
// and performing new health checks after they have connected to one of their previous peers
func TestDynamicDiscovery(t *testing.T) {
	t.Run("8/3/sim", dynamicDiscoverySimulation)
}

// node count must be higher than 3 and higher than disconnect count for now
func dynamicDiscoverySimulation(t *testing.T) {

	// quitC is used to make sure the async sim events select loops exit
	var quitC *chan struct{}
	defer func(quitC *chan struct{}) {
		if quitC != nil {
			close(*quitC)
		}
	}(quitC)

	// this directory will keep the state store dbs
	dir, err := ioutil.TempDir("", "dynamic-discovery")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// discovery service
	dynamicServices := adapters.Services{
		"discovery": newDynamicServices(dir),
	}

	// set up locals we need
	paramstring := strings.Split(t.Name(), "/")
	nodeCount, _ := strconv.ParseInt(paramstring[1], 10, 0)
	numUpDowns, _ := strconv.ParseInt(paramstring[2], 10, 0)
	adapter := paramstring[3]
	bootNodes = make([]*discover.NodeID, bootNodeCount)
	events = make(chan *p2p.PeerEvent)
	ids = make([]discover.NodeID, nodeCount)

	log.Info("starting dynamic test", "nodecount", nodeCount, "adaptertype", adapter)

	// select adapter
	var a adapters.NodeAdapter
	if adapter == "exec" {
		dirname, err := ioutil.TempDir(".", "")
		if err != nil {
			t.Fatal(err)
		}
		a = adapters.NewExecAdapter(dirname)
	} else if adapter == "sock" {
		a = adapters.NewSocketAdapter(dynamicServices)
	} else if adapter == "tcp" {
		a = adapters.NewTCPAdapter(dynamicServices)
	} else if adapter == "sim" {
		a = adapters.NewSimAdapter(dynamicServices)
	}

	// create network
	net := simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "discovery",
	})
	defer net.Shutdown()

	// create simnodes
	for i := 0; i < int(nodeCount); i++ {
		conf := adapters.RandomNodeConfig()
		node, err := net.NewNodeWithConfig(conf)
		if err != nil {
			t.Fatalf("error starting node: %s", err)
		}
		ids[i] = node.ID()
		log.Info("new node", "id", ids[i])
		if i < bootNodeCount {
			bootNodes[i] = &ids[i]
		}
	}

	// sim step 1
	// start nodes, trigger them on node up event from sim
	trigger := make(chan discover.NodeID)
	events := make(chan *simulations.Event)
	sub := net.Events().Subscribe(events)
	q := make(chan struct{})
	quitC = &q

	action := func(ctx context.Context) error {
		go func(quitC chan struct{}) {
			for {
				select {
				case ev := <-events:
					if ev == nil {
						panic("got nil event")
					}
					if ev.Type == simulations.EventTypeNode {
						if ev.Node.Up {
							log.Info("got node up event", "event", ev, "node", ev.Node.Config.ID)
							trigger <- ev.Node.Config.ID
						}
					}

				case <-quitC:
					log.Warn("got quit action up")
					return
				}

			}
		}(*quitC)
		go func() {
			for _, n := range ids {
				if err := net.Start(n); err != nil {
					t.Fatalf("error starting node: %s", err)
				}
				log.Info("network start returned", "node", n)
			}
		}()
		return nil

	}

	check := func(ctx context.Context, nodeId discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		log.Info("trigger expect up", "node", nodeId)
		return true, nil
	}

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result := simulations.NewSimulation(net).Run(ctx, &simulations.Step{
		Action:  action,
		Trigger: trigger,
		Expect: &simulations.Expectation{
			Nodes: ids,
			Check: check,
		},
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
	close(*quitC)

	// sim step 2
	// connect the three bootnodes together
	// connect each of the other nodes to a random bootnode
	q = make(chan struct{})
	quitC = &q
	action = func(ctx context.Context) error {
		go func(quitC chan struct{}) {
			for {
				select {
				case ev := <-events:
					if ev.Type == simulations.EventTypeConn {
						if ev.Conn.Up {
							log.Info(fmt.Sprintf("got conn up event %v", ev))
							trigger <- ev.Conn.One
						}
					}
				case <-quitC:
					log.Debug("got quit action connect")
					return
				}
			}
		}(*quitC)
		go func() {
			for i := range ids {
				var j int
				if i == 0 {
					continue
				}
				if i < len(bootNodes) {
					j = i - 1
				} else {
					j = rand.Intn(len(bootNodes) - 1)
				}

				if err := net.Connect(ids[i], ids[j]); err != nil {
					t.Fatalf("error connecting node %x => bootnode %x: %s", ids[i], ids[j], err)
				}
				log.Info("network connect returned", "one", ids[i], "other", ids[j])
			}
		}()
		return nil
	}

	check = func(ctx context.Context, id discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		log.Info("trigger expect conn", "node", id)
		return true, nil
	}

	timeout = 30 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result = simulations.NewSimulation(net).Run(ctx, &simulations.Step{
		Action:  action,
		Trigger: trigger,
		Expect: &simulations.Expectation{
			Nodes: ids[1:],
			Check: check,
		},
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	close(*quitC)
	*quitC = nil
	sub.Unsubscribe()
	sub = nil

	// check health of all nodes
	ctrl := newNodeCtrl(net, ids)
	for _, n := range net.GetNodes() {
		err := ctrl.checkHealth(n, 0)
		if err != nil {
			t.Fatalf("Node %v failed health check", n)
		}
		log.Info("node health ok", "node", n)
	}

	// asynchronously restart random nodes
	wg := sync.WaitGroup{}
	for i := 0; i < int(numUpDowns); i++ {
		wg.Add(1)
		dur := ctrl.randomDelay(0)
		time.Sleep(dur)
		go func() {
			if err := ctrl.nodeUpDown(); err != nil {
				wg.Done()
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	log.Warn("exiting test")
}

// upNodes and upAddrs: node arrays to calculate peerpot from. After restart, a node is added to the array immediately after starting and connect to bootnode
// readyNodes: node array to choose next node to restart from. After restart, a node is added to the array after it is healthy
type nodeCtrl struct {
	seq                int
	net                *simulations.Network
	readyNodes         []discover.NodeID
	upNodes            []discover.NodeID
	upAddrs            [][]byte
	addrIdx            map[discover.NodeID]string
	pot                map[string]*network.PeerPot
	mu                 sync.Mutex
	healthCheckRetries int
	healthCheckDelay   time.Duration
}

func newNodeCtrl(net *simulations.Network, nodes []discover.NodeID) *nodeCtrl {
	ctrl := &nodeCtrl{
		net:                net,
		healthCheckRetries: defaultHealthCheckRetires,
		healthCheckDelay:   defaultHealthCheckDelay,
		addrIdx:            make(map[discover.NodeID]string),
	}
	for i := 0; i < len(nodes); i++ {
		ctrl.upNodes = append(ctrl.upNodes, nodes[i])
		ctrl.readyNodes = append(ctrl.readyNodes, nodes[i])
		addr := network.ToOverlayAddr(ids[i].Bytes())
		ctrl.upAddrs = append(ctrl.upAddrs, addr)
		ctrl.addrIdx[nodes[i]] = common.Bytes2Hex(addr)
		log.Debug("init nodeCtrl", "idx", i, "upnode", ctrl.upNodes[i], "upaddr", fmt.Sprintf("%x", ctrl.upAddrs[i]))
	}
	ctrl.pot = network.NewPeerPotMap(testMinProxBinSize, ctrl.upAddrs)
	for k, p := range ctrl.pot {
		for i, nn := range p.NNSet {
			log.Debug("init nodeCtrl nnset", "i", i, "node", k, "nn", fmt.Sprintf("%x", nn))
		}
	}
	return ctrl
}

func (self *nodeCtrl) nodeUpDown() error {

	self.mu.Lock()

	if len(self.readyNodes) == 1 {
		self.mu.Unlock()
		return errors.New("uh-oh, spaghettios: ran out of readyNodes")
	}

	// used for logging
	self.seq++
	seq := self.seq

	// choose a random node to restart from nodes not currently stopped or getting health checked
	nodeIdx := rand.Intn(len(self.readyNodes) - 1)
	nodeId := self.readyNodes[nodeIdx]
	self.readyNodes[nodeIdx] = self.readyNodes[len(self.readyNodes)-1]
	self.readyNodes = self.readyNodes[:len(self.readyNodes)-1]

	// find the selected node in the upNodes and upAddrs arrays (they have identical indices)
	var found bool
	for i, up := range self.upNodes {
		if up == nodeId {
			self.upNodes[i] = self.upNodes[len(self.upNodes)-1]
			self.upAddrs[i] = self.upAddrs[len(self.upAddrs)-1]
			self.upNodes = self.upNodes[:len(self.upNodes)-1]
			self.upAddrs = self.upAddrs[:len(self.upAddrs)-1]
			found = true
			break
		}
	}

	// this shouldn't happen
	if !found {
		self.mu.Unlock()
		return fmt.Errorf("node %v listed as ready but not found in uplist", nodeId)
	}

	// stop the node
	log.Info("Restart node: stop", "seq", seq, "id", nodeId, "addr", fmt.Sprintf("%x", self.upAddrs[nodeIdx]), "upnodes", len(self.upNodes), "readynodes", len(self.readyNodes))
	self.pot = network.NewPeerPotMap(testMinProxBinSize, self.upAddrs)

	if err := self.net.Stop(nodeId); err != nil {
		self.mu.Unlock()
		return err
	}
	self.mu.Unlock()

	// wait a bit then bring back up
	time.Sleep(self.randomDelay(0))

	// start the node again
	// wait for the up event and a conn event which involves the node
	self.mu.Lock()
	log.Info("Restart node: start", "seq", seq, "id", nodeId, "addr", fmt.Sprintf("%x", network.ToOverlayAddr(nodeId.Bytes())), "upnodes", len(self.upNodes), "readynodes", len(self.readyNodes))
	self.mu.Unlock()

	var err error
	events := make(chan *simulations.Event)
	sub := self.net.Events().Subscribe(events)
	defer sub.Unsubscribe()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	errC := make(chan error)
	quitC := make(chan struct{})
	go func(errC chan error) {
		for {
			select {
			case <-quitC:
				return
			case ev := <-events:
				if ev == nil {
					errC <- errors.New("got nil event")
					return
				} else if ev.Type == simulations.EventTypeNode {
					if ev.Node.Up && ev.Node.Config.ID == nodeId {
						log.Info("got node up event", "event", ev, "node", ev.Node.Config.ID)
					}
				} else if ev.Type == simulations.EventTypeConn {
					if ev.Conn.Up && (ev.Conn.One == nodeId || ev.Conn.Other == nodeId) {
						log.Info("got node conn event", "event", ev, "one", ev.Conn.One, "other", ev.Conn.Other)
						errC <- nil
					}
				}

			case <-ctx.Done():
				errC <- ctx.Err()
				return
			}

		}
	}(errC)

	if err := self.net.Start(nodeId); err != nil {
		return err
	}
	if err = <-errC; err != nil {
		return err
	}
	close(quitC)

	sub.Unsubscribe()

	// node is now up, we are connected to a peer
	// we can add the node back to the uplist
	self.mu.Lock()
	self.upNodes = append(self.upNodes, nodeId)
	self.upAddrs = append(self.upAddrs, network.ToOverlayAddr(nodeId.Bytes()))
	self.mu.Unlock()

	// then we poll for health
	n := self.net.GetNode(nodeId)
	err = self.checkHealth(n, seq)
	if err != nil {
		return err //"failed health", "node", n, "seq", seq
	}

	// celebrate good times, come on!
	log.Info("Restarted node regained health", "id", nodeId, "seq", seq)

	// now we add the node to the readyNode list, so it can be selected as a node to restart again
	self.mu.Lock()
	self.readyNodes = append(self.readyNodes, nodeId)
	self.mu.Unlock()

	return nil
}

func (self *nodeCtrl) randomDelay(maxDuration int) time.Duration {
	if maxDuration == 0 {
		maxDuration = 1000000000
	}
	timeout := rand.Intn(maxDuration) + 10000000
	ns := fmt.Sprintf("%dns", timeout)
	dur, _ := time.ParseDuration(ns)
	return dur
}

// check health nodeCtrl.healthCheckRetries times with small delays inbetween before giving up
func (self *nodeCtrl) checkHealth(node *simulations.Node, seq int) error {
	client, err := node.Client()
	if err != nil {
		return fmt.Errorf("can't get node rpc for node %v: %v", node.ID().TerminalString(), err)
	}

	i := 0
	for {
		if i > self.healthCheckRetries {
			return fmt.Errorf("max health retries for node %v (seq %d)", node.ID().TerminalString(), seq)
		}
		i++
		time.Sleep(self.healthCheckDelay)
		healthy := &network.Health{}
		self.mu.Lock()
		self.pot = network.NewPeerPotMap(testMinProxBinSize, self.upAddrs)

		if _, ok := self.pot[self.addrIdx[node.ID()]]; !ok {
			self.mu.Unlock()
			return fmt.Errorf("missing node in pot", "nodeid", node.ID(), "seq", seq)
		}
		if err := client.Call(&healthy, "hive_healthy", self.pot[self.addrIdx[node.ID()]]); err != nil {
			self.mu.Unlock()
			return fmt.Errorf("error retrieving node health by rpc for node %v: %v", node.ID(), err)
		}
		self.mu.Unlock()
		if !(healthy.KnowNN && healthy.GotNN && healthy.Full) {
			log.Debug("healthy not yet reached", "id", node.ID(), "addr", self.addrIdx[node.ID()], "attempt", i, "health", healthy.Hive)
			continue
		}
		break
	}
	return nil
}

func newDynamicServices(storePath string) func(*adapters.ServiceContext) (node.Service, error) {
	return func(ctx *adapters.ServiceContext) (node.Service, error) {
		host := adapters.ExternalIP()

		addr := network.NewAddrFromNodeIDAndPort(ctx.Config.ID, host, ctx.Config.Port)

		kp := network.NewKadParams()
		kp.MinProxBinSize = testMinProxBinSize
		kp.MaxBinSize = 3
		kp.MinBinSize = 1
		kp.MaxRetries = 10
		kp.RetryExponent = 2
		kp.RetryInterval = 50000000
		if ctx.Config.Reachable != nil {
			kp.Reachable = func(o network.OverlayAddr) bool {
				return ctx.Config.Reachable(o.(*network.BzzAddr).ID())
			}
		}
		kad := network.NewKademlia(addr.Over(), kp)

		hp := network.NewHiveParams()
		hp.KeepAliveInterval = 200 * time.Millisecond

		config := &network.BzzConfig{
			OverlayAddr:  addr.Over(),
			UnderlayAddr: addr.Under(),
			HiveParams:   hp,
		}

		uuid := uuid.NewUUID()
		stateStore, err := state.NewDBStore(filepath.Join(storePath, fmt.Sprintf("state-store-%s.db", uuid)))
		if err != nil {
			return nil, err
		}
		return network.NewBzz(config, kad, stateStore, nil, nil), nil
	}
}
