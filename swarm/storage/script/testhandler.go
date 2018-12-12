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

package script

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage"
)

// NewTestHandler creates a mock Handler object to be used for testing purposes.
func NewTestHandler(t *testing.T) (handler Handler, cleanup func()) {
	path, err := ioutil.TempDir("", "bzzscript-test")
	if err != nil {
		t.Fatal(err)
	}

	localstoreparams := storage.NewDefaultLocalStoreParams()
	localstoreparams.Init(path)
	localStore, err := storage.NewLocalStore(localstoreparams, nil)
	if err != nil {
		t.Fatalf("localstore create fail, path %s: %v", path, err)
	}

	handler = NewHandler(&HandlerParams{
		ChunkStore: localStore,
	})
	localStore.Validators = append(localStore.Validators, handler)

	return handler, func() {
		os.RemoveAll(path)
	}
}
