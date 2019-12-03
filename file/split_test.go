package file

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/file/hasher"
	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

const (
	sectionSize = 32
	branches    = 128
	chunkSize   = 4096
)

// TestSplit creates a Splitter with a reader with one chunk of serial data and
// a Hasher as the underlying param.SectionWriter
// It verifies the returned result
func TestSplit(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() *bmt.Hasher {
		return bmt.New(poolSync)
	}
	h := hasher.New(sectionSize, branches, dataHashFunc, refHashFunc)

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
