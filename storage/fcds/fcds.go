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
	NextShard() (freeShard []uint8, fallback uint8, err error)
	//NextShardLocked() (freeShard shard, s uint8, offset int64, reclaimed bool, err error)
	ShardSize() (slots []ShardInfo, err error)
	Count() (count int, err error)
	Iterate(func(ch chunk.Chunk) (stop bool, err error)) (err error)
	Close() (err error)
}

var _ Storer = new(Store)

// Number of files that store chunk data.
var ShardCount = uint8(32)

// ErrStoreClosed is returned if store is already closed.
var (
	ErrStoreClosed = errors.New("closed store")
	ErrNextShard   = errors.New("error getting next shard")
)

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
	mtx          sync.Mutex
}

// Option is an optional argument passed to New.
type Option func(*Store)

// WithCache is an optional argument to New constructor that enables
// in memory cache of free chunk data positions in files
func WithCache(yes bool) Option {
	return func(s *Store) {
		if yes {
			s.freeCache = newOffsetCache(ShardCount)
		} else {
			s.freeCache = nil
		}
	}
}

// New constructs a new Store with files at path, with specified max chunk size.
func New(path string, maxChunkSize int, metaStore MetaStore, opts ...Option) (s *Store, err error) {
	s = &Store{
		shards:       make([]shard, ShardCount),
		meta:         metaStore,
		free:         make([]bool, ShardCount),
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
		i := i
		fs, err := sh.f.Stat()
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
	defer sh.mu.Unlock()

	data := make([]byte, m.Size)
	n, err := sh.f.ReadAt(data, m.Offset)
	if err != nil && err != io.EOF {
		metrics.GetOrRegisterCounter("fcds.get.error", nil).Inc(1)

		return nil, err
	}
	if n != int(m.Size) {
		return nil, fmt.Errorf("incomplete chunk data, read %v of %v", n, m.Size)
	}
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

	//sh, shardId, offset, reclaimed, err := s.NextShardLocked()
	shardId, offset, cancel, err := s.so()
	if err != nil {
		return 0, err
	}
	reclaimed := offset >= 0

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
	if reclaimed && s.freeCache != nil {
		s.freeCache.remove(shardId, offset)
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
	m, err := s.getMeta(addr)
	if err != nil {
		return err
	}

	s.markShardWithFreeOffsets(m.Shard, true)
	s.mtx.Lock()
	defer s.mtx.Unlock()

	mu := s.shards[m.Shard].mu
	mu.Lock()
	defer mu.Unlock()

	if s.freeCache != nil {
		s.freeCache.set(m.Shard, m.Offset)
	}

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
	//s.mtx.Lock()
	//defer s.mtx.Unlock()
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

// NextShard gets the next shard to write to.
// Returns a slice of shards with free slots and a fallback shard
// which is the shortest shard in size, but does not guarantee it has
// any free slots.
func (s *Store) NextShard() (freeShards []uint8, fallback uint8, err error) {

	// multiple writers that call this at the same time will get the same shard again and again
	// because the free slot value has not been decremented yet
	slots := s.meta.ShardSlots()
	sort.Sort(byVal(slots))

	for _, v := range slots {
		if v.Val > 0 {
			freeShards = append(freeShards, v.Shard)
		}
	}

	// each element Val is the shard size in bytes
	shardSizes, err := s.ShardSize()
	if err != nil {
		return nil, 0, err
	}

	// sorting them will make the first element the largest shard and the last
	// element the smallest shard; pick the smallest
	sort.Sort(byVal(shardSizes))

	return freeShards, shardSizes[len(shardSizes)-1].Shard, nil
}

func (s *Store) so() (uint8, int64, func(), error) {

	shard, offset, cancel, err := s.meta.FastFreeOffset()
	if err != nil {
		return 0, 0, func() {}, err
	}

	if offset >= 0 {
		return shard, offset, cancel, nil
	}

	// each element Val is the shard size in bytes
	shardSizes, err := s.ShardSize()
	if err != nil {
		return 0, 0, func() {}, err
	}

	// sorting them will make the first element the largest shard and the last
	// element the smallest shard; pick the smallest
	sort.Sort(byVal(shardSizes))

	return shardSizes[len(shardSizes)-1].Shard, -1, func() {}, nil

}

func (s *Store) NextShardLocked() (sh shard, id uint8, offset int64, reclaimed bool, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	// multiple writers that call this at the same time will get the same shard again and again
	// because the free slot value has not been decremented yet
	slots := s.meta.ShardSlots()
	sort.Sort(byVal(slots))
	if slots[0].Val > 0 {
		fmt.Println("slots", slots[0].Val, slots)
		i := slots[0].Shard
		sh = s.shards[i]
		sh.mu.Lock()
		offset, reclaimed, err = s.getOffset(i)
		if err != nil {
			sh.mu.Unlock()
			return sh, 0, 0, false, err
		}

		if offset < 0 {
			panic("got no offset but should have")
			s.shards[i].mu.Unlock()
		} else {
			return sh, i, offset, reclaimed, nil
		}
	}

	// each element Val is the shard size in bytes
	shardSizes, err := s.ShardSize()
	if err != nil {
		return sh, 0, 0, false, err
	}

	// sorting them will make the first element the largest shard and the last
	// element the smallest shard; pick the smallest
	sort.Sort(byVal(shardSizes))

	fallback := shardSizes[len(shardSizes)-1].Shard
	sh = s.shards[fallback]
	sh.mu.Lock()
	offset, reclaimed, err = s.getOffset(fallback)
	if err != nil {
		sh.mu.Unlock()
		return sh, 0, 0, false, err
	}

	return sh, fallback, offset, reclaimed, nil

}

type shard struct {
	f  *os.File
	mu *sync.Mutex
}
