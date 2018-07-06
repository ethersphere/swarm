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

package network

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover"
)

var requestedPeerID = discover.MustHexID("1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439")
var sourcePeerID = discover.MustHexID("2dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439")

// var sourcePeerID2 = discover.MustHexID("3dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439")
// var sourcePeerID3 = discover.MustHexID("4dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439")

// mockRequester pushes every request to the requestC channel when its doRequest function is called
type mockRequester struct {
	// requests []Request
	requestC  chan *Request   // when a request is coming it is pushed to requestC
	waitTimes []time.Duration // with waitTimes[i] you can define how much to wait on the ith request (optional)
	ctr       int             //counts the number of requests
}

func newMockRequester(waitTimes ...time.Duration) *mockRequester {
	return &mockRequester{
		requestC:  make(chan *Request),
		waitTimes: waitTimes,
	}
}

func (m *mockRequester) doRequest(ctx context.Context, request *Request) (*discover.NodeID, chan struct{}, error) {
	waitTime := time.Duration(0)
	if m.ctr < len(m.waitTimes) {
		waitTime = m.waitTimes[m.ctr]
		m.ctr++
	}
	time.Sleep(waitTime)
	m.requestC <- request

	// if there is a Source in the request use that, if not use the global requestedPeerId
	source := request.Source
	if source == nil {
		source = &requestedPeerID
	}
	return source, make(chan struct{}), nil
}

// TestFetcherSingleFetch creates a Fetcher using mockRequester, and run it with a sample set of peers to skip.
// mockRequester pushes a Request on a channel every time the request function is called. Using
// this channel we test if calling Fetcher.fetch calls the request function, and whether it uses
// the correct peers to skip which we provided for the fetcher.run function.
func TestFetcherSingleFetch(t *testing.T) {
	requester := newMockRequester()
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peers := []string{"a", "b", "c", "d"}
	peersToSkip := &sync.Map{}
	for _, p := range peers {
		peersToSkip.Store(p, true)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go fetcher.run(ctx, peersToSkip)

	rctx := context.Background()
	fetcher.fetch(rctx)

	select {
	case request := <-requester.requestC:
		// request should contain all peers from peersToSkip provided to the fetcher
		for _, p := range peers {
			if _, ok := request.PeersToSkip.Load(p); !ok {
				t.Fatalf("request.peersToSkip misses peer")
			}
		}

		// source peer should be also added to peersToSkip eventually
		time.Sleep(100 * time.Millisecond)
		if _, ok := request.PeersToSkip.Load(requestedPeerID.String()); !ok {
			t.Fatalf("request.peersToSkip does not contain peer returned by the request function")
		}

		// fetch should trigger a request, if it doesn't happen in time, test should fail
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("fetch timeout")
	}
}

// TestCancelStopsFetcher tests that a cancelled fetcher does not initiate further requests even if its fetch function is called
func TestCancelStopsFetcher(t *testing.T) {
	requester := newMockRequester()
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peersToSkip := &sync.Map{}

	ctx, cancel := context.WithCancel(context.Background())

	// we start the fetcher, and then we immediately cancel the context
	go fetcher.run(ctx, peersToSkip)
	cancel()

	rctx, rcancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer rcancel()
	// we call fetch with an active context
	fetcher.fetch(rctx)

	// fetcher should not initiate request, we can only check by waiting a bit and making sure no request is happening
	select {
	case <-requester.requestC:
		t.Fatalf("cancelled fetcher initiated request")
	case <-time.After(200 * time.Millisecond):
	}
}

// TestFetchCancelStopsFetch tests that calling a fetch function with a cancelled context does not initiate a request
func TestFetchCancelStopsFetch(t *testing.T) {
	requester := newMockRequester(100 * time.Millisecond)
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peersToSkip := &sync.Map{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// we start the fetcher with an active context
	go fetcher.run(ctx, peersToSkip)

	rctx, rcancel := context.WithCancel(context.Background())
	rcancel()

	// we call fetch with a cancelled context
	fetcher.fetch(rctx)

	// fetcher should not initiate request, we can only check by waiting a bit and making sure no request is happening
	select {
	case <-requester.requestC:
		t.Fatalf("cancelled fetch function initiated request")
	case <-time.After(200 * time.Millisecond):
	}
}

// TestFetchUsesSourceFromContext tests Fetcher request behavior when there is a source in the context.
// In this case there should be 1 (and only one) request initiated from the source peer, and the
// source nodeid from the context should appear in the peersToSkip map.
func TestFetchUsesSourceFromContext(t *testing.T) {
	requester := newMockRequester(100 * time.Millisecond)
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peersToSkip := &sync.Map{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start the fetcher
	go fetcher.run(ctx, peersToSkip)

	// set the source peer on the context and call the fetch function with it
	rctx := context.WithValue(context.Background(), "source", sourcePeerID.String())
	fetcher.fetch(rctx)

	// fetcher should not initiate request
	select {
	case <-requester.requestC:
		t.Fatalf("fetcher initiated request")
	case <-time.After(200 * time.Millisecond):
	}

	// call fetch again with a context without source
	rctx = context.Background()
	fetcher.fetch(rctx)

	// there should be exactly 1 request coming from fetcher
	var request *Request
	select {
	case request = <-requester.requestC:
		if *request.Source != sourcePeerID {
			t.Fatalf("Expected source id %v got %v", sourcePeerID, request.Source)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("fetcher did not initiate request")
	}

	select {
	case <-requester.requestC:
		t.Fatalf("Fetcher number of requests expected 1 got 2")
	case <-time.After(200 * time.Millisecond):
	}

	// source peer should be added to peersToSkip eventually
	time.Sleep(100 * time.Millisecond)
	if _, ok := request.PeersToSkip.Load(sourcePeerID.String()); !ok {
		t.Fatalf("SourcePeerId not added to peersToSkip")
	}
}

func TestSecondFetchUsesSourceFromContext(t *testing.T) {
	requester := newMockRequester(100 * time.Millisecond)
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peersToSkip := &sync.Map{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start the fetcher
	go fetcher.run(ctx, peersToSkip)

	// call fetch again with a context without source
	rctx := context.Background()
	fetcher.fetch(rctx)

	// there should be a request coming from fetcher
	var request *Request
	select {
	case request = <-requester.requestC:
		if request.Source != nil {
			t.Fatalf("Incorrect source peer id, expected nil got %v", request.Source)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("fetcher did not initiate request")
	}

	// set the source peer on the context and call the fetch function with it
	rctx = context.WithValue(context.Background(), "source", sourcePeerID.String())
	fetcher.fetch(rctx)

	// there should be a request coming from fetcher
	select {
	case request = <-requester.requestC:
		if *request.Source != sourcePeerID {
			t.Fatalf("Incorrect source peer id, expected %v got %v", sourcePeerID, request.Source)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("fetcher did not initiate request")
	}

	// source peer should be added to peersToSkip eventually
	time.Sleep(100 * time.Millisecond)
	if _, ok := request.PeersToSkip.Load(sourcePeerID.String()); !ok {
		t.Fatalf("SourcePeerId not added to peersToSkip")
	}
}

// TestFetcherRetryOnTimeout tests that fetch retries after searchTimeOut has passed
func TestFetcherRetryOnTimeout(t *testing.T) {
	requester := newMockRequester()
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peersToSkip := &sync.Map{}

	// set searchTimeOut to low value so the test is quicker
	searchTimeout = 250 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start the fetcher
	go fetcher.run(ctx, peersToSkip)

	// call the fetch function with an active context
	rctx := context.Background()
	fetcher.fetch(rctx)

	// after 100ms the first request should be initiated
	time.Sleep(100 * time.Millisecond)

	select {
	case <-requester.requestC:
	default:
		t.Fatalf("fetch did not initiate request")
	}

	// after another 100ms no new request should be initiated, because search timeout is 250ms
	time.Sleep(100 * time.Millisecond)

	select {
	case <-requester.requestC:
		t.Fatalf("unexpected request from fetcher")
	default:
	}

	// after another 200ms search timeout is over, there should be a new request
	time.Sleep(200 * time.Millisecond)

	select {
	case <-requester.requestC:
	default:
		t.Fatalf("fetch did not retry request")
	}
}

// TestFetcherFactory creates a FetcherFactory and checks if the factory really creates and starts
// a Fetcher when it return a fetch function. We test the fetching functionality just by checking if
// a request is initiated when the fetch function is called
func TestFetcherFactory(t *testing.T) {
	requester := newMockRequester(100 * time.Millisecond)
	addr := make([]byte, 32)
	fetcherFactory := NewFetcherFactory(requester.doRequest, false)

	peersToSkip := &sync.Map{}

	fetchFunction := fetcherFactory.New(context.Background(), addr, peersToSkip)

	fetchFunction(context.Background())

	// check if the created fetchFunction really starts a fetcher and initiates a request
	select {
	case <-requester.requestC:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("fetch timeout")
	}

}
