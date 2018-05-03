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
	"path/filepath"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/storage/mock"
)

var (
	dbStorePutCounter = metrics.NewRegisteredCounter("storage.db.dbstore.put.count", nil)
)

type LocalStoreParams struct {
	*StoreParams
	ChunkDbPath string
	Validators  []ChunkValidator `toml:"-"`
}

func NewDefaultLocalStoreParams() *LocalStoreParams {
	return &LocalStoreParams{
		StoreParams: NewDefaultStoreParams(),
	}
}

//this can only finally be set after all config options (file, cmd line, env vars)
//have been evaluated
func (self *LocalStoreParams) Init(path string) {
	if self.ChunkDbPath == "" {
		self.ChunkDbPath = filepath.Join(path, "chunks")
	}
}

// LocalStore is a combination of inmemory db over a disk persisted db
// implements a Get/Put with fallback (caching) logic using any 2 ChunkStores
type LocalStore struct {
	Validators []ChunkValidator
	memStore   *MemStore
	DbStore    *LDBStore
	mu         sync.Mutex
}

// This constructor uses MemStore and DbStore as components
func NewLocalStore(params *LocalStoreParams, mockStore *mock.NodeStore) (*LocalStore, error) {
	ldbparams := NewLDBStoreParams(params.StoreParams, params.ChunkDbPath)
	dbStore, err := NewMockDbStore(ldbparams, mockStore)
	if err != nil {
		return nil, err
	}
	return &LocalStore{
		memStore:   NewMemStore(params.StoreParams, dbStore),
		DbStore:    dbStore,
		Validators: params.Validators,
	}, nil
}

func NewTestLocalStoreForAddr(params *LocalStoreParams) (*LocalStore, error) {
	ldbparams := NewLDBStoreParams(params.StoreParams, params.ChunkDbPath)
	dbStore, err := NewLDBStore(ldbparams)
	if err != nil {
		return nil, err
	}
	localStore := &LocalStore{
		memStore:   NewMemStore(params.StoreParams, dbStore),
		DbStore:    dbStore,
		Validators: params.Validators,
	}
	return localStore, nil
}

// Put is responsible for doing validation and storage of the chunk
// by using configured ChunkValidators, MemStore and LDBStore.
// If the chunk is not valid, its GetErrored function will
// return ErrChunkInvalid.
// This method will check if the chunk is already in the MemStore
// and it will return it if it is. If there is an error from
// the MemStore.Get, it will be returned by calling GetErrored
// on the chunk.
// This method is responsible for closing Chunk.ReqC channel
// when the chunk is stored in memstore.
// After the LDBStore.Put, it is ensured that the MemStore
// contains the chunk with the same data, but nil ReqC channel.
func (self *LocalStore) Put(ctx context.Context, chunk Chunk) (func(ctx context.Context) error, error) {
	valid := true
	for _, v := range self.Validators {
		if valid = v.Validate(chunk.Address(), chunk.Data()); valid {
			break
		}
	}
	if !valid {
		return nil, ErrChunkInvalid
	}

	log.Trace("localstore.put", "key", chunk.Address())
	self.mu.Lock()
	defer self.mu.Unlock()

	_, err := self.memStore.Get(chunk.Address())
	if err == nil {
		return nil, nil
	}
	if err != nil && err != ErrChunkNotFound {
		return nil, err
	}
	dbStorePutCounter.Inc(1)
	wait, err := self.DbStore.Put(ctx, chunk)
	if err != nil {
		return nil, err
	}
	return wait, nil
}

// Get(chunk *Chunk) looks up a chunk in the local stores
// This method is blocking until the chunk is retrieved
// so additional timeout may be needed to wrap this call if
// ChunkStores are remote and can have long latency
func (self *LocalStore) Get(_ context.Context, key Address) (chunk *chunk, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.get(key)
}

func (self *LocalStore) get(key Address) (chunk *chunk, err error) {
	chunk, err = self.memStore.Get(key)
	if err != nil && err != ErrChunkNotFound {
		return nil, err
	}
	chunk, err = self.DbStore.Get(key)
	if err != nil {
		return nil, err
	}
	self.memStore.Put(chunk)
	return chunk, nil
}

// Close the local store
func (self *LocalStore) Close() {
	self.DbStore.Close()
}
