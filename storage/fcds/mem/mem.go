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

package mem

import (
	"sync"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage/fcds"
)

var _ fcds.MetaStore = new(MetaStore)

// MetaStore is the simplest in-memory implementation of FCDS MetaStore.
// It is meant to be used as the reference implementation.
type MetaStore struct {
	meta map[string]*fcds.Meta
	free map[uint8]map[int64]struct{}
	mtx  sync.RWMutex
}

// NewMetaStore constructs a new MetaStore.
func NewMetaStore() (s *MetaStore) {
	free := make(map[uint8]map[int64]struct{})
	for shard := uint8(0); shard < 255; shard++ {
		free[shard] = make(map[int64]struct{})
	}
	return &MetaStore{
		meta: make(map[string]*fcds.Meta),
		free: free,
	}
}

// Get returns chunk meta information.
func (s *MetaStore) Get(addr chunk.Address) (m *fcds.Meta, err error) {
	s.mtx.RLock()
	m = s.meta[string(addr)]
	s.mtx.RUnlock()
	if m == nil {
		return nil, chunk.ErrChunkNotFound
	}
	return m, nil
}

// Set adds a new chunk meta information for a shard.
// Reclaimed flag denotes that the chunk is at the place of
// already deleted chunk, not appended to the end of the file.
func (s *MetaStore) Set(addr chunk.Address, shard uint8, reclaimed bool, m *fcds.Meta) (err error) {
	s.mtx.Lock()

	if reclaimed {
		delete(s.free[shard], m.Offset)
	}

	s.meta[string(addr)] = m
	s.mtx.Unlock()
	return nil
}

// Remove removes chunk meta information from the shard.
func (s *MetaStore) Remove(addr chunk.Address, shard uint8) (err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	key := string(addr)
	m := s.meta[key]
	if m == nil {
		return chunk.ErrChunkNotFound
	}
	s.free[shard][m.Offset] = struct{}{}

	delete(s.meta, key)
	return nil
}

// ShardSlots gives back a slice of ShardInfo items that represent the number
// of free slots inside each shard.
func (s *MetaStore) ShardSlots() (freeSlots []fcds.ShardInfo) {
	freeSlots = make([]fcds.ShardInfo, fcds.ShardCount)

	s.mtx.RLock()
	for i := uint8(0); i < fcds.ShardCount; i++ {
		slot := fcds.ShardInfo{Shard: i}
		if slots, ok := s.free[i]; ok {
			slot.Val = int64(len(slots))
		}
		freeSlots[i] = slot
	}
	s.mtx.RUnlock()

	return freeSlots
}

// FreeOffset returns an offset that can be reclaimed by
// another chunk. If the returned value is less then 0
// there are no free offset at this shard.
func (s *MetaStore) FreeOffset(shard uint8) (offset int64, err error) {
	s.mtx.RLock()
	for o := range s.free[shard] {
		s.mtx.RUnlock()
		return o, nil
	}
	s.mtx.RUnlock()
	return -1, nil
}

func (s *MetaStore) FastFreeOffset() (uint8, int64, func(), error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for shard, offsets := range s.free {
		for o, _ := range offsets {
			if o >= 0 {
				o := o
				// remove from free offset map, create cancel func, return all values

				delete(offsets, o)
				return shard, o, func() {
					s.mtx.Lock()
					defer s.mtx.Unlock()
					s.free[shard][o] = struct{}{}
				}, nil
			} else {
				panic("wtf mem")
			}
		}
	}

	return 0, -1, func() {}, nil

}

// Count returns a number of chunks in MetaStore.
func (s *MetaStore) Count() (count int, err error) {
	s.mtx.RLock()
	count = len(s.meta)
	s.mtx.RUnlock()
	return count, nil
}

// Iterate iterates over all chunk meta information.
func (s *MetaStore) Iterate(fn func(chunk.Address, *fcds.Meta) (stop bool, err error)) (err error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for a, m := range s.meta {
		stop, err := fn(chunk.Address(a), m)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}
	return nil
}

// Close doesn't do anything.
// It exists to implement fcdb.MetaStore interface.
func (s *MetaStore) Close() (err error) {
	return nil
}
