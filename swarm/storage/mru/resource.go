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
	"time"

	"github.com/ethereum/go-ethereum/swarm/storage"
)

const (
	chunkSize              = 4096 // temporary until we implement FileStore in the resourcehandler
	defaultStoreTimeout    = 4000 * time.Millisecond
	hasherCount            = 8
	resourceHashAlgorithm  = storage.SHA3Hash
	defaultRetrieveTimeout = 100 * time.Millisecond
)

// resource caches resource data. When synced it contains the most recent
// version of the resource data and the metadata of its root chunk.
type resource struct {
	resourceUpdate
	ResourceMetadata
	*bytes.Reader
	lastKey storage.Address
	updated time.Time
}

// TODO Expire content after a defined period (to force resync)
func (r *resource) isSynced() bool {
	return !r.updated.IsZero()
}

//Whether the resource data should be interpreted as multihash
func (r *resourceUpdate) Multihash() bool {
	return r.multihash
}

// implements storage.LazySectionReader
func (r *resource) Size(_ chan bool) (int64, error) {
	if !r.isSynced() {
		return 0, NewError(ErrNotSynced, "Not synced")
	}
	return int64(len(r.resourceUpdate.data)), nil
}

//returns the resource's human-readable name
func (r *resource) Name() string {
	return r.ResourceMetadata.Name
}

// Helper function to calculate the next update period number from the current time, start time and frequency
func getNextPeriod(start uint64, current uint64, frequency uint64) (uint32, error) {
	if current < start {
		return 0, NewErrorf(ErrInvalidValue, "given current time value %d < start time %d", current, start)
	}
	if frequency == 0 {
		return 0, NewError(ErrInvalidValue, "frequency is 0")
	}
	timeDiff := current - start
	period := timeDiff / frequency
	return uint32(period + 1), nil
}
