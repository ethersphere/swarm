package file

import (
	"testing"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

func TestHasherOneFullChunk(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	dataHash := bmt.New(poolSync)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() bmt.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}

	_, data := testutil.SerialData(chunkSize*branches, 255, 0)
	h := New(sectionSize, branches, dataHash, refHashFunc)
	for i := 0; i < chunkSize*branches; i += chunkSize {
		h.Write(data[i : i+chunkSize])
	}
	ref := h.Sum(nil)
	t.Logf("res: %x", ref)
}
