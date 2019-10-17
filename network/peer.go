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
	"sync"
)

// Peer wraps BzzPeer and embeds Kademlia overlay connectivity driver
type Peer struct {
	*BzzPeer
	sentPeers bool            // whether we already sent peer closer to this address
	mtx       sync.RWMutex    // protect peers map
	peers     map[string]bool // tracks node records sent to the peer
	depth     uint8           // the proximity order advertised by remote as depth of saturation
}

// NewPeer constructs a discovery peer
func NewPeer(p *BzzPeer) *Peer {
	d := &Peer{
		BzzPeer: p,
		peers:   make(map[string]bool),
	}
	// record remote as seen so we never send a peer its own record
	d.seen(p.BzzAddr)
	return d
}

// NotifyDepth sends a subPeers Msg to the receiver notifying them about
// a change in the depth of saturation
func (d *Peer) NotifyDepth(po uint8) {
	go d.Send(context.TODO(), &subPeersMsg{Depth: po})
}

// seen takes a peer address and checks if it was sent to a peer already
// if not, marks the peer as sent
func (d *Peer) seen(p *BzzAddr) bool {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	k := string(p.Address())
	if d.peers[k] {
		return true
	}
	d.peers[k] = true
	return false
}

func (d *Peer) getDepth() uint8 {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return d.depth
}

func (d *Peer) setDepth(depth uint8) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.depth = depth
}

func noSortPeers(peers []*BzzAddr) []*BzzAddr {
	return peers
}
