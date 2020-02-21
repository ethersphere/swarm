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
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	bv "github.com/ethersphere/swarm/network/bitvector"
	"github.com/ethersphere/swarm/network/stream/intervals"
	"github.com/ethersphere/swarm/network/timeouts"
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

	headBatchSizeGauge = metrics.GetOrRegisterGauge("network.stream.batch_size_head", nil)
	batchSizeGauge     = metrics.GetOrRegisterGauge("network.stream.batch_size", nil)

	streamPeersCount = metrics.GetOrRegisterGauge("network.stream.peers", nil)

	collectBatchLiveTimer    = metrics.GetOrRegisterResettingTimer("network.stream.server_collect_batch_head.total-time", nil)
	collectBatchHistoryTimer = metrics.GetOrRegisterResettingTimer("network.stream.server_collect_batch.total-time", nil)
	providerGetTimer         = metrics.GetOrRegisterResettingTimer("network.stream.provider_get.total-time", nil)
	providerPutTimer         = metrics.GetOrRegisterResettingTimer("network.stream.provider_put.total-time", nil)
	providerSetTimer         = metrics.GetOrRegisterResettingTimer("network.stream.provider_set.total-time", nil)
	providerNeedDataTimer    = metrics.GetOrRegisterResettingTimer("network.stream.provider_need_data.total-time", nil)

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

	// pause the msgHandler execution, used only for tests
	handleMsgPauser protocols.MsgPauser = nil
)

// Registry is the base type that handles all client/server operations on a node
// it is instantiated once per stream protocol instance, that is, it should have
// one instance per node
type Registry struct {
	mtx                     sync.RWMutex
	intervalsStore          state.Store               // store intervals for all peers
	peers                   map[enode.ID]*Peer        // peers
	address                 *network.BzzAddr          // this node's base address
	providers               map[string]StreamProvider // stream providers by name of stream
	spec                    *protocols.Spec           // this protocol's spec
	quit                    chan struct{}             // signal shutdown
	lastReceivedChunkTimeMu sync.RWMutex              // synchronize access to lastReceivedChunkTime
	lastReceivedChunkTime   time.Time                 // last received chunk time
	logger                  log.Logger                // the logger for the registry. appends base address to all logs
}

// New creates a new stream protocol handler
func New(intervalsStore state.Store, address *network.BzzAddr, providers ...StreamProvider) *Registry {
	r := &Registry{
		intervalsStore: intervalsStore,
		peers:          make(map[enode.ID]*Peer),
		providers:      make(map[string]StreamProvider),
		quit:           make(chan struct{}),
		address:        address,
		logger:         log.New("base", address.ShortString()),
		spec:           Spec,
	}
	for _, p := range providers {
		r.providers[p.StreamName()] = p
	}

	return r
}

// Run is being dispatched when 2 nodes connect
func (r *Registry) Run(bp *network.BzzPeer) error {
	sp := newPeer(bp, r.address, r.intervalsStore, r.providers)
	// enable msg pauser for stream protocol, this is used only in tests
	sp.Peer.SetMsgPauser(handleMsgPauser)
	r.addPeer(sp)
	defer r.removePeer(sp)
	go sp.InitProviders()

	return sp.Peer.Run(r.HandleMsg(sp))
}

// HandleMsg is the main message handler for the stream protocol
func (r *Registry) HandleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		switch msg := msg.(type) {
		case *StreamInfoReq:
			return r.serverHandleStreamInfoReq(ctx, p, msg)
		case *StreamInfoRes:
			return r.clientHandleStreamInfoRes(ctx, p, msg)
		case *GetRange:
			return r.serverHandleGetRange(ctx, p, msg)
		case *OfferedHashes:
			return r.clientHandleOfferedHashes(ctx, p, msg)
		case *WantedHashes:
			return r.serverHandleWantedHashes(ctx, p, msg)
		case *ChunkDelivery:
			return r.clientHandleChunkDelivery(ctx, p, msg)

		default:
			// todo: maybe a special error for unknown message, or at least just log it
			return nil
		}
	}
}

// serverHandleStreamInfoReq handles the StreamInfoReq message on the server side (Peer is the client)
func (r *Registry) serverHandleStreamInfoReq(ctx context.Context, p *Peer, msg *StreamInfoReq) error {
	// illegal to request empty streams, drop peer
	if len(msg.Streams) == 0 {
		return protocols.Break(errors.New("nil streams msg requested"))
	}

	streamRes := &StreamInfoRes{}
	for _, v := range msg.Streams {
		provider := r.getProvider(v)
		if provider == nil {
			return fmt.Errorf("unsupported provider for stream: %s", v)
		}

		// get the current cursor from the data source
		streamCursor, err := provider.Cursor(v.Key)
		if err != nil {
			return protocols.Break(fmt.Errorf("get cursor for stream key failed, name %s, key %s: %w", v.Name, v.Key, err))
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
		return nil
	case <-p.quit:
		// peer has been removed, quit
		return nil
	default:
	}

	if err := p.Send(ctx, streamRes); err != nil {
		return protocols.Break(err)
	}

	return nil
}

// clientHandleStreamInfoRes handles the StreamInfoRes message (Peer is the server)
func (r *Registry) clientHandleStreamInfoRes(ctx context.Context, p *Peer, msg *StreamInfoRes) error {
	if len(msg.Streams) == 0 {
		return protocols.Break(errors.New("message stream was empty"))
	}

	for _, s := range msg.Streams {
		s := s

		// get the provider for this stream
		provider := r.getProvider(s.Stream)
		if provider == nil {
			// at this point of the message exchange unsupported providers are illegal. drop peer
			return protocols.Break(errors.New("peer requested unsupported provider"))
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
					// todo: return DropError
					if err != nil {
						p.Drop("had an error sending initial GetRange for historical stream")
					}
				}()
			}

			// handle stream unboundedness
			if !s.Bounded {
				//constantly fetch the head of the stream
				p.logger.Debug("asking for live stream", "stream", s.Stream, "cursor", s.Cursor)
				// ask the tip (cursor + 1)
				go func() {
					// todo: return DropError
					err := r.clientRequestStreamHead(ctx, p, s.Stream, s.Cursor+1)
					// https://github.com/golang/go/issues/4373 - use of closed network connection
					if err != nil && err != p2p.ErrShuttingDown && !strings.Contains(err.Error(), "use of closed network connection") {
						p.Drop("had an error with initial stream head fetch: %s")
					}
				}()
			}
		}
	}

	return nil
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
		return protocols.Break(err)
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
	s := p.getRangeKey(stream, head)
	if v, ok := p.clientOpenGetRange[s]; ok {
		p.logger.Warn("batch already requested, skipping", "stream", stream, "head", head, "from", from, "to", to, "existing ruid", v)
		p.mtx.Unlock()
		return nil
	}
	p.clientOpenGetRange[s] = g.Ruid

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

	p.logger.Trace("clientCreateSendWant", "ruid", g.Ruid, "stream", g.Stream, "from", g.From, "to", to)

	return p.Send(ctx, g)
}

// serverHandleGetRange is handled by the server and sends in response an OfferedHashes message
// in the case that for the specific interval no chunks exist - the server sends an empty OfferedHashes
// message so that the client could seal the interval and request the next
func (r *Registry) serverHandleGetRange(ctx context.Context, p *Peer, msg *GetRange) error {
	provider := r.getProvider(msg.Stream)
	if provider == nil {
		return protocols.Break(fmt.Errorf("unsupported provider"))
	}

	p.logger.Debug("serverHandleGetRange", "ruid", msg.Ruid, "head?", msg.To == nil)
	p.mtx.Lock()
	s := p.getRangeKey(msg.Stream, msg.To == nil)
	if ruid, exists := p.serverOpenGetRange[s]; exists {
		p.logger.Debug("stream request already ongoing, skipping", "ruid in flight", ruid)
		p.mtx.Unlock()
		return nil
	}
	p.serverOpenGetRange[s] = msg.Ruid
	p.mtx.Unlock()

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
		return protocols.Break(fmt.Errorf("parsing stream key for stream %s: %w", msg.Stream, err))
	}

	// get hashes from the data source for this batch. to is 0 to denote we want whatever comes out of SubscribePull
	to := uint64(0)
	if msg.To != nil {
		to = *msg.To
	}
	h, _, t, e, err := r.serverCollectBatch(ctx, p, provider, key, msg.From, to)
	if err != nil {
		return protocols.Break(fmt.Errorf("getting live batch for stream %s: %w", msg.Stream, err))
	}

	if e {
		// prevent sending an empty batch that resulted from db shutdown or peer quit
		select {
		case <-r.quit:
			return nil
		case <-p.quit:
			return nil
		default:
			// if the batch is empty resulting from a request for the tip
			// the lastIdx is msg.From
			// if the range was defined - then it equals to the top of the requested range - msg.To
			lastIdx := msg.From
			if msg.To != nil {
				lastIdx = *msg.To
			}
			offered := OfferedHashes{
				Ruid:      msg.Ruid,
				LastIndex: lastIdx,
				Hashes:    []byte{},
			}

			if err := p.Send(ctx, offered); err != nil {
				return protocols.Break(fmt.Errorf("sending empty live offered hashes, ruid %d: %w", msg.Ruid, err))
			}
			return nil
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
		p.mtx.Lock()
		delete(p.openOffers, msg.Ruid)
		p.mtx.Unlock()
		return protocols.Break(fmt.Errorf("sending offered hashes, ruid %d: %w", msg.Ruid, err))
	}

	p.mtx.Lock()
	delete(p.serverOpenGetRange, s)
	p.mtx.Unlock()

	return nil
}

// clientHandleOfferedHashes handles the OfferedHashes wire protocol message (Peer is the server)
func (r *Registry) clientHandleOfferedHashes(ctx context.Context, p *Peer, msg *OfferedHashes) error {
	w, err := p.getWant(msg.Ruid)
	if err != nil {
		return protocols.Break(err)
	}
	provider := r.getProvider(w.stream)
	if provider == nil {
		return protocols.Break(fmt.Errorf("unsupported provider"))
	}

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
		return protocols.Break(fmt.Errorf("invalid hashes length: %d, ruid: %d", lenHashes, msg.Ruid))
	}

	w.to = &msg.LastIndex // we can set the open wants upper bound to the index supplied in the msg

	// this code block handles the case of a complete gap on the interval on the server side
	// lenhashes == 0 means there's no hashes in the requested range with the upper bound of
	// the LastIndex on the incoming message. we should seal the interval and request the subsequent
	if lenHashes == 0 {
		if err := p.sealWant(w); err != nil {
			return protocols.Break(fmt.Errorf("persisting interval from %d, to %d: %w", w.from, w.to, err))
		}
		return r.requestSubsequentRange(ctx, p, provider, w, msg.LastIndex)
	}

	// we no longer want this stream. send an empty wanted
	// hashes to upstream peer to not deliver the batch
	// it is important that this case is handled after the `lenHashes==0`
	// block above, since the case above is a special case where the
	// upstream peer already deleted the `offer` object associated with
	// the Ruid, while in this case we must send an empty message back to
	// the upstream peer in order to mitigate a leak on `offer`s
	if !provider.WantStream(p, w.stream) {
		wantedHashesMsg.BitVector = []byte{}
		if err := p.Send(ctx, wantedHashesMsg); err != nil {
			return protocols.Break(fmt.Errorf("sending empty wanted hashes:  %w", err))
		}
		return nil
	}

	want, err := bv.New(lenHashes / HashSize)
	if err != nil {
		return protocols.Break(fmt.Errorf("initialising bitvector, len %d, ruid %d: %w", lenHashes/HashSize, msg.Ruid, err))
	}

	for i := 0; i < lenHashes; i += HashSize {
		hash := msg.Hashes[i : i+HashSize]
		addresses[i/HashSize] = hash
		p.logger.Trace("clientHandleOfferedHashes peer offered hash", "ruid", msg.Ruid, "stream", w.stream, "chunk", addresses[i/HashSize])
	}

	startNeed := time.Now()

	// check which hashes we want
	wants, err := provider.NeedData(ctx, addresses...)
	if err != nil {
		return protocols.Break(err)
	}

	for i, wantChunk := range wants {
		if wantChunk {
			ctr++                                     // increment number of wanted chunks
			want.Set(i)                               // set the bitvector
			w.hashes[addresses[i].Hex()] = struct{}{} // set unsolicited chunks guard
		}
	}

	providerNeedDataTimer.UpdateSince(startNeed)

	// set the number of remaining chunks to ctr
	atomic.AddUint64(&w.remaining, ctr)

	// this handles the case that there are no hashes we are interested in
	// we then seal the current interval and request the next batch
	if ctr == 0 {
		streamEmptyWantedHashes.Inc(1)
		wantedHashesMsg.BitVector = []byte{} // set the bitvector value to an empty slice, this is to signal the server we dont want any hashes
		if err := p.sealWant(w); err != nil {
			return protocols.Break(fmt.Errorf("persisting interval from %d, to %d: %w", w.from, w.to, err))
		}
		if err := p.Send(ctx, wantedHashesMsg); err != nil {
			return protocols.Break(fmt.Errorf("sending wanted hashes: %w", err))
		}

		// request the next range in case no chunks wanted
		return r.requestSubsequentRange(ctx, p, provider, w, msg.LastIndex)
	} else {
		// we want some hashes
		streamWantedHashes.Inc(1)
		wantedHashesMsg.BitVector = want.Bytes() // set to bitvector

		errc = r.clientSealBatch(ctx, p, provider, w) // poll for the completion of the batch in a separate goroutine
	}

	if err := p.Send(ctx, wantedHashesMsg); err != nil {
		return protocols.Break(fmt.Errorf("sending wanted hashes: %w", err))
	}

	select {
	case err := <-errc:
		if err != nil {
			streamBatchFail.Inc(1)
			return protocols.Break(fmt.Errorf("sealing batch from %d, to %d: %w", w.from, w.to, err))
		}

		// seal the interval
		if err := p.sealWant(w); err != nil {
			return protocols.Break(fmt.Errorf("persisting interval from %d, to %d: %w", w.from, w.to, err))
		}
	case <-time.After(timeouts.SyncBatchTimeout):
		p.logger.Error("batch has timed out", "ruid", w.ruid)
		close(w.closeC) // signal the polling goroutine to terminate
		p.mtx.Lock()
		delete(p.openWants, msg.Ruid)
		p.mtx.Unlock()

		// todo: this should happen because of the returned error anyway
		// if the stream is wanted and has timed out
		// then drop the peer. this safeguards the edge
		// case that a batch times out when a kademlia
		// depth change occurs between the call to
		// clientSealBatch and a subsequent chunk delivery
		// message
		if provider.WantStream(p, w.stream) {
			return protocols.Break(errors.New("batch has timed out"))
		}
		return nil
	case <-r.quit:
		return nil
	case <-p.quit:
		return nil
	}
	return r.requestSubsequentRange(ctx, p, provider, w, msg.LastIndex)
}

// serverHandleWantedHashes is handled on the server side (Peer is the client) and is dependent on a preceding OfferedHashes message
// the method is to ensure that all chunks in the requested batch is sent to the client
func (r *Registry) serverHandleWantedHashes(ctx context.Context, p *Peer, msg *WantedHashes) error {
	// get the existing offer for ruid from peer, otherwise drop
	o, err := p.getOffer(msg.Ruid)
	if err != nil {
		return protocols.Break(err)
	}
	provider := r.getProvider(o.stream)
	if provider == nil {
		return protocols.Break(errors.New("unsupported provider"))
	}

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

	var (
		l          = len(o.hashes) / HashSize
		cd         = &ChunkDelivery{Ruid: msg.Ruid}
		wantHashes = []chunk.Address{}
		allHashes  = make([]chunk.Address, l)
	)

	if len(msg.BitVector) == 0 {
		p.logger.Debug("peer does not want any hashes in this range", "ruid", o.ruid)
		for i := 0; i < l; i++ {
			allHashes[i] = o.hashes[i*HashSize : (i+1)*HashSize]
		}
		// set all chunks as synced
		if err := provider.Set(ctx, allHashes...); err != nil {
			return protocols.Break(fmt.Errorf("setting chunk as synced, addrs %s: %w", allHashes, err))
		}
		return nil
	}
	want, err := bv.NewFromBytes(msg.BitVector, l)
	if err != nil {
		return protocols.Break(fmt.Errorf("initialising bitvector, l %d, ll %d: %w", l, len(o.hashes), err))
	}

	maxFrame := MinFrameSize
	if v := BatchSize / 4; v > maxFrame {
		maxFrame = v
	}

	// check which hashes to get from the localstore
	for i := 0; i < l; i++ {
		hash := o.hashes[i*HashSize : (i+1)*HashSize]
		if want.Get(i) {
			metrics.GetOrRegisterCounter("network.stream.handle_wanted.want_get", nil).Inc(1)
			wantHashes = append(wantHashes, hash)
		}
		allHashes[i] = hash
	}
	startGet := time.Now()

	// get the chunks from the provider
	chunks, err := provider.Get(ctx, wantHashes...)
	if err != nil {
		return protocols.Break(fmt.Errorf("get provider: %w", err))
	}

	providerGetTimer.UpdateSince(startGet) // measure how long we spend on getting the chunks

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
				return nil
			case <-r.quit:
				return nil
			default:
			}

			//send the batch and reset chunk delivery message
			if err := p.Send(ctx, cd); err != nil {
				return protocols.Break(fmt.Errorf("sending chunk delivery frame, ruid %d: %w", msg.Ruid, err))

			}
			cd = &ChunkDelivery{
				Ruid: msg.Ruid,
			}
		}
	}

	// send anything that we might have left in the batch
	if len(cd.Chunks) > 0 {
		if err := p.Send(ctx, cd); err != nil {
			return protocols.Break(fmt.Errorf("sending chunk delivery frame failed, ruid %d: %w", msg.Ruid, err))
		}
	}

	startSet := time.Now()

	// set the chunks as synced
	err = provider.Set(ctx, allHashes...)
	if err != nil {
		return protocols.Break(fmt.Errorf("sending chunk as synced, addr: %s: %w", allHashes, err))

	}
	providerSetTimer.UpdateSince(startSet)

	return nil
}

// clientHandleChunkDelivery handles chunk delivery messages
func (r *Registry) clientHandleChunkDelivery(ctx context.Context, p *Peer, msg *ChunkDelivery) error {
	// get the existing want for ruid from peer, otherwise drop
	w, err := p.getWant(msg.Ruid)
	if err != nil {
		streamChunkDeliveryFail.Inc(1)
		return protocols.Break(err)
	}
	provider := r.getProvider(w.stream)
	if provider == nil {
		return protocols.Break(fmt.Errorf("unsupported provider"))
	}

	p.logger.Debug("clientHandleChunkDelivery", "ruid", msg.Ruid)

	// don't process this message if we're no longer
	// interested in this stream
	if !provider.WantStream(p, w.stream) {
		return nil
	}
	processReceivedChunksMsgCount.Inc(1)
	r.setLastReceivedChunkTime() // needed for IsPullSyncing

	defer func(start time.Time) {
		metrics.GetOrRegisterResettingTimer("network.stream.handle_chunk_delivery.total-time", nil).UpdateSince(start)
	}(time.Now())

	chunks := make([]chunk.Chunk, len(msg.Chunks))
	for i, dc := range msg.Chunks {
		chunks[i] = chunk.NewChunk(dc.Addr, dc.Data)
	}

	startPut := time.Now()

	// put the chunks to the local store
	seen, err := provider.Put(ctx, chunks...)
	if err != nil {
		if err == storage.ErrChunkInvalid {
			streamChunkDeliveryFail.Inc(1)
			return protocols.Break(fmt.Errorf("put chunks to provider: %w", err))
		}

		return fmt.Errorf("clientHandleChunkDelivery putting chunk: %w", err)
	}

	providerPutTimer.UpdateSince(startPut)

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
			return nil
		case <-r.quit:
			// shutdown
			return nil
		case <-p.quit:
			// peer quit
			return nil
		}
	}

	return nil
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
					p.logger.Error("got an unsolicited chunk from peer", "peer", p.ShortString(), "caddr", c)
					streamChunkDeliveryFail.Inc(1)
					p.Drop("got an unsolicited chunk from peer")
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
				timer = time.NewTimer(timeouts.BatchTimeout)
			} else {
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeouts.BatchTimeout)
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
func (r *Registry) requestSubsequentRange(ctx context.Context, p *Peer, provider StreamProvider, w *want, lastIndex uint64) error {
	cur, ok := p.getCursor(w.stream)
	if !ok {
		metrics.GetOrRegisterCounter("network.stream.quit_unwanted", nil).Inc(1)
		p.logger.Debug("no longer interested in stream. quitting", "stream", w.stream)
		p.mtx.Lock()
		delete(p.openWants, w.ruid)
		p.mtx.Unlock()
		return nil
	}
	if w.head {
		if err := r.clientRequestStreamHead(ctx, p, w.stream, lastIndex+1); err != nil {
			streamRequestNextIntervalFail.Inc(1)
			return protocols.Break(fmt.Errorf("requesting next interval from peer: %w", err))
		}
	} else {
		if err := r.clientRequestStreamRange(ctx, p, provider, w.stream, cur); err != nil {
			streamRequestNextIntervalFail.Inc(1)
			return protocols.Break(fmt.Errorf("requesting next interval from peer: %w", err))
		}
	}

	return nil
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

// PeerInfo holds information about the peer and it's peers.
type PeerInfo struct {
	Base      string                       `json:"base"` // our node's base address
	Kademlia  string                       `json:"kademlia"`
	Peers     []PeerState                  `json:"peers"`
	Cursors   map[string]map[string]uint64 `json:"cursors"`
	Intervals map[string]string            `json:"intervals"`
}

// PeerState holds information about a connected peer.
type PeerState struct {
	Peer    string            `json:"peer"` // the peer address
	Cursors map[string]uint64 `json:"cursors"`
}

// PeerInfo returns a response in which the queried node's
// peer cursors and intervals are returned
func (r *Registry) PeerInfo() (*PeerInfo, error) {
	info := &PeerInfo{
		Base:    r.address.ShortUnder(),
		Cursors: make(map[string]map[string]uint64),
	}
	for name, p := range r.providers {
		info.Cursors[name] = make(map[string]uint64)
		if name != syncStreamName {
			// support only sync provider, for now
			continue
		}
		if sp, ok := p.(*syncProvider); ok {
			info.Kademlia = sp.kad.String()
		}
		for i := uint8(0); i <= chunk.MaxPO; i++ {
			key, err := p.EncodeKey(i)
			if err != nil {
				return nil, err
			}
			cursor, err := p.Cursor(key)
			if err != nil {
				return nil, err
			}
			info.Cursors[name][key] = cursor
		}
	}
	info.Intervals = make(map[string]string)
	if err := r.intervalsStore.Iterate("", func(key, value []byte) (stop bool, err error) {
		i := new(intervals.Intervals)
		if err := i.UnmarshalBinary(value); err != nil {
			return true, err
		}
		info.Intervals[string(key)] = i.String()
		return false, nil
	}); err != nil {
		return nil, err
	}
	for _, p := range r.peers {
		info.Peers = append(info.Peers, PeerState{
			Peer:    hex.EncodeToString(p.OAddr)[:16],
			Cursors: p.getCursorsCopy(),
		})
	}
	return info, nil
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

	var eg errgroup.Group
	for _, peer := range r.peers {
		peer := peer
		eg.Go(func() error {
			return peer.Stop(5 * time.Second)
		})
	}

	err := eg.Wait()
	if err != nil {
		r.logger.Error("stream closed with still active handlers")
	}

	for _, v := range r.providers {
		v.Close()
	}

	return nil
}
