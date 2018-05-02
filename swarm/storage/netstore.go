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

	lru "github.com/hashicorp/golang-lru"
)

// NetStore
type NetStore struct {
	localStore *LocalStore
	requests   *lru.Cache
	retrieve   func(chunk *Chunk) error
}

func NewNetStore(localStore *LocalStore, retrieve func(chunk *Chunk) error) *NetStore {
	return &NetStore{localStore, retrieve}
}

// Has checks if chunk with hash address ref is stored locally
// if not it returns a fetcher function to be called with a context
// block until item is stotef
func (n *NetStore) Has(ctx context.Context, ref Address) (func(context.Context) error, error) {
	chunk, fetch, err := n.Get(ctx, ref)
	if chunk != nil {
		return nil
	}
	return func(c context.Context) error {
		chunk, err = fetch(c)
		if err != nil {
			return err
		}
		chunk.Stored()
	}()
}

// Get attempts at retrieving the chunk from LocalStore
// if it is not found, attempts at retrieving an existing requests
// if none exists, creates one and saves it in the requests cache
// From here on, all Get will hit on this request until the chunk is delivered
// or all request contexts are done
// it returns a chunk, a fetcher function and an error
// if chunk is nil, fetcher needs to be called with a context to return the chunk
func (n *NetStore) Get(ctx context.Context, ref Address) (Chunk, func(ctx context.Context) (Chunk, error), error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	chunk, err := n.LocalStore.Get(ctx, ref)
	if err == nil {
		return chunk, nil, nil
	}
	request, err := n.getOrCreateRequest(ref)
	if err != nil {
		return nil, nil, err
	}
	return nil, request.Run, nil
}

// getOrCreateRequest attempts at retrieving an existing requests
// if none exists, creates one and saves it in the requests cache
// caller must hold the lock
func (n *NetStore) getOrCreateRequest(ref Address) (*Request, error) {
	r, err := n.requests.Get(ref)
	if err == nil {
		return r, err
	}
	r = NewRequest(n)
	err = n.requests.Add(ref, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Put is the entrypoint for local store requests coming from storeLoop

// Put stores a chunk in localstore, manages the request for the chunk if exists
// by closing the ReqC channel
func (n *NetStore) Put(ctx context.Context, ch Chunk) (func(ctx context.Context) error, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	defer func() {
		r := n.requests.Get(ch.Ref)
		if r != nil {
			r.chunk = ch
			close(r.deliveredC)
			n.requests.Remove(ch.Ref)
		}
	}()
	waitToStore, err := n.LocalStore.Put(ch)
	if err != nil {
		return nil, err
	}
	return waitToStore, nil
}

// Close chunk store
func (self *NetStore) Close() {}
