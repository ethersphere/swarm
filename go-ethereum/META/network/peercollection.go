package network

import (
	"github.com/ethereum/go-ethereum/p2p/protocols"
	
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

type PeerCollection struct {
	Peers []*protocols.Peer
}

func (self *PeerCollection) Add(p *protocols.Peer) error {
	self.Peers = append(self.Peers, p)
	glog.V(logger.Debug).Infof("protopeers now %v", *self.Peers[0])
	return nil
}
