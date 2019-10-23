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
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
)

// NewBufferedReadSeeker runs a series of reads and seeks on
// bufferedReadSeeker instances with various buffer sizes.
func TestBufferedReadSeeker(t *testing.T) {
	multiSizeTester(t, func(t *testing.T, dataSize, bufferSize int) {
		data := randomData(t, dataSize)
		newReadSeekerTester(newBufferedReadSeeker(bytes.NewReader(data), bufferSize), data)(t)
	})
}

// TestReadSeekerTester tests newReadSeekerTester steps against the stdlib's
// bytes.Reader which is used as the reference implementation.
func TestReadSeekerTester(t *testing.T) {
	for _, size := range testDataSizes {
		data := randomData(t, parseDataSize(t, size))
		t.Run(size, newReadSeekerTester(bytes.NewReader(data), data))
	}
}

// newReadSeekerTester returns a new test function that performs a series of
// Read and Seek method calls to validate that provided io.ReadSeeker
// provide the expected functionality while reading data and seeking on it.
// Argument data must be the same as used in io.ReadSeeker as it is used
// in validations.
func newReadSeekerTester(rs io.ReadSeeker, data []byte) func(t *testing.T) {
	return func(t *testing.T) {
		read := func(t *testing.T, size int, want []byte, wantErr error) {
			t.Helper()

			b := make([]byte, size)
			for count := 0; count < len(want); {
				n, err := rs.Read(b[count:])
				if err != wantErr {
					t.Fatalf("got error %v, want %v", err, wantErr)
				}
				count += n
			}
			if !bytes.Equal(b, want) {
				t.Fatal("invalid read data")
			}
		}

		seek := func(t *testing.T, offset, whence, wantPosition int, wantErr error) {
			t.Helper()

			n, err := rs.Seek(int64(offset), whence)
			if err != wantErr {
				t.Fatalf("got error %v, want %v", err, wantErr)
			}
			if n != int64(wantPosition) {
				t.Fatalf("got position %v, want %v", n, wantPosition)
			}
		}

		l := len(data)

		// Test sequential reads
		readSize1 := l / 5
		read(t, readSize1, data[:readSize1], nil)
		readSize2 := l / 6
		read(t, readSize2, data[readSize1:readSize1+readSize2], nil)
		readSize3 := l / 4
		read(t, readSize3, data[readSize1+readSize2:readSize1+readSize2+readSize3], nil)

		// Test seek and read
		seekSize1 := l / 4
		seek(t, seekSize1, io.SeekStart, seekSize1, nil)
		readSize1 = l / 5
		read(t, readSize1, data[seekSize1:seekSize1+readSize1], nil)
		readSize2 = l / 10
		read(t, readSize2, data[seekSize1+readSize1:seekSize1+readSize1+readSize2], nil)

		// Test get size and read from start
		seek(t, 0, io.SeekEnd, l, nil)
		seek(t, 0, io.SeekStart, 0, nil)
		readSize1 = l / 6
		read(t, readSize1, data[:readSize1], nil)

		// Test read end
		seek(t, 0, io.SeekEnd, l, nil)
		read(t, 0, nil, io.EOF)

		// Test read near end
		seekOffset := 1 / 10
		seek(t, seekOffset, io.SeekEnd, l-seekOffset, nil)
		read(t, seekOffset, data[l-seekOffset:], io.EOF)

		// Test seek from current with reads
		seek(t, 0, io.SeekStart, 0, nil)
		seekSize1 = l / 3
		seek(t, seekSize1, io.SeekCurrent, seekSize1, nil)
		readSize1 = l / 8
		read(t, readSize1, data[seekSize1:seekSize1+readSize1], nil)

	}
}

// randomDataCache keeps random data in memory between tests
// to avoid regenerating random data for every test or subtest.
var randomDataCache []byte

// randomData returns a byte slice with random data.
// This function is not safe for concurrent use.
func randomData(t testing.TB, size int) []byte {
	t.Helper()

	if cacheSize := len(randomDataCache); cacheSize < size {
		data := make([]byte, size-cacheSize)
		_, err := rand.Read(data)
		if err != nil {
			t.Fatal(err)
		}
		randomDataCache = append(randomDataCache, data...)
	}

	return randomDataCache[:size]
}

var (
	testDataSizes   = []string{"100", "749", "1k", "128k", "749k", "1M", "10M"}
	testBufferSizes = []string{"1k", "128k", "753k", "1M", "10M", "25M"}
)

// multiSizeTester performs a series of subtests with different data and buffer sizes.
func multiSizeTester(t *testing.T, newTestFunc func(t *testing.T, dataSize, bufferSize int)) {
	t.Helper()

	for _, dataSize := range testDataSizes {
		for _, bufferSize := range testBufferSizes {
			t.Run(fmt.Sprintf("data %s buffer %s", dataSize, bufferSize), func(t *testing.T) {
				newTestFunc(t, parseDataSize(t, dataSize), parseDataSize(t, bufferSize))
			})
		}
	}
}

func parseDataSize(t *testing.T, v string) (s int) {
	t.Helper()

	multiplier := 1
	for suffix, value := range map[string]int{
		"k": 1024,
		"M": 1024 * 1024,
	} {
		if strings.HasSuffix(v, suffix) {
			v = strings.TrimSuffix(v, suffix)
			multiplier = value
			break
		}
	}
	s, err := strconv.Atoi(v)
	if err != nil {
		t.Fatal(err)
	}
	return s * multiplier
}
