package network 

import (
	"sync"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	//"github.com/ethereum/go-ethereum/p2p/adapters"
	
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

const (
	ProtocolName       = "mw"
	Version            = 0x000001
	NetworkId          = 1666 
	ProtocolMaxMsgSize = 10 * 1024 * 1024
)


func METACodeMap(msgs ...interface{}) *protocols.CodeMap {
	ct := protocols.NewCodeMap(ProtocolName, Version, ProtocolMaxMsgSize)
	//ct.Register(&METAHandshake{})
	ct.Register(msgs...)
	return ct
}

type METAMessenger struct {
}

func (METAMessenger) SendMsg(w p2p.MsgWriter, code uint64, msg interface{}) error {
	return p2p.Send(w, code, msg)
}

func (METAMessenger) ReadMsg(r p2p.MsgReader) (p2p.Msg, error) {
	return r.ReadMsg()
}

type Hellofirstnodemsg struct {
	Pmsg string
	Sub protocols.Peer
}

//func newProtocol(pp *p2ptest.TestPeerPool, wg *sync.WaitGroup) func(adapters.NodeAdapter) adapters.ProtoCall {

func METAProtocol(protopeers *PeerCollection, wg *sync.WaitGroup) p2p.Protocol {

//func META(localAddr []byte, hive PeerPool, na adapters.NodeAdapter, m adapters.Messenger, ct *protocols.CodeMap, services func(Node) error) *p2p.Protocol {
	// handle handshake
	
	ct := METACodeMap(&Hellofirstnodemsg{})

	m := METAMessenger{}
	
	run := func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		
		if wg != nil {
			wg.Add(1)
		}
		peer := protocols.NewPeer(p, rw, ct, m, func() { })
		
		peer.Register(&Hellofirstnodemsg{}, func(msg interface{}) error {
			glog.V(logger.Debug).Infof("hellofirstnode got something")
			peer.Send("yo")
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
		Name:     ProtocolName,
		Version:  Version,
		Length:   ct.Length(),
		Run:      run,
	}
}
