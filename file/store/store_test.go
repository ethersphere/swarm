package store

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

const (
	sectionSize = 32
	branches    = 128
	chunkSize   = 4096
)

// wraps storage.FakeChunkStore to intercept incoming chunk
type testChunkStore struct {
	*storage.FakeChunkStore
	chunkC chan<- chunk.Chunk
}

func newTestChunkStore(chunkC chan<- chunk.Chunk) *testChunkStore {
	return &testChunkStore{
		FakeChunkStore: &storage.FakeChunkStore{},
		chunkC:         chunkC,
	}
}

// Put overrides storage.FakeChunkStore.Put
func (s *testChunkStore) Put(_ context.Context, _ chunk.ModePut, chs ...chunk.Chunk) ([]bool, error) {
	for _, ch := range chs {
		s.chunkC <- ch
	}
	return s.FakeChunkStore.Put(nil, 0, chs...)
}

// TestStoreWithHasher writes a single chunk and verifies the asynchronusly received chunk
// through the underlying chunk store
func TestStoreWithHasher(t *testing.T) {
	pool := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	hashFunc := func() param.SectionWriter {
		return bmt.New(pool).NewAsyncWriter(false)
	}

	// initialize chunk store with channel to intercept chunk
	chunkC := make(chan chunk.Chunk)
	store := newTestChunkStore(chunkC)

	// initialize FileStore
	h := New(store)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	h.Init(ctx, nil)
	h.Link(hashFunc)

	// Write data to Store
	_, data := testutil.SerialData(chunkSize, 255, 0)
	span := bmt.LengthToSpan(chunkSize)
	go func() {
		for i := 0; i < chunkSize; i += sectionSize {
			h.Write(i/sectionSize, data[i:i+sectionSize])
		}
		h.Sum(nil, chunkSize, span)
	}()

	// capture chunk and verify contents
	select {
	case ch := <-chunkC:
		if !bytes.Equal(ch.Data()[:8], span) {
			t.Fatalf("chunk span; expected %x, got %x", span, ch.Data()[:8])
		}
		if !bytes.Equal(ch.Data()[8:], data) {
			t.Fatalf("chunk data; expected %x, got %x", data, ch.Data()[8:])
		}
		refHex := ch.Address().Hex()
		correctRefHex := "c10090961e7682a10890c334d759a28426647141213abda93b096b892824d2ef"
		if refHex != correctRefHex {
			t.Fatalf("chunk ref; expected %s, got %s", correctRefHex, refHex)
		}

	case <-ctx.Done():
		t.Fatalf("timeout %v", ctx.Err())
	}
}
