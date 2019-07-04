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
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/storage"
)

const streamName = "SYNC"

type syncProvider struct {
	netStore *storage.NetStore
	kad      *network.Kademlia

	name         string
	initBehavior StreamInitBehavior
	quit         chan struct{}
}

func NewSyncProvider(ns *storage.NetStore, kad *network.Kademlia, initBehavior StreamInitBehavior) *syncProvider {
	s := &syncProvider{
		netStore: ns,
		kad:      kad,
		name:     streamName,

		initBehavior: initBehavior,
		quit:         make(chan struct{}),
	}
	return s
}

func (s *syncProvider) NeedData(ctx context.Context, key []byte) (loaded bool, wait func(context.Context) error) {
	start := time.Now()

	fi, loaded, ok := s.netStore.GetOrCreateFetcher(ctx, key, "syncer")
	if !ok {
		return loaded, nil
	}

	return loaded, func(ctx context.Context) error {
		select {
		case <-fi.Delivered:
			metrics.GetOrRegisterResettingTimer(fmt.Sprintf("fetcher.%s.syncer", fi.CreatedBy), nil).UpdateSince(start)
		case <-time.After(timeouts.SyncerClientWaitTimeout):
			metrics.GetOrRegisterCounter("fetcher.syncer.timeout", nil).Inc(1)
			return fmt.Errorf("chunk not delivered through syncing after %dsec. ref=%s", timeouts.SyncerClientWaitTimeout, fmt.Sprintf("%x", key))
		}
		return nil
	}
}

func (s *syncProvider) Get(ctx context.Context, addr chunk.Address) ([]byte, error) {
	log.Debug("syncProvider.Get")
	//err := p.syncer.netStore.Set(context.Background(), chunk.ModeSetSync, d.Address)
	ch, err := s.netStore.Store.Get(ctx, chunk.ModeGetSync, addr)
	if err != nil {
		return nil, err
	}
	log.Debug("got chunk")
	return ch.Data(), nil
}

func (s *syncProvider) Put(ctx context.Context, addr chunk.Address, data []byte) (exists bool, err error) {
	log.Debug("syncProvider.Put", "addr", addr)
	ch := chunk.NewChunk(addr, data)
	seen, err := s.netStore.Store.Put(ctx, chunk.ModePutSync, ch)
	if seen {
		log.Trace("syncProvider.Put - chunk already seen", "addr", addr)
	}
	return seen, err
}

func (s *syncProvider) Subscribe(ctx context.Context, key interface{}, from, to uint64) (<-chan chunk.Descriptor, func()) {
	// convert the key to the actual value and call SubscribePull
	bin := key.(uint8)
	log.Debug("sync provider subscribing on key", "key", key, "bin", bin)

	return s.netStore.SubscribePull(ctx, bin, from, to)
}

func (s *syncProvider) Cursor(key interface{}) (uint64, error) {
	bin, ok := key.(uint8)
	if !ok {
		return 0, errors.New("error converting stream key to bin index")
	}
	return s.netStore.LastPullSubscriptionBinID(bin)
}

// RunUpdateStreams creates and maintains the streams per peer.
// Runs per peer, in a separate goroutine
// when the depth changes on our node
//  - peer moves from out-of-depth to depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - peer moves from depth to out-of-depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - depth changes, and peer stays in depth, but we need MORE (or LESS) streams (WHY???).. so again -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
// peer connects and disconnects quickly
func (s *syncProvider) RunUpdateStreams(p *Peer) {

	peerPo := chunk.Proximity(s.kad.BaseAddr(), p.BzzAddr.Address())
	depth := s.kad.NeighbourhoodDepth()
	withinDepth := peerPo >= depth
	p.mtx.Lock()
	p.dirtyStreams[s.StreamName()] = true
	p.mtx.Unlock()

	//run in the background:
	/*
		if the streams are dirty - request the streams (lets assume from -1 to depth and set dirty = false(!) that is because we just requested them. the bool is indicative of what should be requested

		wait for the reply:
		 - if the streams are still dirty when the reply comes in - it means they are now incorrect. dump the reply and request them again with the current depth
		 - if the streams are not dirty, it means they are correct - store the cursors and request the streams according to logic
		 - when we write the stream cursors - we must assume, that if an entry is already there - it probably means that there's a goroutine running and fetching the stream in the background
			 so.... if we want to stop that goroutine - we must set the record to be nil. but that does no guaranty that the goroutine will actually stop
	*/

	log.Debug("create streams", "peer", p.BzzAddr.ID(), "base", fmt.Sprintf("%x", s.kad.BaseAddr()[:12]), "withinDepth", withinDepth, "depth", depth, "po", peerPo)

	sub, _ := syncSubscriptionsDiff(peerPo, -1, depth, s.kad.MaxProxDisplay, true)
	log.Debug("sending initial subscriptions message", "self", fmt.Sprintf("%x", s.kad.BaseAddr()[:12]), "peer", p.ID(), "subs", sub)
	doPeerSubUpdate(p, sub, nil)

	subscription, unsubscribe := s.kad.SubscribeToNeighbourhoodDepthChange()
	defer unsubscribe()
	for {
		select {
		case <-subscription:
			p.mtx.Lock()
			p.dirtyStreams[s.StreamName()] = true
			p.mtx.Unlock()

			newDepth := s.kad.NeighbourhoodDepth()
			log.Debug("got kademlia depth change sig", "peer", p.ID(), "peerPo", peerPo, "depth", depth, "newDepth", newDepth, "withinDepth", withinDepth)
			switch {
			case peerPo >= newDepth:
				// peer is within depth
				if !withinDepth {
					log.Debug("peer moved into depth, requesting cursors", "peer", p.ID())
					withinDepth = true // peerPo >= newDepth
					// previous depth is -1 because we did not have any streams with the client beforehand
					sub, _ := syncSubscriptionsDiff(peerPo, -1, newDepth, s.kad.MaxProxDisplay, true)
					log.Debug("getting cursors info from peer", "self", fmt.Sprintf("%x", s.kad.BaseAddr()[:16]), "peer", p.ID(), "subs", sub)
					doPeerSubUpdate(p, sub, nil)
					depth = newDepth
				} else {
					// peer was within depth, but depth has changed. we should request the cursors for the
					// necessary bins and quit the unnecessary ones
					sub, quits := syncSubscriptionsDiff(peerPo, depth, newDepth, s.kad.MaxProxDisplay, true)
					log.Debug("peer was inside depth, checking if needs changes", "peer", p.ID(), "peerPo", peerPo, "depth", depth, "newDepth", newDepth, "subs", sub, "quits", quits)
					log.Debug("getting cursors info from peer", "self", fmt.Sprintf("%x", s.kad.BaseAddr()[:16]), "peer", p.ID(), "subs", sub)
					doPeerSubUpdate(p, sub, quits)
					//p.mtx.Lock()
					/////unset.../
					//p.mtx.Unlock()
					depth = newDepth
				}
			case peerPo < newDepth:
				if withinDepth {
					sub, quits := syncSubscriptionsDiff(peerPo, depth, newDepth, s.kad.MaxProxDisplay, true)
					log.Debug("peer transitioned out of depth", "peer", p.ID(), "subs", sub, "quits", quits)
					log.Debug("getting cursors info from peer", "self", fmt.Sprintf("%x", s.kad.BaseAddr()[:16]), "peer", p.ID(), "subs", sub)
					doPeerSubUpdate(p, sub, quits)
					withinDepth = false
				}
			}
			p.mtx.Lock()
			p.dirtyStreams[s.StreamName()] = false
			p.mtx.Unlock()

		case <-s.quit:
			return
		}
	}
}

func doPeerSubUpdate(p *Peer, subs, quits []uint) {
	if len(subs) > 0 {
		streams := []ID{}
		for _, v := range subs {
			vv := NewID(streamName, fmt.Sprintf("%d", v))
			streams = append(streams, vv)
		}
		streamsMsg := StreamInfoReq{Streams: streams}
		if err := p.Send(context.TODO(), streamsMsg); err != nil {
			log.Error("error establishing subsequent subscription", "err", err)
			p.Drop()
		}
	}

	for _, v := range quits {
		log.Debug("removing cursor info for peer", "peer", p.ID(), "bin", v, "cursors", p.streamCursors, "quits", quits)

		vv := NewID(streamName, fmt.Sprintf("%d", v))

		p.deleteCursor(vv)
	}
}

// syncSubscriptionsDiff calculates to which proximity order bins a peer
// (with po peerPO) needs to be subscribed after kademlia neighbourhood depth
// change from prevDepth to newDepth. Max argument limits the number of
// proximity order bins. Returned values are slices of integers which represent
// proximity order bins, the first one to which additional subscriptions need to
// be requested and the second one which subscriptions need to be quit. Argument
// prevDepth with value less then 0 represents no previous depth, used for
// initial syncing subscriptions.
func syncSubscriptionsDiff(peerPO, prevDepth, newDepth, max int, syncBinsWithinDepth bool) (subBins, quitBins []uint) {
	newStart, newEnd := syncBins(peerPO, newDepth, max, syncBinsWithinDepth)
	if prevDepth < 0 {
		if newStart == -1 && newEnd == -1 {
			return nil, nil
		}
		// no previous depth, return the complete range
		// for subscriptions requests and nothing for quitting
		return intRange(newStart, newEnd), nil
	}

	prevStart, prevEnd := syncBins(peerPO, prevDepth, max, syncBinsWithinDepth)
	if newStart == -1 && newEnd == -1 {
		// this means that we should not have any streams on any bins with this peer
		// get rid of what was established on the previous depth
		quitBins = append(quitBins, intRange(prevStart, prevEnd)...)
		return
	}

	if newStart < prevStart {
		subBins = append(subBins, intRange(newStart, prevStart)...)
	}

	if prevStart < newStart {
		quitBins = append(quitBins, intRange(prevStart, newStart)...)
	}

	if newEnd < prevEnd {
		quitBins = append(quitBins, intRange(newEnd, prevEnd)...)
	}

	if prevEnd < newEnd {
		subBins = append(subBins, intRange(prevEnd, newEnd)...)
	}

	return subBins, quitBins
}

// syncBins returns the range to which proximity order bins syncing
// subscriptions need to be requested, based on peer proximity and
// kademlia neighbourhood depth. Returned range is [start,end), inclusive for
// start and exclusive for end.
func syncBins(peerPO, depth, max int, syncBinsWithinDepth bool) (start, end int) {
	if syncBinsWithinDepth && peerPO < depth {
		// we don't want to request anything from peers outside depth
		return -1, -1
	} else {
		if peerPO < depth {
			// subscribe only to peerPO bin if it is not
			// in the nearest neighbourhood
			return peerPO, peerPO + 1
		}
	}
	// subscribe from depth to max bin if the peer
	// is in the nearest neighbourhood
	return depth, max + 1
}

// intRange returns the slice of integers [start,end). The start
// is inclusive and the end is not.
func intRange(start, end int) (r []uint) {
	for i := start; i < end; i++ {
		r = append(r, uint(i))
	}
	return r
}

func (s *syncProvider) ParseKey(streamKey string) (interface{}, error) {
	b, err := strconv.Atoi(streamKey)
	if err != nil {
		return 0, err
	}
	if b < 0 || b > 16 {
		return 0, errors.New("stream key out of range")
	}
	return uint8(b), nil
}
func (s *syncProvider) EncodeKey(i interface{}) (string, error) {
	v, ok := i.(uint8)
	if !ok {
		return "", errors.New("error encoding key")
	}
	return fmt.Sprintf("%d", v), nil
}
func (s *syncProvider) StreamName() string { return s.name }

func (s *syncProvider) Boundedness() bool { return false }

func (s *syncProvider) StreamBehavior() StreamInitBehavior { return s.initBehavior }
