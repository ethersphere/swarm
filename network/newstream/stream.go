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
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	bv "github.com/ethersphere/swarm/network/bitvector"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
)

var (
	// Compile time interface check
	_ node.Service = (*SlipStream)(nil)

	// Metrics
	processReceivedChunksCount = metrics.NewRegisteredCounter("network.stream.received_chunks.count", nil)
	streamSeenChunkDelivery    = metrics.NewRegisteredCounter("network.stream.seen_chunk_delivery.count", nil)
	streamEmptyWantedHashes    = metrics.NewRegisteredCounter("network.stream.empty_wanted_hashes.count", nil)
	streamWantedHashes         = metrics.NewRegisteredCounter("network.stream.wanted_hashes.count", nil)

	streamBatchFail               = metrics.NewRegisteredCounter("network.stream.batch_fail.count", nil)
	streamChunkDeliveryFail       = metrics.NewRegisteredCounter("network.stream.delivery_fail.count", nil)
	streamRequestNextIntervalFail = metrics.NewRegisteredCounter("network.stream.next_interval_fail.count", nil)
	lastReceivedChunksMsg         = metrics.GetOrRegisterGauge("network.stream.received_chunks", nil)

	// Protocol spec
	Spec = &protocols.Spec{
		Name:       "bzz-stream",
		Version:    8,
		MaxMsgSize: 10 * 1024 * 1024,
		Messages: []interface{}{
			StreamInfoReq{},
			StreamInfoRes{},
			GetRange{},
			OfferedHashes{},
			ChunkDelivery{},
			WantedHashes{},
		},
	}
)

// SlipStream is the base type that handles all client/server operations on a node
// it is instantiated once per stream protocol instance, that is, it should have
// one instance per node
type SlipStream struct {
	mtx            sync.RWMutex
	intervalsStore state.Store
	peers          map[enode.ID]*Peer
	baseKey        []byte

	providers map[string]StreamProvider

	spec *protocols.Spec

	handlersWg sync.WaitGroup // waits for all handlers to finish in Close method
	quit       chan struct{}

	logger log.Logger
}

func New(intervalsStore state.Store, baseKey []byte, providers ...StreamProvider) *SlipStream {
	slipStream := &SlipStream{
		intervalsStore: intervalsStore,
		peers:          make(map[enode.ID]*Peer),
		providers:      make(map[string]StreamProvider),
		quit:           make(chan struct{}),
		baseKey:        baseKey,
		logger:         log.New("base", hex.EncodeToString(baseKey)),
		spec:           Spec,
	}
	for _, p := range providers {
		slipStream.providers[p.StreamName()] = p
	}

	return slipStream
}

func (s *SlipStream) getProvider(stream ID) StreamProvider {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.providers[stream.Name]
}

func (s *SlipStream) getPeer(id enode.ID) *Peer {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	p := s.peers[id]
	return p
}

func (s *SlipStream) addPeer(p *Peer) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.peers[p.ID()] = p
}

func (s *SlipStream) removePeer(p *Peer) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, found := s.peers[p.ID()]; found {
		p.logger.Error("removing peer")
		delete(s.peers, p.ID())
		close(p.quit)
	} else {
		p.logger.Warn("peer was marked for removal but not found")
	}
}

// Run is being dispatched when 2 nodes connect
func (s *SlipStream) Run(bp *network.BzzPeer) error {
	sp := NewPeer(bp, s.baseKey, s.intervalsStore, s.providers)
	s.addPeer(sp)
	defer s.removePeer(sp)

	go sp.InitProviders()

	return sp.Peer.Run(s.HandleMsg(sp))
}

func (s *SlipStream) HandleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		s.mtx.Lock() // ensure that quit read and handlersWg add are locked together
		defer s.mtx.Unlock()

		select {
		case <-s.quit:
			// no message handling if we quit
			return nil
		default:
		}

		s.handlersWg.Add(1)
		go func() {
			defer s.handlersWg.Done()

			switch msg := msg.(type) {
			case *StreamInfoReq:
				s.handleStreamInfoReq(ctx, p, msg)
			case *StreamInfoRes:
				s.handleStreamInfoRes(ctx, p, msg)
			case *GetRange:
				if msg.To == nil {
					// handle live
					s.handleGetRangeHead(ctx, p, msg)
				} else {
					s.handleGetRange(ctx, p, msg)
				}
			case *OfferedHashes:
				s.handleOfferedHashes(ctx, p, msg)
			case *WantedHashes:
				s.handleWantedHashes(ctx, p, msg)
			case *ChunkDelivery:
				s.handleChunkDelivery(ctx, p, msg)
			}
		}()
		return nil
	}
}

// handleStreamInfoReq handles the StreamInfoReq message.
// this message is handled by the SERVER (*Peer is the client in this case)
func (s *SlipStream) handleStreamInfoReq(ctx context.Context, p *Peer, msg *StreamInfoReq) {
	p.logger.Debug("handleStreamInfoReq")
	streamRes := StreamInfoRes{}
	if len(msg.Streams) == 0 {
		p.logger.Error("nil streams msg requested")
		p.Drop()
		return
	}
	for _, v := range msg.Streams {
		v := v
		provider := s.getProvider(v)
		if provider == nil {
			p.logger.Error("unsupported provider", "stream", v)
			// tell the other peer we dont support this stream. this is non fatal
			// this might not be fatal as we might not support all providers.
			return
		}

		streamCursor, err := provider.CursorStr(v.Key)
		if err != nil {
			p.logger.Error("error getting cursor for stream key", "name", v.Name, "key", v.Key, "err", err)
			panic(fmt.Errorf("provider cursor str %q: %v", v.Key, err))
			p.Drop()
			return
		}
		descriptor := StreamDescriptor{
			Stream:  v,
			Cursor:  streamCursor,
			Bounded: provider.Boundedness(),
		}
		streamRes.Streams = append(streamRes.Streams, descriptor)
	}

	select {
	case <-s.quit:
		return
	default:
	}

	if err := p.Send(ctx, streamRes); err != nil {
		p.logger.Error("failed to send StreamInfoRes to client", "err", err)
	}
}

// TODO: provide this option value from StreamProvider?
var streamAutostart = true

// handleStreamInfoRes handles the StreamInfoRes message.
// this message is handled by the CLIENT (*Peer is the server in this case)
func (st *SlipStream) handleStreamInfoRes(ctx context.Context, p *Peer, msg *StreamInfoRes) {
	p.logger.Debug("handleStreamInfoRes")

	if len(msg.Streams) == 0 {
		p.logger.Error("StreamInfo response is empty")
		p.Drop()
		return
	}

	for _, s := range msg.Streams {
		s := s
		provider := st.getProvider(s.Stream)
		if provider == nil {
			// at this point of the message exchange unsupported providers are illegal. drop peer
			p.logger.Error("unsupported provider", "stream", s.Stream)
			p.Drop()
			return
		}

		if !provider.WantStream(p, s.Stream) {
			if _, exists := p.getCursor(s.Stream); exists {
				p.logger.Debug("stream cursor exists but we don't want it - removing", "stream", s.Stream)
				p.deleteCursor(s.Stream)
			}
			continue
		}

		if _, exists := p.getCursor(s.Stream); exists {
			p.logger.Debug("stream cursor already exists, continue to next", "stream", s.Stream)
			continue
		}

		p.logger.Debug("setting stream cursor", "stream", s.Stream, "cursor", s.Cursor)
		p.setCursor(s.Stream, s.Cursor)

		if streamAutostart {
			if s.Cursor > 0 {
				p.logger.Debug("requesting history stream", "stream", s.Stream, "cursor", s.Cursor)

				// fetch everything from beginning till s.Cursor
				go func() {
					err := st.requestStreamRange(ctx, p, s.Stream, s.Cursor)
					if err != nil {
						p.logger.Error("had an error sending initial GetRange for historical stream", "stream", s.Stream, "err", err)
						p.Drop()
					}
				}()
			}

			// handle stream unboundedness
			if !s.Bounded {
				//constantly fetch the head of the stream
				go func() {
					p.logger.Debug("asking for live stream", "stream", s.Stream, "cursor", s.Cursor)

					// ask the tip (cursor + 1)
					err := st.requestStreamHead(ctx, p, s.Stream, s.Cursor+1)
					// https://github.com/golang/go/issues/4373 - use of closed network connection
					if err != nil && err != p2p.ErrShuttingDown && !strings.Contains(err.Error(), "use of closed network connection") {
						p.logger.Error("had an error with initial stream head fetch", "stream", s.Stream, "cursor", s.Cursor+1, "err", err)
						p.Drop()
					}
				}()
			}
		}
	}
}

func (s *SlipStream) createSendWant(ctx context.Context, p *Peer, stream ID, from uint64, to *uint64, head bool) error {
	g := GetRange{
		Ruid:      uint(rand.Uint32()),
		Stream:    stream,
		From:      from,
		To:        to,
		BatchSize: BatchSize,
		Roundtrip: true,
	}

	p.logger.Debug("sending GetRange to peer", "ruid", g.Ruid, "stream", stream)

	w := &want{
		ruid:      g.Ruid,
		stream:    g.Stream,
		from:      g.From,
		to:        to,
		head:      head,
		hashes:    make(map[string]bool),
		requested: time.Now(),
	}
	p.mtx.Lock()
	p.openWants[w.ruid] = w
	p.mtx.Unlock()

	return p.Send(ctx, g)
}

func (s *SlipStream) requestStreamHead(ctx context.Context, p *Peer, stream ID, from uint64) error {
	p.logger.Debug("peer.requestStreamHead", "stream", stream, "from", from)
	return s.createSendWant(ctx, p, stream, from, nil, true)
}

func (s *SlipStream) requestStreamRange(ctx context.Context, p *Peer, stream ID, cursor uint64) error {
	p.logger.Debug("peer.requestStreamRange", "stream", stream, "cursor", cursor)
	provider := s.getProvider(stream)
	if provider == nil {
		// at this point of the message exchange unsupported providers are illegal. drop peer
		p.logger.Error("unsupported provider", "stream", stream)
		p.Drop()
		return nil
	}
	from, _, empty, err := p.nextInterval(stream, 0)
	if err != nil {
		return err
	}
	p.logger.Debug("peer.requestStreamRange nextInterval", "stream", stream, "cursor", cursor, "from", from)
	if from > cursor || empty {
		p.logger.Debug("peer.requestStreamRange stream finished", "stream", stream, "cursor", cursor)
		// stream finished. quit
		return nil
	}
	return s.createSendWant(ctx, p, stream, from, &cursor, false)
}

func (s *SlipStream) handleGetRangeHead(ctx context.Context, p *Peer, msg *GetRange) {
	p.logger.Debug("peer.handleGetRangeHead", "ruid", msg.Ruid)
	provider := s.getProvider(msg.Stream)
	if provider == nil {
		// at this point of the message exchange unsupported providers are illegal. drop peer
		p.logger.Error("unsupported provider", "stream", msg.Stream)
		p.Drop()
		return
	}

	key, err := provider.ParseKey(msg.Stream.Key)
	if err != nil {
		p.logger.Error("erroring parsing stream key", "stream", msg.Stream, "err", err)
		p.Drop()
		return
	}
	h, f, t, e, err := s.serverCollectBatch(ctx, p, provider, key, msg.From, 0)
	p.logger.Debug("peer.serverCollectBatch", "stream", msg.Stream, "len(h)", len(h), "f", f, "t", t, "e", e, "err", err, "ruid", msg.Ruid, "msg.from", msg.From)
	if err != nil {
		p.logger.Error("erroring getting live batch for stream", "stream", msg.Stream, "err", err)
		p.Drop()
		return
	}

	if e {
		select {
		case <-p.quit:
			p.logger.Debug("not sending batch due to shutdown")
			// prevent sending an empty batch that resulted from db shutdown
			return
		default:
			offered := OfferedHashes{
				Ruid:      msg.Ruid,
				LastIndex: msg.From,
				Hashes:    []byte{},
			}
			if err := p.Send(ctx, offered); err != nil {
				p.logger.Error("erroring sending empty live offered hashes", "ruid", msg.Ruid, "err", err)
			}
			return
		}
	}

	p.mtx.Lock()
	p.openOffers[msg.Ruid] = offer{
		ruid:      msg.Ruid,
		stream:    msg.Stream,
		hashes:    h,
		requested: time.Now(),
	}
	p.mtx.Unlock()

	offered := OfferedHashes{
		Ruid:      msg.Ruid,
		LastIndex: t,
		Hashes:    h,
	}
	i := len(h) / HashSize
	p.logger.Debug("server offering batch", "ruid", msg.Ruid, "requestfrom", msg.From, "requestto", msg.To, "hashes", i)
	if err := p.Send(ctx, offered); err != nil {
		p.logger.Error("erroring sending offered hashes", "ruid", msg.Ruid, "err", err)
		p.mtx.Lock()
		delete(p.openOffers, msg.Ruid)
		p.mtx.Unlock()
	}
}

// handleGetRange is handled by the SERVER and sends in response an OfferedHashes message
// in the case that for the specific interval no chunks exist - the server sends an empty OfferedHashes
// message so that the client could seal the interval and request the next
func (s *SlipStream) handleGetRange(ctx context.Context, p *Peer, msg *GetRange) {
	p.logger.Debug("peer.handleGetRange", "ruid", msg.Ruid)
	provider := s.getProvider(msg.Stream)
	if provider == nil {
		// at this point of the message exchange unsupported providers are illegal. drop peer
		p.logger.Error("unsupported provider", "stream", msg.Stream)
		p.Drop()
		return
	}

	key, err := provider.ParseKey(msg.Stream.Key)
	if err != nil {
		p.logger.Error("erroring parsing stream key", "err", err, "stream", msg.Stream)
		p.Drop()
		return
	}
	log.Debug("peer.handleGetRange collecting batch", "from", msg.From, "to", msg.To, "stream", msg.Stream)
	h, f, t, e, err := s.serverCollectBatch(ctx, p, provider, key, msg.From, *msg.To)
	// empty batch can be legit, TODO: check which errors should be handled, if any
	if err != nil {
		log.Error("erroring getting batch for stream", "peer", p.ID(), "stream", msg.Stream, "err", err)
		panic("for now")
		p.Drop()
		return
	}
	if e {
		p.logger.Debug("interval is empty for requested range", "empty?", e, "hashes", len(h)/HashSize, "ruid", msg.Ruid)
		select {
		case <-p.quit:
			// prevent sending an empty batch that resulted from db shutdown
			return
		default:
			offered := OfferedHashes{
				Ruid:      msg.Ruid,
				LastIndex: msg.From,
				Hashes:    []byte{},
			}
			if err := p.Send(ctx, offered); err != nil {
				p.logger.Error("erroring sending empty offered hashes", "ruid", msg.Ruid, "err", err)
			}
			return
		}
	}
	p.logger.Debug("collected hashes for requested range", "hashes", len(h)/HashSize, "ruid", msg.Ruid)
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
		LastIndex: t,
		Hashes:    h,
	}
	l := len(h) / HashSize
	p.logger.Debug("server offering batch", "ruid", msg.Ruid, "requestFrom", msg.From, "From", f, "requestTo", msg.To, "hashes", l)
	if err := p.Send(ctx, offered); err != nil {
		p.logger.Error("erroring sending offered hashes", "ruid", msg.Ruid, "err", err)
	}
}

// handleOfferedHashes handles the OfferedHashes wire protocol message.
// this message is handled by the CLIENT.
func (s *SlipStream) handleOfferedHashes(ctx context.Context, p *Peer, msg *OfferedHashes) {
	p.logger.Debug("stream.handleOfferedHashes", "ruid", msg.Ruid, "msg.lastIndex", msg.LastIndex)
	hashes := msg.Hashes
	lenHashes := len(hashes)
	if lenHashes%HashSize != 0 {
		p.logger.Error("invalid hashes length", "len", lenHashes, "ruid", msg.Ruid)
		p.Drop()
		return
	}

	p.mtx.RLock()
	w, ok := p.openWants[msg.Ruid]
	p.mtx.RUnlock()
	if !ok {
		p.logger.Error("ruid not found, dropping peer")
		p.Drop()
		return
	}
	provider := s.getProvider(w.stream)
	if provider == nil {
		// at this point of the message exchange unsupported providers are illegal. drop peer
		p.logger.Error("unsupported provider", "stream", w.stream)
		p.Drop()
		return
	}

	w.to = &msg.LastIndex

	// this code block handles the case of a gap on the interval on the server side
	// lenhashes == 0 means there's no hashes in the requested range with the upper bound of
	// the LastIndex on the incoming message
	if lenHashes == 0 {
		p.logger.Debug("handling empty offered hashes - sealing empty interval", "ruid", w.ruid)
		if err := p.sealWant(w); err != nil {
			p.logger.Error("error persisting interval", "from", w.from, "to", w.to, "err", err)
			p.Drop()
			return
		}
		cur, ok := p.getCursor(w.stream)
		if !ok {
			metrics.NewRegisteredCounter("network.stream.quit_unwanted.count", nil).Inc(1)
			p.logger.Debug("no longer interested in stream. quitting", "stream", w.stream)
			return
		}
		if w.head {
			if err := s.requestStreamHead(ctx, p, w.stream, msg.LastIndex+1); err != nil {
				streamRequestNextIntervalFail.Inc(1)
				p.logger.Error("error requesting next interval from peer", "err", err)
				p.Drop()
				return
			}
		} else {
			if err := s.requestStreamRange(ctx, p, w.stream, cur); err != nil {
				streamRequestNextIntervalFail.Inc(1)
				p.logger.Error("error requesting next interval from peer", "err", err)
				p.Drop()
				return
			}
		}
		return
	}

	want, err := bv.New(lenHashes / HashSize)
	if err != nil {
		p.logger.Error("error initiaising bitvector", "len", lenHashes/HashSize, "ruid", msg.Ruid, "err", err)
		p.Drop()
		return
	}

	var ctr uint64 = 0

	for i := 0; i < lenHashes; i += HashSize {
		hash := hashes[i : i+HashSize]
		p.logger.Trace("peer offered hash", "ref", fmt.Sprintf("%x", hash), "ruid", msg.Ruid)
		c := chunk.Address(hash)

		if _, wait := provider.NeedData(ctx, hash); wait != nil {
			ctr++
			w.hashes[c.Hex()] = true
			// set the bit, so create a request
			want.Set(i / HashSize)
			p.logger.Trace("need data", "need", "true", "ref", fmt.Sprintf("%x", hash), "ruid", msg.Ruid)
		} else {
			p.logger.Trace("dont need data", "need", "false", "ref", fmt.Sprintf("%x", hash), "ruid", msg.Ruid)
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

	errc := s.clientSealBatch(p, provider, w)

	if ctr == 0 {
		// this handles the case that there are no hashes we are interested in (ctr==0)
		// but some hashes were received by the server. the closed channel will result in
		// clientSealBatch goroutine in returning, then in the following select case below
		// the w.done channel is selected, in turn sealing the interval we are not interested in
		// then requesting the next batch
		p.logger.Debug("sending empty wanted hashes", "ruid", msg.Ruid)
		streamEmptyWantedHashes.Inc(1)
		wantedHashesMsg = WantedHashes{
			Ruid:      msg.Ruid,
			BitVector: []byte{},
		}
		close(w.done)
	} else {
		// there are some hashes in the offer and we want some
		p.logger.Debug("sending non-empty wanted hashes", "ruid", msg.Ruid, "len(bv)", len(want.Bytes()))
		streamWantedHashes.Inc(1)
		wantedHashesMsg = WantedHashes{
			Ruid:      msg.Ruid,
			BitVector: want.Bytes(),
		}
	}

	p.logger.Debug("sending wanted hashes", "offered", lenHashes/HashSize, "want", ctr, "ruid", msg.Ruid)
	if err := p.Send(ctx, wantedHashesMsg); err != nil {
		p.logger.Error("error sending wanted hashes", "err", err)
		p.Drop()
		return
	}

	select {
	case err := <-errc:
		if err != nil {
			streamBatchFail.Inc(1)
			p.logger.Error("got an error while sealing batch", "from", w.from, "to", w.to, "err", err)
			p.Drop()
			return
		}
		if err := p.sealWant(w); err != nil {
			p.logger.Error("error persisting interval", "from", w.from, "to", w.to, "err", err)
			p.Drop()
			return
		}
	case <-time.After(5 * time.Second):
		log.Error("batch has timed out", "ruid", w.ruid)
		close(w.done)
		p.mtx.Lock()
		delete(p.openWants, msg.Ruid)
		p.mtx.Unlock()
	case <-w.done:
		p.logger.Debug("batch empty, sealing interval", "ruid", w.ruid)
		if err := p.sealWant(w); err != nil {
			p.logger.Error("error persisting interval", "from", w.from, "to", w.to, "err", err)
			p.Drop()
			return
		}
	case <-s.quit:
		return
	}
	cur, ok := p.getCursor(w.stream)
	if !ok {
		metrics.NewRegisteredCounter("network.stream.quit_unwanted.count", nil).Inc(1)
		p.logger.Debug("no longer interested in stream. quitting", "stream", w.stream)
		return
	}
	p.logger.Debug("batch finished, requesting next", "ruid", w.ruid, "stream", w.stream)
	if w.head {
		if err := s.requestStreamHead(ctx, p, w.stream, msg.LastIndex+1); err != nil {
			streamRequestNextIntervalFail.Inc(1)
			p.logger.Error("error requesting next interval from peer", "err", err)
			p.Drop()
			return
		}
	} else {
		if err := s.requestStreamRange(ctx, p, w.stream, cur); err != nil {
			streamRequestNextIntervalFail.Inc(1)
			p.logger.Error("error requesting next interval from peer", "err", err)
			p.Drop()
			return
		}
	}
}

// handleWantedHashes is handled on the SERVER side and is dependent on a preceding OfferedHashes message
// the method is to ensure that all chunks in the requested batch is sent to the client
func (s *SlipStream) handleWantedHashes(ctx context.Context, p *Peer, msg *WantedHashes) {
	p.logger.Debug("peer.handleWantedHashes", "ruid", msg.Ruid, "bv", msg.BitVector)

	p.mtx.RLock()
	offer, ok := p.openOffers[msg.Ruid]
	p.mtx.RUnlock()
	if !ok {
		p.logger.Error("ruid does not exist. dropping peer", "ruid", msg.Ruid)
		p.Drop()
		return
	}

	provider, ok := p.providers[offer.stream.Name]
	if !ok {
		p.logger.Error("no provider found for stream, dropping peer", "stream", offer.stream)
		p.Drop()
		return
	}

	l := len(offer.hashes) / HashSize
	if len(msg.BitVector) == 0 {
		p.logger.Debug("peer does not want any hashes in this range", "ruid", offer.ruid)
		p.mtx.Lock()
		delete(p.openOffers, msg.Ruid)
		p.mtx.Unlock()
		return
	}
	want, err := bv.NewFromBytes(msg.BitVector, l)
	if err != nil {
		p.logger.Error("error initiaising bitvector", "l", l, "ll", len(offer.hashes), "err", err)
		p.Drop()
		return
	}

	frameSize := 0
	const maxFrame = BatchSize
	cd := &ChunkDelivery{
		Ruid: msg.Ruid,
	}
	for i := 0; i < l; i++ {
		p.logger.Trace("peer wants hash?", "ruid", offer.ruid, "wants?", want.Get(i))
		if want.Get(i) {
			metrics.GetOrRegisterCounter("peer.handlewantedhashesmsg.actualget", nil).Inc(1)
			hash := offer.hashes[i*HashSize : (i+1)*HashSize]
			data, err := provider.Get(ctx, hash)
			if err != nil {
				p.logger.Error("handleWantedHashesMsg", "hash", hash, "err", err)
				p.Drop()
				return
			}

			chunkD := DeliveredChunk{
				Addr: hash,
				Data: data,
			}

			//collect the chunk into the batch
			frameSize++

			cd.Chunks = append(cd.Chunks, chunkD)
			if frameSize == maxFrame {
				//send the batch
				go func(cd *ChunkDelivery) {
					p.logger.Debug("sending chunk delivery")
					if err := p.Send(ctx, cd); err != nil {
						p.logger.Error("error sending chunk delivery frame", "ruid", msg.Ruid, "error", err)
						p.Drop()
					}
				}(cd)
				frameSize = 0
				cd = &ChunkDelivery{
					Ruid: msg.Ruid,
				}
			}
		}
	}

	// send anything that we might have left in the batch
	if frameSize > 0 {
		if err := p.Send(ctx, cd); err != nil {
			p.logger.Error("error sending chunk delivery frame", "ruid", msg.Ruid, "error", err)
			p.Drop()
		}
	}
}

func (s *SlipStream) handleChunkDelivery(ctx context.Context, p *Peer, msg *ChunkDelivery) {
	p.logger.Debug("peer.handleChunkDelivery", "ruid", msg.Ruid, "chunks", len(msg.Chunks))
	processReceivedChunksCount.Inc(1)
	lastReceivedChunksMsg.Update(time.Now().UnixNano())

	p.mtx.RLock()
	w, ok := p.openWants[msg.Ruid]
	p.mtx.RUnlock()
	if !ok {
		streamChunkDeliveryFail.Inc(1)
		p.logger.Error("no open offers for for ruid", "ruid", msg.Ruid)
		p.Drop()
		return
	}

	p.logger.Debug("delivering chunks for peer", "chunks", len(msg.Chunks))
	for _, dc := range msg.Chunks {
		c := chunk.NewChunk(dc.Addr, dc.Data)
		p.logger.Trace("writing chunk to chunks channel", "caddr", c.Address())
		select {
		case w.chunks <- c:
		case <-s.quit:
			return
		}
	}
	p.logger.Debug("done writing batch to chunks channel")
}

func (s *SlipStream) clientSealBatch(p *Peer, provider StreamProvider, w *want) <-chan error {
	p.logger.Debug("stream.clientSealBatch", "stream", w.stream, "ruid", w.ruid, "from", w.from, "to", *w.to)
	errc := make(chan error)
	go func() {
		for {
			select {
			case c, ok := <-w.chunks:
				if !ok {
					p.logger.Error("want chanks returned on !ok")
				}
				p.mtx.RLock()
				if wants, ok := w.hashes[c.Address().Hex()]; !ok || !wants {
					p.logger.Error("got an unsolicited chunk from peer!", "peer", p.ID(), "caddr", c.Address)
					streamChunkDeliveryFail.Inc(1)
					p.Drop()
					return
				}
				p.mtx.RUnlock()
				cc := chunk.NewChunk(c.Address(), c.Data())
				go func() {
					ctx := context.TODO()
					seen, err := provider.Put(ctx, cc.Address(), cc.Data())
					if err != nil {
						if err == storage.ErrChunkInvalid {
							streamChunkDeliveryFail.Inc(1)
							p.Drop()
							return
						}
					}
					if seen {
						streamSeenChunkDelivery.Inc(1)
						p.logger.Error("chunk already seen!", "caddr", c.Address()) //this is possible when the same chunk is asked from multiple peers
					}
					p.mtx.Lock()
					w.hashes[c.Address().Hex()] = false
					p.mtx.Unlock()
					v := atomic.AddUint64(&w.remaining, ^uint64(0))
					p.logger.Trace("got chunk from peer", "addr", cc.Address(), "left", v)
					if v == 0 {
						p.logger.Debug("done receiving chunks for open want", "ruid", w.ruid)
						close(errc)
						return
					}
				}()
			case <-p.quit:
				return
			case <-w.done:
				return
			}
		}
	}()
	return errc
}
func (s *SlipStream) serverCollectBatch(ctx context.Context, p *Peer, provider StreamProvider, key interface{}, from, to uint64) (hashes []byte, f, t uint64, empty bool, err error) {
	p.logger.Debug("stream.CollectBatch", "from", from, "to", to)
	batchStart := time.Now()

	descriptors, stop := provider.Subscribe(ctx, key, from, to)
	defer stop()

	const batchTimeout = 500 * time.Millisecond

	var (
		batch        []byte
		batchSize    int
		batchStartID *uint64
		batchEndID   uint64
		timer        *time.Timer
		timerC       <-chan time.Time
	)

	defer func(start time.Time) {
		metrics.GetOrRegisterResettingTimer("stream.serverCollectBatch.total-time", nil).UpdateSince(start)
		metrics.GetOrRegisterCounter("stream.serverCollectBatch.batch-size", nil).Inc(int64(batchSize))
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
			s.logger.Trace("got address on subscribe", "address", d.Address.String(), "ident", batchStart, "key", key)
			batch = append(batch, d.Address[:]...)
			batchSize++
			if batchStartID == nil {
				// set batch start id only if
				// this is the first iteration
				batchStartID = &d.BinID
			}
			batchEndID = d.BinID
			if batchSize >= BatchSize {
				iterate = false
				metrics.GetOrRegisterCounter("stream.serverCollectBatch.full-batch", nil).Inc(1)
				p.logger.Trace("pull subscription - batch size reached", "batchSize", batchSize, "batchStartID", *batchStartID, "batchEndID", batchEndID)
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
			// return batch if new chunks are not received after some time
			iterate = false
			metrics.GetOrRegisterCounter("stream.serverCollectBatch.timer-expire", nil).Inc(1)
			p.logger.Trace("pull subscription timer expired", "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
		case <-p.quit:
			iterate = false
			p.logger.Trace("pull subscription - quit received", "batchSize", batchSize, "batchStartID", batchStartID, "batchEndID", batchEndID)
		case <-s.quit:
			iterate = false
			p.logger.Trace("pull subscription - shutting down")
		}
	}
	if batchStartID == nil {
		// if batch start id is not set, it means we timed out
		return nil, 0, 0, true, nil
	}
	return batch, *batchStartID, batchEndID, false, nil
}

func (s *SlipStream) PeerCursors() string {
	rows := []string{}
	rows = append(rows, fmt.Sprintf("peer subscriptions for base address: %s", hex.EncodeToString(s.baseKey)[:16]))
	ctr := 0
	for _, p := range s.peers {
		ctr++
		rows = append(rows, fmt.Sprintf("\tpeer: %s", hex.EncodeToString(p.OAddr)[:16]))
		cursors := p.getCursorsCopy()
		for stream, cursor := range cursors {

			rows = append(rows, fmt.Sprintf("\t\tstream:\t%-5s\t\tcursor: %d", stream, cursor))
		}
	}
	if ctr == 0 {
		rows = append(rows, fmt.Sprintf("\tfound no associated bzz-stream peers on this node"))
	}
	return "\n" + strings.Join(rows, "\n")
}

func (s *SlipStream) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "bzz-stream",
			Version: 1,
			Length:  10 * 1024 * 1024,
			Run:     s.runProtocol,
		},
	}
}

func (s *SlipStream) runProtocol(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, s.spec)
	// TODO: fix, used in tests only. Incorrect, as we do not have access to the overlay address
	bp := network.NewBzzPeer(peer)

	return s.Run(bp)
}

func (s *SlipStream) APIs() []rpc.API {
	return nil
}

func (s *SlipStream) Close() {
	close(s.quit)
	// wait for all handlers to finish
	done := make(chan struct{})
	go func() {
		s.handlersWg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Error("slip stream closed with still active handlers")
	}
}

func (s *SlipStream) Start(server *p2p.Server) error {
	log.Debug("slip stream starting")

	return nil
}

func (s *SlipStream) Stop() error {
	log.Debug("slip stream stopping")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.Close()
	for _, v := range s.providers {
		v.Close()
	}

	return nil
}
