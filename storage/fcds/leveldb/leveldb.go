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

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage/fcds"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var _ fcds.MetaStore = new(MetaStore)

type MetaStore struct {
	db *leveldb.DB
}

func NewMetaStore(filename string) (s *MetaStore, err error) {
	db, err := leveldb.OpenFile(filename, &opt.Options{})
	if err != nil {
		return nil, err
	}
	return &MetaStore{
		db: db,
	}, err
}

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

func (s *MetaStore) Remove(addr chunk.Address, shard uint8) (err error) {
	m, err := s.Get(addr)
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)
	batch.Put(freeKey(shard, m.Offset), nil)
	batch.Delete(chunkKey(addr))
	return s.db.Write(batch, nil)
}

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
