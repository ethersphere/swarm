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

package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/network"
)

var (
	errUnknown = errors.New("unknown error")
)

type mockRetrieve struct {
	fetchers      map[string]int
	searchTimeout time.Duration
}

func NewMockRetrieve(to time.Duration) *mockRetrieve {
	return &mockRetrieve{fetchers: make(map[string]int), searchTimeout: to}
}

func (m *mockRetrieve) retrieve(rctx Request, f *Fetcher) (err error) {
	log.Warn("mock retrieve called", "addr", f.Address().Hex())
	haddr := hex.EncodeToString(rctx.Address())
	time.Sleep(m.searchTimeout)
	m.fetchers[haddr] += 1
	if m.fetchers[haddr] < 6 {
		return fmt.Errorf("error %d", m.fetchers[haddr])
	}
	go func() {
		f.deliver(NewChunk(rctx.Address(), []byte{0, 1}))
	}()
	return nil
}

func TestNetstoreFailedRequest(t *testing.T) {
	// setup
	searchTimeout := 300 * time.Millisecond
	naddr := network.RandomAddr() // tested peers peer address

	// temp datadir
	datadir, err := ioutil.TempDir("", "netstore")
	if err != nil {
		t.Fatal(err)
	}
	params := NewDefaultLocalStoreParams()
	params.Init(datadir)
	params.BaseKey = naddr.Over()
	localStore, err := NewTestLocalStoreForAddr(params)
	if err != nil {
		t.Fatal(err)
	}

	r := NewMockRetrieve(searchTimeout)
	netStore, err := NewNetStore(localStore, r.retrieve)
	if err != nil {
		t.Fatal(err)
	}

	addr := Address(make([]byte, 32))
	n := 4
	for i := 1; i < n; i++ {
		// timeout := time.Duration(i)*searchTimeout + 20*time.Millisecond
		timeout := time.Duration(i+1)*searchTimeout + 20*time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(i)*searchTimeout+100*time.Millisecond)
		// forward  call
		log.Warn("calling netstore get", "timeout", timeout)
		_, err = Get(ctx, netStore, addr)
		expErr := fmt.Errorf("context deadline exceeded")
		if err.Error() != expErr.Error() {
			t.Fatalf("expected to get %v , but got: %s", expErr, err)
		}
		status, ok := err.(*errStatus)
		if !ok {
			t.Fatalf("expected to get errstatus, got %T", err)
		}
		expErr = fmt.Errorf("error %d", i)
		if status.Status().Error() != expErr.Error() {
			t.Fatalf("expected to get %v , but got: %s", expErr, status.Status())
		}
		if got := r.fetchers[hex.EncodeToString(addr)]; got != i {
			t.Fatalf("expected to have called retrieve %v, but got: %v", i, got)
		}
		log.Warn("testing get", "timeout", timeout, "status", status.Status(), "err", err)
		cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(n)*searchTimeout+100*time.Millisecond)
	defer cancel()
	// third call
	chunk, err := Get(ctx, netStore, addr)
	if got := r.fetchers[hex.EncodeToString(addr)]; got != n {
		t.Fatalf("expected to have called retrieve %v times, but got: %v", n, got)
	}
	if err != nil {
		t.Fatalf("expected to get a chunk but got: %v", err)
	}
	if len(chunk.Data()) != 10 {
		t.Fatalf("expected to get a chunk with size 10, but got: %v", len(chunk.Data()))
	}
}
