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
	"github.com/ethersphere/swarm/chunk"
	"github.com/janos/forky"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var _ forky.Interface = new(LevelDBStore)

type LevelDBStore struct {
	db *leveldb.DB
}

func NewLevelDBStore(path string) (s *LevelDBStore, err error) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		OpenFilesCacheCapacity: 128,
	})
	if err != nil {
		return nil, err
	}
	return &LevelDBStore{
		db: db,
	}, nil
}

func (s *LevelDBStore) Get(addr chunk.Address) (c chunk.Chunk, err error) {
	data, err := s.db.Get(addr, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, chunk.ErrChunkNotFound
		}
		return nil, err
	}
	return chunk.NewChunk(addr, data), nil
}

func (s *LevelDBStore) Has(addr chunk.Address) (yes bool, err error) {
	return s.db.Has(addr, nil)
}

func (s *LevelDBStore) Put(ch chunk.Chunk) (err error) {
	return s.db.Put(ch.Address(), ch.Data(), nil)
}

func (s *LevelDBStore) Delete(addr chunk.Address) (err error) {
	return s.db.Delete(addr, nil)
}

func (s *LevelDBStore) Close() error {
	return s.db.Close()
}
