// Copyright 2017 The go-ethereum Authors
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
	"context"
	crand "crypto/rand"
	"fmt"
	"io"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/log"
)

const DefaultChunkCount = 2

var MaxExcessSize = DefaultChunkCount

func TestFakeHasher(t *testing.T) {
	sectionSize := 32
	sizes := []int{0, sectionSize - 1, sectionSize, sectionSize + 1, sectionSize * 4, sectionSize*4 + 1}
	bufSizes := []int{32, 7, sectionSize / 2, sectionSize, sectionSize + 1, sectionSize*4 + 1}
	for _, bsz := range bufSizes {
		for _, sz := range sizes {
			t.Run(fmt.Sprintf("fh-buffersize%d-bytesize%d", bsz, sz), func(t *testing.T) {
				fh := newFakeHasher(bsz, sectionSize, 2*sectionSize)
				s := NewSimpleSplitter(fh, bsz)
				buf := make([]byte, bsz)
				_, err := io.ReadFull(crand.Reader, buf)
				if err != nil {
					t.Fatal(err.Error())
				}
				r := bytes.NewReader(buf)
				_, err = s.ReadFrom(r)
				if err != nil {
					t.Fatal(err.Error())
				}
				h, err := s.Sum(context.TODO())
				if err != nil {
					t.Fatal(err.Error())
				}
				if !bytes.Equal(h, fh.output) {
					t.Fatalf("no match, daddyo, expected %x, got %x", fh.output, h)
				}
			})
		}

	}
}

type fakeHasher struct {
	output      []byte
	sectionSize int
	chunkSize   int
	count       int
	cap         int
	length      int64
	doneC       chan struct{}
}

func newFakeHasher(byteSize int, sectionSize int, chunkSize int) *fakeHasher {
	var count int
	if byteSize > 0 {
		count = ((byteSize - 1) / sectionSize) + 1
	}
	fh := &fakeHasher{
		sectionSize: sectionSize,
		output:      make([]byte, byteSize),
		cap:         count,
		chunkSize:   chunkSize,
		doneC:       make(chan struct{}, count),
	}
	log.Debug("fakehasher create", "cap", count)
	return fh
}

func (fh *fakeHasher) GetBuffer(p int64) ([]byte, error) {
	if fh.count < fh.cap {
		log.Debug("fakehasher cc", "cap", fh.cap, "count", fh.count)
		fh.doneC <- struct{}{}
	}
	fh.count++
	return make([]byte, fh.sectionSize), nil

}

func (fh *fakeHasher) ChunkSize() int {
	return fh.chunkSize
}

func (fh *fakeHasher) SetLength(c int64) {
	fh.length = c
}

func (fh *fakeHasher) Reset() { fh.output = nil; return }

func (fh *fakeHasher) WriteBuffer(offset int64, r io.Reader) (int, error) {
	return 0, nil
}

func (fh *fakeHasher) WriteSection(section int64, data []byte) int {
	log.Warn("wrigint to hasher", "src", section, "data", data)
	pos := section * int64(fh.sectionSize)
	copy(fh.output[pos:], data)
	fh.doneC <- struct{}{}
	return len(data)
}

func (fh *fakeHasher) Size() int {
	return 42
}

func (fh *fakeHasher) BlockSize() int {
	return fh.sectionSize
}

func (fh *fakeHasher) Sum(hash []byte, length int, meta []byte) []byte {
	for i := 0; i < fh.cap; i++ {

		log.Debug("sum", "count", fh.count, "length", length, "i", i)
		<-fh.doneC
	}
	return fh.output
}
