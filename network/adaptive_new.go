package network

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

// Capabilities is the collection of capabilities for a Swarm node
// It is user both to store the capabilities in the node, and
// to communicate the node capabilities to its peers
type Capabilities struct {
	idx  map[CapabilityId]int
	Caps []*Capability
	mu   sync.Mutex
}

// NewCapabilities initializes a new Capabilities object
func NewCapabilities() *Capabilities {
	return &Capabilities{
		idx: make(map[CapabilityId]int),
	}
}

// CapabilityId defines a unique type of capability
// @justelad concrete enough for ya?
type CapabilityId uint64

// Capability contains a bit vector of flags that define what capability a node has in a specific module
// The module is defined by the Id.
type Capability struct {
	Id  CapabilityId
	Cap []bool
}

// NewCapability initializes a new Capability with the given id and specified number of bits in the vector
func NewCapability(id CapabilityId, bitCount int) *Capability {
	return &Capability{
		Id:  id,
		Cap: make([]bool, bitCount),
	}
}

// Set switches the bit at the specified index on
func (c *Capability) Set(idx int) error {
	l := len(c.Cap)
	if idx > l-1 {
		return fmt.Errorf("index %d out of bounds (len=%d)", idx, l)
	}
	c.Cap[idx] = true
	return nil
}

// Unset switches the bit at the specified index off
func (c *Capability) Unset(idx int) error {
	l := len(c.Cap)
	if idx > l-1 {
		return fmt.Errorf("index %d out of bounds (len=%d)", idx, l)
	}
	c.Cap[idx] = false
	return nil
}

// String implements Stringer interface
func (c Capability) String() (s string) {
	s = fmt.Sprintf("%d:", c.Id)
	for _, b := range c.Cap {
		if b {
			s += "1"
		} else {
			s += "0"
		}
	}
	return s
}

// IsSameAs returns true if the given Capability object has the identical bit settings as the receiver
func (c *Capability) IsSameAs(cp *Capability) bool {
	if cp == nil {
		fmt.Println("cp is nil!!!!")
		return false
	}
	fmt.Printf("is same as: %s %s %d %d", c, cp, len(c.Cap), len(cp.Cap))
	if len(c.Cap) != len(cp.Cap) {
		return false
	}
	for i, b := range cp.Cap {
		if b != c.Cap[i] {
			return false
		}
	}
	return true
}

// adds a capability to the Capabilities collection
// TODO: Following API PR with make this available to client code
func (c *Capabilities) add(cp *Capability) error {
	if _, ok := c.idx[cp.Id]; ok {
		return fmt.Errorf("Capability id %d already registered", cp.Id)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Caps = append(c.Caps, cp)
	c.idx[cp.Id] = len(c.Caps) - 1
	return nil
}

// gets the capability with the specified module id
// returns nil if the id doesn't exist
func (c Capabilities) get(id CapabilityId) *Capability {
	idx, ok := c.idx[id]
	if !ok {
		return nil
	}
	return c.Caps[idx]
}

// String Implements Stringer interface
func (c Capabilities) String() (s string) {
	for _, cp := range c.Caps {
		if s != "" {
			s += ","
		}
		s += cp.String()
	}
	return s
}

// DecodeRLP implements rlp.RLPDecoder
// this custom deserializer builds the module id to array index map
// state of receiver is undefined on error
func (c *Capabilities) DecodeRLP(s *rlp.Stream) error {

	idx := make(map[CapabilityId]int)
	// discard the Capabilities struct list item
	_, err := s.List()
	if err != nil {
		return err
	}

	// retrieve the count of elements in the capability array
	elementCount, err := s.List()
	if err != nil {
		return err
	}

	// counter for the Capabilities.Caps array
	i := 0

	// The last two elements are the terminators of the Capabilities and Capabilities.Caps enclosures
	for elementCount > 2 {

		var cap Capability

		// All elements in array should be Capability type
		// if not throw error
		err := s.Decode(&cap)
		if err != nil {
			return err
		}

		// Add the entry to the Capabilities array
		c.Caps = append(c.Caps, &cap)
		idx[cap.Id] = i

		// elementCount decreases with one per flag plus one for the CapabilityId
		elementCount -= uint64(len(cap.Cap) + 1)

		i++
		fmt.Printf("decoded cap: %v (%d,%d,%v)\n", cap, elementCount, i, cap.Id)
	}
	c.idx = idx
	return nil
}
