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

package netsim

import (
	"github.com/ethereum/go-ethereum/p2p/discover"
)

type BucketKey string

func (s *Simulation) ServiceItem(id discover.NodeID, key BucketKey) (value interface{}, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.buckets[id]; !ok {
		return nil, false
	}
	return s.buckets[id].Load(key)
}

func (s *Simulation) SetServiceItem(id discover.NodeID, key BucketKey, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buckets[id].Store(key, value)
}

func (s *Simulation) ServicesItems(key BucketKey) (values []interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.NodeIDs()
	values = make([]interface{}, len(ids))
	for i, id := range ids {
		if _, ok := s.buckets[id]; !ok {
			continue
		}
		if v, ok := s.buckets[id].Load(key); ok {
			values[i] = v
		}
	}
	return values
}

func (s *Simulation) UpServicesItems(key BucketKey) (values []interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.NodeIDs()
	for _, id := range ids {
		if _, ok := s.buckets[id]; !ok {
			continue
		}
		if v, ok := s.buckets[id].Load(key); ok {
			values = append(values, v)
		}
	}
	return values
}
