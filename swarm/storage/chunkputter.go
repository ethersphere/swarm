// Copyright 2017 The go-ethereum Authors
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
)

// chunkPutter implements the ChunkStorer (later chunker.Store) interface
type chunkPutter struct {
	ChunkStore
	storeF func(addr []byte, meta []byte, data []byte)
}

func (cp *chunkPutter) Store(addr []byte, meta []byte, data []byte) {
	go cp.storeF(addr, meta, data)
}

type chunkData []byte

func (cp *chunkPutter) store(ctx context.Context, addr []byte, meta []byte, data []byte) error {
	return cp.Put(ctx, NewChunk(Address(addr), chunkData(append(meta, data...))))
}

// newChunkPutterWithErrors
func newChunkPutterWithErrors(ctx context.Context, cs ChunkStore, errc chan error) *chunkPutter {
	cp := &chunkPutter{
		ChunkStore: cs,
	}
	cp.storeF = func(addr []byte, meta []byte, data []byte) {
		errc <- cp.store(ctx, addr, meta, data)
	}
	return cp
}

// newChunkPutterWithErrors
func newChunkPutter(ctx context.Context, cs ChunkStore, errc chan error) *chunkPutter {
	cp := &chunkPutter{
		ChunkStore: cs,
	}
	cp.storeF = func(addr []byte, meta []byte, data []byte) {
		cp.store(ctx, addr, meta, data)
	}
	return cp
}
