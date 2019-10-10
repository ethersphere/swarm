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
	"io"
	"strings"
	"testing"
	"time"
)

func TestLangosCallsPeekOnlyTwice(t *testing.T) {
	testData := "sometestdata" // len 12

	for _, tc := range []struct {
		name     string
		peekSize int
		numReads int
		expReads int
		expErr   error
	}{
		{
			name:     "2 seq reads, no error",
			peekSize: 6,
			numReads: 2,
			expReads: 3, // additional read detects EOF
			expErr:   nil,
		},
		{
			name:     "3 seq reads, EOF",
			peekSize: 6,
			numReads: 3,
			expReads: 3,
			expErr:   io.EOF,
		},
		{
			name:     "2 seq reads, EOF",
			peekSize: 7,
			numReads: 2,
			expReads: 2,
			expErr:   io.EOF,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tl := &testLangos{
				reader: strings.NewReader(testData),
			}
			l := newLangos(tl, tc.peekSize)
			defer l.Close()

			b := make([]byte, tc.peekSize)
			var err error
			for i := 1; i <= tc.numReads; i++ {
				var wantErr error
				if i == tc.numReads {
					wantErr = tc.expErr
				}
				var n int
				n, err = l.Read(b)
				if err != wantErr {
					t.Fatalf("got read #%v error %v, want %v", i, err, wantErr)
				}
				end := i * tc.peekSize
				if end > len(testData) {
					end = len(testData)
				}
				want := testData[(i-1)*tc.peekSize : end]
				if l := len(want); l != n {
					t.Fatalf("got read count #%v %v, want %v", i, n, l)
				}
				got := string(b[:n])
				if got != want {
					t.Fatalf("got read data #%v %q, want %q", i, got, want)
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
	peekSize := 128
	rdr := strings.NewReader("sometestdata")
	tl := &testLangos{
		reader: rdr,
	}
	l := newLangos(tl, peekSize)

	b := make([]byte, peekSize)
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
