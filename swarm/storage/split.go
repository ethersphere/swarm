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
	"encoding/binary"
	"io"

	"github.com/ethereum/go-ethereum/swarm/log"
)

// SimpleSplitter implements the io.ReaderFrom interface for synchronous read from data
// as data is written to it, it chops the input stream to section size buffers
// and calls the section write on the SectionHasher
type SimpleSplitter struct {
	hasher       SectionHasher
	sectionCount int
	count        int64
	result       chan []byte
	readBuffer   []byte
	writeBuffer  []byte
}

//
func NewSimpleSplitter(h SectionHasher, bufferSize int) *SimpleSplitter {
	writeBufferBytes := make([]byte, 0, h.BlockSize())
	return &SimpleSplitter{
		hasher:      h,
		result:      make(chan []byte),
		readBuffer:  make([]byte, bufferSize),
		writeBuffer: bytes.NewBuffer(writeBufferBytes),
	}
}

func (s *SimpleSplitter) Write(buf []byte) (int, error) {
	for len(buf) > 0 {
		sectionOffset := s.section - s.hasher.BlockSize()
		writeBuffer := s.hasher.getBuffer(s.count)
		c := len(buf)
		if c > len(s.hasher.BlockSize()) {
			c = len(s.hasher.BlockSize())
		}
		s.hasher.Write(s.sectionCount, s.writeBuffer.Bytes())
		s.count += c
		s.sectionCount++
		log.Debug("writer", "c", c)
		buf = buf[c:]
		s.sectionCount++
	}
	return int(s.count), nil
}

func (s *SimpleSplitter) Close() error {
	if s.writeBuffer.Len() > 0 {
		log.Debug("writer flush on close", "c", s.writeBuffer.Len())
		s.hasher.Write(s.sectionCount, s.writeBuffer.Bytes())
	}
	s.count = 0
	return nil
}

func (s *SimpleSplitter) ReadFrom(r io.Reader) (int64, error) {
	//lastChunkIndex := -1
	var buf []byte
	for {
		//chunkIndex := (s.count - 1) / s.hasher.ChunkSize()
		//if lastChunkIndex != chunkIndex {
		buf = s.hasher.getBuffer(s.count)
		//}
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return s.count, err
		}
		//s.Write(s.readBuffer[:n])
		s.count += n
		s.sectionCount++
		log.Debug("readfrom", "c", n)
		if err == io.EOF {
			s.Close()
			go func() {
				meta := make([]byte, 8)
				binary.BigEndian.PutUint64(meta, uint64(s.count))
				s.result <- s.hasher.Sum(nil, int(s.count), meta)
			}()
			return s.count, nil
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
