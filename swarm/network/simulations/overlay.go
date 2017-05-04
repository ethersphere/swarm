// +build none

// You can run this simulation using
//
//    go run ./swarm/network/simulations/overlay.go
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
)

// Network extends simulations.Network with hives for each node.
type SimNetwork struct {
	*simulations.Network
  hives map[discover.NodeID]*network.Hive
}

// SimNode is the adapter used by Swarm simulations.
type SimNode struct {
	hive     *network.Hive
	protocol *p2p.Protocol
}

func (s *SimNode) Protocols() []p2p.Protocol {
	return []p2p.Protocol{*s.protocol}
}

func (s *SimNode) APIs() []rpc.API {
	return nil
}

func (s *SimNode) Addr() []byte {
	return nil
}

// the hive update ticker for hive
func af() <-chan time.Time {
	return time.NewTicker(1 * time.Second).C
}

// Start() starts up the hive
// makes SimNode implement node.Service
func (self *SimNode) Start(server p2p.Server) error {
	return self.hive.Start(server, af)
}

// Stop() shuts down the hive
// makes SimNode implement node.Service
func (self *SimNode) Stop() error {
	self.hive.Stop()
	return nil
}

func (self *SimNode) Info() string {
  return self.hive.String()
}

// NewSimNode creates adapters for nodes in the simulation.
func (self *SimNetwork) NewSimNode(conf *simulations.NodeConfig) adapters.NodeAdapter {
	id := conf.Id
	addr := network.NewPeerAddrFromNodeId(id)
	kp := network.NewKadParams()

	kp.MinProxBinSize = 2
	kp.MaxBinSize = 3
	kp.MinBinSize = 1
	kp.MaxRetries = 1000
	kp.RetryExponent = 2
	kp.RetryInterval = 1000000

	to := network.NewKademlia(addr.OverlayAddr(), kp) // overlay topology driver
	hp := network.NewHiveParams()
	hp.CallInterval = 5000
	pp := network.NewHive(hp, to) // hive
	self.hives[id.NodeID] = pp    // remember hive

	services := func(p network.Peer) error {
		dp := network.NewDiscovery(p, to)
		pp.Add(dp)
		log.Trace(fmt.Sprintf("kademlia on %v", dp))
		p.DisconnectHook(func(err error) {
			pp.Remove(dp)
		})
		return nil
	}

	ct := network.BzzCodeMap(network.DiscoveryMsgs...) // bzz protocol code map

	node := &SimNode{
		hive:     pp,
		protocol: network.Bzz(addr.OverlayAddr(), addr.UnderlayAddr(), ct, services, nil, nil),
	}
	return adapters.NewSimNode(id, node, self.Network)
}


func (self *SimNetwork) simSetup(ids []*adapters.NodeId) {
	for i, id := range ids {
		var peerId *adapters.NodeId
		if i == 0 {
			peerId = ids[i+1]
		} else {
			peerId = ids[i-1]
		}
		err := self.hives[id.NodeID].Register(network.NewPeerAddrFromNodeId(peerId))
		if err != nil {
			panic(err.Error())
		}
	}
}

func SetupNewNetwork() *simulations.Network {
  conf := createNetworkConfig()
  sim := &SimNetwork{
    Network: simulations.NewNetwork(conf),
    hives:   make(map[discover.NodeID]*network.Hive),
  }
	sim.Network.SetNaf(sim.NewSimNode)
  sim.Network.SetInit(sim.simSetup)
  return sim.Network
}


func createNetworkConfig() *simulations.NetworkConfig {
	//setup conf
  conf := &simulations.NetworkConfig{}
	conf.DefaultMockerConfig = simulations.DefaultMockerConfig()
	conf.DefaultMockerConfig.NodeCount      = 10
	conf.DefaultMockerConfig.SwitchonRate   = 100
	conf.DefaultMockerConfig.NodesTarget    = 15
	conf.DefaultMockerConfig.NewConnCount   = 1
	conf.DefaultMockerConfig.DegreeTarget   = 0
	conf.DefaultMockerConfig.UpdateInterval = 2000
	conf.Id = "0"
	conf.Backend = true

  return conf
}

// var server
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	net := SetupNewNetwork()
	c, quitc := simulations.RunDefaultNet(net)

	simulations.StartRestApiServer("8888", c)
	// wait until server shuts down
	<-quitc

}
