// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package mock

import (
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage/fcds"
	"github.com/ethersphere/swarm/storage/mock"
)

var _ fcds.Storer = new(Store)

// Store implements FCDS Interface by using mock
// store for persistence.
type Store struct {
	m *mock.NodeStore
}

// New returns a new store with mock NodeStore
// for storing Chunk data.
func New(m *mock.NodeStore) (s *Store) {
	return &Store{
		m: m,
	}
}

// Get returns a chunk with data.
func (s *Store) Get(addr chunk.Address) (c chunk.Chunk, err error) {
	data, err := s.m.Get(addr)
	if err != nil {
		if err == mock.ErrNotFound {
			return nil, chunk.ErrChunkNotFound
		}
		return nil, err
	}
	return chunk.NewChunk(addr, data), nil
}

// Has returns true if chunk is stored.
func (s *Store) Has(addr chunk.Address) (yes bool, err error) {
	_, err = s.m.Get(addr)
	if err != nil {
		if err == mock.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Put stores chunk data.
func (s *Store) Put(ch chunk.Chunk) (shard uint8, err error) {
	err = s.m.Put(ch.Address(), ch.Data())
	return 0, err
}

// Delete removes chunk data.
func (s *Store) Delete(addr chunk.Address) (err error) {
	return s.m.Delete(addr)
}

// Count returns a number of stored chunks.
func (s *Store) Count() (count int, err error) {
	var startKey []byte
	for {
		keys, err := s.m.Keys(startKey, 0)
		if err != nil {
			return 0, err
		}
		count += len(keys.Keys)
		if keys.Next == nil {
			break
		}
		startKey = keys.Next
	}
	return count, nil
}

// Iterate iterates over stored chunks in no particular order.
func (s *Store) Iterate(fn func(chunk.Chunk) (stop bool, err error)) (err error) {
	var startKey []byte
	for {
		keys, err := s.m.Keys(startKey, 0)
		if err != nil {
			return err
		}
		for _, addr := range keys.Keys {
			data, err := s.m.Get(addr)
			if err != nil {
				return err
			}
			stop, err := fn(chunk.NewChunk(addr, data))
			if err != nil {
				return err
			}
			if stop {
				return nil
			}
		}
		if keys.Next == nil {
			break
		}
		startKey = keys.Next
	}
	return nil
}

func (s *Store) ShardSize() (slots []fcds.ShardSize, err error) {
	i, err := s.Count()
	return []fcds.ShardSize{fcds.ShardSize{Shard: 0, Size: int64(i)}}, err
}

// Close doesn't do anything.
// It exists to implement fcdb.MetaStore interface.
func (s *Store) Close() error {
	return nil
}
