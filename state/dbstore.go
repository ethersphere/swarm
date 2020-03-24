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

package state

import (
	"encoding"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/badger"
	"io/ioutil"
	"os"


)

// ErrNotFound is returned when no results are returned from the database
var ErrNotFound = errors.New("ErrorNotFound")

// Store defines methods required to get, set, delete values for different keys
// and close the underlying resources.
type Store interface {
	Get(key string, i interface{}) (err error)
	Put(key string, i interface{}) (err error)
	Delete(key string) (err error)
	Iterate(prefix string, iterFunc iterFunction) (err error)
	GetBatch() (txn *badger.Txn)
	PutInBatch(key string, i interface{}, batch *badger.Txn) (err error)
	DeleteInBatch(key string, batch *badger.Txn) (err error)
	WriteBatch(batch *badger.Txn) (err error)
	Close() error
}

const (
	syncWrites          = true   // do not fsync entries as they are written
	valueThresholdLimit = 1024   // valuess less than 1K are co-located with the key
	valueLogEntries     = 100000  // maximum no of entries in a value log file
)

// DBStore uses badger to store values.
type DBStore struct {
	db *badger.DB
}

// NewDBStore creates a new instance of DBStore.
func NewDBStore(path string) (s *DBStore, err error) {
	o := badger.DefaultOptions(path)
	o.SyncWrites = syncWrites
	o.ValueThreshold = valueThresholdLimit
	o.ValueLogMaxEntries = valueLogEntries
	o.Logger = nil // don't use the badgers internal logging mechanism
	db, err := badger.Open(o)
	if err != nil {
		return nil, err
	}
	return &DBStore{
		db: db,
	}, nil
}

// NewInmemoryStore returns a new instance of DBStore. To be used only in tests and simulations.
func NewInmemoryStore() *DBStore {

	dir, err := ioutil.TempDir("", "state-test")
	defer os.RemoveAll(dir)
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		panic(err)
	}
	return &DBStore{
		db: db,
	}
}

// Get retrieves a persisted value for a specific key. If there is no results
// ErrNotFound is returned. The provided parameter should be either a byte slice or
// a struct that implements the encoding.BinaryUnmarshaler interface
func (s *DBStore) Get(key string, i interface{}) (err error) {
	var data []byte
	err = s.db.View(func(txn *badger.Txn) (err error) {
		item, err := txn.Get([]byte(key))
		if err != nil {
				return err
		}
		return item.Value(func(val []byte) error {
			data = make([]byte, len(val))
			copy(data, val)
			return nil
		})
	})
	if data == nil {
		return ErrNotFound
	}

	unmarshaler, ok := i.(encoding.BinaryUnmarshaler)
	if !ok {
		return json.Unmarshal(data, i)
	}
	return unmarshaler.UnmarshalBinary(data)
}

// Put stores an object that implements Binary for a specific key.
func (s *DBStore) Put(key string, i interface{}) (err error) {
	var bytes []byte
	if marshaler, ok := i.(encoding.BinaryMarshaler); ok {
		if bytes, err = marshaler.MarshalBinary(); err != nil {
			return err
		}
	} else {
		if bytes, err = json.Marshal(i); err != nil {
			return err
		}
	}
	return s.db.Update(func(txn *badger.Txn) (err error) {
		err = txn.Set([]byte(key), bytes)
		return err
	})
}

// Delete removes entries stored under a specific key.
func (s *DBStore) Delete(key string) (err error) {
	return s.db.Update(func(txn *badger.Txn) (err error) {
		return txn.Delete([]byte(key))
	})
}

// iterFunction is a function called on every key/value pair obtained
// through iterating the store.
// If true is returned in the stop variable, iteration will
// stop, and by returning the error, that error will be
// propagated to the called iterator method on Iterate.
type iterFunction func(key, value []byte) (stop bool, err error)

// Iterate entries (key/value pair) which have keys matching the given prefix
func (s *DBStore) Iterate(prefix string, iterFunc iterFunction) (err error) {
	return s.db.View(func(txn *badger.Txn) (err error) {
		o := badger.DefaultIteratorOptions
		o.PrefetchValues = true
		o.PrefetchSize = 1024
		i := txn.NewIterator(o)
		defer i.Close()

		for i.Seek([]byte(prefix)); i.ValidForPrefix([]byte(prefix)); i.Next() {
			value, err := i.Item().ValueCopy(nil)
			if err != nil {
				return err
			}
			stop, err := iterFunc(i.Item().Key(), value)
			if err != nil {
				return err
			}
			if stop {
				break
			}
		}
		return nil
	})
}

// Close releases the resources used by the underlying LevelDB.
func (s *DBStore) Close() error {
	return s.db.Close()
}


func (s *DBStore) GetBatch() *badger.Txn{
	return s.db.NewTransaction(true)
}



// Put encodes the value and puts a corresponding Put operation into the underlying batch.
// This only returns an error if the encoding failed.
func (s *DBStore) PutInBatch(key string, i interface{}, batch *badger.Txn) (err error) {
	var bytes []byte
	if marshaler, ok := i.(encoding.BinaryMarshaler); ok {
		if bytes, err = marshaler.MarshalBinary(); err != nil {
			return err
		}
	} else {
		if bytes, err = json.Marshal(i); err != nil {
			return err
		}
	}
	return batch.Set([]byte(key), bytes)
}

// Delete adds a delete operation to the underlying batch.
func (b *DBStore) DeleteInBatch(key string, batch *badger.Txn) error {
	return batch.Delete([]byte(key))
}

// WriteBatch executes the batch on the underlying database.
func (s *DBStore) WriteBatch(batch *badger.Txn) error {
	defer batch.Discard()
	err := batch.Commit()
	if err != nil {
		return nil
	}
	return nil
}
