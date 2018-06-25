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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover"
)

var sourcePeerID = discover.MustHexID("1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439")

// mockRequester pushes every request to the requestC channel when its doRequest function is called
type mockRequester struct {
	// requests []Request
	requestC  chan *Request
	waitTimes []time.Duration
	ctr       int
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
	source := request.Source
	if source == nil {
		fmt.Println("source nil")
		source = &sourcePeerID
	}
	return source, make(chan struct{}), nil
}

// This test creates a Fetcher using mockRequester, and run it with a sample set of peers to skip.
// mockRequester pushes a Request on a channel every time the request function is called. Using
// this channel we test if calling Fetcher.fetch calls the request function, and whether it provides
// to correct peers to skip which we provided for the fetcher.run function.
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
		for _, p := range peers {
			if _, ok := request.PeersToSkip.Load(p); !ok {
				t.Fatalf("request.peersToSkip misses peer")
			}
		}
		// wait for the source peer to be added to peersToSkip
		time.Sleep(100 * time.Millisecond)
		if _, ok := request.PeersToSkip.Load(sourcePeerID.String()); !ok {
			t.Fatalf("request.peersToSkip does not contain peer returned by the request function")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("fetch timeout")
	}
}

// Test that a cancelled fetcher does not initiate further requests even if its fetch function is called
func TestCancelStopsFetcher(t *testing.T) {
	requester := newMockRequester()
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peersToSkip := &sync.Map{}

	ctx, cancel := context.WithCancel(context.Background())

	go fetcher.run(ctx, peersToSkip)

	cancel()
	rctx, rcancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer rcancel()
	fetcher.fetch(rctx)

	select {
	case <-requester.requestC:
		t.Fatalf("cancelled fetcher initiated request")
	case <-time.After(200 * time.Millisecond):
	}
}

// Test if calling a fetch function with a cancelled context does not initiate a request
func TestFetchCancelStopsFetch(t *testing.T) {
	requester := newMockRequester(100 * time.Millisecond)
	addr := make([]byte, 32)
	fetcher := NewFetcher(addr, requester.doRequest, true)

	peersToSkip := &sync.Map{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go fetcher.run(ctx, peersToSkip)

	rctx, rcancel := context.WithCancel(context.Background())
	rcancel()
	fetcher.fetch(rctx)

	select {
	case <-requester.requestC:
		t.Fatalf("cancelled fetcher initiated request")
	case <-time.After(200 * time.Millisecond):
	}
}
