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

package localstore

import (
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/shed"
	"github.com/syndtr/goleveldb/leveldb"
)

var _ TagStore = &DB{}

type TagStore interface {
	PutUploadID(uploadId uint64, timestamp int64, uploadName string) error
	PutTags(key, value interface{}) error

	GetTags(addr chunk.Address) ([]uint64, error)
	PutTags(item shed.Item, tags []uint64) ([]uint64, error)
}

func (db *DB) PutUploadID(id uint64, uploadTime int64, uploadName string) (err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()

	batch := new(leveldb.Batch)

	// put to indexes: tag
	db.Index.PutInBatch(batch, itemK, itemV)
	err = db.shed.WriteBatch(batch)
	if err != nil {
		return err
	}
	return nil

}

func (db *DB) GetTags(addr chunk.Address) ([]uint64, error) {
	item := addressToItem(addr)

	out, err = db.retrievalDataIndex.Get(item)
	if err != nil {
		return out, err
	}
	c, err := db.pushIndex.Get(out)
	if err != nil {
		return out, err
	}

	return c.Tags(), nil
}

func (db *DB) PutTags(uploadId, tag uint64, path string) (err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()

	batch := new(leveldb.Batch)

	// put to indexes: tag
	db.tagIndex.PutInBatch(batch, itemK, itemV)
	err = db.shed.WriteBatch(batch)
	if err != nil {
		return err
	}
	return nil
}
