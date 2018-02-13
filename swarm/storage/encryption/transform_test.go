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
	"crypto"
	"fmt"
	"hash"
	"testing"
)

func TestChunkEncryption(t *testing.T) {
	chunkEncryption := &chunkEncryption{
		hashFunc: func() hash.Hash { return crypto.SHA256.New() },
		size:     4096,
	}

	bytes := make([]byte, 4096)

	key := make([]byte, 256)

	transformed := chunkEncryption.transformBytes(bytes, key)
	fmt.Printf("Transformed bytes: %v", transformed)
}
