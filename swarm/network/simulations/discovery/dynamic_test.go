package discovery

import (
	"bytes"
	"context"
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

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/pot"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/state"
)

const (
	bootNodeCount             = 3
	defaultHealthCheckRetries = 10
	defaultHealthCheckDelay   = time.Millisecond * 250
	defaultEventsTimeout      = time.Second
	defaultPruneInterval      = 5000000000 // 1 sec
	defaultMaxBinSize         = 3
)

var (
	mu              sync.Mutex
	bootNodes       []*discover.NodeID
	events          chan *p2p.PeerEvent
	ids             []discover.NodeID
	addrIdx         map[discover.NodeID][]byte
	rpcs            map[discover.NodeID]*rpc.Client
	dynamicServices adapters.Services
)

// Test to verify that restarted nodes will reach healthy state
//
// First, it brings up and connects bootnodes, and connects each of the further nodes
// to a random bootnode
//
// if network is healthy, it proceeds to stop and start a selection of nodes
// upon stop, it checks the health of the remaining nodes in the network that are up
// after starting, it performs new health checks after they have connected
// to one of their previous peers
func TestDynamicDiscovery(t *testing.T) {
	t.Run("32/4/sim", testDynamicDiscoveryRestarts)
}

func testDynamicDiscoveryRestarts(t *testing.T) {

	// this directory will keep the state store dbs
	dir, err := ioutil.TempDir("", "dynamic-discovery")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// discovery service
	dynamicServices := adapters.Services{
		"discovery": newDynamicServices(dir, false),
	}

	// set up locals we need
	paramstring := strings.Split(t.Name(), "/")
	nodeCount, _ := strconv.ParseInt(paramstring[1], 10, 0)
	numUpDowns, _ := strconv.ParseInt(paramstring[2], 10, 0)
	adapter := paramstring[3]

	if nodeCount < bootNodeCount {
		t.Fatalf("nodeCount must be bigger than bootnodeCount (%d < %d)", nodeCount, bootNodeCount)
	}

	bootNodes = make([]*discover.NodeID, bootNodeCount)
	events = make(chan *p2p.PeerEvent)
	ids = make([]discover.NodeID, nodeCount)
	addrIdx = make(map[discover.NodeID][]byte)
	rpcs = make(map[discover.NodeID]*rpc.Client)

	healthCheckDelay := defaultHealthCheckDelay

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
		if i < bootNodeCount {
			bootNodes[i] = &ids[i]
		}
		addrIdx[node.ID()] = network.ToOverlayAddr(node.ID().Bytes())
	}

	// sim step 1
	// start nodes, trigger them on node up event from sim
	trigger := make(chan discover.NodeID)
	events := make(chan *simulations.Event)
	sub := net.Events().Subscribe(events)
	defer sub.Unsubscribe()

	// quitC stops the event listener loops
	// inside the step action method after step is complete
	quitC := make(chan struct{})

	action := func(ctx context.Context) error {
		go func() {
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
				case <-ctx.Done():
					return
				case <-quitC:
					return
				}

			}
		}()
		go func() {
			for _, n := range ids {
				if err := net.Start(n); err != nil {
					t.Fatalf("error starting node: %s", err)
				}
				log.Info("network start returned", "node", n)

				// rpc client is only available after node start
				rpcs[n], err = net.GetNode(n).Client()
				if err != nil {
					t.Fatalf("error getting node rpc: %s", err)
				}
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

	timeout := 10 * time.Second
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

	// sim step 2
	// connect the three bootnodes together 1 -> 2 -> .. -> n
	// connect each of the other nodes to a random bootnode
	// triggers on connection event from sim
	close(quitC)
	quitC = make(chan struct{})
	action = func(ctx context.Context) error {
		go func(quitC chan struct{}) {
			for {
				select {
				case ev := <-events:
					if ev == nil {
						panic("got nil event")
					}
					if ev.Type == simulations.EventTypeConn {
						if ev.Conn.Up {
							log.Info(fmt.Sprintf("got conn up event %v", ev))
							trigger <- ev.Conn.One
						}
					}
				case <-ctx.Done():
					return
				case <-quitC:
					return
				}
			}
		}(quitC)
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

	timeout = 10 * time.Second
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
	sub.Unsubscribe()
	// sim step 3
	// now all nodes are up, all nodes are connected to network
	// so we check health of all nodes
	close(quitC)
	quitC = make(chan struct{})
	action = func(ctx context.Context) error {
		for _, n := range net.GetNodes() {
			go func(n *simulations.Node) {
				tick := time.NewTicker(healthCheckDelay)
				for {
					select {
					case <-tick.C:
						trigger <- n.ID()
					case <-ctx.Done():
						return
					case <-quitC:
						return
					}
				}
			}(n)
		}
		return nil
	}

	check = func(ctx context.Context, id discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		log.Info("health ok", "node", id)
		return checkHealth(net, id)
	}

	timeout = 5 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result = simulations.NewSimulation(net).Run(ctx, &simulations.Step{
		Action:  action,
		Trigger: trigger,
		Expect: &simulations.Expectation{
			Nodes: ids[:],
			Check: check,
		},
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	// sim step 4
	// bring the nodes up and down
	// if any health checks fail, the step will fail
	// check will be triggered when the first node up event is received
	close(quitC)
	quitC = make(chan struct{})
	victimSliceOffset := rand.Intn(len(ids) - int(numUpDowns) - 1)
	victimNodes := ids[victimSliceOffset : victimSliceOffset+int(numUpDowns)]

	events = make(chan *simulations.Event)
	sub = net.Events().Subscribe(events)
	defer sub.Unsubscribe()

	action = func(ctx context.Context) error {
		for _, nid := range victimNodes {
			stopC := make(chan struct{})
			go func(nid discover.NodeID, stopC chan struct{}) {
				var stopped bool
				var upped bool
				for {
					select {
					case ev := <-events:
						if ev == nil {
							panic("got nil event")
						} else if ev.Type == simulations.EventTypeNode {
							if ev.Node.Config.ID == nid {
								if ev.Node.Up && stopped && !upped {
									log.Info(fmt.Sprintf("got node up event %v", ev))
									// rpc client is changed upon new start, we need to get it anew
									mu.Lock()
									rpcs[nid], err = net.GetNode(nid).Client()
									mu.Unlock()
									if err != nil {
										t.Fatal(err)
									}
									upped = true
									trigger <- nid

								} else {
									log.Info(fmt.Sprintf("got node down event %v", ev))
									if !stopped {

										stopped = true
										close(stopC)
									}
								}
							}
						}
					case <-ctx.Done():
						return
					case <-quitC:
						return
					}
				}
			}(nid, stopC)

			// stop the node
			log.Info("restarting: stop", "node", nid, "addr", fmt.Sprintf("%x", addrIdx[nid]))
			err := net.Stop(nid)
			if err != nil {
				t.Fatal(err)
			}

			// wait for the stop event
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-quitC:
				return nil
			case <-stopC:
			}

			// check health of remaining nodes
		OUTER:
			for _, n := range net.GetNodes() {
				if !n.Up {
					if n.ID() != nid {

						log.Warn("a remaining node is down", "stoppednode", nid, "checknode", n.ID())
					}
					continue
				}
				tick := time.NewTicker(healthCheckDelay)
				for i := 0; ; i++ {
					select {
					case <-tick.C:
						log.Debug("health check other node after stop", "stoppednode", nid, "checknode", n.ID(), "attempt", i)
						ok, err := checkHealth(net, n.ID())
						if ok {
							log.Info("health ok other node after stop", "stoppednode", nid, "checknode", n.ID())
							continue OUTER
						} else if err != nil {
							return err
						}
					case <-ctx.Done():
						return fmt.Errorf("health not reached for node %s (addr %s) after stopped node %s (addr %s)", n.ID().TerminalString(), fmt.Sprintf("%x", addrIdx[n.ID()][:8]), nid.TerminalString(), fmt.Sprintf("%x", addrIdx[nid][:8]))
					case <-quitC:
						return nil
					}
				}
			}

			// bring the node back up
			log.Info("restarting: start", "node", nid, "addr", fmt.Sprintf("%x", addrIdx[nid]))
			err = net.Start(nid)
			if err != nil {
				t.Fatal(err)
			}

		}
		return nil
	}

	check = func(ctx context.Context, id discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		// health check for the restarted node
		tick := time.NewTicker(healthCheckDelay)
		for i := 0; ; i++ {
			log.Debug("health check after restart", "node", id, "attempt", i)
			select {
			case <-tick.C:
				ok, err := checkHealth(net, id)
				if ok {
					log.Info("health ok after restart", "node", id)
					return true, nil
				} else if err != nil {
					return false, err
				}
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
		return false, fmt.Errorf("health not reached for node %s (addr %s)", id.TerminalString(), fmt.Sprintf("%x", addrIdx[id][:8]))
	}

	timeout = 300 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result = simulations.NewSimulation(net).Run(ctx, &simulations.Step{
		Action:  action,
		Trigger: trigger,
		Expect: &simulations.Expectation{
			Nodes: victimNodes,
			Check: check,
		},
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	close(quitC)

	log.Warn("exiting dynamic test")
}

func TestDynamicPruning(t *testing.T) {
	t.Run("32/sim", testDynamicPruning)
}

func testDynamicPruning(t *testing.T) {
	// this directory will keep the state store dbs
	dir, err := ioutil.TempDir("", "dynamic-discovery")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// discovery service
	dynamicServices := adapters.Services{
		"discovery": newDynamicServices(dir, true),
	}

	// set up locals we need
	paramstring := strings.Split(t.Name(), "/")
	nodeCount, _ := strconv.ParseInt(paramstring[1], 10, 0)
	adapter := paramstring[2]

	bootNodes = make([]*discover.NodeID, bootNodeCount)
	events = make(chan *p2p.PeerEvent)
	ids = make([]discover.NodeID, nodeCount)
	addrIdx = make(map[discover.NodeID][]byte)
	rpcs = make(map[discover.NodeID]*rpc.Client)

	eventsTimeout := defaultEventsTimeout
	maxBinSize := defaultMaxBinSize

	log.Info("starting pruning test", "nodecount", nodeCount, "adaptertype", adapter)

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
		if i < bootNodeCount {
			bootNodes[i] = &ids[i]
		}
		addrIdx[node.ID()] = network.ToOverlayAddr(node.ID().Bytes())
	}

	// sim step 1
	// start nodes, trigger them on node up event from sim
	trigger := make(chan discover.NodeID)
	events := make(chan *simulations.Event)
	sub := net.Events().Subscribe(events)
	defer sub.Unsubscribe()
	// quitC stops the event listener loops
	// inside the step action method after step is complete
	quitC := make(chan struct{})

	action := func(ctx context.Context) error {
		go func() {
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
				case <-ctx.Done():
					return
				case <-quitC:
					return
				}

			}
		}()
		go func() {
			for _, n := range ids {
				if err := net.Start(n); err != nil {
					t.Fatalf("error starting node: %s", err)
				}
				log.Info("network start returned", "node", n)

				// rpc client is only available after node start
				rpcs[n], err = net.GetNode(n).Client()
				if err != nil {
					t.Fatalf("error getting node rpc: %s", err)
				}
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

	timeout := 10 * time.Second
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

	// sim step 2
	// connect the three bootnodes together 1 -> 2 -> .. -> n
	// connect each of the other nodes to a random bootnode
	// triggers on connection event from sim
	close(quitC)
	quitC = make(chan struct{})
	action = func(ctx context.Context) error {
		go func(quitC chan struct{}) {
			for {
				var lastEvent time.Time
				tick := time.NewTicker(eventsTimeout)
				select {
				case ev := <-events:
					if ev.Type == simulations.EventTypeConn {
						lastEvent = time.Now()
					}
				case <-tick.C:
					if time.Since(lastEvent) > eventsTimeout {
						trigger <- ids[0]
					}
				case <-ctx.Done():
					return
				case <-quitC:
					return
				}
			}
		}(quitC)
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
		return true, nil
	}

	timeout = 20 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result = simulations.NewSimulation(net).Run(ctx, &simulations.Step{
		Action:  action,
		Trigger: trigger,
		Expect: &simulations.Expectation{
			Nodes: ids[:1],
			Check: check,
		},
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	sub.Unsubscribe()
	close(quitC)

	// sort all network events according to po, and appended chronologically
	pof := pot.DefaultPof(256)
	eventMap := make(map[discover.NodeID]map[int][][]byte)
	for _, n := range net.GetNodes() {
		eventMap[n.ID()] = make(map[int][][]byte)
	}
	for _, ev := range result.NetworkEvents {
		if ev.Type == simulations.EventTypeConn {
			if ev.Conn.Up {
				log.Debug("Processing event", "time", ev.Time, "one", ev.Conn.One.TerminalString(), "other", ev.Conn.Other.TerminalString())
				po, _ := pof(addrIdx[ev.Conn.One], addrIdx[ev.Conn.Other], 0)
				eventMap[ev.Conn.One][po] = append(eventMap[ev.Conn.One][po], addrIdx[ev.Conn.Other])
				eventMap[ev.Conn.Other][po] = append(eventMap[ev.Conn.Other][po], addrIdx[ev.Conn.One])
			}
		}
	}

	// slice the newest MaxProxBinSize entries from the events
	// the latest should be the ones in the pruned kademlia
	for k, bin := range eventMap {
		for i, addrs := range bin {
			if len(addrs) > maxBinSize {
				addrs = addrs[len(addrs)-maxBinSize:]
			}
			log.Trace("sliced", "po", 255-i, "node", k.TerminalString(), "addr", network.LogAddrs(addrs))
		}
	}

	time.Sleep(5 * time.Second)

	// get current connections
	// (they should now be pruned)
	connMap := make(map[discover.NodeID][]discover.NodeID)
	potMap := make(map[discover.NodeID]*pot.Pot)
	for _, n := range net.GetNodes() {
		potMap[n.ID()] = pot.NewPot(addrIdx[n.ID()], 0)
	}

	// map all connections to individual nodes
	for _, conn := range net.Conns {
		if conn.Up {
			connMap[conn.One] = append(connMap[conn.One], conn.Other)
			connMap[conn.Other] = append(connMap[conn.Other], conn.One)
			potMap[conn.One], _, _ = pot.Add(potMap[conn.One], addrIdx[conn.Other], pof)
			potMap[conn.Other], _, _ = pot.Add(potMap[conn.Other], addrIdx[conn.One], pof)
		}
	}

	for _, n := range net.GetNodes() {
		var bins [256]int
		potMap[n.ID()].EachFrom(func(v pot.Val, po int) bool {

			log.Debug(fmt.Sprintf("checking peer %08x against node %s", v.([]byte), n.ID().TerminalString()))
			if bins[po] == maxBinSize {
				t.Fatalf("node %s (addr %08x) has more than maxBinSize peers", n.ID().TerminalString(), addrIdx[n.ID()])
			}
			var matchPeer bool
			for _, p := range eventMap[n.ID()][po] {
				if bytes.Equal(v.([]byte), p) {
					matchPeer = true
				}
			}
			if !matchPeer {
				t.Fatalf("node %s (addr %08x) is missing most recent peer %08x", n.ID().TerminalString(), addrIdx[n.ID()], v.([]byte))
			}
			bins[po]++
			log.Info(fmt.Sprintf("%s [%d] -> %x", n.ID().TerminalString(), po, v.([]byte)[:8]))
			return true
		}, 0)

	}
	log.Info("exit pruning test")
}

func randomDelay(maxDuration int) time.Duration {
	if maxDuration == 0 {
		maxDuration = 1000000000
	}
	timeout := rand.Intn(maxDuration) + 10000000
	ns := fmt.Sprintf("%dns", timeout)
	dur, _ := time.ParseDuration(ns)
	return dur
}

func checkHealth(net *simulations.Network, id discover.NodeID) (bool, error) {
	healthy := &network.Health{}

	var upAddrs [][]byte
	for _, n := range net.GetNodes() {
		if n.Up {
			upAddrs = append(upAddrs, addrIdx[n.ID()])
		}
	}
	log.Debug("generating new peerpotmap", "node", id, "addr", fmt.Sprintf("%x", addrIdx[id]))
	hotPot := network.NewPeerPotMap(testMinProxBinSize, upAddrs)
	addrHex := fmt.Sprintf("%x", addrIdx[id])

	if _, ok := hotPot[addrHex]; !ok {
		log.Debug("missing pot", "node", id, "addr", fmt.Sprintf("%x", addrHex[:8]))
		return false, nil
	}
	if err := rpcs[id].Call(&healthy, "hive_healthy", hotPot[addrHex]); err != nil {
		return false, fmt.Errorf("error retrieving node health by rpc for node %v: %v", id, err)
	}
	if !(healthy.KnowNN && healthy.GotNN && healthy.Full) {
		log.Debug(fmt.Sprintf("healthy not yet reached\n%s", healthy.Hive), "id", id, "addr", fmt.Sprintf("%x", addrIdx[id][:8]), "missing", network.LogAddrs(healthy.CulpritsNN), "knowNN", healthy.KnowNN, "gotNN", healthy.GotNN, "countNN", healthy.CountNN, "full", healthy.Full)
		return false, nil
	}
	return true, nil

}

func newDynamicServices(storePath string, pruning bool) func(*adapters.ServiceContext) (node.Service, error) {
	return func(ctx *adapters.ServiceContext) (node.Service, error) {
		host := adapters.ExternalIP()

		addr := network.NewAddrFromNodeIDAndPort(ctx.Config.ID, host, ctx.Config.Port)

		kp := network.NewKadParams()
		kp.MinProxBinSize = testMinProxBinSize
		kp.MaxBinSize = defaultMaxBinSize
		kp.MinBinSize = 1
		kp.MaxRetries = 10
		kp.RetryExponent = 2
		kp.RetryInterval = 50000000
		if pruning {
			kp.PruneInterval = defaultPruneInterval
		}
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

		stateStore, err := state.NewDBStore(filepath.Join(storePath, fmt.Sprintf("state-store-%s.db", ctx.Config.ID)))
		if err != nil {
			return nil, err
		}
		return network.NewBzz(config, kad, stateStore, nil, nil), nil
	}
}
