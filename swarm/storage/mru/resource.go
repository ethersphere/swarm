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
	"fmt"
	"math/big"
	"time"

	"golang.org/x/net/idna"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const (
	chunkSize              = 4096 // temporary until we implement FileStore in the resourcehandler
	defaultStoreTimeout    = 4000 * time.Millisecond
	hasherCount            = 8
	resourceHashAlgorithm  = storage.SHA3Hash
	defaultRetrieveTimeout = 100 * time.Millisecond
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

// resource caches resource data. When synced it contains the most recent
// version of the resource data and the metadata of its root chunk.
type resource struct {
	resourceUpdate
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
func (r *resourceUpdate) Multihash() bool {
	return r.multihash
}

// implements (which?) interface
func (r *resource) Size(ctx context.Context, _ chan bool) (int64, error) {
	if !r.isSynced() {
		return 0, NewError(ErrNotSynced, "Not synced")
	}
	return int64(len(r.resourceUpdate.data)), nil
}

//returns the resource's human-readable name
func (r *resource) Name() string {
	return r.name
}

// Helper function to calculate the next update period number from the current time, start time and frequency
func getNextPeriod(start uint64, current uint64, frequency uint64) (uint32, error) {
	if current < start {
		return 0, NewError(ErrInvalidValue, fmt.Sprintf("given current time value %d < start time %d", current, start))
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
