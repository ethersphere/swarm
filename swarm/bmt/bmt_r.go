// Copyright 2017 The go-ethereum Authors
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

// simple nonconcurrent reference implementation for hashsize segment based
// Binary Merkle tree hash on arbitrary but fixed maximum chunksize
//
// This implementation does not take advantage of any paralellisms and uses
// far more memory than necessary, but it is easy to see that it is correct.
// It can be used for generating test cases for optimized implementations.
// see testBMTHasherCorrectness function in bmt_test.go
package bmt

import (
	"hash"
)

// RefHasher is the non-optimized easy-to-read reference implementation of BMT
type RefHasher struct {
	maxSize int // c * 32, where c = 2 ^ ceil(log2(count)), where count = ceil(length / 32)
	section int // 64
	hasher  hash.Hash
}

// NewRefHasher returns a new RefHasher
func NewRefHasher(hasher BaseHasher, count int) *RefHasher {
	h := hasher()
	hashsize := h.Size()
	c := 2
	for ; c < count; c *= 2 {
	}
	return &RefHasher{
		section: 2 * hashsize,
		maxSize: c * hashsize,
		hasher:  h,
	}
}

// Hash returns the BMT hash of the byte slice
// implements the SwarmHash interface
func (rh *RefHasher) Hash(d []byte) []byte {
	data := make([]byte, rh.maxSize)
	l := len(d)
	if l > rh.maxSize {
		l = rh.maxSize
	}
	copy(data, d[:l])
	return rh.hash(data, rh.maxSize)
}

func (rh *RefHasher) hash(d []byte, l int) []byte {
	var section []byte
	if l == rh.section {
		section = d
	} else {
		l /= 2
		section = append(rh.hash(d[:l], l), rh.hash(d[l:], l)...)
	}
	rh.hasher.Reset()
	rh.hasher.Write(section)
	return rh.hasher.Sum(nil)
}
