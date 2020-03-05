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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethersphere/swarm/log"

	"github.com/ethersphere/swarm/chunk"
)

// Storer specifies methods required for FCDS implementation.
// It can be used where alternative implementations are needed to
// switch at runtime.
type Storer interface {
	Get(addr chunk.Address) (ch chunk.Chunk, err error)
	Has(addr chunk.Address) (yes bool, err error)
	Put(ch chunk.Chunk) (err error)
	Delete(addr chunk.Address) (err error)
	Count() (count int, err error)
	Iterate(func(ch chunk.Chunk) (stop bool, err error)) (err error)
	Close() (err error)
}

var _ Storer = new(Store)

// Number of files that store chunk data.
const shardCount = 32

// ErrStoreClosed is returned if store is already closed.
var ErrStoreClosed = errors.New("closed store")

// Store is the main FCDS implementation. It stores chunk data into
// a number of files partitioned by the last byte of the chunk address.
type Store struct {
	shards       []shard        // relations with shard id and a shard file and their mutexes
	meta         MetaStore      // stores chunk offsets
	free         []bool         // which shards have free offsets
	freeMu       sync.RWMutex   // protects free field
	freeCache    *offsetCache   // optional cache of free offset values
	wg           sync.WaitGroup // blocks Close until all other method calls are done
	maxChunkSize int            // maximal chunk data size
	quit         chan struct{}  // quit disables all operations after Close is called
	quitOnce     sync.Once      // protects quit channel from multiple Close calls
}

// Option is an optional argument passed to New.
type Option func(*Store)

// WithCache is an optional argument to New constructor that enables
// in memory cache of free chunk data positions in files
func WithCache(yes bool) Option {
	return func(s *Store) {
		if yes {
			s.freeCache = newOffsetCache(shardCount)
		} else {
			s.freeCache = nil
		}
	}
}

// New constructs a new Store with files at path, with specified max chunk size.
func New(path string, maxChunkSize int, metaStore MetaStore, opts ...Option) (s *Store, err error) {
	s = &Store{
		shards:       make([]shard, shardCount),
		meta:         metaStore,
		free:         make([]bool, shardCount),
		maxChunkSize: maxChunkSize,
		quit:         make(chan struct{}),
	}
	for _, o := range opts {
		o(s)
	}
	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}
	for i := byte(0); i < shardCount; i++ {
		s.shards[i].f, err = os.OpenFile(filepath.Join(path, fmt.Sprintf("chunks-%v.db", i)), os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		s.shards[i].mu = new(sync.Mutex)
	}
	return s, nil
}

// Get returns a chunk with data.
func (s *Store) Get(addr chunk.Address) (ch chunk.Chunk, err error) {
	if err := s.protect(); err != nil {
		return nil, err
	}
	defer s.unprotect()

	sh := s.shards[getShard(addr)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	m, err := s.getMeta(addr)
	if err != nil {
		return nil, err
	}
	data := make([]byte, m.Size)
	n, err := sh.f.ReadAt(data, m.Offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n != int(m.Size) {
		return nil, fmt.Errorf("incomplete chunk data, read %v of %v", n, m.Size)
	}
	return chunk.NewChunk(addr, data), nil
}

// Has returns true if chunk is stored.
func (s *Store) Has(addr chunk.Address) (yes bool, err error) {
	if err := s.protect(); err != nil {
		return false, err
	}
	defer s.unprotect()

	mu := s.shards[getShard(addr)].mu
	mu.Lock()
	defer mu.Unlock()

	_, err = s.getMeta(addr)
	if err != nil {
		if err == chunk.ErrChunkNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Put stores chunk data.
func (s *Store) Put(ch chunk.Chunk) (err error) {
	if err := s.protect(); err != nil {
		return err
	}
	defer s.unprotect()

	addr := ch.Address()
	data := ch.Data()

	size := len(data)
	if size > s.maxChunkSize {
		return fmt.Errorf("chunk data size %v exceeds %v bytes", size, s.maxChunkSize)
	}

	section := make([]byte, s.maxChunkSize)
	copy(section, data)

	shard := getShard(addr)
	sh := s.shards[shard]

	sh.mu.Lock()
	defer sh.mu.Unlock()

	_, err = s.getMeta(addr)
	switch err {
	case chunk.ErrChunkNotFound:
	case nil:
		return nil
	default:
		return err
	}

	offset, reclaimed, err := s.getOffset(shard)
	if err != nil {
		return err
	}

	if offset < 0 {
		// no free offsets found,
		// append the chunk data by
		// seeking to the end of the file
		offset, err = sh.f.Seek(0, io.SeekEnd)
	} else {
		// seek to the offset position
		// to replace the chunk data at that position
		_, err = sh.f.Seek(offset, io.SeekStart)
	}
	if err != nil {
		return err
	}

	if _, err = sh.f.Write(section); err != nil {
		return err
	}
	if reclaimed && s.freeCache != nil {
		s.freeCache.remove(shard, offset)
	}
	return s.meta.Set(addr, shard, reclaimed, &Meta{
		Size:   uint16(size),
		Offset: offset,
	})
}

// getOffset returns an offset where chunk data can be written to
// and a flag if the offset is reclaimed from a previously removed chunk.
// If offset is less then 0, no free offsets are available.
func (s *Store) getOffset(shard uint8) (offset int64, reclaimed bool, err error) {
	if !s.shardHasFreeOffsets(shard) {
		return -1, false, nil
	}

	offset = -1
	if s.freeCache != nil {
		offset = s.freeCache.get(shard)
	}

	if offset < 0 {
		offset, err = s.meta.FreeOffset(shard)
		if err != nil {
			return 0, false, err
		}
	}
	if offset < 0 {
		s.markShardWithFreeOffsets(shard, false)
		return -1, false, nil
	}

	return offset, true, nil
}

// Delete makes the chunk unavailable.
func (s *Store) Delete(addr chunk.Address) (err error) {
	if err := s.protect(); err != nil {
		return err
	}
	defer s.unprotect()

	shard := getShard(addr)
	s.markShardWithFreeOffsets(shard, true)

	mu := s.shards[shard].mu
	mu.Lock()
	defer mu.Unlock()

	if s.freeCache != nil {
		m, err := s.getMeta(addr)
		if err != nil {
			return err
		}
		s.freeCache.set(shard, m.Offset)
	}
	return s.meta.Remove(addr, shard)
}

// Count returns a number of stored chunks.
func (s *Store) Count() (count int, err error) {
	return s.meta.Count()
}

// Iterate iterates over stored chunks in no particular order.
func (s *Store) Iterate(fn func(chunk.Chunk) (stop bool, err error)) (err error) {
	if err := s.protect(); err != nil {
		return err
	}
	defer s.unprotect()

	for _, sh := range s.shards {
		sh.mu.Lock()
	}
	defer func() {
		for _, sh := range s.shards {
			sh.mu.Unlock()
		}
	}()

	return s.meta.Iterate(func(addr chunk.Address, m *Meta) (stop bool, err error) {
		data := make([]byte, m.Size)
		_, err = s.shards[getShard(addr)].f.ReadAt(data, m.Offset)
		if err != nil {
			return true, err
		}
		return fn(chunk.NewChunk(addr, data))
	})
}

// Close disables of further operations on the Store.
// Every call to its methods will return ErrStoreClosed error.
// Close will wait for all running operations to finish before
// closing its MetaStore and returning.
func (s *Store) Close() (err error) {
	s.quitOnce.Do(func() {
		close(s.quit)
	})

	timeout := 15 * time.Second
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		log.Debug("timeout on waiting chunk store parallel operations to finish", "timeout", timeout)
	}

	for _, sh := range s.shards {
		if err := sh.f.Close(); err != nil {
			return err
		}
	}
	return s.meta.Close()
}

// protect protects Store from executing operations
// after the Close method is called and makes sure
// that Close method will wait for all ongoing operations
// to finish before returning. Method unprotect done
// must be closed to unblock the Close method call.
func (s *Store) protect() (err error) {
	select {
	case <-s.quit:
		return ErrStoreClosed
	default:
	}
	s.wg.Add(1)
	return nil
}

// unprotect removes a protection set by the protect method
// allowing the Close method to unblock.
func (s *Store) unprotect() {
	s.wg.Done()
}

// getMeta returns Meta information from MetaStore.
func (s *Store) getMeta(addr chunk.Address) (m *Meta, err error) {
	return s.meta.Get(addr)
}

func (s *Store) markShardWithFreeOffsets(shard uint8, has bool) {
	s.freeMu.Lock()
	s.free[shard] = has
	s.freeMu.Unlock()
}

func (s *Store) shardHasFreeOffsets(shard uint8) (has bool) {
	s.freeMu.RLock()
	has = s.free[shard]
	s.freeMu.RUnlock()
	return has
}

// getShard returns a shard number for the chunk address.
func getShard(addr chunk.Address) (shard uint8) {
	return addr[len(addr)-1] % shardCount
}

type shard struct {
	f  *os.File
	mu *sync.Mutex
}
