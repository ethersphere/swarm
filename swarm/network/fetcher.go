// Copyright 2016 The go-ethereum Authors
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
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var searchTimeout = 3000 * time.Millisecond

// Fetcher is created when a chunk is not found locally
// it starts a request handler loop once and keeps it
// alive until all active requests complete
// either because the chunk is delivered or requestor cancelled/timed out
// fetcher self destroys
// TODO: cancel all forward requests after termination
type Fetcher struct {
	request   func(context.Context, storage.Address, storage.Address, bool, *sync.Map) (context.Context, error) //
	addr      storage.Address                                                                                   //
	offerC    chan storage.Address
	skipCheck bool
}

func FetchFunc(request func(context.Context, storage.Address, storage.Address, bool, *sync.Map) (context.Context, error), skipCheck bool) storage.FetchFuncConstructor {
	return func(ctx context.Context, addr storage.Address, peers *sync.Map) (fetch storage.FetchFunc) {
		f := NewFetcher(addr, request, skipCheck)
		f.start(ctx, peers)
		return f.fetch
	}
}

// NewFetcher creates a new fetcher for a chunk
// stored in netstore's fetchers (LRU cache) keyed by address
func NewFetcher(addr storage.Address, request func(context.Context, storage.Address, storage.Address, bool, *sync.Map) (context.Context, error), skipCheck bool) *Fetcher {
	return &Fetcher{
		addr:      addr,
		request:   request,
		offerC:    make(chan storage.Address),
		skipCheck: skipCheck,
	}
}

// fetch is called by NetStore evey time there is a request or offer for a chunk
func (f *Fetcher) fetch(ctx context.Context) {
	// put offer/request
	var offer storage.Address
	if offerIF := ctx.Value("offer"); offerIF != nil {
		offer = offerIF.(storage.Address)
	}
	select {
	case f.offerC <- offer:
	case <-ctx.Done():
	}
}

// start prepares the Fetcher
// it keeps the Fetcher alive
func (f *Fetcher) start(ctx context.Context, peers *sync.Map) {
	var (
		dorequest bool // determines if retrieval is initiated in the current iteration
		wait      *time.Timer
		waitC     <-chan time.Time
		offers    []storage.Address
		wanted    bool
	)
F:
	// loop that keeps the Fetcher alive
	for {
		// if forwarding is wanted, call netStore

		select {

		// a request or offer
		case offer := <-f.offerC:
			log.Warn("dpa event received")
			if offer != nil {
				// 1) the chunk is offered
				// launch a request to it iff the chunk is requested
				wanted = true
				offers = append(offers, offer)
				dorequest = wanted
			} else {
				// 2) chunk is requested
				// launch a request iff none been launched yet
				dorequest = !wanted
				wanted = true
			}

			// search timeout
		case <-waitC:
			log.Warn(" search timed out rerequesting")
			dorequest = true

			// all Fetcher context closed, can quit
		case <-ctx.Done():
			log.Warn("quitmain loop")
			break F
		}
		if dorequest {
			var fctx context.Context
			var err error
			var i int
			for i = 0; i < len(offers); i++ {
				fctx, err = f.request(ctx, f.addr, offers[i], f.skipCheck, peers)
				if err == nil {
					break
				}
			}
			if i == len(offers) {
				fctx, err = f.request(ctx, f.addr, nil, f.skipCheck, peers)
			}
			if err == nil {
				go func() {
					select {
					case <-fctx.Done():

					case <-ctx.Done():
					}
				}()
			}
			if wait == nil {
				wait = time.NewTimer(searchTimeout)
				defer wait.Stop()
				waitC = wait.C
			}
			wait.Reset(searchTimeout)
			dorequest = false
		}

	}

}
