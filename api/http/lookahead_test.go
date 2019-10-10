package http

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestLangosCallsPeekOnlyTwice(t *testing.T) {
	defer func(c int) {
		segmentSize = c
	}(segmentSize)

	for _, tc := range []struct {
		name        string
		segmentSize int
		numReads    int
		expReads    int
		expErr      error
	}{
		{
			name:        "2 seq reads, no error",
			segmentSize: 6,
			numReads:    2,
			expReads:    3, // additional read detects EOF
			expErr:      nil,
		},
		{
			name:        "3 seq reads, EOF",
			segmentSize: 6,
			numReads:    3,
			expReads:    3,
			expErr:      io.EOF,
		},
		{
			name:        "2 seq reads, EOF",
			segmentSize: 7,
			numReads:    2,
			expReads:    2,
			expErr:      io.EOF,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			segmentSize = tc.segmentSize

			tl := &testLangos{
				reader: strings.NewReader("sometestdata"), // len 12
			}
			l := newLangos(tl)

			b := make([]byte, segmentSize)
			var err error
			for i := 1; i <= tc.numReads; i++ {
				var wantErr error
				if i == tc.numReads {
					wantErr = tc.expErr
				}
				_, err = l.Read(b)
				if err != wantErr {
					t.Fatalf("got read #%v error %v, want %v", i, err, wantErr)
				}
			}
			if tc.numReads != tc.expReads {
				// wait for peek to finish
				// so that it can be counted
				time.Sleep(10 * time.Millisecond)
			}
			if tl.readCount != tc.expReads {
				t.Fatalf("expected %d call to read func, got %d", tc.expReads, tl.readCount)
			}
		})
	}
}

func TestLangosCallsPeek(t *testing.T) {
	rdr := strings.NewReader("sometestdata")
	tl := &testLangos{
		reader: rdr,
	}
	l := newLangos(tl)

	b := make([]byte, segmentSize)
	_, err := l.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	exp := 2
	time.Sleep(5 * time.Millisecond)
	if tl.readCount != exp {
		t.Fatalf("expected %d call to peek func, got %d", exp, tl.readCount)
	}
}

type testLangos struct {
	reader
	readCount int
}

func (l *testLangos) Read(p []byte) (n int, err error) {
	l.readCount++
	return l.reader.Read(p)
}

func (l *testLangos) ReadAt(p []byte, off int64) (int, error) {
	l.readCount++
	return l.reader.ReadAt(p, off)
}
