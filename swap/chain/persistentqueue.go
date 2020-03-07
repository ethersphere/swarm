package chain

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethersphere/swarm/state"
)

/*
	persistentQueue represents a queue stored in a state store
	Items are enqueued by writing them to the state store with the timestamp as prefix and a nonce so that two items can be queued at the same time
	It provides a (blocking) Next function to wait for a new item to be available. Only a single call to Next may be active at any time
	To allow atomic operations with other state store operations all functions only write to batches instead of writing to the store directly
	The user must ensure that all functions (except Next) are called with the same lock held which is provided externally so multiple queues can use the same
	The queue provides no dequeue function. Instead an item must be deleted by its key
*/

// persistentQueue represents a queue stored in a state store
type persistentQueue struct {
	store   state.Store   // the store backing this queue
	prefix  string        // the prefix for the keys for this queue
	trigger chan struct{} // channel to notify the queue that a new item is available
	nonce   uint64        // increasing nonce. starts with 0 on every startup
}

// NewPersistentQueue creates a structure to interact with a queue with the given prefix
func newPersistentQueue(store state.Store, prefix string) *persistentQueue {
	return &persistentQueue{
		store:   store,
		prefix:  prefix,
		trigger: make(chan struct{}, 1),
		nonce:   0,
	}
}

// queue puts the necessary database operations for enqueueing a new item into the supplied batch
// It returns the generated key and a trigger function which must be called once the batch was successfully written
// This only returns an error if the encoding fails which is an unrecoverable error
// A lock must be held and kept until after the trigger function was called or the batch write failed
func (pq *persistentQueue) queue(b *state.StoreBatch, v interface{}) (key string, trigger func(), err error) {
	// the nonce guarantees keys don't collide if multiple transactions are queued in the same second
	pq.nonce++
	key = fmt.Sprintf("%d_%08d", time.Now().Unix(), pq.nonce)
	if err = b.Put(pq.prefix+key, v); err != nil {
		return "", nil, err
	}

	return key, func() {
		select {
		case pq.trigger <- struct{}{}:
		default:
		}
	}, nil
}

// peek looks at the next item in the queue
// The error returned is either a decode or an io error
// A lock must be held when this is called and should be held afterwards to prevent the item from being removed while processing
func (pq *persistentQueue) peek(i interface{}) (key string, exists bool, err error) {
	err = pq.store.Iterate(pq.prefix, func(k, data []byte) (bool, error) {
		key = string(k)
		unmarshaler, ok := i.(encoding.BinaryUnmarshaler)
		if !ok {
			return true, json.Unmarshal(data, i)
		}
		return true, unmarshaler.UnmarshalBinary(data)
	})
	if err != nil {
		return "", false, err
	}
	if key == "" {
		return "", false, nil
	}
	return strings.TrimPrefix(key, pq.prefix), true, nil
}

// Next looks at the next item in the queue and blocks until an item is available if there is none
// The error returned is either an decode error, an io error or a cancelled context
// No lock should not be held when this is called. Only a single call to next may be active at any time
// If the the key is not "", the value exists, the supplied lock was acquired and must be released by the caller after processing the item
// The supplied lock should be the same that is used for the other functions
func (pq *persistentQueue) next(ctx context.Context, i interface{}, lock *sync.Mutex) (key string, err error) {
	lock.Lock()
	key, exists, err := pq.peek(i)
	if exists {
		return key, nil
	}
	lock.Unlock()
	if err != nil {
		return "", err
	}

	for {
		select {
		case <-pq.trigger:
			lock.Lock()
			key, exists, err = pq.peek(i)
			if exists {
				return key, nil
			}
			lock.Unlock()
			if err != nil {
				return "", err
			}
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// Delete adds the batch operation to delete the queue element with the given key
// A lock must be held when the batch is written
func (pq *persistentQueue) delete(b *state.StoreBatch, key string) {
	b.Delete(pq.prefix + key)
}
