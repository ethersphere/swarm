package encrypt

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"testing"
	"time"

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
		eFunc := New(key, uint32(42))
		eFunc.Init(ctx, errFunc)
		eFunc.Link(cacheFunc)
		return eFunc
	}

	_, data := testutil.SerialData(chunkSize, 255, 0)
	h := hasher.New(sectionSize, branches, encryptFunc)
	h.Init(ctx, func(error) {})
	h.Link(refHashFunc)
	h.Write(0, data)
	ref := h.Sum(nil, 0, nil)

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
