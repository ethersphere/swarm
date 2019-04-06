// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package tagstore

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/syndtr/goleveldb/leveldb"
)

var _ chunk.TagStore = &DB{}

func (db *DB) NewTag(uploadTime int64, uploadName string) (tag uint64, err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()
	tag = db.rng.Uint64()
	batch := new(leveldb.Batch)
	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, uint64(uploadTime))
	val = append(val, []byte(uploadName)...)
	//check that it doesnt exist
	// put to indexes: tag
	err = db.tagIndex.PutInBatch(batch, tag, val)
	if err != nil {
		return tag, err
	}

	err = db.shed.WriteBatch(batch)
	if err != nil {
		return tag, err
	}

	return tag, nil
}

func (db *DB) DeleteTag(tag uint64) error {
	return nil
}

func (db *DB) GetTags() (*chunk.Tags, error) {
	t := chunk.NewTags()
	err := db.tagIndex.Iterate(func(k, v interface{}) (bool, error) {
		keyVal := k.(uint64)
		valBytes := v.([]byte)
		_ = binary.BigEndian.Uint64(valBytes)

		tagName := string(valBytes[8:])
		_, err := t.New(keyVal, tagName, 0)
		if err != nil {
			return true, err
		}
		return false, nil
	}, nil)
	return t, err
}

func (db *DB) GetTag(tag uint64) (chunk.Tag, error) {

	return chunk.Tag{}, nil
}

func (db *DB) ChunkTags(addr chunk.Address) ([]uint64, error) {
	/*item := addressToItem(addr)

	out, err := db.retrievalDataIndex.Get(item)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return []uint64{}, nil
		}

		return nil, err
	}
	c, err := db.pushIndex.Get(out)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return []uint64{}, nil
		}
		return nil, err
	}

	return c.Tags, nil*/
	return []uint64{}, nil
}
