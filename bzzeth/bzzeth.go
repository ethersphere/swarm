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
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/spancontext"
	"github.com/ethersphere/swarm/storage"
	"golang.org/x/sync/errgroup"
)

var (
	errUnsolicitedHeader = errors.New("unsolicited header received")
	errDuplicateHeader   = errors.New("duplicate header received")
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
func New(netStore *storage.NetStore, kad *network.Kademlia) *BzzEth {
	return &BzzEth{
		peers:    newPeers(),
		netStore: netStore,
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
		p.logger.Trace("bzzeth.handleMsg")
		switch msg := msg.(type) {
		case *NewBlockHeaders:
			go b.handleNewBlockHeaders(ctx, p, msg)
		case *BlockHeaders:
			go b.handleBlockHeaders(ctx, p, msg)
		case *GetBlockHeaders:
			go b.handleGetBlockHeaders(ctx, p, msg)
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

// handleNewBlockHeaders handles new header hashes
// only request headers that are in Kad Nearest Neighbourhood
func (b *BzzEth) handleNewBlockHeaders(ctx context.Context, p *Peer, msg *NewBlockHeaders) {
	p.logger.Trace("bzzeth.handleNewBlockHeaders")

	// collect the addresses of blocks that are not in our localstore
	addresses := make([]chunk.Address, len(*msg))
	for i, h := range *msg {
		addresses[i] = h.Hash.Bytes()
		log.Trace("Received hashes ", "Header", hex.EncodeToString(h.Hash.Bytes()))
	}
	yes, err := b.netStore.Store.HasMulti(ctx, addresses...)
	if err != nil {
		log.Error("Error checking hashesh in store", "Reason", err)
		return
	}

	// collect the hashes of block headers we want
	var hashes [][]byte
	for i, y := range yes {
		// ignore hashes already present in localstore
		if y {
			continue
		}

		// collect hash based on proximity
		vhash := addresses[i]
		if wantHeaderFunc(vhash, b.kad) {
			hashes = append(hashes, vhash)
		} else {
			p.logger.Trace("ignoring header. Not in proximity ", "Address", hex.EncodeToString(addresses[i]))
		}
	}

	// request them from the offering peer and deliver in a channel
	deliveries := make(chan []byte)
	req, err := p.getBlockHeaders(ctx, hashes, deliveries, nil)
	if err != nil {
		p.logger.Error("Error sending GetBlockHeader message", "Reason", err)
		return
	}
	defer req.cancel()
	defer close(req.c)

	// this loop blocks until all delivered or context done
	// only needed to log results
	deliveredCnt := 0
	for {
		select {
		case hash, ok := <-deliveries:
			if !ok {
				p.logger.Debug("bzzeth.handleNewBlockHeaders", "delivered", deliveredCnt)
				return
			}
			deliveredCnt++
			p.logger.Trace("bzzeth.handleNewBlockHeaders", "hash", hex.EncodeToString(hash), "delivered", deliveredCnt)
			if deliveredCnt == len(req.hashes) {
				p.logger.Debug("all headers delivered", "count", deliveredCnt)
				finishDeliveryFunc(req.hashes)
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

// finishStorageFunc is used to determine if all the requested headers are stored
// this is used in testing if the headers are indeed stored in the localstore
var finishStorageFunc = finishStorage

func finishStorage(chunks []chunk.Chunk) {
	for _, c := range chunks {
		log.Trace("Header stored", "Address", c.Address().Hex())
	}
}

// finishDeliveryFunc is used to determine if all the requested headers are delivered
// this is used in testing .. otherwise it just logs trace
var finishDeliveryFunc = finishDelivery

func finishDelivery(hashes map[string]bool) {
	for addr := range hashes {
		log.Trace("Header delivered", "Address", addr)
	}
}

// handleBlockHeaders handles block headers message
func (b *BzzEth) handleBlockHeaders(ctx context.Context, p *Peer, msg *BlockHeaders) {
	p.logger.Debug("bzzeth.handleBlockHeaders", "id", msg.Rid)

	// retrieve the request for this id
	req, ok := p.requests.get(msg.Rid)
	if !ok {
		p.logger.Warn("bzzeth.handleBlockHeaders: nonexisting request id", "id", msg.Rid)
		p.Drop("nonexisting request id")
		return
	}

	// convert rlp.RawValue to bytes
	for _, h := range msg.Headers {
		displayHeader(h)
	}

	// convert rlp.RawValue to bytes
	headers := make([][]byte, len(msg.Headers))
	for i, h := range msg.Headers {
		headers[i] = h
	}

	err := b.deliverAndStoreAll(ctx, req, headers)
	if err != nil {
		p.logger.Warn("bzzeth.handleBlockHeaders: fatal dropping peer", "id", msg.Rid, "err", err)
		p.Drop("error on deliverAndStoreAll")
	}
}

// debug function to display header contents
func displayHeader(h []byte) {
	var hdr types.Header
	err := rlp.DecodeBytes(h, &hdr)
	if err != nil {
		log.Error("Could not decode header")
		return
	}
	log.Trace("Header ", "ParentHash", hdr.ParentHash.Hex())
	log.Trace("Header ", "UncleHash", hdr.UncleHash.Hex())
	log.Trace("Header ", "Coinbase", hdr.Coinbase.Hex())
	log.Trace("Header ", "Root", hdr.Root.Hex())
	log.Trace("Header ", "TxHash", hdr.TxHash.Hex())
	log.Trace("Header ", "ReceiptHash", hdr.ReceiptHash.Hex())
	log.Trace("Header ", "MixDigest", hdr.MixDigest.Hex())

	log.Trace("Header ", "Difficulty", hdr.Difficulty)
	log.Trace("Header ", "Number", hdr.Number)
	log.Trace("Header ", "GasLimit", hdr.GasLimit)
	log.Trace("Header ", "GasUsed", hdr.GasUsed)
	log.Trace("Header ", "Time", time.Unix(int64(hdr.Time), 0))
}

// Validates and headers asynchronously and stores the valid chunks in one go
func (b *BzzEth) deliverAndStoreAll(ctx context.Context, req *request, headers [][]byte) error {
	chunks := make([]chunk.Chunk, 0)
	var chunkL sync.RWMutex
	var wg errgroup.Group
	for _, h := range headers {
		hdr := make([]byte, len(h))
		copy(hdr, h)
		wg.Go(func() error {
			ch, err := b.validateHeader(ctx, hdr, req)
			if err != nil {
				return err
			}
			chunkL.Lock()
			defer chunkL.Unlock()
			chunks = append(chunks, ch)
			return nil
		})
	}
	// finish storage is used mostly in testing
	// in normal scenario.. it just logs Trace
	defer finishStorageFunc(chunks)

	// wait for all validations to get over and close the channels
	err := wg.Wait()

	// Store all the valid header chunks in one shot
	storeErr := b.storeChunks(ctx, chunks)
	if storeErr != nil {
		return err
	}

	// Pass on the validation error if any
	if err != nil {
		return err
	}

	return nil
}

func (b *BzzEth) storeChunks(ctx context.Context, chunks []chunk.Chunk) error {
	// Store all the valid header chunks in one shot
	results, err := b.netStore.Put(ctx, chunk.ModePutUpload, chunks...)
	if err != nil {
		for i := range results {
			log.Error("bzzeth.store", "hash", chunks[i].Address().Hex(), "err", err)
			// ignore all other errors, but invalid chunk incurs peer drop
			if err == chunk.ErrChunkInvalid {
				return err
			}
		}
	}
	log.Debug("Stored all headers ", "count", len(chunks))
	return nil
}

// validateHeader check for correctness and validity of the header
// this also informs the delivery channel about the received header
func (b *BzzEth) validateHeader(ctx context.Context, header []byte, req *request) (chunk.Chunk, error) {
	ch := newChunk(header)
	headerAlreadyReceived, expected := isHeaderExpected(req, ch.Address().Hex())
	if expected {
		if headerAlreadyReceived {
			// header already received
			return nil, errDuplicateHeader
		} else {
			setHeaderAsReceived(req, ch.Address().Hex())
			// This channel is used ot track deliveries
			if req.c != nil {
				req.c <- ch.Address()
			}
			// This channel is used to give back the header to the requesting eth node
			if req.giveBackC != nil {
				req.giveBackC <- header
			}
			return ch, nil
		}
	} else {
		// header is not present in the request hash.
		return nil, errUnsolicitedHeader
	}
}

// Checks if the given hash is expected in this request
func isHeaderExpected(req *request, addr string) (rcvdFlag bool, ok bool) {
	req.lock.RLock()
	defer req.lock.RUnlock()
	rcvdFlag, ok = req.hashes[addr]
	return rcvdFlag, ok
}

// Set the given hash as received in the request
func setHeaderAsReceived(req *request, addr string) {
	req.lock.Lock()
	defer req.lock.Unlock()
	req.hashes[addr] = true
}

// newChunk creates a new content addressed chunk from data using Keccak256  SHA3 hash
func newChunk(data []byte) chunk.Chunk {
	hash := crypto.Keccak256(data)
	return chunk.NewChunk(hash, data)
}

var arrangeHeaderFunc = arrangeHeader

// arrangeHeader is used in testing the response headers delivered to the light client
// This function does nothing in normal operation, but in test case, it arranges the headers
// as per the position of the hashes received, soas to become predictable
func arrangeHeader(hashes [][]byte, headers [][]byte) [][]byte {
	return headers
}

// handles GetBlockHeader requests, in the protocol handler this call is asynchronous
// so it is safe to have it run until delivery is finished
func (b *BzzEth) handleGetBlockHeaders(ctx context.Context, p *Peer, msg *GetBlockHeaders) {
	p.logger.Debug("bzzeth.handleGetBlockHeaders", "id", msg.Rid, "hash", hex.EncodeToString(msg.Hashes[0]))
	total := len(msg.Hashes)
	ctx, osp := spancontext.StartSpan(ctx, "bzzeth.handleGetBlockHeaders")
	defer osp.Finish()

	deliveries := make(chan []byte)
	defer close(deliveries)
	trigger := make(chan chan [][]byte)
	defer close(trigger)
	batches := make(chan [][]byte)
	defer close(batches)

	// deliver in batches, this blocks until total number of requests are delivered or considered not found
	go readToBatches(deliveries, trigger)

	// asynchronously request all headers as swarm chunks
	go b.requestAll(ctx, deliveries, msg.Hashes, p)

	// Send a trigger to create a batch
	trigger <- batches

	deliveredCnt := 0
	var err error
	// this loop terminates if
	// - batches channel is closed (because the underlying deliveries channel is closed) OR
	// - context is done
	// the implementation aspires to send as many as possible as early as possible
DELIVERY:
	for headers := range batches {
		deliveredCnt += len(headers)
		headers = arrangeHeaderFunc(msg.Hashes, headers)

		// convert bytes to rlp.RawValue
		rawHeaders := make([]rlp.RawValue, len(headers))
		for i, h := range headers {
			rawHeaders[i] = h
			displayHeader(h)
		}

		p.logger.Debug("sending headers", "count", len(rawHeaders))
		if err = p.Send(ctx, &BlockHeaders{
			Rid:     uint32(msg.Rid),
			Headers: rawHeaders,
		}); err != nil { // in case of a send error, the peer will disconnect so can safely return
			break DELIVERY
		}
		// Break if all the headers are delivered
		if deliveredCnt >= total {
			break DELIVERY
		}
		select {
		case trigger <- batches: // signal that we are ready for another batch
		case <-ctx.Done():
			break DELIVERY
		}
	}
	p.logger.Debug("bzzeth.handleGetBlockHeaders", "id", msg.Rid, "total", total, "delivered", deliveredCnt, "err", err)
	if err == nil && deliveredCnt < total { // if there was no send error and we deliver less than requested
		err := p.Send(ctx, &BlockHeaders{Rid: uint32(msg.Rid)}) // it is prudent to send an empty BlockHeaders message
		if err != nil {
			p.logger.Error("could not send empty BlockHeader")
		}
	}
	p.logger.Debug("bzzeth.handleGetBlockHeaders: sent all headers", "id", msg.Rid)
}

var batchWait = 100 * time.Millisecond // time to wait for collecting headers in a batch
var minBatchSize = 1                   // minimum headers in a batch

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
			if len(buffer) >= minBatchSize { // if buffer is not empty
				trigger = out  // allow sending batch
				continue BATCH // wait till next batch can send
			}
			time.Sleep(batchWait) // otherwise wait and continue
		}
	}
}

var skipHeaderFromSwarm = false

// getBlockHeaderBzz retrieves a block header by its hash from swarm
func (b *BzzEth) getBlockHeaderBzz(ctx context.Context, hash []byte) ([]byte, error) {
	if skipHeaderFromSwarm {
		return nil, errors.New("ignoring headers in swarm because of testcase")
	}
	req := &storage.Request{
		Addr:   hash,
		Origin: b.netStore.LocalID,
	}
	chunk, err := b.netStore.Get(ctx, chunk.ModeGetRequest, req)
	if err != nil {
		return nil, err
	}
	return chunk.Data(), nil
}

// requestAll requests each hash and channel
func (b *BzzEth) requestAll(ctx context.Context, deliveries chan []byte, hashes [][]byte, rcvdPeer *Peer) {
	ctx, cancel := context.WithTimeout(ctx, timeouts.FetcherGlobalTimeout)
	defer cancel()

	// missingHeaders collects hashes of headers not found within swarm
	// ie., the hashes to request from the eth full nodes
	missingHeaders := make(chan []byte)
	defer close(missingHeaders)
	var wg sync.WaitGroup

BZZ:
	for _, h := range hashes {
		hdr := make([]byte, len(h))
		copy(hdr, h)
		wg.Add(1)

		go func() {
			defer wg.Done()
			header, err := b.getBlockHeaderBzz(ctx, hdr)
			if err != nil {
				log.Debug("bzzeth.requestAll: netstore.Get can not retrieve chunk", "ref", hex.EncodeToString(h), "err", err)
				select {
				case missingHeaders <- hdr: // fallback: request header from eth peers
				case <-ctx.Done():
				}
				return
			}
			// deliver the headers received from Swarm
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
	wg.Add(1)
	go b.getBlockHeadersEth(ctx, missingHeaders, deliveries, &wg, rcvdPeer)

	// wait till all hashes are requested from swarm OR from another eth node,
	wg.Wait()

}

// getBlockHeadersEth manages fetching headers from ethereum bzzeth nodes
// This is part of the response to GetBlockHeaders requests by bzzeth light/syncing nodes
// As a fallback after header retrieval from local storage and swarm network are unsuccessful
// When called, it
// - reads requested header hashes from a channel (headerC) and
// - creates batch requests and sends them to an adequate bzzeth peer
// - channels the responses into a delivery channel (deliveries)
func (b *BzzEth) getBlockHeadersEth(ctx context.Context, headersC, giveBackC chan []byte, pwg *sync.WaitGroup, rcvdPeer *Peer) {
	log.Debug("getting missing headers from another ETH node", "count", len(giveBackC))
	defer pwg.Done() // unblock the parent so that i can continue

	// read header requests into batches
	readNext := make(chan chan [][]byte)
	batches := make(chan [][]byte)
	go readToBatches(headersC, readNext)
	readNext <- batches

	// send GetBlockHeader requests to adequate bzzeth peers
	// this loop terminates when batches channel is closed as a result of input headersC being closed
	requiredCount := 0
	var wg sync.WaitGroup
	deliveryC := make(chan []byte)
	total := len(headersC)
	defer close(deliveryC)
	for header := range batches {
		p := b.peers.getEth(rcvdPeer) // find candidate peer to serve the headers
		if p == nil {                 // if no peer found just skip the batch TODO: smarter retry?
			continue
		}
		// initiate request with the chosen peer
		fmt.Println("trying to get the header from peer ", p.ID())
		req, err := p.getBlockHeaders(ctx, header, deliveryC, giveBackC)
		if err != nil { // in case of failure, no retries TODO: smarter retry?
			continue
		}

		wg.Add(1)
		go b.getDelivery(ctx, req, &wg)
		requiredCount++

		if requiredCount >= total {
			break
		}
	}
	wg.Wait()
}

func (b *BzzEth) getDelivery(ctx context.Context, req *request, wg *sync.WaitGroup) {
	defer wg.Done()
	defer req.cancel()
	for {
		select {
		case hash, ok := <-req.c:
			if !ok {
				log.Debug("could not get delivery of a missing header")
				return
			}
			if _, ok := req.hashes[hex.EncodeToString(hash)]; ok {
				log.Debug("delivered missing header ", "header", hex.EncodeToString(hash))
				return
			}
		case <-ctx.Done():
			return
		}
	}
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
