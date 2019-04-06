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

package tagstore

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/shed"
	"github.com/ethereum/go-ethereum/swarm/storage/mock"
)

// DB implements chunk.Store.
var _ chunk.TagStore = &DB{}

var (
	// Default value for Capacity DB option.
	defaultCapacity uint64 = 5000000
)

// DB is the tag store implementation and holds
// database related objects.
type DB struct {
	shed *shed.DB

	// schema name of loaded data
	schemaName shed.StringField

	// capacity
	capacity uint64

	// retrieval indexes
	// tags index maintains a mapping between tags and upload time, chunk status and tag name
	tagIndex shed.GenericIndex

	baseKey []byte

	batchMu sync.Mutex

	// this channel is closed when close function is called
	// to terminate other goroutines
	close chan struct{}
}

// Options struct holds optional parameters for configuring DB.
type Options struct {
	// MockStore is a mock node store that is used to store
	// chunk data in a central store. It can be used to reduce
	// total storage space requirements in testing large number
	// of swarm nodes with chunk data deduplication provided by
	// the mock global store.
	MockStore *mock.NodeStore
	// Capacity is a limit that triggers garbage collection when
	// number of items in gcIndex equals or exceeds it.
	Capacity uint64
	// MetricsPrefix defines a prefix for metrics names.
	MetricsPrefix string
}

// New returns a new DB.  All fields and indexes are initialized
// and possible conflicts with schema from existing database is checked.
// One goroutine for writing batches is created.
func New(path string, o *Options) (db *DB, err error) {
	if o == nil {
		// default options
		o = &Options{
			Capacity: 5000000,
		}
	}
	db = &DB{
		capacity: o.Capacity,
		// channel collectGarbageTrigger
		// needs to be buffered with the size of 1
		// to signal another event if it
		// is triggered during already running function
		close: make(chan struct{}),
	}
	if db.capacity <= 0 {
		db.capacity = defaultCapacity
	}

	db.shed, err = shed.NewDB(path, o.MetricsPrefix)
	if err != nil {
		return nil, err
	}
	// Identify current storage schema by arbitrary name.
	db.schemaName, err = db.shed.NewStringField("schema-name")
	if err != nil {
		return nil, err
	}

	// Functions for retrieval data index.
	/*	var (
		encodeValueFunc func(fields shed.Item) (value []byte, err error)
		decodeValueFunc func(keyItem shed.Item, value []byte) (e shed.Item, err error)
	)*/

	/*db.pushIndex, err = db.shed.NewIndex("StoreTimestamp|Hash->Tags", shed.IndexFuncs{
		EncodeKey: func(fields shed.Item) (key []byte, err error) {
			key = make([]byte, 40)
			binary.BigEndian.PutUint64(key[:8], uint64(fields.StoreTimestamp))
			copy(key[8:], fields.Address[:])
			return key, nil
		},
		DecodeKey: func(key []byte) (e shed.Item, err error) {
			e.Address = key[8:]
			e.StoreTimestamp = int64(binary.BigEndian.Uint64(key[:8]))
			return e, nil
		},
		EncodeValue: func(fields shed.Item) (value []byte, err error) {
			valueBuffer := []byte{}
			buf := make([]byte, binary.MaxVarintLen64)

			for _, v := range fields.Tags {
				n := binary.PutUvarint(buf, v)
				if n > 0 {
					valueBuffer = append(valueBuffer, buf[:n]...)
				}
			}

			return valueBuffer, nil
		},
		DecodeValue: func(keyItem shed.Item, value []byte) (e shed.Item, err error) {
			for {
				v, n := binary.Uvarint(value)
				if n > 0 {
					e.Tags = append(e.Tags, v)
				} else {
					break
				}
				value = value[n:]
			}
			return e, nil
		},
	})
	if err != nil {
		return nil, err
	}*/

	db.tagIndex, err = db.shed.NewGenericIndex("Tag->UploadTime|UploadName", shed.GenericIndexFuncs{
		EncodeKey: func(fields interface{}) (key []byte, err error) {
			return nil, nil
		},
		DecodeKey: func(key []byte) (e interface{}, err error) {
			return nil, nil
		},
		EncodeValue: func(fields interface{}) (value []byte, err error) {
			return nil, nil
		},
		DecodeValue: func(keyItem interface{}, value []byte) (e interface{}, err error) {
			return nil, nil
		},
	})
	if err != nil {
		return nil, err
	}
	return db, err
}

// Close closes the underlying database.
func (db *DB) Close() (err error) {
	close(db.close)

	return db.shed.Close()
}

// chunkToItem creates new Item with data provided by the Chunk.
func chunkToItem(ch chunk.Chunk) shed.Item {
	return shed.Item{
		Address: ch.Address(),
		Data:    ch.Data(),
		Tags:    ch.Tags(),
	}
}

// addressToItem creates new Item with a provided address.
func addressToItem(addr chunk.Address) shed.Item {
	return shed.Item{
		Address: addr,
	}
}

// now is a helper function that returns a current unix timestamp
// in UTC timezone.
// It is set in the init function for usage in production, and
// optionally overridden in tests for data validation.
var now func() int64

func init() {
	// set the now function
	now = func() (t int64) {
		return time.Now().UTC().UnixNano()
	}
}
