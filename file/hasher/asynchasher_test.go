package hasher

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

// tests order-neutral concurrent writes with entire max size written in one go
func TestAsyncCorrectness(t *testing.T) {
	data := testutil.RandomBytes(1, bufferSize)
	hasher := sha3.NewLegacyKeccak256
	size := hasher().Size()
	whs := []whenHash{first, last, random}

	for _, double := range []bool{false, true} {
		for _, wh := range whs {
			for _, count := range counts {
				t.Run(fmt.Sprintf("double_%v_hash_when_%v_segments_%v", double, wh, count), func(t *testing.T) {
					max := count * size
					var incr int
					capacity := 1
					pool := bmt.NewTreePool(hasher, count, capacity)
					defer pool.Drain(0)
					for n := 1; n <= max; n += incr {
						incr = 1 + rand.Intn(5)
						bmtobj := bmt.New(pool)
						d := data[:n]
						rbmtobj := bmt.NewRefHasher(hasher, count)
						expNoMeta := rbmtobj.Hash(d)
						h := hasher()
						h.Write(bmt.ZeroSpan)
						h.Write(expNoMeta)
						exp := h.Sum(nil)
						got := syncHash(bmtobj, 0, d)
						if !bytes.Equal(got, exp) {
							t.Fatalf("wrong sync hash (syncpart) for datalength %v: expected %x (ref), got %x", n, exp, got)
						}
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						sw := NewAsyncHasher(ctx, bmtobj, double, nil)
						got = asyncHashRandom(sw, 0, d, wh)
						if !bytes.Equal(got, exp) {
							t.Fatalf("wrong async hash (asyncpart) for datalength %v: expected %x, got %x", n, exp, got)
						}
					}
				})
			}
		}
	}
}

func BenchmarkBMTAsync(t *testing.B) {
	whs := []whenHash{first, last, random}
	for size := 4096; size >= 128; size /= 2 {
		for _, wh := range whs {
			for _, double := range []bool{false, true} {
				t.Run(fmt.Sprintf("double_%v_hash_when_%v_size_%v", double, wh, size), func(t *testing.B) {
					benchmarkBMTAsync(t, size, wh, double)
				})
			}
		}
	}
}

// benchmarks BMT hasher with asynchronous concurrent segment/section writes
func benchmarkBMTAsync(t *testing.B, n int, wh whenHash, double bool) {
	data := testutil.RandomBytes(1, n)
	hasher := sha3.NewLegacyKeccak256
	pool := bmt.NewTreePool(hasher, segmentCount, bmt.PoolSize)
	bmth := bmt.New(pool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bmtobj := NewAsyncHasher(ctx, bmth, double, nil)
	idxs, segments := splitAndShuffle(bmtobj.SectionSize(), data)
	rand.Shuffle(len(idxs), func(i int, j int) {
		idxs[i], idxs[j] = idxs[j], idxs[i]
	})

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		asyncHash(bmtobj, 0, n, wh, idxs, segments)
	}
}

// splits the input data performs a random shuffle to mock async section writes
func asyncHashRandom(bmtobj *AsyncHasher, spanLength int, data []byte, wh whenHash) (s []byte) {
	idxs, segments := splitAndShuffle(bmtobj.SectionSize(), data)
	return asyncHash(bmtobj, spanLength, len(data), wh, idxs, segments)
}

// mock for async section writes for file.SectionWriter
// requires a permutation (a random shuffle) of list of all indexes of segments
// and writes them in order to the appropriate section
// the Sum function is called according to the wh parameter (first, last, random [relative to segment writes])
func asyncHash(bmtobj *AsyncHasher, spanLength int, l int, wh whenHash, idxs []int, segments [][]byte) (s []byte) {
	bmtobj.Reset()
	if l == 0 {
		bmtobj.SetLength(l)
		bmtobj.SetSpan(spanLength)
		return bmtobj.SumIndexed(nil)
	}
	c := make(chan []byte, 1)
	hashf := func() {
		bmtobj.SetLength(l)
		bmtobj.SetSpan(spanLength)
		c <- bmtobj.SumIndexed(nil)
	}
	maxsize := len(idxs)
	var r int
	if wh == random {
		r = rand.Intn(maxsize)
	}
	for i, idx := range idxs {
		bmtobj.WriteIndexed(idx, segments[idx])
		if (wh == first || wh == random) && i == r {
			go hashf()
		}
	}
	if wh == last {
		bmtobj.SetLength(l)
		bmtobj.SetSpan(spanLength)
		return bmtobj.SumIndexed(nil)
	}
	return <-c
}

// TestUseAsyncAsOrdinaryHasher verifies that the bmt.Hasher can be used with the hash.Hash interface
func TestUseAsyncAsOrdinaryHasher(t *testing.T) {
	hasher := sha3.NewLegacyKeccak256
	pool := bmt.NewTreePool(hasher, segmentCount, bmt.PoolSize)
	sbmt := bmt.New(pool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	abmt := NewAsyncHasher(ctx, sbmt, false, nil)
	abmt.SetSpan(3)
	abmt.Write([]byte("foo"))
	res := abmt.Sum(nil)
	refh := bmt.NewRefHasher(hasher, 128)
	resh := refh.Hash([]byte("foo"))
	hsub := hasher()
	span := bmt.LengthToSpan(3)
	hsub.Write(span)
	hsub.Write(resh)
	refRes := hsub.Sum(nil)
	if !bytes.Equal(res, refRes) {
		t.Fatalf("normalhash; expected %x, got %x", refRes, res)
	}
}

// COPIED FROM bmt test package
// MERGE LATER

// Hash hashes the data and the span using the bmt hasher
func syncHash(h *bmt.Hasher, spanLength int, data []byte) []byte {
	h.Reset()
	h.SetSpan(spanLength)
	h.Write(data)
	return h.Sum(nil)
}

func splitAndShuffle(secsize int, data []byte) (idxs []int, segments [][]byte) {
	l := len(data)
	n := l / secsize
	if l%secsize > 0 {
		n++
	}
	for i := 0; i < n; i++ {
		idxs = append(idxs, i)
		end := (i + 1) * secsize
		if end > l {
			end = l
		}
		section := data[i*secsize : end]
		segments = append(segments, section)
	}
	rand.Shuffle(n, func(i int, j int) {
		idxs[i], idxs[j] = idxs[j], idxs[i]
	})
	return idxs, segments
}

const (
	// segmentCount is the maximum number of segments of the underlying chunk
	// Should be equal to max-chunk-data-size / hash-size
	// Currently set to 128 == 4096 (default chunk size) / 32 (sha3.keccak256 size)
	segmentCount = 128
)

const bufferSize = 4128

type whenHash = int

const (
	first whenHash = iota
	last
	random
)

var counts = []int{1, 2, 3, 4, 5, 8, 9, 15, 16, 17, 32, 37, 42, 53, 63, 64, 65, 111, 127, 128}
