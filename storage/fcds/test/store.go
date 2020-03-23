// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package test

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	chunktesting "github.com/ethersphere/swarm/chunk/testing"
	"github.com/ethersphere/swarm/storage/fcds"
)

var (
	chunksFlag      = flag.Int("chunks", 100, "Number of chunks to use in tests.")
	concurrencyFlag = flag.Int("concurrency", 8, "Maximal number of parallel operations.")
)

// Main parses custom cli flags automatically on test runs.
func Main(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// RunAll runs all available tests for a Store implementation.
func RunAll(t *testing.T, newStoreFunc func(t *testing.T) (fcds.Storer, func())) {
	RunStd(t, newStoreFunc)

	t.Run("no grow", func(t *testing.T) {
		runNoGrow(t, newStoreFunc)
	})
}

// RunStd runs all standard tests that are agnostic to specific Store implementation details.
func RunStd(t *testing.T, newStoreFunc func(t *testing.T) (fcds.Storer, func())) {
	t.Run("empty", func(t *testing.T) {
		RunStore(t, &RunStoreOptions{
			ChunkCount:   *chunksFlag,
			NewStoreFunc: newStoreFunc,
		})
	})

	t.Run("cleaned", func(t *testing.T) {
		RunStore(t, &RunStoreOptions{
			ChunkCount:   *chunksFlag,
			NewStoreFunc: newStoreFunc,
			Cleaned:      true,
		})
	})

	for _, tc := range []struct {
		name        string
		deleteSplit int
	}{
		{
			name:        "delete-all",
			deleteSplit: 1,
		},
		{
			name:        "delete-half",
			deleteSplit: 2,
		},
		{
			name:        "delete-fifth",
			deleteSplit: 5,
		},
		{
			name:        "delete-tenth",
			deleteSplit: 10,
		},
		{
			name:        "delete-percent",
			deleteSplit: 100,
		},
		{
			name:        "delete-permill",
			deleteSplit: 1000,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			RunStore(t, &RunStoreOptions{
				ChunkCount:   *chunksFlag,
				DeleteSplit:  tc.deleteSplit,
				NewStoreFunc: newStoreFunc,
			})
		})
	}

	t.Run("iterator", func(t *testing.T) {
		RunIterator(t, newStoreFunc)
	})

}

func runNoGrow(t *testing.T, newStoreFunc func(t *testing.T) (fcds.Storer, func())) {
	defer func(s uint8) {
		fcds.ShardCount = s
	}(fcds.ShardCount)

	fcds.ShardCount = 4

	db, clean := newStoreFunc(t)

	defer clean()

	chunkCount := 1000
	chunks := getChunks(chunkCount)

	chunkShards := make(map[string]uint8)

	for _, ch := range chunks {
		if shard, err := db.Put(ch); err != nil {
			t.Fatal(err)
		} else {
			chunkShards[ch.Address().String()] = shard
		}
	}

	// delete 4,3,2,1 chunks from shards 0,1,2,3
	del := 4
	deleteChunks := []string{}

	for i := uint8(0); i < fcds.ShardCount; i++ {
		d := del
		for addr, storedOn := range chunkShards {
			if storedOn == i {

				// delete the chunk to make a free slot on the shard
				c := unmarshalAddressString(t, addr)

				if err := db.Delete(*c); err != nil {
					t.Fatal(err)
				}
				deleteChunks = append(deleteChunks, addr)
				if len(deleteChunks) == d {
					break
				}
			}
		}
		for _, v := range deleteChunks {
			delete(chunkShards, v)
		}
		deleteChunks = []string{}

		del--
	}

	ins := 4 + 3 + 2 + 1

	freeSlots := []int{4, 3, 2, 1}

	for i := 0; i < ins; i++ {
		cc := chunktesting.GenerateTestRandomChunk()
		if shard, err := db.Put(cc); err != nil {
			t.Fatal(err)
		} else {
			freeSlots[shard]--
			if freeSlots[shard] < 0 {
				t.Fatalf("shard %d slots went negative", shard)
			}
			chunkShards[cc.Address().String()] = shard
		}
	}

	slots, err := db.ShardSize()
	if err != nil {
		t.Fatal(err)
	}

	sum := 0
	for _, v := range slots {
		sum += int(v.Size)
	}

	if sum != 4096*1000 {
		t.Fatal(sum)
	}

	// now for each new chunk, we should first check all shard
	// sizes, and locate the smallest shard
	// for each Put we should get that shard as next

	insNew := 10000
	for i := 0; i < insNew; i++ {
		slots, err := db.ShardSize()
		if err != nil {
			t.Fatal(err)
		}

		minSize, minSlot := slots[0].Size, uint8(0)
		for i, v := range slots {
			// take the _last_ minimum
			if v.Size <= minSize {
				minSize = v.Size
				minSlot = uint8(i)
			}
		}
		cc := chunktesting.GenerateTestRandomChunk()
		if shard, err := db.Put(cc); err != nil {
			t.Fatal(err)
		} else {
			if shard != minSlot {
				t.Fatalf("next slot expected to be %d but got %d. chunk number %d", minSlot, shard, i)
			}
		}
	}
}

// RunStoreOptions define parameters for Store test function.
type RunStoreOptions struct {
	NewStoreFunc func(t *testing.T) (fcds.Storer, func())
	ChunkCount   int
	DeleteSplit  int
	Cleaned      bool
}

// RunStore tests a single Store implementation for its general functionalities.
// Subtests are deliberately separated into sections that can have timings
// printed on test runs for each of them.
func RunStore(t *testing.T, o *RunStoreOptions) {
	db, clean := o.NewStoreFunc(t)
	defer clean()

	chunks := getChunks(o.ChunkCount)

	if o.Cleaned {
		t.Run("clean", func(t *testing.T) {
			sem := make(chan struct{}, *concurrencyFlag)
			var wg sync.WaitGroup

			wg.Add(o.ChunkCount)
			for _, ch := range chunks {
				sem <- struct{}{}

				go func(ch chunk.Chunk) {
					defer func() {
						<-sem
						wg.Done()
					}()

					if _, err := db.Put(ch); err != nil {
						panic(err)
					}
				}(ch)
			}
			wg.Wait()

			wg = sync.WaitGroup{}

			wg.Add(o.ChunkCount)
			for _, ch := range chunks {
				sem <- struct{}{}

				go func(ch chunk.Chunk) {
					defer func() {
						<-sem
						wg.Done()
					}()

					if err := db.Delete(ch.Address()); err != nil {
						panic(err)
					}
				}(ch)
			}
			wg.Wait()
		})
	}

	rand.Shuffle(o.ChunkCount, func(i, j int) {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	})

	var deletedChunks sync.Map

	t.Run("write", func(t *testing.T) {
		sem := make(chan struct{}, *concurrencyFlag)
		var wg sync.WaitGroup
		var wantCount int
		var wantCountMu sync.Mutex
		wg.Add(o.ChunkCount)
		for i, ch := range chunks {
			sem <- struct{}{}

			go func(i int, ch chunk.Chunk) {
				defer func() {
					<-sem
					wg.Done()
				}()

				if _, err := db.Put(ch); err != nil {
					panic(err)
				}
				if o.DeleteSplit > 0 && i%o.DeleteSplit == 0 {
					if err := db.Delete(ch.Address()); err != nil {
						panic(err)
					}
					deletedChunks.Store(string(ch.Address()), nil)
				} else {
					wantCountMu.Lock()
					wantCount++
					wantCountMu.Unlock()
				}
			}(i, ch)
		}
		wg.Wait()
	})

	rand.Shuffle(o.ChunkCount, func(i, j int) {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	})

	t.Run("read", func(t *testing.T) {
		sem := make(chan struct{}, *concurrencyFlag)
		var wg sync.WaitGroup

		wg.Add(o.ChunkCount)
		for i, ch := range chunks {
			sem <- struct{}{}

			go func(i int, ch chunk.Chunk) {
				defer func() {
					<-sem
					wg.Done()
				}()

				got, err := db.Get(ch.Address())

				if _, ok := deletedChunks.Load(string(ch.Address())); ok {
					if err != chunk.ErrChunkNotFound {
						panic(fmt.Errorf("got error %v, want %v", err, chunk.ErrChunkNotFound))
					}
				} else {
					if err != nil {
						panic(fmt.Errorf("chunk %v %s: %v", i, ch.Address().Hex(), err))
					}
					if !bytes.Equal(got.Address(), ch.Address()) {
						panic(fmt.Errorf("got chunk %v address %x, want %x", i, got.Address(), ch.Address()))
					}
					if !bytes.Equal(got.Data(), ch.Data()) {
						panic(fmt.Errorf("got chunk %v data %x, want %x", i, got.Data(), ch.Data()))
					}
				}
			}(i, ch)
		}
		wg.Wait()
	})
}

// RunIterator validates behaviour of Iterate and Count methods on a Store.
func RunIterator(t *testing.T, newStoreFunc func(t *testing.T) (fcds.Storer, func())) {
	chunkCount := 1000

	db, clean := newStoreFunc(t)
	defer clean()

	chunks := getChunks(chunkCount)

	for _, ch := range chunks {
		if _, err := db.Put(ch); err != nil {
			t.Fatal(err)
		}
	}

	gotCount, err := db.Count()
	if err != nil {
		t.Fatal(err)
	}
	if gotCount != chunkCount {
		t.Fatalf("got %v count, want %v", gotCount, chunkCount)
	}

	var iteratedCount int
	if err := db.Iterate(func(ch chunk.Chunk) (stop bool, err error) {
		for _, c := range chunks {
			if bytes.Equal(c.Address(), ch.Address()) {
				if !bytes.Equal(c.Data(), ch.Data()) {
					t.Fatalf("invalid data in iterator for key %s", c.Address())
				}
				iteratedCount++
				return false, nil
			}
		}
		return false, nil
	}); err != nil {
		t.Fatal(err)
	}
	if iteratedCount != chunkCount {
		t.Fatalf("iterated on %v chunks, want %v", iteratedCount, chunkCount)
	}
}

// NewFCDSStore is a test helper function that constructs
// a new Store for testing purposes into which a specific MetaStore can be injected.
func NewFCDSStore(t *testing.T, path string, metaStore fcds.MetaStore) (s *fcds.Store, clean func()) {
	t.Helper()

	s, err := fcds.New(path, chunk.DefaultSize, metaStore)
	if err != nil {
		os.RemoveAll(path)
		t.Fatal(err)
	}
	return s, func() {
		s.Close()
		os.RemoveAll(path)
	}
}

// chunkCache reduces the work done by generating random chunks
// by getChunks function by keeping storing them for future reuse.
var chunkCache []chunk.Chunk

// getChunk returns a number of chunks with random data for testing purposes.
// By calling it multiple times, it will return same chunks from the cache.
func getChunks(count int) []chunk.Chunk {
	l := len(chunkCache)
	if l == 0 {
		chunkCache = make([]chunk.Chunk, count)
		for i := 0; i < count; i++ {
			chunkCache[i] = chunktesting.GenerateTestRandomChunk()
		}
		return chunkCache
	}
	if l < count {
		for i := 0; i < count-l; i++ {
			chunkCache = append(chunkCache, chunktesting.GenerateTestRandomChunk())
		}
		return chunkCache
	}
	return chunkCache[:count]
}

func unmarshalAddressString(t *testing.T, s string) *chunk.Address {
	t.Helper()
	v, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}

	if len(v) != 32 {
		t.Fatalf("address length mistmatch. got %d bytes but expected %d", len(v), 32)
	}
	a := new(chunk.Address)
	*a = make([]byte, 32)
	copy(*a, v)

	return a
}
