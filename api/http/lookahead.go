package http

import "io"

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
	s            reader
	size         int64
	buf          []byte
	lastSegment  int
	lastReadSize int
	lastErr      error
	peekDone     chan struct{}
}

func newLangos(r reader) *langos {
	l := &langos{
		s:        r,
		buf:      make([]byte, segmentSize),
		peekDone: make(chan struct{}, 1),
	}
	return l
}

func (l *langos) Read(p []byte) (n int, err error) {
	l.lastSegment++
	if l.lastSegment == 1 {
		n, err := l.s.Read(p)
		if err != nil {
			return n, err
		}
		go l.peek()
		return n, err
	}
	select {
	case _, ok := <-l.peekDone:
		if l.lastErr == nil || l.lastErr == io.EOF {
			copy(p, l.buf[:l.lastReadSize])
		}
		if l.lastErr != nil || !ok {
			break
		}
		go l.peek()
	}

	return l.lastReadSize, l.lastErr
}

func (l *langos) Seek(offset int64, whence int) (int64, error) {
	panic("not implemented")
}

func (l *langos) peek() {
	n, err := l.s.ReadAt(l.buf, int64(l.lastSegment*segmentSize))
	l.lastReadSize = n
	l.lastErr = err
	if err == io.EOF {
		close(l.peekDone)
		return
	}
	l.peekDone <- struct{}{}
}

type reader interface {
	io.ReadSeeker
	io.ReaderAt
}
