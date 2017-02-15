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
		
		peer.Register(&METATmpName{}, func(msg interface{}) error {
			t := msg.(*METATmpName)	
			METATmpSwarmRegistryLookup[t.Name] = [2]string{fmt.Sprintf("%v",t.Node), fmt.Sprintf("%v",t.Swarmhash)}
			// get duplicates when using pointer, even if the pointer is to an element in a persistent array, why?
			// cant use storage.Key as map key, why?
			//METATmpSwarmRegistryLookup[t.Node] = fmt.Sprintf("%v",t.Swarmhash)
			//METATmpSwarmRegistryLookupReverse[fmt.Sprintf("%v",t.Swarmhash)] = t.Node
			/*
			METATmpSwarmRegistryKeys = append(METATmpSwarmRegistryKeys, t.Swarmhash)
			swarmhashp := &METATmpSwarmRegistryKeys[len(METATmpSwarmRegistryKeys)-1]
			METATmpSwarmRegistryLookup[t.Node] = *swarmhashp
			METATmpSwarmRegistryLookupReverse[*swarmhashp] = t.Node
			*/ 
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
