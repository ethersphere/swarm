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

// Handler is the API for Mutable Resources
// It enables creating, updating, syncing and retrieving resources and their update data
package mru

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

type Handler struct {
	chunkStore      *storage.NetStore
	HashSize        int
	resources       map[uint64]*cacheEntry
	resourceLock    sync.RWMutex
	storeTimeout    time.Duration
	queryMaxPeriods uint32
}

// HandlerParams pass parameters to the Handler constructor NewHandler
// Signer and TimestampProvider are mandatory parameters
type HandlerParams struct {
	QueryMaxPeriods uint32
}

// hashPool contains a pool of ready hashers
var hashPool sync.Pool

// init initializes the package and hashPool
func init() {
	hashPool = sync.Pool{
		New: func() interface{} {
			return storage.MakeHashFunc(resourceHashAlgorithm)()
		},
	}
}

// NewHandler creates a new Mutable Resource API
func NewHandler(params *HandlerParams) *Handler {
	rh := &Handler{
		resources:       make(map[uint64]*cacheEntry),
		queryMaxPeriods: params.QueryMaxPeriods,
	}

	for i := 0; i < hasherCount; i++ {
		hashfunc := storage.MakeHashFunc(resourceHashAlgorithm)()
		if rh.HashSize == 0 {
			rh.HashSize = hashfunc.Size()
		}
		hashPool.Put(hashfunc)
	}

	return rh
}

// SetStore sets the store backend for the Mutable Resource API
func (h *Handler) SetStore(store *storage.NetStore) {
	h.chunkStore = store
}

// Validate is a chunk validation method
// If it looks like a resource update, the chunk address is checked against the userAddr of the update's signature
// It implements the storage.ChunkValidator interface
func (h *Handler) Validate(chunkAddr storage.Address, data []byte) bool {
	dataLength := len(data)
	if dataLength < minimumSignedUpdateLength {
		return false
	}

	// check if it is a properly formatted update chunk with
	// valid signature and proof of ownership of the resource it is trying
	// to update

	// First, deserialize the chunk
	var r Request
	if err := r.fromChunk(chunkAddr, data); err != nil {
		log.Debug("Invalid resource chunk", "addr", chunkAddr.Hex(), "err", err.Error())
		return false
	}

	// Verify signatures and that the signer actually owns the resource
	// If it fails, it means either the signature is not valid, data is corrupted
	// or someone is trying to update someone else's resource.
	if err := r.Verify(); err != nil {
		log.Debug("Invalid signature", "err", err)
		return false
	}

	return true
}

// GetContent retrieves the data payload of the last synced update of the Mutable Resource
func (h *Handler) GetContent(view *View) (storage.Address, []byte, error) {
	rsrc := h.get(view)
	if rsrc == nil {
		return nil, nil, NewError(ErrNotFound, " does not exist")
	}
	return rsrc.lastKey, rsrc.data, nil
}

// GetLastPeriod retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetLastPeriod(view *View) (uint32, error) {
	rsrc := h.get(view)
	if rsrc == nil {
		return 0, NewError(ErrNotFound, " does not exist")
	}

	return rsrc.Period, nil
}

// GetVersion retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetVersion(view *View) (uint32, error) {
	rsrc := h.get(view)
	if rsrc == nil {
		return 0, NewError(ErrNotFound, " does not exist")
	}
	return rsrc.Version, nil
}

// NewUpdateRequest prepares an UpdateRequest structure with all the necessary information to
// just add the desired data and sign it.
// The resulting structure can then be signed and passed to Handler.Update to be verified and sent
func (h *Handler) NewUpdateRequest(ctx context.Context, view *View) (updateRequest *Request, err error) {

	if view == nil {
		return nil, NewError(ErrInvalidValue, "view cannot be nil")
	}

	now := TimestampProvider.Now()

	updateRequest = new(Request)
	updateRequest.Period, err = getNextPeriod(view.StartTime.Time, now.Time, view.Frequency)
	if err != nil {
		return nil, err
	}

	// check if there is already an update in this period

	rsrc, err := h.lookup(LookupLatestVersionInPeriod(view, updateRequest.Period))
	if err != nil {
		if err.(*Error).code != ErrNotFound {
			return nil, err
		}
		// not finding updates means that there is a network error
		// or that the resource really does not have updates in this period.
	}

	updateRequest.View = *view

	// if we already have an update for this period then increment version
	if rsrc != nil {
		updateRequest.Version = rsrc.Version + 1
	} else {
		updateRequest.Version = 1
	}

	return updateRequest, nil
}

// Lookup retrieves a specific or latest version of the resource update with metadata chunk at params.Root
// Lookup works differently depending on the configuration of `LookupParams`
// See the `LookupParams` documentation and helper functions:
// `LookupLatest`, `LookupLatestVersionInPeriod` and `LookupVersion`
// When looking for the latest update, it starts at the next period after the current time.
// upon failure tries the corresponding keys of each previous period until one is found
// (or startTime is reached, in which case there are no updates).
func (h *Handler) Lookup(ctx context.Context, params *LookupParams) (*cacheEntry, error) {
	return h.lookup(params)
}

// LookupPrevious returns the resource before the one currently loaded in the resource cache
// This is useful where resource updates are used incrementally in contrast to
// merely replacing content.
// Requires a cached resource object to determine the current state of the resource.
func (h *Handler) LookupPrevious(ctx context.Context, params *LookupParams) (*cacheEntry, error) {
	rsrc := h.get(&params.View)
	if rsrc == nil {
		return nil, NewError(ErrNothingToReturn, "resource not loaded")
	}
	var version, period uint32
	if rsrc.Version > 1 {
		version = rsrc.Version - 1
		period = rsrc.Period
	} else if rsrc.Period == 1 {
		return nil, NewError(ErrNothingToReturn, "Current update is the oldest")
	} else {
		version = 0
		period = rsrc.Period - 1
	}
	return h.lookup(NewLookupParams(&params.View, period, version, params.Limit))
}

// base code for public lookup methods
func (h *Handler) lookup(params *LookupParams) (*cacheEntry, error) {

	lp := *params
	// we can't look for anything without a store
	if h.chunkStore == nil {
		return nil, NewError(ErrInit, "Call Handler.SetStore() before performing lookups")
	}

	var specificperiod bool
	if lp.Period > 0 {
		specificperiod = true
	} else {
		// get the current time and the next period
		now := TimestampProvider.Now()

		var period uint32
		period, err := getNextPeriod(params.View.StartTime.Time, now.Time, params.View.Frequency)
		if err != nil {
			return nil, err
		}
		lp.Period = period
	}

	// start from the last possible period, and iterate previous ones
	// (unless we want a specific period only) until we find a match.
	// If we hit startTime we're out of options
	var specificversion bool
	if lp.Version > 0 {
		specificversion = true
	} else {
		lp.Version = 1
	}

	var hops uint32
	if lp.Limit == 0 {
		lp.Limit = h.queryMaxPeriods
	}
	log.Trace("resource lookup", "period", lp.Period, "version", lp.Version, "limit", lp.Limit)
	for lp.Period > 0 {
		if lp.Limit != 0 && hops > lp.Limit {
			return nil, NewErrorf(ErrPeriodDepth, "Lookup exceeded max period hops (%d)", lp.Limit)
		}
		updateAddr := lp.UpdateAddr()
		chunk, err := h.chunkStore.GetWithTimeout(context.TODO(), updateAddr, defaultRetrieveTimeout)
		if err == nil {
			if specificversion {
				return h.updateCache(&params.View, chunk)
			}
			// check if we have versions > 1. If a version fails, the previous version is used and returned.
			log.Trace("rsrc update version 1 found, checking for version updates", "period", lp.Period, "updateAddr", updateAddr)
			for {
				newversion := lp.Version + 1
				updateAddr := lp.UpdateAddr()
				newchunk, err := h.chunkStore.GetWithTimeout(context.TODO(), updateAddr, defaultRetrieveTimeout)
				if err != nil {
					return h.updateCache(&params.View, chunk)
				}
				chunk = newchunk
				lp.Version = newversion
				log.Trace("version update found, checking next", "version", lp.Version, "period", lp.Period, "updateAddr", updateAddr)
			}
		}
		if specificperiod {
			break
		}
		log.Trace("rsrc update not found, checking previous period", "period", lp.Period, "updateAddr", updateAddr)
		lp.Period--
		hops++
	}
	return nil, NewError(ErrNotFound, "no updates found")
}

// update mutable resource cache map with specified content
func (h *Handler) updateCache(view *View, chunk *storage.Chunk) (*cacheEntry, error) {

	// retrieve metadata from chunk data and check that it matches this mutable resource
	var r Request
	if err := r.fromChunk(chunk.Addr, chunk.SData); err != nil {
		return nil, err
	}
	log.Trace("resource cache update", "topic", view.Topic.Hex(), "updatekey", chunk.Addr, "period", r.Period, "version", r.Version)

	rsrc := h.get(view)
	if rsrc == nil {
		rsrc = &cacheEntry{}
		h.set(view, rsrc)
	}

	// update our rsrcs entry map
	rsrc.lastKey = chunk.Addr
	rsrc.ResourceUpdate = r.ResourceUpdate
	rsrc.Reader = bytes.NewReader(rsrc.data)
	return rsrc, nil
}

// Update adds an actual data update
// Uses the Mutable Resource metadata currently loaded in the resources map entry.
// It is the caller's responsibility to make sure that this data is not stale.
// Note that a Mutable Resource update cannot span chunks, and thus has a MAX NET LENGTH 4096, INCLUDING update header data and signature. An error will be returned if the total length of the chunk payload will exceed this limit.
// Update can only check if the caller is trying to overwrite the very last known version, otherwise it just puts the update
// on the network.
func (h *Handler) Update(ctx context.Context, r *Request) (storage.Address, error) {
	return h.update(ctx, r)
}

// create and commit an update
func (h *Handler) update(ctx context.Context, r *Request) (updateAddr storage.Address, err error) {

	// we can't update anything without a store
	if h.chunkStore == nil {
		return nil, NewError(ErrInit, "Call Handler.SetStore() before updating")
	}

	rsrc := h.get(&r.View)
	if rsrc != nil && rsrc.Period != 0 && rsrc.Version != 0 && // This is the only cheap check we can do for sure
		rsrc.Period == r.Period && rsrc.Version >= r.Version {

		return nil, NewError(ErrInvalidValue, "A former update in this period is already known to exist")
	}

	chunk, err := r.toChunk() // Serialize the update into a chunk. Fails if data is too big
	if err != nil {
		return nil, err
	}

	// send the chunk
	h.chunkStore.Put(ctx, chunk)
	log.Trace("resource update", "updateAddr", r.updateAddr, "lastperiod", r.Period, "version", r.Version, "data", chunk.SData)

	// update our resources map cache entry if the new update is older than the one we have, if we have it.
	if rsrc != nil && (r.Period > rsrc.Period || (rsrc.Period == r.Period && r.Version > rsrc.Version)) {
		rsrc.Period = r.Period
		rsrc.Version = r.Version
		rsrc.data = make([]byte, len(r.data))
		rsrc.lastKey = r.updateAddr
		copy(rsrc.data, r.data)
		rsrc.Reader = bytes.NewReader(rsrc.data)
	}
	return r.updateAddr, nil
}

// Retrieves the resource cache value for the given nameHash
func (h *Handler) get(view *View) *cacheEntry {
	if view == nil {
		log.Warn("Handler.get with invalid View")
		return nil
	}
	mapKey := view.mapKey()
	h.resourceLock.RLock()
	defer h.resourceLock.RUnlock()
	rsrc := h.resources[mapKey]
	return rsrc
}

// Sets the resource cache value for the given View
func (h *Handler) set(view *View, rsrc *cacheEntry) {
	if view == nil {
		log.Warn("Handler.set with invalid View")
		return
	}
	mapKey := view.mapKey()
	h.resourceLock.Lock()
	defer h.resourceLock.Unlock()
	h.resources[mapKey] = rsrc
}

// Checks if we already have an update on this resource, according to the value in the current state of the resource cache
func (h *Handler) hasUpdate(view *View, period uint32) bool {
	rsrc := h.get(view)
	return rsrc != nil && rsrc.Period == period
}
