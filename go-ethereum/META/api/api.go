package api

import (
	"fmt"
	"github.com/ethereum/go-ethereum/META/network"
	//"github.com/ethereum/go-ethereum/p2p/protocols"
	
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

type Api struct {
	C string //for now nothing here
}

func NewApi() (self *Api) {
	self = &Api{
		C: "abcdef",
	}
	return
}

type ApiErrorSimple struct {
	Errorstring string
	Errorstate bool
}

// serialisable info about META
type Info struct {
	*Config
}

func (i *Info) Infoo() interface{} {
	return i
}

func NewInfo(c *Config) *Info {
	return &Info{c}
}

type ParrotNode struct {
	peers *network.PeerCollection
}

func (p *ParrotNode) Hellofirstnode(msg string) interface{} {
	glog.V(logger.Debug).Infof("hellofirstnode is %v got %v", p, msg)
	
	if len(p.peers.Peers) == 0 {
		return &ApiErrorSimple{Errorstate: true, Errorstring: "no peers"}
	}
	//protomsg := &network.Hellofirstnodemsg{C: &HelloFirstNodeReply{Message: msg, Code: fmt.Sprintf("%x",p.peers)}}
	protomsg := &network.Hellofirstnodemsg{Pmsg: msg, Sub: *p.peers.Peers[0]}
	
	err := p.peers.Peers[0].Send(protomsg)
	if err != nil {
		return &ApiErrorSimple{Errorstate: true, Errorstring: fmt.Sprintf("couldnt send %v to %v", protomsg, p.peers.Peers[0])}
	}
	
	return &ApiErrorSimple{Errorstate: false, Errorstring: fmt.Sprintf("sent '%v' to %v!", protomsg, p.peers.Peers[0])}
}

func NewParrotNode(self *network.PeerCollection) *ParrotNode {
	return &ParrotNode{peers: self}
	
}

