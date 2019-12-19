// Copyright 2016 The go-ethereum Authors
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

package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/state"
)

const connectionsKey = "conns"
const addressesKey = "peers"

/*
Hive is the logistic manager of the swarm

When the hive is started, a forever loop is launched that
asks the  kademlia nodetable
to suggest peers to bootstrap connectivity
*/

// HiveParams holds the config options to hive
type HiveParams struct {
	Discovery             bool  // if want discovery of not
	DisableAutoConnect    bool  // this flag disables the auto connect loop
	PeersBroadcastSetSize uint8 // how many peers to use when relaying
	MaxPeersPerRequest    uint8 // max size for peer address batches
	KeepAliveInterval     time.Duration
}

// NewHiveParams returns hive config with only the
func NewHiveParams() *HiveParams {
	return &HiveParams{
		Discovery:             true,
		PeersBroadcastSetSize: 3,
		MaxPeersPerRequest:    5,
		KeepAliveInterval:     500 * time.Millisecond,
	}
}

// Hive manages network connections of the swarm node
type Hive struct {
	*HiveParams                   // settings
	*Kademlia                     // the overlay connectiviy driver
	Store       state.Store       // storage interface to save peers across sessions
	addPeer     func(*enode.Node) // server callback to connect to a peer
	// bookkeeping
	lock    sync.Mutex
	peers   map[enode.ID]*BzzPeer
	ticker  *time.Ticker
	done    chan struct{}
	started bool
}

// NewHive constructs a new hive
// HiveParams: config parameters
// Kademlia: connectivity driver using a network topology
// StateStore: to save peers across sessions
func NewHive(params *HiveParams, kad *Kademlia, store state.Store) *Hive {
	return &Hive{
		HiveParams: params,
		Kademlia:   kad,
		Store:      store,
		peers:      make(map[enode.ID]*BzzPeer),
	}
}

// Start stars the hive, receives p2p.Server only at startup
// server is used to connect to a peer based on its NodeID or enode URL
// these are called on the p2p.Server which runs on the node
func (h *Hive) Start(server *p2p.Server) error {
	return h.start(server, server.AddPeer)
}

// start stars the hive, receives p2p.Server only at startup
// server is used to connect to a peer based on its NodeID or enode URL
// these are called on the p2p.Server which runs on the node
func (h *Hive) start(server *p2p.Server, addPeerFunc func(*enode.Node)) error {
	log.Info("Starting hive", "baseaddr", fmt.Sprintf("%x", h.BaseAddr()[:4]))
	// assigns the p2p.Server#AddPeer function to connect to peers
	h.addPeer = addPeerFunc
	// if state store is specified, load peers to prepopulate the overlay address book
	if h.Store != nil {
		log.Info("Detected an existing store. trying to load peers")
		if err := h.loadPeers(); err != nil {
			log.Error(fmt.Sprintf("%08x hive encoutered an error trying to load peers", h.BaseAddr()[:4]))
			return err
		}
	}
	// ticker to keep the hive alive
	h.ticker = time.NewTicker(h.KeepAliveInterval)
	// done channel to signal the connect goroutine to return after Stop
	h.done = make(chan struct{})
	// this loop is doing bootstrapping and maintains a healthy table
	if !h.DisableAutoConnect {
		go h.connect()
	}
	h.started = true
	return nil
}

// Stop terminates the updateloop and saves the peers
func (h *Hive) Stop() error {
	log.Info(fmt.Sprintf("%08x hive stopping, saving peers", h.BaseAddr()[:4]))
	if !h.started {
		return nil
	}
	if h.ticker != nil {
		h.ticker.Stop()
	}
	close(h.done)
	if h.Store != nil {
		if err := h.savePeers(); err != nil {
			return fmt.Errorf("could not save peers to persistence store: %v", err)
		}
		if err := h.Store.Close(); err != nil {
			return fmt.Errorf("could not close file handle to persistence store: %v", err)
		}
	}
	log.Info(fmt.Sprintf("%08x hive stopped, dropping peers", h.BaseAddr()[:4]))
	h.EachConn(nil, 255, func(p *Peer, _ int) bool {
		p.Drop("hive stopping")
		return true
	})

	log.Info(fmt.Sprintf("%08x all peers dropped", h.BaseAddr()[:4]))

	h.started = false
	return nil
}

// connect is a forever loop
// at each iteration, ask the overlay driver to suggest the most preferred peer to connect to
// as well as advertises saturation depth if needed
func (h *Hive) connect() {
	for {
		select {
		case <-h.ticker.C:
			h.tickHive()
		case <-h.done:
			return
		}
	}
}

func (h *Hive) tickHive() {
	addr, depth, changed := h.SuggestPeer()
	if h.Discovery && changed {
		h.NotifyDepth(uint8(depth))
	}
	if addr != nil {
		log.Trace(fmt.Sprintf("%08x hive connect() suggested %08x", h.BaseAddr()[:4], addr.Address()[:4]))
		underA := addr.Under()
		s := string(underA)
		under, err := enode.ParseV4(s)
		if err != nil {
			log.Warn(fmt.Sprintf("%08x unable to connect to bee %08x: invalid node URL: %v", h.BaseAddr()[:4], addr.Address()[:4], err))
			return
		}
		log.Trace(fmt.Sprintf("%08x attempt to connect to bee %08x", h.BaseAddr()[:4], addr.Address()[:4]))
		h.addPeer(under)
	}
}

// Run protocol run function
func (h *Hive) Run(p *BzzPeer) error {
	h.trackPeer(p)
	defer h.untrackPeer(p)

	dp := NewPeer(p, h.Kademlia)
	depth, changed := h.On(dp)
	// if we want discovery, advertise change of depth
	if h.Discovery {
		if changed {
			// if depth changed, send to all peers
			h.NotifyDepth(depth)
		} else {
			// otherwise just send depth to new peer
			dp.NotifyDepth(depth)
		}
		h.NotifyPeer(p.BzzAddr)
	}
	defer h.Off(dp)
	return dp.Run(h.handleMsg(dp))
}

func (h *Hive) trackPeer(p *BzzPeer) {
	h.lock.Lock()
	h.peers[p.ID()] = p
	h.lock.Unlock()
}

func (h *Hive) untrackPeer(p *BzzPeer) {
	h.lock.Lock()
	delete(h.peers, p.ID())
	h.lock.Unlock()
}

// NodeInfo function is used by the p2p.server RPC interface to display
// protocol specific node information
func (h *Hive) NodeInfo() interface{} {
	return h.String()
}

// PeerInfo function is used by the p2p.server RPC interface to display
// protocol specific information any connected peer referred to by their NodeID
func (h *Hive) PeerInfo(id enode.ID) interface{} {
	p := h.Peer(id)

	if p == nil {
		return nil
	}
	// TODO this is bogus, the overlay address will not be correct
	addr := NewBzzAddrFromEnode(p.Node())
	return struct {
		OAddr hexutil.Bytes
		UAddr hexutil.Bytes
	}{
		OAddr: addr.OAddr,
		UAddr: addr.UAddr,
	}
}

// Peer returns a bzz peer from the Hive. If there is no peer
// with the provided enode id, a nil value is returned.
func (h *Hive) Peer(id enode.ID) *BzzPeer {
	h.lock.Lock()
	defer h.lock.Unlock()

	return h.peers[id]
}

// loadPeers, savePeer implement persistence callback/
func (h *Hive) loadPeers() error {
	var as []*BzzAddr
	err := h.Store.Get(addressesKey, &as)
	if err != nil {
		if err == state.ErrNotFound {
			log.Info(fmt.Sprintf("hive %08x: no persisted peers found", h.BaseAddr()[:4]))
			return nil
		}
		return err
	}
	// workaround for old node stores not containing capabilities
	for i := range as {
		if as[i].Capabilities == nil {
			caps := capability.NewCapabilities()
			caps.Add(fullCapability)
			as[i] = as[i].WithCapabilities(caps)
		}
	}
	log.Info(fmt.Sprintf("hive %08x: peers loaded", h.BaseAddr()[:4]))
	errRegistering := h.Register(as...)
	var conns []*BzzAddr
	err = h.Store.Get(connectionsKey, &conns)
	if err != nil {
		if err == state.ErrNotFound {
			log.Info(fmt.Sprintf("hive %08x: no persisted peer connections found", h.BaseAddr()[:4]))
		} else {
			log.Warn(fmt.Sprintf("hive %08x: error loading connections: %v", h.BaseAddr()[:4], err))
		}

	} else {
		go h.connectInitialPeers(conns)
	}
	return errRegistering
}

func (h *Hive) connectInitialPeers(conns []*BzzAddr) {
	log.Info(fmt.Sprintf("%08x hive connectInitialPeers() With %v saved connections", h.BaseAddr()[:4], len(conns)))
	for _, addr := range conns {
		log.Trace(fmt.Sprintf("%08x hive connect() suggested initial %08x", h.BaseAddr()[:4], addr.Address()[:4]))
		under, err := enode.ParseV4(string(addr.Under()))
		if err != nil {
			log.Warn(fmt.Sprintf("%08x unable to connect to bee %08x: invalid node URL: %v", h.BaseAddr()[:4], addr.Address()[:4], err))
			continue
		}
		log.Trace(fmt.Sprintf("%08x attempt to connect to bee %08x", h.BaseAddr()[:4], addr.Address()[:4]))
		h.addPeer(under)
	}
}

// savePeers, savePeer implement persistence callback/
func (h *Hive) savePeers() error {
	var peers []*BzzAddr
	var conns []*BzzAddr
	h.Kademlia.EachAddr(nil, 256, func(pa *BzzAddr, i int) bool {
		if pa == nil {
			log.Warn(fmt.Sprintf("empty addr: %v", i))
			return true
		}
		log.Trace("saving peer", "peer", pa)
		peers = append(peers, pa)
		return true
	})

	h.Kademlia.EachConn(nil, 256, func(p *Peer, i int) bool {
		log.Trace("saving connected peer", "OAddr", hexutil.Encode(p.OAddr), "UAddr", p.UAddr)
		conns = append(conns, p.BzzAddr)
		return true
	})
	if err := h.Store.Put(addressesKey, peers); err != nil {
		return fmt.Errorf("could not save peers: %v", err)
	}

	if err := h.Store.Put(connectionsKey, conns); err != nil {
		return fmt.Errorf("could not save peer connections: %v", err)
	}
	return nil
}

var sortPeers = noSortPeers

// handleMsg is the message handler that delegates incoming messages
func (h *Hive) handleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		switch msg := msg.(type) {
		case *peersMsg:
			return h.handlePeersMsg(p, msg)
		case *subPeersMsg:
			return h.handleSubPeersMsg(ctx, p, msg)
		}

		return fmt.Errorf("unknown message type: %T", msg)
	}
}

// NotifyDepth sends a message to all connections if depth of saturation is changed
func (h *Hive) NotifyDepth(depth uint8) {
	f := func(val *Peer, po int) bool {
		val.NotifyDepth(depth)
		return true
	}
	h.EachConn(nil, 255, f)
}

// NotifyPeer informs all peers about a newly added node
func (h *Hive) NotifyPeer(p *BzzAddr) {
	f := func(val *Peer, po int) bool {
		val.NotifyPeer(p, uint8(po))
		return true
	}
	h.EachConn(p.Address(), 255, f)
}

// handlePeersMsg called by the protocol when receiving peerset (for target address)
// list of nodes ([]PeerAddr in peersMsg) is added to the overlay db using the
// Register interface method
func (h *Hive) handlePeersMsg(d *Peer, msg *peersMsg) error {
	// register all addresses
	if len(msg.Peers) == 0 {
		return nil
	}
	for _, a := range msg.Peers {
		d.seen(a)
		h.NotifyPeer(a)
	}
	return h.Register(msg.Peers...)
}

// handleSubPeersMsg handles incoming subPeersMsg
// this message represents the saturation depth of the remote peer
// saturation depth is the radius within which the peer subscribes to peers
// the first time this is received we send peer info on all
// our connected peers that fall within peers saturation depth
// otherwise this depth is just recorded on the peer, so that
// subsequent new connections are sent iff they fall within the radius
func (h *Hive) handleSubPeersMsg(ctx context.Context, d *Peer, msg *subPeersMsg) error {
	d.setDepth(msg.Depth)
	// only send peers after the initial subPeersMsg
	h.lock.Lock()
	if d.sentPeers {
		h.lock.Unlock()
		return nil
	}

	d.sentPeers = true
	h.lock.Unlock()

	var peers []*BzzAddr
	// iterate connection in ascending order of distance from the remote address
	h.EachConn(d.Over(), 255, func(p *Peer, po int) bool {
		// terminate if we are beyond the radius
		if uint8(po) < msg.Depth {
			return false
		}
		if !d.seen(p.BzzAddr) { // here just records the peer sent
			peers = append(peers, p.BzzAddr)
		}
		return true
	})
	// if useful  peers are found, send them over
	if len(peers) > 0 {
		go d.Send(ctx, &peersMsg{Peers: sortPeers(peers)})
	}
	return nil
}
