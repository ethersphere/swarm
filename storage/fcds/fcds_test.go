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
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethersphere/swarm/chunk"
	chunktesting "github.com/ethersphere/swarm/chunk/testing"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/testutil"
)

func init() {
	testutil.Init()
}

func TestStoreGrow(t *testing.T) {
	path, err := ioutil.TempDir("", "swarm-fcds")
	if err != nil {
		t.Fatal(err)
	}
	defer func(sc uint8) {
		ShardCount = sc
	}(ShardCount)

	ShardCount = 16
	capacity := 100000
	gcTarget := 7000
	insert := 10000000000

	s, err := New(path, chunk.DefaultSize, newMetaStore(), WithCache(false))
	if err != nil {
		os.RemoveAll(path)
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		os.RemoveAll(path)
	}()
	inserted := 0
	gcRuns := 0
	var mtx sync.Mutex
	sem := make(chan struct{}, 1)
	for i := 0; i < insert; i++ {
		ch := chunktesting.GenerateTestRandomChunk()
		err = s.Put(ch)
		if err != nil {
			t.Fatal(err)
		}
		mtx.Lock()
		inserted++
		mtx.Unlock()
		if inserted >= capacity {
			select {
			case sem <- struct{}{}:
				gcRuns++
				go func() {
					defer func() {
						time.Sleep(1 * time.Second)
						<-sem
					}()
					count := 0
					var wg sync.WaitGroup
					err := s.Iterate(func(c chunk.Chunk) (stop bool, err error) {
						count++
						wg.Add(1)
						go func(c chunk.Address) {
							defer wg.Done()
							e := s.Delete(c)
							if e != nil {
								fmt.Println("error deleteing", e)
							}
							mtx.Lock()
							inserted--
							mtx.Unlock()
						}(c.Address())
						if count >= gcTarget {
							return true, nil
						}
						return false, nil
					})
					if err != nil {
						fmt.Println("iterator err", err)
					}
					wg.Wait()
				}()
			default:

			}

		}
		if i%10000 == 0 {
			mtx.Lock()
			ss := getShardsSum(s.shards)
			ssmb := ss / (1024 * 1024)
			insertedmb := i * 4096 / (1024 * 1024)
			expectedSum := capacity * 4096 / (1024 * 1024)

			fmt.Println("inserted", i, "insertedMB", insertedmb, "expectedSum", expectedSum, "shardsum", ss, "mb", ssmb, "gcruns", gcRuns)
			mtx.Unlock()
		}
	}

}

func getShardsSum(s []shard) int {
	sum := 0

	for _, sh := range s {
		v, err := sh.f.Stat()
		if err != nil {
			panic(err)
		}
		//log.Error("summing", "i", i, "sum", sum, "size", v.Size())
		sum += int(v.Size())
	}

	return sum
}

type metaStore struct {
	meta map[string]*Meta
	free map[uint8]map[int64]struct{}
	mu   sync.RWMutex
}

// NewMetaStore constructs a new MetaStore.
func newMetaStore() (s *metaStore) {
	free := make(map[uint8]map[int64]struct{})
	for shard := uint8(0); shard < 255; shard++ {
		free[shard] = make(map[int64]struct{})
	}
	return &metaStore{
		meta: make(map[string]*Meta),
		free: free,
	}
}

// Get returns chunk meta information.
func (s *metaStore) Get(addr chunk.Address) (m *Meta, err error) {
	s.mu.RLock()
	m = s.meta[string(addr)]
	s.mu.RUnlock()
	if m == nil {
		return nil, chunk.ErrChunkNotFound
	}
	return m, nil
}

// Set adds a new chunk meta information for a shard.
// Reclaimed flag denotes that the chunk is at the place of
// already deleted chunk, not appended to the end of the file.
func (s *metaStore) Set(addr chunk.Address, shard uint8, reclaimed bool, m *Meta) (err error) {
	s.mu.Lock()
	if reclaimed {
		delete(s.free[shard], m.Offset)
	}
	s.meta[string(addr)] = m
	s.mu.Unlock()
	return nil
}

// Remove removes chunk meta information from the shard.
func (s *metaStore) Remove(addr chunk.Address, shard uint8) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := string(addr)
	m := s.meta[key]
	if m == nil {
		panic(0)
		return chunk.ErrChunkNotFound
	}
	log.Error("setting offset as free", "offset", m.Offset, "shard", shard)
	if _, v := s.free[shard][m.Offset]; v == true {
		panic(v)
	}
	s.free[shard][m.Offset] = struct{}{}
	delete(s.meta, key)
	return nil
}

// FreeOffset returns an offset that can be reclaimed by
// another chunk. If the returned value is less then 0
// there are no free offset at this shard.
func (s *metaStore) FreeOffset(shard uint8) (offset int64, err error) {
	s.mu.RLock()
	//spew.Dump("free slots", s.free[shard])
	for o := range s.free[shard] {
		s.mu.RUnlock()
		return o, nil
	}
	s.mu.RUnlock()
	return -1, nil
}

// Count returns a number of chunks in MetaStore.
func (s *metaStore) Count() (count int, err error) {
	s.mu.RLock()
	count = len(s.meta)
	s.mu.RUnlock()
	return count, nil
}

// Iterate iterates over all chunk meta information.
func (s *metaStore) Iterate(fn func(chunk.Address, *Meta) (stop bool, err error)) (err error) {
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

// Close doesn't do anything.
// It exists to implement fcdb.MetaStore interface.
func (s *metaStore) Close() (err error) {
	return nil
}
