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
	"github.com/syndtr/goleveldb/leveldb"
)

var _ TagStore = &DB{}

type TagStore interface {
	PutUploadID(uploadId uint64, timestamp int64, uploadName string) error

	GetChunkTags(addr chunk.Address) ([]uint64, error)
	PutTag(uploadId, tag uint64, path string) error
}

func (db *DB) PutUploadID(id uint64, uploadTime int64, uploadName string) (err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()

	batch := new(leveldb.Batch)
	var k, v interface{}

	// put to indexes: tag
	db.uploadIndex.PutInBatch(batch, k, v)
	err = db.shed.WriteBatch(batch)
	if err != nil {
		return err
	}
	return nil

}

func (db *DB) GetChunkTags(addr chunk.Address) ([]uint64, error) {
	item := addressToItem(addr)

	out, err := db.retrievalDataIndex.Get(item)
	if err != nil {
		return nil, err
	}
	c, err := db.pushIndex.Get(out)
	if err != nil {
		return nil, err
	}

	return c.Tags, nil
}

func (db *DB) PutTag(uploadId, tag uint64, path string) (err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()

	batch := new(leveldb.Batch)
	var k, v interface{}
	// put to indexes: tag
	db.tagIndex.PutInBatch(batch, k, v)
	err = db.shed.WriteBatch(batch)
	if err != nil {
		return err
	}
	return nil
}
