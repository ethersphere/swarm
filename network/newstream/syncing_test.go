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

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

// TestTwoNodesFullSync connects two nodes, uploads content to one node and expects the
// uploader node's chunks to be synced to the second node. This is expected behaviour since although
// both nodes might share address bits, due to kademlia depth=0 when under ProxBinSize - this will
// eventually create subscriptions on all bins between the two nodes, causing a full sync between them
// The test checks that:
// 1. All subscriptions are created
// 2. All chunks are transferred from one node to another (asserted by summing and comparing bin indexes on both nodes)
func TestTwoNodesFullSync(t *testing.T) {
	var (
		chunkCount = 20000
		syncTime   = 3 * time.Second
	)
	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(0),
	})
	defer sim.Close()

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}
	//defer profile.Start(profile.CPUProfile).Stop()

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		nodeIDs := sim.UpNodeIDs()

		item := sim.NodeItem(sim.UpNodeIDs()[0], bucketKeyFileStore)

		log.Debug("subscriptions on all bins exist between the two nodes, proceeding to check bin indexes")
		log.Debug("uploader node", "enode", nodeIDs[0])
		item = sim.NodeItem(nodeIDs[0], bucketKeyFileStore)
		store := item.(chunk.Store)

		//put some data into just the first node
		filesize := chunkCount * 4096
		cctx := context.Background()
		_, wait, err := item.(*storage.FileStore).Store(cctx, testutil.RandomReader(0, filesize), int64(filesize), false)
		if err != nil {
			return err
		}
		if err := wait(cctx); err != nil {
			return err
		}

		id, err := sim.AddNodes(1)
		if err != nil {
			return err
		}
		err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
		if err != nil {
			return err
		}
		nodeIDs = sim.UpNodeIDs()
		syncingNodeId := nodeIDs[1]

		uploaderNodeBinIDs := make([]uint64, 17)

		log.Debug("checking pull subscription bin ids")
		for po := 0; po <= 16; po++ {
			until, err := store.LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			log.Debug("uploader node got bin index", "bin", po, "binIndex", until)

			uploaderNodeBinIDs[po] = until
		}

		// wait for syncing
		<-time.After(syncTime)

		// check that the sum of bin indexes is equal

		log.Debug("compare to", "enode", syncingNodeId)
		item = sim.NodeItem(syncingNodeId, bucketKeyFileStore)
		db := item.(chunk.Store)

		uploaderSum, otherNodeSum := 0, 0
		for po, uploaderUntil := range uploaderNodeBinIDs {
			shouldUntil, err := db.LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			otherNodeSum += int(shouldUntil)
			uploaderSum += int(uploaderUntil)
		}
		if uploaderSum != otherNodeSum {
			return fmt.Errorf("bin indice sum mismatch. got %d want %d", otherNodeSum, uploaderSum)
		}
		return nil
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

	chunkCount := func(t *testing.T, store chunk.Store) (c uint64) {
		t.Helper()

		for po := 0; po <= chunk.MaxPO; po++ {
			last, err := store.LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				t.Fatal(err)
			}
			c += last
		}
		return c
	}

	waitChunks := func(t *testing.T, store chunk.Store, want uint64) {
		t.Helper()

		for i := 49; i >= 0; i-- {
			time.Sleep(100 * time.Millisecond)

			syncedChunkCount := chunkCount(t, store)
			if syncedChunkCount != want {
				if i == 0 {
					t.Errorf("got synced chunks %d, want %d", syncedChunkCount, want)
				} else {
					continue
				}
			}
			break
		}
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

			totalChunkCount := chunkCount(t, uploadStore)

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

			waitChunks(t, syncStore, totalChunkCount-removedCount)

			if tc.liveChunkCount > 0 {
				chunks = append(chunks, uploadChunks(t, ctx, uploadStore, tc.liveChunkCount)...)

				totalChunkCount = chunkCount(t, uploadStore)

				if want := tc.chunkCount + tc.liveChunkCount; totalChunkCount != want {
					t.Errorf("uploaded %v chunks, want %v", totalChunkCount, want)
				}

				removedCount += removeChunks(t, ctx, uploadStore, tc.liveGaps, chunks)

				waitChunks(t, syncStore, totalChunkCount-removedCount)
			}
		})
	}
}

// TestTwoNodesFullSyncLive brings up one node, adds chunkCount * 4096 bytes to its localstore, then connects to it another fresh node.
// it then waits for syncTime and checks that they have both synced correctly. It then adds another chunkCount to the uploader node
// and waits for another syncTime, then checks for the correct sync by bin indexes
func TestTwoNodesFullSyncLive(t *testing.T) {
	var (
		chunkCount = 10000
		syncTime   = 3 * time.Second
	)
	sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
		"bzz-sync": newBzzSyncWithLocalstoreDataInsertion(0),
	})
	defer sim.Close()

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		nodeIDs := sim.UpNodeIDs()

		uploaderNodeStore := sim.NodeItem(sim.UpNodeIDs()[0], bucketKeyFileStore)

		log.Debug("uploader node", "enode", nodeIDs[0])

		//put some data into just the first node
		filesize := chunkCount * 4096
		cctx := context.Background()
		_, wait, err := uploaderNodeStore.(*storage.FileStore).Store(cctx, testutil.RandomReader(0, filesize), int64(filesize), false)
		if err != nil {
			return err
		}
		if err := wait(cctx); err != nil {
			return err
		}

		id, err := sim.AddNodes(1)
		if err != nil {
			return err
		}
		err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
		if err != nil {
			return err
		}
		nodeIDs = sim.UpNodeIDs()
		syncingNodeStore := sim.NodeItem(nodeIDs[1], bucketKeyFileStore)

		uploaderNodeBinIDs := make([]uint64, 17)

		log.Debug("checking pull subscription bin ids")
		for po := 0; po <= 16; po++ {
			until, err := uploaderNodeStore.(chunk.Store).LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			log.Debug("uploader node got bin index", "bin", po, "binIndex", until)

			uploaderNodeBinIDs[po] = until
		}

		// wait for syncing
		<-time.After(syncTime)

		// check that the sum of bin indexes is equal
		log.Debug("compare to", "enode", nodeIDs[1])

		uploaderSum, otherNodeSum := 0, 0
		for po, uploaderUntil := range uploaderNodeBinIDs {
			shouldUntil, err := syncingNodeStore.(chunk.Store).LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			otherNodeSum += int(shouldUntil)
			uploaderSum += int(uploaderUntil)
		}
		if uploaderSum != otherNodeSum {
			return fmt.Errorf("bin indice sum mismatch. got %d want %d", otherNodeSum, uploaderSum)
		}

		// now add stuff so that we fetch it with live syncing

		_, wait, err = uploaderNodeStore.(*storage.FileStore).Store(cctx, testutil.RandomReader(101010, filesize), int64(filesize), false)
		if err != nil {
			return err
		}
		if err = wait(cctx); err != nil {
			return err
		}

		// count the content in the bins again
		for po := 0; po <= 16; po++ {
			until, err := uploaderNodeStore.(chunk.Store).LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			log.Debug("uploader node got bin index", "bin", po, "binIndex", until)

			uploaderNodeBinIDs[po] = until
		}

		// wait for live syncing
		<-time.After(syncTime)

		for po, uploaderUntil := range uploaderNodeBinIDs {
			shouldUntil, err := syncingNodeStore.(chunk.Store).LastPullSubscriptionBinID(uint8(po))
			if err != nil {
				return err
			}
			otherNodeSum += int(shouldUntil)
			uploaderSum += int(uploaderUntil)
		}
		if uploaderSum != otherNodeSum {
			return fmt.Errorf("live sync bin indice sum mismatch. got %d want %d", otherNodeSum, uploaderSum)
		}

		return nil
	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}
}
