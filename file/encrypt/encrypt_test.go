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

	hashFunc := testutillocal.NewBMTHasherFunc(0)

	e, err := New(nil, 42, hashFunc)
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
	cacheFunc := func(_ context.Context) param.SectionWriter {
		return cache
	}
	e, err = New(testKey, 42, cacheFunc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(testKey, e.key) {
		t.Fatalf("key seed; expected %x, got %x", testKey, e.key)
	}
	e.SetWriter(cacheFunc)

	_, data := testutil.SerialData(chunkSize, 255, 0)
	e.Write(data) // 0
	e.SetLength(chunkSize)
	doubleRef := e.Sum(nil)
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

	hashFunc := testutillocal.NewBMTHasherFunc(128 * 128)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	errFunc := func(error) {}

	cache := testutillocal.NewCache()
	cache.Init(ctx, errFunc)
	cache.SetWriter(hashFunc)
	cacheFunc := func(_ context.Context) param.SectionWriter {
		return cache
	}

	encryptFunc := func(_ context.Context) param.SectionWriter {
		eFunc, err := New(testKey, uint32(42), cacheFunc)
		if err != nil {
			t.Fatal(err)
		}
		eFunc.SetWriter(cacheFunc)
		eFunc.Init(ctx, errFunc)
		return eFunc
	}

	_, data := testutil.SerialData(chunkSize, 255, 0)
	h := hasher.New(encryptFunc)
	h.Init(ctx, func(error) {})
	h.Write(data) //0
	doubleRef := h.Sum(nil)

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

	bmtTreePool := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	hc := bmt.New(bmtTreePool)
	hc.Reset()
	hc.SetLength(len(cipherText))
	hc.Write(cipherText)
	cipherRef := hc.Sum(nil)
	dataRef := doubleRef[encryption.KeyLength:]
	if !bytes.Equal(dataRef, cipherRef) {
		t.Fatalf("encrypt ref; expected %x, got %x", cipherRef, dataRef)
	}
}

func TestEncryptChunkWholeAndSections(t *testing.T) {
	hashFunc := testutillocal.NewBMTHasherFunc(128 * 128)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	errFunc := func(error) {}

	cache := testutillocal.NewCache()
	cache.Init(ctx, errFunc)
	cache.SetWriter(hashFunc)
	cacheFunc := func(_ context.Context) param.SectionWriter {
		return cache
	}

	e, err := New(testKey, uint32(42), cacheFunc)
	if err != nil {
		t.Fatal(err)
	}
	e.Init(ctx, errFunc)

	_, data := testutil.SerialData(chunkSize, 255, 0)
	e.Write(data) // 0
	e.SetLength(chunkSize)
	e.Sum(nil)

	cacheCopy := make([]byte, chunkSize)
	copy(cacheCopy, cache.Get(0))
	cache.Delete(0)

	e, err = New(testKey, uint32(42), cacheFunc)
	if err != nil {
		t.Fatal(err)
	}
	e.Init(ctx, errFunc)

	for i := 0; i < chunkSize; i += sectionSize {
		e.SeekSection(i / sectionSize)
		e.Write(data[i : i+sectionSize])
	}
	e.SetLength(chunkSize)
	e.Sum(nil)

	for i := 0; i < chunkSize; i += sectionSize {
		chunked := cacheCopy[i : i+sectionSize]
		sectioned := cache.Get(i / sectionSize)
		if !bytes.Equal(chunked, sectioned) {
			t.Fatalf("encrypt chunk full and section idx %d; expected %x, got %x", i/sectionSize, chunked, sectioned)
		}
	}
}

func TestEncryptIntermediateChunk(t *testing.T) {
	hashFunc := testutillocal.NewBMTHasherFunc(128 * 128)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	errFunc := func(err error) {
		log.Error("filehasher pipeline error", "err", err)
		cancel()
	}

	cache := testutillocal.NewCache()
	cache.Init(ctx, errFunc)
	cache.SetWriter(hashFunc)
	cacheFunc := func(_ context.Context) param.SectionWriter {
		return cache
	}

	encryptRefFunc := func(_ context.Context) param.SectionWriter {
		eFunc, err := New(testKey, uint32(42), cacheFunc)
		if err != nil {
			t.Fatal(err)
		}
		eFunc.Init(ctx, errFunc)
		return eFunc
	}

	h := hasher.New(encryptRefFunc)

	_, data := testutil.SerialData(chunkSize*branches, 255, 0)
	for i := 0; i < chunkSize*branches; i += chunkSize {
		h.SeekSection(i / chunkSize)
		h.Write(data[i : i+chunkSize])
	}
	h.SetLength(chunkSize * branches)
	ref := h.Sum(nil)
	select {
	case <-ctx.Done():
		t.Fatalf("ctx done: %v", ctx.Err())
	default:
	}
	t.Logf("%x", ref)
}
