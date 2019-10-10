package http

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestLangosCallsPeekOnlyTwice(t *testing.T) {
	// data length is 12 bytes
	rdr := strings.NewReader("sometestdata")
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
			expReads:    2,
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
				reader: &nopSeeker{
					rdr,
					rdr,
				},
			}
			l := newLangos(tl)

			b := make([]byte, segmentSize)
			var err error
			for i := 1; i < tc.numReads; i++ {
				_, err = l.Read(b)
				time.Sleep(5 * time.Millisecond)
			}

			if err != tc.expErr {
				t.Fatal(err)
			}
			time.Sleep(5 * time.Millisecond)
			if tl.readCount != tc.expReads {
				t.Fatalf("expected %d call to read func, got %d", tc.expReads, tl.readCount)
			}
			//_, err = l.Read(b)
			//if err != nil {
			//t.Fatal(err)
			//}

			//time.Sleep(5 * time.Millisecond)
			//if tl.readCount != exp {
			//t.Fatalf("expected %d call to peek func, got %d", exp, tl.readCount)
			//}
		})
	}
}

func TestLangosCallsPeek(t *testing.T) {
	rdr := strings.NewReader("sometestdata")
	tl := &testLangos{
		reader: &nopSeeker{
			rdr,
			rdr,
		},
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

type nopSeeker struct {
	io.Reader
	//io.Seeker
	io.ReaderAt
}

func (nopSeeker) Seek(offset int64, whence int) (int64, error) { return 0, nil }

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
