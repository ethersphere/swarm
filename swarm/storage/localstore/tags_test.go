// Copyright 2019 The go-ethereum Authors
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
	"testing"
	"time"
)

// tests that new tag is created, iterated over (one or all) and deleted in the database
func TestTags(t *testing.T) {
	db, cleanupFunc := newTestDB(t, nil)
	defer cleanupFunc()

	tag := db.NewTag(time.Now().Unix(), "path/to/directory")

	/*	c := generateTestRandomChunkWithTags([]uint64{tag})

		err := db.Put(context.Background(), chunk.ModePutUpload, c)
		if err != nil {
			t.Fatal(err)
		}
	*/

	existingTags = db.GetTags()

	//expect tag to be in existingTags

	oneTag = db.GetTag(tag)

	// expect to exist

	//delete tag
	err = db.DeleteTag(tag)
	if err != nil {
		t.Fatal(err)
	}

	tagShouldNotExist, err := db.GetTag(tag)
	if err == nil {
		t.Fatal("tag should not exist")
	}

}

/*func TestPutTag(t *testing.T) {
	db, cleanupFunc := newTestDB(t, nil)
	defer cleanupFunc()

	wantTimestamp := time.Now().UTC().UnixNano()
	defer setNow(func() (t int64) {
		return wantTimestamp
	})()

	ch := generateTestRandomChunk()

	err := db.Put(context.Background(), chunk.ModePutTags, ch)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve indexes", newRetrieveIndexesTest(db, ch, wantTimestamp, 0))

	t.Run("pull index", newPullIndexTest(db, ch, 1, nil))

}*/
