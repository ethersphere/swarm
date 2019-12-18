package forward

import (
	"fmt"
	"sync"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
)

type Session struct {
	kademlia        *network.KademliaLoadBalancer
	pivot           []byte
	id              int
	capabilityIndex string
	nextC           chan struct{}
	getC            chan *ForwardPeer
}

type SessionManager struct {
	sessions map[int]*Session
	kademlia *network.KademliaLoadBalancer
	lastId   int // starts at 1 to make create from context easier
	mu       sync.Mutex
}

func NewSessionManager(kademlia *network.KademliaLoadBalancer) *SessionManager {
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
		nextC:           make(chan struct{}),
		getC:            make(chan *ForwardPeer),
	}
	if pivot == nil {
		s.pivot = m.kademlia.BaseAddr()
	} else {
		s.pivot = pivot
	}
	go s.load()
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

func (s *Session) Get(numPeers int) ([]*ForwardPeer, error) {
	var result []*ForwardPeer
	for i := 0; i < numPeers; i++ {
		s.nextC <- struct{}{}
		p, ok := <-s.getC
		if !ok {
			break
		}
		result = append(result, p)
	}
	return result, nil
}

func (s *Session) load() error {
	err := s.kademlia.EachBinFiltered(s.pivot, s.capabilityIndex, func(bin network.LBBin) bool {
		for _, p := range bin.LBPeers {
			_, ok := <-s.nextC
			if !ok {
				return false
			}
			s.getC <- &ForwardPeer{Peer: p.Peer}
		}
		return true
	})
	close(s.getC)
	return err
}

func (s *Session) Close() {
	close(s.nextC)
}
