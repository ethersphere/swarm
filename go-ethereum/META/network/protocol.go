package network 

import (
	"fmt"
	"sync"
	"time"
	
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	
)

const (
	ProtocolPrefix     = "mw"
	Version            = 0x000001
	NetworkId          = 1666 
	ProtocolMaxMsgSize = 10 * 1024 * 1024
)

const (
	MessageRequest = iota
	MessageReply
)

func METACodeMap1(msgs ...interface{}) *protocols.CodeMap {
	ct := protocols.NewCodeMap(ProtocolPrefix + "1", Version, ProtocolMaxMsgSize)
	ct.Register(msgs...)
	return ct
}

/***
 * \todo implement module handshake demo
 */
func METACodeMap2(msgs ...interface{}) *protocols.CodeMap {
	ct := protocols.NewCodeMap(ProtocolPrefix + "2", Version, ProtocolMaxMsgSize)
	ct.Register(msgs...)
	return ct
}

type Hellofirstnodemsg struct {
	Type uint
	Pmsg string
	Now time.Time
}

type Helloallnodemsg struct {
	Type uint
	Pmsg string
}

type Whoareyoumsg struct {
	Who *p2p.Peer
}

func METAProtocol1(protopeers *PeerCollection, wg *sync.WaitGroup, consolechan chan string) p2p.Protocol {

	ct := METACodeMap1(&Hellofirstnodemsg{}, &Helloallnodemsg{})

	m := adapters.RLPxMessenger{}
	
	run := func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		
		glog.V(logger.Debug).Infof("Registering protocol 1 for peer %v", p.ID().String())
		
		if wg != nil {
			wg.Add(1)
		}
		peer := protocols.NewPeer(p, rw, ct, m, func() { })
		
		
		peer.Register(&Hellofirstnodemsg{}, func(msg interface{}) error {
			
			hm := msg.(*Hellofirstnodemsg)
			if hm.Type == MessageRequest  {
				hm := &Hellofirstnodemsg{Type: MessageReply, Pmsg: "received!", Now: time.Now()}
				peer.Send(hm)
			} else {
				consolechan <- fmt.Sprintf("peer %v received %v", peer.ID().String(), hm.Pmsg)
			}
			return nil
		})
		
		peer.Register(&Helloallnodemsg{}, func(msg interface{}) error {
			hm := msg.(*Helloallnodemsg)
			glog.V(logger.Debug).Infof("peerindex of %v has answersbroadcast %v", p.ID().String(), protopeers.Peers[PeerIndex[peer]].Answersbroadcast)
			if protopeers.Peers[PeerIndex[peer]].Answersbroadcast != true { // get peer in collection by peer pointer address
				consolechan <- ""
				return nil
			}
			
			if hm.Type == MessageRequest  {
				hm := &Helloallnodemsg{Type: MessageReply, Pmsg: "received!"}
				peer.Send(hm)
			} else {
				consolechan <- fmt.Sprintf("peer %v received %v", peer.ID().String(), hm.Pmsg)
			}
			
			return nil
		})
		
		protopeers.Add(peer)
		
		err := peer.Run()
		if wg != nil {
			wg.Done()
		}
		return err
	}		
	
	return p2p.Protocol{
		Name:     ProtocolPrefix + "1",
		Version:  Version,
		Length:   ct.Length(),
		Run:      run,
	}
}

func METAProtocol2(protopeers *PeerCollection) p2p.Protocol {

	ct := METACodeMap2(&Whoareyoumsg{})

	m := adapters.RLPxMessenger{}
	
	run := func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		
		glog.V(logger.Debug).Infof("Registering protocol 2 for peer %v", p.ID().String())
		
		peer := protocols.NewPeer(p, rw, ct, m, func() { })
		
		peer.Register(&Whoareyoumsg{}, func(msg interface{}) error {
			hm := msg.(*Whoareyoumsg)
			if hm.Who == nil {
				hm = &Whoareyou{Who: p}
				peer.Send(hm)
			}
			return nil
		})
		
		protopeers.Add(peer)
		
		err := peer.Run()

		return err
	}		
	
	return p2p.Protocol{
		Name:     ProtocolPrefix + "2",
		Version:  Version,
		Length:   ct.Length(),
		Run:      run,
	}
}
