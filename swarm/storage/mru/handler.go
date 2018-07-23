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
	"unsafe"

	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

type Handler struct {
	chunkStore      *storage.NetStore
	HashSize        int
	resources       map[uint64]*resource
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
		resources:       make(map[uint64]*resource),
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
// If it looks like a resource update, the chunk address is checked against the ownerAddr of the update's signature
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
	var r SignedResourceUpdate
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
func (h *Handler) GetContent(viewID *ResourceViewID) (storage.Address, []byte, error) {
	rsrc := h.get(viewID)
	if rsrc == nil {
		return nil, nil, NewError(ErrNotFound, " does not exist")
	}
	return rsrc.lastKey, rsrc.data, nil
}

// GetLastPeriod retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetLastPeriod(viewID *ResourceViewID) (uint32, error) {
	rsrc := h.get(viewID)
	if rsrc == nil {
		return 0, NewError(ErrNotFound, " does not exist")
	}

	return rsrc.period, nil
}

// GetVersion retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetVersion(viewID *ResourceViewID) (uint32, error) {
	rsrc := h.get(viewID)
	if rsrc == nil {
		return 0, NewError(ErrNotFound, " does not exist")
	}
	return rsrc.version, nil
}

/*
// New creates a new metadata chunk out of the request passed in.
func (h *Handler) New(ctx context.Context, request *Request) error {

	// frequency 0 is invalid
	if request.metadata.Frequency == 0 {
		return NewError(ErrInvalidValue, "frequency cannot be 0 when creating a resource")
	}

	// make sure owner is set to something
	if request.metadata.Owner == zeroAddr {
		return NewError(ErrInvalidValue, "ownerAddr must be set to create a new metadata chunk")
	}

	// create the meta chunk and store it in swarm
	chunk, metaHash, err := request.metadata.newChunk()
	if err != nil {
		return err
	}
	if request.metaHash != nil && !bytes.Equal(request.metaHash, metaHash) ||
		request.rootAddr != nil && !bytes.Equal(request.rootAddr, chunk.Addr) {
		return NewError(ErrInvalidValue, "metaHash in UpdateRequest does not match actual metadata")
	}

	request.metaHash = metaHash
	request.rootAddr = chunk.Addr

	h.chunkStore.Put(ctx, chunk)
	log.Debug("new resource", "name", request.metadata.Topic, "startTime", request.metadata.StartTime, "frequency", request.metadata.Frequency, "owner", request.metadata.Owner)

	// create the internal index for the resource and populate it with its metadata
	rsrc := &resource{
		resourceUpdate: resourceUpdate{
			updateHeader: updateHeader{
				UpdateLookup: UpdateLookup{
					rootAddr: chunk.Addr,
				},
			},
		},
		ResourceID: request.metadata,
		updated:    time.Now(),
	}
	h.set(chunk.Addr, rsrc)

	return nil
}
*/

// NewUpdateRequest prepares an UpdateRequest structure with all the necessary information to
// just add the desired data and sign it.
// The resulting structure can then be signed and passed to Handler.Update to be verified and sent
func (h *Handler) NewUpdateRequest(ctx context.Context, viewID *ResourceViewID) (updateRequest *Request, err error) {

	if viewID == nil {
		return nil, NewError(ErrInvalidValue, "viewID cannot be nil")
	}

	now := TimestampProvider.Now()

	updateRequest = new(Request)
	updateRequest.period, err = getNextPeriod(viewID.resourceID.StartTime.Time, now.Time, viewID.resourceID.Frequency)
	if err != nil {
		return nil, err
	}

	// check if there is already an update in this period

	rsrc, err := h.lookup(LookupLatestVersionInPeriod(viewID, updateRequest.period))
	if err != nil {
		if err.(*Error).code != ErrNotFound {
			return nil, err
		}
		// not finding updates means that there is a network error
		// or that the resource really does not have updates in this period.
	}

	updateRequest.viewID = *viewID

	// if we already have an update for this period then increment version
	if rsrc != nil {
		updateRequest.version = rsrc.version + 1
	} else {
		updateRequest.version = 1
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
func (h *Handler) Lookup(ctx context.Context, params *LookupParams) (*resource, error) {
	return h.lookup(params)
}

// LookupPrevious returns the resource before the one currently loaded in the resource cache
// This is useful where resource updates are used incrementally in contrast to
// merely replacing content.
// Requires a cached resource object to determine the current state of the resource.
func (h *Handler) LookupPrevious(ctx context.Context, params *LookupParams) (*resource, error) {
	rsrc := h.get(&params.viewID)
	if rsrc == nil {
		return nil, NewError(ErrNothingToReturn, "resource not loaded")
	}
	var version, period uint32
	if rsrc.version > 1 {
		version = rsrc.version - 1
		period = rsrc.period
	} else if rsrc.period == 1 {
		return nil, NewError(ErrNothingToReturn, "Current update is the oldest")
	} else {
		version = 0
		period = rsrc.period - 1
	}
	return h.lookup(NewLookupParams(&params.viewID, period, version, params.Limit))
}

// base code for public lookup methods
func (h *Handler) lookup(params *LookupParams) (*resource, error) {

	lp := *params
	// we can't look for anything without a store
	if h.chunkStore == nil {
		return nil, NewError(ErrInit, "Call Handler.SetStore() before performing lookups")
	}

	var specificperiod bool
	if lp.period > 0 {
		specificperiod = true
	} else {
		// get the current time and the next period
		now := TimestampProvider.Now()

		var period uint32
		period, err := getNextPeriod(params.viewID.resourceID.StartTime.Time, now.Time, params.viewID.resourceID.Frequency)
		if err != nil {
			return nil, err
		}
		lp.period = period
	}

	// start from the last possible period, and iterate previous ones
	// (unless we want a specific period only) until we find a match.
	// If we hit startTime we're out of options
	var specificversion bool
	if lp.version > 0 {
		specificversion = true
	} else {
		lp.version = 1
	}

	var hops uint32
	if lp.Limit == 0 {
		lp.Limit = h.queryMaxPeriods
	}
	log.Trace("resource lookup", "period", lp.period, "version", lp.version, "limit", lp.Limit)
	for lp.period > 0 {
		if lp.Limit != 0 && hops > lp.Limit {
			return nil, NewErrorf(ErrPeriodDepth, "Lookup exceeded max period hops (%d)", lp.Limit)
		}
		updateAddr := lp.UpdateAddr()
		chunk, err := h.chunkStore.GetWithTimeout(context.TODO(), updateAddr, defaultRetrieveTimeout)
		if err == nil {
			if specificversion {
				return h.updateIndex(&params.viewID, chunk)
			}
			// check if we have versions > 1. If a version fails, the previous version is used and returned.
			log.Trace("rsrc update version 1 found, checking for version updates", "period", lp.period, "updateAddr", updateAddr)
			for {
				newversion := lp.version + 1
				updateAddr := lp.UpdateAddr()
				newchunk, err := h.chunkStore.GetWithTimeout(context.TODO(), updateAddr, defaultRetrieveTimeout)
				if err != nil {
					return h.updateIndex(&params.viewID, chunk)
				}
				chunk = newchunk
				lp.version = newversion
				log.Trace("version update found, checking next", "version", lp.version, "period", lp.period, "updateAddr", updateAddr)
			}
		}
		if specificperiod {
			break
		}
		log.Trace("rsrc update not found, checking previous period", "period", lp.period, "updateAddr", updateAddr)
		lp.period--
		hops++
	}
	return nil, NewError(ErrNotFound, "no updates found")
}

/*
// Load retrieves the Mutable Resource metadata chunk stored at rootAddr
// Upon retrieval it creates/updates the index entry for it with metadata corresponding to the chunk contents
func (h *Handler) Load(ctx context.Context, rootAddr storage.Address) (*resource, error) {
	chunk, err := h.chunkStore.GetWithTimeout(ctx, rootAddr, defaultRetrieveTimeout)
	if err != nil {
		return nil, NewError(ErrNotFound, err.Error())
	}

	// create the index entry
	rsrc := &resource{}

	if err := rsrc.ResourceID.binaryGet(chunk.SData); err != nil { // Will fail if this is not really a metadata chunk
		return nil, err
	}rootAddr

	rsrc.rootAddr, rsrc.metaHash = metadataHash(chunk.SData)
	if !bytes.Equal(rsrc.rootAddr, rootAddr) {
		return nil, NewError(ErrCorruptData, "Corrupt metadata chunk")
	}
	h.set(rootAddr, rsrc)
	log.Trace("resource index load", "rootkey", rootAddr, "name", rsrc.ResourceID.Topic, "starttime", rsrc.ResourceID.StartTime, "frequency", rsrc.ResourceID.Frequency)
	return rsrc, nil
}
*/

// update mutable resource index map with specified content
func (h *Handler) updateIndex(viewID *ResourceViewID, chunk *storage.Chunk) (*resource, error) {

	// retrieve metadata from chunk data and check that it matches this mutable resource
	var r SignedResourceUpdate
	if err := r.fromChunk(chunk.Addr, chunk.SData); err != nil {
		return nil, err
	}
	log.Trace("resource index update", "topic", viewID.resourceID.Topic.Hex(), "updatekey", chunk.Addr, "period", r.period, "version", r.version)

	rsrc := h.get(viewID)
	if rsrc == nil {
		rsrc = &resource{}
		h.set(viewID, rsrc)
	}

	// update our rsrcs entry map
	rsrc.lastKey = chunk.Addr
	rsrc.resourceUpdate = r.resourceUpdate
	rsrc.Reader = bytes.NewReader(rsrc.data)
	return rsrc, nil
}

// Update adds an actual data update
// Uses the Mutable Resource metadata currently loaded in the resources map entry.
// It is the caller's responsibility to make sure that this data is not stale.
// Note that a Mutable Resource update cannot span chunks, and thus has a MAX NET LENGTH 4096, INCLUDING update header data and signature. An error will be returned if the total length of the chunk payload will exceed this limit.
// Update can only check if the caller is trying to overwrite the very last known version, otherwise it just puts the update
// on the network.
func (h *Handler) Update(ctx context.Context, r *SignedResourceUpdate) (storage.Address, error) {
	return h.update(ctx, r)
}

// create and commit an update
func (h *Handler) update(ctx context.Context, r *SignedResourceUpdate) (updateAddr storage.Address, err error) {

	// we can't update anything without a store
	if h.chunkStore == nil {
		return nil, NewError(ErrInit, "Call Handler.SetStore() before updating")
	}

	rsrc := h.get(&r.viewID)
	if rsrc != nil && rsrc.period != 0 && rsrc.version != 0 && // This is the only cheap check we can do for sure
		rsrc.period == r.period && rsrc.version >= r.version {

		return nil, NewError(ErrInvalidValue, "A former update in this period is already known to exist")
	}

	chunk, err := r.toChunk() // Serialize the update into a chunk. Fails if data is too big
	if err != nil {
		return nil, err
	}

	// send the chunk
	h.chunkStore.Put(ctx, chunk)
	log.Trace("resource update", "updateAddr", r.updateAddr, "lastperiod", r.period, "version", r.version, "data", chunk.SData)

	// update our resources map entry if the new update is older than the one we have, if we have it.
	if rsrc != nil && (r.period > rsrc.period || (rsrc.period == r.period && r.version > rsrc.version)) {
		rsrc.period = r.period
		rsrc.version = r.version
		rsrc.data = make([]byte, len(r.data))
		rsrc.lastKey = r.updateAddr
		copy(rsrc.data, r.data)
		rsrc.Reader = bytes.NewReader(rsrc.data)
	}
	return r.updateAddr, nil
}

// Retrieves the resource index value for the given nameHash
func (h *Handler) get(viewID *ResourceViewID) *resource {
	if viewID == nil {
		log.Warn("Handler.get with invalid ViewID")
		return nil
	}
	viewIDKey := viewID.ResourceViewIDAddr()
	hashKey := *(*uint64)(unsafe.Pointer(&viewIDKey[0]))
	h.resourceLock.RLock()
	defer h.resourceLock.RUnlock()
	rsrc := h.resources[hashKey]
	return rsrc
}

// Sets the resource index value for the given nameHash
func (h *Handler) set(viewID *ResourceViewID, rsrc *resource) {
	if viewID == nil {
		log.Warn("Handler.set with invalid ViewID")
		return
	}
	viewIDKey := viewID.ResourceViewIDAddr()
	hashKey := *(*uint64)(unsafe.Pointer(&viewIDKey[0]))
	h.resourceLock.Lock()
	defer h.resourceLock.Unlock()
	h.resources[hashKey] = rsrc
}

// Checks if we already have an update on this resource, according to the value in the current state of the resource index
func (h *Handler) hasUpdate(viewID *ResourceViewID, period uint32) bool {
	rsrc := h.get(viewID)
	return rsrc != nil && rsrc.period == period
}
