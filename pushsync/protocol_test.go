// Copyright 2019 The Swarm Authors
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

package pushsync

import (
	"context"
	"encoding/binary"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
)

type testStore struct {
	store *sync.Map
}

func (t *testStore) Put(_ context.Context, _ chunk.ModePut, chs ...chunk.Chunk) ([]bool, error) {
	exists := make([]bool, len(chs))
	for i, ch := range chs {
		idx := binary.BigEndian.Uint64(ch.Address()[:8])
		var storedCnt uint32 = 1
		if v, loaded := t.store.LoadOrStore(idx, &storedCnt); loaded {
			atomic.AddUint32(v.(*uint32), 1)
			exists[i] = loaded
		}
		log.Debug("testStore put", "idx", idx)
	}
	return exists, nil
}

// TestProtocol tests the push sync protocol
// push syncer node communicate with storers via mock loopback PubSub
func TestProtocol(t *testing.T) {
	timeout := 10 * time.Second
	chunkCnt := 1024
	tagCnt := 4
	storerCnt := 4

	sent := &sync.Map{}
	store := &sync.Map{}
	// mock pubsub messenger
	lb := newLoopBack()

	// set up a number of storers
	storers := make([]*Storer, storerCnt)
	for i := 0; i < storerCnt; i++ {
		// every chunk is closest to exactly one storer
		j := i
		isClosestTo := func(addr []byte) bool {
			n := int(binary.BigEndian.Uint64(addr[:8]))
			log.Debug("closest node?", "n", n, "n%storerCnt", n%storerCnt, "storer", j)
			return n%storerCnt == j
		}
		storers[j] = NewStorer(&testStore{store}, &testPubSub{lb, isClosestTo})
	}

	tags, tagIDs := setupTags(chunkCnt, tagCnt)
	// construct the mock push sync index iterator
	tp := newTestPushSyncIndex(chunkCnt, tagIDs, tags, sent)
	// isClosestTo function mocked
	isClosestTo := func([]byte) bool { return false }
	// start push syncing in a go routine
	p := NewPusher(tp, &testPubSub{lb, isClosestTo}, tags)
	defer p.Close()

	synced := make(map[int]int)
	for {
		select {
		case idx := <-tp.synced:
			n := synced[idx]
			synced[idx] = n + 1
			if len(synced) == chunkCnt {
				expTotal := int64(chunkCnt / tagCnt)
				checkTags(t, expTotal, tagIDs[:tagCnt-1], tags)
				for i := uint64(0); i < uint64(chunkCnt); i++ {
					if n := synced[int(i)]; n != 1 {
						t.Fatalf("expected to receive exactly 1 receipt for chunk %v, got %v", i, n)
					}
					v, ok := store.Load(i)
					if !ok {
						t.Fatalf("chunk %v not stored", i)
					}
					if cnt := *(v.(*uint32)); cnt < uint32(storerCnt) {
						t.Fatalf("chunk %v expected to be saved at least %v times, got %v", i, storerCnt, cnt)
					}
				}
				return
			}
		case <-time.After(timeout):
			t.Fatalf("timeout waiting for all chunks to be synced")
		}
	}
}
