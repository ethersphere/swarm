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

package feed

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
)

const (
	testDbDirName = "feeds"
)

// TestHandler is a stuct to be used for testing purposes
type TestHandler struct {
	*Handler
}

// NewTestHandler creates a new TestHandler
func NewTestHandler(datadir string, params *HandlerParams) (*TestHandler, error) {
	path := filepath.Join(datadir, testDbDirName)
	fh := NewHandler(params)

	db, err := localstore.New(filepath.Join(path, "chunks"), make([]byte, 32), nil)
	if err != nil {
		return nil, err
	}

	localStore := chunk.NewValidatorStore(db, storage.NewContentAddressValidator(storage.MakeHashFunc(feedsHashAlgorithm)), fh)

	netStore := storage.NewNetStore(localStore, network.NewBzzAddr(make([]byte, 32), nil))
	netStore.RemoteGet = func(ctx context.Context, req *storage.Request, localID enode.ID) (*enode.ID, func(), error) {
		return nil, func() {}, errors.New("not found")
	}
	fh.SetStore(netStore)
	return &TestHandler{fh}, nil
}

// NewTestHandlerWithStore creates a new TestHandler with a set chunk store
func NewTestHandlerWithStore(datadir string, db chunk.Store, params *HandlerParams) (*TestHandler, error) {
	fh := NewHandler(params)
	return newTestHandlerWithStore(fh, datadir, db, params)
}

func newTestHandlerWithStore(fh *Handler, datadir string, db chunk.Store, params *HandlerParams) (*TestHandler, error) {
	localStore := chunk.NewValidatorStore(db, storage.NewContentAddressValidator(storage.MakeHashFunc(feedsHashAlgorithm)), fh)

	netStore := storage.NewNetStore(localStore, network.NewBzzAddr(make([]byte, 32), nil))
	netStore.RemoteGet = func(ctx context.Context, req *storage.Request, localID enode.ID) (*enode.ID, func(), error) {
		return nil, func() {}, errors.New("not found")
	}
	fh.SetStore(netStore)
	return &TestHandler{fh}, nil
}

// Close closes the chunk store field for the test handler receiver
func (t *TestHandler) Close() {
	t.chunkStore.Close()
}

// DummyHandler is a test handler with dummy Lookup and GetContent funcs
type DummyHandler struct {
	cacheEntry *cacheEntry // mock content for feed retrieval
}

// NewDummyHandler initializes and returns a DummyHandler struct
func NewDummyHandler() *DummyHandler {
	d := &DummyHandler{}
	d.cacheEntry = &cacheEntry{}

	return d
}

// Lookup is the dummy func for DummyHandler structs
func (d *DummyHandler) Lookup(ctx context.Context, query *Query) (*cacheEntry, error) {
	return d.cacheEntry, nil
}

// GetContent is the dummy func for DummyHandler structs
func (d *DummyHandler) GetContent(feed *Feed) (storage.Address, []byte, error) {
	return d.cacheEntry.lastKey, d.cacheEntry.data, nil
}

// SetContent sets the binary data to be returned by a dummy handler
func (d *DummyHandler) SetContent(b []byte) {
	key := chunk.Address([]byte{1}) // dummy key
	d.cacheEntry.lastKey = key
	d.cacheEntry.data = b
}
