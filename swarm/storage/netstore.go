// Copyright 2016 The go-ethereum Authors
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

package storage

import (
	"context"
	"encoding/hex"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	lru "github.com/hashicorp/golang-lru"
)

// NetStore
type NetStore struct {
	mu       sync.Mutex
	store    ChunkStore
	fetchers *lru.Cache
	// retrieve func(Request, *Fetcher) (context.Context, error)
	newFetcher func(addr Address) Fetcher
}

func NewNetStore(store ChunkStore, newFetcher func(addr Address) Fetcher) (*NetStore, error) {
	fetchers, err := lru.New(defaultChunkRequestsCacheCapacity)
	if err != nil {
		return nil, err
	}
	return &NetStore{
		store:      store,
		fetchers:   fetchers,
		newFetcher: newFetcher,
	}, nil
}

// Put stores a chunk in localstore, returns a wait function to wait for
// storage unless it is found
func (n *NetStore) Put(ch Chunk) (func(ctx context.Context) error, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	wait, err := n.store.Put(ch)
	if err != nil {
		return nil, err
	}
	// if chunk was already in store (wait f is nil)
	if wait == nil {
		return nil, nil
	}
	// if chunk is now put in store, check if there was an active fetcher
	f, _ := n.fetchers.Get(ch.Address())
	// if there is, deliver the chunk to requestors via fetcher
	if f != nil {
		f.(*fetcher).deliver(ch)
	}
	return wait, nil
}

// Has checks if chunk with hash address ref is stored locally
// if not it returns a fetcher function to be called with a context
// block until item is stored
func (n *NetStore) Has(ref Address) (func(Request) error, error) {
	chunk, fetch, err := n.get(ref)
	if chunk != nil {
		return nil, nil
	}
	wait := func(rctx Request) error {
		_, err = fetch(rctx)
		// TODO: exact logic for waiting till stored
		return err
	}
	return wait, nil
}

// get attempts at retrieving the chunk from LocalStore
// if it is not found, attempts at retrieving an existing Fetchers
// if none exists, creates one and saves it in the Fetchers cache
// From here on, all Get will hit on this Fetcher until the chunk is delivered
// or all Fetcher contexts are done
// it returns a chunk, a fetcher function and an error
// if chunk is nil, fetcher needs to be called with a context to return the chunk
func (n *NetStore) get(ref Address) (Chunk, func(context.Context) (Chunk, error), error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	chunk, err := n.store.Get(ref)
	if err == nil {
		return chunk, nil, nil
	}
	f, err := n.getOrCreateFetcher(ref)
	if err != nil {
		return nil, nil, err
	}
	return nil, f.fetch, nil
}

// getOrCreateFetcher attempts at retrieving an existing Fetchers
// if none exists, creates one and saves it in the Fetchers cache
// caller must hold the lock
func (n *NetStore) getOrCreateFetcher(ref Address) (*fetcher, error) {
	key := hex.EncodeToString(ref)
	log.Debug("getOrCreateFetcher", "fetchers.Len", n.fetchers.Len())
	f, ok := n.fetchers.Get(key)
	if ok {
		log.Debug("getOrCreateFetcher found fetcher")
		return f.(*fetcher), nil
	}
	log.Debug("getOrCreateFetcher new fetcher")
	fetcher := newFetcher(n.newFetcher(ref), func() {
		n.fetchers.Remove(key)
	})
	n.fetchers.Add(key, fetcher)

	return fetcher, nil
}

// Get retrieves the chunk from the NetStore DPA synchronously
// it calls NetStore get. If the chunk is not in local Storage
// it calls fetch with the request, which blocks until the chunk
// arrived or context is done
func (n *NetStore) Get(rctx Request, ref Address) (Chunk, error) {
	chunk, fetch, err := n.get(ref)
	if err != nil {
		return nil, err
	}
	if chunk != nil {
		return chunk, nil
	}
	return fetch(rctx)
}

// // waitAndRemoveFetcher waits till all actual Requests are closed, removes the Fetcher from fetchers
// // and stops the fetcher.
// func (n *NetStore) waitAndRemoveFetcher(f Fetcher) {
// 	log.Debug("waitAndRemoveFetcher started")
// 	f.wait()
// 	log.Debug("waitAndRemoveFetcher after wait")
// 	log.Warn("remove fetcher")
// 	n.fetchers.Remove(hex.EncodeToString(f.addr))
// 	f.stop()
// }

// Close chunk store
func (n *NetStore) Close() {
	n.store.Close()
}

// Request is an extention of context.Context and is handed by client to the fetcher
type Request interface {
	context.Context
	Address() Address
}

type Fetcher interface {
	Fetch(rctx context.Context) (Chunk, error)
	Stop()
	Status() error
}

type fetcher struct {
	fetcher    Fetcher
	chunk      Chunk
	requestCnt int32
	deliveredC chan struct{} // chan signalling chunk delivery to requests
	remove     func()
}

func newFetcher(f Fetcher, remove func()) *fetcher {
	return &fetcher{
		fetcher:    f,
		deliveredC: make(chan struct{}),
		remove:     remove,
	}
}

func (f *fetcher) fetch(rctx context.Context) (Chunk, error) {
	atomic.AddInt32(&f.requestCnt, 1)
	log.Debug("fetch", "requestCnt", f.requestCnt)
	defer func() {
		log.Debug("fetch defer", "requestCnt", f.requestCnt)
		if atomic.AddInt32(&f.requestCnt, -1) == 0 {
			f.remove()
			f.fetcher.Stop()
		}
	}()

	f.fetcher.Fetch(rctx)

	select {
	case <-rctx.Done():
		log.Warn("context done")
		return nil, &errStatus{rctx.Err(), f.fetcher.Status()}
	case <-f.deliveredC:
		return f.chunk, nil
	}
}

func (f *fetcher) deliver(ch Chunk) {
	f.chunk = ch
	close(f.deliveredC)
}

type errStatus struct {
	error
	status error
}

func (e *errStatus) Status() error {
	return e.status
}
