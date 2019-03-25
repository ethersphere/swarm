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
	"bytes"
	"context"
	"errors"
	"runtime"
	"strconv"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/spancontext"
	"github.com/ethereum/go-ethereum/swarm/storage"
	olog "github.com/opentracing/opentracing-go/log"
)

const (
	swarmChunkServerStreamName = "RETRIEVE_REQUEST"
	deliveryCap                = 32
)

var (
	processReceivedChunksCount    = metrics.NewRegisteredCounter("network.stream.received_chunks.count", nil)
	handleRetrieveRequestMsgCount = metrics.NewRegisteredCounter("network.stream.handle_retrieve_request_msg.count", nil)
	retrieveChunkFail             = metrics.NewRegisteredCounter("network.stream.retrieve_chunks_fail.count", nil)
)

type Delivery struct {
	chunkStore storage.ChunkStore
	kad        *network.Kademlia
	getPeer    func(enode.ID) *Peer
}

func NewDelivery(kad *network.Kademlia, chunkStore storage.ChunkStore) *Delivery {
	return &Delivery{
		chunkStore: chunkStore,
		kad:        kad,
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

// SessionIndex returns zero in all cases for SwarmChunkServer.
func (s *SwarmChunkServer) SessionIndex() (uint64, error) {
	return 0, nil
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
	HopCount  uint8
}

func (d *Delivery) handleRetrieveRequestMsg(ctx context.Context, sp *Peer, req *RetrieveRequestMsg) error {
	log.Trace("handle retrieve request", "peer", sp.ID(), "hash", req.Addr)
	handleRetrieveRequestMsgCount.Inc(1)

	ctx, osp := spancontext.StartSpan(
		ctx,
		"handle.retrieve.request")

	osp.LogFields(olog.String("ref", req.Addr.String()))

	s, err := sp.getServer(NewStream(swarmChunkServerStreamName, "", true))
	if err != nil {
		return err
	}
	streamer := s.Server.(*SwarmChunkServer)

	ctx, cancel := context.WithTimeout(context.WithValue(ctx, "hopCount", req.HopCount+1), network.RequestTimeout)

	go func() {
		select {
		case <-ctx.Done():
		case <-streamer.quit:
		}
		cancel()
	}()

	go func() {
		defer osp.Finish()
		chunk, err := d.chunkStore.Get(ctx, req.Addr)
		if err != nil {
			retrieveChunkFail.Inc(1)
			log.Debug("ChunkStore.Get can not retrieve chunk", "peer", sp.ID().String(), "addr", req.Addr, "hopcount", req.HopCount, "err", err)
			return
		}
		if req.SkipCheck {
			syncing := false
			osp.LogFields(olog.Bool("skipCheck", true))

			log.Trace("retrieve request, delivery", "ref", req.Addr, "peer", sp.ID())
			err = sp.Deliver(ctx, chunk, s.priority, syncing)
			if err != nil {
				log.Warn("ERROR in handleRetrieveRequestMsg", "err", err)
			}
			osp.LogFields(olog.Bool("delivered", true))
			return
		}
		metrics.GetOrRegisterCounter("handleRetrieveRequest.skipcheck", nil).Inc(1)

		osp.LogFields(olog.Bool("skipCheck", false))
		select {
		case streamer.deliveryC <- chunk.Address()[:]:
			metrics.GetOrRegisterCounter("handleRetrieveRequest.skipcheck.deliveryC", nil).Inc(1)
		case <-streamer.quit:
		}

	}()

	return nil
}

//Chunk delivery always uses the same message type....
type ChunkDeliveryMsg struct {
	Addr  storage.Address
	SData []byte // the stored chunk Data (incl size)
	peer  *Peer  // set in handleChunkDeliveryMsg
}

//...but swap accounting needs to disambiguate if it is a delivery for syncing or for retrieval
//as it decides based on message type if it needs to account for this message or not

//defines a chunk delivery for retrieval (with accounting)
type ChunkDeliveryMsgRetrieval ChunkDeliveryMsg

//defines a chunk delivery for syncing (without accounting)
type ChunkDeliveryMsgSyncing ChunkDeliveryMsg

// chunk delivery msg is response to retrieverequest msg
func (d *Delivery) handleChunkDeliveryMsg(ctx context.Context, sp *Peer, req *ChunkDeliveryMsg) error {
	rid := getGID()

	processReceivedChunksCount.Inc(1)

	go func() {
		log.Trace("handle.chunk.delivery", "ref", req.Addr, "peer", sp.ID(), "rid", rid)

		err := d.chunkStore.Put(ctx, storage.NewChunk(req.Addr, req.SData))
		if err != nil {
			if err == storage.ErrChunkInvalid {
				log.Error("invalid chunk delivered", "peer", sp.ID(), "chunk", req.Addr)
			}

			log.Error("err", err.Error(), "peer", sp.ID(), "chunk", req.Addr)
		}

		log.Trace("handle.chunk.delivery, done put", "ref", req.Addr, "peer", sp.ID(), "err", err, "rid", rid)
	}()

	return nil
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// RequestFromPeers sends a chunk retrieve request to a peer
// The closest peer that hasn't already been sent to is chosen
func (d *Delivery) RequestFromPeers(ctx context.Context, req *network.Request) (*enode.ID, error) {
	metrics.GetOrRegisterCounter("delivery.requestfrompeers", nil).Inc(1)

	var sp *Peer

	d.kad.EachConn(req.Addr[:], 255, func(p *network.Peer, po int) bool {
		id := p.ID()

		// skip light nodes
		if p.LightNode {
			return true
		}

		// skip peers that we have already tried
		if req.SkipPeer(id.String()) {
			rid := getGID()
			log.Trace("Delivery.RequestFromPeers: skip peer", "peer", id, "ref", req.Addr.String(), "rid", rid)
			return true
		}

		sp = d.getPeer(id)

		// sp is nil, when we encounter a peer that is not registered for delivery, i.e. doesn't support the `stream` protocol
		if sp == nil {
			return true
		}

		return false
	})

	if sp == nil {
		return nil, errors.New("no peer found") // TODO: maybe clear the peers to skip and try again, or return a failure?
	}

	// setting this value in the context creates a new span that can persist across the sendpriority queue and the network roundtrip
	// this span will finish only when delivery is handled (or times out)
	log.Trace("sending retrieve request", "ref", req.Addr, "peer", sp.ID().String())
	err := sp.Send(ctx, &RetrieveRequestMsg{
		Addr:      req.Addr,
		HopCount:  req.HopCount,
		SkipCheck: true, // this has something to do with old syncing
	})
	if err != nil {
		return nil, err
	}

	spID := sp.ID()
	return &spID, nil
}
