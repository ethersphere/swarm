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

type TestHandler struct {
	*Handler
}

func (t *TestHandler) Close() {
	t.chunkStore.Close()
}

// NewTestHandler creates Handler object to be used for testing purposes.
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
