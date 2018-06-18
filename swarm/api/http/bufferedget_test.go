// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package http

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// Value of io.Copy internal buffer size
const ioCopyBufferSize = 32 * 1024

// TestBufferedGetBzz tests the change of download time with different
// buffer sizes for requests to bzz scheme.
func TestBufferedGetBzz(t *testing.T) {
	testBufferedGet(t, "bzz")
}

// TestBufferedGetBzzRaw tests the change of download time with different
// buffer sizes for requests to bzz-raw scheme.
func TestBufferedGetBzzRaw(t *testing.T) {
	testBufferedGet(t, "bzz-raw")
}

// Upload one file with random data and compare download times for
// different buffers set on LazySectionReader.
// This test uses linear regression to detect the slope of interpolated
// durations. The slope should be a negative number as download durations
// get lower with increasing buffer size.
func testBufferedGet(t *testing.T, scheme string) {
	srv, err := newTestServer(100 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	key, data, err := upload(client, srv.URL+"/"+scheme+":/", 16*ioCopyBufferSize)
	if err != nil {
		t.Fatal(err)
	}

	// Test buffer sizes from the smallest of default that uses io.Copy
	// to the size of the uploaded data.
	// If buffer size is less then the io.Copy's or more
	// then the size of uploaded data, there should be no significant changes
	// in download time.
	bufferSizes := []int{
		ioCopyBufferSize,
		2 * ioCopyBufferSize,
		4 * ioCopyBufferSize,
		8 * ioCopyBufferSize,
		16 * ioCopyBufferSize,
	}

	durations, err := downloadWithBufferSizes(client, srv.URL+"/"+scheme+":/"+key+"/", bufferSizes, data)
	if err != nil {
		t.Fatal(err)
	}

	slope, _ := linearRegressionFloat64(durations...)

	// Negative slope indicates that the reported durations are in regression
	// by raising the buffer.
	// Value of -0.32 indicates that the slope is steep enough for significant
	// differences durations.
	expectedSlope := -0.32
	if slope > expectedSlope {
		t.Errorf("got slope %v, expected it less then %v", slope, expectedSlope)
	} else {
		t.Logf("durations slope %v, target less then %v", slope, expectedSlope)
	}
}

// TestUnderbufferedGetBzz tests that there are no changes in
// download times if the buffer is smaller the io.Copy's one, for bzz scheme.
func TestUnderbufferedGetBzz(t *testing.T) {
	testUnderbufferedGet(t, "bzz")
}

// TestUnderbufferedGetBzzRaw tests that there are no changes in
// download times if the buffer is smaller the io.Copy's one, for bzz-raw scheme.
func TestUnderbufferedGetBzzRaw(t *testing.T) {
	testUnderbufferedGet(t, "bzz-raw")
}

// Upload one file with random data and compare download times for
// different buffers set on LazySectionReader.
// All buffer sizes are less or equal then the one that io.Copy
// is using internally, so there should be no change in download times.
// This is detected with linear regression's slope being very close to 0.
func testUnderbufferedGet(t *testing.T, scheme string) {
	srv, err := newTestServer(200 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	key, data, err := upload(client, srv.URL+"/"+scheme+":/", 4*ioCopyBufferSize)
	if err != nil {
		t.Fatal(err)
	}

	bufferSizes := []int{
		ioCopyBufferSize / 16,
		ioCopyBufferSize / 8,
		ioCopyBufferSize / 4,
		ioCopyBufferSize / 2,
		ioCopyBufferSize,
	}

	durations, err := downloadWithBufferSizes(client, srv.URL+"/"+scheme+":/"+key+"/", bufferSizes, data)
	if err != nil {
		t.Fatal(err)
	}

	slope, _ := linearRegressionFloat64(durations...)

	if slope > 0.05 || slope < -0.05 {
		t.Errorf("got slope %v, expected 0 +- 0.05", slope)
	} else {
		t.Logf("durations slope %v, target 0 +- 0.05", slope)
	}
}

// TestOverbufferedGetBzz tests that there are no changes in
// download times if the buffer larger then the uploaded data, for bzz scheme.
func TestOverbufferedGetBzz(t *testing.T) {
	testOverbufferedGet(t, "bzz")
}

// TestOverbufferedGetBzzRaw tests that there are no changes in
// download times if the buffer larger then the uploaded data, for bzz-raw scheme.
func TestOverbufferedGetBzzRaw(t *testing.T) {
	testOverbufferedGet(t, "bzz-raw")
}

// Upload one file with random data and compare download times for
// different buffers set on LazySectionReader.
// All buffer sizes are greater or equal then the uploaded data size,
// so there should be no change in download times.
// This is detected with linear regression's slope being very close to 0.
func testOverbufferedGet(t *testing.T, scheme string) {
	srv, err := newTestServer(50 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	key, data, err := upload(client, srv.URL+"/"+scheme+":/", 4*ioCopyBufferSize)
	if err != nil {
		t.Fatal(err)
	}

	bufferSizes := []int{
		4 * ioCopyBufferSize,
		8 * ioCopyBufferSize,
		16 * ioCopyBufferSize,
		32 * ioCopyBufferSize,
		64 * ioCopyBufferSize,
	}

	durations, err := downloadWithBufferSizes(client, srv.URL+"/"+scheme+":/"+key+"/", bufferSizes, data)
	if err != nil {
		t.Fatal(err)
	}

	slope, _ := linearRegressionFloat64(durations...)

	if slope > 0.05 || slope < -0.05 {
		t.Errorf("got slope %v, expected 0 +- 0.05", slope)
	} else {
		t.Logf("durations slope %v, target 0 +- 0.05", slope)
	}
}

// testServer is a wrapper around httptest.Server for testing buffered responses.
type testServer struct {
	*httptest.Server
	fileStore *storage.FileStore
	closeFunc func()
}

func newTestServer(getFuncDelay time.Duration) (*testServer, error) {
	dir, err := ioutil.TempDir("", "swarm-http-server")
	if err != nil {
		return nil, err
	}
	storeParams := storage.NewDefaultLocalStoreParams()
	storeParams.Init(dir)
	localStore, err := storage.NewLocalStore(storeParams, nil)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	fileStore := storage.NewFileStore(
		&slowLocalStore{
			LocalStore:   localStore,
			getFuncDelay: getFuncDelay,
		},
		storage.NewFileStoreParams(),
	)

	srv := httptest.NewServer(&Server{
		api: api.NewAPI(fileStore, nil, nil),
	})
	return &testServer{
		Server:    srv,
		fileStore: fileStore,
		closeFunc: func() {
			srv.Close()
			os.RemoveAll(dir)
		},
	}, nil
}

func (t *testServer) Close() {
	t.closeFunc()
}

// slowLocalStore wraps storage.LocalStore to slow down
// Get function in order to detect changes in download
// time by changing buffer size resulting more parallel
// chunk get function calls.
type slowLocalStore struct {
	*storage.LocalStore
	getFuncDelay time.Duration
}

func (s slowLocalStore) Get(addr storage.Address) (chunk *storage.Chunk, err error) {
	time.Sleep(s.getFuncDelay)
	return s.LocalStore.Get(addr)
}

func upload(client *http.Client, url string, size int) (key string, data []byte, err error) {
	data = make([]byte, size)
	rand.Seed(time.Now().UnixNano())
	rand.Read(data)

	resp, err := client.Post(url, "text/plain", bytes.NewReader(data))
	if err != nil {
		return "", nil, fmt.Errorf("http post: %v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response body: %v", err)
	}

	return string(b), data, nil
}

func downloadAndCheck(client *http.Client, url string, checkData []byte) (d time.Duration, err error) {
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("http get: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %v", resp.Status)
	}

	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read body: %v", err)
	}

	if !bytes.Equal(checkData, got) {
		return 0, errors.New("uploaded and downloaded data differ")
	}
	return time.Since(start), nil
}

func downloadWithBufferSizes(client *http.Client, url string, bufferSizes []int, checkData []byte) (durations []float64, err error) {
	durations = make([]float64, len(bufferSizes))

	// set getFileBufferSize to the initial value at the end of the test
	defer func(s int) { getFileBufferSize = s }(getFileBufferSize)

	for i, b := range bufferSizes {
		getFileBufferSize = b

		d, err := downloadAndCheck(client, url, checkData)
		if err != nil {
			return nil, err
		}

		durations[i] = d.Seconds()
	}

	return durations, nil
}

// Interpolate provided float 64 points giving back
// the parameters for x = y*a + b linear equation.
// Parameter a is the slope and b the shift on the y-axes.
func linearRegressionFloat64(points ...float64) (a float64, b float64) {
	n := float64(len(points))

	var sumX, sumY, sumXY, sumXX = 0.0, 0.0, 0.0, 0.0

	for i, y := range points {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	base := (n*sumXX - sumX*sumX)
	a = (n*sumXY - sumX*sumY) / base
	b = (sumXX*sumY - sumXY*sumX) / base

	return
}
