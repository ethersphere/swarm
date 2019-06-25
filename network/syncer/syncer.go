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

package syncer

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
)

// SwarmSyncer implements node.Service
var _ node.Service = (*SwarmSyncer)(nil)
var (
	pollTime           = 1 * time.Second
	createStreamsDelay = 50 * time.Millisecond //to avoid a race condition where we send a message to a server that hasnt set up yet
)

var SyncerSpec = &protocols.Spec{
	Name:       "bzz-sync",
	Version:    8,
	MaxMsgSize: 10 * 1024 * 1024,
	Messages: []interface{}{
		StreamInfoReq{},
		StreamInfoRes{},
		GetRange{},
		OfferedHashes{},
		WantedHashes{},
	},
}

// SwarmSyncer is the base type that handles all client/server operations on a node
// it is instantiated once per stream protocol instance, that is, it should have
// one instance per node
type SwarmSyncer struct {
	mtx            sync.RWMutex
	intervalsStore state.Store //every protocol would make use of this
	peers          map[enode.ID]*Peer
	netStore       *storage.NetStore
	kad            *network.Kademlia
	started        bool

	spec    *protocols.Spec   //this protocol's spec
	balance protocols.Balance //implements protocols.Balance, for accounting
	prices  protocols.Prices  //implements protocols.Prices, provides prices to accounting

	quit chan struct{} // terminates registry goroutines
}

func NewSwarmSyncer(me enode.ID, intervalsStore state.Store, kad *network.Kademlia, ns *storage.NetStore) *SwarmSyncer {
	syncer := &SwarmSyncer{
		intervalsStore: intervalsStore,
		peers:          make(map[enode.ID]*Peer),
		kad:            kad,
		netStore:       ns,
		quit:           make(chan struct{}),
	}

	syncer.spec = SyncerSpec

	return syncer
}

func (s *SwarmSyncer) addPeer(p *Peer) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.peers[p.ID()] = p
}

func (s *SwarmSyncer) removePeer(p *Peer) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, found := s.peers[p.ID()]; found {
		delete(s.peers, p.ID())
		p.Left()

	} else {
		log.Warn("peer was marked for removal but not found")
		panic("shouldnt happen")
	}
}

// Run is being dispatched when 2 nodes connect
func (s *SwarmSyncer) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, s.spec)
	bp := network.NewBzzPeer(peer)

	np := network.NewPeer(bp, s.kad)
	s.kad.On(np)
	defer s.kad.Off(np)

	sp := NewPeer(bp, s)
	s.addPeer(sp)
	defer s.removePeer(sp)
	go s.CreateStreams(sp)
	return peer.Run(sp.HandleMsg)
}

// CreateStreams creates and maintains the streams per peer.
// Runs per peer, in a separate goroutine
// when the depth changes on our node
//  - peer moves from out-of-depth to depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - peer moves from depth to out-of-depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - depth changes, and peer stays in depth, but we need MORE (or LESS) streams (WHY???).. so again -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
// peer connects and disconnects quickly
func (s *SwarmSyncer) CreateStreams(p *Peer) {
	peerPo := chunk.Proximity(s.kad.BaseAddr(), p.BzzAddr.Address())
	depth := s.kad.NeighbourhoodDepth()
	withinDepth := peerPo >= depth

	log.Debug("create streams", "peer", p.BzzAddr, "base", s.kad.BaseAddr(), "withinDepth", withinDepth, "depth", depth, "po", peerPo)

	if withinDepth {
		sub, _ := syncSubscriptionsDiff(peerPo, -1, depth, s.kad.MaxProxDisplay, true)

		streamsMsg := StreamInfoReq{Streams: sub}
		log.Debug("sending subscriptions message", "bins", sub)
		time.Sleep(createStreamsDelay)
		if err := p.Send(context.TODO(), streamsMsg); err != nil {
			log.Error("err establishing initial subscription", "err", err)
		}
	}

	subscription, unsubscribe := s.kad.SubscribeToNeighbourhoodDepthChange()
	defer unsubscribe()
	for {
		select {
		case <-subscription:
			switch newDepth := s.kad.NeighbourhoodDepth(); {
			case newDepth == depth:
				continue
			case peerPo >= newDepth:
				// peer is within depth
				if !withinDepth {
					log.Debug("peer moved into depth, requesting cursors")

					withinDepth = peerPo >= newDepth
					// previous depth is -1 because we did not have any streams with the client beforehand
					sub, _ := syncSubscriptionsDiff(peerPo, -1, newDepth, s.kad.MaxProxDisplay, true)
					streamsMsg := StreamInfoReq{Streams: sub}
					if err := p.Send(context.TODO(), streamsMsg); err != nil {
						log.Error("error establishing subsequent subscription", "err", err)
						p.Drop()
					}
					depth = newDepth
				} else {
					// peer was within depth, but depth has changed. we should request the cursors for the
					// necessary bins and quit the unnecessary ones
					depth = newDepth
				}
			case peerPo < newDepth:
				if withinDepth {
					log.Debug("peer transitioned out of depth, removing cursors")
					for k, _ := range p.streamCursors {
						delete(p.streamCursors, k)

						if hs, ok := p.historicalStreams[k]; ok {
							close(hs.quit)
							// todo: wait for the hs.done to close?
							delete(p.historicalStreams, k)
						} else {
							// this could happen when the cursor was 0 thus the historical stream was not created - do nothing
						}
					}
					withinDepth = false
				}
			}

		case <-s.quit:
			return
		}
	}
}

func (s *SwarmSyncer) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "bzz-sync",
			Version: 1,
			Length:  10 * 1024 * 1024,
			Run:     s.Run,
		},
	}
}

func (r *SwarmSyncer) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "bzz-sync",
			Version:   "1.0",
			Service:   NewAPI(r),
			Public:    false,
		},
	}
}

// Additional public methods accessible through API for pss
type API struct {
	*SwarmSyncer
}

func NewAPI(s *SwarmSyncer) *API {
	return &API{SwarmSyncer: s}
}

func (s *SwarmSyncer) Start(server *p2p.Server) error {
	log.Info("syncer starting")
	return nil
}

func (s *SwarmSyncer) Stop() error {
	log.Info("syncer shutting down")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	close(s.quit)
	return nil
}

func (s *SwarmSyncer) GetBinsForPeer(p *Peer) (bins []uint, depth int) {
	peerPo := chunk.Proximity(s.kad.BaseAddr(), p.BzzAddr.Address())
	depth = s.kad.NeighbourhoodDepth()
	sub, _ := syncSubscriptionsDiff(peerPo, -1, depth, s.kad.MaxProxDisplay, true)
	return sub, depth
}
