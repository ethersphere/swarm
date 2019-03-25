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
	"bytes"
	"context"
	"errors"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/spancontext"
	"github.com/ethereum/go-ethereum/swarm/storage"
	olog "github.com/opentracing/opentracing-go/log"
)

// FailedPeerSkipDelay is the time we consider a peer to be skipped for a particular request/chunk,
// because this peer failed to deliver it during the SearchTimeout interval
var FailedPeerSkipDelay = 10 * time.Second

// RequestTimeout is the max time for which we try to find a chunk while handling a retrieve request
var RequestTimeout = 10 * time.Second

// FetcherTimeout is the max time a node tries to find a chunk for a client, after which it returns a 404
// Basically this is the amount of time a singleflight request for a given chunk lives
var FetcherTimeout = 10 * time.Second

// SearchTimeout is the max time we wait for a peer to deliver a chunk we requests, after which we try another peer
var SearchTimeout = 1 * time.Second

var RemoteGet func(ctx context.Context, req *Request) (*enode.ID, error)

type Request struct {
	Addr        storage.Address // chunk address
	PeersToSkip sync.Map        // peers not to request chunk from (only makes sense if source is nil)
	HopCount    uint8           // number of forwarded requests (hops)
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func RemoteFetch(ctx context.Context, ref storage.Address, fi *storage.FetcherItem) error {
	// while we haven't timed-out, and while we don't have a chunk,
	// iterate over peers and try to find a chunk
	metrics.GetOrRegisterCounter("remote.fetch", nil).Inc(1)
	gt := time.After(FetcherTimeout)

	var hopCount uint8
	hopCount, _ = ctx.Value("hopCount").(uint8)

	req := NewRequest(ref, hopCount)
	rid := getGID()

	for {
		metrics.GetOrRegisterCounter("remote.fetch.inner", nil).Inc(1)

		innerCtx, osp := spancontext.StartSpan(
			ctx,
			"remote.fetch")
		osp.LogFields(olog.String("ref", ref.String()))

		log.Trace("remote.fetch", "ref", ref, "rid", rid)
		currentPeer, err := RemoteGet(innerCtx, req)
		if err != nil {
			log.Error(err.Error(), "ref", ref, "rid", rid)
			osp.Finish()
			return err
		}

		// add peer to the set of peers to skip from now
		log.Trace("remote.fetch, adding peer to skip", "ref", ref, "peer", currentPeer.String(), "rid", rid)
		req.PeersToSkip.Store(currentPeer.String(), time.Now())

		select {
		case <-fi.Delivered:
			log.Trace("remote.fetch, chunk delivered", "ref", ref, "rid", rid)

			osp.LogFields(olog.Bool("delivered", true))
			osp.Finish()
			return nil
		case <-time.After(SearchTimeout):
			osp.LogFields(olog.Bool("timeout", true))
			osp.Finish()
			break
		case <-gt:
			log.Trace("remote.fetch, fail", "ref", ref, "rid", rid)

			osp.LogFields(olog.Bool("fail", true))
			osp.Finish()
			return errors.New("chunk couldnt be retrieved from remote nodes")
		}
	}
}

// NewRequest returns a new instance of Request based on chunk address skip check and
// a map of peers to skip.
func NewRequest(addr storage.Address, hopCount uint8) *Request {
	return &Request{
		Addr:        addr,
		HopCount:    hopCount,
		PeersToSkip: sync.Map{},
	}
}

// SkipPeer returns if the peer with nodeID should not be requested to deliver a chunk.
// Peers to skip are kept per Request and for a time period of FailedPeerSkipDelay.
func (r *Request) SkipPeer(nodeID string) bool {
	val, ok := r.PeersToSkip.Load(nodeID)
	if !ok {
		return false
	}
	t, ok := val.(time.Time)
	if ok && time.Now().After(t.Add(FailedPeerSkipDelay)) {
		r.PeersToSkip.Delete(nodeID)
		return false
	}
	return true
}
