package http

import (
	"io"
	"sync"

	"github.com/ethersphere/swarm/log"
)

// langos is a reader with a lookahead peekBuffer
// this is the most naive implementation of a lookahead peekBuffer
// it should issue a lookahead Read when a Read is called, hence
// the name - langos
// |--->====>>------------|
//    cur   topmost
// the first read is not a lookahead but the rest are
// so, it could be that a lookahead read might need to wait for a previous read to finish
// due to resource pooling
//
// Limitations:
//  - Read and Seek methods are not concurrent safe and must be called synchronously.
//  - After io.EOF error is returned by the Read method, no more calls on Read or Seek are allowed.
//  - Close method can be called only once.
type langos struct {
	r            reader        // reader needs to implement io.ReadSeeker and io.ReaderAt interfaces
	cursor       int64         // current read position
	cursorMu     sync.Mutex    // cursorMu protects cursor on concurrent peek goroutne
	peekBuf      []byte        // peeked data
	peekReadSize int           // peeked data length
	peekErr      error         // error returned by ReadAt on peeking
	peekDone     chan struct{} // signals that the peek is done so that Read can copy peekBuf data (set after the first Read)
	closed       chan struct{} // terminates peek goroutine and unblocks Read method
}

// newLangos bakes a new yummy langos that peeks
// on provider reader when its Read or Seek methods are called.
// Argument peekSize sets the length of peeks.
func newLangos(r reader, peekSize int) *langos {
	return &langos{
		r:       r,
		peekBuf: make([]byte, peekSize),
		closed:  make(chan struct{}),
	}
}

func (l *langos) Read(p []byte) (n int, err error) {
	log.Debug("l.Read", "cursor", l.cursor)

	// first read, no peeking happened before
	if l.peekDone == nil {
		// note: calling Seek(0, io.SeekStart) is safe to call
		// where checking l.cursor for first read would result
		// in double peek on the same range
		log.Debug("firstRead")
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
	case _, ok := <-l.peekDone:
		if (l.peekErr == nil || l.peekErr == io.EOF) && l.peekReadSize > 0 {
			log.Debug("copying")
			copy(p, l.peekBuf[:l.peekReadSize])
		}
		if l.peekErr == nil && ok {
			// peek from the current cursor
			go l.peek(l.cursor)
		}
		return l.peekReadSize, l.peekErr
	case <-l.closed:
		return 0, io.EOF
	}
}

func (l *langos) Seek(offset int64, whence int) (int64, error) {
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

func (l *langos) peek(offset int64) {
	log.Debug("l.peek", "offset", offset, "lastN", l.peekReadSize, "peekErr", l.peekErr)
	n, err := l.r.ReadAt(l.peekBuf, offset)

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
	log.Debug("peek readat returned", "offset", offset, "n", n, "err", l.peekErr)
	if err == io.EOF {
		log.Debug("peek EOF")
		// no more peeking when EOF is reached
		close(l.peekDone)
		return
	}
	l.cursor += int64(n)
	select {
	case l.peekDone <- struct{}{}:
	case <-l.closed:
	}
}

func (l *langos) Close() (err error) {
	close(l.closed)
	return nil
}

type reader interface {
	io.ReadSeeker
	io.ReaderAt
}
