package forward

import (
	"time"

	"github.com/ethersphere/swarm/network"
)

var (
	zeroTime = time.Unix(0, 0)
)

type ForwardPeer struct {
	*network.BzzPeer
}

type SessionInterface interface {
	Subscribe() <-chan ForwardPeer
	Get(numberOfPeers int) ([]ForwardPeer, error)
	Close()
}

// also implements context.Context
type SessionContext struct {
	CapabilityIndex string
	SessionId       int
	Address         []byte
}

func NewSessionContext() *SessionContext {
	sctx := newSessionContext("", sessionId, nil)
	return sctx
}

func newSessionContext(capabilityIndex string, sessionId int, addr []byte) *SessionContext {
	return &SessionContext{
		CapabilityIndex: capabilityIndex,
		SessionId:       sessionId,
		Address:         addr,
	}
}

func (c *SessionContext) Deadline() (time.Time, bool) {
	return zeroTime, false
}

func (c *SessionContext) Done() <-chan struct{} {
	return nil
}

func (c *SessionContext) Err() error {
	return nil
}

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

func (c *SessionContext) SetAddress(addr []byte) {
	c.Address = addr
}

func (c *SessionContext) SetCapability(capabilityIndex string) {
	c.CapabilityIndex = capabilityIndex
}
