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
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

// NetStore
type NetStore struct {
	mu       sync.Mutex
	store    ChunkStore
	requests *lru.Cache
	retrieve func(ctx context.Context, r *Request) (chan struct{}, error)
}

func NewNetStore(store ChunkStore, retrieve func(ctx context.Context, r *Request) (chan struct{}, error)) (*NetStore, error) {
	requests, err := lru.New(defaultChunkRequestsCacheCapacity)
	if err != nil {
		return nil, err
	}
	return &NetStore{
		store:    store,
		requests: requests,
		retrieve: retrieve,
	}, nil
}

// Has checks if chunk with hash address ref is stored locally
// if not it returns a fetcher function to be called with a context
// block until item is stored
func (n *NetStore) Has(ref Address) (func(context.Context) error, error) {
	chunk, fetch, err := n.Get(ref)
	if chunk != nil {
		return nil, nil
	}
	return func(ctx context.Context) error {
		_, err = fetch(ctx)
		// TODO: exact logic for waiting till stored
		return err
	}, nil
}

// Get attempts at retrieving the chunk from LocalStore
// if it is not found, attempts at retrieving an existing requests
// if none exists, creates one and saves it in the requests cache
// From here on, all Get will hit on this request until the chunk is delivered
// or all request contexts are done
// it returns a chunk, a fetcher function and an error
// if chunk is nil, fetcher needs to be called with a context to return the chunk
func (n *NetStore) Get(ref Address) (Chunk, func(context.Context) (Chunk, error), error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	chunk, err := n.store.Get(ref)
	if err == nil {
		return chunk, nil, nil
	}
	request, err := n.getOrCreateRequest(ref)
	if err != nil {
		return nil, nil, err
	}
	f := func(ctx context.Context) (Chunk, error) {
		return n.request(ctx, request)
	}
	return nil, f, nil
}

// getOrCreateRequest attempts at retrieving an existing requests
// if none exists, creates one and saves it in the requests cache
// caller must hold the lock
func (n *NetStore) getOrCreateRequest(ref Address) (*Request, error) {
	ch, ok := n.requests.Get(ref)
	if ok {
		return ch.(*Request), nil
	}
	r := NewRequest()
	n.requests.Add(ref, r)
	return r, nil
}

func Get(ctx context.Context, dpa DPA, ref Address) (Chunk, error) {
	chunk, fetch, err := dpa.Get(ref)
	if err != nil {
		return nil, err
	}
	if chunk != nil {
		return chunk, nil
	}
	return fetch(ctx)
}

// Put is the entrypoint for local store requests coming from storeLoop

// Put stores a chunk in localstore, manages the request for the chunk if exists
// by closing the ReqC channel
func (n *NetStore) Put(ch Chunk) (func(ctx context.Context) error, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	defer func() {
		r, _ := n.requests.Get(ch.Address())
		if r != nil {
			r.(*Request).chunk = ch.Chunk()
		}
	}()
	wait, err := n.store.Put(ch)
	if err != nil {
		return nil, err
	}
	return wait, nil
}

// Close chunk store
func (n *NetStore) Close() {
	n.store.Close()
}

func (r *Request) SetData(addr Address, data []byte) {
	r.chunk = NewChunk(addr, data)
}

// request is a fetcher function to be called
// it launches the fetching only once by calling
// the retrieve function
func (n *NetStore) request(ctx context.Context, r *Request) (Chunk, error) {
	r.wg.Add(1)
	n.run(ctx, r)
	defer r.wg.Done()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.deliveredC:
		return r.chunk, nil
	}
}

// run prepares the request
// it keeps the request alive by rerequesting
// * after a search timeouted if request was successful
// * after retryInterval if request was unsuccessful
// * if an upstream sync client offers the chunk
func (n *NetStore) run(ctx context.Context, r *Request) {
	wait := time.NewTimer(0)
	var doRetrieve bool
	var waitC <-chan time.Time
	var quitC chan struct{}
	var err error

	// wait till all actual requests a closed
	go func() {
		r.wg.Wait()
		close(r.quitC)
	}()

	// loop that keeps the request alive
	go func() {
		// remove the request from the cache when all requests
		// contexts closed
		defer func() {
			n.requests.Remove(r.chunk.Address())
		}()
	F:
		for {
			if doRetrieve {
				quitC, err = n.retrieve(ctx, r)
				if err != nil {
					// retrieve error, wait before retry
					wait.Reset(retryInterval)
				} else {
					// otherwise wait for response
					wait.Reset(searchTimeout)
				}
				waitC = wait.C
				doRetrieve = false
			}
			select {
			case ctx = <-r.triggerC:
				//
				if ctx != nil && r.wanted {
					doRetrieve = true

				}
				if ctx == nil && !r.wanted {
					r.wanted = true
					doRetrieve = true
				}
			case <-waitC:
				// search or retry timeout; rerequest
				doRetrieve = true
			case <-quitC:
				// requested downstream peer disconnected; rerequest
				doRetrieve = true
			case <-r.quitC:
				// all request context closed, can quit
				break F
			}
		}
	}()

}
