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
	lastSegment  int
	lastReadSize int
	lastErr      error
	peekDone     chan struct{}
}

func newLangos(r reader) *langos {
	l := &langos{
		r:        r,
		buf:      make([]byte, segmentSize),
		peekDone: make(chan struct{}),
	}
	return l
}

func (l *langos) Read(p []byte) (n int, err error) {
	log.Debug("l.Read", "last", l.lastSegment)
	if l.lastSegment == 0 {
		l.lastSegment++
		log.Debug("firstRead")
		n, err := l.r.Read(p)
		if err != nil {
			return n, err
		}
		go l.peek()
		return n, err
	}
	select {
	case _, ok := <-l.peekDone:
		if l.lastErr == nil || l.lastErr == io.EOF {
			log.Debug("copying")
			copy(p, l.buf[:l.lastReadSize])
		}
		if l.lastErr == nil && ok {
			go l.peek()
		}
	}

	return l.lastReadSize, l.lastErr
}

func (l *langos) Seek(offset int64, whence int) (int64, error) {
	// todo: handle peek buffer invalidation
	return l.r.Seek(offset, whence)
}

func (l *langos) peek() {
	log.Debug("l.peek", "last", l.lastSegment, "lastN", l.lastReadSize, "lastErr", l.lastErr)
	n, err := l.r.ReadAt(l.buf, int64(l.lastSegment*segmentSize))
	l.lastReadSize = n
	l.lastErr = err
	log.Debug("peek readat returned", "lastSegment", l.lastSegment, "err", l.lastErr)
	if err == io.EOF {
		log.Debug("peek EOF")
		close(l.peekDone)
		return
	}
	l.peekDone <- struct{}{}
	l.lastSegment++
}

type reader interface {
	io.ReadSeeker
	io.ReaderAt
}
