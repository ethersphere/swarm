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
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/bitvector"
	bv "github.com/ethersphere/swarm/network/bitvector"
	"github.com/ethersphere/swarm/network/stream/intervals"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
)

var ErrEmptyBatch = errors.New("empty batch")

const (
	HashSize  = 32
	BatchSize = 16
)

// Peer is the Peer extension for the streaming protocol
type Peer struct {
	*network.BzzPeer
	mtx            sync.Mutex
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

func (p *Peer) Left() {
	close(p.quit)
}

func (p *Peer) InitProviders() {
	log.Debug("peer.InitProviders")

	for _, sp := range p.providers {
		if sp.StreamBehavior() != StreamIdle {
			go sp.RunUpdateStreams(p)
		}
	}
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
	case *ChunkDelivery:
		go p.handleChunkDelivery(ctx, msg)
	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}
	return nil
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
	to        uint64
	stream    ID
	hashes    map[string]bool
	bv        *bitvector.BitVector
	requested time.Time
	remaining uint64
	chunks    chan chunk.Chunk
	done      chan error
}

// handleStreamInfoReq handles the StreamInfoReq message.
// this message is handled by the SERVER (*Peer is the client in this case)
func (p *Peer) handleStreamInfoReq(ctx context.Context, msg *StreamInfoReq) {
	log.Debug("handleStreamInfoReq", "peer", p.ID(), "msg", msg)
	streamRes := StreamInfoRes{}
	if len(msg.Streams) == 0 {
		panic("nil streams msg requested")
	}
	for _, v := range msg.Streams {
		if provider, ok := p.providers[v.Name]; ok {
			key, err := provider.ParseKey(v.Key)
			if err != nil {
				// error parsing the stream key,
				log.Error("error parsing the stream key", "peer", p.ID(), "key", key)
				p.Drop()
			}
			streamCursor, err := provider.Cursor(key)
			if err != nil {
				log.Error("error getting cursor for stream key", "peer", p.ID(), "name", v.Name, "key", key, "err", err)
				panic("shouldnt happen")
			}
			descriptor := StreamDescriptor{
				Stream:  v,
				Cursor:  streamCursor,
				Bounded: provider.Boundedness(),
			}
			streamRes.Streams = append(streamRes.Streams, descriptor)
		} else {
			// tell the other peer we dont support this stream. this is non fatal
		}
	}
	if err := p.Send(ctx, streamRes); err != nil {
		log.Error("failed to send StreamInfoRes to client", "requested keys", msg.Streams)
	}
}

// handleStreamInfoRes handles the StreamInfoRes message.
// this message is handled by the CLIENT (*Peer is the server in this case)
func (p *Peer) handleStreamInfoRes(ctx context.Context, msg *StreamInfoRes) {
	log.Debug("handleStreamInfoRes", "peer", p.ID(), "msg", msg)

	if len(msg.Streams) == 0 {
		log.Error("StreamInfo response is empty")
		panic("panic for now - this shouldnt happen") //p.Drop()
	}

	for _, s := range msg.Streams {
		if provider, ok := p.providers[s.Stream.Name]; ok {
			// check the stream integrity
			_, err := provider.ParseKey(s.Stream.Key)
			if err != nil {
				log.Error("error parsing stream", "stream", s.Stream)
				p.Drop()
			}
			log.Debug("setting stream cursor", "peer", p.ID(), "stream", s.Stream.String(), "cursor", s.Cursor)
			p.setCursor(s.Stream, s.Cursor)

			if provider.StreamBehavior() == StreamAutostart {
				if s.Cursor > 0 {
					log.Debug("got cursor > 0 for stream. requesting history", "stream", s.Stream.String(), "cursor", s.Cursor)
					stID := NewID(s.Stream.Name, s.Stream.Key)

					c := p.getCursor(s.Stream)
					if s.Cursor == 0 {
						panic("wtf")
					}
					// fetch everything from beginning till s.Cursor
					go func(stream ID, cursor uint64) {
						err := p.requestStreamRange(ctx, stID, c)
						if err != nil {
							log.Error("had an error sending initial GetRange for historical stream", "peer", p.ID(), "stream", s.Stream.String(), "err", err)
							p.Drop()
						}
					}(stID, c)
				}

				// handle stream unboundedness
				if !s.Bounded {
					// constantly fetch the head of the stream
				}
			}
		} else {
			log.Error("got a StreamInfoRes message for a provider which I dont support")
			panic("shouldn't happen, replace with p.Drop()")
		}
	}
}

func (p *Peer) requestStreamRange(ctx context.Context, stream ID, cursor uint64) error {
	log.Debug("peer.requestStreamRange", "peer", p.ID(), "stream", stream.String(), "cursor", cursor)
	if _, ok := p.providers[stream.Name]; ok {
		peerIntervalKey := p.peerStreamIntervalKey(stream)
		interval, err := p.getOrCreateInterval(peerIntervalKey)
		if err != nil {
			return err
		}
		from, to, empty := interval.Next(cursor)
		log.Debug("peer.requestStreamRange nextInterval", "peer", p.ID(), "stream", stream.String(), "cursor", cursor, "from", from, "to", to)
		if from > cursor || empty {
			log.Debug("peer.requestStreamRange stream finished", "peer", p.ID(), "stream", stream.String(), "cursor", cursor)
			// stream finished. quit
			return nil
		}

		if from == 0 {
			panic("no")
		}

		if to-from > BatchSize-1 {
			log.Debug("limiting TO to HistoricalStreamPageSize", "to", to, "new to", from+BatchSize)
			to = from + BatchSize - 1 //because the intervals are INCLUSIVE, it means we get also FROM and TO
		}

		g := GetRange{
			Ruid:      uint(rand.Uint32()),
			Stream:    stream,
			From:      from,
			To:        to,
			BatchSize: BatchSize,
			Roundtrip: true,
		}

		log.Debug("sending GetRange to peer", "peer", p.ID(), "ruid", g.Ruid, "stream", stream.String(), "cursor", cursor, "GetRange", g)

		if err := p.Send(ctx, g); err != nil {
			return err
		}

		w := &want{
			ruid:   g.Ruid,
			stream: g.Stream,
			from:   g.From,
			to:     g.To,

			hashes:    make(map[string]bool),
			requested: time.Now(),
		}
		p.mtx.Lock()
		p.openWants[w.ruid] = w
		p.mtx.Unlock()
		return nil
	} else {
		panic("wtf")
		//got a message for an unsupported provider
	}
	return nil
}

// handleGetRange is handled by the SERVER and sends in response an OfferedHashes message
// in the case that for the specific interval no chunks exist - the server sends an empty OfferedHashes
// message so that the client could seal the interval and request the next
func (p *Peer) handleGetRange(ctx context.Context, msg *GetRange) {
	log.Debug("peer.handleGetRange", "peer", p.ID(), "msg", msg)
	if provider, ok := p.providers[msg.Stream.Name]; ok {
		key, err := provider.ParseKey(msg.Stream.Key)
		if err != nil {
			log.Error("erroring parsing stream key", "err", err, "stream", msg.Stream.String())
			p.Drop()
		}
		log.Debug("peer.handleGetRange collecting batch", "from", msg.From, "to", msg.To)
		h, f, t, err := p.collectBatch(ctx, provider, key, msg.From, msg.To)
		if err != nil {
			log.Error("erroring getting batch for stream", "peer", p.ID(), "stream", msg.Stream, "err", err)
			s := fmt.Sprintf("erroring getting batch for stream. peer %s, stream %s, error %v", p.ID().String(), msg.Stream.String(), err)
			panic(s)
			//p.Drop()
		}
		log.Debug("collected hashes for requested range", "hashes", len(h)/HashSize, "msg", msg)
		o := offer{
			ruid:      msg.Ruid,
			stream:    msg.Stream,
			hashes:    h,
			requested: time.Now(),
		}

		p.mtx.Lock()
		p.openOffers[msg.Ruid] = o
		p.mtx.Unlock()

		offered := OfferedHashes{
			Ruid:      msg.Ruid,
			LastIndex: uint(t),
			Hashes:    h,
		}
		l := len(h) / HashSize
		//if
		log.Debug("server offering batch", "peer", p.ID(), "ruid", msg.Ruid, "requestFrom", msg.From, "From", f, "requestTo", msg.To, "hashes", h, "l", l)
		if err := p.Send(ctx, offered); err != nil {
			log.Error("erroring sending offered hashes", "peer", p.ID(), "ruid", msg.Ruid, "err", err)
		}
	} else {
		panic("wtf")
		// unsupported proto
	}
}

// handleOfferedHashes handles the OfferedHashes wire protocol message.
// this message is handled by the CLIENT.
func (p *Peer) handleOfferedHashes(ctx context.Context, msg *OfferedHashes) {
	log.Debug("peer.handleOfferedHashes", "peer", p.ID(), "msg.ruid", msg.Ruid, "msg", msg)
	hashes := msg.Hashes
	lenHashes := len(hashes)
	if lenHashes%HashSize != 0 {
		log.Error("error invalid hashes length", "len", lenHashes, "msg.ruid", msg.Ruid)
	}

	p.mtx.Lock()
	w, ok := p.openWants[msg.Ruid]
	p.mtx.Unlock()
	if !ok {
		log.Error("ruid not found, dropping peer")
		p.Drop()
	}

	provider, ok := p.providers[w.stream.Name]
	if !ok {
		log.Error("got offeredHashes for unsupported protocol, dropping peer", "peer", p.ID())
		p.Drop()
	}
	want, err := bv.New(lenHashes / HashSize)
	if err != nil {
		log.Error("error initiaising bitvector", "len", lenHashes/HashSize, "msg.ruid", msg.Ruid, "err", err)
		p.Drop()
	}

	var ctr uint64 = 0

	for i := 0; i < lenHashes; i += HashSize {
		hash := hashes[i : i+HashSize]
		log.Trace("checking offered hash", "ref", fmt.Sprintf("%x", hash))
		c := chunk.Address(hash)

		if _, wait := provider.NeedData(ctx, hash); wait != nil {
			ctr++
			w.hashes[c.Hex()] = true
			// set the bit, so create a request
			want.Set(i / HashSize)
			log.Trace("need data", "ref", fmt.Sprintf("%x", hash), "request", true)
		} else {
			w.hashes[c.Hex()] = false
		}
	}
	cc := make(chan chunk.Chunk)
	dc := make(chan error)

	atomic.AddUint64(&w.remaining, ctr)
	w.bv = want
	w.chunks = cc
	w.done = dc

	var wantedHashesMsg WantedHashes

	errc := p.sealBatch(provider, w)

	if ctr == 0 && lenHashes == 0 {
		wantedHashesMsg = WantedHashes{
			Ruid:      msg.Ruid,
			BitVector: []byte{},
		}
	} else {
		wantedHashesMsg = WantedHashes{
			Ruid:      msg.Ruid,
			BitVector: want.Bytes(),
		}
	}

	log.Debug("sending wanted hashes", "peer", p.ID(), "offered", lenHashes/HashSize, "want", ctr, "msg", wantedHashesMsg)
	if err := p.Send(ctx, wantedHashesMsg); err != nil {
		log.Error("error sending wanted hashes", "peer", p.ID(), "w", wantedHashesMsg)
		p.Drop()
	}

	p.mtx.Lock()
	p.openWants[msg.Ruid] = w
	p.mtx.Unlock()

	stream := w.stream
	peerIntervalKey := p.peerStreamIntervalKey(w.stream)
	select {
	case err := <-errc:
		if err != nil {
			log.Error("got an error while sealing batch", "peer", p.ID(), "from", w.from, "to", w.to, "err", err)
			p.Drop()
		}
		err = p.addInterval(w.from, w.to, peerIntervalKey)
		if err != nil {
			log.Error("error persisting interval", "peer", p.ID(), "peerIntervalKey", peerIntervalKey, "from", w.from, "to", w.to)
		}
		p.mtx.Lock()
		delete(p.openWants, msg.Ruid)
		p.mtx.Unlock()

		//TODO BATCH TIMEOUT?
	}

	f, t, empty, err := p.nextInterval(peerIntervalKey, p.getCursor(stream))
	if empty {
		log.Debug("range ended, quitting")
	}
	log.Debug("next interval", "f", f, "t", t, "err", err, "intervalsKey", peerIntervalKey, "w", w)
	if err := p.requestStreamRange(ctx, stream, p.getCursor(stream)); err != nil {
		log.Error("error requesting next interval from peer", "peer", p.ID(), "err", err)
		p.Drop()
	}
}

func (p *Peer) addInterval(start, end uint64, peerStreamKey string) (err error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	i := &intervals.Intervals{}
	if err = p.intervalsStore.Get(peerStreamKey, i); err != nil {
		return err
	}
	i.Add(start, end)
	return p.intervalsStore.Put(peerStreamKey, i)
}

func (p *Peer) nextInterval(peerStreamKey string, ceil uint64) (start, end uint64, empty bool, err error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	i := &intervals.Intervals{}
	err = p.intervalsStore.Get(peerStreamKey, i)
	if err != nil {
		return 0, 0, false, err
	}
	start, end, empty = i.Next(ceil)
	return start, end, empty, nil
}

func (p *Peer) getOrCreateInterval(key string) (*intervals.Intervals, error) {
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

func (p *Peer) sealBatch(provider StreamProvider, w *want) <-chan error {
	log.Debug("peer.sealBatch", "ruid", w.ruid)
	errc := make(chan error)
	go func() {

		for {
			select {
			case c, ok := <-w.chunks:
				if !ok {
					log.Error("want chanks rreturned on !ok")
				}
				//p.mtx.Lock()
				//if wants, ok := want.hashes[c.Address().Hex()]; !ok || !wants {
				//log.Error("got an unwanted chunk from peer!", "peer", p.ID(), "caddr", c.Address)
				//}
				cc := chunk.NewChunk(c.Address(), c.Data())
				go func() {
					ctx := context.TODO()
					seen, err := provider.Put(ctx, cc.Address(), cc.Data())
					if err != nil {
						if err == storage.ErrChunkInvalid {
							p.Drop()
						}
					}
					if seen {
						log.Error("chunk already seen!", "peer", p.ID(), "caddr", c.Address()) //this is possible when the same chunk is asked from multiple peers
					}
					//want.hashes[c.Address().Hex()] = false //todo: should by sync map
					v := atomic.AddUint64(&w.remaining, ^uint64(0))
					//p.mtx.Unlock()
					if v == 0 {
						close(errc)
						return
					}
				}()
			case <-p.quit:
				return
			}
		}
	}()
	return errc
}

// handleWantedHashes is handled on the SERVER side and is dependent on a preceding OfferedHashes message
// the method is to ensure that all chunks in the requested batch is sent to the client
func (p *Peer) handleWantedHashes(ctx context.Context, msg *WantedHashes) {
	log.Debug("peer.handleWantedHashes", "peer", p.ID(), "ruid", msg.Ruid)
	// Get the length of the original offer from state
	// get the offered hashes themselves
	p.mtx.Lock()
	offer, ok := p.openOffers[msg.Ruid]
	p.mtx.Unlock()
	if !ok {
		log.Error("ruid does not exist. dropping peer", "ruid", msg.Ruid, "peer", p.ID())
		p.Drop()
	}

	provider, ok := p.providers[offer.stream.Name]
	if !ok {
		log.Error("no provider found for stream, dropping peer", "peer", p.ID(), "stream", offer.stream.String())
		p.Drop()
	}

	l := len(offer.hashes) / HashSize
	lll := len(msg.BitVector)
	log.Debug("bitvector", "l", lll, "h", offer.hashes)
	want, err := bv.NewFromBytes(msg.BitVector, l)
	if err != nil {
		log.Error("error initiaising bitvector", "l", l, "ll", len(offer.hashes), "err", err)
		panic("err")
	}
	log.Debug("iterate over wanted hashes", "l", len(offer.hashes))

	frameSize := 0
	const maxFrame = BatchSize
	cd := ChunkDelivery{
		Ruid:      msg.Ruid,
		LastIndex: 0,
	}

	for i := 0; i < l; i++ {
		if want.Get(i) {
			frameSize++

			metrics.GetOrRegisterCounter("peer.handlewantedhashesmsg.actualget", nil).Inc(1)

			hash := offer.hashes[i*HashSize : (i+1)*HashSize]
			data, err := provider.Get(ctx, hash)
			if err != nil {
				log.Error("handleWantedHashesMsg", "hash", hash, "err", err)
				p.Drop()
			}

			chunkD := DeliveredChunk{
				Addr: hash,
				Data: data,
			}
			//collect the chunk into the batch

			cd.Chunks = append(cd.Chunks, chunkD)
			if frameSize == maxFrame {
				//send the batch
				go func(cd ChunkDelivery) {
					log.Debug("sending chunk delivery")
					if err := p.Send(ctx, cd); err != nil {
						log.Error("error sending chunk delivery frame", "peer", p.ID(), "ruid", msg.Ruid, "error", err)
					}
				}(cd)
				frameSize = 0
				cd = ChunkDelivery{
					Ruid:      msg.Ruid,
					LastIndex: 0,
				}
			}
		}
	}

	// send anything that we might have left in the batch
	if frameSize > 0 {
		if err := p.Send(ctx, cd); err != nil {
			log.Error("error sending chunk delivery frame", "peer", p.ID(), "ruid", msg.Ruid, "error", err)
		}
	}
}

func (p *Peer) handleChunkDelivery(ctx context.Context, msg *ChunkDelivery) {
	log.Debug("peer.handleChunkDelivery", "peer", p.ID(), "chunks", len(msg.Chunks))

	p.mtx.Lock()
	w, ok := p.openWants[msg.Ruid]
	p.mtx.Unlock()
	if !ok {
		log.Error("no open offers for for ruid", "peer", p.ID(), "ruid", msg.Ruid)
		panic("should not happen")
	}
	if len(msg.Chunks) == 0 {
		log.Error("no chunks in msg!", "peer", p.ID(), "ruid", msg.Ruid)
		panic("should not happen")
	}
	log.Debug("delivering chunks for peer", "peer", p.ID(), "chunks", len(msg.Chunks))
	for _, dc := range msg.Chunks {
		c := chunk.NewChunk(dc.Addr, dc.Data)
		log.Debug("writing chunk to chunks channel", "peer", p.ID(), "caddr", c.Address())
		w.chunks <- c
	}
	log.Debug("done writing batch to chunks channel", "peer", p.ID())
}

func (p *Peer) collectBatch(ctx context.Context, provider StreamProvider, key interface{}, from, to uint64) (hashes []byte, f, t uint64, err error) {
	log.Debug("collectBatch", "peer", p.ID(), "from", from, "to", to)
	batchStart := time.Now()

	descriptors, stop := provider.Subscribe(ctx, key, from, to)
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
			log.Debug("got a chunk on key", "key", key)
			batch = append(batch, d.Address[:]...)
			batchSize++
			if batchStartID == nil {
				// set batch start id only if
				// this is the first iteration
				batchStartID = &d.BinID
			}
			log.Debug("got bin id", "id", d.BinID)
			batchEndID = d.BinID
			if batchSize >= BatchSize {
				iterate = false
				metrics.GetOrRegisterCounter("syncer.set-next-batch.full-batch", nil).Inc(1)
				log.Trace("syncer pull subscription - batch size reached", "batchSize", batchSize, "batchStartID", *batchStartID, "batchEndID", batchEndID)
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
			log.Trace("syncer pull subscription timer expired", "peer", p.ID(), "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
		case <-p.quit:
			iterate = false
			log.Trace("syncer pull subscription - quit received", "peer", p.ID(), "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
		}
	}
	if batchStartID == nil {
		// if batch start id is not set, it means we timed out
		return nil, 0, 0, ErrEmptyBatch
	}
	return batch, *batchStartID, batchEndID, nil

}

func (p *Peer) peerStreamIntervalKey(stream ID) string {
	k := fmt.Sprintf("%s|%s", p.ID().String(), stream.String())
	return k
}
