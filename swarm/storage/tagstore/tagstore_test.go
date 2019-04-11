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

package tagstore

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/swarm/state"
)

// TestDB validates if the chunk can be uploaded and
// correctly retrieved.
func TestTagPersistence(t *testing.T) {
	testDb, cleanup := newTestDb(t)
	defer cleanup()

	store, err := New(testDb)

	timeNow := time.Now().Unix()
	// should create and persist the tag
	tag, err := store.NewTag(timeNow, "test/upload")
	if err != nil {
		t.Fatal(err)
	}
	persistedTag, err := store.Load(tag.GetUid())
	if err != nil {
		t.Fatal(err)
	}
	if persistedTag.GetUid() != tag.GetUid() {
		t.Fatalf("persisted tag and created tag uids not equal. want %d got %d", tag.GetUid(), persistedTag.GetUid())
	}
	if persistedTag.GetName() != tag.GetName() {
		t.Fatalf("tag names dont match. want '%s' got '%s'", tag.GetName(), persistedTag.GetName())
	}
}

func newTestDb(t *testing.T) (state.Store, func()) {
	t.Helper()
	dir, err := ioutil.TempDir("", "tagstore-test")
	if err != nil {
		t.Fatal(err)
	}

	store, err := state.NewDBStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	cleanupFunc := func() {
		err := store.Close()
		if err != nil {
			t.Error(err)
		}
		os.RemoveAll(dir)
	}

	return store, cleanupFunc

}
