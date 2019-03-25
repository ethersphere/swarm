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

package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/spancontext"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/sync/singleflight"
)

const (
	// maximum number of forwarded requests (hops), to make sure requests are not
	// forwarded forever in peer loops
	maxHopCount uint8 = 20
)

var requestGroup singleflight.Group

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

var RemoteFetch func(ctx context.Context, ref Address, fi *FetcherItem) error

// FetcherItem are stored in fetchers map and signal to all interested parties if a given chunk is delivered
// the mutex controls who closes the channel, and make sure we close the channel only once
type FetcherItem struct {
	Delivered chan struct{} // when closed, it means that the chunk this FetcherItem refers to is delivered

	// it is possible for multiple actors to be delivering the same chunk,
	// for example through syncing and through retrieve request. however we want the `Delivered` channel to be closed only
	// once, even if we put the same chunk multiple times in the NetStore.
	once sync.Once
}

func NewFetcherItem() *FetcherItem {
	return &FetcherItem{make(chan struct{}), sync.Once{}}
}

func (fi *FetcherItem) SafeClose() {
	fi.once.Do(func() {
		close(fi.Delivered)
	})
}

// NetStore is an extension of LocalStore
// it implements the ChunkStore interface
// on request it initiates remote cloud retrieval
type NetStore struct {
	store    *LocalStore
	fetchers sync.Map
	putMu    sync.Mutex
}

// NewNetStore creates a new NetStore object using the given local store. newFetchFunc is a
// constructor function that can create a fetch function for a specific chunk address.
func NewNetStore(store *LocalStore) *NetStore {
	return &NetStore{
		store:    store,
		fetchers: sync.Map{},
		putMu:    sync.Mutex{},
	}
}

// Put stores a chunk in localstore, and delivers to all requestor peers using the fetcher stored in
// the fetchers cache
func (n *NetStore) Put(ctx context.Context, chunk Chunk) error {
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
		fii := fi.(*FetcherItem)
		fii.SafeClose()
		log.Trace("netstore.put chunk delivered and stored", "ref", chunk.Address().String(), "rid", rid)

		n.fetchers.Delete(chunk.Address().String())
	}

	return nil
}

func (n *NetStore) BinIndex(po uint8) uint64 {
	return n.store.BinIndex(po)
}

func (n *NetStore) Iterator(from uint64, to uint64, po uint8, f func(Address, uint64) bool) error {
	return n.store.Iterator(from, to, po, f)
}

// Close chunk store
func (n *NetStore) Close() {
	n.store.Close()
}

// Get retrieves a chunk
// If it is not found in the LocalStore then it uses RemoteGet to fetch from the network.
func (n *NetStore) Get(ctx context.Context, ref Address) (Chunk, error) {
	metrics.GetOrRegisterCounter("netstore.get", nil).Inc(1)
	rid := getGID()

	log.Trace("netstore.get", "ref", ref.String(), "rid", rid)

	chunk, err := n.store.Get(ctx, ref)
	if err != nil {
		// TODO: fix comparison - we should be comparing against leveldb.ErrNotFound, this error should be wrapped.
		if err != ErrChunkNotFound && err != leveldb.ErrNotFound {
			log.Error("got error from LocalStore other than leveldb.ErrNotFound or ErrChunkNotFound", "err", err)
		}

		var hopCount uint8
		hopCount, _ = ctx.Value("hopCount").(uint8)

		if hopCount >= maxHopCount {
			return nil, fmt.Errorf("reach %v hop counts for ref=%s", maxHopCount, fmt.Sprintf("%x", hopCount))
		}

		log.Trace("netstore.chunk-not-in-localstore", "ref", ref.String(), "hopCount", hopCount, "rid", rid)
		v, err, _ := requestGroup.Do(ref.String(), func() (interface{}, error) {
			has, fi := n.HasWithCallback(ctx, ref)
			if !has {
				err := RemoteFetch(ctx, ref, fi)
				if err != nil {
					return nil, err
				}
			}

			chunk, err := n.store.Get(ctx, ref)
			if err != nil {
				log.Error(err.Error(), "ref", ref, "rid", rid)
				return nil, errors.New("item should have been in localstore, but it is not")
			}

			return chunk, nil
		})

		res, _ := v.(Chunk)

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

// Has is the storage layer entry point to query the underlying
// database to return if it has a chunk or not.
func (n *NetStore) Has(ctx context.Context, ref Address) bool {
	return n.store.Has(ctx, ref)
}

func (n *NetStore) HasWithCallback(ctx context.Context, ref Address) (bool, *FetcherItem) {
	n.putMu.Lock()
	defer n.putMu.Unlock()

	if n.store.Has(ctx, ref) {
		return true, nil
	}

	fi := NewFetcherItem()
	v, loaded := n.fetchers.LoadOrStore(ref.String(), fi)
	log.Trace("netstore.has-with-callback.loadorstore", "ref", ref.String(), "loaded", loaded)
	if loaded {
		fi = v.(*FetcherItem)
	}
	return false, fi
}
