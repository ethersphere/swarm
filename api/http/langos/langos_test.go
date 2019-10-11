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

package langos_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethersphere/swarm/api/http/langos"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
			tl := newCounterReader(strings.NewReader(testData))
			l := langos.NewLangos(tl, tc.peekSize)

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
			if readCount := tl.ReadCount(); readCount != tc.expReads {
				t.Fatalf("expected %d call to read func, got %d", tc.expReads, readCount)
			}
		})
	}
}

func TestLangosCallsPeek(t *testing.T) {
	peekSize := 128
	tl := newCounterReader(strings.NewReader("sometestdata"))
	l := langos.NewLangos(tl, peekSize)

	b := make([]byte, peekSize)
	_, err := l.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	exp := 2
	// wait for the peek goroutine to finish
	time.Sleep(5 * time.Millisecond)
	if readCount := tl.ReadCount(); readCount != exp {
		t.Fatalf("expected %d call to read func, got %d", exp, readCount)
	}
}

// counterReader counts the number of Read or ReadAt calls.
type counterReader struct {
	langos.Reader
	readCount int
	mu        sync.Mutex
}

func newCounterReader(r langos.Reader) (c *counterReader) {
	return &counterReader{
		Reader: r,
	}
}

func (l *counterReader) Read(p []byte) (n int, err error) {
	l.mu.Lock()
	l.readCount++
	l.mu.Unlock()
	return l.Reader.Read(p)
}

func (l *counterReader) ReadAt(p []byte, off int64) (int, error) {
	l.mu.Lock()
	l.readCount++
	l.mu.Unlock()
	return l.Reader.ReadAt(p, off)
}

func (l *counterReader) ReadCount() (c int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.readCount
}

// BenchmarkDelayedReaders performs benchmarks on reader with deterministic
// delays on every Read method call. Function ioutil.ReadAll is used for reading.
//
//  - direct: a baseline on plain reader
//  - buffered: reading through bufio.Reader
//  - langos: reading through buffered langos
//
// goos: darwin
// goarch: amd64
// pkg: github.com/ethersphere/swarm/api/http/langos
// BenchmarkDelayedReaders/direct-8         	      25	  40113969 ns/op	33552685 B/op	      18 allocs/op
// BenchmarkDelayedReaders/buffered-8       	      40	  30294152 ns/op	33683763 B/op	      21 allocs/op
// BenchmarkDelayedReaders/langos-8         	      86	  11935154 ns/op	33908081 B/op	     985 allocs/op
func BenchmarkDelayedReaders(b *testing.B) {
	dataSize := 10 * 1024 * 1024
	bufferSize := 4 * 32 * 1024

	data := make([]byte, dataSize)
	_, err := rand.Read(data)
	if err != nil {
		b.Fatal(err)
	}

	delays := []time.Duration{
		2 * time.Millisecond,
		0, 0, 0,
		5 * time.Millisecond,
		0, 0,
		10 * time.Millisecond,
		0, 0,
	}

	for _, bc := range []struct {
		name      string
		newReader func() langos.Reader
	}{
		{
			name: "direct",
			newReader: func() langos.Reader {
				return newDelayedReaderStatic(bytes.NewReader(data), delays)
			},
		},
		{
			name: "buffered",
			newReader: func() langos.Reader {
				return langos.NewBufferedReadSeeker(newDelayedReaderStatic(bytes.NewReader(data), delays), bufferSize)
			},
		},
		{
			name: "langos",
			newReader: func() langos.Reader {
				return langos.NewBufferedLangos(newDelayedReaderStatic(bytes.NewReader(data), delays), bufferSize)
			},
		},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.StopTimer()
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				got, err := ioutil.ReadAll(bc.newReader())
				b.StopTimer()

				if err != nil {
					b.Fatal(err)
				}
				if !bytes.Equal(got, data) {
					b.Fatalf("got invalid data (lengths: got %v, want %v)", len(got), len(data))
				}
			}
		})
	}
}

type delayedReaderFunc func(i int) (delay time.Duration)

type delayedReader struct {
	langos.Reader
	f delayedReaderFunc
	i int
}

// func newDelayedReader(r langos.Reader, f delayedReaderFunc) *delayedReader {
// 	return &delayedReader{
// 		Reader: r,
// 		f:      f,
// 	}
// }

func newDelayedReaderStatic(r langos.Reader, delays []time.Duration) *delayedReader {
	l := len(delays)
	return &delayedReader{
		Reader: r,
		f: func(i int) (delay time.Duration) {
			return delays[i%l]
		},
	}
}

func (d *delayedReader) Read(p []byte) (n int, err error) {
	time.Sleep(d.f(d.i))
	d.i++
	return d.Reader.Read(p)
}
