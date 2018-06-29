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
package storage

import (
	"context"
	"time"
)

var putTimeout = 30 * time.Second

// PutChunks adds chunks  to localstore
// It waits for receive on the stored channel
// It logs but does not fail on delivery error
func PutChunks(store *LocalStore, chunks ...Chunk) error {
	i := 0
	f := func(n int64) Chunk {
		chunk := chunks[i]
		i++
		return chunk
	}
	_, err := put(store, len(chunks), f)
	return err
}

func put(store *LocalStore, n int, f func(i int64) Chunk) (hs []Address, err error) {
	// put to localstore and wait for stored channel
	// does not check delivery error state
	done := make(chan struct{})
	errc := make(chan error)
	ctx, _ := context.WithTimeout(context.Background(), putTimeout)
	// defer cancel()
	defer close(done)
	for i := int64(0); i < int64(n); i++ {
		chunk := f(DefaultChunkSize)
		go func() {
			select {
			case errc <- store.Put(ctx, chunk):
			case <-done:
			}
		}()
		hs = append(hs, chunk.Address())
	}

	// wait for all chunks to be stored
	for i := 0; i < n; i++ {
		err := <-errc
		if err != nil {
			return nil, err
		}
	}
	return hs, nil
}
