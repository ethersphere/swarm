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

type MetaStore struct {
	meta map[string]*fcds.Meta
	free map[uint8]map[int64]struct{}
	mu   sync.RWMutex
}

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

func (s *MetaStore) Get(addr chunk.Address) (m *fcds.Meta, err error) {
	s.mu.RLock()
	m = s.meta[string(addr)]
	s.mu.RUnlock()
	if m == nil {
		return nil, chunk.ErrChunkNotFound
	}
	return m, nil
}

func (s *MetaStore) Set(addr chunk.Address, shard uint8, reclaimed bool, m *fcds.Meta) (err error) {
	s.mu.Lock()
	if reclaimed {
		delete(s.free[shard], m.Offset)
	}
	s.meta[string(addr)] = m
	s.mu.Unlock()
	return nil
}

func (s *MetaStore) Remove(addr chunk.Address, shard uint8) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := string(addr)
	m := s.meta[key]
	if m == nil {
		return chunk.ErrChunkNotFound
	}
	s.free[shard][m.Offset] = struct{}{}
	delete(s.meta, key)
	return nil
}

func (s *MetaStore) FreeOffset(shard uint8) (offset int64, err error) {
	s.mu.RLock()
	for o := range s.free[shard] {
		s.mu.RUnlock()
		return o, nil
	}
	s.mu.RUnlock()
	return -1, nil
}

func (s *MetaStore) Count() (count int, err error) {
	s.mu.RLock()
	count = len(s.meta)
	s.mu.RUnlock()
	return count, nil
}

func (s *MetaStore) Iterate(fn func(chunk.Address, *fcds.Meta) (stop bool, err error)) (err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
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

func (s *MetaStore) Close() (err error) {
	return nil
}
