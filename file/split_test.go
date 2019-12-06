package file

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/file/hasher"
	"github.com/ethersphere/swarm/file/store"
	"github.com/ethersphere/swarm/file/testutillocal"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

const (
	sectionSize = 32
	branches    = 128
	chunkSize   = 4096
)

func init() {
	testutil.Init()
}

var (
	errFunc = func(err error) {
		log.Error("split writer pipeline error", "err", err)
	}
)

// TestSplit creates a Splitter with a reader with one chunk of serial data and
// a Hasher as the underlying param.SectionWriter
// It verifies the returned result
func TestSplit(t *testing.T) {

	hashFunc := testutillocal.NewBMTHasherFunc(0)
	h := hasher.New(hashFunc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Init(ctx, errFunc)

	r, _ := testutil.SerialData(chunkSize, 255, 0)
	s := NewSplitter(r, h)
	ref, err := s.Split()
	if err != nil {
		t.Fatal(err)
	}
	refHex := hexutil.Encode(ref)
	correctRefHex := "0xc10090961e7682a10890c334d759a28426647141213abda93b096b892824d2ef"
	if refHex != correctRefHex {
		t.Fatalf("split, expected %s, got %s", correctRefHex, refHex)
	}
}

// TestSplitWithDataFileStore verifies chunk.Store sink result for data hashing
func TestSplitWithDataFileStore(t *testing.T) {
	hashFunc := testutillocal.NewBMTHasherFunc(128)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	chunkStore := &storage.FakeChunkStore{}
	storeFunc := func(_ context.Context) param.SectionWriter {
		h := store.New(chunkStore, hashFunc)
		h.Init(ctx, errFunc)
		return h
	}

	h := hasher.New(storeFunc)
	h.Init(ctx, errFunc)

	r, _ := testutil.SerialData(chunkSize, 255, 0)
	s := NewSplitter(r, h)
	ref, err := s.Split()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	refHex := hexutil.Encode(ref)
	correctRefHex := "0xc10090961e7682a10890c334d759a28426647141213abda93b096b892824d2ef"
	if refHex != correctRefHex {
		t.Fatalf("split, expected %s, got %s", correctRefHex, refHex)
	}
}
