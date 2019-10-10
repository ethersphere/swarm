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

package langos

import (
	"io"
	"sync"

	"github.com/ethersphere/swarm/log"
)

type Reader interface {
	io.ReadSeeker
	io.ReaderAt
}

// Langos is a reader with a lookahead peekBuffer
// this is the most naive implementation of a lookahead peekBuffer
// it should issue a lookahead Read when a Read is called, hence
// the name - langos
// |--->====>>------------|
//    cur   topmost
// the first read is not a lookahead but the rest are
// so, it could be that a lookahead read might need to wait for a previous read to finish
// due to resource pooling
//
// All Read and Seek method call must be synchronous.
type Langos struct {
	r            Reader        // reader needs to implement io.ReadSeeker and io.ReaderAt interfaces
	cursor       int64         // current read position
	cursorMu     sync.Mutex    // cursorMu protects cursor on concurrent peek goroutne
	peekBuf      []byte        // peeked data
	peekReadSize int           // peeked data length
	peekErr      error         // error returned by ReadAt on peeking
	peekDone     chan struct{} // signals that the peek is done so that Read can copy peekBuf data (set after the first Read)
	closed       chan struct{} // terminates peek goroutine and unblocks Read method
	closeOnce    sync.Once     // protects closed channel on multiple calls to Close method
}

// NewLangos bakes a new yummy langos that peeks
// on provider reader when its Read or Seek methods are called.
// Argument peekSize sets the length of peeks.
func NewLangos(r Reader, peekSize int) *Langos {
	return &Langos{
		r:       r,
		peekBuf: make([]byte, peekSize),
		closed:  make(chan struct{}),
	}
}

func NewBufferedLangos(r Reader, bufferSize int) Reader {
	return NewBufferedReader(NewLangos(r, bufferSize), bufferSize)
}

func (l *Langos) Read(p []byte) (n int, err error) {
	log.Debug("langos Read", "cursor", l.cursor)

	// first read, no peeking happened before
	if l.peekDone == nil {
		// note: calling Seek(0, io.SeekStart) is safe to call
		// where checking l.cursor for first read would result
		// in double peek on the same range
		n, err := l.r.Read(p)
		if err != nil {
			return n, err
		}
		l.cursor = int64(n)
		l.peekDone = make(chan struct{}, 1)

		// peek for the second read
		go l.peek(l.cursor)
		return n, err
	}

	// second and further Read calls are waiting for peeks to finish
	select {
	case <-l.peekDone:
		if (l.peekErr == nil || l.peekErr == io.EOF) && l.peekReadSize > 0 {
			copy(p, l.peekBuf[:l.peekReadSize])
		}
		if l.peekErr != io.EOF {
			// peek from the current cursor
			go l.peek(l.cursor)
		}
		return l.peekReadSize, l.peekErr
	case <-l.closed:
		return 0, io.EOF
	}
}

func (l *Langos) Seek(offset int64, whence int) (int64, error) {
	n, err := l.r.Seek(offset, whence)
	if err != nil {
		return n, err
	}

	// protect cursor from peek method call
	// in different goroutine
	l.cursorMu.Lock()
	l.cursor = n
	l.cursorMu.Unlock()

	// get the peek from the new cursor
	// current peek result will be ignored
	go l.peek(n)
	return n, err
}

func (l *Langos) ReadAt(p []byte, off int64) (int, error) {
	return l.r.ReadAt(p, off)
}

func (l *Langos) peek(offset int64) {
	log.Debug("langos peek", "offset", offset, "peekReadSize", l.peekReadSize, "peekErr", l.peekErr)
	n, err := l.r.ReadAt(l.peekBuf, offset)
	log.Debug("langos peek ReadAt returned", "offset", offset, "n", n, "err", l.peekErr)

	// protect cursor from Seek method call
	// in different goroutine
	l.cursorMu.Lock()
	defer l.cursorMu.Unlock()

	// check if seek has been called
	// to disregard this peek result
	if l.cursor != offset {
		return
	}

	l.peekReadSize = n
	l.peekErr = err
	l.cursor += int64(n)

	select {
	// allow the Read method to return a copy of current peekBuf
	case l.peekDone <- struct{}{}:
	case <-l.closed:
	}
}

func (l *Langos) Close() (err error) {
	l.closeOnce.Do(func() {
		close(l.closed)
	})
	return nil
}
