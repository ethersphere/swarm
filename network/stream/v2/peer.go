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

package stream

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ethersphere/swarm/chunk"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/stream/intervals"
	"github.com/ethersphere/swarm/state"
)

// Peer is the Peer extension for the streaming protocol
type Peer struct {
	*network.BzzPeer
	mtx            sync.RWMutex
	providers      map[string]StreamProvider
	intervalsStore state.Store //move to stream

	logger log.Logger

	streamCursorsMu sync.Mutex
	streamCursors   map[string]uint64 // key: Stream ID string representation, value: session cursor. Keeps cursors for all streams. when unset - we are not interested in that bin
	openWants       map[uint]*want    // maintain open wants on the client side
	openOffers      map[uint]offer    // maintain open offers on the server side

	quit chan struct{} // closed when peer is going offline
}

// NewPeer is the constructor for Peer
func NewPeer(peer *network.BzzPeer, baseKey []byte, i state.Store, providers map[string]StreamProvider) *Peer {
	p := &Peer{
		BzzPeer:        peer,
		providers:      providers,
		intervalsStore: i,
		streamCursors:  make(map[string]uint64),
		openWants:      make(map[uint]*want),
		openOffers:     make(map[uint]offer),
		quit:           make(chan struct{}),
		logger:         log.New("base", hex.EncodeToString(baseKey)[:16], "peer", peer.ID().String()[:16]),
	}
	return p
}

func (p *Peer) cursorsCount() int {
	p.streamCursorsMu.Lock()
	defer p.streamCursorsMu.Unlock()

	return len(p.streamCursors)
}

func (p *Peer) getCursorsCopy() map[string]uint64 {
	p.streamCursorsMu.Lock()
	defer p.streamCursorsMu.Unlock()

	c := make(map[string]uint64, len(p.streamCursors))
	for k, v := range p.streamCursors {
		c[k] = v
	}
	return c
}

func (p *Peer) getCursor(stream ID) (uint64, bool) {
	p.streamCursorsMu.Lock()
	defer p.streamCursorsMu.Unlock()
	val, ok := p.streamCursors[stream.String()]
	return val, ok
}

func (p *Peer) setCursor(stream ID, cursor uint64) {
	p.streamCursorsMu.Lock()
	defer p.streamCursorsMu.Unlock()

	p.streamCursors[stream.String()] = cursor
}

func (p *Peer) deleteCursor(stream ID) {
	p.streamCursorsMu.Lock()
	defer p.streamCursorsMu.Unlock()

	delete(p.streamCursors, stream.String())
}

// InitProviders initializes a provider for a certain peer
func (p *Peer) InitProviders() {
	p.logger.Debug("peer.InitProviders")

	for _, sp := range p.providers {
		go sp.InitPeer(p)
	}
}

// offer represents an open offer from a server to a client as a result of a GetRange message
// it is stored for reference to requests on the peer.openOffers map
type offer struct {
	ruid      uint      // the request uid
	stream    ID        // the stream id
	hashes    []byte    // all hashes offered to the client
	requested time.Time // requested at time
}

// want represents an open want for a hash range from a client to a server
// it is stored on the peer.openWants
type want struct {
	ruid      uint                // the request uid
	from      uint64              // want from index
	to        *uint64             // want to index, nil signifies top of range not yet known
	head      bool                // is this the head of the stream? (bound versus tip of the stream; true is tip)
	stream    ID                  // the stream id
	hashes    map[string]struct{} // key: chunk address, value: wanted yes/no, used to prevent unsolicited chunks
	requested time.Time           // requested at time
	remaining uint64              // number of remaining chunks to deliver
	chunks    chan chunk.Address  // chunk arrived notification channel
	closeC    chan error          // signal polling goroutine to terminate due to empty batch or timeout
}

// getOfferOrDrop gets on open offer for the requested ruid
// in case the offer is not found - the peer is dropped
func (p *Peer) getOfferOrDrop(ruid uint) (o offer, shouldBreak bool) {
	p.mtx.RLock()
	o, ok := p.openOffers[ruid]
	p.mtx.RUnlock()
	if !ok {
		p.logger.Error("ruid not found, dropping peer", "ruid", ruid)
		p.Drop()
		return o, true
	}
	return o, false
}

// getWantOrDrop gets on open want for the requested ruid
// in case the want is not found - the peer is dropped
func (p *Peer) getWantOrDrop(ruid uint) (w *want, shouldBreak bool) {
	p.mtx.RLock()
	w, ok := p.openWants[ruid]
	p.mtx.RUnlock()
	if !ok {
		p.logger.Error("ruid not found, dropping peer", "ruid", ruid)
		p.Drop()
		return nil, true
	}
	return w, false
}

func (p *Peer) addInterval(stream ID, start, end uint64) (err error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	peerStreamKey := p.peerStreamIntervalKey(stream)
	i := &intervals.Intervals{}
	if err = p.intervalsStore.Get(peerStreamKey, i); err != nil {
		return err
	}
	i.Add(start, end)
	return p.intervalsStore.Put(peerStreamKey, i)
}

func (p *Peer) nextInterval(stream ID, ceil uint64) (start, end uint64, empty bool, err error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	i := &intervals.Intervals{}
	err = p.intervalsStore.Get(p.peerStreamIntervalKey(stream), i)
	if err != nil {
		return 0, 0, false, err
	}

	start, end, empty = i.Next(ceil)
	return start, end, empty, nil
}

func (p *Peer) sealWant(w *want) error {
	err := p.addInterval(w.stream, w.from, *w.to)
	if err != nil {
		return err
	}
	p.mtx.Lock()
	delete(p.openWants, w.ruid)
	p.mtx.Unlock()
	return nil
}

func (p *Peer) getOrCreateInterval(key string) (*intervals.Intervals, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// check that an interval entry exists
	i := &intervals.Intervals{}
	err := p.intervalsStore.Get(key, i)
	switch err {
	case nil:
	case state.ErrNotFound:
		// key interval values are ALWAYS > 0
		i = intervals.NewIntervals(1)
		if err := p.intervalsStore.Put(key, i); err != nil {
			return nil, err
		}
	default:
		p.logger.Error("unknown error while getting interval for peer", "err", err)
		return nil, err
	}
	return i, nil
}

func (p *Peer) peerStreamIntervalKey(stream ID) string {
	k := fmt.Sprintf("%s|%s", hex.EncodeToString(p.BzzAddr.OAddr), stream.String())
	return k
}
