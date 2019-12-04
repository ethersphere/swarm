package encrypt

import (
	"bytes"
	"context"
	crand "crypto/rand"
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
	key := [32]byte{}
	key[0] = 0x2a
	e, err = New(key[:], 42)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(key[:], e.key) {
		t.Fatalf("key seed; expected %x, got %x", key, e.key)
	}

	_, data := testutil.SerialData(chunkSize, 255, 0)
	e.Link(cacheFunc)
	e.Write(0, data)
	span := bmt.LengthToSpan(chunkSize)
	doubleRef := e.Sum(nil, chunkSize, span)
	refKey := doubleRef[:encryption.KeyLength]
	if !bytes.Equal(refKey, key[:]) {
		t.Fatalf("returned ref key, expected %x, got %x", key, refKey)
	}

	correctNextKeyHex := "0xd83b8137defe4bdaf5e1243b3175dc49b0a19c9d1f68044b7bf261db9f006233"
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

	key := make([]byte, encryption.KeyLength)
	c, err := crand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	if c != encryption.KeyLength {
		t.Fatalf("short read %d", c)
	}
	encryptFunc := func() param.SectionWriter {
		eFunc, err := New(key, uint32(42))
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
	h.Sum(nil, 0, nil)

	enc := encryption.New(key, 0, 42, sha3.NewLegacyKeccak256)
	cipherText, err := enc.Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}
	cacheText := cache.Get(0)
	if !bytes.Equal(cipherText, cacheText) {
		log.Trace("data mismatch", "expect", cipherText, "got", cacheText)
		t.Fatalf("encrypt onechunk; data mismatch")
	}
}

//func TestEncryptIntermediateChunk(t *testing.T) {
//	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
//	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
//	refHashFunc := func() param.SectionWriter {
//		return bmt.New(poolAsync).NewAsyncWriter(false)
//	}
//	dataHashFunc := func() param.SectionWriter {
//		return hasher.NewBMTSyncSectionWriter(bmt.New(poolSync))
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
//	defer cancel()
//	errFunc := func(error) {}
//
//	cache := testutillocal.NewCache()
//	cache.Init(ctx, errFunc)
//	cache.Link(dataHashFunc)
//	cacheFunc := func() param.SectionWriter {
//		return cache
//	}
//
//	key := make([]byte, encryption.KeyLength)
//	c, err := crand.Read(key)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if c != encryption.KeyLength {
//		t.Fatalf("short read %d", c)
//	}
//	encryptFunc := func() param.SectionWriter {
//		eFunc := New(key, uint32(42))
//		eFunc.Init(ctx, errFunc)
//		eFunc.Link(cacheFunc)
//		return eFunc
//	}
//}
