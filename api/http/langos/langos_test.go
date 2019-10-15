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
			numReads: 1,
			expReads: 2,
			expErr:   nil,
		},
		{
			name:     "3 seq reads, EOF",
			peekSize: 6,
			numReads: 3,
			expReads: 4,
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
			cr := newCounterReader(strings.NewReader(testData))
			l := langos.NewLangos(cr, tc.peekSize)

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

			testReadCount(t, cr, tc.expReads)
		})
	}
}

func TestLangosCallsPeek(t *testing.T) {
	peekSize := 128
	cr := newCounterReader(strings.NewReader("sometestdata"))
	l := langos.NewLangos(cr, peekSize)

	b := make([]byte, peekSize)
	_, err := l.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	testReadCount(t, cr, 2)
}

// counterReader counts the number of Read or ReadAt calls.
type counterReader struct {
	langos.Reader
	readCount int
	mu        sync.Mutex
}

func newCounterReader(r langos.Reader) (cr *counterReader) {
	return &counterReader{
		Reader: r,
	}
}

func (cr *counterReader) Read(p []byte) (n int, err error) {
	cr.mu.Lock()
	cr.readCount++
	cr.mu.Unlock()
	return cr.Reader.Read(p)
}

func (cr *counterReader) ReadAt(p []byte, off int64) (int, error) {
	cr.mu.Lock()
	cr.readCount++
	cr.mu.Unlock()
	return cr.Reader.ReadAt(p, off)
}

func (cr *counterReader) ReadCount() (c int) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.readCount
}

func testReadCount(t *testing.T, cr *counterReader, want int) {
	t.Helper()

	var got int
	// retry for 2s to give the peek goroutine time to finish
	for i := 0; i < 400; i++ {
		got = cr.ReadCount()
		if got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("got %d, want %d call to read func", got, want)
}

// BenchmarkDelayedReaders performs benchmarks on reader with deterministic and random
// delays on every Read method call. Function ioutil.ReadAll is used for reading.
//
//  - direct: a baseline on plain reader
//  - buffered: reading through bufio.Reader
//  - langos: reading through regular langos
//  - bufferd langos: reading through buffered langos
//
// goos: darwin
// goarch: amd64
// pkg: github.com/ethersphere/swarm/api/http/langos
// BenchmarkDelayedReaders/static_direct-8                      	      28	  38975129 ns/op	33552515 B/op	      18 allocs/op
// BenchmarkDelayedReaders/static_buffered-8                    	      40	  31634832 ns/op	33683734 B/op	      21 allocs/op
// BenchmarkDelayedReaders/static_langos-8                      	      92	  12413853 ns/op	33828891 B/op	    1507 allocs/op
// BenchmarkDelayedReaders/static_buffered_langos-8             	     100	  12357659 ns/op	33948417 B/op	    1389 allocs/op
// BenchmarkDelayedReaders/random_direct-8                      	      13	 107649125 ns/op	33552476 B/op	      17 allocs/op
// BenchmarkDelayedReaders/random_buffered-8                    	      19	  62106573 ns/op	33683683 B/op	      20 allocs/op
// BenchmarkDelayedReaders/random_langos-8                      	      72	  17809501 ns/op	33829083 B/op	    1508 allocs/op
// BenchmarkDelayedReaders/random_buffered_langos-8             	      84	  17836451 ns/op	33948630 B/op	    1391 allocs/op
func BenchmarkDelayedReaders(b *testing.B) {
	dataSize := 10 * 1024 * 1024
	bufferSize := 4 * 32 * 1024

	data := randomData(b, dataSize)

	for _, bc := range []struct {
		name      string
		newReader func() langos.Reader
	}{
		{
			name: "static direct",
			newReader: func() langos.Reader {
				return newDelayedReaderStatic(bytes.NewReader(data), defaultStaticDelays)
			},
		},
		{
			name: "static buffered",
			newReader: func() langos.Reader {
				return langos.NewBufferedReadSeeker(newDelayedReaderStatic(bytes.NewReader(data), defaultStaticDelays), bufferSize)
			},
		},
		{
			name: "static langos",
			newReader: func() langos.Reader {
				return langos.NewLangos(newDelayedReaderStatic(bytes.NewReader(data), defaultStaticDelays), bufferSize)
			},
		},
		{
			name: "static buffered langos",
			newReader: func() langos.Reader {
				return langos.NewBufferedLangos(newDelayedReaderStatic(bytes.NewReader(data), defaultStaticDelays), bufferSize)
			},
		},
		{
			name: "random direct",
			newReader: func() langos.Reader {
				return newDelayedReader(bytes.NewReader(data), randomDelaysFunc)
			},
		},
		{
			name: "random buffered",
			newReader: func() langos.Reader {
				return langos.NewBufferedReadSeeker(newDelayedReader(bytes.NewReader(data), randomDelaysFunc), bufferSize)
			},
		},
		{
			name: "random langos",
			newReader: func() langos.Reader {
				return langos.NewLangos(newDelayedReader(bytes.NewReader(data), randomDelaysFunc), bufferSize)
			},
		},
		{
			name: "random buffered langos",
			newReader: func() langos.Reader {
				return langos.NewBufferedLangos(newDelayedReader(bytes.NewReader(data), randomDelaysFunc), bufferSize)
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

func newDelayedReader(r langos.Reader, f delayedReaderFunc) *delayedReader {
	return &delayedReader{
		Reader: r,
		f:      f,
	}
}

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

var (
	defaultStaticDelays = []time.Duration{
		2 * time.Millisecond,
		0, 0, 0,
		5 * time.Millisecond,
		0, 0,
		10 * time.Millisecond,
		0, 0,
	}
	randomDelaysFunc = func(_ int) (delay time.Duration) {
		// max delay 10ms
		return time.Duration(rand.Intn(10 * int(time.Millisecond)))
	}
)

func randomData(t testing.TB, size int) (data []byte) {
	t.Helper()

	data = make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
