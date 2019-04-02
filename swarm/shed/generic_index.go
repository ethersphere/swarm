// Copyright 2018 The go-ethereum Authors
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

package shed

import (
	"github.com/syndtr/goleveldb/leveldb"
)

// Index represents a set of LevelDB key value pairs that have common
// prefix. It holds functions for encoding and decoding keys and values
// to provide transparent actions on saved data which inclide:
// - getting a particular Item
// - saving a particular Item
// - iterating over a sorted LevelDB keys
type GenericIndex struct {
	db              *DB
	prefix          []byte
	encodeKeyFunc   func(item interface{}) (key []byte, err error)
	decodeKeyFunc   func(key []byte) (item interface{}, err error)
	encodeValueFunc func(fields interface{}) (value []byte, err error)
	decodeValueFunc func(keyFields interface{}, value []byte) (e interface{}, err error)
}

// GenericIndexFuncs structure defines functions for encoding and decoding
// LevelDB keys and values for a specific index.
type GenericIndexFuncs struct {
	EncodeKey   func(fields interface{}) (key []byte, err error)
	DecodeKey   func(key []byte) (e interface{}, err error)
	EncodeValue func(fields interface{}) (value []byte, err error)
	DecodeValue func(keyFields interface{}, value []byte) (e interface{}, err error)
}

// NewIndex returns a new Index instance with defined name and
// encoding functions. The name must be unique and will be validated
// on database schema for a key prefix byte.
func (db *DB) NewGenericIndex(name string, funcs GenericIndexFuncs) (f GenericIndex, err error) {
	id, err := db.schemaIndexPrefix(name)
	if err != nil {
		return f, err
	}
	prefix := []byte{id}
	return GenericIndex{
		db:     db,
		prefix: prefix,
		// This function adjusts Index LevelDB key
		// by appending the provided index id byte.
		// This is needed to avoid collisions between keys of different
		// indexes as all index ids are unique.
		encodeKeyFunc: func(e interface{}) (key []byte, err error) {
			key, err = funcs.EncodeKey(e)
			if err != nil {
				return nil, err
			}
			return append(append(make([]byte, 0, len(key)+1), prefix...), key...), nil
		},
		// This function reverses the encodeKeyFunc constructed key
		// to transparently work with index keys without their index ids.
		// It assumes that index keys are prefixed with only one byte.
		decodeKeyFunc: func(key []byte) (e interface{}, err error) {
			return funcs.DecodeKey(key[1:])
		},
		encodeValueFunc: funcs.EncodeValue,
		decodeValueFunc: funcs.DecodeValue,
	}, nil
}

// Get accepts key fields represented as Item to retrieve a
// value from the index and return maximum available information
// from the index represented as another Item.
func (f *GenericIndex) Get(keyFields interface{}) (out interface{}, err error) {
	key, err := f.encodeKeyFunc(keyFields)
	if err != nil {
		return out, err
	}
	value, err := f.db.Get(key)
	if err != nil {
		return out, err
	}
	out, err = f.decodeValueFunc(keyFields, value)
	if err != nil {
		return out, err
	}
	return out, nil
}

// Has accepts key fields represented as Item to check
// if there this Item's encoded key is stored in
// the index.
func (f GenericIndex) Has(keyFields interface{}) (bool, error) {
	key, err := f.encodeKeyFunc(keyFields)
	if err != nil {
		return false, err
	}
	return f.db.Has(key)
}

// Put accepts Item to encode information from it
// and save it to the database.
func (f GenericIndex) Put(k, v interface{}) (err error) {
	key, err := f.encodeKeyFunc(k)
	if err != nil {
		return err
	}
	value, err := f.encodeValueFunc(v)
	if err != nil {
		return err
	}
	return f.db.Put(key, value)
}

// PutInBatch is the same as Put method, but it just
// saves the key/value pair to the batch instead
// directly to the database.
func (f GenericIndex) PutInBatch(batch *leveldb.Batch, k, v interface{}) (err error) {
	key, err := f.encodeKeyFunc(k)
	if err != nil {
		return err
	}
	value, err := f.encodeValueFunc(v)
	if err != nil {
		return err
	}
	batch.Put(key, value)
	return nil
}

// Delete accepts Item to remove a key/value pair
// from the database based on its fields.
func (f GenericIndex) Delete(keyFields interface{}) (err error) {
	key, err := f.encodeKeyFunc(keyFields)
	if err != nil {
		return err
	}
	return f.db.Delete(key)
}

// DeleteInBatch is the same as Delete just the operation
// is performed on the batch instead on the database.
func (f GenericIndex) DeleteInBatch(batch *leveldb.Batch, keyFields interface{}) (err error) {
	key, err := f.encodeKeyFunc(keyFields)
	if err != nil {
		return err
	}
	batch.Delete(key)
	return nil
}
