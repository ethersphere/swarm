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

// SimpleSplitter implements the io.ReaderFrom interface for synchronous read from data
// as data is written to it, it chops the input stream to section size buffers
// and calls the section write on the SectionHasher
type SimpleSplitter struct {
	hasher  Hash
	bufsize int
	result  chan []byte
}

func (s *SimpleSplitter) Hash(ctx context.Context, r io.Reader) ([]byte, error) {
	errc := make(chan error)
	go func() {
		select {
		case errc <- s.ReadFrom(r):
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}()

}

//
func NewSimpleSplitter(h Hash, bufsize int) *SimpleSplitter {
	return &SimpleSplitter{
		hasher:  h,
		bufsize: bufsize,
		result:  make(chan []byte),
	}
}

//
func (s *SimpleSplitter) ReadFrom(r io.Reader) error {
	var read int64
	buf := make([]byte, s.bufsize)
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		s.hasher.Write(buf[:n])
		read += int64(n)
		if err == io.EOF {
			go func() {
				s.result <- s.hasher.Sum(read)
			}()
			return nil
		}
	}
}

func (s *SimpleSplitter) Sum(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case sum := <-s.result:
		return sum, nil
	}
}
