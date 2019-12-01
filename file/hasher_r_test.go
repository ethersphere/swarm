package file

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

// TestManualDanglingChunk is a test script explicitly hashing and writing every individual level in the dangling chunk edge case
// we use a balanced tree with data size of chunkSize*branches, and a single chunk of data
// this case is chosen because it produces the wrong result in the pyramid hasher at the time of writing (master commit hash 4928d989ebd0854d993c10c194e61a5a5455e4f9)
func TestManualDanglingChunk(t *testing.T) {
	pool := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	h := bmt.New(pool)

	// to execute the job we need buffers with the following capacities:
	// level 0: chunkSize*branches+chunkSize
	// level 1: chunkSize
	// level 2: sectionSize * 2
	var levels [][]byte
	levels = append(levels, nil)
	levels = append(levels, make([]byte, chunkSize))
	levels = append(levels, make([]byte, sectionSize*2))

	// hash the balanced tree portion of the data level and write to level 1
	_, levels[0] = testutil.SerialData(chunkSize*branches+chunkSize, 255, 0)
	span := lengthToSpan(chunkSize)
	for i := 0; i < chunkSize*branches; i += chunkSize {
		h.ResetWithLength(span)
		h.Write(levels[0][i : i+chunkSize])
		copy(levels[1][i/branches:], h.Sum(nil))
	}
	refHex := hexutil.Encode(levels[1][:sectionSize])
	correctRefHex := "0xc10090961e7682a10890c334d759a28426647141213abda93b096b892824d2ef"
	if refHex != correctRefHex {
		t.Fatalf("manual dangling single chunk; expected %s, got %s", correctRefHex, refHex)
	}

	// write a single section of the dangling chunk
	// hash it and write the reference on the second section of level 3
	span = lengthToSpan(chunkSize)
	h.ResetWithLength(span)
	h.Write(levels[0][chunkSize*branches:])
	copy(levels[2][sectionSize:], h.Sum(nil))
	refHex = hexutil.Encode(levels[2][sectionSize:])
	correctRefHex = "0x81b31d9a7f6c377523e8769db021091df23edd9fd7bd6bcdf11a22f518db6006"
	if refHex != correctRefHex {
		t.Fatalf("manual dangling single chunk; expected %s, got %s", correctRefHex, refHex)
	}

	// hash the chunk on level 2 and write into the first section of level 3
	span = lengthToSpan(chunkSize * branches)
	h.ResetWithLength(span)
	h.Write(levels[1])
	copy(levels[2], h.Sum(nil))
	refHex = hexutil.Encode(levels[2][:sectionSize])
	correctRefHex = "0x3047d841077898c26bbe6be652a2ec590a5d9bd7cd45d290ea42511b48753c09"
	if refHex != correctRefHex {
		t.Fatalf("manual dangling balanced tree; expected %s, got %s", correctRefHex, refHex)
	}

	// hash the two sections on level 3 to obtain the root hash
	span = lengthToSpan(chunkSize*branches + chunkSize)
	h.ResetWithLength(span)
	h.Write(levels[2])
	ref := h.Sum(nil)
	refHex = hexutil.Encode(ref)
	correctRefHex = "0xb8e1804e37a064d28d161ab5f256cc482b1423d5cd0a6b30fde7b0f51ece9199"
	if refHex != correctRefHex {
		t.Fatalf("manual dangling root; expected %s, got %s", correctRefHex, refHex)
	}
}

// TestReferenceFileHasher executes the file hasher algorithms on serial input data of periods of 0-254
// of lengths defined in common_test.go
//
// the "expected" array in common_test.go is generated by this implementation, and test failure due to
// result mismatch is nothing else than an indication that something has changed in the reference filehasher
// or the underlying hashing algorithm
func TestReferenceFileHasherVector(t *testing.T) {
	pool := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	h := bmt.New(pool)
	var mismatch int
	for i := start; i < end; i++ {
		dataLength := dataLengths[i]
		log.Info("start", "i", i, "len", dataLength)
		fh := NewReferenceFileHasher(h, branches)
		r, data := testutil.SerialData(dataLength, 255, 0)
		refHash := fh.Hash(r, len(data))
		eq := true
		if expected[i] != fmt.Sprintf("%x", refHash) {
			mismatch++
			eq = false
		}
		t.Logf("[%7d+%4d]\t%v\tref: %x\texpect: %s", dataLength/chunkSize, dataLength%chunkSize, eq, refHash, expected[i])
	}
	if mismatch > 0 {
		t.Fatalf("mismatches: %d/%d", mismatch, end-start)
	}
}

// BenchmarkReferenceHasher establishes a baseline for a fully synchronous file hashing operation
// it will be vastly inefficient
func BenchmarkReferenceHasher(b *testing.B) {
	for i := start; i < end; i++ {
		b.Run(fmt.Sprintf("%d", dataLengths[i]), benchmarkReferenceFileHasher)
	}
}

func benchmarkReferenceFileHasher(b *testing.B) {
	params := strings.Split(b.Name(), "/")
	dataLength, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		b.Fatal(err)
	}
	pool := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, bmt.PoolSize)
	log.Trace("running reference bench", "l", dataLength)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, data := testutil.SerialData(int(dataLength), 255, 0)
		h := bmt.New(pool)
		fh := NewReferenceFileHasher(h, branches)
		fh.Hash(r, len(data))
	}
}
