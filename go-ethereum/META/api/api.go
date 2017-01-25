package api

import (
	"fmt"
	"time"
	"math/rand"
	
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

type peerNotAvailableError struct {
}
func (e *peerNotAvailableError) ErrorCode() int { return 0xFF010001 }
func (e *peerNotAvailableError) Error() string { return "No available peers" }

type sendToPeerError struct {
	details string
}
func (e *sendToPeerError) ErrorCode() int { return 0xFF010002 }
func (e *sendToPeerError) Error() string { return fmt.Sprintf("Could not sent to peer: %v", e.details) }

type peerIndexError struct {
	details string
}
func (e *peerIndexError) ErrorCode() int { return 0xFF010003 }
func (e *peerIndexError) Error() string { return fmt.Sprintf("No peer on that index") }

// serialisable info about META
type Info struct {
	*Config
}

func (i *Info) Infoo() (interface{}, error) {
	return i, nil
}

func NewInfo(c *Config) *Info {
	return &Info{c}
}

type ParrotNode struct {
	peers *network.PeerCollection
	consolechan chan string
}

func (self *ParrotNode) Hellonode(peerindex int, msg string) (interface{}, error) {
	if peerindex > len(self.peers.Peers) {
		return nil, &peerIndexError{}
	}
	
	if len(self.peers.Peers) == 0 {
		return nil, &peerNotAvailableError{}
	}

	protomsg := &network.Hellofirstnodemsg{Pmsg: msg, Now: time.Now()}
	
	err := self.peers.Peers[peerindex - 1].Send(protomsg)
	if err != nil {
		return nil, &sendToPeerError{details: err.Error()}
	}
	
	peerresponse := <-self.consolechan
	
	return fmt.Sprintf("sent '%v' to %v, response was %v", protomsg, self.peers.Peers[peerindex - 1], peerresponse), nil
}

func NewParrotNode(self *network.PeerCollection, l_consolechan chan string) *ParrotNode {
	return &ParrotNode{peers: self, consolechan: l_consolechan}
	
}

type ParrotCrowd struct {
	peers *network.PeerCollection
	consolechan chan string
}

/***
 *
 * \todo Move the channel down the stack and create a loop listener to output to console
 *  
 */
 
func (self *ParrotCrowd) Helloallnode(msg string) (interface{}, error) {
	
	rand.Seed(int64(time.Now().Nanosecond()))
	
	glog.V(logger.Debug).Infof("hellofirstnode is %v got %v", self, msg)
	
	if len(self.peers.Peers) == 0 {
		return nil, &peerNotAvailableError{}
	}

	protomsg := &network.Helloallnodemsg{Pmsg: fmt.Sprintf("%v + rnd %v", msg, rand.Int())}
	
	for _,p := range self.peers.Peers {
		err := p.Send(protomsg)
		if err != nil {
			return nil, &sendToPeerError{details: err.Error()}
		}
	}
	
	peerresponse,ok := <-self.consolechan
	if ok == false {
		glog.V(logger.Error).Infof("consolechan is closed!")
	}
	
	return fmt.Sprintf("sent '%v' to %v, response was %v", protomsg, self.peers.Peers[0], peerresponse), nil
}

func NewParrotCrowd(self *network.PeerCollection, l_consolechan chan string) *ParrotCrowd {
	return &ParrotCrowd{peers: self, consolechan: l_consolechan}
}

type PeerBroadcastSwitch struct {
	peers *network.PeerCollection
}

func (self *PeerBroadcastSwitch) Peerbroadcast(peerindex int, value bool) (interface{}, error) {
	if peerindex > len(self.peers.Peers) {
		return nil, &peerIndexError{}
	}
	if len(self.peers.Peers) == 0 {
		return nil, &peerNotAvailableError{}
	}
	self.peers.Peers[peerindex - 1].Answersbroadcast = value
	return "ok", nil
}

func NewPeerBroadcastSwitch(self *network.PeerCollection) *PeerBroadcastSwitch {
	return &PeerBroadcastSwitch{peers: self}
}

type WhoAreYou struct {
	peers *network.PeerCollection
}

func (self *WhoAreYou) Whoareyou(peerindex int) (interface{}, error) {
	if peerindex > len(self.peers.Peers) {
		return nil, &peerIndexError{}
	}
	if len(self.peers.Peers) == 0 {
		return nil, &peerNotAvailableError{}
	}
	
	protomsg := &network.Whoareyoumsg{Who: nil}
	
	err := self.peers.Peers[peerindex - 1].Send(protomsg)
	if err != nil {
		return nil, &sendToPeerError{details: err.Error()}
	}

	return "ok", nil
}

func NewWhoAreYou(self *network.PeerCollection) *WhoAreYou {
	return &WhoAreYou{peers: self}
}
