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

package stream

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/network/timeouts"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const (
	BatchSize = 8
)

// SwarmSyncerServer implements an Server for history syncing on bins
// offered streams:
// * live request delivery with or without checkback
// * (live/non-live historical) chunk syncing per proximity bin
type SwarmSyncerServer struct {
	po       uint8
	netStore *network.NetStore
	quit     chan struct{}
}

// NewSwarmSyncerServer is constructor for SwarmSyncerServer
func NewSwarmSyncerServer(po uint8, netStore *network.NetStore) (*SwarmSyncerServer, error) {
	return &SwarmSyncerServer{
		po:       po,
		netStore: netStore,
		quit:     make(chan struct{}),
	}, nil
}

func RegisterSwarmSyncerServer(streamer *Registry, netStore *network.NetStore) {
	streamer.RegisterServerFunc("SYNC", func(_ *Peer, t string, _ bool) (Server, error) {
		po, err := ParseSyncBinKey(t)
		if err != nil {
			return nil, err
		}
		return NewSwarmSyncerServer(po, netStore)
	})
	// streamer.RegisterServerFunc(stream, func(p *Peer) (Server, error) {
	// 	return NewOutgoingProvableSwarmSyncer(po, db)
	// })
}

// Close needs to be called on a stream server
func (s *SwarmSyncerServer) Close() {
	close(s.quit)
}

// GetData retrieves the actual chunk from netstore
func (s *SwarmSyncerServer) GetData(ctx context.Context, key []byte) ([]byte, error) {
	//TODO: this should be localstore, not netstore?
	r := &network.Request{
		Addr:     storage.Address(key),
		Origin:   enode.ID{},
		HopCount: 0,
	}

	// this timeout shouldn't be necessary as syncer server is supposed to go straight to localstore,
	// but if a chunk is garbage collected while we actually offered it, it is possible for this
	// to trigger a network request
	ctx, cancel := context.WithTimeout(ctx, timeouts.FetcherGlobalTimeout)
	defer cancel()

	chunk, err := s.netStore.Get(ctx, r)
	if err != nil {
		return nil, err
	}
	return chunk.Data(), nil
}

// SessionIndex returns current storage bin (po) index.
func (s *SwarmSyncerServer) SessionIndex() (uint64, error) {
	return s.netStore.BinIndex(s.po), nil
}

// GetBatch retrieves the next batch of hashes from the dbstore
func (s *SwarmSyncerServer) SetNextBatch(from, to uint64) ([]byte, uint64, uint64, *HandoverProof, error) {
	var batch []byte
	i := 0

	var ticker *time.Ticker
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()
	var wait bool
	for {
		if wait {
			if ticker == nil {
				ticker = time.NewTicker(1000 * time.Millisecond)
			}
			select {
			case <-ticker.C:
			case <-s.quit:
				return nil, 0, 0, nil, nil
			}
		}

		metrics.GetOrRegisterCounter("syncer.setnextbatch.iterator", nil).Inc(1)
		err := s.netStore.Iterator(from, to, s.po, func(key storage.Address, idx uint64) bool {
			select {
			case <-s.quit:
				return false
			default:
			}
			batch = append(batch, key[:]...)
			i++
			to = idx
			return i < BatchSize
		})
		if err != nil {
			return nil, 0, 0, nil, err
		}
		if len(batch) > 0 {
			break
		}
		wait = true
	}

	log.Trace("SwarmSyncerServer offer batch", "po", s.po, "len", i, "from", from, "to", to, "current store count", s.netStore.BinIndex(s.po))
	return batch, from, to, nil, nil
}

// SwarmSyncerClient
type SwarmSyncerClient struct {
	store  *network.NetStore
	peer   *Peer
	stream Stream
}

// NewSwarmSyncerClient is a contructor for provable data exchange syncer
func NewSwarmSyncerClient(p *Peer, netStore *network.NetStore, stream Stream) (*SwarmSyncerClient, error) {
	return &SwarmSyncerClient{
		store:  netStore,
		peer:   p,
		stream: stream,
	}, nil
}

// RegisterSwarmSyncerClient registers the client constructor function for
// to handle incoming sync streams
func RegisterSwarmSyncerClient(streamer *Registry, netStore *network.NetStore) {
	streamer.RegisterClientFunc("SYNC", func(p *Peer, t string, live bool) (Client, error) {
		return NewSwarmSyncerClient(p, netStore, NewStream("SYNC", t, live))
	})
}

func (s *SwarmSyncerClient) NeedData(ctx context.Context, key []byte) (loaded bool, wait func(context.Context) error) {
	start := time.Now()

	has, fi, loaded := s.store.HasWithCallback(ctx, key, "syncer")
	if has {
		return loaded, nil
	}

	return loaded, func(ctx context.Context) error {
		select {
		case <-fi.Delivered:
			metrics.GetOrRegisterResettingTimer(fmt.Sprintf("fetcher.%s.syncer", fi.CreatedBy), nil).UpdateSince(start)
		case <-time.After(20 * time.Second):
			// TODO: whats the proper timeout here? it is not the global fetcher timeout,
			// since we don't do NetStore.Get(), but just wait for chunk delivery via offered/wanted hashes
			metrics.GetOrRegisterCounter("fetcher.syncer.timeout", nil).Inc(1)
			return fmt.Errorf("chunk not delivered through syncing after 20sec. ref=%s", fmt.Sprintf("%x", key))
		}
		return nil
	}
}

func (s *SwarmSyncerClient) BatchDone(stream Stream, from uint64, hashes []byte, root []byte) func() (*TakeoverProof, error) {
	// TODO: reenable this with putter/getter refactored code
	// if s.chunker != nil {
	// 	return func() (*TakeoverProof, error) { return s.TakeoverProof(stream, from, hashes, root) }
	// }
	return nil
}

func (s *SwarmSyncerClient) Close() {}

// base for parsing and formating sync bin key
// it must be 2 <= base <= 36
const syncBinKeyBase = 36

// FormatSyncBinKey returns a string representation of
// Kademlia bin number to be used as key for SYNC stream.
func FormatSyncBinKey(bin uint8) string {
	return strconv.FormatUint(uint64(bin), syncBinKeyBase)
}

// ParseSyncBinKey parses the string representation
// and returns the Kademlia bin number.
func ParseSyncBinKey(s string) (uint8, error) {
	bin, err := strconv.ParseUint(s, syncBinKeyBase, 8)
	if err != nil {
		return 0, err
	}
	return uint8(bin), nil
}
