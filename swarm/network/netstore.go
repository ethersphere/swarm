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
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/network/timeouts"
	"github.com/ethereum/go-ethereum/swarm/spancontext"
	"github.com/ethereum/go-ethereum/swarm/storage"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/sync/singleflight"
)

const (
	// maximum number of forwarded requests (hops), to make sure requests are not
	// forwarded forever in peer loops
	maxHopCount uint8 = 10
)

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// FetcherItem are stored in fetchers map and signal to all interested parties if a given chunk is delivered
// the mutex controls who closes the channel, and make sure we close the channel only once
type FetcherItem struct {
	Delivered chan struct{} // when closed, it means that the chunk this FetcherItem refers to is delivered

	// it is possible for multiple actors to be delivering the same chunk,
	// for example through syncing and through retrieve request. however we want the `Delivered` channel to be closed only
	// once, even if we put the same chunk multiple times in the NetStore.
	once sync.Once

	CreatedAt time.Time // timestamp when the fetcher was created, used for metrics measuring lifetime of fetchers
	CreatedBy string    // who created the fethcer - "request" or "syncing", used for metrics measuring lifecycle of fetchers
}

func NewFetcherItem() *FetcherItem {
	return &FetcherItem{make(chan struct{}), sync.Once{}, time.Now(), ""}
}

func (fi *FetcherItem) SafeClose() {
	fi.once.Do(func() {
		close(fi.Delivered)
	})
}

type RemoteGetFunc func(ctx context.Context, req *Request, localID enode.ID) (*enode.ID, error)

// NetStore is an extension of LocalStore
// it implements the ChunkStore interface
// on request it initiates remote cloud retrieval
type NetStore struct {
	localID      enode.ID // our local enode - used when issuing RetrieveRequests
	store        *storage.LocalStore
	fetchers     sync.Map
	putMu        sync.Mutex
	requestGroup singleflight.Group
	//RemoteGet    func(ctx context.Context, req *Request, localID enode.ID) (*enode.ID, error)
	RemoteGet RemoteGetFunc
}

// NewNetStore creates a new NetStore object using the given local store. newFetchFunc is a
// constructor function that can create a fetch function for a specific chunk address.
func NewNetStore(store *storage.LocalStore, localID enode.ID) *NetStore {
	return &NetStore{
		localID:      localID,
		store:        store,
		fetchers:     sync.Map{},
		putMu:        sync.Mutex{},
		requestGroup: singleflight.Group{},
	}
}

// Put stores a chunk in localstore, and delivers to all requestor peers using the fetcher stored in
// the fetchers cache
func (n *NetStore) Put(ctx context.Context, chunk storage.Chunk) error {
	n.putMu.Lock()
	defer n.putMu.Unlock()

	rid := getGID()
	log.Trace("netstore.put", "ref", chunk.Address().String(), "rid", rid)

	// put the chunk to the localstore, there should be no error
	err := n.store.Put(ctx, chunk)
	if err != nil {
		return err
	}

	// TODO: probably safe to put this in a go-routine
	// notify RemoteGet about a chunk being stored
	fi, ok := n.fetchers.Load(chunk.Address().String())
	if ok {
		// we need SafeClose, because it is possible for a chunk to both be
		// delivered through syncing and through a retrieve request
		fii, ok := fi.(*FetcherItem)
		if !ok {
			panic("loaded item from n.fetchers is not *FetcherItem")
		}
		if fii == nil {
			panic("fii is nil")
		}
		fii.SafeClose()
		log.Trace("netstore.put chunk delivered and stored", "ref", chunk.Address().String(), "rid", rid)

		metrics.GetOrRegisterResettingTimer(fmt.Sprintf("netstore.fetcher.lifetime.%s", fii.CreatedBy), nil).UpdateSince(fii.CreatedAt)

		if time.Since(fii.CreatedAt) > 5*time.Second {
			log.Trace("netstore.put slow chunk delivery", "ref", chunk.Address().String(), "rid", rid)
		}

		n.fetchers.Delete(chunk.Address().String())
	}

	return nil
}

func (n *NetStore) BinIndex(po uint8) uint64 {
	return n.store.BinIndex(po)
}

func (n *NetStore) Iterator(from uint64, to uint64, po uint8, f func(storage.Address, uint64) bool) error {
	return n.store.Iterator(from, to, po, f)
}

// Close chunk store
func (n *NetStore) Close() {
	n.store.Close()
}

// Get retrieves a chunk
// If it is not found in the LocalStore then it uses RemoteGet to fetch from the network.
func (n *NetStore) Get(ctx context.Context, req *Request) (storage.Chunk, error) {
	metrics.GetOrRegisterCounter("netstore.get", nil).Inc(1)
	start := time.Now()

	ref := req.Addr
	rid := getGID()

	log.Trace("netstore.get", "ref", ref.String(), "rid", rid)

	chunk, err := n.store.Get(ctx, ref)
	if err != nil {
		// TODO: fix comparison - we should be comparing against leveldb.ErrNotFound, this error should be wrapped.
		if err != storage.ErrChunkNotFound && err != leveldb.ErrNotFound {
			log.Error("got error from LocalStore other than leveldb.ErrNotFound or ErrChunkNotFound", "err", err)
		}

		if req.HopCount >= maxHopCount {
			return nil, fmt.Errorf("reach %v hop counts for ref=%s", maxHopCount, fmt.Sprintf("%x", req.HopCount))
		}

		log.Trace("netstore.chunk-not-in-localstore", "ref", ref.String(), "hopCount", req.HopCount, "rid", rid)
		v, err, _ := n.requestGroup.Do(ref.String(), func() (interface{}, error) {
			//TODO: decide if we want to issue a retrieve request if a fetcher
			// has already been created by a syncer for that particular chunk.
			// for now we issue a retrieve request, so it is possible to
			// have 2 in-flight requests for the same chunk - one by a
			// syncer (offered/wanted/deliver flow) and one from
			// here - retrieve request!
			has, fi, _ := n.HasWithCallback(ctx, ref, "request")
			if !has {
				err := n.RemoteFetch(ctx, req, fi)
				if err != nil {
					return nil, err
				}
			}

			chunk, err := n.store.Get(ctx, ref)
			if err != nil {
				log.Error(err.Error(), "ref", ref, "rid", rid)
				return nil, errors.New("item should have been in localstore, but it is not")
			}

			// fi could be nil if the chunk was added to the NetStore inbetween n.store.Get and the call to n.HasWithCallback
			if fi != nil {
				metrics.GetOrRegisterResettingTimer(fmt.Sprintf("fetcher.%s.request", fi.CreatedBy), nil).UpdateSince(start)
			}

			return chunk, nil
		})

		res, _ := v.(storage.Chunk)

		log.Trace("netstore.singleflight returned", "ref", ref.String(), "err", err, "rid", rid)

		if err != nil {
			log.Error(err.Error(), "ref", ref, "rid", rid)
			return nil, err
		}

		log.Trace("netstore return", "ref", ref.String(), "chunk len", len(res.Data()), "rid", rid)

		return res, nil
	}

	ctx, ssp := spancontext.StartSpan(
		ctx,
		"localstore.get")
	defer ssp.Finish()

	return chunk, nil
}

func (n *NetStore) RemoteFetch(ctx context.Context, req *Request, fi *FetcherItem) error {
	// while we haven't timed-out, and while we don't have a chunk,
	// iterate over peers and try to find a chunk
	metrics.GetOrRegisterCounter("remote.fetch", nil).Inc(1)

	ref := req.Addr

	rid := getGID()

	for {
		metrics.GetOrRegisterCounter("remote.fetch.inner", nil).Inc(1)

		innerCtx, osp := spancontext.StartSpan(
			ctx,
			"remote.fetch")
		osp.LogFields(olog.String("ref", ref.String()))

		log.Trace("remote.fetch", "ref", ref, "rid", rid)
		currentPeer, err := n.RemoteGet(innerCtx, req, n.localID)
		if err != nil {
			log.Error(err.Error(), "ref", ref, "rid", rid)
			osp.LogFields(olog.String("err", err.Error()))
			osp.Finish()
			return err
		}
		osp.LogFields(olog.String("peer", currentPeer.String()))

		// add peer to the set of peers to skip from now
		log.Trace("remote.fetch, adding peer to skip", "ref", ref, "peer", currentPeer.String(), "rid", rid)
		req.PeersToSkip.Store(currentPeer.String(), time.Now())

		select {
		case <-fi.Delivered:
			log.Trace("remote.fetch, chunk delivered", "ref", ref, "rid", rid)

			osp.LogFields(olog.Bool("delivered", true))
			osp.Finish()
			return nil
		case <-time.After(timeouts.SearchTimeout):
			metrics.GetOrRegisterCounter("remote.fetch.timeout.search", nil).Inc(1)

			osp.LogFields(olog.Bool("timeout", true))
			osp.Finish()
			break
		case <-ctx.Done(): // global fetcher timeout
			log.Trace("remote.fetch, fail", "ref", ref, "rid", rid)
			metrics.GetOrRegisterCounter("remote.fetch.timeout.global", nil).Inc(1)

			osp.LogFields(olog.Bool("fail", true))
			osp.Finish()
			return errors.New("chunk couldnt be retrieved from remote nodes")
		}
	}
}

// Has is the storage layer entry point to query the underlying
// database to return if it has a chunk or not.
func (n *NetStore) Has(ctx context.Context, ref storage.Address) bool {
	return n.store.Has(ctx, ref)
}

func (n *NetStore) HasWithCallback(ctx context.Context, ref storage.Address, interestedParty string) (bool, *FetcherItem, bool) {
	n.putMu.Lock()
	defer n.putMu.Unlock()

	if n.store.Has(ctx, ref) {
		return true, nil, false
	}

	fi := NewFetcherItem()
	v, loaded := n.fetchers.LoadOrStore(ref.String(), fi)
	log.Trace("netstore.has-with-callback.loadorstore", "ref", ref.String(), "loaded", loaded)
	if loaded {
		var ok bool
		fi, ok = v.(*FetcherItem)
		if !ok {
			panic("loaded item from n.fetchers is not *FetcherItem")
		}
	} else {
		fi.CreatedBy = interestedParty
	}
	if fi == nil {
		panic("fi is nil")
	}
	return false, fi, loaded
}
