package hasher

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethersphere/swarm/file"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

// tests order-neutral concurrent writes with entire max size written in one go
func TestAsyncCorrectness(t *testing.T) {
	data := testutil.RandomBytes(1, BufferSize)
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
					pool := NewTreePool(hasher, count, capacity)
					defer pool.Drain(0)
					for n := 1; n <= max; n += incr {
						incr = 1 + rand.Intn(5)
						bmt := New(pool)
						d := data[:n]
						rbmt := NewRefHasher(hasher, count)
						expNoMeta := rbmt.Hash(d)
						h := hasher()
						h.Write(zeroSpan)
						h.Write(expNoMeta)
						exp := h.Sum(nil)
						got := syncHash(bmt, 0, d)
						if !bytes.Equal(got, exp) {
							t.Fatalf("wrong sync hash (syncpart) for datalength %v: expected %x (ref), got %x", n, exp, got)
						}
						sw := bmt.NewAsyncWriter(double)
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
	pool := NewTreePool(hasher, segmentCount, PoolSize)
	bmt := New(pool).NewAsyncWriter(double)
	idxs, segments := splitAndShuffle(bmt.SectionSize(), data)
	rand.Shuffle(len(idxs), func(i int, j int) {
		idxs[i], idxs[j] = idxs[j], idxs[i]
	})

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		asyncHash(bmt, 0, n, wh, idxs, segments)
	}
}

// splits the input data performs a random shuffle to mock async section writes
func asyncHashRandom(bmt file.SectionWriter, spanLength int, data []byte, wh whenHash) (s []byte) {
	idxs, segments := splitAndShuffle(bmt.SectionSize(), data)
	return asyncHash(bmt, spanLength, len(data), wh, idxs, segments)
}

// mock for async section writes for file.SectionWriter
// requires a permutation (a random shuffle) of list of all indexes of segments
// and writes them in order to the appropriate section
// the Sum function is called according to the wh parameter (first, last, random [relative to segment writes])
func asyncHash(bmt file.SectionWriter, spanLength int, l int, wh whenHash, idxs []int, segments [][]byte) (s []byte) {
	bmt.Reset()
	if l == 0 {
		bmt.SetLength(l)
		bmt.SetSpan(spanLength)
		return bmt.Sum(nil)
	}
	c := make(chan []byte, 1)
	hashf := func() {
		bmt.SetLength(l)
		bmt.SetSpan(spanLength)
		c <- bmt.Sum(nil)
	}
	maxsize := len(idxs)
	var r int
	if wh == random {
		r = rand.Intn(maxsize)
	}
	for i, idx := range idxs {
		bmt.SeekSection(idx)
		bmt.Write(segments[idx])
		if (wh == first || wh == random) && i == r {
			go hashf()
		}
	}
	if wh == last {
		bmt.SetLength(l)
		bmt.SetSpan(spanLength)
		return bmt.Sum(nil)
	}
	return <-c
}
