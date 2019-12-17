package forward

import (
	"github.com/ethersphere/swarm/network"
)

var (
	sessionId = 0
	sessions  []*Session
)

type Session struct {
	kademlia        *network.Kademlia
	pivot           []byte
	id              int
	capabilityIndex string
}

func New(kad *network.Kademlia, capabilityIndex string, pivot []byte) *Session {
	s := &Session{
		kademlia:        kad,
		id:              sessionId,
		capabilityIndex: capabilityIndex,
	}
	if pivot == nil {
		s.pivot = kad.BaseAddr()
	} else {
		s.pivot = pivot
	}
	sessionId++
	return s
}

//func NewFromContext(sctx *SessionContext, kad *network.Kademlia) *Session {
//	s := &Session{
//		kademlia: kad,
//	}
//
//	s.id = sctx.Value("id").(int)
//
//	addr := sctx.Value("address")
//	if addr == nil {
//		s.pivot = kad.BaseAddr()
//	} else {
//		s.pivot = addr.([]byte)
//	}
//
//	capabilityIndex := sctx.Value("capability")
//	if capabilityIndex != nil {
//		s.capabilityIndex = capabilityIndex.(string)
//	}
//
//	return s
//}

func (s *Session) Get(numPeers int) ([]ForwardPeer, error) {
	var result []ForwardPeer

	return result, nil
}
