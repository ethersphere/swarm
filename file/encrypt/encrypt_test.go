package encrypt

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/file/hasher"
	"github.com/ethersphere/swarm/file/testutillocal"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/storage/encryption"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

const (
	sectionSize = 32
	branches    = 128
	chunkSize   = 4096
)

var (
	testKey = append(make([]byte, encryption.KeyLength-1), byte(0x2a))
)

func init() {
	testutil.Init()
}

func TestKey(t *testing.T) {

	e, err := New(nil, 42)
	if err != nil {
		t.Fatal(err)
	}
	if e.key == nil {
		t.Fatalf("new key nil; expected not nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	errFunc := func(error) {}
	cache := testutillocal.NewCache()
	cache.Init(ctx, errFunc)
	cacheFunc := func() param.SectionWriter {
		return cache
	}
	e, err = New(testKey, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(testKey, e.key) {
		t.Fatalf("key seed; expected %x, got %x", testKey, e.key)
	}

	_, data := testutil.SerialData(chunkSize, 255, 0)
	e.Link(cacheFunc)
	e.Write(0, data)
	span := bmt.LengthToSpan(chunkSize)
	doubleRef := e.Sum(nil, chunkSize, span)
	refKey := doubleRef[:encryption.KeyLength]
	if !bytes.Equal(refKey, testKey) {
		t.Fatalf("returned ref key, expected %x, got %x", testKey, refKey)
	}

	correctNextKeyHex := "0xbeced09521047d05b8960b7e7bcc1d1292cf3e4b2a6b63f48335cbde5f7545d2"
	nextKeyHex := hexutil.Encode(e.key)
	if nextKeyHex != correctNextKeyHex {
		t.Fatalf("key next; expected %s, got %s", correctNextKeyHex, nextKeyHex)
	}
}

func TestEncryptOneChunk(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return hasher.NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	errFunc := func(error) {}

	cache := testutillocal.NewCache()
	cache.Init(ctx, errFunc)
	cache.Link(dataHashFunc)
	cacheFunc := func() param.SectionWriter {
		return cache
	}

	encryptFunc := func() param.SectionWriter {
		eFunc, err := New(testKey, uint32(42))
		if err != nil {
			t.Fatal(err)
		}
		eFunc.Init(ctx, errFunc)
		eFunc.Link(cacheFunc)
		return eFunc
	}

	_, data := testutil.SerialData(chunkSize, 255, 0)
	h := hasher.New(sectionSize, branches, encryptFunc)
	h.Init(ctx, func(error) {})
	h.Link(refHashFunc)
	h.Write(0, data)
	doubleRef := h.Sum(nil, 0, nil)

	enc := encryption.New(testKey, 0, 42, sha3.NewLegacyKeccak256)
	cipherText, err := enc.Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}
	cacheText := cache.Get(0)
	if !bytes.Equal(cipherText, cacheText) {
		log.Trace("data mismatch", "expect", cipherText, "got", cacheText)
		t.Fatalf("encrypt onechunk; data mismatch")
	}

	hc := bmt.New(poolSync)
	span := bmt.LengthToSpan(len(cipherText))
	hc.ResetWithLength(span)
	hc.Write(cipherText)
	cipherRef := hc.Sum(nil)
	dataRef := doubleRef[encryption.KeyLength:]
	if !bytes.Equal(dataRef, cipherRef) {
		t.Fatalf("encrypt ref; expected %x, got %x", cipherRef, dataRef)
	}
}

func TestEncryptChunkWholeAndSections(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return hasher.NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	errFunc := func(error) {}

	cache := testutillocal.NewCache()
	cache.Init(ctx, errFunc)
	cache.Link(dataHashFunc)
	cacheFunc := func() param.SectionWriter {
		return cache
	}

	e, err := New(testKey, uint32(42))
	if err != nil {
		t.Fatal(err)
	}
	e.Init(ctx, errFunc)
	e.Link(cacheFunc)

	_, data := testutil.SerialData(chunkSize, 255, 0)
	e.Write(0, data)
	span := bmt.LengthToSpan(chunkSize)
	e.Sum(nil, chunkSize, span)

	cacheCopy := make([]byte, chunkSize)
	copy(cacheCopy, cache.Get(0))
	cache.Delete(0)

	cache.Link(refHashFunc)
	e, err = New(testKey, uint32(42))
	if err != nil {
		t.Fatal(err)
	}
	e.Init(ctx, errFunc)
	e.Link(cacheFunc)

	for i := 0; i < chunkSize; i += sectionSize {
		e.Write(i/sectionSize, data[i:i+sectionSize])
	}
	e.Sum(nil, chunkSize, span)

	for i := 0; i < chunkSize; i += sectionSize {
		chunked := cacheCopy[i : i+sectionSize]
		sectioned := cache.Get(i / sectionSize)
		if !bytes.Equal(chunked, sectioned) {
			t.Fatalf("encrypt chunk full and section idx %d; expected %x, got %x", i/sectionSize, chunked, sectioned)
		}
	}
}

func TestEncryptIntermediateChunk(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return hasher.NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	errFunc := func(err error) {
		log.Error("filehasher pipeline error", "err", err)
		cancel()
	}

	cache := testutillocal.NewCache()
	cache.Init(ctx, errFunc)
	cache.Link(refHashFunc)
	cacheFunc := func() param.SectionWriter {
		return cache
	}

	encryptRefFunc := func() param.SectionWriter {
		eFunc, err := New(testKey, uint32(42))
		if err != nil {
			t.Fatal(err)
		}
		eFunc.Init(ctx, errFunc)
		eFunc.Link(cacheFunc)
		return eFunc
	}

	encryptDataFunc := func() param.SectionWriter {
		eFunc, err := New(nil, uint32(42))
		if err != nil {
			t.Fatal(err)
		}
		eFunc.Init(ctx, errFunc)
		eFunc.Link(dataHashFunc)
		return eFunc
	}

	h := hasher.New(sectionSize, branches, encryptDataFunc)
	h.Link(encryptRefFunc)

	_, data := testutil.SerialData(chunkSize*branches, 255, 0)
	for i := 0; i < chunkSize*branches; i += chunkSize {
		h.Write(i/chunkSize, data[i:i+chunkSize])
	}
	span := bmt.LengthToSpan(chunkSize * branches)
	ref := h.Sum(nil, chunkSize*branches, span)
	select {
	case <-ctx.Done():
		t.Fatalf("ctx done: %v", ctx.Err())
	default:
	}
	t.Logf("%x", ref)
}
