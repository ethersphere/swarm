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
package network

import (
	"bytes"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/pubsubchannel"
	"github.com/ethersphere/swarm/network/resourceusestats"
)

// KademliaBackend is the required interface of KademliaLoadBalancer.
type KademliaBackend interface {
	SubscribeToPeerChanges() *pubsubchannel.Subscription
	BaseAddr() []byte
	EachBinDesc(base []byte, minProximityOrder int, consumer PeerBinConsumer)
	EachBinDescFiltered(base []byte, capKey string, minProximityOrder int, consumer PeerBinConsumer) error
	EachConn(base []byte, o int, f func(*Peer, int) bool)
}

// Creates a new KademliaLoadBalancer from a KademliaBackend.
// If useNearestNeighbourInit is true the nearest neighbour peer use count will be used when a peer is initialized.
// If not, least used peer use count in same bin as new peer will be used. It is not clear which one is better, when
// this load balancer would be used in several use cases we could do take some decision.
func NewKademliaLoadBalancer(kademlia KademliaBackend, useNearestNeighbourInit bool) *KademliaLoadBalancer {
	onOffPeerSub := kademlia.SubscribeToPeerChanges()
	quitC := make(chan struct{})
	klb := &KademliaLoadBalancer{
		kademlia:         kademlia,
		resourceUseStats: resourceusestats.NewResourceUseStats(quitC),
		onOffPeerSub:     onOffPeerSub,
		quitC:            quitC,
	}
	if useNearestNeighbourInit {
		klb.initCountFunc = klb.nearestNeighbourUseCount
	} else {
		klb.initCountFunc = klb.leastUsedCountInBin
	}

	go klb.listenOnOffPeers()
	return klb
}

// Consumer functions. A consumer is a function that uses an element returned by an iterator. It usually also returns
// a boolean signaling if it wants to iterate more or not. We created an alias for consumer function (LBBinConsumer)
// for code clarity.

// An LBPeer represents a peer with a AddUseCount() function to signal that the peer has been used in order
// to account it for LB sorting criteria.
type LBPeer struct {
	Peer  *Peer
	stats *resourceusestats.ResourceUseStats
}

// AddUseCount is called to account a use for these peer. Should be called if the peer is actually used.
func (lbPeer *LBPeer) AddUseCount() {
	lbPeer.stats.AddUse(lbPeer.Peer)
}

// LBBin represents a Bin of LBPeer's
type LBBin struct {
	LBPeers        []LBPeer
	ProximityOrder int
}

// LBBinConsumer will be provided with a list of LBPeer's in LB criteria ordering (currently in least used ordering).
// Should return true if it must continue iterating LBBin's or stops if false.
type LBBinConsumer func(bin LBBin) bool

// KademliaLoadBalancer tries to balance request to the peers in Kademlia returning the peers sorted
// by least recent used whenever several will be returned with the same po to a particular address.
// The user of KademliaLoadBalancer should signal if the returned element (LBPeer) has been used with the
// function lbPeer.AddUseCount()
type KademliaLoadBalancer struct {
	kademlia         KademliaBackend                    // kademlia to obtain bins of peers
	resourceUseStats *resourceusestats.ResourceUseStats // a resourceUseStats to count uses
	onOffPeerSub     *pubsubchannel.Subscription        // a pubsub channel to be notified of on/off peers in kademlia
	quitC            chan struct{}

	initCountFunc func(peer *Peer, po int) int //Function to use for initializing a new peer count
}

// Stop unsubscribe from notifiers
func (klb *KademliaLoadBalancer) Stop() {
	klb.onOffPeerSub.Unsubscribe()
	close(klb.quitC)
}

// EachBinNodeAddress calls EachBinDesc with the base address of kademlia (the node address)
func (klb *KademliaLoadBalancer) EachBinNodeAddress(consumeBin LBBinConsumer) {
	klb.EachBinDesc(klb.kademlia.BaseAddr(), consumeBin)
}

// EachBinFiltered returns all bins in descending order from the perspective of base address.
// Only peers with the provided capabilities capKey are considered.
// All peers in that bin will be provided to the LBBinConsumer sorted by least used first.
func (klb *KademliaLoadBalancer) EachBinFiltered(base []byte, capKey string, consumeBin LBBinConsumer) error {
	return klb.kademlia.EachBinDescFiltered(base, capKey, 0, func(peerBin *PeerBin) bool {
		peers := klb.peerBinToPeerList(peerBin)
		return consumeBin(LBBin{LBPeers: peers, ProximityOrder: peerBin.ProximityOrder})
	})
}

// EachBinDesc returns all bins in descending order from the perspective of base address.
// All peers in that bin will be provided to the LBBinConsumer sorted by least used first.
func (klb *KademliaLoadBalancer) EachBinDesc(base []byte, consumeBin LBBinConsumer) {
	klb.kademlia.EachBinDesc(base, 0, func(peerBin *PeerBin) bool {
		peers := klb.peerBinToPeerList(peerBin)
		return consumeBin(LBBin{LBPeers: peers, ProximityOrder: peerBin.ProximityOrder})
	})
}

func (klb *KademliaLoadBalancer) peerBinToPeerList(bin *PeerBin) []LBPeer {
	resources := make([]resourceusestats.Resource, bin.Size)
	var i int
	bin.PeerIterator(func(entry *entry) bool {
		resources[i] = entry.conn
		i++
		return true
	})
	return klb.resourcesToLbPeers(resources)
}

func (klb *KademliaLoadBalancer) resourcesToLbPeers(resources []resourceusestats.Resource) []LBPeer {
	sorted := klb.resourceUseStats.SortResources(resources)
	peers := klb.toLBPeers(sorted)
	return peers
}

func (klb *KademliaLoadBalancer) listenOnOffPeers() {
	for {
		select {
		case <-klb.quitC:
			return
		case msg, ok := <-klb.onOffPeerSub.ReceiveChannel():
			if !ok {
				log.Debug("listenOnOffPeers closed channel, finishing subscriber to on/off peers")
				return
			}
			signal, ok := msg.(onOffPeerSignal)
			if !ok {
				log.Warn("listenOnOffPeers received message is not a on/off peer signal!")
				continue
			}
			//log.Warn("OnOff peer", "key", signal.peer.Key(), "on", signal.on)
			if signal.on {
				klb.addedPeer(signal.peer, signal.po)
			} else {
				klb.resourceUseStats.RemoveResource(signal.peer)
			}
		}
	}
}

// addedPeer is called back when a new peer is added to the kademlia. Its uses will be initialized
// to the use count of the least used peer in its bin. The po of the new peer is passed to avoid having
// to calculate it again.
func (klb *KademliaLoadBalancer) addedPeer(peer *Peer, po int) {
	initCount := klb.initCountFunc(peer, 0)
	log.Debug("Adding peer", "key", peer.Label(), "initCount", initCount)
	klb.resourceUseStats.InitKey(peer.Key(), initCount)
}

// leastUsedCountInBin returns the use count for the least used peer in this bin excluding the excludePeer.
func (klb *KademliaLoadBalancer) leastUsedCountInBin(excludePeer *Peer, po int) int {
	addr := klb.kademlia.BaseAddr()
	peersInSamePo := klb.getPeersForPo(addr, po)
	leastUsedCount := 0
	for i := 0; i < len(peersInSamePo); i++ {
		leastUsed := peersInSamePo[i]
		if leastUsed.Peer.Key() != excludePeer.Key() {
			leastUsedCount = klb.resourceUseStats.GetUses(leastUsed.Peer)
			log.Debug("Least used peer is", "peer", leastUsed.Peer.Label(), "leastUsedCount", leastUsedCount)
			break
		}
	}
	return leastUsedCount
}

// nearestNeighbourUseCount returns the use count for the closest peer count.
func (klb *KademliaLoadBalancer) nearestNeighbourUseCount(newPeer *Peer, _ int) int {
	var count int
	klb.kademlia.EachConn(newPeer.Address(), 255, func(peer *Peer, po int) bool {
		if !bytes.Equal(peer.OAddr, newPeer.OAddr) {
			count = klb.resourceUseStats.GetUses(peer)
			log.Debug("Nearest neighbour is", "peer", peer.Label(), "count", count)
			return false
		}
		return true
	})
	return count
}

func (klb *KademliaLoadBalancer) toLBPeers(resources []resourceusestats.Resource) []LBPeer {
	peers := make([]LBPeer, len(resources))
	for i, res := range resources {
		peer := res.(*Peer)
		peers[i].Peer = peer
		peers[i].stats = klb.resourceUseStats
	}
	return peers
}

func (klb *KademliaLoadBalancer) getPeersForPo(base []byte, po int) []LBPeer {
	resources := make([]resourceusestats.Resource, 0)
	klb.kademlia.EachBinDesc(base, po, func(bin *PeerBin) bool {
		if bin.ProximityOrder == po {
			return bin.PeerIterator(func(entry *entry) bool {
				resources = append(resources, entry.conn)
				return true
			})
		} else {
			return true
		}
	})
	return klb.resourcesToLbPeers(resources)
}
