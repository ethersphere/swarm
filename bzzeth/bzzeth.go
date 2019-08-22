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

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/storage"
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
	log.Warn("handshake", "hs", handshake, "peer", bp)
	// with another swarm node the protocol goes into idle
	if isSwarmNodeFunc(bp) {
		<-b.quit
		return nil
	}
	b.peers.add(bp)
	defer b.peers.remove(bp)

	return peer.Run(b.handleMsg(bp))
}

// handleMsg is the message handler that delegates incoming messages
// handlers are called asynchronously so handler calls do not block incoming msg processing
func (b *BzzEth) handleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		p.logger.Debug("bzzeth.handleMsg")
		switch msg.(type) {
		default:
			log.Info("Received a message ")
		}
		return nil
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
