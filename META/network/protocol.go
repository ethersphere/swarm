package network 

import (
	"sync"
	//"time"
	
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	//"github.com/ethereum/go-ethereum/swarm/storage"
	
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	
)

func init() {
	glog.SetV(6)
	glog.SetToStderr(true)
}

const (
	ProtocolName     = "mw"
	ProtocolVersion  = 0x000003
	NetworkId          = 1666 
	ProtocolMaxMsgSize = 10 * 1024 * 1024
)

func NewMETACodeMap(msgs ...interface{}) *protocols.CodeMap {
	ct := protocols.NewCodeMap(ProtocolName, ProtocolVersion, ProtocolMaxMsgSize)
	ct.Register(msgs...)
	return ct
}

func METAProtocol(protopeers *PeerCollection, ct *protocols.CodeMap, na adapters.NodeAdapter, wg *sync.WaitGroup) p2p.Protocol {

	run := func(peer *protocols.Peer) error {
		
		if wg != nil {
			wg.Add(1)
		}
		
		peer.Register(&METATmpName{}, func(msg interface{}) error {
			glog.V(logger.Debug).Infof("Beforeparse")
			t := msg.(*METATmpName)	
			glog.V(logger.Debug).Infof("Peer received tmpname name '%s'", t.Name)
			return nil
		})
		
		/*peer.Register(&METAAssetNotification{}, func(msg interface{}) error {
			hm := msg.(*METAAssetNotification)	
			glog.V(logger.Debug).Infof("Peer received asset notification %v", METAAssetType[hm.Typ])
			return nil
		})*/
		
		protopeers.Add(peer)
		
		err := peer.Run()
		if wg != nil {
			wg.Done()
		}
		glog.V(logger.Debug).Infof("protocol died!! %v", err)
		return err
	}		
	
	p := protocols.NewProtocol(ProtocolName, ProtocolVersion, run, na, ct)
	return *p
}
