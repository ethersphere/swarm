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

package encryption

import (
	"hash"
)

type chunkEncryption struct {
	hashFunc func() hash.Hash
	size     int
}

func (c *chunkEncryption) Encrypt(chunk Chunk) (Key, error) {
	// TODO: implement
	return nil, nil
}

func (c *chunkEncryption) Decrypt(chunk Chunk) error {
	// TODO: implement
	return nil
}

func (c *chunkEncryption) transform(chunk Chunk, key Key) {
	transformedData := c.transformBytes(chunk.Data(), key)
	chunk.SetData(transformedData)
	//TODO: transform span too
}

func (c *chunkEncryption) transformBytes(bytes []byte, key Key) []byte {
	length := len(bytes)
	transformedBytes := make([]byte, length)
	hasher := c.hashFunc()
	var ctr uint8
	for i := 0; i < length; i += c.size {
		hasher.Write(key)
		hasher.Write([]byte{ctr})
		ctrHash := hasher.Sum(nil)
		hasher.Reset()
		hasher.Write(ctrHash)
		segmentKey := hasher.Sum(nil)
		hasher.Reset()
		for j := 0; j < c.size; j++ {
			transformedBytes[i+j] = bytes[i+j] ^ segmentKey[j]
		}
		ctr++
	}
	return transformedBytes
}
