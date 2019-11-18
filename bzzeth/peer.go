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
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/p2p/protocols"
)

// Peer extends p2p/protocols Peer and represents a conrete protocol connection
type Peer struct {
	*protocols.Peer            // embeds protocols.Peer
	serveHeaders    bool       // if the remote serves headers
	requests        *requests  // per-peer pool of open requests
	logger          log.Logger // custom logger for peer
}

// NewPeer is the constructor for Peer
func NewPeer(peer *protocols.Peer) *Peer {
	return &Peer{
		Peer:     peer,
		requests: newRequests(),
		logger:   log.New("peer", peer.ID()),
	}
}

// peers represents the bzzeth specific peer pool
type peers struct {
	mtx   sync.RWMutex
	peers map[enode.ID]*Peer
}

func newPeers() *peers {
	return &peers{peers: make(map[enode.ID]*Peer)}
}

func (p *peers) get(id enode.ID) *Peer {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return p.peers[id]
}

func (p *peers) add(peer *Peer) {
	p.mtx.Lock()
	p.peers[peer.ID()] = peer
	p.mtx.Unlock()
}

func (p *peers) remove(peer *Peer) {
	p.mtx.Lock()
	delete(p.peers, peer.ID())
	p.mtx.Unlock()
}

// getEthPeer finds a peer that serves headers and calls the function argument on this peer
func (p *peers) getEth() *Peer {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	for _, peer := range p.peers {
		if peer.serveHeaders {
			return peer
		}
	}
	return nil
}

// requests represents the peer specific pool of open requests
type requests struct {
	mtx sync.RWMutex        // for concurrent access to requests
	r   map[uint32]*request // requests open for peer
}

type request struct {
	hashes map[string]bool // remembers the block hashes that are requested in this connection
	lock   sync.RWMutex    // mutex to update the hashes received status
	c      chan []byte     // channel in which the received block headers are passed on
	cancel func()          // function to call in case of cancellation of the GetBlockHeaders event
}

// newRequestIDFunc is used to generate unique ID for requests
// tests can reassign for deterministic ids
var newRequestIDFunc = newRequestID

// newRequestID generates a 32-bit random number to be used as unique id for s
// no reuse of id across peers
func newRequestID() uint32 {
	return rand.Uint32()
}

func newRequests() *requests {
	return &requests{
		r: make(map[uint32]*request),
	}
}

// create constructs a new request
// registers it on the peer request pool
// request.cancel() should be called to cleanup
func (r *requests) create(hashes [][]byte, c chan []byte) (*request, uint32) {
	req := &request{
		hashes: make(map[string]bool),
		c:      c,
	}
	id := newRequestIDFunc()
	for _, h := range hashes {
		req.hashes[hex.EncodeToString(h)] = false
	}
	req.cancel = func() { r.remove(id) }
	r.add(id, req)
	return req, id
}

func (r *requests) add(id uint32, req *request) {
	r.mtx.Lock()
	r.r[id] = req
	r.mtx.Unlock()
}

func (r *requests) remove(id uint32) {
	r.mtx.Lock()
	delete(r.r, id)
	r.mtx.Unlock()
}

func (r *requests) get(id uint32) (*request, bool) {
	r.mtx.RLock()
	req, ok := r.r[id]
	r.mtx.RUnlock()
	return req, ok
}

// getBlockHeaders sends a GetBlockHeaders message to the remote peer requesting headers by their _hashes_
// and delivers the actual block header responses to the deliveries channel
func (p *Peer) getBlockHeaders(ctx context.Context, hashes [][]byte, deliveries chan []byte) (*request, error) {
	req, id := p.requests.create(hashes, deliveries)
	err := p.Send(ctx, &GetBlockHeaders{
		Rid:    id,
		Hashes: hashes,
	})
	if err != nil {
		req.cancel()
		return nil, err
	}
	return req, nil
}

// this function is called to check if the remote peer is another swarm node
// in which case the protocol is idle
// can be reassigned in test to mock a swarm node
var isSwarmNodeFunc = isSwarmNode

func isSwarmNode(p *Peer) bool {
	return p.HasCap("bzz")
}
