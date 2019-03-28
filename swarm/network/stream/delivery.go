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
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/chunk"
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
	processReceivedChunksCount    = metrics.GetOrRegisterCounter("network.stream.received_chunks.count", nil)
	handleRetrieveRequestMsgCount = metrics.GetOrRegisterCounter("network.stream.handle_retrieve_request_msg.count", nil)
	retrieveChunkFail             = metrics.GetOrRegisterCounter("network.stream.retrieve_chunks_fail.count", nil)

	lastReceivedChunksMsg = metrics.GetOrRegisterGauge("network.stream.received_chunks", nil)
)

type Delivery struct {
	netStore *network.NetStore
	kad      *network.Kademlia
	getPeer  func(enode.ID) *Peer
}

func NewDelivery(kad *network.Kademlia, netStore *network.NetStore) *Delivery {
	return &Delivery{
		netStore: netStore,
		kad:      kad,
	}
}

// SwarmChunkServer implements Server
type SwarmChunkServer struct {
	deliveryC  chan []byte
	batchC     chan []byte
	netStore   *network.NetStore
	currentLen uint64
	quit       chan struct{}
}

// NewSwarmChunkServer is SwarmChunkServer constructor
func NewSwarmChunkServer(chunkStore *network.NetStore) *SwarmChunkServer {
	s := &SwarmChunkServer{
		deliveryC: make(chan []byte, deliveryCap),
		batchC:    make(chan []byte),
		netStore:  chunkStore,
		quit:      make(chan struct{}),
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
	//TODO: this should be localstore, not netstore?
	r := &network.Request{
		Addr:     storage.Address(key),
		Origin:   enode.ID{},
		HopCount: 0,
	}
	chunk, err := s.netStore.Get(ctx, r)
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

	ctx, cancel := context.WithTimeout(ctx, network.FetcherGlobalTimeout)

	go func() {
		select {
		case <-ctx.Done():
		case <-streamer.quit:
		}
		cancel()
	}()

	go func() {
		defer osp.Finish()

		r := &network.Request{
			Addr:     req.Addr,
			Origin:   sp.ID(),
			HopCount: req.HopCount,
		}
		chunk, err := d.netStore.Get(ctx, r)
		if err != nil {
			retrieveChunkFail.Inc(1)
			log.Debug("ChunkStore.Get can not retrieve chunk", "peer", sp.ID().String(), "addr", req.Addr, "hopcount", req.HopCount, "err", err)
			return
		}

		log.Trace("retrieve request, delivery", "ref", req.Addr, "peer", sp.ID())
		err = sp.Deliver(ctx, chunk, s.priority, false)
		if err != nil {
			log.Warn("ERROR in handleRetrieveRequestMsg", "err", err)
		}
		osp.LogFields(olog.Bool("delivered", true))
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

	// record the last time we received a chunk delivery message
	lastReceivedChunksMsg.Update(time.Now().UnixNano())

	go func() {
		log.Trace("handle.chunk.delivery", "ref", req.Addr, "peer", sp.ID(), "rid", rid)

		err := d.netStore.Put(ctx, storage.NewChunk(req.Addr, req.SData))
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
func (d *Delivery) RequestFromPeers(ctx context.Context, req *network.Request, localID enode.ID) (*enode.ID, error) {
	metrics.GetOrRegisterCounter("delivery.requestfrompeers", nil).Inc(1)

	rid := getGID()

	var sp *Peer

	var err error

	depth := d.kad.NeighbourhoodDepth()

	d.kad.EachConn(req.Addr[:], 255, func(p *network.Peer, po int) bool {
		id := p.ID()

		// skip light nodes
		if p.LightNode {
			return true
		}

		// do not send request back to peer who asked us. maybe merge with SkipPeer at some point
		if req.Origin.String() == id.String() {
			return true
		}

		// skip peers that we have already tried
		if req.SkipPeer(id.String()) {
			rid := getGID()
			log.Trace("Delivery.RequestFromPeers: skip peer", "peer", id, "ref", req.Addr.String(), "rid", rid)
			return true
		}

		// if origin is farther away from req.Addr and origin is not in our depth
		prox := chunk.Proximity(req.Addr, d.kad.BaseAddr())
		// proximity between the req.Addr and our base addr
		if po < depth && prox >= depth {
			log.Trace("Delivery.RequestFromPeers: skip peer because depth", "po", po, "depth", depth, "peer", id, "ref", req.Addr.String(), "rid", rid)

			err = fmt.Errorf("not going outside of depth; ref=%s po=%v depth=%v prox=%v", req.Addr.String(), po, depth, prox)
			return false
		}

		sp = d.getPeer(id)

		// sp is nil, when we encounter a peer that is not registered for delivery, i.e. doesn't support the `stream` protocol
		if sp == nil {
			return true
		}

		return false
	})

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if sp == nil {
		return nil, errors.New("no peer found") // TODO: maybe clear the peers to skip and try again, or return a failure?
	}

	// setting this value in the context creates a new span that can persist across the sendpriority queue and the network roundtrip
	// this span will finish only when delivery is handled (or times out)
	r := &RetrieveRequestMsg{
		Addr:      req.Addr,
		HopCount:  req.HopCount + 1,
		SkipCheck: true, // this has something to do with old syncing
	}
	log.Trace("sending retrieve request", "ref", r.Addr, "peer", sp.ID().String(), "origin", localID)
	err = sp.Send(ctx, r)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	spID := sp.ID()
	return &spID, nil
}
