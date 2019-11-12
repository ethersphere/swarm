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

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/pot"
)

// Peer wraps BzzPeer and embeds Kademlia overlay connectivity driver
type Peer struct {
	*BzzPeer
	kad       *Kademlia
	sentPeers bool            // whether we already sent peer closer to this address
	mtx       sync.RWMutex    // protect peers map
	peers     map[string]bool // tracks node records sent to the peer
	depth     uint8           // the proximity order advertised by remote as depth of saturation
	key       string          // peer key. Hex form of Address()
}

// NewPeer constructs a discovery peer
func NewPeer(p *BzzPeer, kad *Kademlia) *Peer {
	d := &Peer{
		kad:     kad,
		BzzPeer: p,
		peers:   make(map[string]bool),
		key:     hexutil.Encode(p.Address()),
	}
	// record remote as seen so we never send a peer its own record
	d.seen(p.BzzAddr)
	return d
}

// Key returns a string representation of this peer to be used in maps.
func (d *Peer) Key() string {
	return d.key
}

// Label returns a short string representation for debugging purposes
func (d *Peer) Label() string {
	return d.key[:4]
}

// NotifyPeer notifies the remote node (recipient) about a peer if
// the peer's PO is within the recipients advertised depth
// OR the peer is closer to the recipient than self
// unless already notified during the connection session
func (d *Peer) NotifyPeer(a *BzzAddr, po uint8) {
	// immediately return
	if (po < d.getDepth() && pot.ProxCmp(d.kad.BaseAddr(), d.Address(), a.Address()) != 1) || d.seen(a) {
		return
	}
	resp := &peersMsg{
		Peers: []*BzzAddr{a},
	}
	go d.Send(context.TODO(), resp)
}

// NotifyDepth sends a subPeers Msg to the receiver notifying them about
// a change in the depth of saturation
func (d *Peer) NotifyDepth(po uint8) {
	go d.Send(context.TODO(), &subPeersMsg{Depth: po})
}

/*
peersMsg is the message to pass peer information
It is always a response to a peersRequestMsg

The encoding of a peer address is identical the devp2p base protocol peers
messages: [IP, Port, NodeID],
Note that a node's FileStore address is not the NodeID but the hash of the NodeID.

TODO:
To mitigate against spurious peers messages, requests should be remembered
and correctness of responses should be checked

If the proxBin of peers in the response is incorrect the sender should be
disconnected
*/

// peersMsg encapsulates an array of peer addresses
// used for communicating about known peers
// relevant for bootstrapping connectivity and updating peersets
type peersMsg struct {
	Peers []*BzzAddr
}

// DecodeRLP implements rlp.Decoder interface
func (p *peersMsg) DecodeRLP(s *rlp.Stream) error {
	_, err := s.List()
	if err != nil {
		return err
	}
	_, err = s.List()
	if err != nil {
		return err
	}
	for {
		var addr BzzAddr
		err = s.Decode(&addr)
		if err != nil {
			break
		}
		p.Peers = append(p.Peers, &addr)
	}
	return nil
}

// String pretty prints a peersMsg
func (msg peersMsg) String() string {
	return fmt.Sprintf("%T: %v", msg, msg.Peers)
}

// subPeers msg is communicating the depth of the overlay table of a peer
type subPeersMsg struct {
	Depth uint8
}

// String returns the pretty printer
func (msg subPeersMsg) String() string {
	return fmt.Sprintf("%T: request peers > PO%02d. ", msg, msg.Depth)
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
