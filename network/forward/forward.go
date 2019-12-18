package forward

import (
	"fmt"
	"sync"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
)

type Session struct {
	kademlia        *network.Kademlia
	pivot           []byte
	id              int
	capabilityIndex string
}

type SessionManager struct {
	sessions map[int]*Session
	kademlia *network.Kademlia
	lastId   int // starts at 1 to make create from context easier
	mu       sync.Mutex
}

func NewSessionManager(kademlia *network.Kademlia) *SessionManager {
	return &SessionManager{
		sessions: make(map[int]*Session),
		kademlia: kademlia,
	}
}

func (m *SessionManager) add(s *Session) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastId++
	log.Trace("adding session", "id", m.lastId)
	s.id = m.lastId
	m.sessions[m.lastId] = s
	return s
}

func (m *SessionManager) New(capabilityIndex string, pivot []byte) *Session {
	s := &Session{
		capabilityIndex: capabilityIndex,
		kademlia:        m.kademlia,
	}
	if pivot == nil {
		s.pivot = m.kademlia.BaseAddr()
	} else {
		s.pivot = pivot
	}
	return m.add(s)
}

func (m *SessionManager) ToContext(id int) (*SessionContext, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("No such session %d", id)
	}
	return &SessionContext{
		CapabilityIndex: s.capabilityIndex,
		SessionId:       s.id,
		Address:         s.pivot,
	}, nil
}

func (m *SessionManager) FromContext(sctx *SessionContext) (*Session, error) {

	sessionId, ok := sctx.Value("id").(int)
	if ok {
		s, ok := m.sessions[sessionId]
		if !ok {
			return nil, fmt.Errorf("No such session %d", sessionId)
		}
		return s, nil
	}

	addr, _ := sctx.Value("address").([]byte)
	capabilityIndex, _ := sctx.Value("capability").(string)
	return m.New(capabilityIndex, addr), nil
}

func (s *Session) Get(numPeers int) ([]ForwardPeer, error) {
	var result []ForwardPeer

	i := 0
	err := s.kademlia.EachConnFiltered(s.pivot, s.capabilityIndex, 255, func(p *network.Peer, po int) bool {
		result = append(result, ForwardPeer{Peer: p})
		i++
		if i == numPeers {
			return false
		}
		return true
	})
	return result, err
}
