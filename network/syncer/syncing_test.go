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

package syncer

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

// TestTwoNodesFullSync connects two nodes, uploads content to one node and expects the
// uploader node's chunks to be synced to the second node. This is expected behaviour since although
// both nodes might share address bits, due to kademlia depth=0 when under ProxBinSize - this will
// eventually create subscriptions on all bins between the two nodes, causing a full sync between them
// The test checks that:
// 1. All subscriptions are created
// 2. All chunks are transferred from one node to another (asserted by summing and comparing bin indexes on both nodes)
func TestTwoNodesFullSync(t *testing.T) {
	var (
		chunkCount = 10000
		syncTime   = 1 * time.Second
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
		time.Sleep(syncTime)

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

type chunkProxData struct {
	addr            chunk.Address
	uploaderNodePO  int
	nodeProximities map[enode.ID]int
	closestNode     enode.ID
	closestNodePO   int
}
