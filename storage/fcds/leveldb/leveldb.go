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

package leveldb

import (
	"encoding/binary"
	"sync"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage/fcds"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var _ fcds.MetaStore = new(MetaStore)

// MetaStore implements FCDS MetaStore with LevelDB
// for persistence.
type MetaStore struct {
	db   *leveldb.DB
	free map[uint8]map[int64]struct{} // free slots map. root map key is shard id
	mtx  sync.Mutex                   // synchronise free slots
}

// NewMetaStore returns new MetaStore at path.
func NewMetaStore(path string) (s *MetaStore, err error) {
	db, err := leveldb.OpenFile(path, &opt.Options{})
	if err != nil {
		return nil, err
	}

	ms := &MetaStore{
		db:   db,
		free: make(map[uint8]map[int64]struct{}),
	}

	for i := uint8(0); i < fcds.ShardCount; i++ {
		ms.free[i] = make(map[int64]struct{})
	}

	// caution - this _will_ break if we one day decide to
	// decrease the shard count
	ms.iterateFree(func(shard uint8, offset int64) {
		ms.free[shard][offset] = struct{}{}
	})

	return ms, nil
}

// Get returns chunk meta information.
func (s *MetaStore) Get(addr chunk.Address) (m *fcds.Meta, err error) {
	data, err := s.db.Get(chunkKey(addr), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, chunk.ErrChunkNotFound
		}
		return nil, err
	}
	m = new(fcds.Meta)
	if err := m.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return m, nil
}

// Set adds a new chunk meta information for a shard.
// Reclaimed flag denotes that the chunk is at the place of
// already deleted chunk, not appended to the end of the file.
// Caller expected to hold the shard lock.
func (s *MetaStore) Set(addr chunk.Address, shard uint8, reclaimed bool, m *fcds.Meta) (err error) {
	batch := new(leveldb.Batch)
	if reclaimed {
		batch.Delete(freeKey(shard, m.Offset))
	}
	meta, err := m.MarshalBinary()
	if err != nil {
		return err
	}
	batch.Put(chunkKey(addr), meta)
	return s.db.Write(batch, nil)
}

// Remove removes chunk meta information from the shard.
func (s *MetaStore) Remove(addr chunk.Address, shard uint8) (err error) {
	m, err := s.Get(addr)
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)
	batch.Put(freeKey(shard, m.Offset), nil)
	batch.Delete(chunkKey(addr))

	err = s.db.Write(batch, nil)
	if err != nil {
		return err
	}

	s.mtx.Lock()
	s.free[shard][m.Offset] = struct{}{}
	s.mtx.Unlock()

	return nil
}

// FreeOffset returns an offset that can be reclaimed by
// another chunk. If the returned value is less then 0
// there are no free offsets on any shards and the chunk must be
// appended to the shortest shard
func (s *MetaStore) FreeOffset() (shard uint8, offset int64, cancel func()) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for shard, offsets := range s.free {
		for offset, _ = range offsets {
			delete(offsets, offset)
			return shard, offset, func() {
				s.mtx.Lock()
				defer s.mtx.Unlock()
				s.free[shard][offset] = struct{}{}
			}
		}
	}

	return 0, -1, func() {}
}

// Count returns a number of chunks in MetaStore.
// This operation is slow for larger numbers of chunks.
func (s *MetaStore) Count() (count int, err error) {
	it := s.db.NewIterator(nil, nil)
	defer it.Release()

	for ok := it.First(); ok; ok = it.Next() {
		value := it.Value()
		if len(value) == 0 {
			continue
		}
		key := it.Key()
		if len(key) < 1 {
			continue
		}
		count++
	}
	return count, it.Error()
}

// Iterate iterates over all chunk meta information.
func (s *MetaStore) Iterate(fn func(chunk.Address, *fcds.Meta) (stop bool, err error)) (err error) {
	it := s.db.NewIterator(nil, nil)
	defer it.Release()

	for ok := it.First(); ok; ok = it.Next() {
		value := it.Value()
		if len(value) == 0 {
			continue
		}
		key := it.Key()
		if len(key) < 1 {
			continue
		}
		m := new(fcds.Meta)
		if err := m.UnmarshalBinary(value); err != nil {
			return err
		}
		b := make([]byte, len(key)-1)
		copy(b, key[1:])
		stop, err := fn(chunk.Address(b), m)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}
	return it.Error()
}

// iterateFree iterates over all free slot entries in leveldb
// and calls the defined callback function on each entry found.
func (s *MetaStore) iterateFree(fn func(shard uint8, offset int64)) {
	i := s.db.NewIterator(nil, nil)
	defer i.Release()

	for ok := i.Seek([]byte{freePrefix}); ok; ok = i.Next() {
		key := i.Key()
		if key == nil || key[0] != freePrefix {
			return
		}
		shard := uint8(key[1])
		offset := int64(binary.BigEndian.Uint64(key[2:10]))
		fn(shard, offset)
	}
}

// Close closes the underlaying LevelDB instance.
func (s *MetaStore) Close() (err error) {
	return s.db.Close()
}

const (
	chunkPrefix = 0
	freePrefix  = 1
)

func chunkKey(addr chunk.Address) (key []byte) {
	return append([]byte{chunkPrefix}, addr...)
}

func freeKey(shard uint8, offset int64) (key []byte) {
	key = make([]byte, 10)
	key[0] = freePrefix
	key[1] = shard
	binary.BigEndian.PutUint64(key[2:10], uint64(offset))
	return key
}
