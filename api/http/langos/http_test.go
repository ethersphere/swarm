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
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethersphere/swarm/api/http/langos"
)

// TestHTTPResponse validates that the langos returns correct data
// over http test server and ServeContent function.
func TestHTTPResponse(t *testing.T) {
	dataSize := 10 * 1024 * 1024
	bufferSize := 4 * 32 * 1024

	data := make([]byte, dataSize)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "test", time.Now(), langos.NewBufferedLangos(bytes.NewReader(data), bufferSize))
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, data) {
		t.Fatalf("got invalid data (lengths: got %v, want %v)", len(got), len(data))
	}
}

// TestHTTPResponse validates that the langos returns correct data
// over http test server and ServeContent function for http range requests.
func TestHTTPRangeResponse(t *testing.T) {
	dataSize := 10 * 1024 * 1024
	bufferSize := 4 * 32 * 1024

	data := make([]byte, dataSize)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "test", time.Now(), langos.NewBufferedLangos(bytes.NewReader(data), bufferSize))
	}))
	defer ts.Close()

	for i := 0; i < 100; i++ {
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		start := rand.Intn(dataSize)
		end := rand.Intn(dataSize-start) + start
		rangeHeader := fmt.Sprintf("bytes=%v-%v", start, end-1)
		if i == 0 {
			// test open ended range
			end = dataSize
			rangeHeader = fmt.Sprintf("bytes=%v-", start)
		}
		req.Header.Add("Range", rangeHeader)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		got, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		want := data[start:end]
		if !bytes.Equal(got, want) {
			t.Fatalf("got invalid data for range %s (lengths: got %v, want %v)", rangeHeader, len(got), len(want))
		}
	}
}

// BenchmarkHTTPDelayedReaders measures time needed by test http server to serve the body
// using different readers.
//
// goos: darwin
// goarch: amd64
// pkg: github.com/ethersphere/swarm/api/http/langos
// BenchmarkHTTPDelayedReaders/direct-8         	       8	 127489432 ns/op	 8390421 B/op	      26 allocs/op
// BenchmarkHTTPDelayedReaders/buffered-8       	      38	  28823581 ns/op	 8388504 B/op	      20 allocs/op
// BenchmarkHTTPDelayedReaders/langos-8         	     397	   3409622 ns/op	 8406115 B/op	     201 allocs/op
func BenchmarkHTTPDelayedReaders(b *testing.B) {
	dataSize := 2 * 1024 * 1024
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

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeContent(w, r, "test", time.Now(), bc.newReader())
			}))
			defer ts.Close()

			for i := 0; i < b.N; i++ {
				res, err := http.Get(ts.URL)
				if err != nil {
					b.Fatal(err)
				}

				b.StartTimer()
				got, err := ioutil.ReadAll(res.Body)
				b.StopTimer()

				res.Body.Close()
				if err != nil {
					b.Fatal(err)
				}
				if !bytes.Equal(got, data) {
					b.Fatalf("%v got invalid data (lengths: got %v, want %v)", i, len(got), len(data))
				}
			}
		})
	}
}
