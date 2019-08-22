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

package bzzeth

import (
	"context"
	"errors"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/spancontext"
	"github.com/ethersphere/swarm/storage"
	// "github.com/ethersphere/swarm/storage/localstore"
)

var (
	errRcvdMsgFromSwarmNode = errors.New("received message from Swarm node")
)

// BzzEth implements node.Service
var _ node.Service = &BzzEth{}

// BzzEth is a global module handling ethereum state on swarm
type BzzEth struct {
	peers    *peers            // bzzeth peer pool
	netStore *storage.NetStore // netstore to retrieve and store
	kad      *network.Kademlia // kademlia to determine if a header chunk belongs to us
	quit     chan struct{}     // quit channel to close go routines
}

// New constructs the BzzEth node service
func New(ns *storage.NetStore, kad *network.Kademlia) *BzzEth {
	return &BzzEth{
		peers:    newPeers(),
		netStore: ns,
		kad:      kad,
		quit:     make(chan struct{}),
	}
}

// Run is the bzzeth protocol run function.
// - creates a peer
// - checks if it is a swarm node, put the protocol in idle mode
// - performs handshake
// - adds peer to the peerpool
// - starts incoming message handler loop
func (b *BzzEth) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, Spec)
	bp := NewPeer(peer)

	// perform handshake and register if peer serves headers
	handshake, err := bp.Handshake(context.TODO(), Handshake{ServeHeaders: true}, nil)
	if err != nil {
		return err
	}
	bp.serveHeaders = handshake.(*Handshake).ServeHeaders
	log.Debug("handshake", "hs", handshake, "peer", bp)

	b.peers.add(bp)
	defer b.peers.remove(bp)

	// This protocol is all about interaction between an Eth node and a Swarm Node.
	// If another swarm node tries to connect then the protocol goes into idle
	if isSwarmNodeFunc(bp) {
		return peer.Run(b.handleMsgFromSwarmNode(bp))
	}

	return peer.Run(b.handleMsg(bp))
}

// handleMsg is the message handler that delegates incoming messages
// handlers are called asynchronously so handler calls do not block incoming msg processing
func (b *BzzEth) handleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		p.logger.Debug("bzzeth.handleMsg")
		switch msg := msg.(type) {
		case *NewBlockHeaders:
			go b.handleNewBlockHeaders(ctx, p, msg)
		case *GetBlockHeaders:
			go b.handleGetBlockHeaders(ctx, p, msg)
		case *BlockHeaders:
			go b.handleBlockHeaders(ctx, p, msg)
		}
		return nil
	}
}


// handleMsgFromSwarmNode is used in the case if this node is connected to a Swarm node
// If any message is received in this case, the peer needs to be dropped
func (b *BzzEth) handleMsgFromSwarmNode(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		p.logger.Warn("bzzeth.handleMsgFromSwarmNode")
		return errRcvdMsgFromSwarmNode
	}
}

// handles new header hashes - strategy; only request headers that are in Kad Nearest Neighbourhood
func (b *BzzEth) handleNewBlockHeaders(ctx context.Context, p *Peer, msg *NewBlockHeaders) {
	p.logger.Debug("bzzeth.handleNewBlockHeaders")
	// collect the hashes of block headers we want
	var hashes [][]byte
	for _, h := range msg.Headers {
		if wantHeaderFunc(h.Hash, b.kad) {
			hashes = append(hashes, h.Hash)
		}
	}

	// request them from the offering peer and deliver in a channel
	deliveries := make(chan []byte)
	req, err := p.getBlockHeaders(ctx, hashes, deliveries)
	defer req.cancel()
	deliveredCnt := 0
	// this loop blocks until all delivered or context done
	// only needed to log results
	for {
		select {
		case _, ok := <-deliveries:
			if !ok {
				p.logger.Debug("bzzeth.handleNewBlockHeaders", "delivered", deliveredCnt)
				return
			}
			deliveredCnt++
			if deliveredCnt == len(hashes) {
				p.logger.Debug("bzzeth.handleNewBlockHeaders", "delivered", deliveredCnt)
				return
			}
		case <-ctx.Done():
			p.logger.Debug("bzzeth.handleNewBlockHeaders", "delivered", deliveredCnt, "err", err)
			return
		}
	}
}

// wantHeaderFunc is used to determine if we need a particular header offered as latest
// by an eth fullnode
// tests reassign this to control
var wantHeaderFunc = wantHeader

// wantHeader returns true iff the hash argument falls in the NN of kademlia
func wantHeader(hash []byte, kad *network.Kademlia) bool {
	return chunk.Proximity(kad.BaseAddr(), hash) >= kad.NeighbourhoodDepth()
}

// requestAll requests each hash and channel
func (b *BzzEth) requestAll(ctx context.Context, deliveries chan []byte, hashes [][]byte) {
	ctx, cancel := context.WithTimeout(ctx, timeouts.FetcherGlobalTimeout)
	defer cancel()

	// missingHeaders collects hashes of headers not found within swarm
	// ie., the hashes to request from the eth full nodes
	missingHeaders := make(chan []byte)

	wg := sync.WaitGroup{}
	defer close(deliveries)
	wg.Add(1)
BZZ:
	for _, h := range hashes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			header, err := b.getBlockHeaderBzz(ctx, h)
			if err != nil {
				log.Debug("bzzeth.requestAll: netstore.Get can not retrieve chunk", "ref", hex.EncodeToString(h), "err", err)
				select {
				case missingHeaders <- h: // fallback: request header from eth peers
				case <-ctx.Done():
				}
				return
			}
			select {
			case deliveries <- header:
			case <-ctx.Done():
			}
		}()
		select {
		case <-ctx.Done():
			break BZZ
		default:
		}
	}

	// fall back to retrieval from eth clients
	// collect missing block header hashes
	// terminates after missingHeaders is read and closed or context is done
	go b.getBlockHeadersEth(ctx, missingHeaders, deliveries)

	// wait till all hashes are requested from swarm, then close missingHeaders channel
	// this cannot block as this function is called async
	wg.Done()
	wg.Wait()
	close(missingHeaders)
}

// getBlockHeadersEth manages fetching headers from ethereum bzzeth nodes
// This is part of the response to GetBlockHeaders requests by bzzeth light/syncing nodes
// As a fallback after header retrieval from local storage and swarm network are unsuccessful
// When called, it
// - reads requested header hashes from a channel (headerC) and
// - creates batch requests and sends them to an adequate bzzeth peer
// - channels the responses into a delivery channel (deliveries)
func (b *BzzEth) getBlockHeadersEth(ctx context.Context, headersC, deliveries chan []byte) {
	// read header requests into batches
	readNext := make(chan chan [][]byte)
	batches := make(chan [][]byte)
	go readToBatches(headersC, readNext)
	readNext <- batches

	// send GetBlockHeader requests to adequate bzzeth peers
	// this loop terminates when batches channel is closed as a result of input headersC being closed
	var reqs []*request
	for headers := range batches {
		p := b.peers.getEth() // find candidate peer to serve the headers
		if p == nil {         // if no peer found just skip the batch TODO: smarter retry?
			continue
		}
		// initiate request with the chosen peer
		req, err := p.getBlockHeaders(ctx, headers, deliveries)
		if err != nil { // in case of failure, no retries TODO: smarter retry?
			continue
		}
		reqs = append(reqs, req) // remember the request so that it can be cancelled
	}
	cancelAll(reqs...)
}

// cancelAll cancels all requests given as arguments
func cancelAll(reqs ...*request) {
	for _, req := range reqs {
		req.cancel()
	}
}

// getBlockHeaderBzz retrieves a block header by its hash from swarm
func (b *BzzEth) getBlockHeaderBzz(ctx context.Context, hash []byte) ([]byte, error) {
	req := &storage.Request{
		Addr: hash,
		// Origin: b.ID(),
	}
	chunk, err := b.netStore.Get(ctx, chunk.ModeGetRequest, req)
	if err != nil {
		return nil, err
	}
	return chunk.Data(), nil
}

// handles GetBlockHeader requests, in the protocol handler this call is asynchronous
// so it is safe to have it run until delivery is finished
func (b *BzzEth) handleGetBlockHeaders(ctx context.Context, p *Peer, msg *GetBlockHeaders) {
	total := len(msg.Hashes)
	p.logger.Debug("bzzeth.handleGetBlockHeaders", "id", msg.ID)
	ctx, osp := spancontext.StartSpan(ctx, "bzzeth.handleGetBlockHeaders")
	defer osp.Finish()

	// deliver in batches, this blocks until total number of requests are delivered or considered not found
	deliveries := make(chan []byte)
	trigger := make(chan chan [][]byte)
	batches := make(chan [][]byte)
	defer close(trigger)
	go readToBatches(deliveries, trigger)

	// asynchronously request all headers as swarm chunks
	go b.requestAll(ctx, deliveries, msg.Hashes)
	deliveredCnt := 0
	var err error
	// this loop terminates if
	// - batches channel is closed (because the underlying deliveries channel is closed) OR
	// - context is done
	// the implementation aspires to send as many as possible as early as possible
DELIVERY:
	for headers := range batches {
		deliveredCnt += len(headers)
		if err = p.Send(ctx, &BlockHeaders{
			ID:      msg.ID,
			Headers: headers,
		}); err != nil { // in case of a send error, the peer will disconnect so can safely return
			break DELIVERY
		}
		select {
		case trigger <- batches: // signal that we are ready for another batch
		case <-ctx.Done():
			break DELIVERY
		}
	}

	p.logger.Debug("bzzeth.handleGetBlockHeaders", "id", msg.ID, "total", total, "delivered", deliveredCnt, "err", err)

	if err == nil && deliveredCnt < total { // if there was no send error and we deliver less than requested
		p.Send(ctx, &BlockHeaders{ID: msg.ID}) // it is prudent to send an empty BlockHeaders message
	}
}

// handleBlockHeaders handles block headers message
func (b *BzzEth) handleBlockHeaders(ctx context.Context, p *Peer, msg *BlockHeaders) {
	p.logger.Debug("bzzeth.handleBlockHeaders", "id", msg.ID)

	// retrieve the request for this id :TODO:
	req, ok := p.requests.get(msg.ID)
	if !ok {
		p.logger.Warn("bzzeth.handleBlockHeaders: nonexisting request id", "id", msg.ID)
		p.Drop()
		return
	}
	err := b.deliverAll(ctx, req.c, msg.Headers)
	if err != nil {
		p.logger.Warn("bzzeth.handleBlockHeaders: fatal dropping peer", "id", msg.ID, "err", err)
		p.Drop()
	}
}

// store delivery
func (b *BzzEth) deliverAll(ctx context.Context, deliveries chan []byte, headers [][]byte) error {
	errc := make(chan error, 1)       // only the first error propagetes
	go b.storeAll(ctx, errc, headers) // storing all heades, pro
	return <-errc
}

// stores all headers asynchronously, reports store error on errc
func (b *BzzEth) storeAll(ctx context.Context, errc chan error, headers [][]byte) {
	defer close(errc)
	for _, h := range headers {
		h := h
		go func() {
			// TODO: unsolicited header validation should come here
			// TODO: header validation should come here
			if err := b.store(ctx, h); err != nil {
				select {
				case errc <- err: // buffered channel,
				default: //  there is already an error, ignore
				}
			}
		}()
	}
}

// store stores a header as a chunk, returns error if and only if invalid chunk
func (b *BzzEth) store(ctx context.Context, header []byte) error {
	ch := newChunk(header)
	_, err := b.netStore.Put(ctx, chunk.ModePutSync, ch)
	if err != nil {
		log.Warn("bzzeth.store", "hash", ch.Address().Hex(), "err", err)
		// ignore all other errors, but invalid chunk incurs peer drop
		if err == chunk.ErrChunkInvalid {
			return err
		}
	}
	return nil
}

// newChunk creates a new content addressed chunk from data using Keccak256  SHA3 hash
func newChunk(data []byte) chunk.Chunk {
	hash := crypto.Keccak256(data)
	return chunk.NewChunk(hash, data)

}

// Protocols returns the p2p protocol
func (b *BzzEth) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    Spec.Name,
			Version: Spec.Version,
			Length:  Spec.Length(),
			Run:     b.Run,
		},
	}
}

// APIs return APIs defined on the node service
func (b *BzzEth) APIs() []rpc.API {
	return nil
}

// Start starts the BzzEth node service
func (b *BzzEth) Start(server *p2p.Server) error {
	log.Info("bzzeth starting...")
	return nil
}

// Stop stops the BzzEth node service
func (b *BzzEth) Stop() error {
	log.Info("bzzeth shutting down...")
	close(b.quit)
	return nil
}

var batchWait = 100 * time.Millisecond

// readToBatches reads items from an input channel into a buffer and
// sends non-empty buffers on a channel read from the out
func readToBatches(in chan []byte, out chan chan [][]byte) {
	var buffer [][]byte
	var trigger chan chan [][]byte
BATCH:
	for {
		select {
		case batches := <-trigger: // new batch channel available
			if batches == nil { // terminate if batches channel is closed, no more batches accepted
				return
			}
			batches <- buffer // otherwise write buffer into batch channel
			if in == nil {    // terminate if in channel is already closed, sent last batch
				return
			}
			buffer = nil  // otherwise start new buffer
			trigger = nil // block this case: disallow new batches until enough in buffer

		case item, more := <-in: // reading input
			if !more {
				in = nil       // block this case: disallow read from closed channel
				continue BATCH // wait till last batch can send
			}
			// otherwise collect item in buffer
			buffer = append(buffer, item)

		default:
			if len(buffer) > 0 { // if buffer is not empty
				trigger = out  // allow sending batch
				continue BATCH // wait till next batch can send
			}
			time.Sleep(batchWait) // otherwise wait and continue
		}
	}
}
