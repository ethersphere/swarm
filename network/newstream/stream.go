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
	"encoding/json"
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

const (
	HashSize     = 32
	BatchSize    = 64
	MinFrameSize = 16
)

var (
	// Compile time interface check
	_ node.Service = (*Registry)(nil)

	// Metrics
	processReceivedChunksMsgCount = metrics.GetOrRegisterCounter("network.stream.received_chunks_msg", nil)
	processReceivedChunksCount    = metrics.GetOrRegisterCounter("network.stream.received_chunks_handled", nil)
	streamSeenChunkDelivery       = metrics.GetOrRegisterCounter("network.stream.seen_chunk_delivery", nil)
	streamEmptyWantedHashes       = metrics.GetOrRegisterCounter("network.stream.empty_wanted_hashes", nil)
	streamWantedHashes            = metrics.GetOrRegisterCounter("network.stream.wanted_hashes", nil)

	streamBatchFail               = metrics.GetOrRegisterCounter("network.stream.batch_fail", nil)
	streamChunkDeliveryFail       = metrics.GetOrRegisterCounter("network.stream.delivery_fail", nil)
	streamRequestNextIntervalFail = metrics.GetOrRegisterCounter("network.stream.next_interval_fail", nil)

	headBatchSizeGauge    = metrics.GetOrRegisterGauge("network.stream.batch_size_head", nil)
	batchSizeGauge        = metrics.GetOrRegisterGauge("network.stream.batch_size", nil)
	lastReceivedChunksMsg = metrics.GetOrRegisterGauge("network.stream.received_chunks", nil)

	streamPeersCount = metrics.GetOrRegisterGauge("network.stream.peers", nil)

	collectBatchLiveTimer    = metrics.GetOrRegisterResettingTimer("network.stream.server_collect_batch_head.total-time", nil)
	collectBatchHistoryTimer = metrics.GetOrRegisterResettingTimer("network.stream.server_collect_batch.total-time", nil)
	activeBatchTimeout       = 20 * time.Second

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

// Registry is the base type that handles all client/server operations on a node
// it is instantiated once per stream protocol instance, that is, it should have
// one instance per node
type Registry struct {
	mtx                     sync.RWMutex
	intervalsStore          state.Store               // store intervals for all peers
	peers                   map[enode.ID]*Peer        // peers
	baseKey                 []byte                    // this node's base address
	providers               map[string]StreamProvider // stream providers by name of stream
	spec                    *protocols.Spec           // this protocol's spec
	handlersWg              sync.WaitGroup            // waits for all handlers to finish in Close method
	quit                    chan struct{}             // signal shutdown
	lastReceivedChunkTimeMu sync.RWMutex              // synchronize access to lastReceivedChunkTime
	lastReceivedChunkTime   time.Time                 // last received chunk time
	logger                  log.Logger                // the logger for the registry. appends base address to all logs
}

// New creates a new stream protocol handler
func New(intervalsStore state.Store, baseKey []byte, providers ...StreamProvider) *Registry {
	r := &Registry{
		intervalsStore: intervalsStore,
		peers:          make(map[enode.ID]*Peer),
		providers:      make(map[string]StreamProvider),
		quit:           make(chan struct{}),
		baseKey:        baseKey,
		logger:         log.New("base", hex.EncodeToString(baseKey)),
		spec:           Spec,
	}
	for _, p := range providers {
		r.providers[p.StreamName()] = p
	}

	return r
}

// Run is being dispatched when 2 nodes connect
func (r *Registry) Run(bp *network.BzzPeer) error {
	sp := NewPeer(bp, r.baseKey, r.intervalsStore, r.providers)
	r.addPeer(sp)
	defer r.removePeer(sp)

	go sp.InitProviders()

	return sp.Peer.Run(r.HandleMsg(sp))
}

// HandleMsg is the main message handler for the stream protocol
func (r *Registry) HandleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		r.mtx.Lock() // ensure that quit read and handlersWg add are locked together
		defer r.mtx.Unlock()

		select {
		case <-r.quit:
			// no message handling if we quit
			return nil
		case <-p.quit:
			// peer has been removed, quit
			return nil
		default:
		}

		// handleMsgPauser should not be nil only in tests.
		// It does not use mutex lock protection and because of that
		// it must be set before the Registry is constructed and
		// reset when it is closed, in tests.
		// Production performance impact can be considered as
		// neglectable as nil check is a ns order operation.
		if handleMsgPauser != nil {
			handleMsgPauser.wait()
		}

		r.handlersWg.Add(1)
		go func() {
			defer r.handlersWg.Done()

			switch msg := msg.(type) {
			case *StreamInfoReq:
				r.serverHandleStreamInfoReq(ctx, p, msg)
			case *StreamInfoRes:
				if len(msg.Streams) == 0 {
					p.logger.Error("StreamInfo response is empty")
					p.Drop()
					return
				}

				r.clientHandleStreamInfoRes(ctx, p, msg)
			case *GetRange:
				provider := r.getProvider(msg.Stream)
				if provider == nil {
					p.logger.Error("unsupported provider", "stream", msg.Stream)
					p.Drop()
					return
				}
				r.serverHandleGetRange(ctx, p, msg, provider)
			case *OfferedHashes:
				// get the existing want for ruid from peer, otherwise drop
				w, exit := p.getWantOrDrop(msg.Ruid)
				if exit {
					return
				}
				provider := r.getProvider(w.stream)
				if provider == nil {
					p.logger.Error("unsupported provider", "stream", w.stream)
					p.Drop()
					return
				}
				r.clientHandleOfferedHashes(ctx, p, msg, w, provider)
			case *WantedHashes:
				// get the existing offer for ruid from peer, otherwise drop
				o, exit := p.getOfferOrDrop(msg.Ruid)
				if exit {
					return
				}
				provider := r.getProvider(o.stream)
				if provider == nil {
					p.logger.Error("unsupported provider", "stream", o.stream)
					p.Drop()
					return
				}
				r.serverHandleWantedHashes(ctx, p, msg, o, provider)
			case *ChunkDelivery:
				// get the existing want for ruid from peer, otherwise drop
				w, exit := p.getWantOrDrop(msg.Ruid)
				if exit {
					streamChunkDeliveryFail.Inc(1)
					return
				}
				provider := r.getProvider(w.stream)
				if provider == nil {
					p.logger.Error("unsupported provider", "stream", w.stream)
					p.Drop()
					return
				}
				r.clientHandleChunkDelivery(ctx, p, msg, w, provider)
			}
		}()
		return nil
	}
}

// Used to pause any message handling in tests for
// synchronizing desired states.
var handleMsgPauser pauser

type pauser interface {
	pause()
	resume()
	wait()
}

// serverHandleStreamInfoReq handles the StreamInfoReq message on the server side (Peer is the client)
func (r *Registry) serverHandleStreamInfoReq(ctx context.Context, p *Peer, msg *StreamInfoReq) {
	// illegal to request empty streams, drop peer
	if len(msg.Streams) == 0 {
		p.logger.Error("nil streams msg requested")
		p.Drop()
		return
	}

	streamRes := &StreamInfoRes{}
	for _, v := range msg.Streams {
		v := v
		provider := r.getProvider(v)
		if provider == nil {
			p.logger.Error("unsupported provider", "stream", v)
			// TODO: tell the other peer we dont support this stream? this is non fatal
			// this need not be fatal as we might not support all providers
			return
		}

		// get the current cursor from the data source
		streamCursor, err := provider.Cursor(v.Key)
		if err != nil {
			p.logger.Error("error getting cursor for stream key", "name", v.Name, "key", v.Key, "err", err)
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

	// don't send the message in case we're shutting down or the peer left
	select {
	case <-r.quit:
		// shutdown
		return
	case <-p.quit:
		// peer has been removed, quit
		return
	default:
	}

	if err := p.Send(ctx, streamRes); err != nil {
		p.logger.Error("failed to send StreamInfoRes to peer", "err", err)
		p.Drop()
	}
}

// clientHandleStreamInfoRes handles the StreamInfoRes message (Peer is the server)
func (r *Registry) clientHandleStreamInfoRes(ctx context.Context, p *Peer, msg *StreamInfoRes) {
	for _, s := range msg.Streams {
		s := s

		// get the provider for this stream
		provider := r.getProvider(s.Stream)
		if provider == nil {
			// at this point of the message exchange unsupported providers are illegal. drop peer
			p.logger.Error("peer requested unsupported provider. illegal, dropping peer")
			p.Drop()
			return
		}

		// check if we still want the requested stream. due to the fact that under certain conditions we might not
		// want to handle the stream by the time that StreamInfoRes has been received in response to StreamInfoReq
		if !provider.WantStream(p, s.Stream) {
			if _, exists := p.getCursor(s.Stream); exists {
				p.logger.Debug("stream cursor exists but we don't want it - removing", "stream", s.Stream)
				p.deleteCursor(s.Stream)
			}
			continue
		}

		// if the stream cursors exists for this peer - it means that a GetRange operation on it is already in progress
		if _, exists := p.getCursor(s.Stream); exists {
			p.logger.Debug("stream cursor already exists, continue to next", "stream", s.Stream)
			continue
		}

		p.logger.Debug("setting stream cursor", "stream", s.Stream, "cursor", s.Cursor)
		p.setCursor(s.Stream, s.Cursor)

		if provider.Autostart() {
			// don't request historical ranges for streams with cursor == 0
			if s.Cursor > 0 {
				p.logger.Debug("requesting history stream", "stream", s.Stream, "cursor", s.Cursor)
				// fetch everything from beginning till s.Cursor

				go func() {
					err := r.clientRequestStreamRange(ctx, p, provider, s.Stream, s.Cursor)
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
					err := r.clientRequestStreamHead(ctx, p, s.Stream, s.Cursor+1)
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

// clientRequestStreamHead sends a GetRange message to the server requesting
// new chunks from the supplied cursor position
func (r *Registry) clientRequestStreamHead(ctx context.Context, p *Peer, stream ID, from uint64) error {
	p.logger.Debug("clientRequestStreamHead", "stream", stream, "from", from)
	return r.clientCreateSendWant(ctx, p, stream, from, nil, true)
}

// clientRequestStreamRange sends a GetRange message to the server requesting
// a bound interval of chunks starting from the current stored interval in the
// interval store and ending at most in the supplied cursor position
func (r *Registry) clientRequestStreamRange(ctx context.Context, p *Peer, provider StreamProvider, stream ID, cursor uint64) error {
	p.logger.Debug("clientRequestStreamRange", "stream", stream, "cursor", cursor)

	// get the next interval from the intervals store
	from, _, empty, err := p.nextInterval(stream, 0)
	if err != nil {
		return err
	}

	// nothing to do - the next interval is bigger than the cursor or theinterval is empty
	if from > cursor || empty {
		p.logger.Debug("peer.requestStreamRange stream finished", "stream", stream, "cursor", cursor)
		return nil
	}
	return r.clientCreateSendWant(ctx, p, stream, from, &cursor, false)
}

func (r *Registry) clientCreateSendWant(ctx context.Context, p *Peer, stream ID, from uint64, to *uint64, head bool) error {
	g := GetRange{
		Ruid:      uint(rand.Uint32()),
		Stream:    stream,
		From:      from,
		To:        to,
		BatchSize: BatchSize,
	}

	p.mtx.Lock()
	p.openWants[g.Ruid] = &want{
		ruid:   g.Ruid,
		stream: g.Stream,
		from:   g.From,
		to:     to,
		head:   head,
		hashes: make(map[string]struct{}),
		chunks: make(chan chunk.Address),
		closeC: make(chan error),

		requested: time.Now(),
	}
	p.mtx.Unlock()

	return p.Send(ctx, g)
}

// serverHandleGetRange is handled by the server and sends in response an OfferedHashes message
// in the case that for the specific interval no chunks exist - the server sends an empty OfferedHashes
// message so that the client could seal the interval and request the next
func (r *Registry) serverHandleGetRange(ctx context.Context, p *Peer, msg *GetRange, provider StreamProvider) {
	p.logger.Debug("serverHandleGetRange", "ruid", msg.Ruid, "head?", msg.To == nil)
	start := time.Now()
	defer func(start time.Time) {
		if msg.To == nil {
			metrics.GetOrRegisterResettingTimer("network.stream.handle_get_range_head.total-time", nil).UpdateSince(start)
		} else {
			metrics.GetOrRegisterResettingTimer("network.stream.handle_get_range.total-time", nil).UpdateSince(start)
		}
	}(start)

	key, err := provider.ParseKey(msg.Stream.Key)
	if err != nil {
		p.logger.Error("erroring parsing stream key", "stream", msg.Stream, "err", err)
		p.Drop()
		return
	}

	// get hashes from the data source for this batch. to is 0 to denote we want whatever comes out of SubscribePull
	to := uint64(0)
	if msg.To != nil {
		to = *msg.To
	}
	h, _, t, e, err := r.serverCollectBatch(ctx, p, provider, key, msg.From, to)
	if err != nil {
		p.logger.Error("erroring getting live batch for stream", "stream", msg.Stream, "err", err)
		p.Drop()
		return
	}

	if e {
		// prevent sending an empty batch that resulted from db shutdown or peer quit
		select {
		case <-r.quit:
			return
		case <-p.quit:
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

	// store the offer for the peer
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
	l := len(h) / HashSize
	if msg.To == nil {
		headBatchSizeGauge.Update(int64(l))
	} else {
		batchSizeGauge.Update(int64(l))
	}
	if err := p.Send(ctx, offered); err != nil {
		p.logger.Error("erroring sending offered hashes", "ruid", msg.Ruid, "err", err)
		p.mtx.Lock()
		delete(p.openOffers, msg.Ruid)
		p.mtx.Unlock()
		p.Drop()
	}
}

// clientHandleOfferedHashes handles the OfferedHashes wire protocol message (Peer is the server)
func (r *Registry) clientHandleOfferedHashes(ctx context.Context, p *Peer, msg *OfferedHashes, w *want, provider StreamProvider) {
	p.logger.Debug("clientHandleOfferedHashes", "ruid", msg.Ruid, "msg.lastIndex", msg.LastIndex)
	start := time.Now()
	defer func(start time.Time) {
		metrics.GetOrRegisterResettingTimer("network.stream.handle_offered_hashes.total-time", nil).UpdateSince(start)
	}(start)

	var (
		lenHashes                    = len(msg.Hashes)
		ctr             uint64       = 0                                         // the number of chunks wanted out of the batch
		addresses                    = make([]chunk.Address, lenHashes/HashSize) // the address slice for MultiHas
		wantedHashesMsg              = WantedHashes{Ruid: msg.Ruid}              // the message to send back to the server
		errc            <-chan error                                             // channel to signal end of batch
	)

	if lenHashes%HashSize != 0 {
		p.logger.Error("invalid hashes length", "len", lenHashes, "ruid", msg.Ruid)
		p.Drop()
		return
	}

	w.to = &msg.LastIndex // now that we know the range of the batch we can set the upped bound of the interval to the open want

	// this code block handles the case of a complete gap on the interval on the server side
	// lenhashes == 0 means there's no hashes in the requested range with the upper bound of
	// the LastIndex on the incoming message. we should seal the interval and request the subsequent
	if lenHashes == 0 {
		if err := p.sealWant(w); err != nil {
			p.logger.Error("error persisting interval", "from", w.from, "to", w.to, "err", err)
			p.Drop()
			return
		}
		r.requestSubsequentRange(ctx, p, provider, w, msg.LastIndex)
		return
	}

	want, err := bv.New(lenHashes / HashSize)
	if err != nil {
		p.logger.Error("error initiaising bitvector", "len", lenHashes/HashSize, "ruid", msg.Ruid, "err", err)
		p.Drop()
		return
	}

	for i := 0; i < lenHashes; i += HashSize {
		hash := msg.Hashes[i : i+HashSize]
		addresses[i/HashSize] = hash
	}

	// check which hashes we want
	if hasses, err := provider.NeedData(ctx, addresses...); err == nil {
		for i, has := range hasses {
			if !has {
				ctr++                                     // increment number of wanted chunks
				want.Set(i)                               // set the bitvector
				w.hashes[addresses[i].Hex()] = struct{}{} // set unsolicited chunks guard
			}
		}
	} else {
		p.logger.Error("multi need data returned an error, dropping peer", "err", err)
		p.Drop()
		return
	}

	// set the number of remaining chunks to ctr
	atomic.AddUint64(&w.remaining, ctr)

	// this handles the case that there are no hashes we are interested in
	// we then seal the current interval and request the next batch
	if ctr == 0 {
		streamEmptyWantedHashes.Inc(1)
		wantedHashesMsg.BitVector = []byte{} // set the bitvector value to an empty slice, this is to signal the server we dont want any hashes
		if err := p.sealWant(w); err != nil {
			p.logger.Error("error persisting interval", "from", w.from, "to", *w.to, "err", err)
			p.Drop()
			return
		}
	} else {
		// we want some hashes
		streamWantedHashes.Inc(1)
		wantedHashesMsg.BitVector = want.Bytes() // set to bitvector

		errc = r.clientSealBatch(ctx, p, provider, w) // poll for the completion of the batch in a separate goroutine
	}

	if err := p.Send(ctx, wantedHashesMsg); err != nil {
		p.logger.Error("error sending wanted hashes", "err", err)
		p.Drop()
		return
	}
	if ctr == 0 {
		// request the next range in case no chunks wanted
		r.requestSubsequentRange(ctx, p, provider, w, msg.LastIndex)
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

		// seal the interval
		if err := p.sealWant(w); err != nil {
			p.logger.Error("error persisting interval", "from", w.from, "to", w.to, "err", err)
			p.Drop()
			return
		}
	case <-time.After(activeBatchTimeout):
		p.logger.Error("batch has timed out", "ruid", w.ruid)
		close(w.closeC) // signal the polling goroutine to terminate
		p.mtx.Lock()
		delete(p.openWants, msg.Ruid)
		p.mtx.Unlock()
		p.Drop()
		return
	case <-r.quit:
		return
	case <-p.quit:
		return
	}
	r.requestSubsequentRange(ctx, p, provider, w, msg.LastIndex)
}

// serverHandleWantedHashes is handled on the server side (Peer is the client) and is dependent on a preceding OfferedHashes message
// the method is to ensure that all chunks in the requested batch is sent to the client
func (r *Registry) serverHandleWantedHashes(ctx context.Context, p *Peer, msg *WantedHashes, o offer, provider StreamProvider) {
	p.logger.Debug("serverHandleWantedHashes", "ruid", msg.Ruid)
	start := time.Now()
	defer func(start time.Time) {
		metrics.GetOrRegisterResettingTimer("network.stream.handle_wanted_hashes.total-time", nil).UpdateSince(start)
	}(start)

	defer func() {
		p.mtx.Lock()
		delete(p.openOffers, msg.Ruid)
		p.mtx.Unlock()
	}()

	l := len(o.hashes) / HashSize
	if len(msg.BitVector) == 0 {
		p.logger.Debug("peer does not want any hashes in this range", "ruid", o.ruid)
		return
	}
	want, err := bv.NewFromBytes(msg.BitVector, l)
	if err != nil {
		p.logger.Error("error initiaising bitvector", "l", l, "ll", len(o.hashes), "err", err)
		p.Drop()
		return
	}

	var (
		cd         = &ChunkDelivery{Ruid: msg.Ruid}
		wantHashes = []chunk.Address{}
	)

	maxFrame := MinFrameSize
	if v := BatchSize / 4; v > maxFrame {
		maxFrame = v
	}

	// check which hashes to get from the localstore
	for i := 0; i < l; i++ {
		if want.Get(i) {
			metrics.GetOrRegisterCounter("network.stream.handle_wanted.want_get", nil).Inc(1)
			hash := o.hashes[i*HashSize : (i+1)*HashSize]
			wantHashes = append(wantHashes, hash)
		}
	}

	// get the chunks from the provider
	chunks, err := provider.Get(ctx, wantHashes...)
	if err != nil {
		p.logger.Error("handleWantedHashesMsg", "err", err)
		p.Drop()
		return
	}

	// append the chunks to the chunk delivery message. when reaching maxFrameSize send the current batch
	for _, v := range chunks {
		chunkD := DeliveredChunk{
			Addr: v.Address(),
			Data: v.Data(),
		}
		cd.Chunks = append(cd.Chunks, chunkD)

		if len(cd.Chunks) == maxFrame {
			// prevent sending batch on shutdown or peer dropout
			select {
			case <-p.quit:
				return
			case <-r.quit:
				return
			default:
			}

			//send the batch and reset chunk delivery message
			if err := p.Send(ctx, cd); err != nil {
				p.logger.Error("error sending chunk delivery frame", "ruid", msg.Ruid, "error", err)
				p.Drop()
				return
			}
			cd = &ChunkDelivery{
				Ruid: msg.Ruid,
			}
		}
	}

	// send anything that we might have left in the batch
	if len(cd.Chunks) > 0 {
		if err := p.Send(ctx, cd); err != nil {
			p.logger.Error("error sending chunk delivery frame", "ruid", msg.Ruid, "error", err)
			p.Drop()
		}
	}

	// set the chunks as synced
	err = provider.Set(ctx, wantHashes...)
	if err != nil {
		p.logger.Error("error setting chunk as synced", "addrs", wantHashes, "err", err)
		p.Drop()
		return
	}
}

// clientHandleChunkDelivery handles chunk delivery messages
func (r *Registry) clientHandleChunkDelivery(ctx context.Context, p *Peer, msg *ChunkDelivery, w *want, provider StreamProvider) {
	p.logger.Debug("clientHandleChunkDelivery", "ruid", msg.Ruid)
	processReceivedChunksMsgCount.Inc(1)
	lastReceivedChunksMsg.Update(time.Now().UnixNano())
	r.setLastReceivedChunkTime() // needed for IsPullSyncing
	defer func(start time.Time) {
		metrics.GetOrRegisterResettingTimer("network.stream.handle_chunk_delivery.total-time", nil).UpdateSince(start)
	}(time.Now())

	chunks := make([]chunk.Chunk, len(msg.Chunks))
	for i, dc := range msg.Chunks {
		chunks[i] = chunk.NewChunk(dc.Addr, dc.Data)
	}

	// put the chunks to the local store
	seen, err := provider.Put(ctx, chunks...)
	if err != nil {
		if err == storage.ErrChunkInvalid {
			streamChunkDeliveryFail.Inc(1)
			p.Drop()
			return
		}
		p.logger.Error("clientHandleChunkDelivery error putting chunk", "err", err)
		return
	}

	// increment seen chunk delivery metric. duplicate delivery is possible when the same chunk is asked from multiple peers, we currently do not limit this
	for _, v := range seen {
		if v {
			streamSeenChunkDelivery.Inc(1)
		}
	}

	for _, dc := range chunks {
		select {
		case w.chunks <- dc.Address():
			// send the chunk address to the goroutine polling end of batch (clientSealBatch)
		case <-w.closeC:
			// batch timeout
			return
		case <-r.quit:
			// shutdown
			return
		case <-p.quit:
			// peer quit
			return
		}
	}
}

// clientSealBatch seals a given batch (want). it launches a separate goroutine that check every chunk being delivered on the given ruid
// if an unsolicited chunk is received it drops the peer
func (r *Registry) clientSealBatch(ctx context.Context, p *Peer, provider StreamProvider, w *want) <-chan error {
	p.logger.Debug("clientSealBatch", "stream", w.stream, "ruid", w.ruid, "from", w.from, "to", *w.to)
	errc := make(chan error)
	go func() {
		start := time.Now()
		defer func(start time.Time) {
			metrics.GetOrRegisterResettingTimer("network.stream.client_seal_batch.total-time", nil).UpdateSince(start)
		}(start)
		for {
			select {
			case c, ok := <-w.chunks:
				if !ok {
					return
				}
				processReceivedChunksCount.Inc(1)
				p.mtx.Lock()
				if _, ok := w.hashes[c.Hex()]; !ok {
					p.logger.Error("got an unsolicited chunk from peer!", "peer", p.ID(), "caddr", c)
					streamChunkDeliveryFail.Inc(1)
					p.Drop()
					p.mtx.Unlock()
					return
				}
				delete(w.hashes, c.Hex())
				p.mtx.Unlock()
				v := atomic.AddUint64(&w.remaining, ^uint64(0))
				if v == 0 {
					p.logger.Trace("done receiving chunks for open want", "ruid", w.ruid)
					close(errc)
					return
				}
			case <-p.quit:
				// peer quit
				return
			case <-w.closeC:
				// batch timeout was signalled
				return
			case <-r.quit:
				// shutdown
				return
			}
		}
	}()
	return errc
}

// serverCollectBatch collects a batch of hashes in response for a GetRange message
// it will block until at least one hash is received from the provider
func (r *Registry) serverCollectBatch(ctx context.Context, p *Peer, provider StreamProvider, key interface{}, from, to uint64) (hashes []byte, f, t uint64, empty bool, err error) {
	p.logger.Debug("serverCollectBatch", "from", from, "to", to)

	const batchTimeout = 1 * time.Second

	var (
		batch        []byte
		batchSize    int
		batchStartID *uint64
		batchEndID   uint64
		timer        *time.Timer
		timerC       <-chan time.Time
	)

	defer func(start time.Time) {
		if to == 0 {
			collectBatchLiveTimer.UpdateSince(start)
		} else {
			collectBatchHistoryTimer.UpdateSince(start)
		}
		if timer != nil {
			timer.Stop()
		}
	}(time.Now())

	descriptors, stop := provider.Subscribe(ctx, key, from, to)
	defer stop()

	for iterate := true; iterate; {
		select {
		case d, ok := <-descriptors:
			if !ok {
				iterate = false
				break
			}
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
				metrics.GetOrRegisterCounter("network.stream.server_collect_batch.full-batch", nil).Inc(1)
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
			metrics.GetOrRegisterCounter("network.stream.server_collect_batch.timer-expire", nil).Inc(1)
		case <-p.quit:
			iterate = false
		case <-r.quit:
			iterate = false
		}
	}
	if batchStartID == nil {
		// if batch start id is not set, it means we timed out
		return nil, 0, 0, true, nil
	}
	return batch, *batchStartID, batchEndID, false, nil
}

// requestSubsequentRange checks the cursor for the current stream, and in case needed - requests the next range
func (r *Registry) requestSubsequentRange(ctx context.Context, p *Peer, provider StreamProvider, w *want, lastIndex uint64) {
	cur, ok := p.getCursor(w.stream)
	if !ok {
		metrics.GetOrRegisterCounter("network.stream.quit_unwanted", nil).Inc(1)
		p.logger.Debug("no longer interested in stream. quitting", "stream", w.stream)
		p.mtx.Lock()
		delete(p.openWants, w.ruid)
		p.mtx.Unlock()
		return
	}
	if w.head {
		if err := r.clientRequestStreamHead(ctx, p, w.stream, lastIndex+1); err != nil {
			streamRequestNextIntervalFail.Inc(1)
			p.logger.Error("error requesting next interval from peer", "err", err)
			p.Drop()
			return
		}
	} else {
		if err := r.clientRequestStreamRange(ctx, p, provider, w.stream, cur); err != nil {
			streamRequestNextIntervalFail.Inc(1)
			p.logger.Error("error requesting next interval from peer", "err", err)
			p.Drop()
			return
		}
	}
}

func (r *Registry) getProvider(stream ID) StreamProvider {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.providers[stream.Name]
}

func (r *Registry) getPeer(id enode.ID) *Peer {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	p := r.peers[id]
	return p
}

func (r *Registry) addPeer(p *Peer) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.peers[p.ID()] = p

	streamPeersCount.Update(int64(len(r.peers)))
}

func (r *Registry) removePeer(p *Peer) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if _, found := r.peers[p.ID()]; found {
		p.logger.Error("removing peer")
		delete(r.peers, p.ID())
		close(p.quit)
	}
	streamPeersCount.Update(int64(len(r.peers)))
}

// PeerCurosrs returns a JSON response in which the queried node's
// peer cursors are returned
func (r *Registry) PeerCursors() string {
	type peerCurs struct {
		Peer    string            `json:"peer"` // the peer address
		Cursors map[string]uint64 `json:"cursors"`
	}
	curs := struct {
		Base  string     `json:"base"` // our node's base address
		Peers []peerCurs `json:"peers"`
	}{
		Base: hex.EncodeToString(r.baseKey)[:16],
	}

	for _, p := range r.peers {
		pcur := peerCurs{
			Peer:    hex.EncodeToString(p.OAddr)[:16],
			Cursors: p.getCursorsCopy(),
		}
		curs.Peers = append(curs.Peers, pcur)
	}
	pc, err := json.Marshal(&curs)
	if err != nil {
		return ""
	}
	return string(pc)
}

// LastReceivedChunkTime returns the time when the last chunk
// was received by syncing. This method is used in api.Inspector
// to detect when the syncing is complete.
func (r *Registry) LastReceivedChunkTime() time.Time {
	r.lastReceivedChunkTimeMu.RLock()
	defer r.lastReceivedChunkTimeMu.RUnlock()
	return r.lastReceivedChunkTime
}

func (r *Registry) setLastReceivedChunkTime() {
	r.lastReceivedChunkTimeMu.Lock()
	r.lastReceivedChunkTime = time.Now()
	r.lastReceivedChunkTimeMu.Unlock()
}

func (r *Registry) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "bzz-stream",
			Version: 1,
			Length:  10 * 1024 * 1024,
			Run:     r.runProtocol,
		},
	}
}

func (r *Registry) runProtocol(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, r.spec)
	// TODO: fix, used in tests only. Incorrect, as we do not have access to the overlay address
	bp := network.NewBzzPeer(peer)

	return r.Run(bp)
}

func (r *Registry) APIs() []rpc.API {
	return nil
}

func (r *Registry) Start(server *p2p.Server) error {
	r.logger.Debug("stream registry starting")

	return nil
}

func (r *Registry) Stop() error {
	log.Debug("stream registry stopping")
	r.mtx.Lock()
	defer r.mtx.Unlock()

	close(r.quit)
	// wait for all handlers to finish
	done := make(chan struct{})
	go func() {
		r.handlersWg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Error("stream closed with still active handlers")
	}

	for _, v := range r.providers {
		v.Close()
	}

	return nil
}
