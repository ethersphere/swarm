// Copyright 2020 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package pss

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	trojan "github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/shed"
	"github.com/ethersphere/swarm/storage/localstore"
)

// TestTrojanChunkRetrieval creates a trojan chunk
// mocks the localstore
// calls pss.Send method and verifies it's properly stored
func TestTrojanChunkRetrieval(t *testing.T) {
	var err error
	ctx := context.TODO()

	localStore := newMockLocalStore(t)

	testTargets := [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}
	payload := []byte("RECOVERY CHUNK")
	topic := trojan.NewTopic("RECOVERY")

	var ch chunk.Chunk

	pss := NewPss(localStore)

	// call Send to store trojan chunk in localstore
	if ch, err = pss.Send(ctx, testTargets, topic, payload); err != nil {
		t.Fatal(err)
	}

	// verify store, that trojan chunk has been stored correctly
	var storedChunk chunk.Chunk
	if storedChunk, err = localStore.Get(ctx, chunk.ModeGetRequest, ch.Address()); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ch, storedChunk) {
		t.Fatalf("store chunk does not match sent chunk")
	}

	// check if pinning makes a difference

}

func TestPssMonitor(t *testing.T) {
	var err error
	ctx := context.TODO()

	db, cleanupFunc := newTestDB(t, &Options{Tags: chunk.NewTags()})
	defer cleanupFunc()

	testTargets := [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}
	payload := []byte("RECOVERY CHUNK")
	topic := trojan.NewTopic("RECOVERY")

	var ch chunk.Chunk

	pss := NewPss(db.localStore)

	// call Send to store trojan chunk in localstore
	if ch, err = pss.Send(ctx, testTargets, topic, payload); err != nil {
		t.Fatal(err)
	}

	item, err := db.localStore.pullIndex.Get(shed.Item{
		Address: ch.Address(),
		BinID:   1,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.localStore.Set(context.Background(), chunk.ModeSetSyncPull, ch.Address())
	if err != nil {
		t.Fatal(err)
	}

	checkTag(t, tag, 0, 1, 0, 1, 0, 1)

}

// CheckTag checks the first tag in the api struct to be in a certain state
// TODO: reuse existing tag CheckTag instead of this
func checkTag(t *testing.T, tag *chunk.Tag, split, stored, seen, sent, synced, total int64) {
	t.Helper()
	if tag == nil {
		t.Fatal("no tag found")
	}
	tSplit := tag.Get(chunk.StateSplit)
	if tSplit != split {
		t.Fatalf("should have had split chunks, got %d want %d", tSplit, split)
	}

	tSeen := tag.Get(chunk.StateSeen)
	if tSeen != seen {
		t.Fatalf("should have had seen chunks, got %d want %d", tSeen, seen)
	}

	tStored := tag.Get(chunk.StateStored)
	if tStored != stored {
		t.Fatalf("mismatch stored chunks, got %d want %d", tStored, stored)
	}

	tSent := tag.Get(chunk.StateSent)
	if tStored != stored {
		t.Fatalf("mismatch sent chunks, got %d want %d", tSent, sent)
	}

	tSynced := tag.Get(chunk.StateSynced)
	if tSynced != synced {
		t.Fatalf("mismatch synced chunks, got %d want %d", tSynced, synced)
	}

	tTotal := tag.TotalCounter()
	if tTotal != total {
		t.Fatalf("mismatch total chunks, got %d want %d", tTotal, total)
	}
}

// newTestDB is a helper function that constructs a
// temporary database and returns a cleanup function that must
// be called to remove the data.
// TODO: refactor into common newTestDB for all test
func newTestDB(t testing.TB, o *Options) (db *DB, cleanupFunc func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "localstore-test")
	if err != nil {
		t.Fatal(err)
	}
	cleanupFunc = func() { os.RemoveAll(dir) }
	baseKey := make([]byte, 32)
	if _, err := rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}
	db, err = New(dir, baseKey, o)
	if err != nil {
		cleanupFunc()
		t.Fatal(err)
	}
	cleanupFunc = func() {
		err := db.Close()
		if err != nil {
			t.Error(err)
		}
		os.RemoveAll(dir)
	}
	return db, cleanupFunc
}

func newMockLocalStore(t *testing.T) *localstore.DB {
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	if _, err = rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}

	return localStore
}

// TODO: later test could be a simulation test for 2 nodes, localstore + netstore
