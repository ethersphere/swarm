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
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel   = flag.Int("loglevel", 3, "verbosity of logs")
	putTimeout = 30 * time.Second
	getTimeout = 30 * time.Second
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

type brokenLimitedReader struct {
	lr    io.Reader
	errAt int
	off   int
	size  int
}

func brokenLimitReader(data io.Reader, size int, errAt int) *brokenLimitedReader {
	return &brokenLimitedReader{
		lr:    data,
		errAt: errAt,
		size:  size,
	}
}

func newLDBStore(t *testing.T) (*LDBStore, func()) {
	dir, err := ioutil.TempDir("", "bzz-storage-test")
	if err != nil {
		t.Fatal(err)
	}
	log.Trace("memstore.tempdir", "dir", dir)

	ldbparams := NewLDBStoreParams(NewDefaultStoreParams(), dir)
	db, err := NewLDBStore(ldbparams)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		db.Close()
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, cleanup
}

func mputRandomChunks(store ChunkStore, n int, chunksize int64) ([]Address, error) {
	return mput(store, n, GenerateRandomChunk)
}

func mputChunks(store ChunkStore, chunks ...Chunk) error {
	i := 0
	f := func(n int64) Chunk {
		chunk := chunks[i]
		i++
		return chunk
	}
	_, err := mput(store, len(chunks), f)
	return err
}

func mput(store ChunkStore, n int, f func(i int64) Chunk) (hs []Address, err error) {
	// put to localstore and wait for stored channel
	// does not check delivery error state
	done := make(chan struct{})
	errc := make(chan error)
	ctx, _ := context.WithTimeout(context.Background(), putTimeout)
	// defer cancel()
	defer close(done)
	for i := int64(0); i < int64(n); i++ {
		chunk := f(DefaultChunkSize)
		wait, err := store.Put(chunk)
		if err != nil {
			return nil, err
		}
		go func() {
			select {
			case errc <- wait(ctx):
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

func mget(store ChunkStore, hs []Address, f func(h Address, chunk Chunk) error) error {
	wg := sync.WaitGroup{}
	wg.Add(len(hs))
	errc := make(chan error)

	for _, k := range hs {
		go func(h Address) {
			defer wg.Done()
			chunk, err := store.Get(h)
			if err != nil {
				errc <- err
				return
			}
			if f != nil {
				err = f(h, chunk)
				if err != nil {
					errc <- err
					return
				}
			}
		}(k)
	}
	go func() {
		wg.Wait()
		close(errc)
	}()
	var err error
	select {
	case err = <-errc:
	case <-time.NewTimer(5 * time.Second).C:
		err = fmt.Errorf("timed out after 5 seconds")
	}
	return err
}

func testDataReader(l int) (r io.Reader) {
	return io.LimitReader(rand.Reader, int64(l))
}

func (r *brokenLimitedReader) Read(buf []byte) (int, error) {
	if r.off+len(buf) > r.errAt {
		return 0, fmt.Errorf("Broken reader")
	}
	r.off += len(buf)
	return r.lr.Read(buf)
}

func testStoreRandom(m ChunkStore, n int, chunksize int64, t *testing.T) {
	hs, err := mputRandomChunks(m, n, chunksize)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = mget(m, hs, nil)
	if err != nil {
		t.Fatalf("testStore failed: %v", err)
	}
}

func testStoreCorrect(m ChunkStore, n int, chunksize int64, t *testing.T) {
	hs, err := mputRandomChunks(m, n, chunksize)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	f := func(h Address, chunk Chunk) error {
		if !bytes.Equal(h, chunk.Address()) {
			return fmt.Errorf("key does not match retrieved chunk Address")
		}
		hasher := MakeHashFunc(DefaultHash)()
		hasher.ResetWithLength(chunk.SpanBytes())
		hasher.Write(chunk.Payload())
		exp := hasher.Sum(nil)
		if !bytes.Equal(h, exp) {
			return fmt.Errorf("key is not hash of chunk data")
		}
		return nil
	}
	err = mget(m, hs, f)
	if err != nil {
		t.Fatalf("testStore failed: %v", err)
	}
}

func benchmarkStorePut(store ChunkStore, n int, chunksize int64, b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mputRandomChunks(store, n, chunksize)
	}
}

func benchmarkStoreGet(store ChunkStore, n int, chunksize int64, b *testing.B) {
	hs, err := mputRandomChunks(store, n, chunksize)
	if err != nil {
		b.Fatalf("expected no error, got %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mget(store, hs, nil)
		if err != nil {
			b.Fatalf("mget failed: %v", err)
		}
	}
}

// MapChunkStore is a very simple ChunkStore implementation to store chunks in a map in memory.
type MapChunkStore struct {
	chunks map[string]Chunk
	mu     sync.RWMutex
}

func NewMapChunkStore() *MapChunkStore {
	return &MapChunkStore{
		chunks: make(map[string]Chunk),
	}
}

func (m *MapChunkStore) Put(ch Chunk) (func(context.Context) error, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chunks[ch.Address().Hex()] = ch
	return func(context.Context) error { return nil }, nil
}

func (m *MapChunkStore) Get(ref Address) (Chunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	chunk := m.chunks[ref.Hex()]
	if chunk == nil {
		return nil, ErrChunkNotFound
	}
	return chunk, nil
}

func (m *MapChunkStore) Close() {
}

// fakeChunkStore doesn't store anything, just implements the ChunkStore interface
// It can be used to inject into a hasherStore if you don't want to actually store data just do the
// hashing
type fakeChunkStore struct {
}

// Put doesn't store anything it is just here to implement ChunkStore
func (f *fakeChunkStore) Put(ch Chunk) (func(context.Context) error, error) {
	return func(context.Context) error { return nil }, nil
}

// Gut doesn't store anything it is just here to implement ChunkStore
func (f *fakeChunkStore) Get(ref Address) (Chunk, error) {
	panic("FakeChunkStore doesn't support Get")
}

// Close doesn't store anything it is just here to implement ChunkStore
func (f *fakeChunkStore) Close() {
}

func NewRandomChunk(chunkSize uint64) Chunk {
	data := make([]byte, chunkSize+8) // SData should be chunkSize + 8 bytes reserved for length

	rand.Read(data)

	hasher := MakeHashFunc(SHA3Hash)()
	hasher.Write(data)
	return NewChunk(hasher.Sum(nil), data)
}

type fakeDPA struct {
	store ChunkStore
}

func (f *fakeDPA) Get(rctx context.Context, ref Address) (ch Chunk, err error) {
	return f.store.Get(ref)
}

func (f *fakeDPA) Put(ch Chunk) (waitToStore func(ctx context.Context) error, err error) {
	return f.store.Put(ch)
}

func (f *fakeDPA) Has(ref Address) (waitToStore func(context.Context) error, err error) {
	_, err = f.store.Get(ref)
	return func(context.Context) error { return nil }, err
}

func (f *fakeDPA) Close() {

}

type chunkMemStore struct {
	*MemStore
}

func (m *chunkMemStore) Put(c Chunk) (waitToStore func(ctx context.Context) error, err error) {
	m.MemStore.Put(c)
	return func(context.Context) error { return nil }, nil
}
