package network

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

type Capabilities struct {
	idx  map[CapabilityId]int
	caps []Capability
	mu   sync.Mutex
}

func NewCapabilities() Capabilities {
	c := Capabilities{
		idx: make(map[CapabilityId]int),
	}
	return c
}

type CapabilityId uint64

type Capability struct {
	id  CapabilityId
	cap []bool
}

func NewCapability(id CapabilityId, bitCount int) Capability {
	return Capability{
		id:  id,
		cap: make([]bool, bitCount),
	}
}

func (c *Capability) Set(idx int) error {
	l := len(c.cap)
	if idx > l-1 {
		return fmt.Errorf("index %d out of bounds (len=%d)", idx, l)
	}
	c.cap[idx] = true
	return nil
}

func (c *Capability) Unset(idx int) error {
	l := len(c.cap)
	if idx > l-1 {
		return fmt.Errorf("index %d out of bounds (len=%d)", idx, l)
	}
	c.cap[idx] = false
	return nil
}

// implements encoding/BinaryMarshaler interface
func (c Capability) MarshalBinary() ([]byte, error) {

	// serialize bit vector length
	l := make([]byte, 8)
	csz := len(c.cap)
	lsz := binary.PutUvarint(l, uint64(csz))

	// serialize cap id
	id := make([]byte, 8)
	idsz := binary.PutUvarint(id, uint64(c.id))

	// create storage array (size rounded up to nearest byte threshold)
	s := make([]byte, csz/8+1+lsz+idsz)

	// prefix with length
	idx := 0
	copy(s[idx:], l[:lsz])

	idx += lsz
	copy(s[idx:], id[:idsz])

	// iterate all bit flags and set bits in data portion of array accordingly
	idx += idsz
	for i, b := range c.cap {
		if b {
			ri := uint8(7 - (i % 8))
			s[idx+i/8] |= 1 << ri
		}
	}

	return s, nil
}

// implements encoding/BinaryUnmarshaler interface
func (c Capability) UnmarshalBinary(s []byte) error {

	// retrieve the bit vector length
	idx := 0
	csz, lsz := binary.Uvarint(s[idx:])

	idx += lsz
	id, idsz := binary.Uvarint(s[idx:])

	idx += idsz

	// audit that enough bytes exist in the data to contain the indicated bit vector length
	tsz := csz/8 + 1 + uint64(lsz) + uint64(idsz)
	if tsz != uint64(len(s)) {
		return fmt.Errorf("wrong data length. expected length prefix %d bytes + bit vector length %d bits = %d bytes, got %d bytes", lsz, csz, tsz, len(s))
	}

	// iterate all the bit flags and set the bools in the capability object accordingly
	c.id = CapabilityId(id)
	c.cap = make([]bool, csz)
	for i := uint64(0); i < csz; i++ {
		ri := uint(7 - (i % 8))
		if s[i/8+uint64(idx)]&1<<ri > 0 {
			c.cap[i] = true
		}
	}
	return nil
}

func (c Capability) String() (s string) {
	for _, b := range c.cap {
		if b {
			s += "1"
		} else {
			s += "0"
		}
	}
	return s
}

func (c *Capabilities) add(cp Capability) error {
	if _, ok := c.idx[cp.id]; ok {
		return fmt.Errorf("Capability id %d already registered", cp.id)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.caps = append(c.caps, cp)
	c.idx[cp.id] = len(c.caps) - 1
	return nil
}

func (c Capabilities) get(id CapabilityId) Capability {
	fmt.Fprintf(os.Stderr, "id %v idx %v", id, c.idx[id])
	return c.caps[c.idx[id]]
}

func (c Capabilities) String() (s string) {
	for _, cp := range c.caps {
		if s != "" {
			s += ","
		}
		s += cp.String()
	}
	return s
}

func _main() {
	c := NewCapability(42, 9)
	fmt.Println(c)
	c.Set(1)
	c.Set(8)
	fmt.Println(c)
	m, _ := c.MarshalBinary()
	fmt.Println(m)
	err := c.UnmarshalBinary(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	fmt.Println(c.String())

	var cs Capabilities
	cs.add(c)

	c2 := NewCapability(157, 3)
	c2.Set(2)
	cs.add(c2)

	fmt.Printf("%s\n", cs)
}
