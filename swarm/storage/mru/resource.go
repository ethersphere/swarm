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

	"golang.org/x/net/idna"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/multihash"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const (
	signatureLength         = 65
	metadataChunkOffsetSize = 16 + common.AddressLength
	chunkSize               = 4096 // temporary until we implement FileStore in the resourcehandler
	defaultStoreTimeout     = 4000 * time.Millisecond
	hasherCount             = 8
	resourceHash            = storage.SHA3Hash
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

// Signature is an alias for a static byte array with the size of a signature
type Signature [signatureLength]byte

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

// encapsulates an specific resource update. When synced it contains the most recent
// version of the resource update data.
type resource struct {
	*bytes.Reader
	Multihash  bool
	name       string
	nameHash   common.Hash
	startBlock uint64
	lastPeriod uint32
	lastKey    storage.Address
	frequency  uint64
	version    uint32
	data       []byte
	owner      common.Address
	updated    time.Time
}

func (r *resource) Context() context.Context {
	return context.TODO()
}

// TODO Expire content after a defined period (to force resync)
func (r *resource) isSynced() bool {
	return !r.updated.IsZero()
}

func (r *resource) NameHash() common.Hash {
	return r.nameHash
}

// implements (which?) interface
func (r *resource) Size(ctx context.Context, _ chan bool) (int64, error) {
	if !r.isSynced() {
		return 0, NewError(ErrNotSynced, "Not synced")
	}
	return int64(len(r.data)), nil
}

func (r *resource) Name() string {
	return r.name
}

func (r *resource) UnmarshalBinary(data []byte) error {
	r.startBlock = binary.LittleEndian.Uint64(data)
	r.frequency = binary.LittleEndian.Uint64(data[8:])
	copy(r.owner[:], data[16:16+common.AddressLength])
	r.name = string(data[16+common.AddressLength:])
	return nil
}

func (r *resource) MarshalBinary() ([]byte, error) {
	b := make([]byte, 16+len(r.name))
	binary.LittleEndian.PutUint64(b, r.startBlock)
	binary.LittleEndian.PutUint64(b[8:], r.frequency)
	copy(b[16:], r.owner[:])
	copy(b[16+common.AddressLength:], []byte(r.name))
	return b, nil
}

type headerGetter interface {
	HeaderByNumber(context.Context, string, *big.Int) (*types.Header, error)
}

// Handler is the API for Mutable Resources
// It enables creating, updating, syncing and retrieving resources and their update data
type Handler struct {
	chunkStore      *storage.NetStore
	HashSize        int
	signer          Signer
	headerGetter    headerGetter
	resources       map[string]*resource
	hashPool        sync.Pool
	resourceLock    sync.RWMutex
	storeTimeout    time.Duration
	queryMaxPeriods uint32
}

// HandlerParams pass parameters to the Handler constructor NewHandler
// Signer and HeaderGetter are mandatory parameters
type HandlerParams struct {
	QueryMaxPeriods uint32
	Signer          Signer
	HeaderGetter    headerGetter
}

// NewHandler creates a new Mutable Resource API
func NewHandler(params *HandlerParams) (*Handler, error) {
	if params.Signer == nil {
		return nil, NewError(ErrInit, "Signer cannot be nil")
	}
	if params.HeaderGetter == nil {
		return nil, NewError(ErrInit, "HeaderGetter cannot be nil")
	}
	rh := &Handler{
		headerGetter: params.HeaderGetter,
		resources:    make(map[string]*resource),
		storeTimeout: defaultStoreTimeout,
		signer:       params.Signer,
		hashPool: sync.Pool{
			New: func() interface{} {
				return storage.MakeHashFunc(resourceHash)()
			},
		},
		queryMaxPeriods: params.QueryMaxPeriods,
	}

	for i := 0; i < hasherCount; i++ {
		hashfunc := storage.MakeHashFunc(resourceHash)()
		if rh.HashSize == 0 {
			rh.HashSize = hashfunc.Size()
		}
		rh.hashPool.Put(hashfunc)
	}

	return rh, nil
}

// SetStore sets the store backend for the Mutable Resource API
func (h *Handler) SetStore(store *storage.NetStore) {
	h.chunkStore = store
}

// Validate is a chunk validation method
// If it's a resource update, the chunk address is checked against the public key of the update's signature
// It implements the storage.ChunkValidator interface
func (h *Handler) Validate(addr storage.Address, data []byte) bool {

	// does the chunk data make sense?
	// This is not 100% safe evaluation, since content addressed data can coincidentally produce data with length header matching content size
	signature, period, version, name, parseddata, _, err := h.parseUpdate(data)
	if err != nil {
		log.Error("Invalid resource chunk")
		return false
	}

	// Check if signature is valid against the chunk address
	// Since we force signatures to be used, we can know be sure that this is valid data
	digest := h.keyDataHash(addr, parseddata)
	publicKey, err := crypto.SigToPub(digest.Bytes(), signature[:])
	if err != nil {
		log.Error("Unparseable signature in resource chunk %s", addr)
		return false
	}
	signAddr := crypto.PubkeyToAddress(*publicKey)
	nameHash := ens.EnsNode(name)
	checkAddr := h.resourceHash(period, version, nameHash, signAddr)
	if !bytes.Equal(checkAddr, addr) {
		log.Error("Invalid signature on resource chunk")
		return false
	}
	return true
}

// create the resource update digest used in signatures
func (h *Handler) keyDataHash(addr storage.Address, data []byte) common.Hash {
	hasher := h.hashPool.Get().(storage.SwarmHash)
	defer h.hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(addr[:])
	hasher.Write(data)
	return common.BytesToHash(hasher.Sum(nil))
}

// GetContent retrieves the data payload of the last synced update of the Mutable Resource
func (h *Handler) GetContent(addr storage.Address) (storage.Address, []byte, error) {
	rsrc := h.get(addr.Hex())
	if rsrc == nil || !rsrc.isSynced() {
		return nil, nil, NewError(ErrNotFound, " does not exist or is not synced")
	}
	return rsrc.lastKey, rsrc.data, nil
}

// GetLastPeriod retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetLastPeriod(addr storage.Address) (uint32, error) {
	rsrc := h.get(addr.Hex())
	if rsrc == nil {
		return 0, NewError(ErrNotFound, " does not exist")
	} else if !rsrc.isSynced() {
		return 0, NewError(ErrNotSynced, " is not synced")
	}
	return rsrc.lastPeriod, nil
}

// GetVersion retrieves the period of the last synced update of the Mutable Resource
func (h *Handler) GetVersion(addr storage.Address) (uint32, error) {
	rsrc := h.get(addr.Hex())
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
// It uses the public key from the Signer registered with the Handler object
// The start block of the resource update will be the actual current block height of the connected network.
func (h *Handler) New(ctx context.Context, name string, frequency uint64) (storage.Address, *resource, error) {
	return h.NewWithAddress(ctx, name, frequency, h.signer.Address())
}

// NewWithPublicKey performs the same action as New, but enables a custom public key to be used for the Mutable Resource
func (h *Handler) NewWithAddress(ctx context.Context, name string, frequency uint64, ownerAddr common.Address) (storage.Address, *resource, error) {

	// frequency 0 is invalid
	if frequency == 0 {
		return nil, nil, NewError(ErrInvalidValue, "frequency cannot be 0")
	}

	// make sure name only contains ascii values
	if !isSafeName(name) {
		return nil, nil, NewError(ErrInvalidValue, fmt.Sprintf("invalid name: '%s'", name))
	}

	// get our blockheight at this time
	currentblock, err := h.getBlock(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	// create the meta chunk and store it in swarm
	chunk := h.newMetaChunk(name, currentblock, frequency, ownerAddr)
	h.chunkStore.Put(ctx, chunk)

	nameHash := ens.EnsNode(name)
	log.Debug("new resource", "name", name, "key", nameHash, "startBlock", currentblock, "frequency", frequency, "owner", ownerAddr)

	// create the internal index for the resource and populate it with the data of the first version
	rsrc := &resource{
		startBlock: currentblock,
		frequency:  frequency,
		name:       name,
		nameHash:   nameHash,
		updated:    time.Now(),
	}
	copy(rsrc.owner[:], ownerAddr[:])
	h.set(chunk.Addr.Hex(), rsrc)

	return chunk.Addr, rsrc, nil
}

// creates a metadata chunk
func (h *Handler) newMetaChunk(name string, startBlock uint64, frequency uint64, ownerAddr common.Address) *storage.Chunk {
	// the metadata chunk points to data of first blockheight + update frequency
	// from this we know from what blockheight we should look for updates, and how often
	// it also contains the name of the resource, so we know what resource we are working with
	data := make([]byte, metadataChunkOffsetSize+len(name))

	// root block has first two bytes both set to 0, which distinguishes from update bytes
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val, startBlock)
	copy(data[:8], val)
	binary.LittleEndian.PutUint64(val, frequency)
	copy(data[8:16], val)
	copy(data[16:], ownerAddr[:])
	copy(data[16+common.AddressLength:], []byte(name))

	// the key of the metadata chunk is content-addressed
	// if it wasn't we couldn't replace it later
	// resolving this relationship is left up to external agents (for example ENS)
	hasher := h.hashPool.Get().(storage.SwarmHash)
	hasher.Reset()
	hasher.Write(data)
	key := hasher.Sum(nil)
	h.hashPool.Put(hasher)

	// make the chunk and send it to swarm
	chunk := storage.NewChunk(key, nil)
	chunk.SData = make([]byte, metadataChunkOffsetSize+len(name))
	copy(chunk.SData, data)
	return chunk
}

// LookupLatest retrieves the latest version of the resource update with metadata chunk at params.Root
// It starts at the next period after the current block height, and upon failure
// tries the corresponding keys of each previous period until one is found
// (or startBlock is reached, in which case there are no updates).
func (h *Handler) Lookup(ctx context.Context, params *LookupParams) (*resource, error) { //nameHash common.Hash, refresh bool, maxLookup *LookupParams) (*resource, error) {

	rsrc := h.get(params.Root.Hex())
	if rsrc == nil {
		return nil, NewError(ErrNothingToReturn, "resource not loaded")
	}
	if params.Period == 0 {
		// get our blockheight at this time and the next block of the update period
		currentblock, err := h.getBlock(ctx, rsrc.name)
		if err != nil {
			return nil, err
		}
		var period uint32
		period, err = getNextPeriod(rsrc.startBlock, currentblock, rsrc.frequency)
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
func (h *Handler) LookupPrevious(ctx context.Context, params *LookupParams) (*resource, error) { //nameHash common.Hash, maxLookup *LookupParams) (*resource, error) {
	rsrc := h.get(params.Root.Hex())
	if rsrc == nil {
		return nil, NewError(ErrNothingToReturn, "resource not loaded")
	}
	if !rsrc.isSynced() {
		return nil, NewError(ErrNotSynced, "LookupPrevious requires synced resource.")
	} else if rsrc.lastPeriod == 0 {
		return nil, NewError(ErrNothingToReturn, " not found")
	}
	if rsrc.version > 1 {
		rsrc.version--
	} else if rsrc.lastPeriod == 1 {
		return nil, NewError(ErrNothingToReturn, "Current update is the oldest")
	} else {
		rsrc.version = 0
		rsrc.lastPeriod--
	}
	return h.lookup(rsrc, rsrc.lastPeriod, rsrc.version, params.Limit)
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
		key := h.resourceHash(period, version, rsrc.nameHash, rsrc.owner)
		chunk, err := h.chunkStore.GetWithTimeout(context.TODO(), key, defaultRetrieveTimeout)
		if err == nil {
			if specificversion {
				return h.updateIndex(rsrc, chunk)
			}
			// check if we have versions > 1. If a version fails, the previous version is used and returned.
			log.Trace("rsrc update version 1 found, checking for version updates", "period", period, "key", key)
			for {
				newversion := version + 1
				key := h.resourceHash(period, newversion, rsrc.nameHash, rsrc.owner)
				newchunk, err := h.chunkStore.GetWithTimeout(context.TODO(), key, defaultRetrieveTimeout)
				if err != nil {
					return h.updateIndex(rsrc, chunk)
				}
				chunk = newchunk
				version = newversion
				log.Trace("version update found, checking next", "version", version, "period", period, "key", key)
			}
		}
		log.Trace("rsrc update not found, checking previous period", "period", period, "key", key)
		period--
		hops++
	}
	return nil, NewError(ErrNotFound, "no updates found")
}

// Load retrieves the Mutable Resource metadata chunk stored at addr
// Upon retrieval it creates/updates the index entry for it with metadata corresponding to the chunk contents
func (h *Handler) Load(ctx context.Context, addr storage.Address) (*resource, error) {
	chunk, err := h.chunkStore.GetWithTimeout(ctx, addr, defaultRetrieveTimeout)
	if err != nil {
		return nil, NewError(ErrNotFound, err.Error())
	}

	// \TODO this is not enough to make sure the data isn't bogus. A normal content addressed chunk could still satisfy these criteria
	if len(chunk.SData) <= metadataChunkOffsetSize {
		return nil, NewError(ErrNothingToReturn, fmt.Sprintf("Invalid chunk length %d, should be minimum %d", len(chunk.SData), metadataChunkOffsetSize+1))
	}

	// create the index entry
	rsrc := &resource{}
	rsrc.UnmarshalBinary(chunk.SData[:])
	rsrc.nameHash = ens.EnsNode(rsrc.name)
	h.set(addr.Hex(), rsrc)
	log.Trace("resource index load", "rootkey", addr, "name", rsrc.name, "namehash", rsrc.nameHash, "startblock", rsrc.startBlock, "frequency", rsrc.frequency)
	return rsrc, nil
}

// update mutable resource index map with specified content
func (h *Handler) updateIndex(rsrc *resource, chunk *storage.Chunk) (*resource, error) {

	// retrieve metadata from chunk data and check that it matches this mutable resource
	signature, period, version, name, data, multihash, err := h.parseUpdate(chunk.SData)
	if rsrc.name != name {
		return nil, NewError(ErrNothingToReturn, fmt.Sprintf("Update belongs to '%s', but have '%s'", name, rsrc.name))
	}
	log.Trace("resource index update", "name", rsrc.name, "namehash", rsrc.nameHash, "updatekey", chunk.Addr, "period", period, "version", version)

	// check signature (if signer algorithm is present)
	// \TODO maybe this check is redundant if also checked upon retrieval of chunk
	if signature != nil {
		digest := h.keyDataHash(chunk.Addr, data)
		_, err = getAddressFromDataSig(digest, *signature)
		if err != nil {
			return nil, NewError(ErrUnauthorized, fmt.Sprintf("Invalid signature: %v", err))
		}
	}

	// update our rsrcs entry map
	rsrc.lastKey = chunk.Addr
	rsrc.lastPeriod = period
	rsrc.version = version
	rsrc.updated = time.Now()
	rsrc.data = make([]byte, len(data))
	rsrc.Multihash = multihash
	rsrc.Reader = bytes.NewReader(rsrc.data)
	copy(rsrc.data, data)
	log.Debug(" synced", "name", rsrc.name, "key", chunk.Addr, "period", rsrc.lastPeriod, "version", rsrc.version)
	h.set(chunk.Addr.Hex(), rsrc)
	return rsrc, nil
}

// retrieve update metadata from chunk data
// mirrors newUpdateChunk()
func (h *Handler) parseUpdate(chunkdata []byte) (*Signature, uint32, uint32, string, []byte, bool, error) {
	// absolute minimum an update chunk can contain:
	// 14 = header + one byte of name + one byte of data
	if len(chunkdata) < 14 {
		return nil, 0, 0, "", nil, false, NewError(ErrNothingToReturn, "chunk less than 13 bytes cannot be a resource update chunk")
	}
	cursor := 0
	headerlength := binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2])
	cursor += 2
	datalength := binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2])
	cursor += 2

	if int(headerlength+datalength) > len(chunkdata) {
		return nil, 0, 0, "", nil, false, NewError(ErrNothingToReturn, "length specified in header is greater than actual chunk size")
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
			return nil, 0, 0, "", nil, false, NewError(ErrCorruptData, errstr)

		}
		r, err = binary.ReadUvarint(uvarintbuf)
		if err != nil {
			errstr := fmt.Sprintf("corrupt multihash, hash length field could not be read: %v", err)
			log.Warn(errstr)
			return nil, 0, 0, "", nil, false, NewError(ErrCorruptData, errstr)

		}
		exclsignlength = int(headerlength + uint16(r))
	} else {
		exclsignlength = int(headerlength + datalength + 4)
	}

	// the total length excluding signature is headerlength and datalength fields plus the length of the header and the data given in these fields
	exclsignlength = int(headerlength + datalength + 4)
	if exclsignlength > len(chunkdata) || exclsignlength < 14 {
		return nil, 0, 0, "", nil, false, NewError(ErrNothingToReturn, fmt.Sprintf("Reported headerlength %d + datalength %d longer than actual chunk data length %d", headerlength, exclsignlength, len(chunkdata)))
	} else if exclsignlength < 14 {
		return nil, 0, 0, "", nil, false, NewError(ErrNothingToReturn, fmt.Sprintf("Reported headerlength %d + datalength %d is smaller than minimum valid resource chunk length %d", headerlength, datalength, 14))
	}

	// at this point we can be satisfied that the data integrity is ok
	var period uint32
	var version uint32
	var name string
	var data []byte
	period = binary.LittleEndian.Uint32(chunkdata[cursor : cursor+4])
	cursor += 4
	version = binary.LittleEndian.Uint32(chunkdata[cursor : cursor+4])
	cursor += 4
	namelength := int(headerlength) - cursor + 4
	if l := len(chunkdata); l < cursor+namelength {
		return nil, 0, 0, "", nil, false, NewError(ErrNothingToReturn, fmt.Sprintf("chunk less than %v bytes is too short to read the name", l))
	}
	name = string(chunkdata[cursor : cursor+namelength])
	cursor += namelength

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
			return nil, 0, 0, "", nil, false, err
		}
		intdatalength += intheaderlength
		multihashboundary := cursor + intdatalength
		if len(chunkdata) != multihashboundary && len(chunkdata) < multihashboundary+signatureLength {
			log.Debug("multihash error", "chunkdatalen", len(chunkdata), "multihashboundary", multihashboundary)
			return nil, 0, 0, "", nil, false, errors.New("Corrupt multihash data")
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
	if h.signer != nil {
		sigdata := chunkdata[cursor : cursor+signatureLength]
		if len(sigdata) > 0 {
			signature = &Signature{}
			copy(signature[:], sigdata)
		}
	}

	return signature, period, version, name, data, ismultihash, nil
}

// Adds an actual data update
//
// Uses the data currently loaded in the resources map entry.
// It is the caller's responsibility to make sure that this data is not stale.
//
// A resource update cannot span chunks, and thus has max length 4096
func (h *Handler) UpdateMultihash(ctx context.Context, addr storage.Address, data []byte) (storage.Address, error) {
	// \TODO perhaps this check should be in newUpdateChunk()
	if _, _, err := multihash.GetMultihashLength(data); err != nil {
		return nil, NewError(ErrNothingToReturn, err.Error())
	}
	return h.update(ctx, addr, data, true)
}

// Update adds an actual data update
// Uses the Mutable Resource metadata currently loaded in the resources map entry.
// It is the caller's responsibility to make sure that this data is not stale.
// Note that a Mutable Resource update cannot span chunks, and thus has a MAX NET LENGTH 4096, INCLUDING update header data and signature. An error will be returned if the total length of the chunk payload will exceed this limit.
func (h *Handler) Update(ctx context.Context, addr storage.Address, data []byte) (storage.Address, error) {
	return h.update(ctx, addr, data, false)
}

// create and commit an update
func (h *Handler) update(ctx context.Context, addr storage.Address, data []byte, multihash bool) (storage.Address, error) {

	// zero-length updates are bogus
	if len(data) == 0 {
		return nil, NewError(ErrInvalidValue, "I refuse to waste swarm space for updates with empty values, amigo (data length is 0)")
	}

	// we can't update anything without a store
	if h.chunkStore == nil {
		return nil, NewError(ErrInit, "Call Handler.SetStore() before updating")
	}

	// get the cached information
	addrHex := addr.Hex()
	rsrc := h.get(addrHex)
	if rsrc == nil {
		return nil, NewError(ErrNotFound, fmt.Sprintf(" object '%s' not in index", rsrc.name))
	} else if !rsrc.isSynced() {
		return nil, NewError(ErrNotSynced, " object not in sync")
	}

	// an update can be only one chunk long; data length less header and signature data
	// 12 = length of header and data length fields (2xuint16) plus period and frequency value fields (2xuint32)
	datalimit := h.chunkSize() - int64(signatureLength-len(rsrc.name)-12)
	if int64(len(data)) > datalimit {
		return nil, NewError(ErrDataOverflow, fmt.Sprintf("Data overflow: %d / %d bytes", len(data), datalimit))
	}

	// get our blockheight at this time and the next block of the update period
	currentblock, err := h.getBlock(ctx, rsrc.name)
	if err != nil {
		return nil, NewError(ErrIO, fmt.Sprintf("Could not get block height: %v", err))
	}
	nextperiod, err := getNextPeriod(rsrc.startBlock, currentblock, rsrc.frequency)
	if err != nil {
		return nil, err
	}

	// if we already have an update for this block then increment version
	// resource object MUST be in sync for version to be correct, but we checked this earlier in the method already
	var version uint32
	if h.hasUpdate(addr, nextperiod) {
		version = rsrc.version
	}
	version++

	// calculate the chunk key
	key := h.resourceHash(nextperiod, version, rsrc.nameHash, rsrc.owner)

	// if we have a signing function, sign the update
	// \TODO this code should probably be consolidated with corresponding code in New()
	var signature *Signature
	if h.signer != nil {
		// sign the data hash with the key
		digest := h.keyDataHash(key, data)
		sig, err := h.signer.Sign(digest)
		if err != nil {
			return nil, NewError(ErrInvalidSignature, fmt.Sprintf("Sign fail: %v", err))
		}
		signature = &sig

		// get the address of the signer (which also checks that it's a valid signature)
		_, err = getAddressFromDataSig(digest, *signature)
		if err != nil {
			return nil, NewError(ErrInvalidSignature, fmt.Sprintf("Invalid data/signature: %v", err))
		}
	}

	// a datalength field set to 0 means the content is a multihash
	var datalength int
	if !multihash {
		datalength = len(data)
	}
	chunk := newUpdateChunk(key, signature, nextperiod, version, rsrc.name, data, datalength)

	// send the chunk
	h.chunkStore.Put(ctx, chunk)
	log.Trace("resource update", "name", rsrc.name, "key", key, "currentblock", currentblock, "lastperiod", nextperiod, "version", version, "data", chunk.SData, "multihash", multihash)

	// update our resources map entry and return the new key
	rsrc.lastPeriod = nextperiod
	rsrc.version = version
	rsrc.data = make([]byte, len(data))
	copy(rsrc.data, data)
	return key, nil
}

// gets the current block height
func (h *Handler) getBlock(ctx context.Context, name string) (uint64, error) {
	blockheader, err := h.headerGetter.HeaderByNumber(ctx, name, nil)
	if err != nil {
		return 0, err
	}
	return blockheader.Number.Uint64(), nil
}

// BlockToPeriod calculates the period index (aka major version number) from a given block number
func (h *Handler) BlockToPeriod(name string, blocknumber uint64) (uint32, error) {
	return getNextPeriod(h.resources[name].startBlock, blocknumber, h.resources[name].frequency)
}

// PeriodToBlock calculates the block number from a given period index (aka major version number)
func (h *Handler) PeriodToBlock(name string, period uint32) uint64 {
	return h.resources[name].startBlock + (uint64(period) * h.resources[name].frequency)
}

// Retrieves the resource index value for the given nameHash
func (h *Handler) get(nameHash string) *resource {
	h.resourceLock.RLock()
	defer h.resourceLock.RUnlock()
	rsrc := h.resources[nameHash]
	return rsrc
}

// Sets the resource index value for the given nameHash
func (h *Handler) set(nameHash string, rsrc *resource) {
	h.resourceLock.Lock()
	defer h.resourceLock.Unlock()
	h.resources[nameHash] = rsrc
}

// used for chunk keys
//func (h *Handler) resourceHash(period uint32, version uint32, nameHash common.Hash, publicKeyBytes [publicKeyLength]byte) storage.Address {
func (h *Handler) resourceHash(period uint32, version uint32, nameHash common.Hash, ownerAddr common.Address) storage.Address {
	hasher := h.hashPool.Get().(storage.SwarmHash)
	defer h.hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(NewResourceHash(period, version, nameHash, ownerAddr))
	return hasher.Sum(nil)
}

// Checks if we already have an update on this resource, according to the value in the current state of the resource index
func (h *Handler) hasUpdate(addr storage.Address, period uint32) bool {
	return h.resources[addr.Hex()].lastPeriod == period
}

func getAddressFromDataSig(datahash common.Hash, signature Signature) (common.Address, error) {
	pub, err := crypto.SigToPub(datahash.Bytes(), signature[:])
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pub), nil
}

// create an update chunk
func newUpdateChunk(addr storage.Address, signature *Signature, period uint32, version uint32, name string, data []byte, datalength int) *storage.Chunk {

	// no signatures if no validator
	var signaturelength int
	if signature != nil {
		signaturelength = signatureLength
	}

	// prepend version and period to allow reverse lookups
	headerlength := len(name) + 4 + 4

	actualdatalength := len(data)
	chunk := storage.NewChunk(addr, nil)
	chunk.SData = make([]byte, 4+signaturelength+headerlength+actualdatalength) // initial 4 are uint16 length descriptors for headerlength and datalength

	// data header length does NOT include the header length prefix bytes themselves
	cursor := 0
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(headerlength))
	cursor += 2

	// data length
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(datalength))
	cursor += 2

	// header = period + version + name
	binary.LittleEndian.PutUint32(chunk.SData[cursor:], period)
	cursor += 4

	binary.LittleEndian.PutUint32(chunk.SData[cursor:], version)
	cursor += 4

	namebytes := []byte(name)
	copy(chunk.SData[cursor:], namebytes)
	cursor += len(namebytes)

	// add the data
	copy(chunk.SData[cursor:], data)

	// if signature is present it's the last item in the chunk data
	if signature != nil {
		cursor += actualdatalength
		copy(chunk.SData[cursor:], signature[:])
	}

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

// NewResourceHash will create a deterministic address from the update metadata
// format is: hash(period|version|publickey|namehash)
func NewResourceHash(period uint32, version uint32, namehash common.Hash, ownerAddr common.Address) []byte {
	buf := bytes.NewBuffer(nil)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, period)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, version)
	buf.Write(b)
	buf.Write(ownerAddr[:])
	buf.Write(namehash[:])
	return buf.Bytes()
}
