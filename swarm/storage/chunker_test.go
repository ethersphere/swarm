// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/sha3"
)

/*
Tests TreeChunker by splitting and joining a random byte slice
*/

type test interface {
	Fatalf(string, ...interface{})
	Logf(string, ...interface{})
}

type chunkerTester struct {
	inputs map[uint64][]byte
	t      test
}

type testChunkStore struct {
	chunks map[string]*Chunk
	mu     sync.RWMutex
}

func (t *testChunkStore) Put(chunk *Chunk) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.chunks[chunk.Key.Hex()] = chunk
}

func (t *testChunkStore) Get(key Key) (*Chunk, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.chunks[key.Hex()], nil
}

func (t *testChunkStore) Close() {
	return
}

func newTestChunkStore() *testChunkStore {
	return &testChunkStore{
		chunks: make(map[string]*Chunk),
	}
}

func newTestHasherStore() *hasherStore {
	return NewHasherStore(newTestChunkStore(), MakeHashFunc(BMTHash), false)
}

func testRandomBrokenData(splitter Splitter, n int, tester *chunkerTester) {
	data := io.LimitReader(rand.Reader, int64(n))
	brokendata := brokenLimitReader(data, n, n/2)

	buf := make([]byte, n)
	_, err := brokendata.Read(buf)
	if err == nil || err.Error() != "Broken reader" {
		tester.t.Fatalf("Broken reader is not broken, hence broken. Returns: %v", err)
	}

	data = io.LimitReader(rand.Reader, int64(n))
	brokendata = brokenLimitReader(data, n, n/2)

	putGetter := newTestHasherStore()

	expectedError := fmt.Errorf("Broken reader")
	key, _, err := splitter.Split(brokendata, int64(n), putGetter)
	if err == nil || err.Error() != expectedError.Error() {
		tester.t.Fatalf("Not receiving the correct error! Expected %v, received %v", expectedError, err)
	}
	tester.t.Logf(" Key = %v\n", key)
}

func testRandomData(chunker Chunker, n int, tester *chunkerTester) Key {
	if tester.inputs == nil {
		tester.inputs = make(map[uint64][]byte)
	}
	input, found := tester.inputs[uint64(n)]
	var data io.Reader
	if !found {
		data, input = generateRandomData(n)
		tester.inputs[uint64(n)] = input
	} else {
		data = io.LimitReader(bytes.NewReader(input), int64(n))
	}

	putGetter := newTestHasherStore()

	key, wait, err := chunker.Split(data, int64(n), putGetter)
	if err != nil {
		tester.t.Fatalf(err.Error())
	}
	tester.t.Logf(" Key = %v\n", key)
	wait()

	reader := chunker.Join(key, putGetter, 0)
	output := make([]byte, n)
	r, err := reader.Read(output)
	if r != n || err != io.EOF {
		tester.t.Fatalf("read error  read: %v  n = %v  err = %v\n", r, n, err)
	}
	if input != nil {
		if !bytes.Equal(output, input) {
			tester.t.Fatalf("input and output mismatch\n IN: %v\nOUT: %v\n", input, output)
		}
	}

	return key
}

func TestSha3ForCorrectness(t *testing.T) {
	tester := &chunkerTester{t: t}

	size := 4096
	input := make([]byte, size+8)
	binary.LittleEndian.PutUint64(input[:8], uint64(size))

	io.LimitReader(bytes.NewReader(input[8:]), int64(size))

	rawSha3 := sha3.NewKeccak256()
	rawSha3.Reset()
	rawSha3.Write(input)
	rawSha3Output := rawSha3.Sum(nil)

	sha3FromMakeFunc := MakeHashFunc(SHA3Hash)()
	sha3FromMakeFunc.ResetWithLength(input[:8])
	sha3FromMakeFunc.Write(input[8:])
	sha3FromMakeFuncOutput := sha3FromMakeFunc.Sum(nil)

	if len(rawSha3Output) != len(sha3FromMakeFuncOutput) {
		tester.t.Fatalf("Original SHA3 and abstracted Sha3 has different length %v:%v\n", len(rawSha3Output), len(sha3FromMakeFuncOutput))
	}

	if !bytes.Equal(rawSha3Output, sha3FromMakeFuncOutput) {
		tester.t.Fatalf("Original SHA3 and abstracted Sha3 mismatch %v:%v\n", rawSha3Output, sha3FromMakeFuncOutput)
	}

}

func XTestDataAppend(t *testing.T) {
	sizes := []int{1, 1, 1, 4095, 4096, 4097, 1, 1, 1, 123456, 2345678, 2345678}
	appendSizes := []int{4095, 4096, 4097, 1, 1, 1, 8191, 8192, 8193, 9000, 3000, 5000}

	tester := &chunkerTester{t: t}
	for i := range sizes {
		n := sizes[i]
		m := appendSizes[i]

		if tester.inputs == nil {
			tester.inputs = make(map[uint64][]byte)
		}
		input, found := tester.inputs[uint64(n)]
		var data io.Reader
		if !found {
			data, input = generateRandomData(n)
			tester.inputs[uint64(n)] = input
		} else {
			data = io.LimitReader(bytes.NewReader(input), int64(n))
		}

		chunker := NewPyramidChunker(NewChunkerParams())
		putGetter := newTestHasherStore()
		key, wait, err := chunker.Split(data, int64(n), putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
		wait()

		//create a append data stream
		appendInput, found := tester.inputs[uint64(m)]
		var appendData io.Reader
		if !found {
			appendData, appendInput = generateRandomData(m)
			tester.inputs[uint64(m)] = appendInput
		} else {
			appendData = io.LimitReader(bytes.NewReader(appendInput), int64(m))
		}

		newKey, wait, err := chunker.Append(key, appendData, putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
		wait()

		reader := chunker.Join(newKey, putGetter, 0)
		newOutput := make([]byte, n+m)
		r, err := reader.Read(newOutput)
		if r != (n + m) {
			tester.t.Fatalf("read error  read: %v  n = %v  m = %v  err = %v\n", r, n, m, err)
		}

		newInput := append(input, appendInput...)
		if !bytes.Equal(newOutput, newInput) {
			tester.t.Fatalf("input and output mismatch\n IN: %v\nOUT: %v\n", newInput, newOutput)
		}
	}
}

func TestRandomData(t *testing.T) {
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289, 123456, 2345678}
	tester := &chunkerTester{t: t}

	// TODO: only tree chunker is implemented well
	chunker := NewTreeChunker(NewChunkerParams())
	// pyramid := NewTreePyramidChunker(NewChunkerParams())
	for _, s := range sizes {
		testRandomData(chunker, s, tester)
		// testRandomData(pyramid, s, tester)
		// if treeChunkerKey.String() != pyramidChunkerKey.String() {
		// 	tester.t.Fatalf("tree chunker and pyramid chunker key mismatch for size %v\n TC: %v\n PC: %v\n", s, treeChunkerKey.String(), pyramidChunkerKey.String())
		// }
	}
}

func TestRandomBrokenData(t *testing.T) {
	sizes := []int{1, 60, 83, 179, 253, 1024, 4095, 4096, 4097, 8191, 8192, 8193, 12287, 12288, 12289, 123456, 2345678}
	tester := &chunkerTester{t: t}
	chunker := NewTreeChunker(NewChunkerParams())
	for _, s := range sizes {
		testRandomBrokenData(chunker, s, tester)
	}
}

func benchReadAll(reader LazySectionReader) {
	size, _ := reader.Size(nil)
	output := make([]byte, 1000)
	for pos := int64(0); pos < size; pos += 1000 {
		reader.ReadAt(output, pos)
	}
}

func benchmarkJoin(n int, t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		chunker := NewPyramidChunker(NewChunkerParams())
		tester := &chunkerTester{t: t}
		data := testDataReader(n)

		// chunkC := make(chan *Chunk, 1000)
		putGetter := newTestHasherStore()
		key, wait, err := chunker.Split(data, int64(n), putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
		wait()
		// chunkC = make(chan *Chunk, 1000)
		reader := chunker.Join(key, putGetter, i)
		benchReadAll(reader)
	}
}

func benchmarkSplitTreeSHA3(n int, t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		chunker := NewPyramidChunker(NewChunkerParams())
		tester := &chunkerTester{t: t}
		data := testDataReader(n)
		putGetter := newTestHasherStore()
		_, _, err := chunker.Split(data, int64(n), putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
	}
}

func benchmarkSplitTreeBMT(n int, t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		cp := NewChunkerParams()
		cp.Hash = BMTHash
		chunker := NewPyramidChunker(cp)
		tester := &chunkerTester{t: t}
		data := testDataReader(n)
		putGetter := newTestHasherStore()
		_, _, err := chunker.Split(data, int64(n), putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
	}
}

func benchmarkSplitPyramidSHA3(n int, t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		splitter := NewPyramidChunker(NewChunkerParams())
		tester := &chunkerTester{t: t}
		data := testDataReader(n)
		putGetter := newTestHasherStore()
		_, _, err := splitter.Split(data, int64(n), putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}

	}
}

func benchmarkSplitPyramidBMT(n int, t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		cp := NewChunkerParams()
		cp.Hash = BMTHash
		splitter := NewPyramidChunker(cp)
		tester := &chunkerTester{t: t}
		data := testDataReader(n)
		putGetter := newTestHasherStore()
		_, _, err := splitter.Split(data, int64(n), putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
	}
}

func benchmarkAppendPyramid(n, m int, t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		chunker := NewPyramidChunker(NewChunkerParams())
		tester := &chunkerTester{t: t}
		data := testDataReader(n)
		data1 := testDataReader(m)

		putGetter := newTestHasherStore()
		key, wait, err := chunker.Split(data, int64(n), putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
		wait()

		_, wait, err = chunker.Append(key, data1, putGetter)
		if err != nil {
			tester.t.Fatalf(err.Error())
		}
		wait()
	}
}

func BenchmarkJoin_2(t *testing.B) { benchmarkJoin(100, t) }
func BenchmarkJoin_3(t *testing.B) { benchmarkJoin(1000, t) }
func BenchmarkJoin_4(t *testing.B) { benchmarkJoin(10000, t) }
func BenchmarkJoin_5(t *testing.B) { benchmarkJoin(100000, t) }
func BenchmarkJoin_6(t *testing.B) { benchmarkJoin(1000000, t) }
func BenchmarkJoin_7(t *testing.B) { benchmarkJoin(10000000, t) }

// func BenchmarkJoin_8(t *testing.B) { benchmarkJoin(100000000, t) }

func BenchmarkSplitTreeSHA3_2(t *testing.B)  { benchmarkSplitTreeSHA3(100, t) }
func BenchmarkSplitTreeSHA3_2h(t *testing.B) { benchmarkSplitTreeSHA3(500, t) }
func BenchmarkSplitTreeSHA3_3(t *testing.B)  { benchmarkSplitTreeSHA3(1000, t) }
func BenchmarkSplitTreeSHA3_3h(t *testing.B) { benchmarkSplitTreeSHA3(5000, t) }
func BenchmarkSplitTreeSHA3_4(t *testing.B)  { benchmarkSplitTreeSHA3(10000, t) }
func BenchmarkSplitTreeSHA3_4h(t *testing.B) { benchmarkSplitTreeSHA3(50000, t) }
func BenchmarkSplitTreeSHA3_5(t *testing.B)  { benchmarkSplitTreeSHA3(100000, t) }
func BenchmarkSplitTreeSHA3_6(t *testing.B)  { benchmarkSplitTreeSHA3(1000000, t) }
func BenchmarkSplitTreeSHA3_7(t *testing.B)  { benchmarkSplitTreeSHA3(10000000, t) }

// func BenchmarkSplitTreeSHA3_8(t *testing.B)  { benchmarkSplitTreeSHA3(100000000, t) }

func BenchmarkSplitTreeBMT_2(t *testing.B)  { benchmarkSplitTreeBMT(100, t) }
func BenchmarkSplitTreeBMT_2h(t *testing.B) { benchmarkSplitTreeBMT(500, t) }
func BenchmarkSplitTreeBMT_3(t *testing.B)  { benchmarkSplitTreeBMT(1000, t) }
func BenchmarkSplitTreeBMT_3h(t *testing.B) { benchmarkSplitTreeBMT(5000, t) }
func BenchmarkSplitTreeBMT_4(t *testing.B)  { benchmarkSplitTreeBMT(10000, t) }
func BenchmarkSplitTreeBMT_4h(t *testing.B) { benchmarkSplitTreeBMT(50000, t) }
func BenchmarkSplitTreeBMT_5(t *testing.B)  { benchmarkSplitTreeBMT(100000, t) }
func BenchmarkSplitTreeBMT_6(t *testing.B)  { benchmarkSplitTreeBMT(1000000, t) }
func BenchmarkSplitTreeBMT_7(t *testing.B)  { benchmarkSplitTreeBMT(10000000, t) }

// func BenchmarkSplitTreeBMT_8(t *testing.B)  { benchmarkSplitTreeBMT(100000000, t) }

func BenchmarkSplitPyramidSHA3_2(t *testing.B)  { benchmarkSplitPyramidSHA3(100, t) }
func BenchmarkSplitPyramidSHA3_2h(t *testing.B) { benchmarkSplitPyramidSHA3(500, t) }
func BenchmarkSplitPyramidSHA3_3(t *testing.B)  { benchmarkSplitPyramidSHA3(1000, t) }
func BenchmarkSplitPyramidSHA3_3h(t *testing.B) { benchmarkSplitPyramidSHA3(5000, t) }
func BenchmarkSplitPyramidSHA3_4(t *testing.B)  { benchmarkSplitPyramidSHA3(10000, t) }
func BenchmarkSplitPyramidSHA3_4h(t *testing.B) { benchmarkSplitPyramidSHA3(50000, t) }
func BenchmarkSplitPyramidSHA3_5(t *testing.B)  { benchmarkSplitPyramidSHA3(100000, t) }
func BenchmarkSplitPyramidSHA3_6(t *testing.B)  { benchmarkSplitPyramidSHA3(1000000, t) }
func BenchmarkSplitPyramidSHA3_7(t *testing.B)  { benchmarkSplitPyramidSHA3(10000000, t) }

// func BenchmarkSplitPyramidSHA3_8(t *testing.B)  { benchmarkSplitPyramidSHA3(100000000, t) }

func BenchmarkSplitPyramidBMT_2(t *testing.B)  { benchmarkSplitPyramidBMT(100, t) }
func BenchmarkSplitPyramidBMT_2h(t *testing.B) { benchmarkSplitPyramidBMT(500, t) }
func BenchmarkSplitPyramidBMT_3(t *testing.B)  { benchmarkSplitPyramidBMT(1000, t) }
func BenchmarkSplitPyramidBMT_3h(t *testing.B) { benchmarkSplitPyramidBMT(5000, t) }
func BenchmarkSplitPyramidBMT_4(t *testing.B)  { benchmarkSplitPyramidBMT(10000, t) }
func BenchmarkSplitPyramidBMT_4h(t *testing.B) { benchmarkSplitPyramidBMT(50000, t) }
func BenchmarkSplitPyramidBMT_5(t *testing.B)  { benchmarkSplitPyramidBMT(100000, t) }
func BenchmarkSplitPyramidBMT_6(t *testing.B)  { benchmarkSplitPyramidBMT(1000000, t) }
func BenchmarkSplitPyramidBMT_7(t *testing.B)  { benchmarkSplitPyramidBMT(10000000, t) }

// func BenchmarkSplitPyramidBMT_8(t *testing.B)  { benchmarkSplitPyramidBMT(100000000, t) }

func BenchmarkAppendPyramid_2(t *testing.B)  { benchmarkAppendPyramid(100, 1000, t) }
func BenchmarkAppendPyramid_2h(t *testing.B) { benchmarkAppendPyramid(500, 1000, t) }
func BenchmarkAppendPyramid_3(t *testing.B)  { benchmarkAppendPyramid(1000, 1000, t) }
func BenchmarkAppendPyramid_4(t *testing.B)  { benchmarkAppendPyramid(10000, 1000, t) }
func BenchmarkAppendPyramid_4h(t *testing.B) { benchmarkAppendPyramid(50000, 1000, t) }
func BenchmarkAppendPyramid_5(t *testing.B)  { benchmarkAppendPyramid(1000000, 1000, t) }
func BenchmarkAppendPyramid_6(t *testing.B)  { benchmarkAppendPyramid(1000000, 1000, t) }
func BenchmarkAppendPyramid_7(t *testing.B)  { benchmarkAppendPyramid(10000000, 1000, t) }

// func BenchmarkAppendPyramid_8(t *testing.B)  { benchmarkAppendPyramid(100000000, 1000, t) }

// go test -timeout 20m -cpu 4 -bench=./swarm/storage -run no
// If you dont add the timeout argument above .. the benchmark will timeout and dump
