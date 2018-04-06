package discovery

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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
	upNodes     []*discover.NodeID
	upNodesLast int
	mu          sync.Mutex
	bootNodes   []*discover.NodeID
	rpcs        map[discover.NodeID]*rpc.Client
	subs        map[discover.NodeID]*rpc.ClientSubscription
	events      chan *p2p.PeerEvent
	ids         []discover.NodeID
)

func TestDynamicDiscovery(t *testing.T) {
	t.Run("10/10/sim", dynamicDiscoverySimulation)
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

	ctrl := newNodeCtrl(net, ids)

	trigger := make(chan discover.NodeID)
	// run a simulation which connects the nodes to the bootnodes
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

	// construct the peer pot, so that kademlia health can be checked
	check := func(ctx context.Context, id discover.NodeID) (bool, error) {
		log.Warn("checking", "id", id)
		var err error
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case e := <-events:
			if e.Type == p2p.PeerEventTypeAdd {
				log.Info("check got add", "id", id)
				n := net.GetNode(id)
				err := ctrl.checkHealth(n)
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
	upNodes            []discover.NodeID
	upAddrs            [][]byte
	nodeCursor         int
	pot                map[discover.NodeID]*network.PeerPot
	mu                 sync.Mutex
	healthCheckRetries int
	healthCheckDelay   time.Duration
}

func newNodeCtrl(net *simulations.Network, nodes []discover.NodeID) *nodeCtrl {
	ctrl := &nodeCtrl{
		net:                net,
		healthCheckRetries: defaultHealthCheckRetires,
		healthCheckDelay:   defaultHealthCheckDelay,
	}
	for ctrl.nodeCursor = 0; ctrl.nodeCursor < len(nodes); ctrl.nodeCursor++ {
		ctrl.upNodes = append(ctrl.upNodes, nodes[ctrl.nodeCursor])
		ctrl.upAddrs = append(ctrl.upAddrs, network.ToOverlayAddr(ids[ctrl.nodeCursor].Bytes()))
		log.Debug("init nodeCtrl", "idx", ctrl.nodeCursor, "upnode", ctrl.upNodes[ctrl.nodeCursor], "upaddr", fmt.Sprintf("%x", ctrl.upAddrs[ctrl.nodeCursor]))
	}
	ctrl.pot = network.NewPeerPot(testMinProxBinSize, ctrl.upNodes, ctrl.upAddrs)
	for k, p := range ctrl.pot {
		for i, nn := range p.NNSet {
			log.Debug("init nodeCtrl nnset", "i", i, "node", k.TerminalString(), "nn", fmt.Sprintf("%x", nn))
		}
	}
	return ctrl
}

func (self *nodeCtrl) nodeUpDown() error {
	mu.Lock()
	self.seq++
	nodeIdx := rand.Intn(len(self.upNodes) - 1)

	nodeId := self.upNodes[nodeIdx]
	log.Info("Restart node: stop", "id", nodeId, "addr", fmt.Sprintf("%x", self.upAddrs[nodeIdx]), "upnodes", len(self.upNodes))
	n := self.net.GetNode(nodeId)
	self.upNodes[nodeIdx] = self.upNodes[len(self.upNodes)-1]
	self.upAddrs[nodeIdx] = self.upAddrs[len(self.upAddrs)-1]
	self.upNodes = self.upNodes[:len(self.upNodes)-1]
	self.upAddrs = self.upAddrs[:len(self.upAddrs)-1]
	self.nodeCursor--
	self.pot = network.NewPeerPot(testMinProxBinSize, self.upNodes, self.upAddrs)
	//	if err := self.net.Stop(nodeId); err != nil {
	//		return err
	//	}
	for _, c := range self.net.Conns {
		if (c.One == nodeId || c.Other == nodeId) && c.Up {
			err := self.net.Disconnect(c.One, c.Other)
			if err != nil {
				mu.Unlock()
				return fmt.Errorf("Could not disconnect %v =/=> %v: %v", c.One.TerminalString(), c.Other.TerminalString(), err)
			}
		}
	}
	mu.Unlock()

	// wait a bit then bring back up
	time.Sleep(self.randomDelay(0))

	mu.Lock()
	log.Info("Restart node: start", "id", nodeId, "addr", fmt.Sprintf("%x", network.ToOverlayAddr(nodeId.Bytes()), "upnodes", len(self.upNodes)))
	//	if err := self.net.Start(nodeId); err != nil {
	//		return err
	//	}

	// wait a bit then bring back up
	//time.Sleep(time.Second * 1)

	// add it back into the uplist
	self.upNodes = append(self.upNodes, nodeId)
	self.upAddrs = append(self.upAddrs, network.ToOverlayAddr(nodeId.Bytes()))
	self.nodeCursor++
	self.pot = network.NewPeerPot(testMinProxBinSize, self.upNodes, self.upAddrs)
	if err := self.checkHealth(n); err != nil {
		log.Debug("failed health", "node", n, "seq", self.seq)
		return err
	}
	mu.Unlock()

	log.Debug("Restarted node regained health", "id", nodeId)
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

func (self *nodeCtrl) checkHealth(node *simulations.Node) error {
	client, err := node.Client()
	if err != nil {
		return fmt.Errorf("can't get node rpc for node %v: %v", node.ID().TerminalString(), err)
	}

	i := 0
	for {
		if i > self.healthCheckRetries {
			return fmt.Errorf("max health retries for node %v", node.ID().TerminalString())
		}
		time.Sleep(self.healthCheckDelay)
		healthy := &network.Health{}
		self.mu.Lock()
		if _, ok := self.pot[node.ID()]; !ok {
			log.Error("Missing node in pot", "nodeid", node.ID())
			for i, n := range self.upNodes {
				log.Debug("ctrl dump", "node", n.TerminalString(), "addr", fmt.Sprintf("%x", self.upAddrs[i]))
			}
			for k, p := range self.pot {
				log.Debug("pot dump", "node", k.TerminalString(), "pot", p)
			}
			continue
		}
		if err := client.Call(&healthy, "hive_healthy", self.pot[node.ID()]); err != nil {
			self.mu.Unlock()
			return fmt.Errorf("error retrieving node health by rpc for node %v: %v", node.ID(), err)
		}
		self.mu.Unlock()
		//if !(healthy.KnowNN && healthy.GotNN && healthy.Full) {
		if !healthy.KnowNN || !healthy.Full || healthy.CountNN < testMinProxBinSize {
			log.Debug("healthy not yet reached", "id", node.ID(), "attempt", i)
			i++
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
