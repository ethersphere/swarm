// Copyright 2019 The go-ethereum Authors
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

package localstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/shed"
)

// TestExportImport constructs two databases, one to put and export
// chunks and another one to import and validate that all chunks are
// imported.
func TestExportImport(t *testing.T) {
	db1, cleanup1 := newTestDB(t, nil)
	defer cleanup1()

	var (
		chunkCount  = exportTagsFileLimit * 3
		pinnedCount = exportTagsFileLimit*2 + 10
	)

	chunks := make(map[string][]byte, chunkCount)
	pinned := make(map[string]uint64, pinnedCount)
	for i := 0; i < chunkCount; i++ {
		ch := generateTestRandomChunk()

		_, err := db1.Put(context.Background(), chunk.ModePutUpload, ch)
		if err != nil {
			t.Fatal(err)
		}
		chunks[string(ch.Address())] = ch.Data()
		if i < pinnedCount {
			count := uint64(i%84) + 1 // arbitrary pin counter
			for i := uint64(0); i < count; i++ {
				err := db1.Set(context.Background(), chunk.ModeSetPin, ch.Address())
				if err != nil {
					t.Fatal(err)
				}
			}
			pinned[string(ch.Address())] = count
		}
	}

	validtePins := func(t *testing.T, db *DB) {
		t.Helper()

		var got int
		err := db1.pinIndex.Iterate(func(item shed.Item) (stop bool, err error) {
			count, ok := pinned[string(item.Address)]
			if !ok {
				return true, errors.New("chunk not pinned")
			}
			if count != item.PinCounter {
				return true, fmt.Errorf("got pin count %v for chunk %x, want %v", item.PinCounter, item.Address, count)
			}
			got++
			return false, nil
		}, nil)
		if err != nil {
			t.Fatal(err)
		}

		if got != pinnedCount {
			t.Fatalf("got pinned chunks %v, want %v", got, pinnedCount)
		}
	}

	validtePins(t, db1)

	var buf bytes.Buffer

	c, err := db1.Export(&buf)
	if err != nil {
		t.Fatal(err)
	}
	wantChunksCount := int64(len(chunks))
	if c != wantChunksCount {
		t.Errorf("got export count %v, want %v", c, wantChunksCount)
	}

	db2, cleanup2 := newTestDB(t, nil)
	defer cleanup2()

	c, err = db2.Import(&buf, false)
	if err != nil {
		t.Fatal(err)
	}
	if c != wantChunksCount {
		t.Errorf("got import count %v, want %v", c, wantChunksCount)
	}

	for a, want := range chunks {
		addr := chunk.Address([]byte(a))
		ch, err := db2.Get(context.Background(), chunk.ModeGetRequest, addr)
		if err != nil {
			t.Fatal(err)
		}
		got := ch.Data()
		if !bytes.Equal(got, want) {
			t.Fatalf("chunk %s: got data %x, want %x", addr.Hex(), got, want)
		}
	}

	validtePins(t, db2)
}
