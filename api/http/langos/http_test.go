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
	"mime"
	"mime/multipart"
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

	data := randomData(t, dataSize)

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

	data := randomData(t, dataSize)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "test", time.Now(), langos.NewBufferedLangos(bytes.NewReader(data), bufferSize))
	}))
	defer ts.Close()

	for i := 0; i < 100; i++ {
		start := rand.Intn(dataSize - 1)
		end := rand.Intn(dataSize-1-start) + start
		rangeHeader := fmt.Sprintf("bytes=%v-%v", start, end)
		if i == 0 {
			// test open ended range
			end = dataSize - 1
			rangeHeader = fmt.Sprintf("bytes=%v-", start)
		}

		gotRangs := httpRangeRequest(t, ts.URL, rangeHeader)
		got := gotRangs[0]
		want := data[start : end+1]
		if !bytes.Equal(got, want) {
			t.Fatalf("got invalid data for range %s (lengths: got %v, want %v)", rangeHeader, len(got), len(want))
		}
	}
}

// TestHTTPMultipleRangeResponse validates that the langos returns correct data
// over http test server and ServeContent function for http requests with multiple ranges.
func TestHTTPMultipleRangeResponse(t *testing.T) {
	dataSize := 10 * 1024 * 1024
	bufferSize := 4 * 32 * 1024

	data := randomData(t, dataSize)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "test", time.Now(), langos.NewBufferedLangos(bytes.NewReader(data), bufferSize))
	}))
	defer ts.Close()

	for i := 0; i < 100; i++ {
		var ranges [][2]int

		var wantParts [][]byte
		for i := rand.Intn(5); i >= 0; i-- {
			var beginning int
			if l := len(ranges); l > 0 {
				beginning = ranges[l-1][1]
			}
			if beginning >= dataSize {
				break
			}
			start := rand.Intn(dataSize-beginning) + beginning
			var end int
			if dataSize-1-start <= 0 {
				end = dataSize - 1
			} else {
				end = rand.Intn(dataSize-1-start) + start
			}
			if start == end {
				continue
			}
			ranges = append(ranges, [2]int{start, end})
			wantParts = append(wantParts, data[start:end+1])
		}

		rangeHeader := "bytes="
		for i, r := range ranges {
			if i > 0 {
				rangeHeader += ", "
			}
			rangeHeader += fmt.Sprintf("%v-%v", r[0], r[1])
		}

		gotParts := httpRangeRequest(t, ts.URL, rangeHeader)

		for i, want := range wantParts {
			got := gotParts[i]
			if !bytes.Equal(got, want) {
				t.Fatalf("got invalid data for range #%v %s (lengths: got %v, want %v)", i+1, rangeHeader, len(got), len(want))
			}
		}
	}
}

func httpRangeRequest(t *testing.T, url, rangeHeader string) (parts [][]byte) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Range", rangeHeader)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	mimetype, params, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if mimetype == "multipart/byteranges" {
		mr := multipart.NewReader(res.Body, params["boundary"])
		for part, err := mr.NextPart(); err == nil; part, err = mr.NextPart() {
			value, err := ioutil.ReadAll(part)
			if err != nil {
				t.Fatal(err)
			}
			parts = append(parts, value)
		}
	} else {
		value, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, value)
	}

	return parts
}

// BenchmarkHTTPDelayedReaders measures time needed by test http server to serve the body
// using different readers.
//
// goos: darwin
// goarch: amd64
// pkg: github.com/ethersphere/swarm/api/http/langos
// BenchmarkHTTPDelayedReaders/static_direct-8         	       8	 126790118 ns/op	 8390170 B/op	      25 allocs/op
// BenchmarkHTTPDelayedReaders/static_buffered-8       	      46	  28732654 ns/op	 8388969 B/op	      21 allocs/op
// BenchmarkHTTPDelayedReaders/static_langos-8         	     415	   2924194 ns/op	 8404185 B/op	     194 allocs/op
// BenchmarkHTTPDelayedReaders/random_direct-8         	       3	 380645816 ns/op	 8389432 B/op	      24 allocs/op
// BenchmarkHTTPDelayedReaders/random_buffered-8       	      12	  95311138 ns/op	 8389156 B/op	      22 allocs/op
// BenchmarkHTTPDelayedReaders/random_langos-8         	     381	   3149494 ns/op	 8404111 B/op	     193 allocs/op
func BenchmarkHTTPDelayedReaders(b *testing.B) {
	dataSize := 2 * 1024 * 1024
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
				return langos.NewBufferedLangos(newDelayedReader(bytes.NewReader(data), randomDelaysFunc), bufferSize)
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
