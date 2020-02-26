// Copyright 2020 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package bmt

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethersphere/swarm/bmt"
	bmttestutil "github.com/ethersphere/swarm/bmt/testutil"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

type whenHash = int

const (
	first whenHash = iota
	last
	random
)

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

// tests order-neutral concurrent writes with entire max size written in one go
func TestAsyncCorrectness(t *testing.T) {
	data := testutil.RandomBytes(1, bmttestutil.BufferSize)
	hasher := sha3.NewLegacyKeccak256
	size := hasher().Size()
	whs := []whenHash{first, last, random}

	for _, double := range []bool{false, true} {
		for _, wh := range whs {
			for _, count := range bmttestutil.Counts {
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
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						sw := NewAsyncHasher(ctx, bmtobj, double, nil)
						got := asyncHashRandom(sw, 0, d, wh)
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
		bmtobj.SetSpan(spanLength)
		return bmtobj.SumIndexed(nil, l)
	}
	c := make(chan []byte, 1)
	hashf := func() {
		bmtobj.SetSpan(spanLength)
		c <- bmtobj.SumIndexed(nil, l)
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
		bmtobj.SetSpan(spanLength)
		return bmtobj.SumIndexed(nil, l)
	}
	return <-c
}

// benchmarks BMT hasher with asynchronous concurrent segment/section writes
func benchmarkBMTAsync(t *testing.B, n int, wh whenHash, double bool) {
	data := testutil.RandomBytes(1, n)
	hasher := sha3.NewLegacyKeccak256
	pool := bmt.NewTreePool(hasher, bmttestutil.SegmentCount, bmt.PoolSize)
	bmth := bmt.New(pool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bmtobj := NewAsyncHasher(ctx, bmth, double, nil)
	idxs, segments := splitAndShuffle(bmtobj.SectionSize(), data)
	rand.Shuffle(len(idxs), func(i int, j int) {
		idxs[i], idxs[j] = idxs[j], idxs[i]
	})

	var r []byte
	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		r = asyncHash(bmtobj, 0, n, wh, idxs, segments)
	}
	bmttestutil.BenchmarkBMTResult = r
}

// TestUseAsyncAsOrdinaryHasher verifies that the bmt.Hasher can be used with the hash.Hash interface
func TestUseAsyncAsOrdinaryHasher(t *testing.T) {
	hasher := sha3.NewLegacyKeccak256
	pool := bmt.NewTreePool(hasher, bmttestutil.SegmentCount, bmt.PoolSize)
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
