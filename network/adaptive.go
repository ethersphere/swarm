package network

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ethersphere/swarm/pot"
)

/*
 Adaptive capability

*/

type Capabilities struct {
	Flags   []capability
	changeC chan<- capability
	mu      sync.Mutex
}

func NewCapabilities(changeC chan<- capability) *Capabilities {
	return &Capabilities{
		changeC: changeC,
	}
}

func (c Capabilities) String() string {
	var caps []string
	for _, cap := range c.Flags {
		caps = append(caps, fmt.Sprintf("%02x:%v", cap[0], pot.ToBin(cap[2:])))
	}
	return strings.Join(caps, ",")
}

func (m Capabilities) destroy() {
	if m.changeC != nil {
		close(m.changeC)
	}
}

func (m Capabilities) notify(c capability) {
	if m.changeC != nil {
		m.changeC <- c
	}
}

func (c Capabilities) get(id uint8) capability {
	if len(c.Flags) == 0 {
		return nil
	}
	for _, caps := range c.Flags {
		if caps[0] == id {
			return caps
		}
	}
	return nil
}

func (c *Capabilities) add(cap capability) {
	c.Flags = append(c.Flags, cap)
}

func (c *Capabilities) set(id uint8, flags []byte) error {
	if len(flags) == 0 {
		return errors.New("flag bytes cannot be empty")
	}
	cap := c.get(id)
	if cap == nil {
		return fmt.Errorf("capability id %d not registered", id)
	}
	c.mu.Lock()
	err := cap.set(flags)
	ccopy := newCapability(cap[0], cap[1])
	ccopy.set(cap[2:])
	c.mu.Unlock()
	if err != nil {
		return err
	}
	c.notify(ccopy)
	return nil
}

func (c *Capabilities) unset(id uint8, flags []byte) error {
	cap := c.get(id)
	if cap == nil {
		return fmt.Errorf("capability id %d not registered", id)
	}
	c.mu.Lock()
	err := cap.unset(flags)
	ccopy := newCapability(cap[0], cap[1])
	ccopy.set(cap[2:])
	c.mu.Unlock()
	if err != nil {
		return err
	}
	c.notify(ccopy)
	return nil
}

func (c *Capabilities) registerModule(id uint8, length uint8) error {
	cap := c.get(id)
	if cap != nil {
		return fmt.Errorf("capability %d already registered", id)
	}
	cap = newCapability(id, length)
	c.add(cap)
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
