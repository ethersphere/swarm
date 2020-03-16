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
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/log"

	"github.com/ethersphere/swarm/chunk"
)

// Storer specifies methods required for FCDS implementation.
// It can be used where alternative implementations are needed to
// switch at runtime.
type Storer interface {
	Get(addr chunk.Address) (ch chunk.Chunk, err error)
	Has(addr chunk.Address) (yes bool, err error)
	Put(ch chunk.Chunk) (shard uint8, err error)
	Delete(addr chunk.Address) (err error)
	ShardSize() (slots []ShardInfo, err error)
	Count() (count int, err error)
	Iterate(func(ch chunk.Chunk) (stop bool, err error)) (err error)
	Close() (err error)
}

var _ Storer = new(Store)

// Number of files that store chunk data.
var ShardCount = uint8(32)

// ErrStoreClosed is returned if store is already closed.
var ErrStoreClosed = errors.New("closed store")

// Store is the main FCDS implementation. It stores chunk data into
// a number of files partitioned by the last byte of the chunk address.
type Store struct {
	shards       []shard        // relations with shard id and a shard file and their mutexes
	meta         MetaStore      // stores chunk offsets
	wg           sync.WaitGroup // blocks Close until all other method calls are done
	maxChunkSize int            // maximal chunk data size
	quit         chan struct{}  // quit disables all operations after Close is called
	quitOnce     sync.Once      // protects quit channel from multiple Close calls
}

// Option is an optional argument passed to New.
type Option func(*Store)

// New constructs a new Store with files at path, with specified max chunk size.
func New(path string, maxChunkSize int, metaStore MetaStore, opts ...Option) (s *Store, err error) {
	s = &Store{
		shards:       make([]shard, ShardCount),
		meta:         metaStore,
		maxChunkSize: maxChunkSize,
		quit:         make(chan struct{}),
	}
	for _, o := range opts {
		o(s)
	}
	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}
	for i := byte(0); i < ShardCount; i++ {
		s.shards[i].f, err = os.OpenFile(filepath.Join(path, fmt.Sprintf("chunks-%v.db", i)), os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		s.shards[i].mu = new(sync.Mutex)
	}
	return s, nil
}

func (s *Store) ShardSize() (slots []ShardInfo, err error) {
	slots = make([]ShardInfo, len(s.shards))
	for i, sh := range s.shards {
		sh.mu.Lock()
		fs, err := sh.f.Stat()
		sh.mu.Unlock()
		if err != nil {
			return nil, err
		}
		slots[i] = ShardInfo{Shard: uint8(i), Val: fs.Size()}
	}

	return slots, nil
}

// Get returns a chunk with data.
func (s *Store) Get(addr chunk.Address) (ch chunk.Chunk, err error) {
	if err := s.protect(); err != nil {
		return nil, err
	}
	defer s.unprotect()

	m, err := s.getMeta(addr)
	if err != nil {
		return nil, err
	}

	sh := s.shards[m.Shard]
	sh.mu.Lock()

	data := make([]byte, m.Size)
	n, err := sh.f.ReadAt(data, m.Offset)
	if err != nil && err != io.EOF {
		metrics.GetOrRegisterCounter("fcds.get.error", nil).Inc(1)

		sh.mu.Unlock()
		return nil, err
	}
	if n != int(m.Size) {
		return nil, fmt.Errorf("incomplete chunk data, read %v of %v", n, m.Size)
	}
	sh.mu.Unlock()

	metrics.GetOrRegisterCounter("fcds.get.ok", nil).Inc(1)

	return chunk.NewChunk(addr, data), nil
}

// Has returns true if chunk is stored.
func (s *Store) Has(addr chunk.Address) (yes bool, err error) {
	if err := s.protect(); err != nil {
		return false, err
	}
	defer s.unprotect()

	_, err = s.getMeta(addr)
	if err != nil {
		if err == chunk.ErrChunkNotFound {
			metrics.GetOrRegisterCounter("fcds.has.no", nil).Inc(1)
			return false, nil
		}
		metrics.GetOrRegisterCounter("fcds.has.err", nil).Inc(1)
		return false, err
	}
	metrics.GetOrRegisterCounter("fcds.has.ok", nil).Inc(1)

	return true, nil
}

// Put stores chunk data.
// Returns the shard number into which the chunk was added.
func (s *Store) Put(ch chunk.Chunk) (uint8, error) {
	if err := s.protect(); err != nil {
		return 0, err
	}
	defer s.unprotect()
	m, err := s.getMeta(ch.Address())
	if err == nil {
		return m.Shard, nil
	}
	addr := ch.Address()
	data := ch.Data()

	size := len(data)
	if size > s.maxChunkSize {
		return 0, fmt.Errorf("chunk data size %v exceeds %v bytes", size, s.maxChunkSize)
	}

	section := make([]byte, s.maxChunkSize)
	copy(section, data)

	shardId, offset, reclaimed, cancel, err := s.getOffset()
	if err != nil {
		return 0, err
	}

	sh := s.shards[shardId]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	if reclaimed {
		metrics.GetOrRegisterCounter("fcds.put.reclaimed", nil).Inc(1)
	}

	if offset < 0 {
		metrics.GetOrRegisterCounter("fcds.put.append", nil).Inc(1)
		// no free offsets found,
		// append the chunk data by
		// seeking to the end of the file
		offset, err = sh.f.Seek(0, io.SeekEnd)
	} else {
		metrics.GetOrRegisterCounter("fcds.put.offset", nil).Inc(1)
		// seek to the offset position
		// to replace the chunk data at that position
		_, err = sh.f.Seek(offset, io.SeekStart)
	}
	if err != nil {
		cancel()
		return 0, err
	}

	if _, err = sh.f.Write(section); err != nil {
		cancel()
		return 0, err
	}

	err = s.meta.Set(addr, shardId, reclaimed, &Meta{
		Size:   uint16(size),
		Offset: offset,
		Shard:  shardId,
	})
	if err != nil {
		cancel()
	}

	return shardId, err
}

// getOffset returns an offset on a shard where chunk data can be written to
// and a flag if the offset is reclaimed from a previously removed chunk.
// If offset is less then 0, no free offsets are available.
func (s *Store) getOffset() (shard uint8, offset int64, reclaimed bool, cancel func(), err error) {
	cancel = func() {}
	shard, offset, cancel = s.meta.FreeOffset()
	if offset >= 0 {
		return shard, offset, true, cancel, nil
	}

	// each element Val is the shard size in bytes
	shardSizes, err := s.ShardSize()
	if err != nil {
		return 0, 0, false, cancel, err
	}

	// sorting them will make the first element the largest shard and the last
	// element the smallest shard; pick the smallest
	sort.Sort(byVal(shardSizes))

	return shardSizes[len(shardSizes)-1].Shard, -1, false, cancel, nil

}

// Delete makes the chunk unavailable.
func (s *Store) Delete(addr chunk.Address) (err error) {
	if err := s.protect(); err != nil {
		return err
	}
	defer s.unprotect()

	m, err := s.getMeta(addr)
	if err != nil {
		return err
	}

	mu := s.shards[m.Shard].mu
	mu.Lock()
	defer mu.Unlock()

	err = s.meta.Remove(addr, m.Shard)
	if err != nil {
		metrics.GetOrRegisterCounter("fcds.delete.fail", nil).Inc(1)
		return err
	}

	metrics.GetOrRegisterCounter("fcds.delete.ok", nil).Inc(1)
	return nil
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
		_, err = s.shards[m.Shard].f.ReadAt(data, m.Offset)
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

type shard struct {
	f  *os.File
	mu *sync.Mutex
}
