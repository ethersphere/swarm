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
	"github.com/ethereum/go-ethereum/swarm/chunk"
)

var _ chunk.TagStore = &DB{}

func (db *DB) NewTag(uploadTime int64, uploadName string) (tag *chunk.Tag, err error) {
	// protect parallel updates
	db.batchMu.Lock()
	defer db.batchMu.Unlock()
	tagUid := db.rng.Uint32()

	tag = chunk.NewTag(tagUid, uploadName, 0)

	err = db.Store(tag)
	return tag, err
}

func (db *DB) Store(tag *chunk.Tag) error {
	// tag key is tag.uid
	key := string(tag.GetUid())

	return db.store.Put(key, tag)
}
func (db *DB) Load(uid uint32) (tag *chunk.Tag, err error) {
	// tag key is uid
	key := string(uid)
	tag = &chunk.Tag{}
	err = db.store.Get(key, tag)
	return tag, err
}

func (db *DB) Delete(tag uint32) error {
	return db.store.Delete(string(tag))
}

func (db *DB) GetTags() ([]*chunk.Tag, error) {
	return nil, nil
}

func (db *DB) GetTag(tag uint32) (*chunk.Tag, error) {
	return nil, nil
}
