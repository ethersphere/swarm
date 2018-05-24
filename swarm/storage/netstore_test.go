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
	"errors"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

var (
	errUnknown = errors.New("unknown error")
)

// type mockRetrieve struct {
// 	fetchers      map[string]int
// 	searchTimeout time.Duration
// 	contextC      chan context.Context
// 	errC          chan error
// }

type mockFetcher struct {
	peersPerFetch [][]Address
}

func (m *mockFetcher) fetch(ctx context.Context) {
	// m.peersPerFetch = append(m.peersPerFetch, peers)
}

func newMockFetcher() *mockFetcher {
	return &mockFetcher{}
}

func (m *mockFetcher) mockFetch(_ context.Context, _ Address, _ *sync.Map) FetchFunc {
	return m.fetch
}

// func (m *mockRetrieve) retrieve(rctx Request, f *Fetcher) (context.Context, error) {
// 	ctx := <-m.contextC
// 	err := <-m.errC
// 	return ctx, err
// }

// func (m *mockRetrieve) feed(ctx context.Context, err error) {
// 	go func() {
// 		m.contextC <- ctx
// 		m.errC <- err
// 	}()
// }

// type mockRetrieveContext struct {
// 	err   error
// 	doneC chan struct{}
// }

// func NewMockFailedRetrieveContext(duration time.Time) *mockRetrieveContext {
// 	doneC := make(chan struct{})
// 	timer := time.NewTimer(duration)
// 	go func() {
// 		<-timer.C
// 		close(doneC)
// 	}
// 	return &mockRetrieveContext{
// 		doneC : doneC,
// 		err: errors.New("retrieve aborted"),
// 	}
// }

// func NewMockRetrieveContext() *mockRetrieveContext {
// 	return &mockRetrieveContext{
// 		doneC : make(chan struct{}),
// 	}
// }

func TestNetStoreFetcherCountPeers(t *testing.T) {
	// setup
	searchTimeout := 500 * time.Millisecond
	naddr := make([]byte, 32)

	// temp datadir
	datadir, err := ioutil.TempDir("", "netstore")
	if err != nil {
		t.Fatal(err)
	}
	params := NewDefaultLocalStoreParams()
	params.Init(datadir)
	params.BaseKey = naddr
	localStore, err := NewTestLocalStoreForAddr(params)
	if err != nil {
		t.Fatal(err)
	}

	fetcher := newMockFetcher()
	netStore, err := NewNetStore(localStore, fetcher.mockFetch)
	if err != nil {
		t.Fatal(err)
	}

	addr := Address(make([]byte, 32))
	ctx, _ := context.WithTimeout(context.Background(), searchTimeout)

	netStore.Get(ctx, addr)
	netStore.Get(ctx, addr)
	netStore.Get(ctx, addr)

	if len(fetcher.peersPerFetch) != 3 {
		t.Fatal()
	}

}

// func TestNetstoreRepeatedFailedRequest(t *testing.T) {
// 	// setup
// 	searchTimeout := 500 * time.Millisecond
// 	naddr := network.RandomAddr()

// 	// temp datadir
// 	datadir, err := ioutil.TempDir("", "netstore")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	params := NewDefaultLocalStoreParams()
// 	params.Init(datadir)
// 	params.BaseKey = naddr.Over()
// 	localStore, err := NewTestLocalStoreForAddr(params)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	r := NewMockRetrieve(searchTimeout)
// 	netStore, err := NewNetStore(localStore, r.retrieve)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	addr := Address(make([]byte, 32))
// 	n := 4
// 	for i := 1; i < n; i++ {
// 		log.Warn("\n\nIteration", "i", i)
// 		timeout := time.Duration(i)*searchTimeout + 300*time.Millisecond
// 		ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 		log.Warn("calling netstore get", "timeout", timeout)
// 		rctx := &localRequest{ctx, addr}
// 		_, err = netStore.Get(rctx, addr)
// 		time.Sleep(100 * time.Millisecond)
// 		// check the error
// 		expErr := fmt.Errorf("context deadline exceeded")
// 		if err == nil || err.Error() != expErr.Error() {
// 			t.Fatalf("expected to get %v , but got: %v", expErr, err)
// 		}

// 		// check retrieve status
// 		status, ok := err.(*errStatus)
// 		if !ok {
// 			t.Fatalf("expected to get errstatus, got %T", err)
// 		}
// 		expErr = fmt.Errorf("error %d", i)
// 		if status.Status() == nil || status.Status().Error() != expErr.Error() {
// 			t.Fatalf("expected to get %v , but got: %v", expErr, status.Status())
// 		}

// 		// check how many times retrieve is called
// 		if got := r.fetchers[hex.EncodeToString(addr)]; got != i {
// 			t.Fatalf("expected to have called retrieve %v, but got: %v", i, got)
// 		}
// 		log.Warn("testing get", "timeout", timeout, "status", status.Status(), "err", err)
// 		cancel()
// 	}

// 	// check if eventually the chunk arrives
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(n)*searchTimeout+15*time.Millisecond)
// 	defer cancel()
// 	rctx := &localRequest{ctx, addr}
// 	ch, err := netStore.Get(rctx, addr)
// 	if got := r.fetchers[hex.EncodeToString(addr)]; got != n {
// 		t.Fatalf("expected to have called retrieve %v times, but got: %v", n, got)
// 	}
// 	if err != nil {
// 		t.Fatalf("expected to get a chunk but got: %v", err)
// 	}
// 	if len(ch.Data()) != 2 {
// 		t.Fatalf("expected to get a chunk with size 10, but got: %v", len(ch.Data()))
// 	}
// }
