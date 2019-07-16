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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

// more tests:
// 1. bring up 3 nodes, on each of them different content, connect them all together and check that all 3 get the union of the 3 localstores

//TODO:
// - write test that brings up a bigger cluster, then tests all individual nodes with localstore get to get the chunks that were
// uploaded to the first node

var timeout = 30 * time.Second

// TestTwoNodesFullSync connects two nodes, uploads content to one node and expects the
// uploader node's chunks to be synced to the second node. This is expected behaviour since although
// both nodes might share address bits, due to kademlia depth=0 when under ProxBinSize - this will
// eventually create subscriptions on all bins between the two nodes, causing a full sync between them
// The test checks that:
// 1. All subscriptions are created
// 2. All chunks are transferred from one node to another (asserted by summing and comparing bin indexes on both nodes)
func TestTwoNodesFullSync(t *testing.T) {
	chunkCount := 10000
	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(chunkCount),
	})

	defer sim.Close()

	id, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}
	nodeIDs := id

	log.Debug("pivot node", "enode", nodeIDs[0])

	//defer profile.Start(profile.CPUProfile).Stop()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		id, err := sim.AddNodes(1)
		if err != nil {
			return err
		}

		nodeIDs := sim.UpNodeIDs()

		err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
		if err != nil {
			return err
		}
		syncingNodeID := nodeIDs[1]
		uploaderNodeID := nodeIDs[0]

		uploaderNodeBinIDs := make([]uint64, 17)
		uploaderStore := sim.NodeItem(uploaderNodeID, bucketKeyFileStore).(chunk.Store)
		log.Debug("checking pull subscription bin ids")
		var uploaderSum uint64
		for po := 0; po <= 16; po++ {
			until, err := uploaderStore.LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			log.Debug("uploader node got bin index", "bin", po, "binIndex", until)

			uploaderNodeBinIDs[po] = until
			uploaderSum += until
		}

		// check that the sum of bin indexes is equal
		log.Debug("compare to", "enode", syncingNodeID)

		return waitChunks(sim.NodeItem(syncingNodeID, bucketKeyFileStore).(chunk.Store), uploaderSum, 10*time.Second)
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

func TestTwoNodesSyncWithGaps(t *testing.T) {
	uploadChunks := func(t *testing.T, ctx context.Context, store chunk.Store, count uint64) (chunks []chunk.Address) {
		t.Helper()

		for i := uint64(0); i < count; i++ {
			c := storage.GenerateRandomChunk(4096)
			exists, err := store.Put(ctx, chunk.ModePutUpload, c)
			if err != nil {
				t.Fatal(err)
			}
			if exists {
				t.Fatal("generated already existing chunk")
			}
			chunks = append(chunks, c.Address())
		}
		return chunks
	}

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
			sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
				"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(0),
			})
			defer sim.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			uploadNode, err := sim.AddNode()
			if err != nil {
				t.Fatal(err)
			}

			uploadStore := sim.NodeItem(uploadNode, bucketKeyFileStore).(chunk.Store)

			chunks := uploadChunks(t, ctx, uploadStore, tc.chunkCount)

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
				chunks = append(chunks, uploadChunks(t, ctx, uploadStore, tc.liveChunkCount)...)

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
	var (
		chunkCount = 20000
	)

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(0),
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

		// now add stuff so that we fetch it with live syncing
		filesize := chunkCount * 4096
		_, wait, err := uploaderNodeStore.Store(ctx, testutil.RandomReader(101010, filesize), int64(filesize), false)
		if err != nil {
			return err
		}
		if err = wait(ctx); err != nil {
			return err
		}

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
	var (
		chunkCount = 10000 // per history and per live upload
	)

	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(0),
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

		// add chunks for historical syncing before new node connection
		filesize := chunkCount * 4096
		_, wait, err := uploaderNodeStore.Store(ctx, testutil.RandomReader(int(time.Now().UnixNano()), filesize), int64(filesize), false)
		if err != nil {
			return err
		}
		if err = wait(ctx); err != nil {
			return err
		}

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
		_, wait, err = uploaderNodeStore.Store(ctx, testutil.RandomReader(int(time.Now().UnixNano()), filesize), int64(filesize), false)
		if err != nil {
			return err
		}
		if err = wait(ctx); err != nil {
			return err
		}

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

func BenchmarkHistoricalStream_1000(b *testing.B)  { benchmarkHistoricalStream(b, 1000) }
func BenchmarkHistoricalStream_2000(b *testing.B)  { benchmarkHistoricalStream(b, 2000) }
func BenchmarkHistoricalStream_3000(b *testing.B)  { benchmarkHistoricalStream(b, 3000) }
func BenchmarkHistoricalStream_5000(b *testing.B)  { benchmarkHistoricalStream(b, 5000) }
func BenchmarkHistoricalStream_10000(b *testing.B) { benchmarkHistoricalStream(b, 10000) }
func BenchmarkHistoricalStream_15000(b *testing.B) { benchmarkHistoricalStream(b, 15000) }
func BenchmarkHistoricalStream_20000(b *testing.B) { benchmarkHistoricalStream(b, 20000) }

func benchmarkHistoricalStream(b *testing.B, chunks int) {
	b.StopTimer()
	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(chunks),
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
		syncingNode, err := sim.AddNode()
		if err != nil {
			b.Fatal(err)
		}

		b.StartTimer()

		err = sim.Net.ConnectNodesStar([]enode.ID{syncingNode}, uploaderNode)
		if err != nil {
			b.Fatal(err)
		}
		syncingNodeStore := sim.NodeItem(syncingNode, bucketKeyFileStore).(chunk.Store)
		if err := waitChunks(syncingNodeStore, uint64(len(uploadedChunks)), 10*time.Second); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		err = sim.Net.Disconnect(syncingNode, uploaderNode)
		if err != nil {
			b.Fatal(err)
		}
	}
}
