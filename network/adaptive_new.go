package network

import (
	"fmt"
	"sync"
)

type Capabilities struct {
	idx  map[CapabilityId]int
	Caps []Capability
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
	Id  CapabilityId
	Cap []bool
}

func NewCapability(id CapabilityId, bitCount int) Capability {
	return Capability{
		Id:  id,
		Cap: make([]bool, bitCount),
	}
}

func (c *Capability) Set(idx int) error {
	l := len(c.Cap)
	if idx > l-1 {
		return fmt.Errorf("index %d out of bounds (len=%d)", idx, l)
	}
	c.Cap[idx] = true
	return nil
}

func (c *Capability) Unset(idx int) error {
	l := len(c.Cap)
	if idx > l-1 {
		return fmt.Errorf("index %d out of bounds (len=%d)", idx, l)
	}
	c.Cap[idx] = false
	return nil
}

//// implements encoding/BinaryMarshaler interface
//func (c Capability) MarshalBinary() ([]byte, error) {
//
//	// serialize bit vector length
//	l := make([]byte, 8)
//	csz := len(c.Cap)
//	lsz := binary.PutUvarint(l, uint64(csz))
//
//	// serialize cap id
//	id := make([]byte, 8)
//	idsz := binary.PutUvarint(id, uint64(c.Id))
//
//	// create storage array (size rounded up to nearest byte threshold)
//	s := make([]byte, csz/8+1+lsz+idsz)
//
//	// prefix with length
//	idx := 0
//	copy(s[idx:], l[:lsz])
//
//	idx += lsz
//	copy(s[idx:], id[:idsz])
//
//	// iterate all bit flags and set bits in data portion of array accordingly
//	idx += idsz
//	for i, b := range c.Cap {
//		if b {
//			ri := uint8(7 - (i % 8))
//			s[idx+i/8] |= 1 << ri
//		}
//	}
//
//	return s, nil
//}
//
//// implements encoding/BinaryUnmarshaler interface
//func (c Capability) UnmarshalBinary(s []byte) error {
//
//	// retrieve the bit vector length
//	idx := 0
//	csz, lsz := binary.Uvarint(s[idx:])
//
//	idx += lsz
//	id, idsz := binary.Uvarint(s[idx:])
//
//	idx += idsz
//
//	// audit that enough bytes exist in the data to contain the indicated bit vector length
//	tsz := csz/8 + 1 + uint64(lsz) + uint64(idsz)
//	if tsz != uint64(len(s)) {
//		return fmt.Errorf("wrong data length. expected length prefix %d bytes + bit vector length %d bits = %d bytes, got %d bytes", lsz, csz, tsz, len(s))
//	}
//
//	// iterate all the bit flags and set the bools in the capability object accordingly
//	c.Id = CapabilityId(id)
//	c.Cap = make([]bool, csz)
//	for i := uint64(0); i < csz; i++ {
//		ri := uint(7 - (i % 8))
//		if s[i/8+uint64(idx)]&1<<ri > 0 {
//			c.Cap[i] = true
//		}
//	}
//	return nil
//}

//func (c *Capability) EncodeRLP(w io.Writer) error {
//	data, err := c.MarshalBinary()
//	if err != nil {
//		return err
//	}
//	return rlp.Encode(w, capabilityRlpHack{B: data})
//}
//
//func (c *Capability) DecodeRLP(s *rlp.Stream) error {
//	fmt.Printf("stream: %v\n\n", s)
//	data, err := s.Bytes()
//	if err != nil {
//		return err
//	}
//	return c.UnmarshalBinary(data)
//}

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

func (c Capability) IsSameAs(cp Capability) bool {
	for i, b := range cp.Cap {
		if b != c.Cap[i] {
			return false
		}
	}
	return true
}

func (c *Capabilities) add(cp Capability) error {
	if _, ok := c.idx[cp.Id]; ok {
		return fmt.Errorf("Capability id %d already registered", cp.Id)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Caps = append(c.Caps, cp)
	c.idx[cp.Id] = len(c.Caps) - 1
	return nil
}

func (c *Capabilities) Add(cp Capability) error {
	return c.add(cp)
}

func (c Capabilities) get(id CapabilityId) Capability {
	return c.Caps[c.idx[id]]
}

func (c Capabilities) String() (s string) {
	for _, cp := range c.Caps {
		if s != "" {
			s += ","
		}
		s += cp.String()
	}
	return s
}
