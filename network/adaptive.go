package network

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethersphere/swarm/pot"
)

/*
 Adaptive capability

*/

type Capabilities [][]byte
type CapabilitiesMsg Capabilities

func (m Capabilities) String() string {
	var caps []string
	for _, c := range m {
		caps = append(caps, fmt.Sprintf("%02x:%v", c[0], pot.ToBin(c[2:])))
	}
	return strings.Join(caps, ",")
}

func (m *Capabilities) get(id uint8) Capability {
	if len(*m) == 0 {
		return nil
	}
	for _, cs := range *m {
		if cs[0] == id {
			return cs
		}
	}
	return nil
}

// TODO: check if code already exists in db
func (m *Capabilities) add(c Capability) {
	*m = append(*m, c)
}

func (m *Capabilities) SetCapability(id uint8, flags []byte) error {
	if len(flags) < 2 {
		return errors.New("invalid capability flag length")
	}
	c := m.get(id)
	if c == nil {
		return fmt.Errorf("capability id %d not registered", id)
	}
	return c.Set(flags)
}

func (m *Capabilities) RemoveCapability(id uint8, flags []byte) error {
	c := m.get(id)
	if c == nil {
		return fmt.Errorf("capability id %d not registered", id)
	}
	return c.Unset(flags)
}

func (m *Capabilities) RegisterCapabilityModule(id uint8, length uint8) error {
	c := m.get(id)
	if c != nil {
		return fmt.Errorf("capability %d already registered", id)
	}
	c = NewCapability(id, length)
	m.add(c)
	return nil
}

type Capability []byte

func NewCapability(code uint8, byteLength uint8) Capability {
	c := make(Capability, byteLength+2)
	c[0] = code
	c[1] = byteLength
	return c
}

func (c *Capability) Set(flag []byte) error {
	if !c.validLength(flag) {
		return fmt.Errorf("Bitfield must be %d bytes long", len(*c))
	}
	for i, b := range flag {
		(*c)[2+i] |= b
	}
	return nil
}

func (c *Capability) Unset(flag []byte) error {
	if !c.validLength(flag) {
		return fmt.Errorf("Bitfield must be %d bytes long", len(*c))
	}
	for i, b := range flag {
		(*c)[2+i] &= ^b
	}
	return nil
}

func (c *Capability) validLength(flag []byte) bool {
	if len(flag) != len(*c)-2 {
		return false
	}
	return true
}
