// For more information on Capabilities, see https://github.com/ethersphere/SWIP/blob/master/lightnode_caps_msg.md
package network

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ethersphere/swarm/pot"
)

// Capabilities is a container that holds all capability bitvector flags for all registered modules
type Capabilities struct {
	Flags   []capability      // the bitvector flags
	changeC chan<- capability // builtin notification of changes for the base swarm protocol
	mu      sync.Mutex        // protects copy of state of flags for notification
}

// NewCapabilities creates a new Capabilities container
//
// It optionally takes a change notification channel as argument. If this is nil, notifications will not be issued.
func NewCapabilities(changeC chan<- capability) *Capabilities {
	return &Capabilities{
		changeC: changeC,
	}
}

// String implements Stringer interface
func (c Capabilities) String() string {
	return capArrayToString(c.Flags)
}

// close the internal notification channel if it exists
func (m Capabilities) destroy() {
	if m.changeC != nil {
		close(m.changeC)
	}
}

// send a notification on the internal notification channel
func (m Capabilities) notify(c capability) {
	if m.changeC != nil {
		m.changeC <- c
	}
}

// get a capability bitvector for the module
// the first two bytes returned are id and length, the following bytes contain the flags
// returns nil if id is not registered
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

// adds a capability module to the bitvector collection
// does not protect against duplicate ids. Calling code should use registerModule instead
func (c *Capabilities) add(cap capability) {
	c.Flags = append(c.Flags, cap)
}

// sets bits on bitvector
// fails if:
// * argument is 0-length byte
// * capability module id is unknown
// * the underlying set fails (see capability.set)
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

// unsets bits on bitvector
// fails if:
// * argument is 0-length byte
// * capability module id is unknown
// * the underlying set fails (see capability.set)
func (c *Capabilities) unset(id uint8, flags []byte) error {
	if len(flags) == 0 {
		return errors.New("flag bytes cannot be empty")
	}
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

// adds a new module to the collection.
// length is the length of the bitvector
// fails if id already exists
func (c *Capabilities) registerModule(id uint8, length uint8) error {
	cap := c.get(id)
	if cap != nil {
		return fmt.Errorf("capability %d already registered", id)
	}
	cap = newCapability(id, length)
	c.add(cap)
	return nil
}

// capability is a single bitvector capabilities module
type capability []byte

// initializes a capability type
// sets it two first bytes to id and length
func newCapability(code uint8, byteLength uint8) capability {
	c := make(capability, byteLength+2)
	c[0] = code
	c[1] = byteLength
	return c
}

// sets bits on bitvector
// fails if length of input flags don't match length in capability type
func (c *capability) set(flag []byte) error {
	if !c.validLength(flag) {
		return fmt.Errorf("Bitfield must be %d bytes long", len(*c))
	}
	for i, b := range flag {
		(*c)[2+i] |= b
	}
	return nil
}

// unsets bits on bitvector
// fails if length of input flags don't match length in capability type
func (c *capability) unset(flag []byte) error {
	if !c.validLength(flag) {
		return fmt.Errorf("Bitfield must be %d bytes long", len(*c))
	}
	for i, b := range flag {
		(*c)[2+i] &= ^b
	}
	return nil
}

// validate flag input against capability type length
func (c *capability) validLength(flag []byte) bool {
	return len(flag) == len(*c)-2
}

func (c *capability) String() string {
	return fmt.Sprintf("%02x:%v", (*c)[0], pot.ToBin((*c)[2:]))
}

type CapabilitiesMsg []capability

func (c Capabilities) toMsg() CapabilitiesMsg {
	var m CapabilitiesMsg
	for _, f := range c.Flags {
		entry := make([]byte, len(f))
		copy(entry, f)
		m = append(m, entry)
	}
	return m
}

func (m CapabilitiesMsg) fromMsg() Capabilities {
	c := Capabilities{}
	for _, entry := range m {
		f := capability(make([]byte, len(entry)))
		copy(f, entry)
		c.add(f)
	}
	return c
}

func (m CapabilitiesMsg) String() string {
	return capArrayToString(m)
}

func capArrayToString(b []capability) string {
	var caps []string
	for _, cap := range b {
		caps = append(caps, cap.String())
	}
	return strings.Join(caps, ",")
}
