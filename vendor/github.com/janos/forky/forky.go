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

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethersphere/swarm/chunk"
)

type Interface interface {
	Get(addr chunk.Address) (ch chunk.Chunk, err error)
	Has(addr chunk.Address) (yes bool, err error)
	Put(ch chunk.Chunk) (err error)
	Delete(addr chunk.Address) (err error)
	Close() (err error)
}

const shardCount = 32

var ErrDBClosed = errors.New("closed database")

var _ Interface = new(Store)

type Store struct {
	shards    map[uint8]*os.File
	shardsMu  map[uint8]*sync.Mutex
	meta      MetaStore
	free      map[uint8]struct{}
	freeMu    sync.RWMutex
	metaCache *metaCache
	freeCache *offsetCache
	wg        sync.WaitGroup
	quit      chan struct{}
	quitOnce  sync.Once
}

func NewStore(path string, metaStore MetaStore, noCache bool) (s *Store, err error) {
	shards := make(map[byte]*os.File, shardCount)
	shardsMu := make(map[uint8]*sync.Mutex)
	for i := byte(0); i < shardCount; i++ {
		shards[i], err = os.OpenFile(filepath.Join(path, fmt.Sprintf("chunks-%v.db", i)), os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		shardsMu[i] = new(sync.Mutex)
	}
	var (
		metaCache *metaCache
		freeCache *offsetCache
	)
	if !noCache {
		metaCache = newMetaCache()
		freeCache = newOffsetCache(shardCount)
	}
	return &Store{
		shards:    shards,
		shardsMu:  shardsMu,
		meta:      metaStore,
		metaCache: metaCache,
		freeCache: freeCache,
		free:      make(map[uint8]struct{}),
		quit:      make(chan struct{}),
	}, nil
}

func (s *Store) Get(addr chunk.Address) (ch chunk.Chunk, err error) {
	done, err := s.protect()
	if err != nil {
		return nil, err
	}
	defer done()

	mu := s.shardsMu[getShard(addr)]
	mu.Lock()
	defer mu.Unlock()

	m, err := s.getMeta(addr)
	if err != nil {
		return nil, err
	}
	data := make([]byte, m.Size)
	_, err = s.shards[getShard(addr)].ReadAt(data, m.Offset)
	if err != nil {
		return nil, err
	}
	return chunk.NewChunk(addr, data), nil
}

func (s *Store) Has(addr chunk.Address) (yes bool, err error) {
	done, err := s.protect()
	if err != nil {
		return false, err
	}
	defer done()

	mu := s.shardsMu[getShard(addr)]
	mu.Lock()
	defer mu.Unlock()

	m, err := s.getMeta(addr)
	if err != nil {
		if err == chunk.ErrChunkNotFound {
			return false, nil
		}
		return false, err
	}
	if s.metaCache != nil {
		s.metaCache.set(addr, m)
	}
	return true, nil
}

func (s *Store) Put(ch chunk.Chunk) (err error) {
	done, err := s.protect()
	if err != nil {
		return err
	}
	defer done()

	addr := ch.Address()
	shard := getShard(addr)
	f := s.shards[shard]
	data := ch.Data()
	section := make([]byte, chunk.DefaultSize)
	copy(section, data)

	s.freeMu.RLock()
	_, hasFree := s.free[shard]
	s.freeMu.RUnlock()

	var offset int64
	var reclaimed bool
	mu := s.shardsMu[shard]
	mu.Lock()
	if hasFree {
		var freeOffset int64 = -1
		if s.freeCache != nil {
			freeOffset = s.freeCache.get(shard)
		}
		if freeOffset < 0 {
			freeOffset, err = s.meta.FreeOffset(shard)
			if err != nil {
				return err
			}
		}
		if freeOffset < 0 {
			offset, err = f.Seek(0, io.SeekEnd)
			if err != nil {
				mu.Unlock()
				return err
			}
			s.freeMu.Lock()
			delete(s.free, shard)
			s.freeMu.Unlock()
		} else {
			offset, err = f.Seek(freeOffset, io.SeekStart)
			if err != nil {
				mu.Unlock()
				return err
			}
			reclaimed = true
		}
	} else {
		offset, err = f.Seek(0, io.SeekEnd)
		if err != nil {
			mu.Unlock()
			return err
		}
	}
	_, err = f.Write(section)
	if err != nil {
		mu.Unlock()
		return err
	}
	if reclaimed {
		if s.freeCache != nil {
			s.freeCache.remove(shard, offset)
		}
		defer mu.Unlock()
	} else {
		mu.Unlock()
	}
	m := &Meta{
		Size:   uint16(len(data)),
		Offset: offset,
	}
	if s.metaCache != nil {
		s.metaCache.set(addr, m)
	}
	return s.meta.Set(addr, shard, reclaimed, m)
}

func (s *Store) Delete(addr chunk.Address) (err error) {
	done, err := s.protect()
	if err != nil {
		return err
	}
	defer done()

	shard := getShard(addr)
	s.freeMu.Lock()
	s.free[shard] = struct{}{}
	s.freeMu.Unlock()

	mu := s.shardsMu[shard]
	mu.Lock()
	defer mu.Unlock()

	if s.freeCache != nil {
		m, err := s.getMeta(addr)
		if err != nil {
			return err
		}
		s.freeCache.set(shard, m.Offset)
	}
	if s.metaCache != nil {
		s.metaCache.remove(addr)
	}
	return s.meta.Remove(addr, shard)
}

func (s *Store) Close() (err error) {
	s.quitOnce.Do(func() {
		close(s.quit)
	})

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
	}

	for _, f := range s.shards {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return s.meta.Close()
}

func (s *Store) protect() (done func(), err error) {
	select {
	case <-s.quit:
		return nil, ErrDBClosed
	default:
	}
	s.wg.Add(1)
	return s.wg.Done, nil
}

func (s *Store) getMeta(addr chunk.Address) (m *Meta, err error) {
	if s.metaCache != nil {
		m = s.metaCache.get(addr)
		if m != nil {
			return m, nil
		}
	}
	m, err = s.meta.Get(addr)
	if err != nil {
		return nil, err
	}
	if s.metaCache != nil {
		s.metaCache.set(addr, m)
	}
	return m, nil
}

func getShard(addr chunk.Address) (shard uint8) {
	return addr[len(addr)-1] % shardCount
}

type MetaStore interface {
	Get(addr chunk.Address) (*Meta, error)
	Set(addr chunk.Address, shard uint8, reclaimed bool, m *Meta) error
	Remove(addr chunk.Address, shard uint8) error
	FreeOffset(shard uint8) (int64, error)
	Close() error
}

type Meta struct {
	Size   uint16
	Offset int64
}

func (m *Meta) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 10)
	binary.BigEndian.PutUint64(data[:8], uint64(m.Offset))
	binary.BigEndian.PutUint16(data[8:10], uint16(m.Size))
	return data, nil
}

func (m *Meta) UnmarshalBinary(data []byte) error {
	m.Offset = int64(binary.BigEndian.Uint64(data[:8]))
	m.Size = binary.BigEndian.Uint16(data[8:10])
	return nil
}

func (m *Meta) String() (s string) {
	if m == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Size: %v, Offset %v}", m.Size, m.Offset)
}
