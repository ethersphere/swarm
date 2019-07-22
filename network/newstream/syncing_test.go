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

package newstream

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

var timeout = 30 * time.Second

// TestTwoNodesFullSync connects two nodes, uploads content to one node and expects the
// uploader node's chunks to be synced to the second node. This is expected behaviour since although
// both nodes might share address bits, due to kademlia depth=0 when under ProxBinSize - this will
// eventually create subscriptions on all bins between the two nodes, causing a full sync between them
// The test checks that:
// 1. All subscriptions are created
// 2. All chunks are transferred from one node to another (asserted by summing and comparing bin indexes on both nodes)
func TestTwoNodesFullSync(t *testing.T) {
	const chunkCount = 10000
	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newSyncSimServiceFunc(nil),
	})

	defer sim.Close()
	defer catchDuplicateChunkSync(t)()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	uploaderNode, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("pivot node", "enode", uploaderNode)
	uploadStore := sim.NodeItem(uploaderNode, bucketKeyFileStore).(chunk.Store)

	chunks := mustUploadChunks(ctx, t, uploadStore, chunkCount)

	uploaderNodeBinIDs := make([]uint64, 17)
	uploaderStore := sim.NodeItem(uploaderNode, bucketKeyFileStore).(chunk.Store)
	var uploaderSum uint64
	for po := 0; po <= 16; po++ {
		until, err := uploaderStore.LastPullSubscriptionBinID(uint8(po))
		if err != nil {
			t.Fatal(err)
		}
		uploaderNodeBinIDs[po] = until
		uploaderSum += until
	}

	if uint64(len(chunks)) != uploaderSum {
		t.Fatalf("uploader node chunk number mismatch. got %d want %d", len(chunks), uploaderSum)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		syncingNode, err := sim.AddNode()
		if err != nil {
			return err
		}
		log.Debug("syncing node", "enode", syncingNode)

		err = sim.Net.ConnectNodesChain([]enode.ID{syncingNode, uploaderNode})
		if err != nil {
			return err
		}
		return waitChunks(sim.NodeItem(syncingNode, bucketKeyFileStore).(chunk.Store), uploaderSum, 10*time.Second)
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

func TestTwoNodesSyncWithGaps(t *testing.T) {
	removeChunks := func(t *testing.T, ctx context.Context, store chunk.Store, gaps [][2]uint64, chunks []chunk.Address) (removedCount uint64) {
		t.Helper()

		for _, gap := range gaps {
			for i := gap[0]; i < gap[1]; i++ {
				c := chunks[i]
				if err := store.Set(ctx, chunk.ModeSetRemove, c); err != nil {
					t.Fatal(err)
				}
				removedCount++
			}
		}
		return removedCount
	}

	for _, tc := range []struct {
		name           string
		chunkCount     uint64
		gaps           [][2]uint64
		liveChunkCount uint64
		liveGaps       [][2]uint64
	}{
		{
			name:       "no gaps",
			chunkCount: 100,
			gaps:       nil,
		},
		{
			name:       "first chunk removed",
			chunkCount: 100,
			gaps:       [][2]uint64{{0, 1}},
		},
		{
			name:       "one chunk removed",
			chunkCount: 100,
			gaps:       [][2]uint64{{60, 61}},
		},
		{
			name:       "single gap at start",
			chunkCount: 100,
			gaps:       [][2]uint64{{0, 5}},
		},
		{
			name:       "single gap",
			chunkCount: 100,
			gaps:       [][2]uint64{{5, 10}},
		},
		{
			name:       "multiple gaps",
			chunkCount: 100,
			gaps:       [][2]uint64{{0, 1}, {10, 21}},
		},
		{
			name:       "big gaps",
			chunkCount: 100,
			gaps:       [][2]uint64{{0, 1}, {10, 21}, {50, 91}},
		},
		{
			name:       "remove all",
			chunkCount: 100,
			gaps:       [][2]uint64{{0, 100}},
		},
		{
			name:       "large db",
			chunkCount: 4000,
		},
		{
			name:       "large db with gap",
			chunkCount: 4000,
			gaps:       [][2]uint64{{1000, 3000}},
		},
		{
			name:           "live",
			liveChunkCount: 100,
		},
		{
			name:           "live and history",
			chunkCount:     100,
			liveChunkCount: 100,
		},
		{
			name:           "live and history with history gap",
			chunkCount:     100,
			gaps:           [][2]uint64{{5, 10}},
			liveChunkCount: 100,
		},
		{
			name:           "live and history with live gap",
			chunkCount:     100,
			liveChunkCount: 100,
			liveGaps:       [][2]uint64{{105, 110}},
		},
		{
			name:           "live and history with gaps",
			chunkCount:     100,
			gaps:           [][2]uint64{{5, 10}},
			liveChunkCount: 100,
			liveGaps:       [][2]uint64{{105, 110}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
				"bzz-sync": newSyncSimServiceFunc(nil),
			})
			defer sim.Close()
			defer catchDuplicateChunkSync(t)()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			uploadNode, err := sim.AddNode()
			if err != nil {
				t.Fatal(err)
			}

			uploadStore := sim.NodeItem(uploadNode, bucketKeyFileStore).(chunk.Store)

			chunks := mustUploadChunks(ctx, t, uploadStore, tc.chunkCount)

			totalChunkCount, err := getChunkCount(uploadStore)
			if err != nil {
				t.Fatal(err)
			}

			if totalChunkCount != tc.chunkCount {
				t.Errorf("uploaded %v chunks, want %v", totalChunkCount, tc.chunkCount)
			}

			removedCount := removeChunks(t, ctx, uploadStore, tc.gaps, chunks)

			syncNode, err := sim.AddNode()
			if err != nil {
				t.Fatal(err)
			}
			err = sim.Net.Connect(uploadNode, syncNode)
			if err != nil {
				t.Fatal(err)
			}

			syncStore := sim.NodeItem(syncNode, bucketKeyFileStore).(chunk.Store)

			err = waitChunks(syncStore, totalChunkCount-removedCount, 10*time.Second)
			if err != nil {
				t.Fatal(err)
			}

			if tc.liveChunkCount > 0 {
				chunks = append(chunks, mustUploadChunks(ctx, t, uploadStore, tc.liveChunkCount)...)

				totalChunkCount, err = getChunkCount(uploadStore)
				if err != nil {
					t.Fatal(err)
				}

				if want := tc.chunkCount + tc.liveChunkCount; totalChunkCount != want {
					t.Errorf("uploaded %v chunks, want %v", totalChunkCount, want)
				}

				removedCount += removeChunks(t, ctx, uploadStore, tc.liveGaps, chunks)

				err = waitChunks(syncStore, totalChunkCount-removedCount, time.Minute)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// TestTwoNodesFullSyncLive brings up two nodes, connects them, adds chunkCount
// * 4096 bytes to its localstore, then validates that all chunks are synced.
func TestTwoNodesFullSyncLive(t *testing.T) {
	const (
		chunkCount = 20000
	)

	defer catchDuplicateChunkSync(t)()

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newSyncSimServiceFunc(nil),
	})
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		uploaderNode, err := sim.AddNode()
		if err != nil {
			return err
		}
		uploaderNodeStore := sim.NodeItem(uploaderNode, bucketKeyFileStore).(*storage.FileStore)

		syncingNode, err := sim.AddNode()
		if err != nil {
			return err
		}
		err = sim.Net.Connect(syncingNode, uploaderNode)
		if err != nil {
			return err
		}

		uploaderSum, err := getChunkCount(uploaderNodeStore.ChunkStore)
		if err != nil {
			return err
		}
		if uploaderSum != 0 {
			return fmt.Errorf("got not empty uploader chunk store")
		}

		mustUploadChunks(ctx, t, uploaderNodeStore, chunkCount)

		// count the content in the bins again
		uploadedChunks, err := getChunks(uploaderNodeStore.ChunkStore)
		if err != nil {
			return err
		}
		if len(uploadedChunks) == 0 {
			return fmt.Errorf("got empty uploader chunk store after uploading")
		}

		// wait for all chunks to be synced
		syncingNodeStore := sim.NodeItem(syncingNode, bucketKeyFileStore).(chunk.Store)
		if err := waitChunks(syncingNodeStore, uint64(len(uploadedChunks)), 10*time.Second); err != nil {
			return err
		}

		// validate that all and only all chunks are synced
		syncedChunks, err := getChunks(syncingNodeStore)
		if err != nil {
			return err
		}
		for c := range uploadedChunks {
			if _, ok := syncedChunks[c]; !ok {
				return fmt.Errorf("missing chunk %v", c)
			}
			delete(uploadedChunks, c)
			delete(syncedChunks, c)
		}
		if len(uploadedChunks) != 0 {
			return fmt.Errorf("some of the uploaded chunks are not synced")
		}
		if len(syncedChunks) != 0 {
			return fmt.Errorf("some of the synced chunks are not of uploaded ones")
		}
		return nil
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestTwoNodesFullSyncHistoryAndLive brings up one node, adds chunkCount * 4096
// bytes to its localstore for historical data, then validates that all chunks
// are synced to the newly connected node. After that it adds another chunkCount
// * 4096 bytes to its localstore and validates that all live chunks are synced.
func TestTwoNodesFullSyncHistoryAndLive(t *testing.T) {
	const (
		chunkCount = 10000 // per history and per live upload
	)

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newSyncSimServiceFunc(nil),
	})
	defer sim.Close()

	defer catchDuplicateChunkSync(t)()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		uploaderNode, err := sim.AddNode()
		if err != nil {
			return err
		}
		uploaderNodeStore := nodeFileStore(sim, uploaderNode)

		// add chunks for historical syncing before new node connection
		mustUploadChunks(ctx, t, uploaderNodeStore, chunkCount)

		// add another node
		syncingNode, err := sim.AddNode()
		if err != nil {
			return err
		}
		syncingNodeStore := sim.NodeItem(syncingNode, bucketKeyFileStore).(chunk.Store)
		// and connect it with the uploading node
		err = sim.Net.Connect(syncingNode, uploaderNode)
		if err != nil {
			return err
		}

		// get total number of chunks
		uploaderSum, err := getChunkCount(uploaderNodeStore.ChunkStore)
		if err != nil {
			return err
		}
		// and wait for them to be synced
		if err := waitChunks(syncingNodeStore, uploaderSum, 10*time.Second); err != nil {
			return err
		}

		// get all chunks from the uploading node
		uploadedChunks, err := getChunks(uploaderNodeStore.ChunkStore)
		if err != nil {
			return err
		}
		if len(uploadedChunks) == 0 {
			return fmt.Errorf("got empty uploader chunk store after uploading")
		}
		// get all chunks from the syncing node
		syncedChunks, err := getChunks(syncingNodeStore)
		if err != nil {
			return err
		}
		// validate that all and only all chunks are synced
		// while keeping the record of them for live syncing validation
		historicalChunks := make(map[string]struct{})
		for c := range uploadedChunks {
			if _, ok := syncedChunks[c]; !ok {
				return fmt.Errorf("missing chunk %v", c)
			}
			delete(uploadedChunks, c)
			delete(syncedChunks, c)
			historicalChunks[c] = struct{}{}
		}
		if len(uploadedChunks) != 0 {
			return fmt.Errorf("some of the uploaded historical chunks are not synced")
		}
		if len(syncedChunks) != 0 {
			return fmt.Errorf("some of the synced historical chunks are not of uploaded ones")
		}

		// upload chunks for live syncing
		mustUploadChunks(ctx, t, uploaderNodeStore, chunkCount)

		// get all chunks from the uploader node
		uploadedChunks, err = getChunks(uploaderNodeStore.ChunkStore)
		if err != nil {
			return err
		}
		if len(uploadedChunks) == 0 {
			return fmt.Errorf("got empty uploader chunk store after uploading")
		}

		// wait for all chunks to be synced
		if err := waitChunks(syncingNodeStore, uint64(len(uploadedChunks)), 10*time.Second); err != nil {
			return err
		}

		// get all chunks from the syncing node
		syncedChunks, err = getChunks(syncingNodeStore)
		if err != nil {
			return err
		}
		// remove historical chunks from total uploaded and synced chunks
		for c := range historicalChunks {
			if _, ok := uploadedChunks[c]; !ok {
				return fmt.Errorf("missing uploaded historical chunk: %s", c)
			}
			delete(uploadedChunks, c)
			if _, ok := syncedChunks[c]; !ok {
				return fmt.Errorf("missing synced historical chunk: %s", c)
			}
			delete(syncedChunks, c)
		}
		// validate that all and only all live chunks are synced
		for c := range uploadedChunks {
			if _, ok := syncedChunks[c]; !ok {
				return fmt.Errorf("missing chunk %v", c)
			}
			delete(uploadedChunks, c)
			delete(syncedChunks, c)
		}
		if len(uploadedChunks) != 0 {
			return fmt.Errorf("some of the uploaded live chunks are not synced")
		}
		if len(syncedChunks) != 0 {
			return fmt.Errorf("some of the synced live chunks are not of uploaded ones")
		}
		return nil
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestFullSync performs a series of subtests where a number of nodes are
// connected to the single (chunk uploading) node.
func TestFullSync(t *testing.T) {
	for _, tc := range []struct {
		name          string
		chunkCount    uint64
		syncNodeCount int
		history       bool
		live          bool
	}{
		{
			name:          "sync to two nodes history",
			chunkCount:    5000,
			syncNodeCount: 2,
			history:       true,
		},
		{
			name:          "sync to two nodes live",
			chunkCount:    5000,
			syncNodeCount: 2,
			live:          true,
		},
		{
			name:          "sync to two nodes history and live",
			chunkCount:    2500,
			syncNodeCount: 2,
			history:       true,
			live:          true,
		},
		{
			name:          "sync to 50 nodes history",
			chunkCount:    500,
			syncNodeCount: 50,
			history:       true,
		},
		{
			name:          "sync to 50 nodes live",
			chunkCount:    500,
			syncNodeCount: 50,
			live:          true,
		},
		{
			name:          "sync to 50 nodes history and live",
			chunkCount:    250,
			syncNodeCount: 50,
			history:       true,
			live:          true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
				"bzz-sync": newSyncSimServiceFunc(nil),
			})
			defer sim.Close()

			defer catchDuplicateChunkSync(t)()

			uploaderNode, err := sim.AddNode()
			if err != nil {
				t.Fatal(err)
			}
			uploaderNodeStore := sim.NodeItem(uploaderNode, bucketKeyFileStore).(*storage.FileStore)

			if tc.history {
				mustUploadChunks(context.Background(), t, uploaderNodeStore, tc.chunkCount)
			}

			// add nodes to sync to
			ids, err := sim.AddNodes(tc.syncNodeCount)
			if err != nil {
				t.Fatal(err)
			}
			// connect every new node to the uploading one, so
			// every node will have depth 0 as only uploading node
			// will be in their kademlia tables
			err = sim.Net.ConnectNodesStar(ids, uploaderNode)
			if err != nil {
				t.Fatal(err)
			}

			// count the content in the bins again
			uploadedChunks, err := getChunks(uploaderNodeStore.ChunkStore)
			if err != nil {
				t.Fatal(err)
			}
			if tc.history && len(uploadedChunks) == 0 {
				t.Errorf("got empty uploader chunk store")
			}
			if !tc.history && len(uploadedChunks) != 0 {
				t.Errorf("got non empty uploader chunk store")
			}

			historicalChunks := make(map[enode.ID]map[string]struct{})
			for _, id := range ids {
				wantChunks := make(map[string]struct{}, len(uploadedChunks))
				for k, v := range uploadedChunks {
					wantChunks[k] = v
				}
				// wait for all chunks to be synced
				store := sim.NodeItem(id, bucketKeyFileStore).(chunk.Store)
				if err := waitChunks(store, uint64(len(wantChunks)), 10*time.Second); err != nil {
					t.Fatal(err)
				}

				// validate that all and only all chunks are synced
				syncedChunks, err := getChunks(store)
				if err != nil {
					t.Fatal(err)
				}
				historicalChunks[id] = make(map[string]struct{})
				for c := range wantChunks {
					if _, ok := syncedChunks[c]; !ok {
						t.Errorf("missing chunk %v", c)
					}
					delete(wantChunks, c)
					delete(syncedChunks, c)
					historicalChunks[id][c] = struct{}{}
				}
				if len(wantChunks) != 0 {
					t.Errorf("some of the uploaded chunks are not synced")
				}
				if len(syncedChunks) != 0 {
					t.Errorf("some of the synced chunks are not of uploaded ones")
				}
			}

			if tc.live {
				mustUploadChunks(context.Background(), t, uploaderNodeStore, tc.chunkCount)
			}

			uploadedChunks, err = getChunks(uploaderNodeStore.ChunkStore)
			if err != nil {
				t.Fatal(err)
			}

			for _, id := range ids {
				wantChunks := make(map[string]struct{}, len(uploadedChunks))
				for k, v := range uploadedChunks {
					wantChunks[k] = v
				}
				store := sim.NodeItem(id, bucketKeyFileStore).(chunk.Store)
				// wait for all chunks to be synced
				if err := waitChunks(store, uint64(len(wantChunks)), 10*time.Second); err != nil {
					t.Fatal(err)
				}

				// get all chunks from the syncing node
				syncedChunks, err := getChunks(store)
				if err != nil {
					t.Fatal(err)
				}
				// remove historical chunks from total uploaded and synced chunks
				for c := range historicalChunks[id] {
					if _, ok := wantChunks[c]; !ok {
						t.Errorf("missing uploaded historical chunk: %s", c)
					}
					delete(wantChunks, c)
					if _, ok := syncedChunks[c]; !ok {
						t.Errorf("missing synced historical chunk: %s", c)
					}
					delete(syncedChunks, c)
				}
				// validate that all and only all live chunks are synced
				for c := range wantChunks {
					if _, ok := syncedChunks[c]; !ok {
						t.Errorf("missing chunk %v", c)
					}
					delete(wantChunks, c)
					delete(syncedChunks, c)
				}
				if len(wantChunks) != 0 {
					t.Errorf("some of the uploaded live chunks are not synced")
				}
				if len(syncedChunks) != 0 {
					t.Errorf("some of the synced live chunks are not of uploaded ones")
				}
			}
		})
	}
}

func waitChunks(store chunk.Store, want uint64, staledTimeout time.Duration) (err error) {
	start := time.Now()
	var (
		count  uint64        // total number of chunks
		prev   uint64        // total number of chunks in previous check
		sleep  time.Duration // duration until the next check
		staled time.Duration // duration for when the number of chunks is the same
	)
	for staled < staledTimeout { // wait for some time while staled
		count, err = getChunkCount(store)
		if err != nil {
			return err
		}
		if count >= want {
			break
		}
		if count == prev {
			staled += sleep
		} else {
			staled = 0
		}
		prev = count
		if count > 0 {
			// Calculate sleep time only if there is at least 1% of chunks available,
			// less may produce unreliable result.
			if count > want/100 {
				// Calculate the time required to pass for missing chunks to be available,
				// and divide it by half to perform a check earlier.
				sleep = time.Duration(float64(time.Since(start)) * float64(want-count) / float64(count) / 2)
				log.Debug("expecting all chunks", "in", sleep*2, "want", want, "have", count)
			}
		}
		switch {
		case sleep > time.Minute:
			// next check and speed calculation in some shorter time
			sleep = 500 * time.Millisecond
		case sleep > 5*time.Second:
			// upper limit for the check, do not check too slow
			sleep = 5 * time.Second
		case sleep < 50*time.Millisecond:
			// lower limit for the check, do not check too frequently
			sleep = 50 * time.Millisecond
			if staled > 0 {
				// slow down if chunks are stuck near the want value
				sleep *= 10
			}
		}
		time.Sleep(sleep)
	}

	if count != want {
		return fmt.Errorf("got synced chunks %d, want %d", count, want)
	}
	return nil
}

func getChunkCount(store chunk.Store) (c uint64, err error) {
	for po := 0; po <= chunk.MaxPO; po++ {
		last, err := store.LastPullSubscriptionBinID(uint8(po))
		if err != nil {
			return 0, err
		}
		c += last
	}
	return c, nil
}

func getChunks(store chunk.Store) (chunks map[string]struct{}, err error) {
	chunks = make(map[string]struct{})
	for po := uint8(0); po <= chunk.MaxPO; po++ {
		last, err := store.LastPullSubscriptionBinID(uint8(po))
		if err != nil {
			return nil, err
		}
		if last == 0 {
			continue
		}
		ch, _ := store.SubscribePull(context.Background(), po, 0, last)
		for c := range ch {
			addr := c.Address.Hex()
			if _, ok := chunks[addr]; ok {
				return nil, fmt.Errorf("duplicate chunk %s", addr)
			}
			chunks[addr] = struct{}{}
		}
	}
	return chunks, nil
}

/*
 go test -v -bench . -run BenchmarkHistoricalStream -loglevel 0 -benchtime 10x
BenchmarkHistoricalStream_1000-4    	      10	 119487009 ns/op
BenchmarkHistoricalStream_2000-4    	      10	 236469752 ns/op
BenchmarkHistoricalStream_3000-4    	      10	 371934729 ns/op
BenchmarkHistoricalStream_5000-4    	      10	 638317966 ns/op
BenchmarkHistoricalStream_10000-4   	      10	1359858063 ns/op
BenchmarkHistoricalStream_15000-4   	      10	2485790336 ns/op
BenchmarkHistoricalStream_20000-4   	      10	3382260295 ns/op
*/
func BenchmarkHistoricalStream_1000(b *testing.B)  { benchmarkHistoricalStream(b, 1000) }
func BenchmarkHistoricalStream_2000(b *testing.B)  { benchmarkHistoricalStream(b, 2000) }
func BenchmarkHistoricalStream_3000(b *testing.B)  { benchmarkHistoricalStream(b, 3000) }
func BenchmarkHistoricalStream_5000(b *testing.B)  { benchmarkHistoricalStream(b, 5000) }
func BenchmarkHistoricalStream_10000(b *testing.B) { benchmarkHistoricalStream(b, 10000) }
func BenchmarkHistoricalStream_15000(b *testing.B) { benchmarkHistoricalStream(b, 15000) }
func BenchmarkHistoricalStream_20000(b *testing.B) { benchmarkHistoricalStream(b, 20000) }

func benchmarkHistoricalStream(b *testing.B, chunks uint64) {
	b.StopTimer()
	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newSyncSimServiceFunc(nil),
	})

	defer sim.Close()
	uploaderNode, err := sim.AddNode()
	if err != nil {
		b.Fatal(err)
	}

	if err != nil {
		b.Fatal(err)
	}

	uploaderNodeStore := sim.NodeItem(uploaderNode, bucketKeyFileStore).(*storage.FileStore)
	uploadedChunks, err := getChunks(uploaderNodeStore.ChunkStore)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		syncingNode, err := sim.AddNode()
		if err != nil {
			b.Fatal(err)
		}

		mustUploadChunks(context.Background(), b, nodeFileStore(sim, syncingNode), chunks)

		err = sim.Net.Connect(syncingNode, uploaderNode)
		if err != nil {
			b.Fatal(err)
		}
		syncingNodeStore := sim.NodeItem(syncingNode, bucketKeyFileStore).(chunk.Store)
		if err := waitChunks(syncingNodeStore, uint64(len(uploadedChunks)), 10*time.Second); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		err = sim.Net.Stop(syncingNode)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Function that uses putSeenTestHook to record and report
// if there were duplicate chunk synced between Node id
func catchDuplicateChunkSync(t *testing.T) (validate func()) {
	m := make(map[enode.ID]map[string]int)
	var mu sync.Mutex
	putSeenTestHook = func(addr chunk.Address, id enode.ID) {
		mu.Lock()
		defer mu.Unlock()
		if _, ok := m[id]; !ok {
			m[id] = make(map[string]int)
		}
		m[id][addr.Hex()]++
	}
	return func() {
		// reset the test hook
		putSeenTestHook = nil
		// do the validation
		mu.Lock()
		defer mu.Unlock()
		for nodeID, addrs := range m {
			for addr, count := range addrs {
				t.Errorf("chunk synced %v times to node %s: %v", count, nodeID, addr)
			}
		}
	}
}

// TestStarNetworkSync tests that syncing works on a more elaborate network topology
// the test creates a network of 10 nodes and connects them in a star topology, this causes
// the pivot node to have neighbourhood depth > 0, which in turn means that each individual node
// will only get SOME of the chunks that exist on the uploader node (the pivot node).
// The test checks that EVERY chunk that exists on the pivot node:
//	a. exists on the most proximate node
//	b. exists on the nodes subscribed on the corresponding chunk PO
//	c. does not exist on the peers that do not have that PO subscription
func TestStarNetworkSync(t *testing.T) {
	//t.Skip("flaky test https://github.com/ethersphere/swarm/issues/1457")
	if testutil.RaceEnabled {
		return
	}
	var (
		chunkCount = 500
		nodeCount  = 6
		chunkSize  = 4096
		simTimeout = 60 * time.Second
		syncTime   = 30 * time.Second
		filesize   = chunkCount * chunkSize
	)
	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newSyncSimServiceFunc(nil),
	})
	defer sim.Close()

	// create context for simulation run
	ctx, cancel := context.WithTimeout(context.Background(), simTimeout)
	// defer cancel should come before defer simulation teardown
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		nodeIDs := sim.UpNodeIDs()

		nodeIndex := make(map[enode.ID]int)
		for i, id := range nodeIDs {
			nodeIndex[id] = i
		}
		seed := int(time.Now().Unix())
		randomBytes := testutil.RandomBytes(seed, filesize)

		chunkAddrs, err := getAllRefs(randomBytes[:])
		if err != nil {
			return err
		}
		chunksProx := make([]chunkProxData, 0)
		for _, chunkAddr := range chunkAddrs {
			chunkInfo := chunkProxData{
				addr:            chunkAddr,
				uploaderNodePO:  chunk.Proximity(nodeIDs[0].Bytes(), chunkAddr),
				nodeProximities: make(map[enode.ID]int),
			}
			closestNodePO := 0
			for nodeAddr := range nodeIndex {
				po := chunk.Proximity(nodeAddr.Bytes(), chunkAddr)

				chunkInfo.nodeProximities[nodeAddr] = po
				if po > closestNodePO {
					chunkInfo.closestNodePO = po
					chunkInfo.closestNode = nodeAddr
				}
				log.Trace("processed chunk", "uploaderPO", chunkInfo.uploaderNodePO, "ci", chunkInfo.closestNode, "cpo", chunkInfo.closestNodePO, "cadrr", chunkInfo.addr)
			}
			chunksProx = append(chunksProx, chunkInfo)
		}

		// get the pivot node and pump some data
		item := sim.NodeItem(nodeIDs[0], bucketKeyFileStore)
		fileStore := item.(*storage.FileStore)
		reader := bytes.NewReader(randomBytes[:])
		_, wait1, err := fileStore.Store(ctx, reader, int64(len(randomBytes)), false)
		if err != nil {
			return fmt.Errorf("fileStore.Store: %v", err)
		}

		wait1(ctx)

		// check that chunks with a marked proximate host are where they should be
		count := 0

		// wait to sync
		time.Sleep(syncTime)

		log.Info("checking if chunks are on prox hosts")
		for _, c := range chunksProx {
			// if the most proximate host is set - check that the chunk is there
			if c.closestNodePO > 0 {
				count++
				log.Trace("found chunk with proximate host set, trying to find in localstore", "po", c.closestNodePO, "closestNode", c.closestNode)
				item = sim.NodeItem(c.closestNode, bucketKeyFileStore)
				store := item.(chunk.Store)

				_, err := store.Get(context.TODO(), chunk.ModeGetRequest, c.addr)
				if err != nil {
					return err
				}
			}
		}
		log.Debug("done checking stores", "checked chunks", count, "total chunks", len(chunksProx))
		if count != len(chunksProx) {
			return fmt.Errorf("checked chunks dont match numer of chunks. got %d want %d", count, len(chunksProx))
		}
		// clients are interested in streams according ot their own kademlia depth and not according to the server's kademlia depth
		// this is a major change in comparison to what was in the previous streamer
		// we can possibly maintain this same test vector, but we would have to manually fiddle with the individual nodes (everyone else except the pivot) kademlia
		// by adding artificial nodes -> this way the depth changes and we would be interested in only certain streams from the server, all while preserving the previous test vector
		// another option would be to bring up a cluster (although it might have to be a relatively big one) and make sure that all nodes have depth > 0
		// then measure each node po to each chunk, similarly like in this test vector, then assert on which node it is supposed to be stored. first option seems more feasible
		//uploaderStream := sim.NodeItem(nodeIDs[0], bucketKeyStream)
		//client, err := node.Client()
		//if err != nil {
		//return fmt.Errorf("create node 1 rpc client fail: %v", err)
		//}

		////ask it for subscriptions
		//pstreams := make(map[string][]string)
		//err = client.Call(&pstreams, "stream_getPeerServerSubscriptions")
		//if err != nil {
		//return fmt.Errorf("client call stream_getPeerSubscriptions: %v", err)
		//}

		//create a map of no-subs for a node
		//noSubMap := make(map[enode.ID]map[int]bool)

		//for subscribedNode, streams := range pstreams {
		//id := enode.HexID(subscribedNode)
		//b := make([]bool, 17)
		//for _, sub := range streams {
		//subPO, err := ParseSyncBinKey(strings.Split(sub, "|")[1])
		//if err != nil {
		//return err
		//}
		//b[int(subPO)] = true
		//}
		//noMapMap := make(map[int]bool)
		//for i, v := range b {
		//if !v {
		//noMapMap[i] = true
		//}
		//}
		//noSubMap[id] = noMapMap
		//}

		// iterate over noSubMap, for each node check if it has any of the chunks it shouldn't have
		//for nodeId, nodeNoSubs := range noSubMap {
		//for _, c := range chunksProx {
		//// if the chunk PO is equal to the sub that the node shouldnt have - check if the node has the chunk!
		//if _, ok := nodeNoSubs[c.uploaderNodePO]; ok {
		//count++
		//item = sim.NodeItem(nodeId, bucketKeyFileStore)
		//store := item.(chunk.Store)

		//_, err := store.Get(context.TODO(), chunk.ModeGetRequest, c.addr)
		//if err == nil {
		//return fmt.Errorf("got a chunk where it shouldn't be! addr %s, nodeId %s", c.addr, nodeId)
		//}
		//}
		//}
		//}
		return nil
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

type chunkProxData struct {
	addr            chunk.Address
	uploaderNodePO  int
	nodeProximities map[enode.ID]int
	closestNode     enode.ID
	closestNodePO   int
}

func getAllRefs(testData []byte) (storage.AddressCollection, error) {
	datadir, err := ioutil.TempDir("", "chunk-debug")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir: %v", err)
	}
	defer os.RemoveAll(datadir)
	fileStore, cleanup, err := storage.NewLocalFileStore(datadir, make([]byte, 32), chunk.NewTags())
	if err != nil {
		return nil, err
	}
	defer cleanup()

	reader := bytes.NewReader(testData)
	return fileStore.GetAllReferences(context.Background(), reader, false)
}
