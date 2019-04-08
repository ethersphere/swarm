// Copyright 2019 The go-ethereum Authors
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

package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

type Inspector struct {
	api      *API
	hive     *network.Hive
	netStore *network.NetStore
}

func NewInspector(api *API, hive *network.Hive, netStore *network.NetStore) *Inspector {
	return &Inspector{api, hive, netStore}
}

// Hive prints the kademlia table
func (inspector *Inspector) Hive() string {
	return inspector.hive.String()
}

func (inspector *Inspector) ListKnown() []string {
	res := []string{}
	for _, v := range inspector.hive.Kademlia.ListKnown() {
		res = append(res, fmt.Sprintf("%v", v))
	}
	return res
}

func (inspector *Inspector) Nearest(chunkRef string) string {
	node, err := inspector.hive.Kademlia.Nearest(chunkRef)
	if err != nil {
		log.Error(err.Error())
		return ""
	}
	return node
}

func (inspector *Inspector) AllNearest(chunkRef string) []*network.Peer {
	nodes, err := inspector.hive.Kademlia.AllNearest(chunkRef)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return nodes
}

func (inspector *Inspector) IsSyncing() bool {
	lastReceivedChunksMsg := metrics.GetOrRegisterGauge("network.stream.received_chunks", nil)

	// last received chunks msg time
	lrct := time.Unix(0, lastReceivedChunksMsg.Value())

	// if last received chunks msg time is after now-30sec. (i.e. within the last 30sec.) then we say that the node is still syncing
	// technically this is not correct, because this might have been a retrieve request, but for the time being it works for our purposes
	// because we know we are not making retrieve requests on the node while checking this
	return lrct.After(time.Now().Add(-30 * time.Second))
}

type HasInfo struct {
	Addr string `json:"address"`
	Has  bool   `json:"has"`
}

// Has checks whether each chunk address is present in the underlying datastore,
// the bool in the returned structs indicates if the underlying datastore has
// the chunk stored with the given address (true), or not (false)
func (inspector *Inspector) Has(chunkAddresses []storage.Address) string {
	//TODO: update to a byte slice
	hostChunks := []string{}
	for _, addr := range chunkAddresses {
		res := HasInfo{}
		res.Addr = addr.String()
		res.Has = inspector.netStore.Has(context.Background(), addr)
		if res.Has {
			hostChunks = append(hostChunks, "1")
		} else {
			hostChunks = append(hostChunks, "0")
		}
	}

	return strings.Join(hostChunks, "")
}
