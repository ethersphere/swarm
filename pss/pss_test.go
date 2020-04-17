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
	"time"

	"github.com/ethersphere/swarm/chunk"
	tagtesting "github.com/ethersphere/swarm/chunk/testing"
	trojan "github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage/localstore"
)

// TestTrojanChunkRetrieval creates a trojan chunk
// mocks the localstore
// calls pss.Send method and verifies it's properly stored
func TestTrojanChunkRetrieval(t *testing.T) {
	var err error
	ctx := context.TODO()
	tags := chunk.NewTags()

	localStore := newMockLocalStore(t, tags)

	testTargets := [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}
	payload := []byte("RECOVERY CHUNK")
	topic := trojan.NewTopic("RECOVERY")

	var monitor *Monitor

	pss := NewPss(localStore, tags)

	// call Send to store trojan chunk in localstore
	if monitor, err = pss.Send(ctx, testTargets, topic, payload); err != nil {
		t.Fatal(err)
	}

	// verify store, that trojan chunk has been stored correctly
	var storedChunk chunk.Chunk
	if storedChunk, err = localStore.Get(ctx, chunk.ModeGetRequest, monitor.chunk.Address()); err != nil {
		t.Fatal(err)
	}
	storedChunk = storedChunk.WithTagID(monitor.chunk.TagID())

	if !reflect.DeepEqual(monitor.chunk, storedChunk) {
		t.Fatalf("store chunk does not match sent chunk")
	}

	// check if pinning makes a difference

}

func TestPssMonitor(t *testing.T) {
	var err error
	ctx := context.TODO()
	tags := chunk.NewTags()
	timeout := 10 * time.Second

	localStore := newMockLocalStore(t, tags)

	testTargets := [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}
	payload := []byte("RECOVERY CHUNK")
	topic := trojan.NewTopic("RECOVERY")

	var monitor *Monitor

	pss := NewPss(localStore, tags)

	// call Send to store trojan chunk in localstore
	if monitor, err = pss.Send(ctx, testTargets, topic, payload); err != nil {
		t.Fatal(err)
	}

loop:
	for {
		select {
		case state := <-monitor.state:
			if state == chunk.StateStored {
				t.Log("message has been stored")
			}
			if state == chunk.StateSent {
				t.Log("message has been sent")
			}
			if state == chunk.StateSynced {
				t.Log("message has been synced")
			}
		case <-time.After(timeout):
			t.Log("no message received")
			close(monitor.state)
			break loop
		}
	}

	// we expect the chunk to be stored in localstore
	// with a sent status
	// and the total amount of chunk to be 1
	var split, seen, synced, stored, sent, total int64 = 0, 0, 0, 1, 1, 1

	// verifies if tag has been stored, sent and the total count of chunks
	tagtesting.CheckTag(t, monitor.tag, split, stored, seen, sent, synced, total)

}

func newMockLocalStore(t *testing.T, tags *chunk.Tags) *localstore.DB {
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	if _, err = rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	localStore, err := localstore.New(dir, baseKey, &localstore.Options{Tags: tags})
	if err != nil {
		t.Fatal(err)
	}

	return localStore
}

// TODO: later test could be a simulation test for 2 nodes, localstore + netstore
