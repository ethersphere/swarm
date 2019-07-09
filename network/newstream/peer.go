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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethersphere/swarm/chunk"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/bitvector"
	"github.com/ethersphere/swarm/network/stream/intervals"
	"github.com/ethersphere/swarm/state"
)

var ErrEmptyBatch = errors.New("empty batch")

const (
	HashSize  = 32
	BatchSize = 16
)

// Peer is the Peer extension for the streaming protocol
type Peer struct {
	*network.BzzPeer
	mtx            sync.RWMutex
	providers      map[string]StreamProvider
	intervalsStore state.Store

	streamCursorsMu   sync.Mutex
	streamCursors     map[string]uint64 // key: Stream ID string representation, value: session cursor. Keeps cursors for all streams. when unset - we are not interested in that bin
	dirtyStreams      map[string]bool   // key: stream ID, value: whether cursors for a stream should be updated
	activeBoundedGets map[string]chan struct{}
	openWants         map[uint]*want // maintain open wants on the client side
	openOffers        map[uint]offer // maintain open offers on the server side
	quit              chan struct{}  // closed when peer is going offline
}

// NewPeer is the constructor for Peer
func NewPeer(peer *network.BzzPeer, i state.Store, providers map[string]StreamProvider) *Peer {
	p := &Peer{
		BzzPeer:        peer,
		providers:      providers,
		intervalsStore: i,
		streamCursors:  make(map[string]uint64),
		dirtyStreams:   make(map[string]bool),
		openWants:      make(map[uint]*want),
		openOffers:     make(map[uint]offer),
		quit:           make(chan struct{}),
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

func (p *Peer) getCursor(stream ID) uint64 {
	p.streamCursorsMu.Lock()
	defer p.streamCursorsMu.Unlock()

	return p.streamCursors[stream.String()]
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

func (p *Peer) InitProviders() {
	log.Debug("peer.InitProviders")

	for _, sp := range p.providers {
		if sp.StreamBehavior() != StreamIdle {
			go sp.RunUpdateStreams(p)
		}
	}
}

type offer struct {
	ruid      uint
	stream    ID
	hashes    []byte
	requested time.Time
}

type want struct {
	ruid      uint
	from      uint64
	to        *uint64
	stream    ID
	hashes    map[string]bool
	bv        *bitvector.BitVector
	requested time.Time
	remaining uint64
	chunks    chan chunk.Chunk
	done      chan error
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
		log.Error("unknown error while getting interval for peer", "err", err)
		return nil, err
	}
	return i, nil
}

func (p *Peer) peerStreamIntervalKey(stream ID) string {
	k := fmt.Sprintf("%s|%s", p.ID().String(), stream.String())
	return k
}
