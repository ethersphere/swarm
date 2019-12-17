package forward

import (
	"github.com/ethersphere/swarm/network"
)

type Session struct {
	kademlia        *network.Kademlia
	pivot           []byte
	id              int
	capabilityIndex string
}

func NewFromContext(sctx *SessionContext, kad *network.Kademlia) *Session {
	s := &Session{
		kademlia: kad,
	}

	s.id = sctx.Value("id").(int)

	addr := sctx.Value("address")
	if addr == nil {
		s.pivot = kad.BaseAddr()
	} else {
		s.pivot = addr.([]byte)
	}

	capabilityIndex := sctx.Value("capability")
	if capabilityIndex != nil {
		s.capabilityIndex = capabilityIndex.(string)
	}

	return s
}

func (s *Session) Get(numPeers int) ([]ForwardPeer, error) {
	var result []ForwardPeer

	return result, nil
}
