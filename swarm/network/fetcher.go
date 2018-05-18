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
	"errors"
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
	retrieve func(context.Context) (context.Context, error)
	addr     storage.Address      //
	requestC chan storage.Request // incoming requests
	stoppedC chan struct{}        // closed when terminates
	// requestsWg sync.WaitGroup // wait group on requests
	status error // fetching status
}

func NewFetcherConstructor(retrieve func(storage.Request) (context.Context, error)) func(addr storage.Address, remove func()) *Fetcher {
	return func(addr storage.Address) *Fetcher {
		return NewFetcher(addr, retrieve)
	}
}

// NewFetcher creates a new fetcher for a chunk
// stored in netstore's fetchers (LRU cache) keyed by address
func NewFetcher(addr storage.Address, retrieve func(storage.Request) (context.Context, error)) *Fetcher {
	f := &Fetcher{
		addr:     addr,
		retrieve: retrieve,
		requestC: make(chan storage.Request),
		stoppedC: make(chan struct{}),
	}
	go f.start()
	return f
}

// fetch is a synchronous fetcher function returned
// by NetStore.Get and called if remote fetching is required
func (f *Fetcher) Fetch(rctx storage.Request) (Chunk, error) {
	// put rctx on request channel
	select {
	case f.requestC <- rctx:
	case <-f.stoppedC:
	}
}

// // wait waits till all actual Requests are closed
// func (f *Fetcher) wait() {
// 	<-f.quitC
// 	// remove the Fetcher from the cache when all Fetchers
// 	// contexts closed, self-destruct and remove from fetchers
// }

// // stop stops the Fetcher
// func (f *Fetcher) stop() {
// 	close(f.stoppedC)
// }

// start prepares the Fetcher
// it keeps the Fetcher alive
func (f *Fetcher) start() {
	var (
		doRetrieve bool               // determines if retrieval is initiated in the current iteration
		forwardC   = make(chan error) // timeout/error on one of the forwards
		rctx       Request
	)
F:
	// loop that keeps the Fetcher alive
	for {
		// if forwarding is wanted, call netStore
		if doRetrieve {
			go func() {
				ctx, err := f.retrieve(rctx)
				if err != nil {
					select {
					case forwardC <- err:
						log.Warn("forward result", "err", err)
					case <-f.stoppedC:
					}
				} else {
					go func() {
						timer := time.NewTimer(searchTimeout)
						var err error
						select {
						case <-ctx.Done():
							err = ctx.Err()
						case <-timer.C:
							err = errors.New("search timed out")
						case <-f.stoppedC:
							return
						}
						select {
						case forwardC <- err:
						case <-f.stoppedC:
						}
					}()
				}
			}()
			doRetrieve = false
		}

		select {

		// ready to receive by accept in a request ()
		case rctx = <-f.requestC:
			log.Warn("upstream request received")
			doRetrieve = true

		// rerequest upon forwardC event
		case err := <-forwardC: // if forward request completes
			log.Warn("forward request failed", "err", err)
			f.status = err
			doRetrieve = err != nil

		case <-f.stoppedC:
			log.Warn("quitmain loop")
			// all Fetcher context closed, can quit
			break F
		}
	}

}

func (f *Fetcher) Stop() {
	close(f.stoppedC)
}

func (f *Fetcher) Status() {
	return status
}
