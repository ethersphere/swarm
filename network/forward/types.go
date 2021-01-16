package forward

import (
	"time"

	"github.com/ethersphere/swarm/network"
)

var (
	zeroTime = time.Unix(0, 0)
)

// SessionInterface provides an interface for an individual session object
type SessionInterface interface {
	Subscribe() <-chan *network.Peer
	Get(numberOfPeers int) ([]*network.Peer, error)
}

// SessionContext is a context.Context that can be used to reference existing sessions or create new sessions
type SessionContext struct {
	CapabilityIndex string
	SessionId       int
	Address         []byte
}

// NewSessionContext creates a new SessionContext with the provided capabilityIndex and base address
func NewSessionContext(capabilityIndex string, base []byte) *SessionContext {
	return &SessionContext{
		CapabilityIndex: capabilityIndex,
		Address:         base,
	}
}

// Deadline implements context.Context
func (c *SessionContext) Deadline() (time.Time, bool) {
	return zeroTime, false
}

// Done implements context.Context
func (c *SessionContext) Done() <-chan struct{} {
	return nil
}

// Err implements context.Context
func (c *SessionContext) Err() error {
	return nil
}

// Value implements context.Context
func (c *SessionContext) Value(k interface{}) interface{} {
	ks, ok := k.(string)
	if !ok {
		return nil
	}
	switch ks {
	case "address":
		if c.Address == nil {
			return nil
		}
		return c.Address
	case "capability":
		if c.CapabilityIndex == "" {
			return nil
		}
		return c.CapabilityIndex
	case "id":
		return c.SessionId
	}
	return nil
}
