package network

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pot"
)

/*
 Adaptive capability

*/

type Capabilities struct {
	Flags         []capability
	changeC       chan<- capability
	notifiers     map[rpc.ID]*rpc.Notifier
	lastChangeCId uint64
	mu            sync.RWMutex
}

type CapabilitiesMsg Capabilities

func NewCapabilities(changeC chan<- capability) *Capabilities {
	return &Capabilities{
		changeC:   changeC,
		notifiers: make(map[rpc.ID]*rpc.Notifier),
	}
}

func (m Capabilities) SubscribeChange(ctx context.Context) (*rpc.Subscription, error) {
	notifier, ok := rpc.NotifierFromContext(ctx)
	if !ok {
		return nil, errors.New("notifications not supported")
	}
	sub := notifier.CreateSubscription()
	m.notifiers[sub.ID] = notifier
	go func(sub *rpc.Subscription, notifier *rpc.Notifier) {
		select {
		case err := <-sub.Err():
			log.Warn("rpc capabilities subscription end", "err", err)
		case <-notifier.Closed():
			log.Warn("rpc capabilities notifier closed")
		}
	}(sub, notifier)
	return sub, nil
}

func (m Capabilities) notify(c capability) {
	if m.changeC != nil {
		m.changeC <- c
	}
	for id, notifier := range m.notifiers {
		notifier.Notify(id, c)
	}
}

func (m Capabilities) destroy() {
	if m.changeC != nil {
		close(m.changeC)
	}
}

func (m Capabilities) toMsg() CapabilitiesMsg {
	return CapabilitiesMsg(m)
}

func (m Capabilities) String() string {
	var caps []string
	for _, c := range m.Flags {
		caps = append(caps, fmt.Sprintf("%02x:%v", c[0], pot.ToBin(c[2:])))
	}
	return strings.Join(caps, ",")
}

func (m Capabilities) get(id uint8) capability {
	if len(m.Flags) == 0 {
		return nil
	}
	for _, cs := range m.Flags {
		if cs[0] == id {
			return cs
		}
	}
	return nil
}

func (m *Capabilities) add(c capability) {
	m.Flags = append(m.Flags, c)
}

func (m *Capabilities) SetCapability(id uint8, flags []byte) error {
	if len(flags) == 0 {
		return errors.New("flag bytes cannot be empty")
	}
	c := m.get(id)
	if c == nil {
		return fmt.Errorf("capability id %d not registered", id)
	}
	m.mu.Lock()
	err := c.set(flags)
	ccopy := newCapability(c[0], c[1])
	ccopy.set(c[2:])
	m.mu.Unlock()
	if err != nil {
		return err
	}
	m.notify(ccopy)
	return nil
}

func (m *Capabilities) RemoveCapability(id uint8, flags []byte) error {
	c := m.get(id)
	if c == nil {
		return fmt.Errorf("capability id %d not registered", id)
	}
	m.mu.Lock()
	err := c.unset(flags)
	ccopy := newCapability(c[0], c[1])
	ccopy.set(c[2:])
	m.mu.Unlock()
	if err != nil {
		return err
	}
	m.notify(ccopy)
	return nil
}

func (m *Capabilities) RegisterCapabilityModule(id uint8, length uint8) error {
	c := m.get(id)
	if c != nil {
		return fmt.Errorf("capability %d already registered", id)
	}
	c = newCapability(id, length)
	m.add(c)
	return nil
}

type capability []byte

func newCapability(code uint8, byteLength uint8) capability {
	c := make(capability, byteLength+2)
	c[0] = code
	c[1] = byteLength
	return c
}

func (c *capability) set(flag []byte) error {
	if !c.validLength(flag) {
		return fmt.Errorf("Bitfield must be %d bytes long", len(*c))
	}
	for i, b := range flag {
		(*c)[2+i] |= b
	}
	return nil
}

func (c *capability) unset(flag []byte) error {
	if !c.validLength(flag) {
		return fmt.Errorf("Bitfield must be %d bytes long", len(*c))
	}
	for i, b := range flag {
		(*c)[2+i] &= ^b
	}
	return nil
}

func (c *capability) validLength(flag []byte) bool {
	if len(flag) != len(*c)-2 {
		return false
	}
	return true
}
