package network

import (
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

// TODO: check if code already exists in db
func (m *Capabilities) Add(c Capability) {
	*m = append(*m, c)
}

type Capability []byte

func NewCapability(code uint8, byteLength uint8) Capability {
	c := make(Capability, byteLength+2)
	c[0] = code
	c[1] = byteLength
	return c
}

func (c *Capability) Set(flag []byte) error {
	if len(flag) != len(*c)-2 {
		return fmt.Errorf("Bitfield must be %d bytes long", len(*c))
	}
	for i, b := range flag {
		(*c)[2+i] |= b
	}
	return nil
}
