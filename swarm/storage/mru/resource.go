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

package mru

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/net/idna"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/multihash"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const (
	metadataChunkOffsetSize = 16 + common.AddressLength
	chunkSize               = 4096 // temporary until we implement FileStore in the resourcehandler
	defaultStoreTimeout     = 4000 * time.Millisecond
	hasherCount             = 8
	resourceHashAlgorithm   = storage.SHA3Hash
	defaultRetrieveTimeout  = 100 * time.Millisecond
)

type blockEstimator struct {
	Start   time.Time
	Average time.Duration
}

// NewBlockEstimator returns an object that can be used for retrieving an heuristical block height in the absence of a blockchain connection
// It implements the headerGetter interface
// TODO: Average must  be adjusted when blockchain connection is present and synced
func NewBlockEstimator() *blockEstimator {
	sampleDate, _ := time.Parse(time.RFC3339, "2018-05-04T20:35:22Z")   // from etherscan.io
	sampleBlock := int64(3169691)                                       // from etherscan.io
	ropstenStart, _ := time.Parse(time.RFC3339, "2016-11-20T11:48:50Z") // from etherscan.io
	ns := sampleDate.Sub(ropstenStart).Nanoseconds()
	period := int(ns / sampleBlock)
	parsestring := fmt.Sprintf("%dns", int(float64(period)*1.0005)) // increase the blockcount a little, so we don't overshoot the read block height; if we do, we will never find the updates when getting synced data
	periodNs, _ := time.ParseDuration(parsestring)
	return &blockEstimator{
		Start:   ropstenStart,
		Average: periodNs,
	}
}

// HeaderByNumber retrieves the estimated block number wrapped in a block header struct
func (b *blockEstimator) HeaderByNumber(context.Context, string, *big.Int) (*types.Header, error) {
	return &types.Header{
		Number: big.NewInt(time.Since(b.Start).Nanoseconds() / b.Average.Nanoseconds()),
	}, nil
}

// Error is a the typed error object used for Mutable Resources
type Error struct {
	code int
	err  string
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.err
}

// Code returns the error code
// Error codes are enumerated in the error.go file within the mru package
func (e *Error) Code() int {
	return e.code
}

// NewError creates a new Mutable Resource Error object with the specified code and custom error message
func NewError(code int, s string) error {
	if code < 0 || code >= ErrCnt {
		panic("no such error code!")
	}
	r := &Error{
		err: s,
	}
	switch code {
	case ErrNotFound, ErrIO, ErrUnauthorized, ErrInvalidValue, ErrDataOverflow, ErrNothingToReturn, ErrInvalidSignature, ErrNotSynced, ErrPeriodDepth, ErrCorruptData:
		r.code = code
	}
	return r
}

// LookupParams is used to specify constraints when performing an update lookup
// Limit defines whether or not the lookup should be limited
// If Limit is set to true then Max defines the amount of hops that can be performed
// \TODO this is redundant, just use uint32 with 0 for unlimited hops
type LookupParams struct {
	Period  uint32
	Version uint32
	Root    storage.Address
	Limit   uint32
}

type resourceData struct {
	version   uint32
	period    uint32
	multihash bool
	metaHash  []byte
	rootAddr  storage.Address
	data      []byte
}

// Caches resource data. When synced it contains the most recent
// version of the resource and its metadata.
type resource struct {
	resourceData
	resourceMetadata
	*bytes.Reader
	lastKey storage.Address
	updated time.Time
}

func (r *resource) Context() context.Context {
	return context.TODO()
}

// TODO Expire content after a defined period (to force resync)
func (r *resource) isSynced() bool {
	return !r.updated.IsZero()
}

//Whether the resource data should be interpreted as multihash
func (r *resourceData) Multihash() bool {
	return r.multihash
}

// implements (which?) interface
func (r *resource) Size(ctx context.Context, _ chan bool) (int64, error) {
	if !r.isSynced() {
		return 0, NewError(ErrNotSynced, "Not synced")
	}
	return int64(len(r.resourceData.data)), nil
}

//returns the resource's human-readable name
func (r *resource) Name() string {
	return r.name
}

// Handler is the API for Mutable Resources
// It enables creating, updating, syncing and retrieving resources and their update data
type Handler struct {
	chunkStore        *storage.NetStore
	HashSize          int
	timestampProvider timestampProvider
	resources         map[uint64]*resource
	resourceLock      sync.RWMutex
	storeTimeout      time.Duration
	queryMaxPeriods   uint32
}

// HandlerParams pass parameters to the Handler constructor NewHandler
// Signer and TimestampProvider are mandatory parameters
type HandlerParams struct {
	QueryMaxPeriods   uint32
	TimestampProvider timestampProvider
}

var hashPool sync.Pool

func init() {
	hashPool = sync.Pool{
		New: func() interface{} {
			return storage.MakeHashFunc(resourceHashAlgorithm)()
		},
	}
}

// NewHandler creates a new Mutable Resource API
func NewHandler(params *HandlerParams) (*Handler, error) {

	rh := &Handler{
		timestampProvider: params.TimestampProvider,
		resources:         make(map[uint64]*resource),
		storeTimeout:      defaultStoreTimeout,
		queryMaxPeriods:   params.QueryMaxPeriods,
	}

	if rh.timestampProvider == nil {
		rh.timestampProvider = NewDefaultTimestampProvider()
	}

	for i := 0; i < hasherCount; i++ {
		hashfunc := storage.MakeHashFunc(resourceHashAlgorithm)()
		if rh.HashSize == 0 {
			rh.HashSize = hashfunc.Size()
		}
		hashPool.Put(hashfunc)
	}

	return rh, nil
}

// SetStore sets the store backend for the Mutable Resource API
func (h *Handler) SetStore(store *storage.NetStore) {
	h.chunkStore = store
}

// Validate is a chunk validation method
// If it's a resource update, the chunk address is checked against the public updateAddr of the update's signature
// It implements the storage.ChunkValidator interface
func (h *Handler) Validate(addr storage.Address, data []byte) bool {

	if data[0] == 0 && data[1] == 0 && len(data) > common.AddressLength {
		//metadata chunk
		rootAddr, _ := metadataHash(data)
		valid := bytes.Equal(addr, rootAddr)
		if !valid {
			log.Warn("Invalid root metadata chunk")
		}
		return valid
	}

	// does the chunk data make sense?
	// This is not 100% safe evaluation, since content addressed data can coincidentally produce data with length header matching content size
	_, err := parseUpdate(addr, data)
	if err != nil {
		log.Warn("Invalid resource chunk")
		return false
	}

	return true
}

// GetContent retrieves the data payload of the last synced update of the Mutable Resource
func (h *Handler) GetContent(addr storage.Address) (storage.Address, []byte, error) {
	rsrc := h.get(addr)
	if rsrc == nil || !rsrc.isSynced() {
		return nil, nil, NewError(ErrNotFound, " does not exist or is not synced")
	}
	return rsrc.lastKey, rsrc.data, nil
}

// GetLastPeriod retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetLastPeriod(addr storage.Address) (uint32, error) {
	rsrc := h.get(addr)
	if rsrc == nil {
		return 0, NewError(ErrNotFound, " does not exist")
	} else if !rsrc.isSynced() {
		return 0, NewError(ErrNotSynced, " is not synced")
	}
	return rsrc.period, nil
}

// GetVersion retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetVersion(addr storage.Address) (uint32, error) {
	rsrc := h.get(addr)
	if rsrc == nil {
		return 0, NewError(ErrNotFound, " does not exist")
	} else if !rsrc.isSynced() {
		return 0, NewError(ErrNotSynced, " is not synced")
	}
	return rsrc.version, nil
}

// \TODO should be hashsize * branches from the chosen chunker, implement with FileStore
func (h *Handler) chunkSize() int64 {
	return chunkSize
}

// New creates a new metadata chunk for a Mutable Resource identified by `name` with the specified `frequency`.
// The start block of the resource update will be the actual current block height of the connected network.
func (h *Handler) New(ctx context.Context, request *UpdateRequest) error {

	// frequency 0 is invalid
	if request.frequency == 0 {
		return NewError(ErrInvalidValue, "frequency cannot be 0")
	}

	// make sure name only contains ascii values
	if !isSafeName(request.name) {
		return NewError(ErrInvalidValue, fmt.Sprintf("invalid name: '%s'", request.name))
	}

	// make sure owner is set
	var zeroAddr = common.Address{}
	if request.ownerAddr == zeroAddr {
		return NewError(ErrInvalidValue, "ownerAddr must be set to create a new metadata chunk")
	}

	// get the current time
	if request.startTime == 0 {
		request.startTime = h.getCurrentTime(ctx)
	}
	// create the meta chunk and store it in swarm

	chunk, metaHash := h.newMetaChunk(&request.resourceMetadata)
	h.chunkStore.Put(ctx, chunk)

	request.metaHash = metaHash
	request.rootAddr = chunk.Addr
	request.period = 1
	request.version = 1

	log.Debug("new resource", "name", request.name, "startBlock", request.startTime, "frequency", request.frequency, "owner", request.ownerAddr)

	// create the internal index for the resource and populate it with the data of the first version
	rsrc := &resource{
		resourceData: resourceData{
			rootAddr: chunk.Addr,
		},
		resourceMetadata: resourceMetadata{
			name:      request.name,
			startTime: request.startTime,
			frequency: request.frequency,
		},
		updated: time.Now(),
	}
	copy(rsrc.ownerAddr[:], request.ownerAddr[:])
	h.set(chunk.Addr, rsrc)

	return nil
}

func (h *Handler) NewUpdateRequest(ctx context.Context, rootAddr storage.Address) (*UpdateRequest, error) {

	if rootAddr == nil {
		return nil, NewError(ErrInvalidValue, "rootAddr cannot be nil")
	}

	rsrc, err := h.Load(ctx, rootAddr)
	if err != nil {
		return nil, err
	}

	currentblock := h.getCurrentTime(ctx)

	updateRequest := new(UpdateRequest)
	updateRequest.period, err = getNextPeriod(rsrc.startTime, currentblock, rsrc.frequency)
	if err != nil {
		return nil, err
	}
	if _, err = h.lookup(rsrc, updateRequest.period, 0, 0); err != nil {
		return nil, err
	}

	if !rsrc.isSynced() {
		return nil, NewError(ErrNotSynced, fmt.Sprintf("NewMruRequest: object '%s' not in sync", rootAddr.Hex()))
	}

	updateRequest.multihash = rsrc.multihash
	updateRequest.rootAddr = rsrc.rootAddr
	updateRequest.metaHash = rsrc.metaHash
	updateRequest.resourceMetadata = rsrc.resourceMetadata

	// if we already have an update for this block then increment version
	// resource object MUST be in sync for version to be correct, but we checked this earlier in the method already
	if h.hasUpdate(rootAddr, updateRequest.period) {
		updateRequest.version = rsrc.version + 1
	} else {
		updateRequest.version = 1
	}

	return updateRequest, nil
}

// creates a metadata chunk
func (h *Handler) newMetaChunk(metadata *resourceMetadata) (chunk *storage.Chunk, metaHash []byte) {
	// the metadata chunk points to data of first blockheight + update frequency
	// from this we know from what blockheight we should look for updates, and how often
	// it also contains the name of the resource, so we know what resource we are working with

	// the key (rootAddr) of the metadata chunk is content-addressed
	// if it wasn't we couldn't replace it later
	// resolving this relationship is left up to external agents (for example ENS)
	rootAddr, metaHash, chunkData := metadata.hash()

	// make the chunk and send it to swarm
	chunk = storage.NewChunk(rootAddr, nil)
	chunk.SData = chunkData
	chunk.Size = int64(len(chunkData))

	return chunk, metaHash
}

// LookupLatest retrieves the latest version of the resource update with metadata chunk at params.Root
// It starts at the next period after the current block height, and upon failure
// tries the corresponding keys of each previous period until one is found
// (or startBlock is reached, in which case there are no updates).
func (h *Handler) Lookup(ctx context.Context, params *LookupParams) (*resource, error) {

	rsrc := h.get(params.Root)
	if rsrc == nil {
		return nil, NewError(ErrNothingToReturn, "resource not loaded")
	}
	if params.Period == 0 {
		// get our blockheight at this time and the next block of the update period
		currentblock := h.getCurrentTime(ctx)

		var period uint32
		period, err := getNextPeriod(rsrc.startTime, currentblock, rsrc.frequency)
		if err != nil {
			return nil, err
		}
		params.Period = period
	}
	return h.lookup(rsrc, params.Period, params.Version, params.Limit)
}

// LookupPreviousByName returns the resource before the one currently loaded in the resource index
// This is useful where resource updates are used incrementally in contrast to
// merely replacing content.
// Requires a synced resource object
func (h *Handler) LookupPrevious(ctx context.Context, params *LookupParams) (*resource, error) {
	rsrc := h.get(params.Root)
	if rsrc == nil {
		return nil, NewError(ErrNothingToReturn, "resource not loaded")
	}
	if !rsrc.isSynced() {
		return nil, NewError(ErrNotSynced, "LookupPrevious requires synced resource.")
	} else if rsrc.period == 0 {
		return nil, NewError(ErrNothingToReturn, " not found")
	}
	if rsrc.version > 1 {
		rsrc.version--
	} else if rsrc.period == 1 {
		return nil, NewError(ErrNothingToReturn, "Current update is the oldest")
	} else {
		rsrc.version = 0
		rsrc.period--
	}
	return h.lookup(rsrc, rsrc.period, rsrc.version, params.Limit)
}

// base code for public lookup methods
func (h *Handler) lookup(rsrc *resource, period uint32, version uint32, limit uint32) (*resource, error) {

	// we can't look for anything without a store
	if h.chunkStore == nil {
		return nil, NewError(ErrInit, "Call Handler.SetStore() before performing lookups")
	}

	// period 0 does not exist
	if period == 0 {
		return nil, NewError(ErrInvalidValue, "period must be >0")
	}

	// start from the last possible block period, and iterate previous ones until we find a match
	// if we hit startBlock we're out of options
	var specificversion bool
	if version > 0 {
		specificversion = true
	} else {
		version = 1
	}

	var hops uint32
	if limit == 0 {
		limit = h.queryMaxPeriods
	}
	log.Trace("resource lookup", "period", period, "version", version, "limit", limit)
	for period > 0 {
		if limit != 0 && hops > limit {
			return nil, NewError(ErrPeriodDepth, fmt.Sprintf("Lookup exceeded max period hops (%d)", limit))
		}
		updateAddr := resourceUpdateChunkAddr(period, version, rsrc.rootAddr)
		chunk, err := h.chunkStore.GetWithTimeout(context.TODO(), updateAddr, defaultRetrieveTimeout)
		if err == nil {
			if specificversion {
				return h.updateIndex(rsrc, chunk)
			}
			// check if we have versions > 1. If a version fails, the previous version is used and returned.
			log.Trace("rsrc update version 1 found, checking for version updates", "period", period, "updateAddr", updateAddr)
			for {
				newversion := version + 1
				updateAddr := resourceUpdateChunkAddr(period, newversion, rsrc.rootAddr)
				newchunk, err := h.chunkStore.GetWithTimeout(context.TODO(), updateAddr, defaultRetrieveTimeout)
				if err != nil {
					return h.updateIndex(rsrc, chunk)
				}
				chunk = newchunk
				version = newversion
				log.Trace("version update found, checking next", "version", version, "period", period, "updateAddr", updateAddr)
			}
		}
		log.Trace("rsrc update not found, checking previous period", "period", period, "updateAddr", updateAddr)
		period--
		hops++
	}
	return nil, NewError(ErrNotFound, "no updates found")
}

// Load retrieves the Mutable Resource metadata chunk stored at addr
// Upon retrieval it creates/updates the index entry for it with metadata corresponding to the chunk contents
func (h *Handler) Load(ctx context.Context, rootAddr storage.Address) (*resource, error) {
	chunk, err := h.chunkStore.GetWithTimeout(ctx, rootAddr, defaultRetrieveTimeout)
	if err != nil {
		return nil, NewError(ErrNotFound, err.Error())
	}

	// \TODO this is not enough to make sure the data isn't bogus. A normal content addressed chunk could still satisfy these criteria
	if len(chunk.SData) <= metadataChunkOffsetSize {
		return nil, NewError(ErrNothingToReturn, fmt.Sprintf("Invalid chunk length %d, should be minimum %d", len(chunk.SData), metadataChunkOffsetSize+1))
	}

	// create the index entry
	rsrc := &resource{}
	rsrc.unmarshalBinary(chunk.SData)
	rsrc.rootAddr, rsrc.metaHash = metadataHash(chunk.SData)
	if !bytes.Equal(rsrc.rootAddr, rootAddr) {
		return nil, NewError(ErrCorruptData, "Corrupt metadata chunk")
	}
	h.set(rootAddr, rsrc)
	log.Trace("resource index load", "rootkey", rootAddr, "name", rsrc.name, "startblock", rsrc.startTime, "frequency", rsrc.frequency)
	return rsrc, nil
}

// update mutable resource index map with specified content
func (h *Handler) updateIndex(rsrc *resource, chunk *storage.Chunk) (*resource, error) {

	// retrieve metadata from chunk data and check that it matches this mutable resource
	mru, err := parseUpdate(chunk.Addr, chunk.SData)
	if err != nil {
		return nil, NewError(ErrInvalidSignature, fmt.Sprintf("Invalid resource chunk: %s", err))
	}
	log.Trace("resource index update", "name", rsrc.name, "updatekey", chunk.Addr, "period", mru.period, "version", mru.version)

	// update our rsrcs entry map
	rsrc.lastKey = chunk.Addr
	rsrc.period = mru.period
	rsrc.version = mru.version
	rsrc.updated = time.Now()
	rsrc.data = make([]byte, len(mru.data))
	rsrc.multihash = mru.multihash
	rsrc.Reader = bytes.NewReader(rsrc.data)
	copy(rsrc.data, mru.data)
	log.Debug(" synced", "name", rsrc.name, "updateAddr", chunk.Addr, "period", rsrc.period, "version", rsrc.version)
	h.set(chunk.Addr, rsrc)
	return rsrc, nil
}

// retrieve update metadata from chunk data
// mirrors newUpdateChunk()
func parseUpdate(chunkAddr storage.Address, chunkdata []byte) (*SignedResourceUpdate, error) {
	// absolute minimum an update chunk can contain:
	// 14 = header + one byte of name + one byte of data

	// 2 bytes header Length
	// 2 bytes data length
	// 4 bytes period
	// 4 bytes version
	// 32 bytes rootAddr reference
	// 32 bytes metaHash digest

	if len(chunkdata) < 14 {
		return nil, NewError(ErrNothingToReturn, "chunk less than 13 bytes cannot be a resource update chunk")
	}
	cursor := 0
	headerlength := binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2])
	cursor += 2
	datalength := binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2])
	cursor += 2

	if datalength != 0 && int(2+2+headerlength+datalength+signatureLength) != len(chunkdata) {
		return nil, NewError(ErrNothingToReturn, "length specified in header is different than actual chunk size")
	}

	var exclsignlength int
	// we need extra magic if it's a multihash, since we used datalength 0 in header as an indicator of multihash content
	// retrieve the second varint and set this as the data length
	// TODO: merge with isMultihash code
	if datalength == 0 {
		uvarintbuf := bytes.NewBuffer(chunkdata[headerlength+4:])
		r, err := binary.ReadUvarint(uvarintbuf)
		if err != nil {
			errstr := fmt.Sprintf("corrupt multihash, hash id varint could not be read: %v", err)
			log.Warn(errstr)
			return nil, NewError(ErrCorruptData, errstr)

		}
		r, err = binary.ReadUvarint(uvarintbuf)
		if err != nil {
			errstr := fmt.Sprintf("corrupt multihash, hash length field could not be read: %v", err)
			log.Warn(errstr)
			return nil, NewError(ErrCorruptData, errstr)

		}
		exclsignlength = int(headerlength + uint16(r))
	} else {
		exclsignlength = int(headerlength + datalength + 4)
	}

	// the total length excluding signature is headerlength and datalength fields plus the length of the header and the data given in these fields
	exclsignlength = int(headerlength + datalength + 4)
	if exclsignlength > len(chunkdata) || exclsignlength < 14 {
		return nil, NewError(ErrNothingToReturn, fmt.Sprintf("Reported headerlength %d + datalength %d longer than actual chunk data length %d", headerlength, exclsignlength, len(chunkdata)))
	} else if exclsignlength < 14 {
		return nil, NewError(ErrNothingToReturn, fmt.Sprintf("Reported headerlength %d + datalength %d is smaller than minimum valid resource chunk length %d", headerlength, datalength, 14))
	}

	// at this point we can be satisfied that the data integrity is ok
	var period uint32
	var version uint32

	var data []byte
	period = binary.LittleEndian.Uint32(chunkdata[cursor : cursor+4])
	cursor += 4
	version = binary.LittleEndian.Uint32(chunkdata[cursor : cursor+4])
	cursor += 4

	rootAddr := storage.Address(make([]byte, storage.KeyLength))
	metaHash := make([]byte, storage.KeyLength)
	copy(rootAddr, chunkdata[cursor:cursor+storage.KeyLength])
	cursor += storage.KeyLength
	copy(metaHash, chunkdata[cursor:cursor+storage.KeyLength])
	cursor += storage.KeyLength

	// if multihash content is indicated we check the validity of the multihash
	// \TODO the check above for multihash probably is sufficient also for this case (or can be with a small adjustment) and if so this code should be removed
	var intdatalength int
	var ismultihash bool
	if datalength == 0 {
		var intheaderlength int
		var err error
		intdatalength, intheaderlength, err = multihash.GetMultihashLength(chunkdata[cursor:])
		if err != nil {
			log.Error("multihash parse error", "err", err)
			return nil, err
		}
		intdatalength += intheaderlength
		multihashboundary := cursor + intdatalength
		if len(chunkdata) != multihashboundary && len(chunkdata) < multihashboundary+signatureLength {
			log.Debug("multihash error", "chunkdatalen", len(chunkdata), "multihashboundary", multihashboundary)
			return nil, errors.New("Corrupt multihash data")
		}
		ismultihash = true
	} else {
		intdatalength = int(datalength)
	}
	data = make([]byte, intdatalength)
	copy(data, chunkdata[cursor:cursor+intdatalength])

	// omit signatures if we have no validator
	var signature *Signature
	cursor += intdatalength
	sigdata := chunkdata[cursor : cursor+signatureLength]
	if len(sigdata) > 0 {
		signature = &Signature{}
		copy(signature[:], sigdata)
	}

	r := &SignedResourceUpdate{
		signature:  signature,
		updateAddr: chunkAddr,
		resourceData: resourceData{
			period:    period,
			version:   version,
			rootAddr:  rootAddr,
			metaHash:  metaHash,
			data:      data,
			multihash: ismultihash,
		},
	}

	if err := r.Verify(); err != nil {
		return nil, NewError(ErrUnauthorized, fmt.Sprintf("Invalid signature: %v", err))
	}

	return r, nil

}

// Update adds an actual data update
// Uses the Mutable Resource metadata currently loaded in the resources map entry.
// It is the caller's responsibility to make sure that this data is not stale.
// Note that a Mutable Resource update cannot span chunks, and thus has a MAX NET LENGTH 4096, INCLUDING update header data and signature. An error will be returned if the total length of the chunk payload will exceed this limit.
func (h *Handler) Update(ctx context.Context, rootAddr storage.Address, mru *SignedResourceUpdate) (storage.Address, error) {
	return h.update(ctx, rootAddr, mru)
}

// create and commit an update
func (h *Handler) update(ctx context.Context, rootAddr storage.Address, mru *SignedResourceUpdate) (storage.Address, error) {

	if mru.multihash && isMultihash(mru.data) == 0 {
		return nil, NewError(ErrNothingToReturn, "Invalid multihash")
	}

	// we can't update anything without a store
	if h.chunkStore == nil {
		return nil, NewError(ErrInit, "Call Handler.SetStore() before updating")
	}

	// get the cached information

	rsrc := h.get(rootAddr)
	if rsrc == nil {
		return nil, NewError(ErrNotFound, fmt.Sprintf(" object '%s' not in index", rsrc.name))
	} else if !rsrc.isSynced() {
		return nil, NewError(ErrNotSynced, " object not in sync")
	}

	// an update can be only one chunk long; data length less header and signature data
	// 12 = length of header and data length fields (2xuint16) plus period and frequency value fields (2xuint32)
	datalimit := h.chunkSize() - int64(signatureLength-len(rsrc.name)-12)
	if int64(len(mru.data)) > datalimit {
		return nil, NewError(ErrDataOverflow, fmt.Sprintf("Data overflow: %d / %d bytes", len(mru.data), datalimit))
	}

	if rsrc.period == mru.period {
		if mru.version != rsrc.version+1 {
			return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid version for this period. Expected version=%d", rsrc.version+1))
		}
	} else {
		if !(mru.period > rsrc.period && mru.version == 1) {
			return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid version,period. Expected version=1 and period > %d", rsrc.period))
		}
	}

	chunk := newUpdateChunk(mru)

	// send the chunk
	h.chunkStore.Put(ctx, chunk)
	log.Trace("resource update", "updateAddr", mru.updateAddr, "lastperiod", mru.period, "version", mru.version, "data", chunk.SData, "multihash", mru.multihash)

	// update our resources map entry and return the new updateAddr
	rsrc.period = mru.period
	rsrc.version = mru.version
	rsrc.data = make([]byte, len(mru.data))
	copy(rsrc.data, mru.data)
	return mru.updateAddr, nil
}

// gets the current block height
func (h *Handler) getCurrentTime(ctx context.Context) uint64 {

	return h.timestampProvider.GetCurrentTime()

	/*
		blockheader, err := h.headerGetter.HeaderByNumber(ctx, name, nil)
		if err != nil {
			return 0, err
		}
		return blockheader.Number.Uint64(), nil
	*/
}

/** UNUSED functions
// BlockToPeriod calculates the period index (aka major version number) from a given block number
func (h *Handler) BlockToPeriod(name string, blocknumber uint64) (uint32, error) {
	return getNextPeriod(h.resources[name].startBlock, blocknumber, h.resources[name].frequency)
}

// PeriodToBlock calculates the block number from a given period index (aka major version number)
func (h *Handler) PeriodToBlock(name string, period uint32) uint64 {
	return h.resources[name].startBlock + (uint64(period) * h.resources[name].frequency)
}
**/

// Retrieves the resource index value for the given nameHash
func (h *Handler) get(rootAddr storage.Address) *resource {
	if rootAddr == nil {
		log.Warn("Handler.get with nil rootAddr")
		return nil
	}
	hashKey := *(*uint64)(unsafe.Pointer(&rootAddr[0]))
	h.resourceLock.RLock()
	defer h.resourceLock.RUnlock()
	rsrc := h.resources[hashKey]
	return rsrc
}

// Sets the resource index value for the given nameHash
func (h *Handler) set(rootAddr storage.Address, rsrc *resource) {
	hashKey := *(*uint64)(unsafe.Pointer(&rootAddr[0]))
	h.resourceLock.Lock()
	defer h.resourceLock.Unlock()
	h.resources[hashKey] = rsrc
}

// Checks if we already have an update on this resource, according to the value in the current state of the resource index
func (h *Handler) hasUpdate(rootAddr storage.Address, period uint32) bool {
	rsrc := h.get(rootAddr)
	return rsrc != nil && rsrc.period == period
}

func getAddressFromDataSig(datahash common.Hash, signature Signature) (common.Address, error) {
	pub, err := crypto.SigToPub(datahash.Bytes(), signature[:])
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pub), nil
}

// create an update chunk
func newUpdateChunk(mru *SignedResourceUpdate) *storage.Chunk {

	if mru.rootAddr == nil || mru.metaHash == nil {
		log.Warn("Call to newUpdateChunk with nil rootAddr or metaHash")
		return nil
	}
	// a datalength field set to 0 means the content is a multihash
	var datalength int
	if !mru.multihash {
		datalength = len(mru.data)
	}

	// prepend version, period, metaHash and rootAddr references
	headerlength := 4 + 4 + storage.KeyLength + storage.KeyLength

	actualdatalength := len(mru.data)
	chunk := storage.NewChunk(mru.updateAddr, nil)
	chunk.SData = make([]byte, 2+2+signatureLength+headerlength+actualdatalength) // initial 4 are uint16 length descriptors for headerlength and datalength

	// data header length does NOT include the header length prefix bytes themselves
	cursor := 0
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(headerlength))
	cursor += 2

	// data length
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(datalength))
	cursor += 2

	// header = period + version + rootAddr + metaHash
	binary.LittleEndian.PutUint32(chunk.SData[cursor:], mru.period)
	cursor += 4

	binary.LittleEndian.PutUint32(chunk.SData[cursor:], mru.version)
	cursor += 4

	copy(chunk.SData[cursor:], mru.rootAddr[:storage.KeyLength])
	cursor += storage.KeyLength
	copy(chunk.SData[cursor:], mru.metaHash[:storage.KeyLength])
	cursor += storage.KeyLength

	// add the data
	copy(chunk.SData[cursor:], mru.data)

	// signature is the last item in the chunk data

	cursor += actualdatalength
	copy(chunk.SData[cursor:], mru.signature[:])

	chunk.Size = int64(len(chunk.SData))
	return chunk
}

// Helper function to calculate the next update period number from the current block, start block and frequency
func getNextPeriod(start uint64, current uint64, frequency uint64) (uint32, error) {
	if current < start {
		return 0, NewError(ErrInvalidValue, fmt.Sprintf("given current block value %d < start block %d", current, start))
	}
	blockdiff := current - start
	period := blockdiff / frequency
	return uint32(period + 1), nil
}

// ToSafeName is a helper function to create an valid idna of a given resource update name
func ToSafeName(name string) (string, error) {
	return idna.ToASCII(name)
}

// check that name identifiers contain valid bytes
// Strings created using ToSafeName() should satisfy this check
func isSafeName(name string) bool {
	if name == "" {
		return false
	}
	validname, err := idna.ToASCII(name)
	if err != nil {
		return false
	}
	return validname == name
}

// if first byte is the start of a multihash this function will try to parse it
// if successful it returns the length of multihash data, 0 otherwise
func isMultihash(data []byte) int {
	cursor := 0
	_, c := binary.Uvarint(data)
	if c <= 0 {
		log.Warn("Corrupt multihash data, hashtype is unreadable")
		return 0
	}
	cursor += c
	hashlength, c := binary.Uvarint(data[cursor:])
	if c <= 0 {
		log.Warn("Corrupt multihash data, hashlength is unreadable")
		return 0
	}
	cursor += c
	// we cheekily assume hashlength < maxint
	inthashlength := int(hashlength)
	if len(data[cursor:]) < inthashlength {
		log.Warn("Corrupt multihash data, hash does not align with data boundary")
		return 0
	}
	return cursor + inthashlength
}

// NewResourceHash will create a deterministic address from the update metadata
// format is: hash(period|version|publickey|namehash)
func NewResourceHash(period uint32, version uint32, rootAddr storage.Address) []byte {
	buf := bytes.NewBuffer(nil)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, period)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, version)
	buf.Write(b)
	buf.Write(rootAddr[:])
	return buf.Bytes()
}

func verifyResourceOwnership(ownerAddr common.Address, metaHash []byte, rootAddr storage.Address) bool {
	hasher := hashPool.Get().(storage.SwarmHash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(metaHash)
	hasher.Write(ownerAddr.Bytes())
	rootAddr2 := hasher.Sum(nil)
	return bytes.Equal(rootAddr2, rootAddr)
}
