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
	"bytes"
	"context"
	"crypto/rand"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover"
	ch "github.com/ethereum/go-ethereum/swarm/chunk"

	"github.com/ethereum/go-ethereum/common"
)

var sourcePeerID = discover.MustHexID("2dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439")

type mockNetFetcher struct {
	peers           *sync.Map
	sources         []*discover.NodeID
	peersPerRequest [][]Address
	requestCalled   bool
	offerCalled     bool
	quit            <-chan struct{}
	ctx             context.Context
}

func (m *mockNetFetcher) Offer(ctx context.Context, source *discover.NodeID) {
	m.offerCalled = true
	m.sources = append(m.sources, source)
}

func (m *mockNetFetcher) Request(ctx context.Context) {
	m.requestCalled = true
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

func (m *mockNetFetchFuncFactory) newMockNetFetcher(ctx context.Context, _ Address, peers *sync.Map) NetFetcher {
	m.fetcher.peers = peers
	m.fetcher.quit = ctx.Done()
	m.fetcher.ctx = ctx
	return m.fetcher
}

func mustNewNetStore(t *testing.T) *NetStore {
	netStore, _ := mustNewNetStoreWithFetcher(t)
	return netStore
}

func mustNewNetStoreWithFetcher(t *testing.T) (*NetStore, *mockNetFetcher) {
	t.Helper()

	datadir, err := ioutil.TempDir("", "netstore")
	if err != nil {
		t.Fatal(err)
	}
	naddr := make([]byte, 32)
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
	netStore, err := NewNetStore(localStore, mockNetFetchFuncFactory.newMockNetFetcher)
	if err != nil {
		t.Fatal(err)
	}
	return netStore, fetcher
}

// TestNetStoreGet tests
func TestNetStoreGetAndPut(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	c := make(chan struct{})
	go func() {
		<-c
		time.Sleep(200 * time.Millisecond)
		err := netStore.Put(ctx, chunk)
		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}
	}()

	close(c)
	recChunk, err := netStore.Get(ctx, chunk.Address())
	if err != nil {
		t.Fatalf("Expected no err got %v", err)
	}
	if !bytes.Equal(recChunk.Address(), chunk.Address()) || !bytes.Equal(recChunk.Data(), chunk.Data()) {
		t.Fatalf("Different chunk received than what was put")
	}
	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to remove the fetcher after delivery")
	}

	select {
	case <-fetcher.ctx.Done():
	default:
		t.Fatal("Expected fetcher context to be cancelled")
	}

}

func TestNetStoreGetAfterPut(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := netStore.Put(ctx, chunk)
	if err != nil {
		t.Fatalf("Expected no err got %v", err)
	}

	recChunk, err := netStore.Get(ctx, chunk.Address())
	if err != nil {
		t.Fatalf("Expected no err got %v", err)
	}
	if !bytes.Equal(recChunk.Address(), chunk.Address()) || !bytes.Equal(recChunk.Data(), chunk.Data()) {
		t.Fatalf("Different chunk received than what was put")
	}
	if fetcher.offerCalled || fetcher.requestCalled {
		t.Fatal("NetFetcher.offerCalled or requestCalled not expected to be called")
	}
	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to not have fetcher")
	}

}

func TestNetStoreGetTimeout(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := netStore.Get(ctx, chunk.Address())
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected context.DeadLineExceeded err got %v", err)
	}

	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to remove the fetcher after timeout")
	}

	select {
	case <-fetcher.ctx.Done():
	default:
		t.Fatal("Expected fetcher context to be cancelled")
	}
}

func TestNetStoreGetCancel(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)

	go cancel()

	_, err := netStore.Get(ctx, chunk.Address())

	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled err got %v", err)
	}

	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to remove the fetcher after cancel")
	}

	select {
	case <-fetcher.ctx.Done():
	default:
		t.Fatal("Expected fetcher context to be cancelled")
	}
}

func TestNetStoreMultipleGetAndPut(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		// sleep to make sure Put is called after all the Get
		time.Sleep(500 * time.Millisecond)
		err := netStore.Put(ctx, chunk)
		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}
	}()

	getWG := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		getWG.Add(1)
		go func() {
			defer getWG.Done()
			recChunk, err := netStore.Get(ctx, chunk.Address())
			if err != nil {
				t.Fatalf("Expected no err got %v", err)
			}
			if !bytes.Equal(recChunk.Address(), chunk.Address()) || !bytes.Equal(recChunk.Data(), chunk.Data()) {
				t.Fatalf("Different chunk received than what was put")
			}
		}()
	}

	finishedC := make(chan struct{})
	go func() {
		getWG.Wait()
		close(finishedC)
	}()
	select {
	case <-finishedC:
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout waiting for Get calls to return")
	}

	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to remove the fetcher after delivery")
	}

	select {
	case <-fetcher.ctx.Done():
	default:
		t.Fatal("Expected fetcher context to be cancelled")
	}

}

func TestNetStoreFetchFuncTimeout(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	wait := netStore.FetchFunc(ctx, chunk.Address())
	if wait == nil {
		t.Fatal("Expected wait function to be not nil")
	}

	if netStore.fetchers.Len() != 1 || netStore.getFetcher(chunk.Address()) == nil {
		t.Fatalf("Expected netStore to have one fetcher for the requested chunk")
	}

	err := wait(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected context.DeadLineExceeded err got %v", err)
	}

	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to remove the fetcher after timeout")
	}

	select {
	case <-fetcher.ctx.Done():
	default:
		t.Fatal("Expected fetcher context to be cancelled")
	}
}

func TestNetStoreFetchFuncAfterPut(t *testing.T) {
	netStore := mustNewNetStore(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := netStore.Put(ctx, chunk)
	if err != nil {
		t.Fatalf("Expected no err got %v", err)
	}

	wait := netStore.FetchFunc(ctx, chunk.Address())
	if wait != nil {
		t.Fatal("Expected wait to be nil")
	}

	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to not have fetcher")
	}
}

func TestNetStoreGetCallsRequest(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)

	go cancel()

	_, err := netStore.Get(ctx, chunk.Address())

	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled err got %v", err)
	}

	if !fetcher.requestCalled {
		t.Fatal("Expected NetFetcher.Request to be called")
	}
}

func TestNetStoreGetCallsOffer(t *testing.T) {
	netStore := mustNewNetStore(t)

	fetcher := &mockNetFetcher{}
	netStore.NewNetFetcherFunc = (&mockNetFetchFuncFactory{
		fetcher: fetcher,
	}).newMockNetFetcher

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx := context.WithValue(context.Background(), "source", sourcePeerID.String())
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	chunk, err := netStore.Get(ctx, chunk.Address())

	if err != context.DeadlineExceeded {
		t.Fatalf("Expect error %v got %v", context.DeadlineExceeded, err)
	}

	if !fetcher.offerCalled {
		t.Fatal("Expected NetFetcher.Request to be called")
	}

	if len(fetcher.sources) != 1 {
		t.Fatalf("Expected fetcher sources length 1 got %v", len(fetcher.sources))
	}

	if fetcher.sources[0].String() != sourcePeerID.String() {
		t.Fatalf("Expected fetcher source %v got %v", sourcePeerID, fetcher.sources[0])
	}

}

// TestNetStoreFetcherCountPeers tests multiple NetStore.Get calls with peer in the context.
// There is no Put call, so the Get calls timeout
func TestNetStoreFetcherCountPeers(t *testing.T) {

	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	addr := randomAddr()
	peers := []string{randomAddr().Hex(), randomAddr().Hex(), randomAddr().Hex()}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	errC := make(chan error)
	nrGets := 3
	for i := 0; i < nrGets; i++ {
		peer := peers[i]
		go func() {
			ctx = context.WithValue(ctx, "peer", peer)
			_, err := netStore.Get(ctx, addr)
			errC <- err
		}()
	}

	expectedErr := "context deadline exceeded"
	for i := 0; i < nrGets; i++ {
		err := <-errC
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

func TestNetStoreFetcherLifeCycle(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	wait := netStore.FetchFunc(ctx, chunk.Address())
	if wait == nil {
		t.Fatal("Expected wait function to be not nil")
	}

	if netStore.fetchers.Len() != 1 || netStore.getFetcher(chunk.Address()) == nil {
		t.Fatalf("Expected netStore to have one fetcher for the requested chunk")
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			err := wait(ctx)
			if err != nil {
				t.Fatalf("Expected no err got %v", err)
			}
			wg.Done()
		}()
	}

	// sleep a little so the wait functions are called above
	time.Sleep(100 * time.Millisecond)

	if netStore.fetchers.Len() != 1 || netStore.getFetcher(chunk.Address()) == nil {
		t.Fatal("Expected netStore to have one fetcher for the requested chunk")
	}

	err := netStore.Put(ctx, chunk)
	if err != nil {
		t.Fatalf("Expected no err got %v", err)
	}

	wg.Wait()

	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to remove the fetcher after delivery")
	}

	select {
	case <-fetcher.ctx.Done():
	default:
		t.Fatal("Expected fetcher context to be cancelled")
	}
}

func TestNetStoreFetcherLifeCycleWithTimeout(t *testing.T) {
	netStore, fetcher := mustNewNetStoreWithFetcher(t)

	chunk := GenerateRandomChunk(ch.DefaultSize)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	wait := netStore.FetchFunc(ctx, chunk.Address())
	if wait == nil {
		t.Fatal("Expected wait function to be not nil")
	}

	if netStore.fetchers.Len() != 1 || netStore.getFetcher(chunk.Address()) == nil {
		t.Fatalf("Expected netStore to have one fetcher for the requested chunk")
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rctx, rcancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer rcancel()
			err := wait(rctx)
			if err != context.DeadlineExceeded {
				t.Fatalf("Expected err %v got %v", context.DeadlineExceeded, err)
			}
		}()
	}

	wg.Wait()

	if netStore.fetchers.Len() != 0 {
		t.Fatal("Expected netStore to remove the fetcher after delivery")
	}

	select {
	case <-fetcher.ctx.Done():
	default:
		t.Fatal("Expected fetcher context to be cancelled")
	}
}

func randomAddr() Address {
	addr := make([]byte, 32)
	rand.Read(addr)
	return Address(addr)
}
