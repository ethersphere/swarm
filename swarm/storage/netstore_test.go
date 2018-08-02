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
	"crypto/rand"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover"

	"github.com/ethereum/go-ethereum/common"
)

type mockNetFetcher struct {
	peers           *sync.Map
	peersPerRequest [][]Address
	quit            <-chan struct{}
}

func (m *mockNetFetcher) Offer(ctx context.Context, source *discover.NodeID) {

}

func (m *mockNetFetcher) Request(ctx context.Context) {
	var peers []Address
	m.peers.Range(func(key interface{}, _ interface{}) bool {
		peers = append(peers, common.FromHex(key.(string)))
		return true
	})
	m.peersPerRequest = append(m.peersPerRequest, peers)
}

type mockNetFetchFuncFactory struct {
	fetcher *mockNetFetcher
}

func (m *mockNetFetchFuncFactory) newMockNetFetcherFunc(ctx context.Context, _ Address, peers *sync.Map) NetFetcher {
	m.fetcher.peers = peers
	m.fetcher.quit = ctx.Done()
	return m.fetcher
}

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

	fetcher := &mockNetFetcher{}
	mockNetFetchFuncFactory := &mockNetFetchFuncFactory{
		fetcher: fetcher,
	}
	netStore, err := NewNetStore(localStore, mockNetFetchFuncFactory.newMockNetFetcherFunc)
	if err != nil {
		t.Fatal(err)
	}

	addr := randomAddr()
	peers := []string{randomAddr().Hex(), randomAddr().Hex(), randomAddr().Hex()}

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()
	errC := make(chan error)
	nrGets := 3
	for i := 0; i < nrGets; i++ {
		peer := peers[i]
		go func() {
			ctx = context.WithValue(ctx, "peer", peer)
			_, err = netStore.Get(ctx, addr)
			errC <- err
		}()
	}

	expectedErr := "context deadline exceeded"
	for i := 0; i < nrGets; i++ {
		err = <-errC
		if err.Error() != expectedErr {
			t.Fatalf("Expected \"%v\" error got \"%v\"", expectedErr, err)
		}
	}

	select {
	case <-fetcher.quit:
	case <-time.After(3 * time.Second):
		t.Fatalf("mockNetFetcher not closed after timeout")
	}

	if len(fetcher.peersPerRequest) != nrGets {
		t.Fatalf("Expected 3 got %v", len(fetcher.peersPerRequest))
	}

	for i, peers := range fetcher.peersPerRequest {
		if len(peers) < i+1 {
			t.Fatalf("Expected at least %v got %v", i+1, len(peers))
		}
	}

}

func randomAddr() Address {
	addr := make([]byte, 32)
	rand.Read(addr)
	return Address(addr)
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
