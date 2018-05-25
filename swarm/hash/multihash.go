// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package swarmhash

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	defaultMultihashTypeCode = 0x1b
)

var (
	multihashTypeCode uint8
)

// if first byte is the start of a multihash this function will try to parse it
// if successful it returns the length of multihash data, 0 otherwise
func GetLength(data []byte) (int, error) {
	cursor := 0
	typ, c := binary.Uvarint(data)
	if c <= 0 {
		return 0, fmt.Errorf("unreadable hashtype field")
	}
	if !isSwarmMultihashType(uint8(typ)) {
		return 0, fmt.Errorf("hash code %x is not a swarm hashtype", typ)
	}
	cursor += c
	hashlength, c := binary.Uvarint(data[cursor:])
	if c <= 0 {
		return 0, fmt.Errorf("unreadable length field")
	}
	cursor += c

	// we cheekily assume hashlength < maxint
	inthashlength := int(hashlength)
	if len(data[c:]) < inthashlength {
		return 0, fmt.Errorf("length mismatch")
	}
	return inthashlength, nil
}

func FromMultihash(data []byte) ([]byte, error) {
	hashLength, err := GetLength(data)
	if err != nil {
		return nil, err
	}
	return data[len(data)-hashLength:], nil
}

func NewMultihash(data []byte) ([]byte, error) {
	h := GetHash()
	if h == nil {
		return nil, fmt.Errorf("hasher for %s not initialized", defaultHash)
	}
	hs := h.Hash(data)
	return ToMultihash(hs), nil
}

func ToMultihash(hashData []byte) []byte {
	buf := bytes.NewBuffer(nil)
	b := make([]byte, 8)
	c := binary.PutUvarint(b, uint64(multihashTypeCode))
	buf.Write(b[:c])
	c = binary.PutUvarint(b, uint64(len(hashData)))
	buf.Write(b[:c])
	buf.Write(hashData)
	return buf.Bytes()
}

func isSwarmMultihashType(code uint8) bool {
	return code == multihashTypeCode
}
