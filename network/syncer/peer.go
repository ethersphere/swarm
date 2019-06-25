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
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
)

// Peer is the Peer extension for the streaming protocol
type Peer struct {
	*network.BzzPeer
	mtx          sync.Mutex
	streamsDirty bool // a request for StreamInfo is underway and awaiting reply
	syncer       *SwarmSyncer

	streamCursors     map[uint]uint64           // key: bin, value: session cursor. when unset - we are not interested in that bin
	historicalStreams map[uint]*syncStreamFetch //maintain state for each stream fetcher

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

	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}
	return nil
}

func (p *Peer) handleStreamInfoRes(ctx context.Context, msg *StreamInfoRes) {
	log.Debug("handleStreamInfoRes", "msg", msg)
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if len(msg.Streams) == 0 {
		log.Error("StreamInfo response is empty")
		p.Drop()
	}

	for _, s := range msg.Streams {
		stream := strings.Split(s.Name, "|")
		bin, err := strconv.Atoi(stream[1])
		if err != nil {
			log.Error("got an error parsing stream name", "descriptor", s)
			p.Drop()
		}
		log.Debug("setting bin cursor", "bin", uint(bin), "cursor", s.Cursor)
		p.streamCursors[uint(bin)] = s.Cursor
		if s.Cursor > 0 {
			streamFetch := newSyncStreamFetch(uint(bin))
			p.historicalStreams[uint(bin)] = streamFetch
		}
	}
}

func (p *Peer) handleStreamInfoReq(ctx context.Context, msg *StreamInfoReq) {
	log.Debug("handleStreamInfoReq", "msg", msg)
	p.mtx.Lock()
	defer p.mtx.Unlock()
	streamRes := StreamInfoRes{}

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
