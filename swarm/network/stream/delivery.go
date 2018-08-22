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
	"errors"

	"fmt"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const (
	swarmChunkServerStreamName = "RETRIEVE_REQUEST"
	deliveryCap                = 32
)

var (
	processReceivedChunksCount    = metrics.NewRegisteredCounter("network.stream.received_chunks.count", nil)
	handleRetrieveRequestMsgCount = metrics.NewRegisteredCounter("network.stream.handle_retrieve_request_msg.count", nil)

	requestFromPeersCount     = metrics.NewRegisteredCounter("network.stream.request_from_peers.count", nil)
	requestFromPeersEachCount = metrics.NewRegisteredCounter("network.stream.request_from_peers_each.count", nil)
)

type Delivery struct {
	chunkStore storage.ChunkStore
	overlay    network.Overlay
	getPeer    func(discover.NodeID) *Peer
}

func NewDelivery(overlay network.Overlay, chunkStore storage.ChunkStore) *Delivery {
	return &Delivery{
		chunkStore: chunkStore,
		overlay:    overlay,
	}
}

// SwarmChunkServer implements Server
type SwarmChunkServer struct {
	deliveryC  chan []byte
	batchC     chan []byte
	chunkStore storage.ChunkStore
	currentLen uint64
	quit       chan struct{}
}

// NewSwarmChunkServer is SwarmChunkServer constructor
func NewSwarmChunkServer(chunkStore storage.ChunkStore) *SwarmChunkServer {
	s := &SwarmChunkServer{
		deliveryC:  make(chan []byte, deliveryCap),
		batchC:     make(chan []byte),
		chunkStore: chunkStore,
		quit:       make(chan struct{}),
	}
	go s.processDeliveries()
	return s
}

func (s *SwarmChunkServer) context(req *RetrieveRequestMsg) context.Context {
	var cancel func()
	ctx := context.Background()
	// if req.Timeout > 0 {
	// ctx, cancel = context.WithTimeout(ctx, req.Timeout)
	// }
	ctx, cancel = context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
		case <-s.quit:
			cancel()
		}
	}()

	return ctx
}

// processDeliveries handles delivered chunk hashes
func (s *SwarmChunkServer) processDeliveries() {
	var hashes []byte
	var batchC chan []byte
	for {
		select {
		case <-s.quit:
			return
		case hash := <-s.deliveryC:
			hashes = append(hashes, hash...)
			batchC = s.batchC
		case batchC <- hashes:
			hashes = nil
			batchC = nil
		}
	}
}

// SetNextBatch
func (s *SwarmChunkServer) SetNextBatch(_, _ uint64) (hashes []byte, from uint64, to uint64, proof *HandoverProof, err error) {
	select {
	case hashes = <-s.batchC:
	case <-s.quit:
		return
	}

	from = s.currentLen
	s.currentLen += uint64(len(hashes))
	to = s.currentLen
	return
}

// Close needs to be called on a stream server
func (s *SwarmChunkServer) Close() {
	close(s.quit)
}

// GetData retrives chunk data from db store
func (s *SwarmChunkServer) GetData(ctx context.Context, key []byte) ([]byte, error) {
	chunk, err := s.chunkStore.Get(ctx, storage.Address(key))
	if err != nil {
		return nil, err
	}
	return chunk.Data(), nil
}

// RetrieveRequestMsg is the protocol msg for chunk retrieve requests
type RetrieveRequestMsg struct {
	Addr      storage.Address
	SkipCheck bool
}

// TODO: Fix context SNAFU
func (d *Delivery) handleRetrieveRequestMsg(_ context.Context, sp *Peer, req *RetrieveRequestMsg) error {
	log.Trace("received request", "peer", sp.ID(), "hash", req.Addr)
	handleRetrieveRequestMsgCount.Inc(1)

	//var osp opentracing.Span
	//ctx, osp = spancontext.StartSpan(
	//ctx,
	//"retrieve.request")
	//defer osp.Finish()

	s, err := sp.getServer(NewStream(swarmChunkServerStreamName, "", false))
	if err != nil {
		return err
	}
	streamer := s.Server.(*SwarmChunkServer)

	ctx := streamer.context(req)
	ctx = context.WithValue(ctx, "peer", sp.ID().String())
	go func() {
		chunk, err := d.chunkStore.Get(ctx, req.Addr)
		if err != nil {
			return
		}
		if req.SkipCheck {
			err = sp.Deliver(ctx, chunk, s.priority)
			if err != nil {
				log.Warn("ERROR in handleRetrieveRequestMsg, DROPPING peer!", "err", err)
				return
			}
		}
		streamer.deliveryC <- chunk.Address()[:]
	}()

	return nil
}

type ChunkDeliveryMsg struct {
	Addr  storage.Address
	SData []byte // the stored chunk Data (incl size)
	peer  *Peer  // set in handleChunkDeliveryMsg
}

// TODO: Fix context SNAFU
func (d *Delivery) handleChunkDeliveryMsg(ctx context.Context, sp *Peer, req *ChunkDeliveryMsg) error {
	processReceivedChunksCount.Inc(1)
	go func() {
		req.peer = sp
		err := d.chunkStore.Put(ctx, storage.NewChunk(req.Addr, req.SData))
		if err != nil {
			if err == storage.ErrChunkInvalid {
				req.peer.Drop(err)
			}
		}
	}()
	return nil
}

// RequestFromPeers sends a chunk retrieve request to
func (d *Delivery) RequestFromPeers(ctx context.Context, req *network.Request) (*discover.NodeID, chan struct{}, error) {
	requestFromPeersCount.Inc(1)
	var sp *Peer
	spID := req.Source

	if spID != nil {
		sp = d.getPeer(*spID)
		if sp == nil {
			return nil, nil, fmt.Errorf("source peer %v not found", spID.String())
		}
	} else {
		d.overlay.EachConn(req.Addr[:], 255, func(p network.OverlayConn, po int, nn bool) bool {
			id := p.(network.Peer).ID()
			// TODO: skip light nodes that do not accept retrieve requests
			if _, ok := req.PeersToSkip.Load(id.String()); ok {
				log.Trace("Delivery.RequestFromPeers: skip peer", "peer id", id)
				return true
			}
			sp = d.getPeer(id)
			if sp == nil {
				log.Warn("Delivery.RequestFromPeers: peer not found", "id", id)
				return true
			}
			spID = &id
			return false
		})
		if sp == nil {
			return nil, nil, errors.New("no peer found")
		}
	}

	err := sp.SendPriority(ctx, &RetrieveRequestMsg{
		Addr:      req.Addr,
		SkipCheck: req.SkipCheck,
	}, Top)
	if err != nil {
		return nil, nil, err
	}
	requestFromPeersEachCount.Inc(1)

	return spID, sp.quit, nil
}
