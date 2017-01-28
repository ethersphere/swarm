package api

import (
	"fmt"
	"time"
	
	"github.com/ethereum/go-ethereum/META/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
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

type syntaxError struct {
	details string
}
func (e *syntaxError) ErrorCode() int { return 0xFF010000 }
func (e *syntaxError) Error() string { return fmt.Sprintf("Request syntax error") }


type ZeroKeyBroadcast struct {
	peers *network.PeerCollection
}

func (self *ZeroKeyBroadcast) Sendzeronotification(atype uint8) (interface{}, error) {
	
	sends := 0
	expire,_ := time.Now().Add(network.METADefaultExpireDuration).MarshalBinary()
	
	if len(self.peers.Peers) == 0 {
		return nil, &peerNotAvailableError{}
	}
	if network.METAAssetType[atype] == "" {
		return nil, &syntaxError{}
	}
	
	
	protomsg := &network.METAAssetNotification{Typ: atype, Bzz: storage.ZeroKey, Exp: expire}
	
	for _,p := range self.peers.Peers {
		err := p.Send(protomsg)
		if err != nil {
			return nil, &sendToPeerError{details: err.Error()}
		} else {
			sends++
		}
	}
	
	return fmt.Sprintf("sent to %v peers", sends), nil
}

func NewZeroKeyBroadcast(self *network.PeerCollection) *ZeroKeyBroadcast {
	return &ZeroKeyBroadcast{peers: self}
}
