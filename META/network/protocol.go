package network 

import (
	"sync"
	"time"
	
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	"github.com/ethereum/go-ethereum/swarm/storage"
	
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	
)

const (
	ProtocolName     = "mw"
	Version            = 0x000002
	NetworkId          = 1666 
	ProtocolMaxMsgSize = 10 * 1024 * 1024
)

// METAWire SYN/ACK
const (
	MessageRequest = iota
	MessageReply
)

// METAWire notification enum
const (
	ERN = iota
	DSR
	MLC
)

var METADefaultExpireDuration = time.Hour * 5

var METAAssetType = map[uint8]string{
	ERN: "Eletronic Release Notification",
	DSR: "Digital Sales Report",
	MLC: "Music Licensing Company",
}

func METACodeMap(msgs ...interface{}) *protocols.CodeMap {
	ct := protocols.NewCodeMap(ProtocolName, Version, ProtocolMaxMsgSize)
	ct.Register(msgs...)
	return ct
}
type METAAssetNotification struct {
	Typ uint8
	Bzz storage.Key // this has no set length? Can it be both sha-3 and sha-256?
	Exp []byte // byte marshalled time
}

func METAProtocol(protopeers *PeerCollection, wg *sync.WaitGroup) p2p.Protocol {

	ct := METACodeMap(&METAAssetNotification{})

	m := adapters.RLPxMessenger{}
	
	run := func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		
		if wg != nil {
			wg.Add(1)
		}
		peer := protocols.NewPeer(p, rw, ct, m, func() { })
		
		
		peer.Register(&METAAssetNotification{}, func(msg interface{}) error {
			hm := msg.(*METAAssetNotification)	
			glog.V(logger.Debug).Infof("Peer received asset notification %v", METAAssetType[hm.Typ])
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
