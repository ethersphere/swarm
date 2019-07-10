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

package newstream

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

// should be changed in tests only
var syncBinsWithinDepth = true

type syncProvider struct {
	netStore *storage.NetStore
	kad      *network.Kademlia

	name string
	quit chan struct{}
}

func NewSyncProvider(ns *storage.NetStore, kad *network.Kademlia) *syncProvider {
	s := &syncProvider{
		netStore: ns,
		kad:      kad,
		name:     streamName,

		quit: make(chan struct{}),
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
	ch, err := s.netStore.Store.Get(ctx, chunk.ModeGetSync, addr)
	if err != nil {
		return nil, err
	}

	// mark the chunk as Set in order to allow for garbage collection
	// this can and at some point should be moved to a dedicated method that
	// marks an entire sent batch of chunks as Set once the actual p2p.Send succeeds
	err = s.netStore.Store.Set(context.Background(), chunk.ModeSetSync, addr)
	if err != nil {
		metrics.GetOrRegisterCounter("syncer.set-next-batch.set-sync-err", nil).Inc(1)
		return nil, err
	}
	return ch.Data(), nil
}

func (s *syncProvider) Put(ctx context.Context, addr chunk.Address, data []byte) (exists bool, err error) {
	log.Trace("syncProvider.Put", "addr", addr)
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
	log.Debug("syncProvider.Subscribe", "bin", bin, "from", from, "to", to)

	return s.netStore.SubscribePull(ctx, bin, from, to)
}

func (s *syncProvider) CursorStr(k string) (cursor uint64, err error) {
	key, err := s.ParseKey(k)
	if err != nil {
		// error parsing the stream key,
		log.Error("error parsing the stream key", "key", k)
		return 0, err
	}

	bin, ok := key.(uint8)
	if !ok {
		return 0, errors.New("could not unmarshal key to uint8")
	}
	return s.netStore.LastPullSubscriptionBinID(bin)
}

func (s *syncProvider) Cursor(key interface{}) (uint64, error) {
	bin, ok := key.(uint8)
	if !ok {
		return 0, errors.New("error converting stream key to bin index")
	}
	return s.netStore.LastPullSubscriptionBinID(bin)
}

// InitPeer creates and maintains the streams per peer.
// Runs per peer, in a separate goroutine
// when the depth changes on our node
//  - peer moves from out-of-depth to depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - peer moves from depth to out-of-depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - depth changes, and peer stays in depth, but we need MORE (or LESS) streams (WHY???).. so again -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
// peer connects and disconnects quickly
func (s *syncProvider) InitPeer(p *Peer) {
	po := chunk.Proximity(p.BzzAddr.Over(), s.kad.BaseAddr())
	depth := s.kad.NeighbourhoodDepth()

	wasWithinDepth := po >= depth

	log.Debug("update syncing subscriptions: initial", "peer", p.ID(), "po", po, "depth", depth)

	// initial subscriptions
	subBins, quitBins := syncSubscriptionsDiff(po, -1, depth, s.kad.MaxProxDisplay, syncBinsWithinDepth)
	s.updateSyncSubscriptions(p, subBins, quitBins)

	depthChangeSignal, unsubscribeDepthChangeSignal := s.kad.SubscribeToNeighbourhoodDepthChange()
	defer unsubscribeDepthChangeSignal()

	for {
		select {
		case _, ok := <-depthChangeSignal:
			if !ok {
				return
			}
			newDepth := s.kad.NeighbourhoodDepth()
			if po >= newDepth && !wasWithinDepth {
				// previous depth is -1 because we did not have any streams with the client beforehand
				depth = -1
			}

			subBins, quitBins := syncSubscriptionsDiff(po, depth, newDepth, s.kad.MaxProxDisplay, syncBinsWithinDepth)
			s.updateSyncSubscriptions(p, subBins, quitBins)

			wasWithinDepth = po >= newDepth
			depth = newDepth
		case <-s.quit:
			return
		}
	}
}

// updateSyncSubscriptions accepts two slices of integers, the first one
// representing proximity order bins for required syncing subscriptions
// and the second one representing bins for syncing subscriptions that
// need to be removed.
func (s *syncProvider) updateSyncSubscriptions(p *Peer, subBins, quitBins []int) {
	log.Debug("update syncing subscriptions", "peer", p.ID(), "subscribe", subBins, "quit", quitBins)
	if l := len(subBins); l > 0 {
		streams := make([]ID, l)
		for i, po := range subBins {

			stream := NewID(s.StreamName(), strconv.Itoa(po))
			_, err := p.getOrCreateInterval(p.peerStreamIntervalKey(stream))
			if err != nil {
				log.Error("got an error while trying to register initial streams", "peer", p.ID(), "stream", stream)
			}

			streams[i] = stream
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := p.Send(ctx, StreamInfoReq{Streams: streams}); err != nil {
			log.Error("error establishing subsequent subscription", "err", err)
			p.Drop()
		}
	}
	for _, po := range quitBins {
		log.Debug("removing cursor info for peer", "peer", p.ID(), "bin", po, "cursors", p.streamCursors)
		p.deleteCursor(NewID(streamName, strconv.Itoa(po)))
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
func syncSubscriptionsDiff(peerPO, prevDepth, newDepth, max int, syncBinsWithinDepth bool) (subBins, quitBins []int) {
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
func intRange(start, end int) (r []int) {
	for i := start; i < end; i++ {
		r = append(r, i)
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
