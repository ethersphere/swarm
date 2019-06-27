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
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	bv "github.com/ethersphere/swarm/network/bitvector"
	"github.com/ethersphere/swarm/storage"
)

var ErrEmptyBatch = errors.New("empty batch")

const BatchSize = 128

// Peer is the Peer extension for the streaming protocol
type Peer struct {
	*network.BzzPeer
	mtx          sync.Mutex
	streamsDirty bool // a request for StreamInfo is underway and awaiting reply
	syncer       *SwarmSyncer

	streamCursors     map[uint]uint64           // key: bin, value: session cursor. when unset - we are not interested in that bin
	historicalStreams map[uint]*syncStreamFetch //maintain state for each stream fetcher
	//openOffers map[uint]

	quit chan struct{}
}

// NewPeer is the constructor for Peer
func NewPeer(peer *network.BzzPeer, s *SwarmSyncer) *Peer {
	p := &Peer{
		BzzPeer:           peer,
		streamCursors:     make(map[uint]uint64),
		historicalStreams: make(map[uint]*syncStreamFetch),
		syncer:            s,
		quit:              make(chan struct{}),
	}
	return p
}

func (p *Peer) Left() {
	close(p.quit)
}

// HandleMsg is the message handler that delegates incoming messages
func (p *Peer) HandleMsg(ctx context.Context, msg interface{}) error {
	switch msg := msg.(type) {
	case *StreamInfoReq:
		go p.handleStreamInfoReq(ctx, msg)
	case *StreamInfoRes:
		go p.handleStreamInfoRes(ctx, msg)
	case *GetRange:
		go p.handleGetRange(ctx, msg)
	case *OfferedHashes:
		go p.handleOfferedHashes(ctx, msg)
	case *WantedHashes:
		go p.handleWantedHashes(ctx, msg)

	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}
	return nil
}

// handleStreamInfoReq handles the StreamInfoReq message.
// this message is handled by the SERVER (*Peer is the client in this case)
func (p *Peer) handleStreamInfoReq(ctx context.Context, msg *StreamInfoReq) {
	log.Debug("handleStreamInfoReq", "peer", p.ID(), "msg", msg)
	p.mtx.Lock()
	defer p.mtx.Unlock()
	streamRes := StreamInfoRes{}
	if len(msg.Streams) == 0 {
		panic("nil streams msg requested")
	}
	for _, v := range msg.Streams {
		streamCursor, err := p.syncer.netStore.LastPullSubscriptionBinID(uint8(v))
		if err != nil {
			log.Error("error getting last bin id", "bin", v)
			panic("shouldnt happen")
		}
		descriptor := StreamDescriptor{
			Name:    fmt.Sprintf("SYNC|%d", v),
			Cursor:  streamCursor,
			Bounded: false,
		}
		streamRes.Streams = append(streamRes.Streams, descriptor)
	}
	if err := p.Send(ctx, streamRes); err != nil {
		log.Error("failed to send StreamInfoRes to client", "requested bins", msg.Streams)
	}
}

// handleStreamInfoRes handles the StreamInfoRes message.
// this message is handled by the CLIENT (*Peer is the server in this case)
func (p *Peer) handleStreamInfoRes(ctx context.Context, msg *StreamInfoRes) {
	log.Debug("handleStreamInfoRes", "peer", p.ID(), "msg", msg)
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if len(msg.Streams) == 0 {
		log.Error("StreamInfo response is empty")
		p.Drop()
	}

	for _, s := range msg.Streams {
		bin, err := ParseStream(s.Name)
		if err != nil {
			log.Error("error parsing stream", "stream", s.Name)
			p.Drop()
		}
		log.Debug("setting bin cursor", "peer", p.ID(), "bin", uint(bin), "cursor", s.Cursor)

		p.streamCursors[uint(bin)] = s.Cursor
		if s.Cursor > 0 {
			streamFetch := newSyncStreamFetch(uint(bin))
			p.historicalStreams[uint(bin)] = streamFetch
			g := GetRange{
				Ruid:      uint(rand.Uint32()),
				Stream:    s.Name,
				From:      1, //this should be from the interval store
				To:        s.Cursor,
				BatchSize: 128,
				Roundtrip: true,
			}
			log.Debug("sending first GetRange to peer", "peer", p.ID(), "bin", uint(bin), "cursor", s.Cursor, "GetRange", g)

			if err := p.Send(ctx, g); err != nil {
				log.Error("had an error sending initial GetRange for historical stream", "peer", p.ID(), "stream", s, "GetRange", g, "err", err)
				p.Drop()
			}
		}
	}
}

func (p *Peer) handleGetRange(ctx context.Context, msg *GetRange) {
	log.Debug("peer.handleGetRange", "peer", p.ID(), "msg", msg)
	bin, err := ParseStream(msg.Stream)
	if err != nil {
		log.Error("erroring parsing stream", "err", err, "stream", msg.Stream)
		p.Drop()
	}
	//TODO hard limit for BatchSize
	//TODO check msg integrity
	h, f, t, err := p.collectBatch(ctx, bin, msg.From, msg.To)
	if err != nil {
		log.Error("erroring getting batch for stream", "peer", p.ID(), "bin", bin, "stream", msg.Stream, "err", err)
		p.Drop()
	}

	offered := OfferedHashes{
		Ruid:      msg.Ruid,
		LastIndex: uint(t),
		Hashes:    h,
	}
	l := len(h) / 32
	log.Debug("server offering batch", "peer", p.ID(), "ruid", msg.Ruid, "requestFrom", msg.From, "From", f, "requestTo", msg.To, "hashes", h, "l", l)
	if err := p.Send(ctx, offered); err != nil {
		log.Error("erroring sending offered hashes", "peer", p.ID(), "ruid", msg.Ruid, "err", err)
	}
}

const HashSize = 32

// handleOfferedHashes handles the OfferedHashes wire protocol message.
// this message is handled by the CLIENT.
func (p *Peer) handleOfferedHashes(ctx context.Context, msg *OfferedHashes) {
	log.Debug("peer.handleOfferedHashes", "peer", p.ID(), "msg", msg)
	//TODO if ruid does not exist in state - drop the peer

	hashes := msg.Hashes
	lenHashes := len(hashes)
	if lenHashes%HashSize != 0 {
		log.Error("error invalid hashes length", "len", lenHashes)
	}

	want, err := bv.New(lenHashes / HashSize)
	if err != nil {
		log.Error("error initiaising bitvector", "len", lenHashes/HashSize, "err", err)
		p.Drop()
	}

	ctr := 0

	for i := 0; i < lenHashes; i += HashSize {
		hash := hashes[i : i+HashSize]
		log.Trace("checking offered hash", "ref", fmt.Sprintf("%x", hash))

		if _, wait := p.syncer.NeedData(ctx, hash); wait != nil {
			ctr++

			// set the bit, so create a request
			want.Set(i/HashSize, true)
			log.Trace("need data", "ref", fmt.Sprintf("%x", hash), "request", true)
		}
	}
	// TODO: place goroutine of abstraction to seal off batch HERE
	w := WantedHashes{
		Ruid:      msg.Ruid,
		BitVector: want.Bytes(),
	}
	log.Debug("sending wanted hashes", "peer", p.ID(), "offered", lenHashes/HashSize, "want", ctr)
	if err := p.Send(ctx, w); err != nil {
		log.Error("error sending wanted hashes", "peer", p.ID(), "w", w)
		p.Drop()
	}
}
func (p *Peer) handleWantedHashes(ctx context.Context, msg *WantedHashes) {
	log.Debug("peer.handleWantedHashes", "peer", p.ID(), "ruid", msg.Ruid)
	// Get the length of the original Offer from state
	l := -1
	want, err := bv.NewFromBytes(msg.BitVector, l)
	if err != nil {
		log.Error("error initiaising bitvector", l, err)
	}
	for i := 0; i < l; i++ {
		if want.Get(i) {
			metrics.GetOrRegisterCounter("peer.handlewantedhashesmsg.actualget", nil).Inc(1)

			hash := hashes[i*HashSize : (i+1)*HashSize]
			data, err := p.syncer.GetData(ctx, hash)
			if err != nil {
				log.Error("handleWantedHashesMsg", "hash", hash, "err", err)
				p.Drop()
			}
			chunk := storage.NewChunk(hash, data)
			//collect the chunk into the batch
		}
	}

	//send the batch
	return nil

}

func (p *Peer) collectBatch(ctx context.Context, bin uint, from, to uint64) (hashes []byte, f, t uint64, err error) {
	batchStart := time.Now()
	descriptors, stop := p.syncer.netStore.SubscribePull(ctx, uint8(bin), from, to)
	defer stop()

	const batchTimeout = 2 * time.Second

	var (
		batch        []byte
		batchSize    int
		batchStartID *uint64
		batchEndID   uint64
		timer        *time.Timer
		timerC       <-chan time.Time
	)

	defer func(start time.Time) {
		metrics.GetOrRegisterResettingTimer("syncer.set-next-batch.total-time", nil).UpdateSince(start)
		metrics.GetOrRegisterCounter("syncer.set-next-batch.batch-size", nil).Inc(int64(batchSize))
		if timer != nil {
			timer.Stop()
		}
	}(batchStart)

	for iterate := true; iterate; {
		select {
		case d, ok := <-descriptors:
			if !ok {
				iterate = false
				break
			}
			batch = append(batch, d.Address[:]...)
			// This is the most naive approach to label the chunk as synced
			// allowing it to be garbage collected. A proper way requires
			// validating that the chunk is successfully stored by the peer.
			err := p.syncer.netStore.Set(context.Background(), chunk.ModeSetSync, d.Address)
			if err != nil {
				metrics.GetOrRegisterCounter("syncer.set-next-batch.set-sync-err", nil).Inc(1)
				//log.Debug("syncer pull subscription - err setting chunk as synced", "correlateId", s.correlateId, "err", err)
				return nil, 0, 0, err
			}
			batchSize++
			if batchStartID == nil {
				// set batch start id only if
				// this is the first iteration
				batchStartID = &d.BinID
			}
			batchEndID = d.BinID
			if batchSize >= BatchSize {
				iterate = false
				metrics.GetOrRegisterCounter("syncer.set-next-batch.full-batch", nil).Inc(1)
				log.Trace("syncer pull subscription - batch size reached", "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
			}
			if timer == nil {
				timer = time.NewTimer(batchTimeout)
			} else {
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(batchTimeout)
			}
			timerC = timer.C
		case <-timerC:
			// return batch if new chunks are not
			// received after some time
			iterate = false
			metrics.GetOrRegisterCounter("syncer.set-next-batch.timer-expire", nil).Inc(1)
			//log.Trace("syncer pull subscription timer expired", "correlateId", s.correlateId, "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
		case <-p.syncer.quit:
			iterate = false
			//log.Trace("syncer pull subscription - quit received", "correlateId", s.correlateId, "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
		}
	}
	if batchStartID == nil {
		// if batch start id is not set, it means we timed out
		return nil, 0, 0, ErrEmptyBatch
	}
	return batch, *batchStartID, batchEndID, nil

}

// syncStreamFetch is a struct that holds exposed state used by a separate goroutine that handles stream retrievals
type syncStreamFetch struct {
	bin       uint          //the bin we're working on
	lastIndex uint64        //last chunk bin index that we handled
	quit      chan struct{} //used to signal from other components to quit this stream (i.e. on depth change)
	done      chan struct{} //signaled by the actor on stream fetch done
	err       chan error    //signaled by the actor on error
}

func newSyncStreamFetch(bin uint) *syncStreamFetch {
	return &syncStreamFetch{
		bin:       bin,
		lastIndex: 0,
		quit:      make(chan struct{}),
		done:      make(chan struct{}),
		err:       make(chan error),
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

// CreateStreams creates and maintains the streams per peer.
// Runs per peer, in a separate goroutine
// when the depth changes on our node
//  - peer moves from out-of-depth to depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - peer moves from depth to out-of-depth -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
//  - depth changes, and peer stays in depth, but we need MORE (or LESS) streams (WHY???).. so again -> determine new streams ; init new streams (delete old streams, stop sending get range queries ; graceful shutdown of existing streams)
// peer connects and disconnects quickly
func (s *SwarmSyncer) CreateStreams(p *Peer) {
	defer log.Debug("createStreams closed", "peer", p.ID())

	peerPo := chunk.Proximity(s.kad.BaseAddr(), p.BzzAddr.Address())
	depth := s.kad.NeighbourhoodDepth()
	withinDepth := peerPo >= depth

	log.Debug("create streams", "peer", p.BzzAddr, "base", s.kad.BaseAddr(), "withinDepth", withinDepth, "depth", depth, "po", peerPo)

	if withinDepth {
		sub, _ := syncSubscriptionsDiff(peerPo, -1, depth, s.kad.MaxProxDisplay, true)
		log.Debug("sending initial subscriptions message", "peer", p.ID(), "bins", sub)
		time.Sleep(createStreamsDelay)
		doPeerSubUpdate(p, sub, nil)
		if len(sub) == 0 {
			panic("w00t")
		}
		//if err := p.Send(context.TODO(), streamsMsg); err != nil {
		//log.Error("err establishing initial subscription", "err", err)
		//}
	}

	subscription, unsubscribe := s.kad.SubscribeToNeighbourhoodDepthChange()
	defer unsubscribe()
	for {
		select {
		case <-subscription:
			newDepth := s.kad.NeighbourhoodDepth()
			log.Debug("got kademlia depth change sig", "peer", p.ID(), "peerPo", peerPo, "depth", depth, "newDepth", newDepth, "withinDepth", withinDepth)
			switch {
			//case newDepth == depth:
			//continue
			case peerPo >= newDepth:
				// peer is within depth
				if !withinDepth {
					log.Debug("peer moved into depth, requesting cursors", "peer", p.ID())
					withinDepth = true // peerPo >= newDepth
					// previous depth is -1 because we did not have any streams with the client beforehand
					sub, _ := syncSubscriptionsDiff(peerPo, -1, newDepth, s.kad.MaxProxDisplay, true)
					doPeerSubUpdate(p, sub, nil)
					if len(sub) == 0 {
						panic("w00t")
					}
					depth = newDepth
				} else {
					// peer was within depth, but depth has changed. we should request the cursors for the
					// necessary bins and quit the unnecessary ones
					sub, quits := syncSubscriptionsDiff(peerPo, depth, newDepth, s.kad.MaxProxDisplay, true)
					log.Debug("peer was inside depth, checking if needs changes", "peer", p.ID(), "peerPo", peerPo, "depth", depth, "newDepth", newDepth, "subs", sub, "quits", quits)
					doPeerSubUpdate(p, sub, quits)
					depth = newDepth
				}
			case peerPo < newDepth:
				if withinDepth {
					sub, quits := syncSubscriptionsDiff(peerPo, depth, newDepth, s.kad.MaxProxDisplay, true)
					log.Debug("peer transitioned out of depth", "peer", p.ID(), "subs", sub, "quits", quits)
					doPeerSubUpdate(p, sub, quits)
					withinDepth = false
				}
			}

		case <-s.quit:
			return
		}
	}
}

func doPeerSubUpdate(p *Peer, subs, quits []uint) {
	if len(subs) > 0 {
		log.Debug("getting cursors info from peer", "peer", p.ID(), "subs", subs)
		streamsMsg := StreamInfoReq{Streams: subs}
		if err := p.Send(context.TODO(), streamsMsg); err != nil {
			log.Error("error establishing subsequent subscription", "err", err)
			p.Drop()
		}
	}
	for _, v := range quits {
		log.Debug("removing cursor info for peer", "peer", p.ID(), "bin", v, "cursors", p.streamCursors, "quits", quits)
		delete(p.streamCursors, uint(v))

		if hs, ok := p.historicalStreams[uint(v)]; ok {
			log.Debug("closing historical stream for peer", "peer", p.ID(), "bin", v, "historicalStream", hs)

			close(hs.quit)
			// todo: wait for the hs.done to close?
			delete(p.historicalStreams, uint(v))
		} else {
			// this could happen when the cursor was 0 thus the historical stream was not created - do nothing
		}
	}
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
