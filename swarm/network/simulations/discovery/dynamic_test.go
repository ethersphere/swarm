package discovery

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
)

const (
	bootNodeCount             = 3
	quarantineNodeCount       = 5
	defaultHealthCheckRetires = 10
	defaultHealthCheckDelay   = time.Millisecond * 250
)

var (
	//upNodes     []*discover.NodeID
	//upNodesLast int
	mu        sync.Mutex
	bootNodes []*discover.NodeID
	rpcs      map[discover.NodeID]*rpc.Client
	subs      map[discover.NodeID]*rpc.ClientSubscription
	events    chan *p2p.PeerEvent
	ids       []discover.NodeID
)

// node count should be higher than disconnect count for now
func TestDynamicDiscovery(t *testing.T) {
	t.Run("16/10/sim", dynamicDiscoverySimulation)
}

func dynamicDiscoverySimulation(t *testing.T) {

	paramstring := strings.Split(t.Name(), "/")
	nodeCount, _ := strconv.ParseInt(paramstring[1], 10, 0)
	numUpDowns, _ := strconv.ParseInt(paramstring[2], 10, 0)
	adapter := paramstring[3]

	bootNodes = make([]*discover.NodeID, 3)
	events = make(chan *p2p.PeerEvent)
	subs = make(map[discover.NodeID]*rpc.ClientSubscription)
	rpcs = make(map[discover.NodeID]*rpc.Client)
	ids = make([]discover.NodeID, nodeCount)

	log.Info("dynamic test", "nodecount", nodeCount, "adaptertype", adapter)

	var a adapters.NodeAdapter
	if adapter == "exec" {
		dirname, err := ioutil.TempDir(".", "")
		if err != nil {
			t.Fatal(err)
		}
		a = adapters.NewExecAdapter(dirname)
	} else if adapter == "sock" {
		a = adapters.NewSocketAdapter(services)
	} else if adapter == "tcp" {
		a = adapters.NewTCPAdapter(services)
	} else if adapter == "sim" {
		a = adapters.NewSimAdapter(services)
	}
	// create network
	net := simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: serviceName,
	})
	defer net.Shutdown()

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

		if err = net.Start(node.ID()); err != nil {
			t.Fatalf("error starting node %s: %s", node.ID().TerminalString(), err)
		}
		client, err := node.Client()
		if err != nil {
			t.Fatal(err)
		}
		rpcs[ids[i]] = client
		sub, err := client.Subscribe(context.Background(), "admin", events, "peerEvents")
		if err != nil {
			t.Fatalf("error getting peer events for node %v: %s", ids[i], err)
		}
		subs[ids[i]] = sub
	}

	// run a simulation which connects the nodes to the bootnodes
	trigger := make(chan discover.NodeID)
	action := func(ctx context.Context) error {
		for i := 0; i < int(nodeCount); i++ {
			var j int
			if i < bootNodeCount {
				if i == bootNodeCount-1 {
					j = 0
				} else {
					j = i + 1
				}
			} else {
				j = rand.Intn(bootNodeCount)
			}
			go func(i, j int) {
				net.Connect(ids[i], ids[j])
				// TODO: replace with simevents check
				time.Sleep(time.Second * 2)
				trigger <- ids[i]
			}(i, j)
		}
		return nil
	}

	ctrl := newNodeCtrl(net, ids)
	// construct the peer pot, so that kademlia health can be checked
	check := func(ctx context.Context, id discover.NodeID) (bool, error) {
		var err error
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case e := <-events:
			if e.Type == p2p.PeerEventTypeAdd {
				n := net.GetNode(id)
				err := ctrl.checkHealth(n, 0)
				log.Debug("check health return", "id", id, "err", err)
				return true, err
			}
		case err := <-subs[id].Err():
			if err != nil {
				log.Error(fmt.Sprintf("error getting peer events for node %v", id), "err", err)
			}
		default:
		}
		return false, err
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

// upNodes and upAddrs: node arrays to calculate peerpot from. After restart, a node is added to the array immediately after starting and connect to bootnode
// readyNodes: node array to choose next node to restart from. After restart, a node is added to the array after it is healthy
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

// TODO: move to action/expect in sim step; sync until completed stop. then async sleep and trigger.
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
		return fmt.Errorf("node %v listed as ready but not found in uplist", nodeId)
	}

	// stop the node
	log.Info("Restart node: stop", "id", nodeId, "addr", fmt.Sprintf("%x", self.upAddrs[nodeIdx]), "upnodes", len(self.upNodes), "readynodes", len(self.readyNodes))
	self.pot = network.NewPeerPotMap(testMinProxBinSize, self.upAddrs)
	self.mu.Unlock()
	if err := self.net.Stop(nodeId); err != nil {
		self.mu.Unlock()
		return err
	}

	// wait a bit then bring back up
	time.Sleep(self.randomDelay(0))

	// start the node again
	self.mu.Lock()
	log.Info("Restart node: start", "id", nodeId, "addr", fmt.Sprintf("%x", network.ToOverlayAddr(nodeId.Bytes())), "upnodes", len(self.upNodes), "readynodes", len(self.readyNodes))
	self.mu.Unlock()
	if err := self.net.Start(nodeId); err != nil {
		return err
	}

	// wait for the node to start
	// TODO: how to determine if a node is ready?
	time.Sleep(time.Second * 1)

	// in the meantime, we might temporarily be short of nodes in the network
	// if we are then wait a bit and try again
	for {
		self.mu.Lock()
		if len(self.upNodes)-1 > 1 {
			break
		}
		self.mu.Unlock()
	}

	// now we can add the node back to the uplist
	self.upNodes = append(self.upNodes, nodeId)
	self.upAddrs = append(self.upAddrs, network.ToOverlayAddr(nodeId.Bytes()))

	// we have a new kademlia now. We need to connect to at least one other node
	// we choose it at random from the available ones in the peerpot list
	nodeOtherIdx := rand.Intn(len(self.upNodes) - 1)
	nodeOtherId := self.upNodes[nodeOtherIdx]

	// check if it's incidentally already connected
	// if not, try to connect.
	if cn := self.net.GetConn(nodeId, nodeOtherId); cn != nil {
		if !cn.Up {
			if err := self.net.Connect(nodeId, nodeOtherId); err != nil {
				self.mu.Unlock()
				return err
			}
		}
	}
	self.mu.Unlock()

	// node is now up, we are connected or connecting to a peer
	// so we poll for health
	n := self.net.GetNode(nodeId)
	err := self.checkHealth(n, seq)
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
			//log.Error("Missing node in pot", "nodeid", node.ID(), "seq", seq)
			//			for i, n := range self.upNodes {
			//				log.Trace("ctrl dump", "node", n.TerminalString(), "addr", fmt.Sprintf("%x", self.upAddrs[i]))
			//			}
			//			for k, p := range self.pot {
			//				log.Trace("pot dump", "node", k.TerminalString(), "pot", p)
			//			}
			//			self.mu.Unlock()
			//			continue
		}
		if err := client.Call(&healthy, "hive_healthy", self.pot[self.addrIdx[node.ID()]]); err != nil {
			self.mu.Unlock()
			return fmt.Errorf("error retrieving node health by rpc for node %v: %v", node.ID(), err)
		}
		self.mu.Unlock()
		if !(healthy.KnowNN && healthy.GotNN && healthy.Full) {
			log.Debug("healthy not yet reached", "id", node.ID(), "attempt", i)
			continue
		}
		break
	}
	return nil
}

func newServices(ctx *adapters.ServiceContext) (node.Service, error) {
	host := adapters.ExternalIP()

	addr := network.NewAddrFromNodeIDAndPort(ctx.Config.ID, host, ctx.Config.Port)

	kp := network.NewKadParams()
	kp.MinProxBinSize = testMinProxBinSize
	kp.MaxBinSize = 3
	kp.MinBinSize = 1
	kp.MaxRetries = 4000
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

	return network.NewBzz(config, kad, nil, nil, nil), nil
}
