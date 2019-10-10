package http

import (
	"io"

	"github.com/ethersphere/swarm/log"
)

var segmentSize = 4 * 32 * 1024
var bufferSize = segmentSize // in the future could be 5 * segmentSize

// langos is a reader with a lookahead buffer
// this is the most naive implementation of a lookahead buffer
// it should issue a lookahead Read when a Read is called, hence
// the name - langos
// |--->====>>------------|
//    cur   topmost
// the first read is not a lookahead but the rest are
// so, it could be that a lookahead read might need to wait for a previous read to finish
// due to resource pooling
type langos struct {
	r            reader
	buf          []byte
	cursor       int64
	peekDone     chan struct{}
	peekReadSize int
	peekErr      error
	closed       chan struct{}
}

func newLangos(r reader) *langos {
	l := &langos{
		r:        r,
		buf:      make([]byte, segmentSize),
		peekDone: make(chan struct{}),
		closed:   make(chan struct{}),
	}
	return l
}

func (l *langos) Read(p []byte) (n int, err error) {
	log.Debug("l.Read", "cursor", l.cursor)
	if l.cursor == 0 {
		log.Debug("firstRead")
		n, err := l.r.Read(p)
		if err != nil {
			return n, err
		}
		l.cursor = int64(n)
		go l.peek()
		return n, err
	}
	select {
	case _, ok := <-l.peekDone:
		if (l.peekErr == nil || l.peekErr == io.EOF) && l.peekReadSize > 0 {
			log.Debug("copying")
			copy(p, l.buf[:l.peekReadSize])
		}
		if l.peekErr == nil && ok {
			go l.peek()
		}
	case <-l.closed:
	}

	return l.peekReadSize, l.peekErr
}

func (l *langos) Seek(offset int64, whence int) (int64, error) {
	// todo: handle peek buffer invalidation
	return l.r.Seek(offset, whence)
}

func (l *langos) peek() {
	log.Debug("l.peek", "cursor", l.cursor, "lastN", l.peekReadSize, "peekErr", l.peekErr)
	n, err := l.r.ReadAt(l.buf, l.cursor)
	l.peekReadSize = n
	l.peekErr = err
	log.Debug("peek readat returned", "cursor", l.cursor, "err", l.peekErr)
	if err == io.EOF {
		log.Debug("peek EOF")
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
