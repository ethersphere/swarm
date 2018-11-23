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
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// IndexItem holds fields relevant to Swarm Chunk data and metadata.
// All information required for swarm storage and operations
// on that storage must be defined here.
// This structure is logically connected to swarm storage,
// the only part of this package that is not generalized,
// mostly for performance reasons.
//
// IndexItem is a type that is used for retrieving, storing and encoding
// chunk data and metadata. It is passed as an argument to Index encoding
// functions, get function and put function.
// But it is also returned with additional data from get function call
// and as the argument in iterator function definition.
type IndexItem struct {
	Address         []byte
	Data            []byte
	AccessTimestamp int64
	StoreTimestamp  int64
	// UseMockStore is a pointer to identify
	// an unset state of the field in Join function.
	UseMockStore *bool
}

// Merge is a helper method to construct a new
// IndexItem by filling up fields with default values
// of a particular IndexItem with values from another one.
func (i IndexItem) Merge(i2 IndexItem) (new IndexItem) {
	if i.Address == nil {
		i.Address = i2.Address
	}
	if i.Data == nil {
		i.Data = i2.Data
	}
	if i.AccessTimestamp == 0 {
		i.AccessTimestamp = i2.AccessTimestamp
	}
	if i.StoreTimestamp == 0 {
		i.StoreTimestamp = i2.StoreTimestamp
	}
	if i.UseMockStore == nil {
		i.UseMockStore = i2.UseMockStore
	}
	return i
}

// Index represents a set of LevelDB key value pairs that have common
// prefix. It holds functions for encoding and decoding keys and values
// to provide transparent actions on saved data which inclide:
// - getting a particular IndexItem
// - saving a particular IndexItem
// - iterating over a sorted LevelDB keys
// It implements IndexIteratorInterface interface.
type Index struct {
	db              *DB
	prefix          []byte
	encodeKeyFunc   func(fields IndexItem) (key []byte, err error)
	decodeKeyFunc   func(key []byte) (e IndexItem, err error)
	encodeValueFunc func(fields IndexItem) (value []byte, err error)
	decodeValueFunc func(value []byte) (e IndexItem, err error)

	// triggers are used for signaling
	// subscriptions to continue iterations.
	triggers   map[uint64]chan struct{}
	triggersMu *sync.Mutex
}

// IndexFuncs structure defines functions for encoding and decoding
// LevelDB keys and values for a specific index.
type IndexFuncs struct {
	EncodeKey   func(fields IndexItem) (key []byte, err error)
	DecodeKey   func(key []byte) (e IndexItem, err error)
	EncodeValue func(fields IndexItem) (value []byte, err error)
	DecodeValue func(value []byte) (e IndexItem, err error)
}

// NewIndex returns a new Index instance with defined name and
// encoding functions. The name must be unique and will be validated
// on database schema for a key prefix byte.
func (db *DB) NewIndex(name string, funcs IndexFuncs) (f Index, err error) {
	id, err := db.schemaIndexPrefix(name)
	if err != nil {
		return f, err
	}
	prefix := []byte{id}
	return Index{
		db:     db,
		prefix: prefix,
		// This function adjusts Index LevelDB key
		// by appending the provided index id byte.
		// This is needed to avoid collisions between keys of different
		// indexes as all index ids are unique.
		encodeKeyFunc: func(e IndexItem) (key []byte, err error) {
			key, err = funcs.EncodeKey(e)
			if err != nil {
				return nil, err
			}
			return append(append(make([]byte, 0, len(key)+1), prefix...), key...), nil
		},
		// This function reverses the encodeKeyFunc constructed key
		// to transparently work with index keys without their index ids.
		// It assumes that index keys are prefixed with only one byte.
		decodeKeyFunc: func(key []byte) (e IndexItem, err error) {
			return funcs.DecodeKey(key[1:])
		},
		encodeValueFunc: funcs.EncodeValue,
		decodeValueFunc: funcs.DecodeValue,
		triggers:        make(map[uint64]chan struct{}),
		triggersMu:      new(sync.Mutex),
	}, nil
}

// Get accepts key fields represented as IndexItem to retrieve a
// value from the index and return maximum available information
// from the index represented as another IndexItem.
func (f Index) Get(keyFields IndexItem) (out IndexItem, err error) {
	key, err := f.encodeKeyFunc(keyFields)
	if err != nil {
		return out, err
	}
	value, err := f.db.Get(key)
	if err != nil {
		return out, err
	}
	out, err = f.decodeValueFunc(value)
	if err != nil {
		return out, err
	}
	return out.Merge(keyFields), nil
}

// Put accepts IndexItem to encode information from it
// and save it to the database.
func (f Index) Put(i IndexItem) (err error) {
	key, err := f.encodeKeyFunc(i)
	if err != nil {
		return err
	}
	value, err := f.encodeValueFunc(i)
	if err != nil {
		return err
	}
	return f.db.Put(key, value)
}

// PutInBatch is the same as Put method, but it just
// saves the key/value pair to the batch instead
// directly to the database.
func (f Index) PutInBatch(batch *leveldb.Batch, i IndexItem) (err error) {
	key, err := f.encodeKeyFunc(i)
	if err != nil {
		return err
	}
	value, err := f.encodeValueFunc(i)
	if err != nil {
		return err
	}
	batch.Put(key, value)
	return nil
}

// Delete accepts IndexItem to remove a key/value pair
// from the database based on its fields.
func (f Index) Delete(keyFields IndexItem) (err error) {
	key, err := f.encodeKeyFunc(keyFields)
	if err != nil {
		return err
	}
	return f.db.Delete(key)
}

// DeleteInBatch is the same as Delete just the operation
// is performed on the batch instead on the database.
func (f Index) DeleteInBatch(batch *leveldb.Batch, keyFields IndexItem) (err error) {
	key, err := f.encodeKeyFunc(keyFields)
	if err != nil {
		return err
	}
	batch.Delete(key)
	return nil
}

// IndexIterFunc is a callback on every IndexItem that is decoded
// by iterating on an Index keys.
// By returning a true for stop variable, iteration will
// stop, and by returning the error, that error will be
// propagated to the called iterator method on Index.
type IndexIterFunc func(item IndexItem) (stop bool, err error)

// IterateAll iterates over all keys of the Index.
func (f Index) IterateAll(fn IndexIterFunc) (err error) {
	it := f.db.NewIterator()
	defer it.Release()

	for ok := it.Seek(f.prefix); ok; ok = it.Next() {
		key := it.Key()
		if key[0] != f.prefix[0] {
			break
		}
		keyIndexItem, err := f.decodeKeyFunc(key)
		if err != nil {
			return err
		}
		valueIndexItem, err := f.decodeValueFunc(it.Value())
		if err != nil {
			return err
		}
		stop, err := fn(keyIndexItem.Merge(valueIndexItem))
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return it.Error()
}

// IterateFrom iterates over Index keys starting from the key
// encoded from the provided IndexItem.
func (f Index) IterateFrom(start IndexItem, fn IndexIterFunc) (err error) {
	startKey, err := f.encodeKeyFunc(start)
	if err != nil {
		return err
	}
	it := f.db.NewIterator()
	defer it.Release()

	for ok := it.Seek(startKey); ok; ok = it.Next() {
		key := it.Key()
		if key[0] != f.prefix[0] {
			break
		}
		keyIndexItem, err := f.decodeKeyFunc(key)
		if err != nil {
			return err
		}
		valueIndexItem, err := f.decodeValueFunc(it.Value())
		if err != nil {
			return err
		}
		stop, err := fn(keyIndexItem.Merge(valueIndexItem))
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return it.Error()
}

// Subscription provides methods to control
// and get information about subscription state.
type Subscription struct {
	stopChan chan struct{}
	onceOnce sync.Once
	doneChan chan struct{}
	err      error
	mu       sync.RWMutex
}

// Err returns an error that subscription encountered.
// It should be usually called after the Done is read from.
// It is safe to call this function multiple times.
func (s *Subscription) Err() (err error) {
	s.mu.RLock()
	err = s.err
	s.mu.RUnlock()
	return err
}

// Done returns a read-only channel that will be closed
// when the subscription is stopped or encountered an error.
func (s *Subscription) Done() <-chan struct{} {
	return s.doneChan
}

// Stop terminates the subscription without any error.
// It is safe to call this function multiple times.
func (s *Subscription) Stop() {
	s.onceOnce.Do(func() {
		close(s.stopChan)
	})
}

// NewSubscription starts a new subscription on the index.
// Subscribing is similar to iterating over the index key, but it
// is performend in the background and contiguously over the existing keys
// as well over the new keys when they are put. It is required to signal
// all the iterators to check new keys with TriggerSubscriptions method.
// This provides a greater control over subscription iterators instead to
// trigger subscriptions on Put method or batch writes.
// IndexIterFunc behaves the same as in iterate methods.
// Provided context allows cancellation of created goroutine and Subscription.Err()
// will return appropriate error from context.
func (f Index) NewSubscription(ctx context.Context, fn IndexIterFunc) (s *Subscription, err error) {
	return f.newSubscription(ctx, f.prefix, fn)
}

// NewSubscriptionFrom is the same method as NewSubscription, but it
// iterates over the keys from the provided start index.
func (f Index) NewSubscriptionFrom(ctx context.Context, start IndexItem, fn IndexIterFunc) (s *Subscription, err error) {
	startKey, err := f.encodeKeyFunc(start)
	if err != nil {
		return nil, err
	}
	return f.newSubscription(ctx, startKey, fn)
}

// newSubscription provides base functionality for NewSubscription
// and NewSubscriptionFrom methods.
// It creates a new goroutine which will iterate over existing keys of the index
// and create new iterators when TriggerSubscriptions is called.
func (f Index) newSubscription(ctx context.Context, startKey []byte, fn IndexIterFunc) (s *Subscription, err error) {
	// Create a subscription id to be able to remove the channel from the triggers map.
	f.triggersMu.Lock()
	// Generate new ID.
	id, err := f.newSubscriptionID()
	if err != nil {
		f.triggersMu.Unlock()
		return nil, err
	}
	// trigger is a one size buffered channel for two reasons
	// - to be able to signal the first iteration
	// - performance on sending to channel in TriggerSubscriptions
	trigger := make(chan struct{}, 1)
	f.triggers[id] = trigger
	f.triggersMu.Unlock()

	// send signal for the initial iteration
	trigger <- struct{}{}

	s = &Subscription{
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}

	go func() {
		// this error will be checked in defer
		// and set to Subscription.err
		var err error
		defer func() {
			if err != nil {
				s.mu.Lock()
				s.err = err
				s.mu.Unlock()
			}
			// signal that the subscription is done
			close(s.doneChan)
			f.triggersMu.Lock()
			// clean up the trigger channel
			delete(f.triggers, id)
			f.triggersMu.Unlock()
		}()
		// This flag identifies the first iteration to
		// include the start item in it, and to exclude the
		// start item in the next ones, as they are already sent
		// in previous iterations.
		firstIteration := true
		for {
			select {
			case <-trigger:
				// This closure is to provide a clean defer for
				// iteration release.
				err = func() error {
					it := f.db.NewIterator()
					defer it.Release()

					ok := it.Seek(startKey)
					if firstIteration {
						// The firs iteration will set the flag to false
						// to provide information to all next iterations.
						firstIteration = false
					} else {
						// All iterations but first will start from the
						// startKey+1 as the start key for all non-first
						// iterations is already processed.
						ok = it.Next()
					}

					for ; ok; ok = it.Next() {
						key := it.Key()
						if key[0] != f.prefix[0] {
							break
						}
						keyIndexItem, err := f.decodeKeyFunc(key)
						if err != nil {
							return err
						}
						valueIndexItem, err := f.decodeValueFunc(it.Value())
						if err != nil {
							return err
						}
						stop, err := fn(keyIndexItem.Merge(valueIndexItem))
						if err != nil {
							return err
						}
						startKey = key
						if stop {
							// Q: should the whole subscription stop or just this iteration?
							s.Stop()
							break
						}
						select {
						case <-s.stopChan:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						default:
						}
					}
					return it.Error()
				}()
				if err != nil {
					return
				}
			case <-s.stopChan:
				return
			case <-ctx.Done():
				if err == nil {
					err = ctx.Err()
				}
				return
			}
		}
	}()

	return s, nil
}

// newSubscriptionID generates a new subscription id as a random number.
func (f Index) newSubscriptionID() (id uint64, err error) {
	b := make([]byte, 8)
	newID := func() (uint64, error) {
		_, err = rand.Read(b)
		if err != nil {
			return 0, err
		}
		return binary.BigEndian.Uint64(b), nil
	}
	id, err = newID()
	if err != nil {
		return 0, err
	}
	// check up to 100 times if this id already exists
	for i := 0; i < 100; i++ {
		if _, ok := f.triggers[id]; !ok {
			// this id is unique, return it
			return id, nil
		}
		id, err = newID()
		if err != nil {
			return 0, err
		}
	}
	return 0, errors.New("unable to generate subscription id")
}

// TriggerSubscriptions signals to all index subscriptions
// that they should continue iterating over the index keys
// where they stopped in the last iteration. This method
// should be called when new data is put to the index.
// It is not automatically called by the index Put method
// to allow for optimizations and for signaling of batch
// writes.
func (f Index) TriggerSubscriptions() {
	for _, t := range f.triggers {
		select {
		case t <- struct{}{}:
		default:
		}
	}
}
