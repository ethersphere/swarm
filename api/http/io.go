// Copyright 2019 The Swarm Authors
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

package http

import (
	"bufio"
	"io"
	"net/http"
	"sync"

	"github.com/ethersphere/swarm/log"
)

// The size of buffer used for bufio.Reader on LazyChunkReader passed to
// http.ServeContent in HandleGetFile.
// Warning: This value influences the number of chunk requests and chunker join goroutines
// per file request.
// Recommended value is 4 times the io.Copy default buffer value which is 32kB.
var BufferSize = 4 * 32 * 1024
var Concurrency = 16

var PeekSize = 4 * 32 * 1024

// bufferedReadSeeker wraps bufio.Reader to expose Seek method
// from the provided io.ReadSeeker in newBufferedReadSeeker.
type bufferedReadSeeker struct {
	r *bufio.Reader
	s io.ReadSeeker
}

type logReadSeeker struct {
	io.ReadSeeker
}

func (b *logReadSeeker) Read(p []byte) (int, error) {
	// return b.s.Read(p)
	n, err := b.ReadSeeker.Read(p)
	log.Warn("logReadSeeker read", "len", len(p), "n", n, "err", err)
	return n, err
}

type bufferedReadSeekerPeeker struct {
	bufferedReadSeeker
	peek int
	errc chan error
}

const (
	Baseline int = iota
	Buffered
	Peeker
)

var Mode = Baseline

func newDownloader(s io.ReadSeeker) io.ReadSeeker {
	switch Mode {
	case Baseline:
		return s
	case Buffered:
		return newBufferedReadSeeker(s, BufferSize)
	case Peeker:
		r := s.(reader)

		return newBufferedReadSeeker(newReadPeeker(r, BufferSize, Concurrency), BufferSize)
		// return newBufferedReadSeekerPeeker(s, BufferSize, PeekSize)
	}
	return nil
}

// newBufferedReadSeeker creates a new instance of bufferedReadSeeker,
// out of io.ReadSeeker. Argument `size` is the size of the read buffer.
func newBufferedReadSeeker(readSeeker io.ReadSeeker, size int) bufferedReadSeeker {
	log.Warn("new ", "len", size)
	s := &logReadSeeker{readSeeker}
	return bufferedReadSeeker{
		// r: bufio.NewReader(s),
		r: bufio.NewReaderSize(s, size),
		s: s,
	}
}

// newBufferedReadSeekerPeeker creates a new instance of bufferedReadSeekerPeeker,
// out of io.ReadSeeker. Argument `size` is the size of the read buffer.
// Argument `peek` is the peekahead buffer
func newBufferedReadSeekerPeeker(readSeeker io.ReadSeeker, size, peek int) bufferedReadSeekerPeeker {
	log.Warn("new ", "len", size)
	b := bufferedReadSeekerPeeker{
		bufferedReadSeeker: newBufferedReadSeeker(readSeeker, size),
		peek:               peek,
		errc:               make(chan error, 1),
	}
	b.errc <- nil
	return b
}

func (b bufferedReadSeekerPeeker) goPeek() error {
	n := b.r.Size()
	if _, err := b.r.Peek(n); err != nil {
		log.Error("bufferedReadSeekerPeeker: peek", "size", b.r.Size(), "n", n, "err", err)
	}
	return nil
}

func (b bufferedReadSeekerPeeker) Read(p []byte) (int, error) {
	// err := <-b.errc
	// if err != nil {
	// 	return 0, fmt.Errorf("peekahead error: %v", err)
	// }
	// go func() {
	// 	// <-b.errc
	// 	b.errc <- b.goPeek()
	// }()
	peek := <-b.errc == nil
	n, err := b.r.Read(p)
	// if err == io.EOF {
	// 	err = nil
	// }
	// if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
	// 	return n, err
	// }
	log.Error("bufferedReadSeekerPeeker: read", "size", b.r.Size(), "len", len(p), "n", n, "err", err)
	// if err != nil {
	// 	return n, err
	// }
	if peek {
		go func() {
			b.errc <- b.goPeek()
		}()
	}
	return n, err
}

func (b bufferedReadSeeker) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)

	// if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
	// 	return n, err
	// }
	log.Warn("bufferedReadSeeker read", "size", b.r.Size(), "len", len(p), "buffered", b.r.Buffered(), "n", n, "err", err)
	// panic("oops")
	// if err == io.EOF {
	// 	err = nil
	// }
	return n, err
}

func (b bufferedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	log.Warn("bufferedReadSeeker seek", "offset", offset, "whence", whence)
	n, err := b.s.Seek(offset, whence)
	b.r.Reset(b.s)
	return n, err
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

type reader interface {
	io.ReadSeeker
	io.ReaderAt
}

type segment struct {
	idx     int // sequential index
	err     error
	segment []byte //
}

type readPeeker struct {
	reader      reader
	size        int64
	closed      chan struct{}
	wg          sync.WaitGroup
	cur         int
	segmentIdx  int
	segmentSize int
	buffer      chan *segment // peekcursor
	complete    chan *segment
	segments    map[int]*segment
	segmentPool sync.Pool
}

func newReadPeeker(r reader, segmentSize int, concurrency int) *readPeeker {
	closed := make(chan struct{})
	close(closed)
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		panic(err.Error())
	}
	// var size int64 = 131072
	log.Error("size", "size", size)
	return &readPeeker{
		reader:      r,
		size:        size,
		closed:      closed,
		buffer:      make(chan *segment, concurrency),
		complete:    make(chan *segment, concurrency),
		segments:    make(map[int]*segment),
		segmentSize: segmentSize,
		segmentPool: sync.Pool{
			New: func() interface{} { return &segment{segment: make([]byte, segmentSize)} },
		},
	}
}

func (rp *readPeeker) Seek(offset int64, whence int) (int64, error) {
	log.Warn("bufferedReadSeeker seek", "offset", offset, "whence", whence)
	rp.reset()
	n, err := rp.reader.Seek(offset, whence)
	rp.cur = int(offset / int64(rp.segmentSize))
	rp.segmentIdx = rp.cur
	return n, err
}

func (rp *readPeeker) reset() {
	select {
	case <-rp.closed:
		return
	default:
	}
	rp.wg.Wait()
	close(rp.closed)
	for range rp.buffer {
	}
	for range rp.complete {
	}

	rp.buffer = make(chan *segment, cap(rp.buffer))
	rp.complete = make(chan *segment, cap(rp.complete))
	rp.segmentIdx = rp.cur
	rp.segments = make(map[int]*segment)
}

func (d *readPeeker) peek() {
	d.closed = make(chan struct{})
	defer close(d.complete)
	defer close(d.buffer)
	// reading completed segments
	for i := 0; i < cap(d.complete); i++ {
		d.complete <- nil
	}
	for {
		select {
		case seg := <-d.complete:
			// keep cap(complete) concurrent fetches open
			next := d.segmentPool.Get().(*segment)
			next.idx = d.segmentIdx
			d.segmentIdx++
			if int64(d.segmentIdx)*int64(d.segmentSize) <= d.size {
				d.wg.Add(1)
				go d.readSegment(next)
			}
			if seg == nil {
				continue
			}
			// if this segment is the next segment to be read/buffered, just buffer
			// the longest possible sequence of segments
			if seg.idx == d.cur {
				for ok := true; ok; seg, ok = d.segments[d.cur] {
					d.buffer <- seg
					// select {
					// case d.buffer <- seg:
					// default:
					// 	// noone is reading, bother no further
					// 	log.Error("noone is reading, bother no further")
					// 	close(d.closed)
					// 	return
					// }
					delete(d.segments, d.cur)
					d.cur++
				}
			} else {
				d.segments[seg.idx] = seg
			}
		case <-d.closed:
			return
		}
	}
}

func (d *readPeeker) readSegment(seg *segment) {
	defer d.wg.Done()
	log.Warn("read segment", "idx", seg.idx)
	n, err := d.reader.ReadAt(seg.segment, int64(seg.idx)*int64(d.segmentSize))
	seg.segment = seg.segment[:n]
	seg.err = err
	d.complete <- seg
	log.Warn("segment complete", "idx", seg.idx, "n", n, "err", err)
}

func (d *readPeeker) Read(p []byte) (n int, err error) {
	if d.cur == 0 {
		adv := len(p) / d.segmentSize // >=1 since used within bufio.Reader with segment  size
		n, err := d.reader.Read(p[:adv*d.segmentSize])
		d.cur += n / d.segmentSize
		d.segmentIdx = d.cur
		return n, err
	}
	select {
	case <-d.closed:
		go d.peek()
	default:
	}
	var read int
	for read < len(p) {
		seg := <-d.buffer
		copy(p[read:], seg.segment)
		read += len(seg.segment)
		log.Warn("readPeeker read", "len", len(p), "read", read, "n", len(seg.segment), "err", seg.err)
		if seg.err != nil {
			log.Warn("readPeeker read complete", "err", err)
			break
		}
	}
	return read, err
}
