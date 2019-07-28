package main

import (
	"encoding/binary"
	"fmt"
)

type capabilities []Cap

type CapabilityId uint64

type Cap struct {
	id  CapabilityId
	cap []bool
}

func NewCap(bitCount int, id CapabilityId) Cap {
	return Cap{
		id:  id,
		cap: make([]bool, bitCount),
	}
}

// implements encoding/BinaryMarshaler interface
func (c Cap) MarshalBinary() ([]byte, error) {

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
func (c Cap) UnmarshalBinary(s []byte) error {

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

func (c *Cap) String() (s string) {
	for _, b := range c.cap {
		if b {
			s += "1"
		} else {
			s += "0"
		}
	}
	return s
}

func (c *capabilities) String() {

}

func main() {
	c := NewCap(9, 42)
	fmt.Println(c)
	c.cap[1] = true
	c.cap[8] = true
	fmt.Println(c)
	m, _ := c.MarshalBinary()
	fmt.Println(m)
	err := c.UnmarshalBinary(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	fmt.Println(c.String())
}
