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
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/state"
)

/*
Hive is the logistic manager of the swarm

When the hive is started, a forever loop is launched that
asks the  kademlia nodetable
to suggest peers to bootstrap connectivity
*/

// HiveParams holds the config options to hive
type HiveParams struct {
	Discovery             bool  // if want discovery of not
	PeersBroadcastSetSize uint8 // how many peers to use when relaying
	MaxPeersPerRequest    uint8 // max size for peer address batches
	KeepAliveInterval     time.Duration
	RetryInterval         int64 // initial interval before a peer is first redialed
	RetryExponent         int   // exponent to multiply retry intervals with
}

// NewHiveParams returns hive config with only the
func NewHiveParams() *HiveParams {
	return &HiveParams{
		Discovery:             true,
		PeersBroadcastSetSize: 3,
		MaxPeersPerRequest:    5,
		KeepAliveInterval:     500 * time.Millisecond,
		RetryInterval:         4200000000, // 4.2 sec
		//RetryExponent:         2,
		RetryExponent: 3,
	}
}

// Hive manages network connections of the swarm node
type Hive struct {
	*HiveParams                   // settings
	*Kademlia                     // the overlay connectiviy driver
	Store       state.Store       // storage interface to save peers across sessions
	addPeer     func(*enode.Node) // server callback to connect to a peer
	// bookkeeping
	lock   sync.Mutex
	peers  map[enode.ID]*BzzPeer
	ticker *time.Ticker
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
	log.Info("Starting hive", "baseaddr", fmt.Sprintf("%x", h.BaseAddr()[:4]))
	// if state store is specified, load peers to prepopulate the overlay address book
	if h.Store != nil {
		log.Info("Detected an existing store. trying to load peers")
		if err := h.loadPeers(); err != nil {
			log.Error(fmt.Sprintf("%08x hive encoutered an error trying to load peers", h.BaseAddr()[:4]))
			return err
		}
	}
	// assigns the p2p.Server#AddPeer function to connect to peers
	h.addPeer = server.AddPeer
	// ticker to keep the hive alive
	h.ticker = time.NewTicker(h.KeepAliveInterval)
	// this loop is doing bootstrapping and maintains a healthy table
	go h.connect()
	return nil
}

// Stop terminates the updateloop and saves the peers
func (h *Hive) Stop() error {
	log.Info(fmt.Sprintf("%08x hive stopping, saving peers", h.BaseAddr()[:4]))
	h.ticker.Stop()
	if h.Store != nil {
		if err := h.savePeers(); err != nil {
			return fmt.Errorf("could not save peers to persistence store: %v", err)
		}
		if err := h.Store.Close(); err != nil {
			return fmt.Errorf("could not close file handle to persistence store: %v", err)
		}
	}
	log.Info(fmt.Sprintf("%08x hive stopped, dropping peers", h.BaseAddr()[:4]))
	h.EachConn(nil, 255, func(p *Peer, _ int, _ bool) bool {
		log.Info(fmt.Sprintf("%08x dropping peer %08x", h.BaseAddr()[:4], p.Address()[:4]))
		p.Drop(nil)
		return true
	})

	log.Info(fmt.Sprintf("%08x all peers dropped", h.BaseAddr()[:4]))
	return nil
}

// connect is a forever loop
// at each iteration, ask the overlay driver to suggest the most preferred peer to connect to
// as well as advertises saturation depth if needed
func (h *Hive) connect() {
	for range h.ticker.C {

		addr, depth, changed := h.SuggestPeer()
		if h.Discovery && changed {
			NotifyDepth(uint8(depth), h.Kademlia)
		}
		if addr == nil {
			continue
		}

		log.Trace(fmt.Sprintf("%08x hive connect() suggested %08x", h.BaseAddr()[:4], addr.Address()[:4]))
		under, err := enode.ParseV4(string(addr.Under()))
		if err != nil {
			log.Warn(fmt.Sprintf("%08x unable to connect to bee %08x: invalid node URL: %v", h.BaseAddr()[:4], addr.Address()[:4], err))
			continue
		}
		log.Trace(fmt.Sprintf("%08x attempt to connect to bee %08x", h.BaseAddr()[:4], addr.Address()[:4]))
		h.addPeer(under)
	}
}

// Run protocol run function
func (h *Hive) Run(p *BzzPeer) error {
	dp := NewPeer(p, h.Kademlia)
	depth, changed := h.On(dp)
	// if we want discovery, advertise change of depth
	if h.Discovery {
		if changed {
			// if depth changed, send to all peers
			NotifyDepth(uint8(depth), h.Kademlia)
		} else {
			// otherwise just send depth to new peer
			dp.NotifyDepth(uint8(depth))
		}
		NotifyPeer(p.BzzAddr, h.Kademlia)
	}
	defer h.Off(dp)
	return dp.Run(dp.HandleMsg)
}

// NodeInfo function is used by the p2p.server RPC interface to display
// protocol specific node information
func (h *Hive) NodeInfo() interface{} {
	return h.String()
}

// PeerInfo function is used by the p2p.server RPC interface to display
// protocol specific information any connected peer referred to by their NodeID
func (h *Hive) PeerInfo(id enode.ID) interface{} {
	h.lock.Lock()
	p := h.peers[id]
	h.lock.Unlock()

	if p == nil {
		return nil
	}
	addr := NewAddr(p.Node())
	return struct {
		OAddr hexutil.Bytes
		UAddr hexutil.Bytes
	}{
		OAddr: addr.OAddr,
		UAddr: addr.UAddr,
	}
}

// loadPeers, savePeer implement persistence callback/
func (h *Hive) loadPeers() error {
	var as []*BzzAddr
	err := h.Store.Get("peers", &as)
	if err != nil {
		if err == state.ErrNotFound {
			log.Info(fmt.Sprintf("hive %08x: no persisted peers found", h.BaseAddr()[:4]))
			return nil
		}
		return err
	}
	log.Info(fmt.Sprintf("hive %08x: peers loaded", h.BaseAddr()[:4]))

	return h.Register(as...)
}

// savePeers, savePeer implement persistence callback/
func (h *Hive) savePeers() error {
	var peers []*BzzAddr
	h.Kademlia.EachAddr(nil, 256, func(pa *BzzAddr, i int, _ bool) bool {
		if pa == nil {
			log.Warn(fmt.Sprintf("empty addr: %v", i))
			return true
		}
		log.Trace("saving peer", "peer", pa)
		peers = append(peers, pa)
		return true
	})
	if err := h.Store.Put("peers", peers); err != nil {
		return fmt.Errorf("could not save peers: %v", err)
	}
	return nil
}

// SuggestPeer returns a known peer for the lowest proximity bin for the
// lowest bincount below depth
// naturally if there is an empty row it returns a peer for that
func (h *Hive) SuggestPeer() (a *BzzAddr, o int, want bool) {
	return &BzzAddr{}, 0, false
	//	k.lock.Lock()
	//	defer k.lock.Unlock()
	//	minsize := k.MinBinSize
	//	depth := depthForPot(k.conns, k.MinProxBinSize, k.base)
	//	// if there is a callable neighbour within the current proxBin, connect
	//	// this makes sure nearest neighbour set is fully connected
	//	var ppo int
	//	k.addrs.EachNeighbour(k.base, Pof, func(val pot.Val, po int) bool {
	//		if po < depth {
	//			return false
	//		}
	//		e := val.(*entry)
	//		c := k.callable(e)
	//		if c {
	//			a = e.BzzAddr
	//		}
	//		ppo = po
	//		return !c
	//	})
	//	if a != nil {
	//		log.Trace(fmt.Sprintf("%08x candidate nearest neighbour found: %v (%v)", k.BaseAddr()[:4], a, ppo))
	//		return a, 0, false
	//	}
	//
	//	var bpo []int
	//	prev := -1
	//	k.conns.EachBin(k.base, Pof, 0, func(po, size int, f func(func(val pot.Val, i int) bool) bool) bool {
	//		prev++
	//		for ; prev < po; prev++ {
	//			bpo = append(bpo, prev)
	//			minsize = 0
	//		}
	//		if size < minsize {
	//			bpo = append(bpo, po)
	//			minsize = size
	//		}
	//		return size > 0 && po < depth
	//	})
	//	// all buckets are full, ie., minsize == k.MinBinSize
	//	if len(bpo) == 0 {
	//		return nil, 0, false
	//	}
	//	// as long as we got candidate peers to connect to
	//	// dont ask for new peers (want = false)
	//	// try to select a candidate peer
	//	// find the first callable peer
	//	nxt := bpo[0]
	//	k.addrs.EachBin(k.base, Pof, nxt, func(po, _ int, f func(func(pot.Val, int) bool) bool) bool {
	//		// for each bin (up until depth) we find callable candidate peers
	//		if po >= depth {
	//			return false
	//		}
	//		return f(func(val pot.Val, _ int) bool {
	//			e := val.(*entry)
	//			c := k.callable(e)
	//			if c {
	//				a = e.BzzAddr
	//			}
	//			return !c
	//		})
	//	})
	//	// found a candidate
	//	if a != nil {
	//		return a, 0, false
	//	}
	//	// no candidate peer found, request for the short bin
	//	var changed bool
	//	if nxt < k.depth {
	//		k.depth = nxt
	//		changed = true
	//	}
	//	return a, nxt, changed
}

// calculate the allowed number of retries based on time lapsed since last seen
// NOTE this method has been moved from kademlia.go connect method. The function now spread over two functions
// TODO simplify
func (h *Hive) getRetriesFromDuration(timeAgo time.Duration) int {
	div := int64(h.RetryExponent)
	div += (150000 - rand.Int63n(300000)) * div / 1000000
	var retries int
	for delta := int64(timeAgo); delta > h.RetryInterval; delta /= div {
		retries++
	}
	return retries
}

func (h *Hive) isTimeForRetry(d *Peer) bool {
	timeAgo := time.Since(d.seenAt)
	allowedRetryCountNow := h.getRetriesFromDuration(timeAgo)
	isTime := d.retries < allowedRetryCountNow
	if isTime {
		log.Trace(fmt.Sprintf("%08x: peer %v is callable", Label(d)[:4], d))
	} else {
		log.Trace(fmt.Sprintf("%08x: %v long time since last try (at %v) needed before retry %v, wait only warrants %v", h.BaseAddr()[:4], d, timeAgo, d.retries, allowedRetryCountNow))
	}
	return isTime
}

// callable decides if an address entry represents a callable peer
// this is never called concurrently, so safe to increment
func (h *Hive) callable(d *Peer) bool {
	if d.up || !h.isTimeForRetry(d) {
		return false
	}
	// TODO move to the actual retry call
	d.retries++

	return true
}
