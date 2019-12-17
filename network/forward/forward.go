package forward

import (
	"context"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
)

type Session struct {
	sessionContext context.Context
	kademlia       *network.Kademlia
	pivot          []byte
}

func New(sctx context.Context, kad *network.Kademlia) *Session {
	s := &Session{
		sessionContext: sctx,
		kademlia:       kad,
	}
	addr := sctx.Value("address")
	log.Trace("addr", "addr", addr)
	if addr == nil {
		s.pivot = kad.BaseAddr()
	} else {
		s.pivot = addr.([]byte)
	}
	return s
}

func (s *Session) Get(numPeers int) ([]ForwardPeer, error) {
	var result []ForwardPeer

	return result, nil
}
