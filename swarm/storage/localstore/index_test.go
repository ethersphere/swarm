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
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/chunk"
)

// TestDB_pullIndex validates the ordering of keys in pull index.
// Pull index key contains PO prefix which is calculated from
// DB base key and chunk address. This is not an Item field
// which are checked in Mode tests.
// This test uploads chunks, sorts them in expected order and
// validates that pull index iterator will iterate it the same
// order.
func TestDB_pullIndex(t *testing.T) {
	db, cleanupFunc := newTestDB(t, nil)
	defer cleanupFunc()

	uploader := db.NewPutter(ModePutUpload)

	chunkCount := 50

	chunks := make([]testIndexChunk, chunkCount)

	// upload random chunks
	for i := 0; i < chunkCount; i++ {
		chunk := generateTestRandomChunk()

		err := uploader.Put(chunk)
		if err != nil {
			t.Fatal(err)
		}

		chunks[i] = testIndexChunk{
			Chunk: chunk,
			// this timestamp is not the same as in
			// the index, but given that uploads
			// are sequential and that only ordering
			// of events matter, this information is
			// sufficient
			storeTimestamp: now(),
		}
	}

	testItemsOrder(t, db.pullIndex, chunks, func(i, j int) (less bool) {
		poi := chunk.Proximity(db.baseKey, chunks[i].Address())
		poj := chunk.Proximity(db.baseKey, chunks[j].Address())
		if poi < poj {
			return true
		}
		if poi > poj {
			return false
		}
		if chunks[i].storeTimestamp < chunks[j].storeTimestamp {
			return true
		}
		if chunks[i].storeTimestamp > chunks[j].storeTimestamp {
			return false
		}
		return bytes.Compare(chunks[i].Address(), chunks[j].Address()) == -1
	})
}

// TestDB_gcIndex validates garbage collection index by uploading
// a chunk with and performing operations using synced, access and
// request modes.
func TestDB_gcIndex(t *testing.T) {
	db, cleanupFunc := newTestDB(t, nil)
	defer cleanupFunc()

	uploader := db.NewPutter(ModePutUpload)

	chunkCount := 50

	chunks := make([]testIndexChunk, chunkCount)

	// upload random chunks
	for i := 0; i < chunkCount; i++ {
		chunk := generateTestRandomChunk()

		err := uploader.Put(chunk)
		if err != nil {
			t.Fatal(err)
		}

		chunks[i] = testIndexChunk{
			Chunk: chunk,
		}
	}

	// check if all chunks are stored
	newItemsCountTest(db.pullIndex, chunkCount)(t)

	// check that chunks are not collectable for garbage
	newItemsCountTest(db.gcIndex, 0)(t)

	// set update gc test hook to signal when
	// update gc goroutine is done by sending to
	// testHookUpdateGCChan channel, which is
	// used to wait for indexes change verifications
	testHookUpdateGCChan := make(chan struct{})
	defer setTestHookUpdateGC(func() {
		testHookUpdateGCChan <- struct{}{}
	})()

	t.Run("request unsynced", func(t *testing.T) {
		chunk := chunks[1]

		_, err := db.NewGetter(ModeGetRequest).Get(chunk.Address())
		if err != nil {
			t.Fatal(err)
		}
		// wait for update gc goroutine to be done
		<-testHookUpdateGCChan

		// the chunk is not synced
		// should not be in the garbace collection index
		newItemsCountTest(db.gcIndex, 0)(t)

		newIndexGCSizeTest(db)(t)
	})

	t.Run("sync one chunk", func(t *testing.T) {
		chunk := chunks[0]

		err := db.NewSetter(ModeSetSync).Set(chunk.Address())
		if err != nil {
			t.Fatal(err)
		}

		// the chunk is synced and should be in gc index
		newItemsCountTest(db.gcIndex, 1)(t)

		newIndexGCSizeTest(db)(t)
	})

	t.Run("sync all chunks", func(t *testing.T) {
		setter := db.NewSetter(ModeSetSync)

		for i := range chunks {
			err := setter.Set(chunks[i].Address())
			if err != nil {
				t.Fatal(err)
			}
		}

		testItemsOrder(t, db.gcIndex, chunks, nil)

		newIndexGCSizeTest(db)(t)
	})

	t.Run("request one chunk", func(t *testing.T) {
		i := 6

		_, err := db.NewGetter(ModeGetRequest).Get(chunks[i].Address())
		if err != nil {
			t.Fatal(err)
		}
		// wait for update gc goroutine to be done
		<-testHookUpdateGCChan

		// move the chunk to the end of the expected gc
		c := chunks[i]
		chunks = append(chunks[:i], chunks[i+1:]...)
		chunks = append(chunks, c)

		testItemsOrder(t, db.gcIndex, chunks, nil)

		newIndexGCSizeTest(db)(t)
	})

	t.Run("random chunk request", func(t *testing.T) {
		requester := db.NewGetter(ModeGetRequest)

		rand.Shuffle(len(chunks), func(i, j int) {
			chunks[i], chunks[j] = chunks[j], chunks[i]
		})

		for _, chunk := range chunks {
			_, err := requester.Get(chunk.Address())
			if err != nil {
				t.Fatal(err)
			}
			// wait for update gc goroutine to be done
			<-testHookUpdateGCChan
		}

		testItemsOrder(t, db.gcIndex, chunks, nil)

		newIndexGCSizeTest(db)(t)
	})

	t.Run("remove one chunk", func(t *testing.T) {
		i := 3

		err := db.NewSetter(modeSetRemove).Set(chunks[i].Address())
		if err != nil {
			t.Fatal(err)
		}

		// remove the chunk from the expected chunks in gc index
		chunks = append(chunks[:i], chunks[i+1:]...)

		testItemsOrder(t, db.gcIndex, chunks, nil)

		newIndexGCSizeTest(db)(t)
	})
}

func TestDommy(t *testing.T) {
	c := common.FromHex("0ed0025f8759f09073564aceebea54231a291bb1023e5e12afefb64fcebc9bac")

	//sp0 := common.FromHex("d29e6d")
	sp4 := common.FromHex("3a028e")
	//sp11 := common.FromHex("531e93")
	//sp16 := common.FromHex("64c63a")
	sp1 := common.FromHex("1f21c3")
	sp6 := common.FromHex("16c150")
	sp7 := common.FromHex("1c995f")

	prox1 := chunk.Proximity(sp1, c)
	fmt.Println("prox1:", prox1)

	prox6 := chunk.Proximity(sp6, c)
	fmt.Println("prox6:", prox6)

	prox7 := chunk.Proximity(sp7, c)
	fmt.Println("prox7:", prox7)

	prox4 := chunk.Proximity(sp4, c)
	fmt.Println("prox4:", prox4)

	// prox for all 1 6 and 7: 3
	// depth: 3

	//$ kubectl -n tony exec -it swarm-private-1 -- ./geth attach /root/.ethereum/bzzd.ipc --exec="console.log(bzz.hive)"

	//=========================================================================
	//commit hash: 789bc8662
	//Mon Apr  1 08:25:12 UTC 2019 K???MLI? hive: queen's address: 1f21c3
	//population: 14 (25), NeighbourhoodSize: 2, MinBinSize: 2, MaxBinSize: 4
	//000  3 ce23 ecbf afd2               | 14 ecbf (0) ce23 (0) d29e (0) d334 (0)
	//001  8 42a6 531e 578c 64c6          |  8 578c (0) 531e (0) 42a6 (0) 66bb (0)
	//002  1 3a02                         |  1 3a02 (0)
	//============ DEPTH: 3 ==========================================
	//003  0                              |  0
	//004  1 16c1                         |  1 16c1 (0)
	//005  0                              |  0
	//006  1 1c99                         |  1 1c99 (0)
	//007  0                              |  0
	//008  0                              |  0
	//009  0                              |  0
	//010  0                              |  0
	//011  0                              |  0
	//012  0                              |  0
	//013  0                              |  0
	//014  0                              |  0
	//015  0                              |  0
	//=========================================================================
	//undefined

}
