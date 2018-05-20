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
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// Fetcher is created when a chunk is not found locally
// it starts a request handler loop once and keeps it
// alive until all active requests complete
// either because the chunk is delivered or requestor cancelled/timed out
// fetcher self destroys
// TODO: cancel all forward requests after termination
type Fetcher struct {
	request func(context.Context) (context.Context, error) //
	addr    storage.Address                                //
	eventC  chan *fetchEvent                               // incoming requests
}

type fetchEvent struct {
	typ      fetchEventType
	offer    storage.Address
	requests []storage.Address
}

func FetchFunc(request func(context.Context) (context.Context, error)) func(ctx context.Context, addr storage.Address) func(context.Context, ...storage.Address) {
	return func(ctx context.Context, addr storage.Address) (fetch func(context.Context, ...storage.Address)) {
		f := NewFetcher(addr, request)
		f.start(ctx)
		return f.fetch
	}
}

// NewFetcher creates a new fetcher for a chunk
// stored in netstore's fetchers (LRU cache) keyed by address
func NewFetcher(addr storage.Address, request func(context.Context) (context.Context, error)) *Fetcher {
	return &Fetcher{
		addr:    addr,
		request: request,
		eventC:  make(chan *fetchEvent),
	}
}

// fetch is called by NetStore evey time there is a request or offer for a chunk
func (f *Fetcher) fetch(ctx context.Context, requests ...storage.Address) {
	// put offer/request
	offer := ctx.Value(Offer)
	ev := &fetchEvent{Offer, offer.(storage.Address), requests}
	select {
	case f.eventC <- ev:
	case <-ctx.Done():
	}
}

// start prepares the Fetcher
// it keeps the Fetcher alive
func (f *Fetcher) start(ctx context.Context) {
	var (
		dorequest bool // determines if retrieval is initiated in the current iteration
		wait      *time.Timer
		waitC     <-chan time.Time
		offers    []storage.Address
	)
F:
	// loop that keeps the Fetcher alive
	for {
		// if forwarding is wanted, call netStore

		select {

		// a request or offer
		case ev := <-f.events:
			log.Warn("dpa event received")
			if ev.offer != nil {
				offers = append(offers, ev.offer)
				dorequest = wanted
			} else {
				wanted = true
				dorequest = true
			}
			requests = ev.requests

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
				fctx, err = f.request(ctx, offers[i], requests)
				if err == nil {
					break
				}
			}
			if i == len(offers) {
				fctx, err = f.request(ctx, nil, requests)
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
			}
			wait.Reset(searchTimeout)
			dorequest = false
			wait := waitC
		}

	}

}
