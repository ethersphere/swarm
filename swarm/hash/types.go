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
	"fmt"
	"hash"
	"sync"
)

const (
	DefaultHasherCount = 32 // default number of workers to instantiate in a pool upon creation
)

var (
	defaultHash string
	pools       map[string]*Hasher
	initFunc    sync.Once
)

// the basic swarm hash interface
// an instance represents an individual worker in a pool
type Hash interface {
	hash.Hash
	ResetWithLength([]byte)
}

// function for creating a new individual worker
type HashFunc func() Hash

// Hasher maintains a pool of workers
// It is the entrypoint for invoking hashing jobs
// And to query metadata about the type of worker in its pool
type Hasher struct {
	typ      string
	hashFunc func() Hash
	pool     sync.Pool
	size     int
}

func init() {
	pools = make(map[string]*Hasher)
}

// Init sets the default hash to be used when calling GetHash
// This must be executed before any hashing can commence
// If no hash functions is registered with the identifier, the function will panic
func Init(typ string, size int) {
	initFunc.Do(func() {
		if defaultHash != "" {
			panic("cannot change default hash after initialization")
		} else if _, ok := pools[typ]; !ok {
			panic(fmt.Sprintf("hash %s not registered", typ))
		}
		defaultHash = typ
		setMultihashParamsByName(typ, size)
	})
}

// Add creates a new pool for the provided hash function, with the identifier typ
// hasherCount indicates the number of workers to maintain in the pool. If zero, the default value will be used
// hashFunc is the function to be used to spawn workers
func Add(typ string, hasherCount int, hashFunc HashFunc) error {
	if defaultHash != "" {
		panic("cannot add hash after initialization")
	} else if _, ok := pools[typ]; ok {
		panic(fmt.Sprintf("hash %s already registered", typ))
	}
	if hasherCount == 0 {
		hasherCount = DefaultHasherCount
	}
	h := &Hasher{
		typ: typ,
		pool: sync.Pool{
			New: func() interface{} {
				return hashFunc()
			},
		},
	}
	for i := 0; i < hasherCount; i++ {
		hf := hashFunc()
		if h.size == 0 {
			h.size = hf.Size()
		}
		h.pool.Put(hf)
	}
	pools[typ] = h
	return nil
}

// GetHash returns the hasher pool for a specified hash type
func GetHash() *Hasher {
	return pools[defaultHash]
}

// GetHashByName returns the hasher pool identified by typ
// If the hash entry does not exist in the pool, nil is returned
func GetHashByName(typ string) *Hasher {
	return pools[typ]
}

// backend function for GetHashSize*()
func getHashLength(typ string) int {
	if _, ok := pools[typ]; !ok {
		return 0
	}
	return pools[typ].Size()
}

// GetHashSize returns the digest size of the hash set as default
// It returns 0 if the default hash is not set
func GetHashSize() int {
	return getHashLength(defaultHash)
}

// GetHashSizeByName returns the digets size of the hash identified by typ
// It returns 0 if the hash pool identifier doesn't exist
func GetHashSizeByName(typ string) int {
	return getHashLength(typ)
}

// Returns the digest size the hashers of the hash pool will produce
func (h *Hasher) Size() int {
	return h.size
}

// Hash creates a digest from the registered hash function
// It retrieves a hash instance from the pool, resets it, sequentially writes the provided data to it, and returns the digest
func (h *Hasher) Hash(data ...[]byte) []byte {
	hasher := h.pool.Get().(Hash)
	defer h.pool.Put(hasher)
	hasher.Reset()
	return h.doHash(hasher, data)
}

// HashWithLength does the same as Hash, but first calls ResetWithLength with the provided "length" parameter on the underlying hash
func (h *Hasher) HashWithLength(length []byte, data ...[]byte) []byte {
	hasher := h.pool.Get().(Hash)
	defer h.pool.Put(hasher)
	hasher.ResetWithLength(length)
	return h.doHash(hasher, data)
}

// backend function for Hash*()
func (h *Hasher) doHash(hasher Hash, data [][]byte) []byte {
	for _, d := range data {
		hasher.Write(d)
	}
	return hasher.Sum(nil)
}
