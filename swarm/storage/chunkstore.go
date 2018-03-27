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

package storage

/*
ChunkStore interface is implemented by :

- MemStore: a memory cache
- DbStore: local disk/db store
- LocalStore: a combination (sequence of) memStore and dbStore
- NetStore: cloud storage abstraction layer
- DPA: local requests for swarm storage and retrieval
- FakeChunkStore: dummy store which doesn't store anything just implements the interface
*/
type ChunkStore interface {
	Put(*Chunk) // effectively there is no error even if there is an error
	Get(Key) (*Chunk, error)
	Close()
}

// FakeChunkStore doesn't store anything, just implements the ChunkStore interface
// It can be used to inject into a hasherStore if you don't want to actually store data just do the
// hashing
type FakeChunkStore struct {
}

// Put doesn't store anything it is just here to implement ChunkStore
func (d *FakeChunkStore) Put(*Chunk) {
}

// Gut doesn't store anything it is just here to implement ChunkStore
func (d *FakeChunkStore) Get(Key) (*Chunk, error) {
	return nil, nil
}

// Close doesn't store anything it is just here to implement ChunkStore
func (d *FakeChunkStore) Close() {
}
