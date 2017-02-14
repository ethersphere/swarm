package network

import (
	"github.com/ethereum/go-ethereum/p2p/protocols"
	
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

/***
 * \todo test expenditure if struct will take more memory and/or processing than map
 */
type METAPeer struct {
	*protocols.Peer
	Answersbroadcast bool
}

type PeerCollection struct {
	Peers []*METAPeer
}

var PeerIndex map[*protocols.Peer]int

func NewPeerCollection() *PeerCollection {
	PeerIndex = make(map[*protocols.Peer]int)
	return &PeerCollection{}
}

func (self *PeerCollection) Add(p *protocols.Peer) error {

	self.Peers = append(self.Peers, &METAPeer{Peer: p, Answersbroadcast: true})
	PeerIndex[p] = len(self.Peers) - 1
	glog.V(logger.Debug).Infof("protopeers now has added peers %v, total %v", self.Peers[len(self.Peers) - 1].ID().String(), len(self.Peers))
	return nil
}
/*
func (self *PeerCollection) Remove(rp *protocols.Peer) int {
	var p *METAPeer
	i := -1
	
	for i, p = range self.Peers {
		if p == rp {
			break;
		}
	}

	if i > -1 {
		self.RemoveIndex(i)
	}
	return i
}

func (self *PeerCollection) RemoveIndex(i uint) error {
	
}
*/
