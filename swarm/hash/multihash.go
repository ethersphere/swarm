// Copyright 2018 The go-ethereum Authors
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

// hook used by the master Init() to set the corresponding mulithash code
// currently we naïvely set KECCAK256 (0x1b)
func setMultihashCodeByName(typ string) {
	multihashTypeCode = defaultMultihashTypeCode
}

// check if valid swarm multihash
func isSwarmMultihashType(code uint8) bool {
	return code == multihashTypeCode
}

// GetLength returns the digest length of the provided multihash
// It will fail if the multihash is not a valid swarm mulithash
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

// NewMultihash hashes the provided data with the default hash function and wraps it as a mulithash
// It will fail if the hasher is not initialized
func NewMultihash(data []byte) ([]byte, error) {
	h := GetHash()
	if h == nil {
		return nil, fmt.Errorf("hasher for %s not initialized", defaultHash)
	}
	hs := h.Hash(data)
	return ToMultihash(hs), nil
}

// FromMulithash returns the digest portion of the multihash
// It will fail if the multihash is not a valid swarm multihash
func FromMultihash(data []byte) ([]byte, error) {
	hashLength, err := GetLength(data)
	if err != nil {
		return nil, err
	}
	return data[len(data)-hashLength:], nil
}

// ToMulithash wraps the provided digest data with a swarm mulithash header
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
