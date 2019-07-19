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

package localstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/shed"
	"github.com/syndtr/goleveldb/leveldb"
)

// Get returns a chunk from the database. If the chunk is
// not found chunk.ErrChunkNotFound will be returned.
// All required indexes will be updated required by the
// Getter Mode. Get is required to implement chunk.Store
// interface.
func (db *DB) Get(ctx context.Context, mode chunk.ModeGet, addr chunk.Address) (ch chunk.Chunk, err error) {
	metricName := fmt.Sprintf("localstore.Get.%s", mode)

	metrics.GetOrRegisterCounter(metricName, nil).Inc(1)
	defer totalTimeMetric(metricName, time.Now())

	defer func() {
		if err != nil {
			metrics.GetOrRegisterCounter(metricName+".error", nil).Inc(1)
		}
	}()

	out, err := db.get(mode, addr)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, chunk.ErrChunkNotFound
		}
		return nil, err
	}
	return chunk.NewChunk(out.Address, out.Data), nil
}

// get returns Item from the retrieval index
// and updates other indexes.
func (db *DB) get(mode chunk.ModeGet, addr chunk.Address) (out shed.Item, err error) {
	item := addressToItem(addr)

	out, err = db.retrievalDataIndex.Get(item)
	if err != nil {
		return out, err
	}
	switch mode {
	// update the access timestamp and gc index
	case chunk.ModeGetRequest:
		if db.updateGCSem != nil {
			// wait before creating new goroutines
			// if updateGCSem buffer id full
			db.updateGCSem <- struct{}{}
		}
		db.updateGCWG.Add(1)
		go func() {
			defer db.updateGCWG.Done()
			if db.updateGCSem != nil {
				// free a spot in updateGCSem buffer
				// for a new goroutine
				defer func() { <-db.updateGCSem }()
			}

			metricName := "localstore.updateGC"
			metrics.GetOrRegisterCounter(metricName, nil).Inc(1)
			defer totalTimeMetric(metricName, time.Now())

			err := db.updateGC(out)
			if err != nil {
				metrics.GetOrRegisterCounter(metricName+".error", nil).Inc(1)
				log.Error("localstore update gc", "err", err)
			}
			// if gc update hook is defined, call it
			if testHookUpdateGC != nil {
				testHookUpdateGC()
			}
		}()

	// no updates to indexes
	case chunk.ModeGetSync:
	case chunk.ModeGetLookup:
	default:
		return out, ErrInvalidMode
	}
	return out, nil
}

// updateGC updates garbage collection index for
// a single item. Provided item is expected to have
// only Address and Data fields with non zero values,
// which is ensured by the get function.
func (db *DB) updateGC(item shed.Item) (err error) {
	db.batchMu.Lock()
	defer db.batchMu.Unlock()

	batch := new(leveldb.Batch)

	// update accessTimeStamp in retrieve, gc

	i, err := db.retrievalAccessIndex.Get(item)
	switch err {
	case nil:
		item.AccessTimestamp = i.AccessTimestamp
	case leveldb.ErrNotFound:
		// no chunk accesses
	default:
		return err
	}
	if item.AccessTimestamp == 0 {
		// chunk is not yet synced
		// do not add it to the gc index
		return nil
	}
	// delete current entry from the gc index
	db.gcIndex.DeleteInBatch(batch, item)
	// update access timestamp
	item.AccessTimestamp = now()
	// update retrieve access index
	db.retrievalAccessIndex.PutInBatch(batch, item)
	// add new entry to gc index
	db.gcIndex.PutInBatch(batch, item)

	return db.shed.WriteBatch(batch)
}

// IsPinnedFileRaw checks if a given root hash is pinned as a Raw file or not.
// This infrmation is stored in pinFilesIndex as part of pinning the file
func (db *DB) IsPinnedFileRaw(addr chunk.Address) (bool, error) {
	var item shed.Item
	item.Address = addr
	i, err := db.pinFilesIndex.Get(item)
	if err != nil {
		return false, err
	}
	raw := false
	if i.IsRaw > 0 {
		raw = true
	}
	return raw, nil
}

// IsFilePinned checks if a given root hash is pinned or not.
// It check for the given root hash in the pinFilesIndex
func (db *DB) IsFilePinned(addr chunk.Address) bool {
	var item shed.Item
	item.Address = addr
	has, err := db.pinFilesIndex.Has(item)
	if err != nil {
		return false
	}
	return  has
}

// IsChunkPinned checks of a given chunk id pinned or not.
// it check for an entry in pinIndex and takes a decision.
func (db *DB) IsChunkPinned(addr chunk.Address) bool {
	var item shed.Item
	item.Address = addr
	has, err := db.pinIndex.Has(item)
	if err != nil {
		return false
	}
	return has
}

// GetPinCounterOfChunk returns the number of times a given chunk is pinned.
func (db *DB) GetPinCounterOfChunk(addr chunk.Address) (uint64, error) {
	var item shed.Item
	item.Address = addr
	i, err := db.pinIndex.Get(item)
	if err != nil {
		return 0, err
	}
	return i.PinCounter, nil
}

// GetPinFilesIndex collects all the root hashes that are pinned and returns it.
// This is used in places like ListPinFiles to get the number of chunks
// present in a pinned file.
func (db *DB) GetPinFilesIndex() map[string]uint8 {
	pinnedFiles := make(map[string]uint8)
	_ = db.pinFilesIndex.Iterate(func(item shed.Item) (stop bool, err error) {
		pinnedFiles[hex.EncodeToString(item.Address)] = item.IsRaw
		return false, nil
	}, nil)
	return pinnedFiles
}

// testHookUpdateGC is a hook that can provide
// information when a garbage collection index is updated.
var testHookUpdateGC func()

// GetAllChunksInDB is used only in testing.
// This function returns all the chunks that are present in the DB irrespective of
// whether they are pinned or not. This is used in testing as a truth data set for pinning.
func (db *DB) GetAllChunksInDB() map[string]int {
	chunksInDB := make(map[string]int)
	_ = db.retrievalDataIndex.Iterate(func(item shed.Item) (stop bool, err error) {
		chunksInDB[hex.EncodeToString(item.Address)] = len(item.Data)
		return false, nil
	}, nil)
	return chunksInDB
}
