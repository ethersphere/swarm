package file

import (
	"fmt"
	"strconv"
	"strings"
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

// BenchmarkHasher generates benchmarks that are comparable to the pyramid hasher
func BenchmarkHasher(b *testing.B) {
	for i := start; i < end; i++ {
		b.Run(fmt.Sprintf("%d/%d", i, dataLengths[i]), benchmarkHasher)
	}
}

func benchmarkHasher(b *testing.B) {
	params := strings.Split(b.Name(), "/")
	dataLengthParam, err := strconv.ParseInt(params[2], 10, 64)
	if err != nil {
		b.Fatal(err)
	}
	dataLength := int(dataLengthParam)

	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() bmt.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHash := bmt.New(poolSync)
	_, data := testutil.SerialData(dataLength, 255, 0)

	for j := 0; j < b.N; j++ {
		h := New(sectionSize, branches, dataHash, refHashFunc)
		for i := 0; i < dataLength; i += chunkSize {
			size := chunkSize
			if dataLength-i < chunkSize {
				size = dataLength - i
			}
			h.Write(data[i : i+size])
		}
		h.Sum(nil)
	}
}
