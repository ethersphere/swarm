package network 

import (
	"sync"
	"fmt"
	
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	
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
		
		peer.Register(&METAAnnounce{}, func(msg interface{}) error {
			hm := msg.(*METAAnnounce)	
			glog.V(logger.Debug).Infof("Peer received announce", hm)
			glog.V(logger.Debug).Infof("Command: %v", hm.GetCommand())
			glog.V(logger.Debug).Infof("Uuid: %v", hm.GetUuid())
			for i, p := range hm.Payload {
				glog.V(logger.Debug).Infof("Payload #%d type %d length %d:", i, p.GetType(), p.Length())
				for ii := 0; ii < p.Length(); ii++ {
					label, data := p.GetRawEntry(ii)
					//glog.V(logger.Debug).Infof("- Entry #%d Label %s Data %v:", ii, p.Label[ii], p.Data[ii])
					glog.V(logger.Debug).Infof("- Entry #%d Label %s Data %v:", ii, label, data)
				}
			}
			return nil
		})
		
		peer.Register(&METATmpName{}, func(msg interface{}) error {
			t := msg.(*METATmpName)	
			METATmpSwarmRegistryLookup[t.Name] = [2]string{fmt.Sprintf("%v",t.Node), fmt.Sprintf("%v",t.Swarmhash)}
			return nil
		})
		
	
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
