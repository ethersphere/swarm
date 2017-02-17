package simulations

import (
	"fmt"
	"reflect"
	"sync"
	"bytes"
	"encoding/binary"
	"strconv"

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

type NodeTmpSendSimpleMsgIF struct {
	One uint
	Other uint
	Uuid uint64
	Protocol string
	Command uint8
	Payload []METAPayloadIF
}

type METANameRegisterIF struct {
	Squealernode uint
	Victimnode string
	Name string
	Swarmhash storage.Key
}

type METAPayloadIF struct {
	Type uint8
	Label string
	Numeric bool
	Data string
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
		Ct: METAnetwork.NewMETACodeMap(&METAnetwork.METAAnnounce{}, &METAnetwork.METATmpName{}),
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
									var nodeid *adapters.NodeId
									
									nodeid = p2ptest.RandomNodeId()
									
									ppnet.NewNode(&p2psimulations.NodeConfig{Id: nodeid})
									glog.V(6).Infof("added node %v to network %v", nodeid, ppnet)
									
									//return &p2psimulations.NodeConfig{Id: nodeid}, nil
									return &struct{
										Id *adapters.NodeId
										Index int
									}{
										Id: nodeid,
										Index: len(ppnet.Nodes),
									}, nil
								},
							},
							Retrieve: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									return &NodeResult{Nodes: ppnet.Nodes}, nil
								},
							},
							Update: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									var othernode *p2psimulations.Node
									
									args := msg.(*NodeIF)
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
									
									if ok {							
										if storage.IsZeroKey(args.Swarmhash) { // inputcheck
											glog.V(6).Infof("Name/swarm update update needs swarm hash. dude")
											return &struct{}{}, nil
										}
										protomsg := METAnetwork.NewMETATmpName()
										protomsg.Swarmhash = args.Swarmhash
										protomsg.Name = args.Name
										protomsg.Node = *adapters.NewNodeIdFromHex(args.Victimnode)
										// update local registry
										METAnetwork.METATmpSwarmRegistryLookup[protomsg.Name] = [2]string{fmt.Sprintf("%v",protomsg.Node), fmt.Sprintf("%v",protomsg.Swarmhash)}
										// then pass on to the others
										ppnet.Broadcast(onenode.Id, protomsg)
										
										return &struct{}{}, nil
										
										
									}
									
									
									return &struct{}{}, nil
								},
								Type: reflect.TypeOf(&METANameRegisterIF{}), 
							},
							Retrieve: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
									
									list := []struct{Name string
										Node string
										Swarmhash string}{} // making manual list because serialization of response from map doesn't seem to be implemented
									
									for k, v := range METAnetwork.METATmpSwarmRegistryLookup {
										entry := struct{Name string
											Node string
											Swarmhash string
										}{
											Name: k,
											Node: v[0],
											Swarmhash: v[1],
										}
										list = append(list, entry)
									}
									
									
									return list, nil
									
								}, 
								Type: reflect.TypeOf(&METANameListIF{}), 
							},
						},
					))
					
					nodecontroller.SetResource("msg", p2psimulations.NewResourceContoller (
						&p2psimulations.ResourceHandlers{
							Create: &p2psimulations.ResourceHandler{
								Handle: func(msg interface{}, parent *p2psimulations.ResourceController) (interface{}, error) {
								
									args, ok := msg.(*NodeTmpSendSimpleMsgIF)
									
									if ok {
										onenode := ppnet.Nodes[args.One - 1]
										othernode := ppnet.Nodes[args.Other - 1]
										
										switch args.Protocol {
											case "METAAnnounce":
																						
												announce := METAnetwork.NewMETAAnnounce()
												announce.SetCommand(args.Command)
												announce.SetUuid(args.Uuid)
												
												for _,pl := range args.Payload {
													var data []byte
													if pl.Numeric {
														data = make([]byte, 8)
														i,_ := strconv.ParseInt(pl.Data, 10, 8)
														glog.V(6).Infof("data is numeric ParseInt '%s' => %v", pl.Data, i)
														binary.PutVarint(data, int64(i))
													} else {
														data = bytes.NewBufferString(pl.Data).Bytes()
													}
													
													announce.AddPayload(pl.Type) 
													announce.Payload[0].Add(pl.Label, data)
												}
												
												ppnet.Send(onenode.Id, othernode.Id, 0, announce) //GetNode(onenode.Id).na.(*adapters.SimNode).GetPeer(othernode.Id).SendMsg(0, protomsg)
										}
									}
									
									return &struct{}{}, nil
								},
								Type: reflect.TypeOf(&NodeTmpSendSimpleMsgIF{}), 
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
