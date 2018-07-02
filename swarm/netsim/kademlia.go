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

package netsim

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/network"
)

// BucketKeyKademlia is the key to be used for storing the kademlia
// instance for particuar node, usually inside the ServiceFunc function.
var BucketKeyKademlia BucketKey = "kademlia"

// WaitKademlia is blocking until the health of all kademlias is true.
func (s *Simulation) WaitKademlia(ctx context.Context, kadMinProxSize int) (err error) {
	// Prepare PeerPot map for checking Kademlia health
	var ppmap map[string]*network.PeerPot
	kademlias := s.kademlias()
	addrs := make([][]byte, len(kademlias))
	for i, k := range kademlias {
		addrs[i] = k.BaseAddr()
	}
	ppmap = network.NewPeerPotMap(kadMinProxSize, addrs)

	// Wait for healthy Kademlia on every node before checking files
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			healthy := true
			log.Debug("kademlia health check", "addr count", len(addrs))
			for i, k := range kademlias {
				//PeerPot for this node
				addr := common.Bytes2Hex(k.BaseAddr())
				pp := ppmap[addr]
				//call Healthy RPC
				h := k.Healthy(pp)
				//print info
				log.Debug(k.String())
				log.Debug("kademlia", "empty bins", pp.EmptyBins, "gotNN", h.GotNN, "knowNN", h.KnowNN, "full", h.Full)
				log.Debug("kademlia", "health", h.GotNN && h.KnowNN && h.Full, "addr", fmt.Sprintf("%x", k.BaseAddr()), "i", i)
				log.Debug("kademlia", "ill condition", !h.GotNN || !h.Full, "addr", fmt.Sprintf("%x", k.BaseAddr()), "i", i)
				if !h.GotNN || !h.Full {
					healthy = false
					break
				}
			}
			if healthy {
				return nil
			}
		}
	}
}

// kademlias returns all Kademlia instances that are set
// in simulation bucket.
func (s *Simulation) kademlias() (ks []*network.Kademlia) {
	for _, v := range s.UpNodesItems(BucketKeyKademlia) {
		k, ok := v.(*network.Kademlia)
		if !ok {
			continue
		}
		ks = append(ks, k)
	}
	return ks
}
