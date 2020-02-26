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
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"math/rand"
	"sort"
	"sync"

	"github.com/ethersphere/swarm/chunk"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

var _ MetaStore = new(metaStore)

const probabilisticNextShard = true

// MetaStore implements FCDS MetaStore with LevelDB
// for persistence.
type metaStore struct {
	db   *leveldb.DB
	free map[uint8]int64 // free slots cardinality
	mtx  sync.RWMutex    // synchronise free slots
}

// NewMetaStore returns new MetaStore at path.
func NewMetaStore(path string, inmem bool) (s *metaStore, err error) {
	var (
		db *leveldb.DB
	)

	if inmem {
		db, err = leveldb.Open(storage.NewMemStorage(), &opt.Options{})
	} else {
		db, err = leveldb.OpenFile(path, &opt.Options{})
	}

	if err != nil {
		return nil, err
	}

	// todo: try to get and deserialize the free map from the persisted value on disk

	return &metaStore{
		db:   db,
		free: make(map[uint8]int64),
	}, err
}

// Get returns chunk meta information.
func (s *metaStore) Get(addr chunk.Address) (m *Meta, err error) {
	data, err := s.db.Get(chunkKey(addr), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, chunk.ErrChunkNotFound
		}
		return nil, err
	}
	m = new(Meta)
	if err := m.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return m, nil
}

// Set adds a new chunk meta information for a shard.
// Reclaimed flag denotes that the chunk is at the place of
// already deleted chunk, not appended to the end of the file.
func (s *metaStore) Set(addr chunk.Address, shard uint8, reclaimed bool, m *Meta) (err error) {
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
func (s *metaStore) Remove(addr chunk.Address, shard uint8) (err error) {
	m, err := s.Get(addr)
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	batch.Put(freeKey(shard, m.Offset), nil)
	s.free[shard]++
	batch.Put(freeCountKey(), encodeFreeSlots(s.free))
	batch.Delete(chunkKey(addr))

	err = s.db.Write(batch, nil)
	if err != nil {
		s.free[shard]-- // rollback the value change since the commit did not succeed
		return err
	}

	return nil
}

// NextShard returns an id of the next shard to write to and a boolean
// indicating whether that shard has free offsets
func (s *metaStore) NextShard() (shard uint8, hasFree bool) {

	if probabilisticNextShard {
		// do some magic
		slots := s.shardSlots(false)
		x := make([]int64, len(slots))
		sum := 0
		for _, v := range slots {

			// we need to consider the edge case where no free slots are available
			// we still need to potentially insert 1 chunk and so if all shards have
			// no empty offsets - they all must be considered equally as having at least
			// one empty slot
			if v == 0 {
				v = 1
			}
			sum += v
		}

		magic := rand.Intn(sum)
		movingSum := 0
		for _, v := range slots {
			add := 0
			if v.slots == 0 {
				add = 1
			}
			movingSum += v.slots + add
			if magic <= movingSum {
				// we've reached the shard with the correct id
				return v.shard, v.slots > 0
			}
		}
	} else {
		// get a definitive next shard to write to with a free slot
	}

	// now iterate over all free slots then give weight to each one of them
	// then generate a float and see which one gets picked
	return 0, false

	//return freeSlots[0].shard, freeSlots[0].slots > 0
}

// shardSlots gives back a slice of shardSlot items that represent the number
// of free slots inside each shard. Output can be toggled to be sorted by the
// function argument
func (s *metaStore) shardSlots(toSort bool) (freeSlots []shardSlot) {
	freeSlots = make([]shardSlot, ShardCount)
	s.mtx.Lock()

	for i := 0; uint8(i) < ShardCount; i++ {
		slot := shardSlot{shard: uint8(i)}
		if slots, ok := s.free[uint8(i)]; ok {
			slot.slots = slots

		}
		freeSlots[i] = slot
	}
	s.mtx.Unlock()

	if toSort {
		sort.Sort(BySlots(freeSlots))
	}

	return freeSlots

}

// FreeOffset returns an offset that can be reclaimed by
// another chunk. If the returned value is less then 0
// there are no free offset at this shard.
func (s *metaStore) FreeOffset() (shard uint8, offset int64, err error) {
	i := s.db.NewIterator(nil, nil)
	defer i.Release()

	i.Seek([]byte{freePrefix})
	key := i.Key()
	if key == nil || key[0] != freePrefix {
		return 0, -1, nil
	}
	shard = key[1]
	offset = int64(binary.BigEndian.Uint64(key[2:10]))
	return shard, offset, nil
}

// Count returns a number of chunks in MetaStore.
// This operation is slow for larger numbers of chunks.
func (s *metaStore) Count() (count int, err error) {
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
func (s *metaStore) Iterate(fn func(chunk.Address, *Meta) (stop bool, err error)) (err error) {
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
		m := new(Meta)
		if err := m.UnmarshalBinary(value); err != nil {
			return err
		}
		stop, err := fn(chunk.Address(key[1:]), m)
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
func (s *metaStore) Close() (err error) {
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

func encodeFreeSlots(m map[uint8]int64) []byte {
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	err := e.Encode(m)
	if err != nil {
		panic(err)
	}

	return b.Bytes()
}

func decodeFreeSlots(b []byte) map[uint8]int64 {
	buf := bytes.NewBuffer(b)
	var decodedMap map[uint8]int64
	d := gob.NewDecoder(buf)

	err := d.Decode(&decodedMap)
	if err != nil {
		panic(err)
	}

	return decodedMap
}

type BySlots []shardSlot

func (a BySlots) Len() int           { return len(a) }
func (a BySlots) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySlots) Less(i, j int) bool { return a[j].slots < a[i].slots }

type shardSlot struct {
	shard uint8
	slots int64
}
