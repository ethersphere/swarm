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
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/syndtr/goleveldb/leveldb"
)

var _ chunk.TagStore = &DB{}

/*
type Tag struct {
	uid       uint64 //a unique identifier for this tag
	name      string
	total     uint32     // total chunks belonging to a tag
	split     uint32     // number of chunks already processed by splitter for hashing
	stored    uint32     // number of chunks already stored locally
	sent      uint32     // number of chunks sent for push syncing
	synced    uint32     // number of chunks synced with proof
	startedAt time.Time  // tag started to calculate ETA
	State     chan State // channel to signal completion
}
*/
func (db *DB) Write(tag *chunk.Tag) error {
	return nil
}

func (db *DB) NewTag(uploadTime int64, uploadName string) (tag uint32, err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()
	tag = db.rng.Uint32()
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

func (db *DB) DeleteTag(tag uint32) error {
	return db.tagIndex.Delete(tag)
}

func (db *DB) GetTags() (*chunk.Tags, error) {
	t := chunk.NewTags()
	err := db.tagIndex.Iterate(func(k, v interface{}) (bool, error) {
		tag := tagFromInterface(k, v)

		_, loaded := t.LoadOrStore(tag.GetUid(), tag)
		if loaded {
			return true, fmt.Errorf("tag uid %d already exists", tag.GetUid())
		}
		return false, nil
	}, nil)
	return t, err
}

func (db *DB) GetTag(tag uint32) (*chunk.Tag, error) {
	out, err := db.tagIndex.Get(tag)
	if err != nil {
		return nil, err
	}

	t := tagFromInterface(tag, out)

	return t, nil
}

func (db *DB) ChunkTags(addr chunk.Address) ([]uint32, error) {
	/*item := addressToItem(addr)

	out, err := db.retrievalDataIndex.Get(item)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return []uint32{}, nil
		}

		return nil, err
	}
	c, err := db.pushIndex.Get(out)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return []uint32{}, nil
		}
		return nil, err
	}

	return c.Tags, nil*/
	return []uint32{}, nil
}

func tagFromInterface(k, v interface{}) (*chunk.Tag, error) {
	uid := k.(uint32)
	valBytes := v.([]byte)
	_ = binary.BigEndian.Uint32(valBytes)

	tagName := string(valBytes[8:])

	t := &chunk.Tag{
		uid:       uid,
		name:      s,
		startedAt: time.Now(),
		total:     uint32(total),
		State:     make(chan State, 5),
	}

	return t

}
