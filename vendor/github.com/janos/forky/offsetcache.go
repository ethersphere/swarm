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

package forky

import "sync"

type offsetCache struct {
	m  map[uint8]map[int64]struct{}
	mu sync.RWMutex
}

func newOffsetCache(shardCount uint8) (c *offsetCache) {
	m := make(map[uint8]map[int64]struct{})
	for i := uint8(0); i < shardCount; i++ {
		m[i] = make(map[int64]struct{})
	}
	return &offsetCache{
		m: m,
	}
}

func (c *offsetCache) get(shard uint8) (offset int64) {
	c.mu.RLock()
	for o := range c.m[shard] {
		c.mu.RUnlock()
		return o
	}
	c.mu.RUnlock()
	return -1
}

func (c *offsetCache) set(shard uint8, offset int64) {
	c.mu.Lock()
	c.m[shard][offset] = struct{}{}
	c.mu.Unlock()
}

func (c *offsetCache) remove(shard uint8, offset int64) {
	c.mu.Lock()
	delete(c.m[shard], offset)
	c.mu.Unlock()
}
