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
	BatchSize = 50
	//DeliveryFrameSize        = 128
)

type offer struct {
	Ruid      uint
	stream    ID
	Hashes    []byte
	Requested time.Time
}

type Want struct {
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

// Peer is the Peer extension for the streaming protocol
type Peer struct {
	*network.BzzPeer
	mtx            sync.Mutex
	providers      map[string]StreamProvider
	intervalsStore state.Store

	streamCursors map[string]uint64 // key: Stream ID string representation, value: session cursor. Keeps cursors for all streams. when unset - we are not interested in that bin
	openWants     map[uint]*Want    // maintain open wants on the client side
	openOffers    map[uint]offer    // maintain open offers on the server side
	quit          chan struct{}     // closed when peer is going offline
}

// NewPeer is the constructor for Peer
func NewPeer(peer *network.BzzPeer, i state.Store, providers map[string]StreamProvider) *Peer {
	p := &Peer{
		BzzPeer:        peer,
		providers:      providers,
		intervalsStore: i,
		streamCursors:  make(map[string]uint64),
		openWants:      make(map[uint]*Want),
		openOffers:     make(map[uint]offer),
		quit:           make(chan struct{}),
	}
	return p
}

func (p *Peer) Left() {
	close(p.quit)
}

func (p *Peer) InitProviders() {
	log.Debug("peer.InitProviders")

	for _, sp := range p.providers {

		go sp.RunUpdateStreams(p)
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
	p.mtx.Lock()
	defer p.mtx.Unlock()

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
				panic("w00t")
				p.Drop()
			}
			log.Debug("setting stream cursor", "peer", p.ID(), "stream", s.Stream.String(), "cursor", s.Cursor)
			p.streamCursors[s.Stream.String()] = s.Cursor

			if s.Cursor > 0 {
				// fetch everything from beginning till  s.Cursor
				go func(stream ID, cursor uint64) {
					err := p.requestStreamRange(ctx, s.Stream, cursor)
					if err != nil {
						log.Error("had an error sending initial GetRange for historical stream", "peer", p.ID(), "stream", s.Stream.String(), "err", err)
						p.Drop()
					}
				}(s.Stream, s.Cursor)
			}

			// handle stream unboundedness
			if !s.Bounded {
				// constantly fetch the head of the stream

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
		from, to := interval.Next(cursor)
		log.Debug("peer.requestStreamRange nextInterval", "peer", p.ID(), "stream", stream.String(), "cursor", cursor, "from", from, "to", to)
		if from > cursor {
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
		log.Debug("sending GetRange to peer", "peer", p.ID(), "stream", stream.String(), "cursor", cursor, "GetRange", g)

		if err := p.Send(ctx, g); err != nil {
			return err
		}

		w := &Want{
			ruid:   g.Ruid,
			stream: g.Stream,
			from:   g.From,
			to:     g.To,

			hashes:    make(map[string]bool),
			requested: time.Now(),
		}

		p.openWants[w.ruid] = w
		return nil
	} else {
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
		h, f, t, err := p.collectBatch(ctx, provider, key, msg.From, msg.To)
		if err != nil {
			log.Error("erroring getting batch for stream", "peer", p.ID(), "stream", msg.Stream, "err", err)
			//p.Drop()
		}

		o := offer{
			Ruid:      msg.Ruid,
			stream:    msg.Stream,
			Hashes:    h,
			Requested: time.Now(),
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
		log.Debug("server offering batch", "peer", p.ID(), "ruid", msg.Ruid, "requestFrom", msg.From, "From", f, "requestTo", msg.To, "hashes", h, "l", l)
		if err := p.Send(ctx, offered); err != nil {
			log.Error("erroring sending offered hashes", "peer", p.ID(), "ruid", msg.Ruid, "err", err)
		}
	} else {
		// unsupported proto
	}
}

// handleOfferedHashes handles the OfferedHashes wire protocol message.
// this message is handled by the CLIENT.
func (p *Peer) handleOfferedHashes(ctx context.Context, msg *OfferedHashes) {
	log.Debug("peer.handleOfferedHashes", "peer", p.ID(), "msg.ruid", msg.Ruid)

	hashes := msg.Hashes
	lenHashes := len(hashes)
	if lenHashes%HashSize != 0 {
		log.Error("error invalid hashes length", "len", lenHashes)
	}

	w, ok := p.openWants[msg.Ruid]
	if !ok {
		log.Error("ruid not found, dropping peer")
		panic("drop peer")
		p.Drop()
	}

	provider, ok := p.providers[w.stream.Name]
	if !ok {
		log.Error("got offeredHashes for unsupported protocol, dropping peer", "peer", p.ID())
		p.Drop()
	}

	want, err := bv.New(lenHashes / HashSize)
	if err != nil {
		log.Error("error initiaising bitvector", "len", lenHashes/HashSize, "err", err)
		panic("drop later")
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
			want.Set(i/HashSize, true)
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

	errc := p.sealBatch(provider, msg.Ruid)
	if len(w.hashes) == 0 {
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

	log.Debug("sending wanted hashes", "peer", p.ID(), "offered", lenHashes/HashSize, "want", ctr)
	if err := p.Send(ctx, wantedHashesMsg); err != nil {
		log.Error("error sending wanted hashes", "peer", p.ID(), "w", wantedHashesMsg)
		p.Drop()
	}

	p.openWants[msg.Ruid] = w
	log.Debug("open wants", "ow", p.openWants)
	stream := w.stream
	peerIntervalKey := p.peerStreamIntervalKey(w.stream)
	select {
	case err := <-errc:
		if err != nil {
			log.Error("got an error while sealing batch", "peer", p.ID(), "from", w.from, "to", w.to, "err", err)
			panic(err)
			p.Drop()
		}
		log.Debug("adding interval", "f", w.from, "t", w.to, "key", peerIntervalKey)
		err = p.addInterval(w.from, w.to, peerIntervalKey)
		if err != nil {
			log.Error("error persisting interval", "peer", p.ID(), "peerIntervalKey", peerIntervalKey, "from", w.from, "to", w.to)
		}
		p.mtx.Lock()
		defer p.mtx.Unlock()
		delete(p.openWants, msg.Ruid)

		log.Debug("batch done", "from", w.from, "to", w.to)
		//TODO BATCH TIMEOUT?
	}

	f, t, err := p.nextInterval(peerIntervalKey, p.streamCursors[stream.String()])
	log.Error("next interval", "f", f, "t", t, "err", err, "intervalsKey", peerIntervalKey)
	if err := p.requestStreamRange(ctx, stream, p.streamCursors[stream.String()]); err != nil {
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

func (p *Peer) nextInterval(peerStreamKey string, ceil uint64) (start, end uint64, err error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	i := &intervals.Intervals{}
	err = p.intervalsStore.Get(peerStreamKey, i)
	if err != nil {
		return 0, 0, err
	}
	start, end = i.Next(ceil)
	return start, end, nil
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

func (p *Peer) sealBatch(provider StreamProvider, ruid uint) <-chan error {
	want := p.openWants[ruid]
	errc := make(chan error)
	go func() {
		for {
			select {
			case c, ok := <-want.chunks:
				if !ok {
					log.Error("want chanks rreturned on !ok")
					panic("shouldnt happen")
				}
				//p.mtx.Lock()
				//if wants, ok := want.hashes[c.Address().Hex()]; !ok || !wants {
				//log.Error("got an unwanted chunk from peer!", "peer", p.ID(), "caddr", c.Address)
				//panic("shouldnt happen")
				//}
				go func() {
					ctx := context.TODO()
					seen, err := provider.Put(ctx, c.Address(), c.Data())
					if err != nil {
						if err == storage.ErrChunkInvalid {
							p.Drop()
						}
					}
					if seen {
						log.Error("chunk already seen!", "peer", p.ID(), "caddr", c.Address())
						//panic("shouldnt happen") // this in fact could happen...
					}
					//want.hashes[c.Address().Hex()] = false //todo: should by sync map
					atomic.AddUint64(&want.remaining, ^uint64(0))
					//p.mtx.Unlock()
					v := atomic.LoadUint64(&want.remaining)
					if v == 0 {
						log.Debug("batchdone")
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
	offer, ok := p.openOffers[msg.Ruid]
	if !ok {
		// ruid doesn't exist. error and drop peer
		log.Error("ruid does not exist. dropping peer", "ruid", msg.Ruid, "peer", p.ID())
		p.Drop()
	}

	provider, ok := p.providers[offer.stream.Name]
	if !ok {
		log.Error("no provider found for stream, dropping peer", "peer", p.ID(), "stream", offer.stream.String())
		p.Drop()
	}

	l := len(offer.Hashes) / HashSize
	want, err := bv.NewFromBytes(msg.BitVector, l)
	if err != nil {
		log.Error("error initiaising bitvector", l, err)
	}

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

			hash := offer.Hashes[i*HashSize : (i+1)*HashSize]
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

	w, ok := p.openWants[msg.Ruid]
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
			batch = append(batch, d.Address[:]...)
			// This is the most naive approach to label the chunk as synced
			// allowing it to be garbage collected. A proper way requires
			// validating that the chunk is successfully stored by the peer.
			//err := p.syncer.netStore.Set(context.Background(), chunk.ModeSetSync, d.Address)
			//if err != nil {
			//metrics.GetOrRegisterCounter("syncer.set-next-batch.set-sync-err", nil).Inc(1)
			////log.Debug("syncer pull subscription - err setting chunk as synced", "correlateId", s.correlateId, "err", err)
			//return nil, 0, 0, err
			//}
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
			//log.Trace("syncer pull subscription timer expired", "correlateId", s.correlateId, "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
		case <-p.quit:
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

func (p *Peer) peerStreamIntervalKey(stream ID) string {
	k := fmt.Sprintf("%s|%s", p.ID().String(), stream.String())
	return k
}
