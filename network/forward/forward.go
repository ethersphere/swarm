package forward

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
)

var (
	NoMorePeers = errors.New("no more peers")
)

// Session encapsulates one single peer iteration query
type Session struct {
	kademlia        *network.KademliaLoadBalancer // kademlia backend
	base            []byte                        //
	id              int
	capabilityIndex string
	nextC           chan struct{}
	getC            chan *network.Peer
}

// Id returns the session id
func (s *Session) Id() int {
	return s.id
}

// Get returns up to numPeers peers from the current position of the iterator
// If no further peers are available a NoMorePeers error will be returned
func (s *Session) Get(numPeers int) ([]*network.Peer, error) {
	var result []*network.Peer
	select {
	case <-s.getC:
		return result, NoMorePeers
	default:
	}
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

// starts the iterator and blocks for request for next peer through Get()
func (s *Session) load() error {
	err := s.kademlia.EachBinFiltered(s.base, s.capabilityIndex, func(bin network.LBBin) bool {
		for _, p := range bin.LBPeers {
			_, ok := <-s.nextC
			if !ok {
				return false
			}
			s.getC <- p.Peer
		}
		return true
	})
	close(s.getC)
	return err
}

// frees resources
func (s *Session) destroy() {
	close(s.nextC)
}

// SessionManager is the Session object factory
type SessionManager struct {
	sessions map[int]*Session
	kademlia *network.KademliaLoadBalancer
	lastId   int // starts at 1 to make create from context easier
	mu       sync.Mutex
}

// NewSessionManager is the SessionManager constructor
// Sessions created with the SessionManager will use the provided kademlia backend
// TODO: argument should be network.KademliaBackend, but needs KademliaLoadBalancer to implement this
func NewSessionManager(kademlia *network.KademliaLoadBalancer) *SessionManager {
	return &SessionManager{
		sessions: make(map[int]*Session),
		kademlia: kademlia,
	}
}

// New creates a new Session object with the given capabilityindex and base address
// if capabilityIndex is empty, the global kademlia database will be used
// if base is nil, the kademlia base address will be used as comparator for the iteration
func (m *SessionManager) New(capabilityIndex string, base []byte) *Session {
	s := &Session{
		capabilityIndex: capabilityIndex,
		kademlia:        m.kademlia,
		nextC:           make(chan struct{}),
		getC:            make(chan *network.Peer),
	}
	if base == nil {
		s.base = m.kademlia.BaseAddr()
	} else {
		s.base = base
	}
	go s.load()
	return m.add(s)
}

// Reap frees the Session object resources and removes it from the session index
func (m *SessionManager) Reap(sessionId int) {
	s, ok := m.sessions[sessionId]
	if !ok {
		return
	}
	s.destroy()
}

// ToContext creates a SessionContext from the existing Session matching the provided id
// if the session does not exist an error is returned
func (m *SessionManager) ToContext(id int) (*SessionContext, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("No such session %d", id)
	}
	return &SessionContext{
		CapabilityIndex: s.capabilityIndex,
		SessionId:       s.id,
		Address:         s.base,
	}, nil
}

// FromContext retrieves or creates a Session from a provided context
// If the context has the "id" value set, the corresponding Session is returned, or error if it does not exist
// Otherwise, a new Session is created and returned, optionally with the "address" and/or "capability" values provided in the context
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

// adds a new session to the sessionmanager
func (m *SessionManager) add(s *Session) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastId++
	log.Trace("adding session", "id", m.lastId)
	s.id = m.lastId
	m.sessions[m.lastId] = s
	return s
}
