// +build none

// You can run this simulation using
//
//    go run ./swarm/network/simulations/overlay.go
package main

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	//"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
	METAnetwork "github.com/ethereum/go-ethereum/META/network"
)

// override networkconfig to accept args in sessioncontroller invoke
type NetworkConfig struct {
	*simulations.NetworkConfig
}

type NetworkList struct {
	Current *Network
	Available []*Network
}

type NodeResult struct {
	Nodes []*simulations.Node	
}

type NodeIF struct {
	One uint
	Other uint
	MessageType uint8
}

// Network extends simulations.Network with hives for each node.
type Network struct {
	*simulations.Network
	protopeers *METAnetwork.PeerCollection
	messenger func(p2p.MsgReadWriter) adapters.Messenger
	Id string
}

func (self *Network) String() string {
	return self.Id
}

// SimNode is the adapter used by Swarm simulations.
type SimNode struct {
	adapters.NodeAdapter
	protopeers METAnetwork.PeerCollection 
}

// NewSimNode creates adapters for nodes in the simulation.
func (self *Network) NewSimNode(conf *simulations.NodeConfig) adapters.NodeAdapter {
	wg := sync.WaitGroup{}
	id := conf.Id
	na := adapters.NewSimNode(id, self.Network, self.messenger)
	na.Run = METAnetwork.METAProtocol(self.protopeers, na, &wg).Run
	return na
}

func NewNetwork(network *simulations.Network, messenger func(p2p.MsgReadWriter) adapters.Messenger, id string) *Network {
	n := &Network{
		// hives:
		Network:   network,
		messenger: messenger,
		Id: id,
	}
	n.protopeers = METAnetwork.NewPeerCollection()
	n.SetNaf(n.NewSimNode)
	return n
}

// NewSessionController sits as the top-most controller for this simulation
// creates an inprocess simulation of basic node running their own bzz+hive
func NewSessionController() (*simulations.ResourceController, chan bool) {
	networks := &NetworkList{}
	quitc := make(chan bool)
	return simulations.NewResourceContoller(
		&simulations.ResourceHandlers{
			// POST /
			Create: &simulations.ResourceHandler{
				Handle: func(msg interface{}, parent *simulations.ResourceController) (interface{}, error) {
					conf := msg.(*NetworkConfig)
					messenger := adapters.NewSimPipe
					net := simulations.NewNetwork(nil, &event.TypeMux{})
					ppnet := NewNetwork(net, messenger, conf.Id)
					c := simulations.NewNetworkController(conf.NetworkConfig, net.Events(), simulations.NewJournal())
					if len(conf.Id) == 0 {
						conf.Id = fmt.Sprintf("%d", 0)
					}
					glog.V(6).Infof("new network controller on %v", conf.Id)
					if parent != nil {
						parent.SetResource(conf.Id, c)
					}
					networks.Available = append(networks.Available, ppnet)
					networks.Current = ppnet
					return struct{}{}, nil
				},
				Type: reflect.TypeOf(&NetworkConfig{}),
			},
			// GET /
			Retrieve: &simulations.ResourceHandler{
				Handle: func(msg interface{}, parent *simulations.ResourceController) (interface{}, error) {
					return networks, nil
				},
				Type: reflect.TypeOf(&NetworkList{}),
			},
			// DELETE /
			Destroy: &simulations.ResourceHandler{
				Handle: func(msg interface{}, parent *simulations.ResourceController) (interface{}, error) {
					glog.V(6).Infof("destroy handler called")
					// this can quit the entire app (shut down the backend server)
					quitc <- true
					return struct{}{}, nil
				},
			},
		},
	), quitc
}

// var server
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	glog.SetV(6)
	glog.SetToStderr(true)

	c, quitc := NewSessionController()
	
	// this needs to be moved to sub of networkcontroller
	c.SetResource("node", simulations.NewResourceContoller(
		&simulations.ResourceHandlers{
			Create: &simulations.ResourceHandler{
				Handle: func(msg interface{}, parent *simulations.ResourceController) (interface{}, error) {
					var networks *NetworkList
					var nodeid *adapters.NodeId
					
					t_network, _ := c.Retrieve.Handle(nil, nil) // parent is the sessioncontroller
					networks = t_network.(*NetworkList)
					nodeid = p2ptest.RandomNodeId()
					
					networks.Current.NewNode(&simulations.NodeConfig{Id: nodeid})
					glog.V(6).Infof("added node %v to network %v", nodeid, networks.Current)
					
					return &simulations.NodeConfig{Id: nodeid}, nil
					
				},
				Type: reflect.TypeOf(&simulations.NodeConfig{}),
			},
			Retrieve: &simulations.ResourceHandler{
				Handle: func(msg interface{}, parent *simulations.ResourceController) (interface{}, error) {
					var networks *NetworkList
					t_network, _ := c.Retrieve.Handle(nil, nil) // parent is the sessioncontroller
					networks = t_network.(*NetworkList)
					
					//return &NodeResult{Nodes: networks.Current.Nodes}, nil
					return &NodeResult{Nodes: networks.Current.Nodes}, nil
				},
				Type: reflect.TypeOf(&NodeIF{}), // this is input not output param structure
			},
			Update: &simulations.ResourceHandler{
				Handle: func(msg interface{}, parent *simulations.ResourceController) (interface{}, error) {
					var networks *NetworkList
					var othernode *simulations.Node
					t_network, _ := c.Retrieve.Handle(nil, nil) // parent is the sessioncontroller
					networks = t_network.(*NetworkList)
					
					args := msg.(*NodeIF)
					onenode := networks.Current.Nodes[args.One - 1]
					
					if args.Other == 0 {
						if networks.Current.Start(onenode.Id) != nil {
							networks.Current.Stop(onenode.Id)	
						}
						return &NodeResult{Nodes: []*simulations.Node{onenode}}, nil
						//return &struct{Result string}{Result: fmt.Sprintf("%v", onenode)}, nil
					} else {
						othernode = networks.Current.Nodes[args.Other - 1]
						if (args.MessageType == 0) {
							networks.Current.Connect(onenode.Id, othernode.Id)
							return &NodeResult{Nodes: []*simulations.Node{onenode, othernode}}, nil
							//return &struct{Result string}{Result: fmt.Sprintf("%v -> %v", othernode)}, nil
						} else {
							return &struct{}{}, nil // should have sent protocol message to peer, but don't know how to yet
						}
					}
				},
				Type: reflect.TypeOf(&NodeIF{}), // this is input not output param structure
			},
		},
	))
	simulations.StartRestApiServer("8888", c)
	// wait until server shuts down
	<-quitc

}
