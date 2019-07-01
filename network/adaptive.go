package network

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pot"
)

/*
 Adaptive capability

*/

type Capabilities struct {
	Flags   []capability
	changeC chan []byte
}

type CapabilitiesMsg []capability

func NewCapabilities() *Capabilities {
	return &Capabilities{
		changeC: make(chan []byte),
	}
}

func (m Capabilities) Destroy() {
	close(m.changeC)
}

func (m Capabilities) toMsg() CapabilitiesMsg {
	return CapabilitiesMsg(m.Flags)
}

func (m Capabilities) String() string {
	var caps []string
	for _, c := range m.Flags {
		caps = append(caps, fmt.Sprintf("%02x:%v", c[0], pot.ToBin(c[2:])))
	}
	return strings.Join(caps, ",")
}

func (m Capabilities) get(id uint8) capability {
	log.Error("get", "c", id, "m", m.Flags, "l", len(m.Flags))
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

// TODO: check if code already exists in db
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
	err := c.set(flags)
	if err != nil {
		return err
	}
	chanBytes := make([]byte, len(c))
	copy(chanBytes, c)
	m.changeC <- chanBytes
	return nil
}

func (m *Capabilities) RemoveCapability(id uint8, flags []byte) error {
	c := m.get(id)
	if c == nil {
		return fmt.Errorf("capability id %d not registered", id)
	}
	err := c.unset(flags)
	if err != nil {
		return err
	}
	chanBytes := make([]byte, len(c))
	copy(chanBytes, c)
	m.changeC <- chanBytes
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
	log.Error("set", "flag", flag, "c", (*c)[2:])
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
