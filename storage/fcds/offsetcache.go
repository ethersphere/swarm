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

package fcds

import (
	"sync"
	"time"
)

// offsetCache is a simple cache of offset integers
// by shard files.
type offsetCache struct {
	m        map[uint8]map[int64]time.Time
	ttl      time.Duration
	mu       sync.RWMutex
	quit     chan struct{}
	quitOnce sync.Once
}

// newOffsetCache constructs offsetCache for a fixed number of shards.
func newOffsetCache(shardCount uint8, ttl time.Duration) (c *offsetCache) {
	m := make(map[uint8]map[int64]time.Time)
	for i := uint8(0); i < shardCount; i++ {
		m[i] = make(map[int64]time.Time)
	}
	c = &offsetCache{
		m:    m,
		quit: make(chan struct{}),
	}
	if ttl > 0 {
		go c.cleanup(30 * time.Second)
	}
	return c
}

// get returns a free offset in a shard. If the returned
// value is less then 0, there are no free offset in that
// shard.
func (c *offsetCache) get(shard uint8) (offset int64) {
	c.mu.RLock()
	for o := range c.m[shard] {
		c.mu.RUnlock()
		return o
	}
	c.mu.RUnlock()
	return -1
}

// set sets a free offset for a shard file.
func (c *offsetCache) set(shard uint8, offset int64) {
	c.mu.Lock()
	c.m[shard][offset] = time.Now().Add(c.ttl)
	c.mu.Unlock()
}

// remove removes a free offset for a shard file.
func (c *offsetCache) remove(shard uint8, offset int64) {
	c.mu.Lock()
	delete(c.m[shard], offset)
	c.mu.Unlock()
}

// close stops parallel processing created
// by offsetCache.
func (c *offsetCache) close() {
	c.quitOnce.Do(func() {
		close(c.quit)
	})
}

func (c *offsetCache) cleanup(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for _, s := range c.m {
				for offset, expiration := range s {
					if now.After(expiration) {
						delete(s, offset)
					}
				}
			}
			c.mu.Unlock()
		case <-c.quit:
			return
		}
	}
}
