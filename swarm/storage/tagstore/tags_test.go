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

package tagstore

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/swarm/chunk"
)

// tests that new tag is created, iterated over (one or all) and deleted in the database
func TestTags(t *testing.T) {
	db, cleanupFunc := newTestDB(t, nil)
	defer cleanupFunc()
	timeNow := time.Now().Unix()
	testTags := []struct {
		tag  uint32
		path string
	}{
		{path: "path/to/dir1"},
		{path: "path/to/dir2"},
		{path: "another/path"},
	}

	tagMap := make(map[uint32]string)

	for _, v := range testTags {
		localTag, err := db.NewTag(timeNow, v.path)
		if err != nil {
			t.Fatal(err)
		}
		tagMap[localTag] = v.path
	}

	existingTags, err := db.GetTags()
	if err != nil {
		t.Fatal(err)
	}

	existingTags.Range(func(k, v interface{}) bool {
		keyVal := k.(uint32)
		vv := v.(*chunk.Tag)
		if vv.GetName() != tagMap[keyVal] {
			t.Fatal("tag not equal")
		}
		return true
	})

	//expect tag to be in existingTags

	//oneTag, err := db.GetTag(tag)
	if err != nil {
		t.Fatal(err)
	}
	// expect to exist

	//delete tag
	/*err = db.DeleteTag(tag)
	if err != nil {
		t.Fatal(err)
	}

	tagShouldNotExist, err := db.GetTag(tag)
	if err == nil {
		t.Fatal("tag should not exist")
	}
	*/
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
