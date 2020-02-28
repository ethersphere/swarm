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
	"bytes"
	"encoding/binary"
	"encoding/gob"
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
	free map[uint8]int64 // free slots cardinality
	mtx  sync.RWMutex    // synchronise free slots
}

// NewMetaStore returns new MetaStore at path.
func NewMetaStore(path string) (s *MetaStore, err error) {
	db, err := leveldb.OpenFile(path, &opt.Options{})
	if err != nil {
		return nil, err
	}

	ms := &MetaStore{
		db:   db,
		free: make(map[uint8]int64),
	}

	data, err := ms.db.Get(freeCountKey(), nil)
	if err != nil {
		// key doesn't exist since this is a new db
		// write an empty set into it
		b, err := encodeFreeSlots(ms.free)
		if err != nil {
			return nil, err
		}

		err = ms.db.Put(freeCountKey(), b, nil)
		if err != nil {
			return nil, err
		}
	}

	ms.free, err = decodeFreeSlots(data)
	if err != nil {
		return nil, err
	}

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

	s.mtx.Lock()
	defer s.mtx.Unlock()
	b, err := encodeFreeSlots(s.free)
	if err != nil {
		return err
	}
	batch.Put(freeCountKey(), b)
	batch.Delete(chunkKey(addr))

	err = s.db.Write(batch, nil)
	if err != nil {
		return err
	}

	s.free[shard]++

	return nil
}

// ShardSlots gives back a slice of ShardSlot items that represent the number
// of free slots inside each shard.
func (s *MetaStore) ShardSlots() (freeSlots []fcds.ShardSlot) {
	freeSlots = make([]fcds.ShardSlot, fcds.ShardCount)

	s.mtx.RLock()
	for i := uint8(0); i < fcds.ShardCount; i++ {
		slot := fcds.ShardSlot{Shard: i}
		if slots, ok := s.free[i]; ok {
			slot.Slots = slots
		}
		freeSlots[i] = slot
	}
	s.mtx.RUnlock()

	return freeSlots
}

// FreeOffset returns an offset that can be reclaimed by
// another chunk. If the returned value is less then 0
// there are no free offset at this shard.
func (s *MetaStore) FreeOffset(shard uint8) (offset int64, err error) {
	i := s.db.NewIterator(nil, nil)
	defer i.Release()

	i.Seek([]byte{freePrefix, shard})
	key := i.Key()
	if key == nil || key[0] != freePrefix || key[1] != shard {
		return -1, nil
	}
	offset = int64(binary.BigEndian.Uint64(key[2:10]))
	return offset, nil
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

// Close closes the underlaying LevelDB instance.
func (s *MetaStore) Close() (err error) {
	return s.db.Close()
}

const (
	chunkPrefix = 0
	freePrefix  = 1
	freeCount   = 2
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

func freeCountKey() (key []byte) {
	return []byte{freeCount}
}

func encodeFreeSlots(m map[uint8]int64) ([]byte, error) {
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	err := e.Encode(m)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func decodeFreeSlots(b []byte) (map[uint8]int64, error) {
	buf := bytes.NewBuffer(b)
	var decodedMap map[uint8]int64
	d := gob.NewDecoder(buf)

	err := d.Decode(&decodedMap)
	if err != nil {
		return nil, err
	}

	return decodedMap, nil
}
