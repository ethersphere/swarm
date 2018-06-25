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
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var searchTimeout = 3000 * time.Millisecond

type RequestFunc func(context.Context, *Request) (*discover.NodeID, chan struct{}, error)

// Fetcher is created when a chunk is not found locally
// it starts a request handler loop once and keeps it
// alive until all active requests complete
// either because the chunk is delivered or requestor cancelled/timed out
// fetcher self destroys
// NetStore instantiates the
// TODO: cancel all forward requests after termination
type Fetcher struct {
	request   RequestFunc           // request function fetcher calls to issue retrieve request for a chunk
	addr      storage.Address       // the address of the chunk to be fetched
	sourceC   chan *discover.NodeID // channel of sources (peer node id strings)
	skipCheck bool
}

type Request struct {
	Addr        storage.Address  // chunk address
	Source      *discover.NodeID // nodeID of peer to request from (can be nil)
	SkipCheck   bool             // whether to offer the chunk first or deliver directly
	PeersToSkip *sync.Map        // peers not to request chunk from (only makes sense if source is nil)
}

// FetcherFactory is initialised with a request function and can create fetchers
type FetcherFactory struct {
	request   RequestFunc
	skipCheck bool
}

// NewFetcherFactory takes a request function and skip check parameter
func NewFetcherFactory(request RequestFunc, skipCheck bool) *FetcherFactory {
	return &FetcherFactory{
		request:   request,
		skipCheck: skipCheck,
	}
}

// New contructs a new fetcher, starts it and returns a fetch function
// to be called by the NetStore when handling requests and offers
func (f *FetcherFactory) New(ctx context.Context, source storage.Address, peersToSkip *sync.Map) storage.FetchFunc {
	fetcher := NewFetcher(source, f.request, f.skipCheck)
	go fetcher.run(ctx, peersToSkip)
	return fetcher.fetch
}

// NewFetcher creates a new fetcher for a chunk
// stored in netstore's fetchers (LRU cache) keyed by address
func NewFetcher(addr storage.Address, request RequestFunc, skipCheck bool) *Fetcher {
	return &Fetcher{
		addr:      addr,
		request:   request,
		sourceC:   make(chan *discover.NodeID),
		skipCheck: skipCheck,
	}
}

// fetch is called by NetStore evey time there is a request or offer for a chunk
func (f *Fetcher) fetch(ctx context.Context) {
	// put source/request
	var source *discover.NodeID
	if sourceIF := ctx.Value("source"); sourceIF != nil {
		id := discover.MustHexID(sourceIF.(string))
		source = &id
	}
	select {
	case f.sourceC <- source:
	case <-ctx.Done():
	}
}

// start prepares the Fetcher
// it keeps the Fetcher alive within the lifecycle of the passed context
func (f *Fetcher) run(ctx context.Context, peers *sync.Map) {
	var (
		doRequest bool               // determines if retrieval is initiated in the current iteration
		wait      *time.Timer        // timer for search timeout
		waitC     <-chan time.Time   // timer channel
		sources   []*discover.NodeID // known sources, ie. peers that offered the chunk
		requested bool               // true if the chunk was actually requested
	)
	gone := make(chan *discover.NodeID) // channel to signal that a peer we requested from disconnected

	// loop that keeps the fetching process alive
	// after every request a timer is set. If this goes off we request again from another peer
	// note that the previous request is still alive and has the chance to deliver, so
	// rerequesting extends the search. ie.,
	// if a peer we requested from is gone we issue a new request, so the number of active
	// requests never decreases
	for {
		select {

		// accept a request or source.
		case source := <-f.sourceC:
			if source != nil {
				log.Debug("new source", "peer addr", source, "request addr", f.addr)
				// 1) the chunk is offered by a syncing peer
				// add to known sources
				requested = true
				sources = append(sources, source)
				// launch a request to the source iff the chunk was requested (not just expected because its offered by a syncing peer)
				doRequest = requested
			} else {
				log.Debug("new request", "request addr", f.addr)
				// 2) chunk is requested, set requested flag
				// launch a request iff none been launched yet
				doRequest = !requested
				requested = true
			}

			// peer we requested from is gone. fall back to another
			// and remove the peer from the peers map
		case id := <-gone:
			log.Debug("peer gone", "peer id", id.String(), "request addr", f.addr)
			peers.Delete(id.String())
			doRequest = true

		// search timeout: too much time passed since the last request,
		// extend the search to a new peer if we can find one
		case <-waitC:
			log.Debug("search timed out: rerequesting", "request addr", f.addr)
			doRequest = true

			// all Fetcher context closed, can quit
		case <-ctx.Done():
			log.Debug("terminate fetcher", "request addr", f.addr)
			// TODO: send cancelations to all peers left over in peers map (i.e., those we requested from)
			return
		}

		// need to issue a new request
		if doRequest {
			var err error
			sources, err = f.doRequest(ctx, gone, peers, sources)
			if err != nil {
				log.Debug("unable to request", "request addr", f.addr, "err", err)
			}
		}

		// if wait channel is not set, set it to a timer
		if wait == nil {
			wait = time.NewTimer(searchTimeout)
			defer wait.Stop()
			waitC = wait.C
		}
		// reset the timer to go off after searchTimeout
		wait.Reset(searchTimeout)
		doRequest = false
	}
}

// doRequest attempts at finding a peer to request the chunk from
// * first it tries to request explicitly from peers that are known to have offered the chunk
// * if there are no such peers (available) it tries to request it from a peer closest to the chunk address
//   excluding those in the peersToSkip map
// * if no such peer is found an error is returned
//
// if a request is successful,
// * the peer's address is added to the set of peers to skip
// * the peer's address is removed from prospective sources, and
// * a go routine is started that reports on the gone channel if the peer is disconnected (or terminated their streamer)
func (f *Fetcher) doRequest(ctx context.Context, gone chan *discover.NodeID, peersToSkip *sync.Map, sources []*discover.NodeID) ([]*discover.NodeID, error) {
	var i int
	var sourceID *discover.NodeID
	var quit chan struct{}
	var err error

	req := &Request{
		Addr:        f.addr,
		SkipCheck:   f.skipCheck,
		PeersToSkip: peersToSkip,
	}
	// iterate over known sources
	for i = 0; i < len(sources); i++ {
		req.Source = sources[i]
		sourceID, quit, err = f.request(ctx, req)
		if err == nil {
			// remove the peer from known sources
			// Note: we can modify the source although we are looping on it, because we break from the loop immediately
			sources = append(sources[:i], sources[i+1:]...)
			break
		}
	}

	// if there are no known sources, or none available, we try request from a closest node
	if i == len(sources) {
		req.Source = nil
		sourceID, quit, err = f.request(ctx, req)
		if err != nil {
			// if no peers found to request from
			return sources, err
		}
	}
	// add peer to the set of peers to skip from now
	peersToSkip.Store(sourceID.String(), true)

	// if the quit channel is closed, it indicates that the source peer we requested from
	// disconnected or terminated its streamer
	// here start a go routine that watches this channel and reports the source peer on the gone channel
	// this go routine quits if the fetcher global context is done to prevent process leak
	go func() {
		select {
		case <-quit:
			gone <- sourceID
		case <-ctx.Done():
		}
	}()
	return sources, nil
}
