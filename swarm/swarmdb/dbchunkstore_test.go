// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package swarmdb_test

import (
	"bytes"
	"fmt"
	"swarmdb"
	"testing"
	"time"
)

func BenchmarkStoreChunk(b *testing.B) {
	config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	swarmdb.NewKeyManager(config)
	u := config.GetSWARMDBUser()

	store, err := swarmdb.NewDBChunkStore(config, swarmdb.NewNetstats(config))
	if err != nil {
		fmt.Printf("%s\n", err)
		b.Fatal("Failure to open NewDBChunkStore")
	}

	b.ResetTimer()
	enc := 0
	for i := 0; i < b.N; i++ {
		r := []byte(fmt.Sprintf("randombytes%s-%d", time.Now(), i))
		v := make([]byte, 4096)
		copy(v, r)
		_, err := store.StoreChunk(u, v, enc)
		if err != nil {
			fmt.Printf("%s\n", err)
			b.Fatal("StoreChunk")
		}
	}
}

func TestDBChunkStore(t *testing.T) {
	config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	swarmdb.NewKeyManager(config)
	u := config.GetSWARMDBUser()

	store, err := swarmdb.NewDBChunkStore(config, swarmdb.NewNetstats(config))
	if err != nil {
		t.Fatal("Failure to open NewDBChunkStore")
	}

	// StoreChunk
	for enc := 0; enc < 2; enc++ {
		r := []byte(fmt.Sprintf("randombytes%s-%d", time.Now(), enc))
		v := make([]byte, 4096)
		copy(v, r)

		k, err := store.StoreChunk(u, r, enc)
		if err == nil {
			t.Fatal("Failure to generate StoreChunk Err", k, v)
		} else {
			fmt.Printf("SUCCESS in StoreChunk Err (input only has %d bytes)\n", len(r))
		}

		k, err1 := store.StoreChunk(u, v, enc)
		if err1 != nil {
			t.Fatal("Failure to StoreChunk", k, v, err1)
		} else {
			fmt.Printf("SUCCESS in StoreChunk:  %x => %v\n", string(k), string(v))
		}
		// RetrieveChunk
		val, err := store.RetrieveChunk(u, k)
		if err != nil {
			fmt.Printf("%s\n", err)
			t.Fatal("Failure to RetrieveChunk: Failure to retrieve", k, v, val)
		}
		if bytes.Compare(val, v) != 0 {
			t.Fatal("Failure to RetrieveChunk: Incorrect match", k, v, val)
		} else {
			fmt.Printf("SUCCESS in RetrieveChunk:  %x => %v\n", string(k), string(v))
		}
	}
}
