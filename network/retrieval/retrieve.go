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

package retrieval

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/spancontext"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/swap"
)

var (
	// Compile time interface check
	_ node.Service = &Retrieval{}

	// Metrics
	processReceivedChunksCount    = metrics.NewRegisteredCounter("network.retrieve.received_chunks_handled", nil)
	handleRetrieveRequestMsgCount = metrics.NewRegisteredCounter("network.retrieve.handle_retrieve_request_msg", nil)
	retrieveChunkFail             = metrics.NewRegisteredCounter("network.retrieve.retrieve_chunks_fail", nil)
	unsolicitedChunkDelivery      = metrics.NewRegisteredCounter("network.retrieve.unsolicited_delivery", nil)

	retrievalPeers = metrics.GetOrRegisterGauge("network.retrieve.peers", nil)

	spec = &protocols.Spec{
		Name:       "bzz-retrieve",
		Version:    2,
		MaxMsgSize: 10 * 1024 * 1024,
		Messages: []interface{}{
			ChunkDelivery{},
			RetrieveRequest{},
		},
	}

	ErrNoPeerFound = errors.New("no peer found")
)

// Price is the method through which a message type marks itself
// as implementing the protocols.Price protocol and thus
// as swap-enabled message
func (rr *RetrieveRequest) Price() *protocols.Price {
	return &protocols.Price{
		Value:   swap.RetrieveRequestPrice,
		PerByte: false,
		Payer:   protocols.Sender,
	}
}

// Price is the method through which a message type marks itself
// as implementing the protocols.Price protocol and thus
// as swap-enabled message
func (cd *ChunkDelivery) Price() *protocols.Price {
	return &protocols.Price{
		Value:   swap.ChunkDeliveryPrice,
		PerByte: true,
		Payer:   protocols.Receiver,
	}
}

// Retrieval holds state and handles protocol messages for the `bzz-retrieve` protocol
type Retrieval struct {
	netStore    *storage.NetStore
	baseAddress *network.BzzAddr
	kad         *network.Kademlia
	kademliaLB  *network.KademliaLoadBalancer
	mtx         sync.RWMutex       // protect peer map
	peers       map[enode.ID]*Peer // compatible peers
	spec        *protocols.Spec    // protocol spec
	logger      log.Logger         // custom logger to append a basekey
	quit        chan struct{}      // shutdown channel
}

// New returns a new instance of the retrieval protocol handler
func New(kad *network.Kademlia, ns *storage.NetStore, baseKey *network.BzzAddr, balance protocols.Balance) *Retrieval {
	r := &Retrieval{
		netStore:    ns,
		baseAddress: baseKey,
		kad:         kad,
		kademliaLB:  network.NewKademliaLoadBalancer(kad, false),
		peers:       make(map[enode.ID]*Peer),
		spec:        spec,
		logger:      log.NewBaseAddressLogger(baseKey.ShortString()),
		quit:        make(chan struct{}),
	}
	if balance != nil && !reflect.ValueOf(balance).IsNil() {
		// swap is enabled, so setup the hook
		r.spec.Hook = protocols.NewAccounting(balance)
	}
	return r
}

func (r *Retrieval) addPeer(p *Peer) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.peers[p.ID()] = p
	retrievalPeers.Update(int64(len(r.peers)))
}

func (r *Retrieval) removePeer(p *Peer) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	delete(r.peers, p.ID())
	retrievalPeers.Update(int64(len(r.peers)))
}

func (r *Retrieval) getPeer(id enode.ID) *Peer {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.peers[id]
}

// Run is being dispatched when 2 nodes connect
func (r *Retrieval) Run(bp *network.BzzPeer) error {
	sp := NewPeer(bp, r.baseAddress)
	r.addPeer(sp)
	defer r.removePeer(sp)

	return sp.Peer.Run(r.handleMsg(sp))
}

func (r *Retrieval) handleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		switch msg := msg.(type) {
		case *RetrieveRequest:
			return r.handleRetrieveRequest(ctx, p, msg)
		case *ChunkDelivery:
			return r.handleChunkDelivery(ctx, p, msg)
		}
		return nil
	}
}

// getOriginPo returns the originPo if the incoming Request has an Origin
// if our node is the first node that requests this chunk, then we don't have an Origin,
// and return -1
// this is used only for tracing, and can probably be refactor so that we don't have to
// iterater over Kademlia
func (r *Retrieval) getOriginPo(req *storage.Request) int {
	r.logger.Trace("retrieval.getOriginPo", "req.Addr", req.Addr)
	originPo := -1

	r.kad.EachConn(req.Addr[:], 255, func(p *network.Peer, po int) bool {
		id := p.ID()

		// get po between chunk and origin
		if bytes.Equal(req.Origin.Bytes(), id.Bytes()) {
			originPo = po
			return false
		}

		return true
	})

	return originPo
}

// findPeerLB finds a peer we need to ask for a specific chunk from according to our kademlia load balancer
func (r *Retrieval) findPeerLB(ctx context.Context, req *storage.Request) (retPeer *network.Peer, err error) {
	r.logger.Trace("retrieval.findPeer", "req.Addr", req.Addr)
	osp, _ := ctx.Value("remote.fetch").(opentracing.Span)

	// originPo - proximity of the node that made the request; -1 if the request originator is our node;
	// myPo - this node's proximity with the requested chunk
	// selectedPeerPo - kademlia suggested node's proximity with the requested chunk (computed further below)
	originPo := r.getOriginPo(req)
	myPo := chunk.Proximity(req.Addr, r.kad.BaseAddr())
	selectedPeerPo := -1

	depth := r.kad.NeighbourhoodDepth()

	if osp != nil {
		osp.LogFields(olog.Int("originPo", originPo))
		osp.LogFields(olog.Int("depth", depth))
		osp.LogFields(olog.Int("myPo", myPo))
	}

	// do not forward requests if origin proximity is bigger than our node's proximity
	// this means that origin is closer to the chunk
	if originPo > myPo {
		return nil, errors.New("not forwarding request, origin node is closer to chunk than this node")
	}

	r.kademliaLB.EachBinDesc(req.Addr, func(bin network.LBBin) bool {
		for _, lbPeer := range bin.LBPeers {
			id := lbPeer.Peer.ID()

			// skip peer that does not support retrieval
			if !lbPeer.Peer.HasCap(r.spec.Name) {
				continue
			}

			// do not send request back to peer who asked us. maybe merge with SkipPeer at some point
			if bytes.Equal(req.Origin.Bytes(), id.Bytes()) {
				continue
			}

			// skip peers that we have already tried
			if req.SkipPeer(id.String()) {
				continue
			}

			if myPo < depth { //  chunk is NOT within the neighbourhood
				if bin.ProximityOrder <= myPo { // always choose a peer strictly closer to chunk than us
					return false
				}
			} else { // chunk IS WITHIN neighbourhood
				if bin.ProximityOrder < depth { // do not select peer outside the neighbourhood. But allows peers further from the chunk than us
					return false
				} else if bin.ProximityOrder <= originPo { // avoid loop in neighbourhood, so not forward when a request comes from the neighbourhood
					return false
				}
			}

			// if selected peer is not in the depth (2nd condition; if depth <= po, then peer is in nearest neighbourhood)
			// and they have a lower po than ours, return error
			if bin.ProximityOrder < myPo && depth > bin.ProximityOrder {
				err = fmt.Errorf("not asking peers further away from origin; ref=%s originpo=%v po=%v depth=%v myPo=%v", req.Addr.String(), originPo, bin.ProximityOrder, depth, myPo)
				return false
			}

			// if chunk falls in our nearest neighbourhood (1st condition), but suggested peer is not in
			// the nearest neighbourhood (2nd condition), don't forward the request to suggested peer
			if depth <= myPo && depth > bin.ProximityOrder {
				err = fmt.Errorf("not going outside of depth; ref=%s originpo=%v po=%v depth=%v myPo=%v", req.Addr.String(), originPo, bin.ProximityOrder, depth, myPo)
				return false
			}

			retPeer = lbPeer.Peer

			// sp could be nil, if we encountered a peer that is not registered for delivery, i.e. doesn't support the `stream` protocol
			// if sp is not nil, then we have selected the next peer and we stop iterating
			// if sp is nil, we continue iterating
			if retPeer != nil {
				selectedPeerPo = bin.ProximityOrder
				lbPeer.AddUseCount()

				return false
			}
		}

		return true
	})

	if osp != nil {
		osp.LogFields(olog.Int("selectedPeerPo", selectedPeerPo))
	}

	if err != nil {
		return nil, err
	}

	if retPeer == nil {
		return nil, ErrNoPeerFound
	}

	return retPeer, nil
}

// handleRetrieveRequest handles an incoming retrieve request from a certain Peer
// if the chunk is found in the localstore it is served immediately, otherwise
// it results in a new retrieve request to candidate peers in our kademlia
func (r *Retrieval) handleRetrieveRequest(ctx context.Context, p *Peer, msg *RetrieveRequest) error {
	p.logger.Debug("retrieval.handleRetrieveRequest", "ref", msg.Addr)
	handleRetrieveRequestMsgCount.Inc(1)

	ctx, osp := spancontext.StartSpan(
		ctx,
		"handle.retrieve.request")

	osp.LogFields(olog.String("ref", msg.Addr.String()))

	defer osp.Finish()

	ctx, cancel := context.WithTimeout(ctx, timeouts.FetcherGlobalTimeout)
	defer cancel()

	req := &storage.Request{
		Addr:   msg.Addr,
		Origin: p.ID(),
	}
	chunk, err := r.netStore.Get(ctx, chunk.ModeGetRequest, req)
	if err != nil {
		retrieveChunkFail.Inc(1)
		return fmt.Errorf("netstore.Get can not retrieve chunk for ref %s: %w", msg.Addr, err)
	}

	p.logger.Trace("retrieval.handleRetrieveRequest - delivery", "ref", msg.Addr)

	deliveryMsg := &ChunkDelivery{
		Ruid:  msg.Ruid,
		Addr:  chunk.Address(),
		SData: chunk.Data(),
	}

	err = p.Send(ctx, deliveryMsg)
	if err != nil {
		return fmt.Errorf("retrieval.handleRetrieveRequest - peer delivery for ref %s: %w", msg.Addr, err)
	}
	osp.LogFields(olog.Bool("delivered", true))

	return nil
}

// handleChunkDelivery handles a ChunkDelivery message from a certain peer
// if the chunk proximity order in relation to our base address is within depth
// we treat the chunk as a chunk received in syncing
func (r *Retrieval) handleChunkDelivery(ctx context.Context, p *Peer, msg *ChunkDelivery) error {
	p.logger.Debug("retrieval.handleChunkDelivery", "ref", msg.Addr)
	err := p.checkRequest(msg.Ruid, msg.Addr)
	if err != nil {
		unsolicitedChunkDelivery.Inc(1)
		return protocols.Break(fmt.Errorf("unsolicited chunk delivery from peer, ruid %d, addr %s: %w", msg.Ruid, msg.Addr, err))
	}
	var osp opentracing.Span
	ctx, osp = spancontext.StartSpan(
		ctx,
		"handle.chunk.delivery")

	processReceivedChunksCount.Inc(1)

	// count how many chunks we receive for retrieve requests per peer
	peermetric := fmt.Sprintf("network.retrieve.chunk.delivery.%x", p.BzzAddr.Over()[:16])
	metrics.GetOrRegisterCounter(peermetric, nil).Inc(1)

	peerPO := chunk.Proximity(p.BzzAddr.Over(), msg.Addr)
	po := chunk.Proximity(r.kad.BaseAddr(), msg.Addr)
	depth := r.kad.NeighbourhoodDepth()
	var mode chunk.ModePut
	// chunks within the area of responsibility should always sync
	// https://github.com/ethersphere/go-ethereum/pull/1282#discussion_r269406125
	if po >= depth || peerPO < po {
		mode = chunk.ModePutSync
	} else {
		// do not sync if peer that is sending us a chunk is closer to the chunk then we are
		mode = chunk.ModePutRequest
	}
	defer osp.Finish()

	_, err = r.netStore.Put(ctx, mode, storage.NewChunk(msg.Addr, msg.SData))
	if err != nil {
		if err == storage.ErrChunkInvalid {
			return protocols.Break(fmt.Errorf("netstore putting chunk to localstore: %w", err))
		}

		return fmt.Errorf("netstore putting chunk to localstore: %w", err)
	}

	return nil
}

// RequestFromPeers sends a chunk retrieve request to the next found peer.
// returns the next peer to try, a cleanup function to expire retrievals that were never delivered
func (r *Retrieval) RequestFromPeers(ctx context.Context, req *storage.Request, localID enode.ID) (*enode.ID, func(), error) {
	r.logger.Debug("retrieval.requestFromPeers", "req.Addr", req.Addr, "localID", localID)
	metrics.GetOrRegisterCounter("network.retrieve.request_from_peers", nil).Inc(1)

	const maxFindPeerRetries = 5
	retries := 0

FINDPEER:
	sp, err := r.findPeerLB(ctx, req)
	if err != nil {
		r.logger.Trace(err.Error())
		return nil, func() {}, err
	}

	protoPeer := r.getPeer(sp.ID())
	if protoPeer == nil {
		r.logger.Warn("findPeer returned a peer to skip", "peer", sp.String(), "retry", retries, "ref", req.Addr)
		req.PeersToSkip.Store(sp.ID().String(), time.Now())
		retries++
		if retries == maxFindPeerRetries {
			r.logger.Error("max find peer retries reached", "max retries", maxFindPeerRetries, "ref", req.Addr)
			return nil, func() {}, ErrNoPeerFound
		}

		goto FINDPEER
	}

	ret := &RetrieveRequest{
		Ruid: uint(rand.Uint32()),
		Addr: req.Addr,
	}
	protoPeer.logger.Trace("sending retrieve request", "ref", ret.Addr, "origin", localID, "ruid", ret.Ruid)
	protoPeer.addRetrieval(ret.Ruid, ret.Addr)
	cleanup := func() {
		protoPeer.expireRetrieval(ret.Ruid)
	}
	err = protoPeer.Send(ctx, ret)
	if err != nil {
		protoPeer.logger.Error("error sending retrieve request to peer", "ruid", ret.Ruid, "err", err)
		cleanup()
		return nil, func() {}, err
	}

	spID := protoPeer.ID()
	return &spID, cleanup, nil
}

func (r *Retrieval) Start(server *p2p.Server) error {
	r.logger.Info("starting bzz-retrieve")
	return nil
}

func (r *Retrieval) Stop() error {
	r.logger.Info("shutting down bzz-retrieve")
	close(r.quit)
	r.kademliaLB.Stop()
	return nil
}

func (r *Retrieval) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    r.spec.Name,
			Version: r.spec.Version,
			Length:  r.spec.Length(),
			Run:     r.runProtocol,
		},
	}
}

func (r *Retrieval) runProtocol(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, r.spec)
	bp := network.NewBzzPeer(peer)

	return r.Run(bp)
}

func (r *Retrieval) APIs() []rpc.API {
	return nil
}

func (r *Retrieval) Spec() *protocols.Spec {
	return r.spec
}
