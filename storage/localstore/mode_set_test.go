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
	"testing"
	"time"

	"github.com/ethersphere/swarm/chunk"
	"github.com/syndtr/goleveldb/leveldb"
)

// TestModeSetAccess validates ModeSetAccess index values on the provided DB.
func TestModeSetAccess(t *testing.T) {
	for _, tc := range multiChunkTestCases {
		t.Run(tc.name, func(t *testing.T) {
			db, cleanupFunc := newTestDB(t, nil)
			defer cleanupFunc()

			chunks := generateTestRandomChunks(tc.count)

			wantTimestamp := time.Now().UTC().UnixNano()
			defer setNow(func() (t int64) {
				return wantTimestamp
			})()

			err := db.Set(context.Background(), chunk.ModeSetAccess, chunkAddresses(chunks)...)
			if err != nil {
				t.Fatal(err)
			}

			binIDs := make(map[uint8]uint64)

			for _, ch := range chunks {
				po := db.po(ch.Address())
				binIDs[po]++

				newPullIndexTest(db, ch, binIDs[po], nil)(t)
				newGCIndexTest(db, ch, wantTimestamp, wantTimestamp, binIDs[po], nil)(t)
			}

			t.Run("gc index count", newItemsCountTest(db.gcIndex, tc.count))

			t.Run("pull index count", newItemsCountTest(db.pullIndex, tc.count))

			t.Run("gc size", newIndexGCSizeTest(db))
		})
	}
}

// TestModeSetSyncPull validates ModeSetSyncPull index values on the provided DB.
func TestModeSetSyncPull(t *testing.T) {
	defer func(f func() uint32) {
		chunk.TagUidFunc = f
	}(chunk.TagUidFunc)

	chunk.TagUidFunc = func() uint32 { return 0 }

	for _, mtc := range []struct {
		name            string
		anonymous       bool
		expErrPushIndex error
		expErrGCIndex   error
	}{
		{
			name:            "normal tag",
			anonymous:       false,
			expErrPushIndex: nil,
			expErrGCIndex:   leveldb.ErrNotFound,
		},
		{
			name:            "anonymous tag",
			anonymous:       true,
			expErrPushIndex: leveldb.ErrNotFound,
			expErrGCIndex:   nil,
		},
	} {
		t.Run(mtc.name, func(t *testing.T) {
			for _, tc := range multiChunkTestCases {
				t.Run(tc.name, func(t *testing.T) {
					db, cleanupFunc := newTestDB(t, &Options{Tags: chunk.NewTags()})
					defer cleanupFunc()

					tag, err := db.tags.Create(mtc.name, int64(tc.count), mtc.anonymous)
					if err != nil {
						t.Fatal(err)
					}
					if tag.Uid != 0 {
						t.Fatal("expected mock tag uid")
					}

					chunks := generateTestRandomChunks(tc.count)

					wantTimestamp := time.Now().UTC().UnixNano()
					defer setNow(func() (t int64) {
						return wantTimestamp
					})()

					_, err = db.Put(context.Background(), chunk.ModePutUpload, chunks...)
					if err != nil {
						t.Fatal(err)
					}

					err = db.Set(context.Background(), chunk.ModeSetSyncPull, chunkAddresses(chunks)...)
					if err != nil {
						t.Fatal(err)
					}

					binIDs := make(map[uint8]uint64)

					for _, ch := range chunks {
						po := db.po(ch.Address())
						binIDs[po]++

						newRetrieveIndexesTestWithAccess(db, ch, wantTimestamp, wantTimestamp)(t)
						newPullIndexTest(db, ch, binIDs[po], nil)(t)
						newPushIndexTest(db, ch, wantTimestamp, mtc.expErrPushIndex)(t)
						newGCIndexTest(db, ch, wantTimestamp, wantTimestamp, binIDs[po], mtc.expErrGCIndex)(t)

						// if the upload is anonymous then we expect to see some values in the gc index
						if mtc.anonymous {
							t.Run("gc index count", newItemsCountTest(db.gcIndex, tc.count))
						}
					}

					t.Run("gc size", newIndexGCSizeTest(db))
				})
			}
		})
	}
}

// TestModeSetSyncPush validates ModeSetSyncPush index values on the provided DB.
func TestModeSetSyncPush(t *testing.T) {
	defer func(f func() uint32) {
		chunk.TagUidFunc = f
	}(chunk.TagUidFunc)

	chunk.TagUidFunc = func() uint32 { return 0 }

	for _, mtc := range []struct {
		name      string
		anonymous bool
		expErr    error
	}{
		{
			name:      "normal tag",
			anonymous: false,
			expErr:    nil,
		},
		{
			name:      "anonymous tag",
			anonymous: true,
			expErr:    leveldb.ErrNotFound,
		},
	} {
		t.Run(mtc.name, func(t *testing.T) {
			for _, tc := range multiChunkTestCases {
				t.Run(tc.name, func(t *testing.T) {
					db, cleanupFunc := newTestDB(t, &Options{Tags: chunk.NewTags()})
					defer cleanupFunc()

					tag, err := db.tags.Create(mtc.name, int64(tc.count), mtc.anonymous)
					if err != nil {
						t.Fatal(err)
					}
					if tag.Uid != 0 {
						t.Fatal("expected mock tag uid")
					}

					chunks := generateTestRandomChunks(tc.count)

					wantTimestamp := time.Now().UTC().UnixNano()
					defer setNow(func() (t int64) {
						return wantTimestamp
					})()

					_, err = db.Put(context.Background(), chunk.ModePutUpload, chunks...)
					if err != nil {
						t.Fatal(err)
					}

					err = db.Set(context.Background(), chunk.ModeSetSyncPush, chunkAddresses(chunks)...)
					if err != nil {
						t.Fatal(err)
					}

					binIDs := make(map[uint8]uint64)

					for _, ch := range chunks {
						po := db.po(ch.Address())
						binIDs[po]++

						newRetrieveIndexesTestWithAccess(db, ch, wantTimestamp, wantTimestamp)(t)
						newPullIndexTest(db, ch, binIDs[po], nil)(t)
						newPushIndexTest(db, ch, wantTimestamp, leveldb.ErrNotFound)(t)
						newGCIndexTest(db, ch, wantTimestamp, wantTimestamp, binIDs[po], nil)(t)
					}

					t.Run("gc index count", newItemsCountTest(db.gcIndex, tc.count))

					t.Run("gc size", newIndexGCSizeTest(db))
				})
			}
		})
	}
}

// TestModeSetRemove validates ModeSetRemove index values on the provided DB.
func TestModeSetRemove(t *testing.T) {
	for _, tc := range multiChunkTestCases {
		t.Run(tc.name, func(t *testing.T) {
			db, cleanupFunc := newTestDB(t, nil)
			defer cleanupFunc()

			chunks := generateTestRandomChunks(tc.count)

			_, err := db.Put(context.Background(), chunk.ModePutUpload, chunks...)
			if err != nil {
				t.Fatal(err)
			}

			err = db.Set(context.Background(), chunk.ModeSetRemove, chunkAddresses(chunks)...)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("retrieve indexes", func(t *testing.T) {
				for _, ch := range chunks {
					wantErr := leveldb.ErrNotFound
					_, err := db.retrievalDataIndex.Get(addressToItem(ch.Address()))
					if err != wantErr {
						t.Errorf("got error %v, want %v", err, wantErr)
					}

					// access index should not be set
					_, err = db.retrievalAccessIndex.Get(addressToItem(ch.Address()))
					if err != wantErr {
						t.Errorf("got error %v, want %v", err, wantErr)
					}
				}

				t.Run("retrieve data index count", newItemsCountTest(db.retrievalDataIndex, 0))

				t.Run("retrieve access index count", newItemsCountTest(db.retrievalAccessIndex, 0))
			})

			for _, ch := range chunks {
				newPullIndexTest(db, ch, 0, leveldb.ErrNotFound)(t)
			}

			t.Run("pull index count", newItemsCountTest(db.pullIndex, 0))

			t.Run("gc index count", newItemsCountTest(db.gcIndex, 0))

			t.Run("gc size", newIndexGCSizeTest(db))
		})
	}
}
