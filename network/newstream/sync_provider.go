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
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/storage"
	lru "github.com/hashicorp/golang-lru"
)

const syncStreamName = "SYNC"
const cacheCapacity = 10000

type syncProvider struct {
	netStore                *storage.NetStore // netstore
	kad                     *network.Kademlia // kademlia
	name                    string            // name of the stream we are responsible for
	syncBinsOnlyWithinDepth bool              // true means streams are established only within depth, false means outside of depth too
	autostart               bool              // start fetching streams automatically when cursors arrive from peer
	quit                    chan struct{}     // shutdown
	cacheMtx                sync.RWMutex      // synchronization primitive to protect cache
	cache                   *lru.Cache        // cache to minimize load on netstore
	logger                  log.Logger        // logger that appends the base address to loglines
}

// NewSyncProvider creates a new sync provider that is used by the stream protocol to sink data and control its behaviour
// syncOnlyWithinDepth toggles stream establishment in reference to kademlia. When true - streams are
// established only within depth ( >=depth ). This is needed for Push Sync. When set to false, the streams are
// established on all bins as they did traditionally with Pull Sync.
func NewSyncProvider(ns *storage.NetStore, kad *network.Kademlia, autostart bool, syncOnlyWithinDepth bool) StreamProvider {
	c, err := lru.New(cacheCapacity)
	if err != nil {
		panic(err)
	}

	return &syncProvider{
		netStore:                ns,
		kad:                     kad,
		syncBinsOnlyWithinDepth: syncOnlyWithinDepth,
		autostart:               autostart,
		name:                    syncStreamName,
		quit:                    make(chan struct{}),
		cache:                   c,
		logger:                  log.New("base", hex.EncodeToString(kad.BaseAddr()[:16])),
	}
}

// NeedData checks if we need to retrieve the supplied addrs from the upstream peer
func (s *syncProvider) NeedData(ctx context.Context, addrs ...chunk.Address) ([]bool, error) {
	var (
		start   = time.Now()
		wants   = make([]bool, len(addrs)) // which addresses we want
		check   = make([]chunk.Address, 0) // which addresses to check in localstore
		indexes = make([]int, 0)
	)

	// don't check if we're shutting down
	select {
	case <-s.quit:
		return wants, nil
	default:
	}

	// if the cache contains the chunk key - it is most probable to exist in the localstore
	// therefore we do not want the chunk
	// when the chunk is not in the cache - we check the localstore, and if it does not
	// exist there - we ask for it
	s.cacheMtx.RLock()
	for i, addr := range addrs {
		if !s.cache.Contains(addr.Hex()) {
			// chunk is not in the cache - check in the localstore and if its not there - we want it
			check = append(check, addr)
			indexes = append(indexes, i)
			metrics.GetOrRegisterCounter("network.stream.sync_provider.multi_need_data.cachemiss", nil).Inc(1)
		} else {
			// chunk is in the cache - we don't want it
			wants[i] = false
			metrics.GetOrRegisterCounter("network.stream.sync_provider.multi_need_data.cachehit", nil).Inc(1)
		}
	}
	s.cacheMtx.RUnlock()

	// check localstore for the remaining chunks
	has, err := s.netStore.Store.HasMulti(ctx, check...)
	if err != nil {
		return nil, err
	}

	// inspect results
	for i, have := range has {
		if !have {
			wants[indexes[i]] = true // if we dont have it - we want it
			fi, _, ok := s.netStore.GetOrCreateFetcher(ctx, check[i], "syncer")
			if !ok {
				continue
			}

			go func() {
				select {
				case <-fi.Delivered:
					metrics.GetOrRegisterResettingTimer(fmt.Sprintf("fetcher.%s.syncer", fi.CreatedBy), nil).UpdateSince(start)
				case <-time.After(timeouts.SyncerClientWaitTimeout):
					metrics.GetOrRegisterCounter("fetcher.syncer.timeout", nil).Inc(1)
				}
			}()
		} else {
			// if we have it - we dont want it
			wants[indexes[i]] = false
		}
	}
	return wants, nil
}

// Get the supplied addresses for delivery
func (s *syncProvider) Get(ctx context.Context, addr ...chunk.Address) ([]chunk.Chunk, error) {
	var (
		start     = time.Now()                     // start time
		retChunks = make([]chunk.Chunk, len(addr)) //the chunks we want to Get
		lsChunks  = make([]chunk.Address, 0)       // the chunks that we need to Get from localstore
		indices   = make([]int, 0)                 // backreferences to glue retChunks and lsChunks together
	)

	defer func(start time.Time) {
		metrics.GetOrRegisterResettingTimer("network.stream.sync_provider.get.total-time", nil).UpdateSince(start)
	}(start)

	s.cacheMtx.RLock()
	// iterate over the array - if it is in the cache - pull it out
	// if not - save in a slice and fallback later to localstore in one go
	for i, a := range addr {
		if v, ok := s.cache.Get(a.Hex()); ok {
			retChunks[i] = chunk.NewChunk(a, v.([]byte))
			metrics.GetOrRegisterCounter("network.stream.sync_provider.get.cachehit", nil).Inc(1)
		} else {
			lsChunks = append(lsChunks, a)
			indices = append(indices, i)
			metrics.GetOrRegisterCounter("network.stream.sync_provider.get.cachemiss", nil).Inc(1)
		}
	}
	s.cacheMtx.RUnlock()

	// get the rest from localstore
	chunks, err := s.netStore.GetMulti(ctx, chunk.ModeGetSync, lsChunks...)
	if err != nil {
		return nil, err
	}
	s.cacheMtx.Lock()
	defer s.cacheMtx.Unlock()

	// merge the results together
	for i, ch := range chunks {
		ch := ch
		s.cache.Add(ch.Address().Hex(), ch.Data())
		retChunks[indices[i]] = ch
	}
	return retChunks, nil
}

// Set the supplied addrs as synced in order to allow for garbage collection
func (s *syncProvider) Set(ctx context.Context, addrs ...chunk.Address) error {
	err := s.netStore.Set(ctx, chunk.ModeSetSync, addrs...)
	if err != nil {
		metrics.GetOrRegisterCounter("syncProvider.set-sync-err", nil).Inc(1)
		return err
	}
	return nil
}

// Put the given chunks to the local storage
func (s *syncProvider) Put(ctx context.Context, ch ...chunk.Chunk) (exists []bool, err error) {
	seen, err := s.netStore.Put(ctx, chunk.ModePutSync, ch...)
	for i, v := range seen {
		if v {
			if putSeenTestHook != nil {
				// call the test function if it is set
				putSeenTestHook(ch[i].Address(), s.netStore.LocalID)
			}
		}
	}
	go func(chunks ...chunk.Chunk) {
		s.cacheMtx.Lock()
		defer s.cacheMtx.Unlock()
		for _, c := range chunks {
			s.cache.Add(c.Address().Hex(), c.Data())
		}
	}(ch...)
	return seen, err
}

// Function used only in tests to detect chunks that are synced
// multiple times within the same stream. This function pointer must be
// nil in production.
var putSeenTestHook func(addr chunk.Address, id enode.ID)

// Subscribe wraps SubscribePull to retrieve chunks within a certain interval
func (s *syncProvider) Subscribe(ctx context.Context, key interface{}, from, to uint64) (<-chan chunk.Descriptor, func()) {
	// convert the key to the actual value and call SubscribePull
	bin := key.(uint8)
	log.Debug("syncProvider.Subscribe", "bin", bin, "from", from, "to", to)

	return s.netStore.SubscribePull(ctx, bin, from, to)
}

// Cursor gets the cursor from the localstore for a given stream key
func (s *syncProvider) Cursor(k string) (cursor uint64, err error) {
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

// WantStream checks if we are interested in a given stream for a peer
func (s *syncProvider) WantStream(p *Peer, streamID ID) bool {
	p.logger.Debug("syncProvider.WantStream", "stream", streamID)
	po := chunk.Proximity(p.BzzAddr.Over(), s.kad.BaseAddr())
	depth := s.kad.NeighbourhoodDepth()

	// check all subscriptions that should exist for this peer
	subBins, _ := syncSubscriptionsDiff(po, -1, depth, s.kad.MaxProxDisplay, s.syncBinsOnlyWithinDepth)
	v, err := parseSyncKey(streamID.Key)
	if err != nil {
		return false
	}
	return checkKeyInSlice(int(v), subBins)
}

var (
	SyncInitBackoff = 500 * time.Millisecond
)

// InitPeer creates and maintains the streams per peer.
// Runs per peer, in a separate goroutine
// when the depth changes on our node
//  - peer moves from out-of-depth to depth
//  - peer moves from depth to out-of-depth
//  - depth changes, and peer stays in depth, but we need more or less
// peer connects and disconnects quickly
func (s *syncProvider) InitPeer(p *Peer) {
	p.logger.Debug("syncProvider.InitPeer")
	timer := time.NewTimer(SyncInitBackoff)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-p.quit:
		return
	}

	po := chunk.Proximity(p.BzzAddr.Over(), s.kad.BaseAddr())
	depth := s.kad.NeighbourhoodDepth()

	p.logger.Debug("update syncing subscriptions: initial", "po", po, "depth", depth)

	subBins, quitBins := syncSubscriptionsDiff(po, -1, depth, s.kad.MaxProxDisplay, s.syncBinsOnlyWithinDepth)
	s.updateSyncSubscriptions(p, subBins, quitBins)

	depthChangeSignal, unsubscribeDepthChangeSignal := s.kad.SubscribeToNeighbourhoodDepthChange()
	defer unsubscribeDepthChangeSignal()

	for {
		select {
		case _, ok := <-depthChangeSignal:
			if !ok {
				return
			}

			// update subscriptions for this peer when depth changes
			ndepth := s.kad.NeighbourhoodDepth()
			subs, quits := syncSubscriptionsDiff(po, depth, ndepth, s.kad.MaxProxDisplay, s.syncBinsOnlyWithinDepth)
			p.logger.Debug("update syncing subscriptions", "po", po, "depth", depth, "sub", subs, "quit", quits)
			s.updateSyncSubscriptions(p, subs, quits)
			depth = ndepth
		case <-s.quit:
			return
		case <-p.quit:
			return
		}

	}
}

// updateSyncSubscriptions accepts two slices of integers, the first one
// representing proximity order bins for required syncing subscriptions
// and the second one representing bins for syncing subscriptions that
// need to be removed.
func (s *syncProvider) updateSyncSubscriptions(p *Peer, subBins, quitBins []int) {
	p.logger.Debug("syncProvider.updateSyncSubscriptions", "subBins", subBins, "quitBins", quitBins)
	if l := len(subBins); l > 0 {
		streams := make([]ID, l)
		for i, po := range subBins {

			stream := NewID(s.StreamName(), encodeSyncKey(uint8(po)))
			_, err := p.getOrCreateInterval(p.peerStreamIntervalKey(stream))
			if err != nil {
				p.logger.Error("got an error while trying to register initial streams", "stream", stream)
			}

			streams[i] = stream
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := p.Send(ctx, StreamInfoReq{Streams: streams}); err != nil {
			p.logger.Error("error establishing subsequent subscription", "err", err)
			p.Drop()
			return
		}
	}
	for _, po := range quitBins {
		p.logger.Debug("stream unwanted, removing cursor info for peer", "bin", po)
		p.deleteCursor(NewID(syncStreamName, encodeSyncKey(uint8(po))))
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
// syncBinsOnlyWithinDepth toggles between having requested streams only within depth(true)
// or rather with the old stream establishing logic (false)
func syncSubscriptionsDiff(peerPO, prevDepth, newDepth, max int, syncBinsOnlyWithinDepth bool) (subBins, quitBins []int) {
	newStart, newEnd := syncBins(peerPO, newDepth, max, syncBinsOnlyWithinDepth)
	if prevDepth < 0 {
		if newStart == -1 && newEnd == -1 {
			return nil, nil
		}
		// no previous depth, return the complete range
		// for subscriptions requests and nothing for quitting
		return intRange(newStart, newEnd), nil
	}

	prevStart, prevEnd := syncBins(peerPO, prevDepth, max, syncBinsOnlyWithinDepth)
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
// syncBinsOnlyWithinDepth toggles between having requested streams only within depth(true)
// or rather with the old stream establishing logic (false)
func syncBins(peerPO, depth, max int, syncBinsOnlyWithinDepth bool) (start, end int) {
	if syncBinsOnlyWithinDepth && peerPO < depth {
		// we don't want to request anything from peers outside depth
		return -1, -1
	}
	if peerPO < depth {
		// subscribe only to peerPO bin if it is not
		// in the nearest neighbourhood
		return peerPO, peerPO + 1
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

func checkKeyInSlice(k int, slice []int) (found bool) {
	for _, v := range slice {
		if v == k {
			found = true
		}
	}
	return
}

func (s *syncProvider) ParseKey(streamKey string) (interface{}, error) {
	return parseSyncKey(streamKey)
}

func (s *syncProvider) EncodeKey(i interface{}) (string, error) {
	v, ok := i.(uint8)
	if !ok {
		return "", errors.New("error encoding key")
	}
	return encodeSyncKey(v), nil
}

func (s *syncProvider) StreamName() string { return s.name }

func (s *syncProvider) Boundedness() bool { return false }

func (s *syncProvider) Autostart() bool { return s.autostart }

func (s *syncProvider) Close() { close(s.quit) }

func parseSyncKey(streamKey string) (uint8, error) {
	b, err := strconv.ParseUint(streamKey, 36, 8)
	if err != nil {
		return 0, err
	}
	if b < 0 || b > chunk.MaxPO {
		return 0, fmt.Errorf("stream key %v out of range", b)
	}
	return uint8(b), nil
}

func encodeSyncKey(i uint8) string {
	return strconv.FormatUint(uint64(i), 36)
}
