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
	"context"
	"io"
)

const DefaultChunkCount = 2
var MaxExcessSize = DefaultChunkCount

func TestAsyncWriteFromReaderCorrectness(t *testing.T) {
  data := make([]byte, DefaultChunkSize*DefaultChunkCount+rand.Intn(MaxExcessSize))
  reader := bytes.NewReader(b)
  fh := &fakeHasher{}
  splitter := NewSimpleSplitter(fh, bufsize)

  n, err := io.Copy(splitter, reader)
  if err != nil {
    if err == io.EOF {
      got = <-fh.result
  }

}

type fakeBaseHasherJoiner struct {
  input []byte
}

func (fh *fakeBaseHasherJoiner) Reset() { fh.input = nil; return}
func (fh *fakeBaseHasherJoiner) Write(b []byte) { fh.input = append(fh.input, b...) }
func (fh *fakeBaseHasherJoiner) Sum([]byte) []byte { return fh.input }
func (fh *fakeBaseHasherJoiner) BlockSize() int { return 64 }
func (fh *fakeBaseHasherJoiner) Size() int { return 32 }

type fakeHasher struct {
  input []byte
  output []byte
}

func newFakeHasher() *fakeHasher {
  return &fakeHasher{}
}

func (fh *fakeHasher) Reset() { fh.input = nil; return}
func (fh *fakeHasher) Write([]byte)
