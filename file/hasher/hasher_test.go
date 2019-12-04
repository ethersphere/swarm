package hasher

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

// TestHasherJobTopHash verifies that the top hash on the first level is correctly set even though the Hasher writes asynchronously to the underlying job
func TestHasherJobTopHash(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	_, data := testutil.SerialData(chunkSize*branches, 255, 0)
	h := New(sectionSize, branches, dataHashFunc)
	h.Link(refHashFunc)
	var i int
	for i = 0; i < chunkSize*branches; i += chunkSize {
		h.Write(i, data[i:i+chunkSize])
	}
	h.Sum(nil, i, nil)
	levelOneTopHash := hexutil.Encode(h.index.GetTopHash(1))
	correctLevelOneTopHash := "0xc10090961e7682a10890c334d759a28426647141213abda93b096b892824d2ef"
	if levelOneTopHash != correctLevelOneTopHash {
		t.Fatalf("tophash; expected %s, got %s", correctLevelOneTopHash, levelOneTopHash)
	}

}

// TestHasherOneFullChunk verifies the result of writing a single data chunk to Hasher
func TestHasherOneFullChunk(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	_, data := testutil.SerialData(chunkSize*branches, 255, 0)
	h := New(sectionSize, branches, dataHashFunc)
	h.Link(refHashFunc)
	var i int
	for i = 0; i < chunkSize*branches; i += chunkSize {
		h.Write(i, data[i:i+chunkSize])
	}
	ref := h.Sum(nil, i, nil)
	correctRootHash := "0x3047d841077898c26bbe6be652a2ec590a5d9bd7cd45d290ea42511b48753c09"
	rootHash := hexutil.Encode(ref)
	if rootHash != correctRootHash {
		t.Fatalf("roothash; expected %s, got %s", correctRootHash, rootHash)
	}
}

// TestHasherOneFullChunk verifies that Hasher creates new jobs on branch thresholds
func TestHasherJobChange(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	_, data := testutil.SerialData(chunkSize*branches*branches, 255, 0)
	h := New(sectionSize, branches, dataHashFunc)
	h.Link(refHashFunc)
	jobs := make(map[string]int)
	for i := 0; i < chunkSize*branches*branches; i += chunkSize {
		h.Write(i, data[i:i+chunkSize])
		jobs[h.job.String()]++
	}
	i := 0
	for _, v := range jobs {
		if v != branches {
			t.Fatalf("jobwritecount writes: expected %d, got %d", branches, v)
		}
		i++
	}
	if i != branches {
		t.Fatalf("jobwritecount jobs: expected %d, got %d", branches, i)
	}
}

// TestHasherONeFullLevelOneChunk verifies the result of writing branches times data chunks to Hasher
func TestHasherOneFullLevelOneChunk(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	_, data := testutil.SerialData(chunkSize*branches*branches, 255, 0)
	h := New(sectionSize, branches, dataHashFunc)
	h.Link(refHashFunc)
	var i int
	for i = 0; i < chunkSize*branches*branches; i += chunkSize {
		h.Write(i, data[i:i+chunkSize])
	}
	ref := h.Sum(nil, i, nil)
	correctRootHash := "0x522194562123473dcfd7a457b18ee7dee8b7db70ed3cfa2b73f348a992fdfd3b"
	rootHash := hexutil.Encode(ref)
	if rootHash != correctRootHash {
		t.Fatalf("roothash; expected %s, got %s", correctRootHash, rootHash)
	}
}

func TestHasherVector(t *testing.T) {
	poolSync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return NewBMTSyncSectionWriter(bmt.New(poolSync))
	}

	var mismatch int
	for i, dataLength := range dataLengths {
		log.Info("hashervector start", "i", i, "l", dataLength)
		eq := true
		h := New(sectionSize, branches, dataHashFunc)
		h.Link(refHashFunc)
		_, data := testutil.SerialData(dataLength, 255, 0)
		for j := 0; j < dataLength; j += chunkSize {
			size := chunkSize
			if dataLength-j < chunkSize {
				size = dataLength - j
			}
			h.Write(j, data[j:j+size])
		}
		ref := h.Sum(nil, dataLength, nil)
		correctRefHex := "0x" + expected[i]
		refHex := hexutil.Encode(ref)
		if refHex != correctRefHex {
			mismatch++
			eq = false
		}
		t.Logf("[%7d+%4d]\t%v\tref: %x\texpect: %s", dataLength/chunkSize, dataLength%chunkSize, eq, ref, expected[i])
	}
	if mismatch > 0 {
		t.Fatalf("mismatches: %d/%d", mismatch, end-start)
	}
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
	refHashFunc := func() param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	dataHashFunc := func() param.SectionWriter {
		return NewBMTSyncSectionWriter(bmt.New(poolSync))
	}
	_, data := testutil.SerialData(dataLength, 255, 0)

	for j := 0; j < b.N; j++ {
		h := New(sectionSize, branches, dataHashFunc)
		h.Link(refHashFunc)
		for i := 0; i < dataLength; i += chunkSize {
			size := chunkSize
			if dataLength-i < chunkSize {
				size = dataLength - i
			}
			h.Write(i, data[i:i+size])
		}
		h.Sum(nil, dataLength, nil)
	}
}
