package simulations

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/event"
	//"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	p2psimulations "github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/swarm/storage"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
	METAnetwork "github.com/ethereum/go-ethereum/META/network"
)

func init() {
	glog.SetV(6)
	glog.SetToStderr(true)
}

// override networkconfig to accept args in sessioncontroller invoke
type NetworkConfig struct {
	*p2psimulations.NetworkConfig
}

type NetworkList struct {
	Current *Network
	Available []*Network
}

type NodeResult struct {
	Nodes []*p2psimulations.Node	
}

type NodeIF struct {
	One uint
	Other uint
}

type METANameRegisterIF struct {
	Squealernode uint
	Victimnode string
	Name string
	Swarmhash storage.Key
}

type METANameListIF struct {
	Reverse bool
}

type Network struct {
	*p2psimulations.Network
	messenger func(p2p.MsgReadWriter) adapters.Messenger
	Id string
	Ct *protocols.CodeMap
	Peers map[*adapters.NodeId]*METAnetwork.PeerCollection
}

func (self *Network) String() string {
	return self.Id
}

func (self *Network) NewSimNode(conf *p2psimulations.NodeConfig) adapters.NodeAdapter {
	wg := sync.WaitGroup{}
	id := conf.Id
	na := adapters.NewSimNode(id, self.Network, self.messenger)
	pp := METAnetwork.NewPeerCollection()
	na.Run = METAnetwork.METAProtocol(pp, self.Ct, na, &wg).Run
	self.Peers[na.Id] = pp
	return na
}

func NewNetwork(network *p2psimulations.Network, messenger func(p2p.MsgReadWriter) adapters.Messenger, id string) *Network {
	n := &Network{
		// hives:
		Network:   network,
		messenger: messenger,
		Id: id,
		Ct: METAnetwork.NewMETACodeMap(&METAnetwork.METATmpName{}),
		Peers: make(map[*adapters.NodeId]*METAnetwork.PeerCollection),
	}
	n.SetNaf(n.NewSimNode)
	return n
}

func (self *Network) Broadcast(senderid *adapters.NodeId, protomsg interface{}) {
	
	msg := &p2psimulations.Msg{
		One:   senderid,
		Code:  12345, // TODO get this from lookup msg structure somehow
	}

	for _,peer := range self.Peers[senderid].Peers {
		peer.Send(protomsg)
		a:= &adapters.NodeId{
			peer.ID(),
		}
		msg.Other = a
		//self.Network.events.Post(msg.event())         
	}
	//self.GetNode(senderid).na.(*adapters.SimNode)   //.GetPeer(receiverid).SendMsg(msgcode, protomsg) // phew!
	
}

func NewSessionController() (*p2psimulations.ResourceController, chan bool) {
	networks := &NetworkList{}
	quitc := make(chan bool)
	return p2psimulations.NewResourceContoller(
		&p2psimulations.ResourceHandlers{
			// POST /
			Create: &p2psimulations.ResourceHandler{
				Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
					conf := msg.(*NetworkConfig)
					messenger := adapters.NewSimPipe
					net := p2psimulations.NewNetwork(nil, &event.TypeMux{})
					ppnet := NewNetwork(net, messenger, conf.Id)
					journal := p2psimulations.NewJournal()
					c := p2psimulations.NewNetworkController(conf.NetworkConfig, net.Events(), journal)
					if len(conf.Id) == 0 {
						conf.Id = fmt.Sprintf("%d", 0)
					}
					glog.V(6).Infof("new network controller on %v", conf.Id)
					if parent != nil {
						parent.SetResource(conf.Id, c)
					}
					networks.Available = append(networks.Available, ppnet)
					networks.Current = ppnet
					
					c.SetResource("debug", p2psimulations.NewResourceContoller(
						&p2psimulations.ResourceHandlers{
							Create: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									journaldump := []string{}
									eventfmt := func(e *event.Event) bool {
										journaldump = append(journaldump, fmt.Sprintf("%v", e))
										return true
									}
									journal.Read(eventfmt)
									return struct{Results []string}{Results: journaldump,}, nil
								},
							},
						},
					))
					
					nodecontroller := p2psimulations.NewResourceContoller (
						&p2psimulations.ResourceHandlers{
							Create: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									//var networks *NetworkList
									var nodeid *adapters.NodeId
									
									//t_network, _ := c.Retrieve.Handle(nil, nil) // parent is the sessioncontroller
									//networks = t_network.(*NetworkList)
									nodeid = p2ptest.RandomNodeId()
									
									//networks.Current.NewNode(&p2psimulations.NodeConfig{Id: nodeid})
									ppnet.NewNode(&p2psimulations.NodeConfig{Id: nodeid})
									glog.V(6).Infof("added node %v to network %v", nodeid, ppnet)
									
									return &p2psimulations.NodeConfig{Id: nodeid}, nil
									
								},
							},
							Retrieve: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									//var networks *NetworkList
									//t_network, _ := c.Retrieve.Handle(nil, nil) // parent is the sessioncontroller
									//networks = t_network.(*NetworkList)
									
									//return &NodeResult{Nodes: networks.Current.Nodes}, nil
									return &NodeResult{Nodes: ppnet.Nodes}, nil
								},
							},
							Update: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									//var networks *NetworkList
									var othernode *p2psimulations.Node
									//t_network, _ := c.Retrieve.Handle(nil, nil) // parent is the sessioncontroller
									//networks = t_network.(*NetworkList)
									
									args := msg.(*NodeIF)
									//onenode := networks.Current.Nodes[args.One - 1]
									onenode := ppnet.Nodes[args.One - 1]
									
									if args.Other == 0 {
										if ppnet.Start(onenode.Id) != nil {
											ppnet.Stop(onenode.Id)	
										}
										return &NodeResult{Nodes: []*p2psimulations.Node{onenode}}, nil
									} else {
										othernode = ppnet.Nodes[args.Other - 1]
										ppnet.Connect(onenode.Id, othernode.Id)
										return &NodeResult{Nodes: []*p2psimulations.Node{onenode, othernode}}, nil	
									} 
									
									return struct{}{}, nil
								},
								Type: reflect.TypeOf(&NodeIF{}), // this is input not output param structure
							},
						},
					)
					
					c.SetResource("node", nodecontroller)
					
					nodecontroller.SetResource("tmpname", p2psimulations.NewResourceContoller (
						&p2psimulations.ResourceHandlers{
							Create: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									
									args,ok := msg.(*METANameRegisterIF)
									onenode := ppnet.Nodes[args.Squealernode - 1]
									//othernode := ppnet.Nodes[1]
									
									if ok {							
										if storage.IsZeroKey(args.Swarmhash) { // inputcheck
											glog.V(6).Infof("Name/swarm update update needs swarm hash. dude")
											return &struct{}{}, nil
										}
										//squealer := ppnet.Nodes[args.Squealernode - 1]
										protomsg := METAnetwork.NewMETATmpName()
										protomsg.Swarmhash = args.Swarmhash
										protomsg.Name = args.Name
										protomsg.Node = *adapters.NewNodeIdFromHex(args.Victimnode)
										//protomsg.Node = othernode.Id
										glog.V(6).Infof("Broadcasting update: %v", protomsg)
										ppnet.Broadcast(onenode.Id, protomsg)
									}
									return &struct{}{}, nil
								},
								Type: reflect.TypeOf(&METANameRegisterIF{}), 
							},
							Retrieve: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									
									list := &struct{Names [][2]string}{} // making manual list because serialization of response from map doesn't seem to be implemented
									args,ok := msg.(*METANameListIF)
									
									if ok {
										if args.Reverse {
											//return &struct{List map[*storage.Key]*adapters.NodeId}{List: METAnetwork.METATmpSwarmRegistryLookupReverse}, nil
											for k, v := range METAnetwork.METATmpSwarmRegistryLookupReverse {
												list.Names = append(list.Names, [2]string{fmt.Sprintf("%v",k),fmt.Sprintf("%v",v)})
											}
											return list, nil
										}
									}
									
									//return &struct{List map[*adapters.NodeId]*storage.Key}{List: METAnetwork.METATmpSwarmRegistryLookup}, nil
									for k, v := range METAnetwork.METATmpSwarmRegistryLookup {
										list.Names = append(list.Names, [2]string{fmt.Sprintf("%v",k),fmt.Sprintf("%v",v)})
									}
									
									return list, nil
									
								}, 
								Type: reflect.TypeOf(&METANameListIF{}), 
							},
						},
					))
					
					return struct{}{}, nil
				},
				Type: reflect.TypeOf(&NetworkConfig{}),
			},
			// GET /
			Retrieve: &p2psimulations.ResourceHandler{
				Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
					return networks, nil
				},
			},
			// DELETE /
			Destroy: &p2psimulations.ResourceHandler{
				Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
					glog.V(6).Infof("destroy handler called")
					// this can quit the entire app (shut down the backend server)
					quitc <- true
					return struct{}{}, nil
				},
			},
		},
	), quitc
}
