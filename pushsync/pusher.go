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

package pushsync

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/spancontext"
	"github.com/ethersphere/swarm/storage"
	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
)

// DB interface implemented by localstore
type DB interface {
	// subscribe to chunk to be push synced - iterates from earliest to newest
	SubscribePush(context.Context) (<-chan storage.Chunk, func())
	// called to set a chunk as synced - and allow it to be garbage collected
	// TODO this should take ... last argument to delete many in one batch
	Set(context.Context, chunk.ModeSet, ...storage.Address) error
}

// Pusher takes care of the push syncing
type Pusher struct {
	store         DB                     // localstore DB
	tags          *chunk.Tags            // tags to update counts
	quit          chan struct{}          // channel to signal quitting on all loops
	closed        chan struct{}          // channel to signal sync loop terminated
	pushed        map[string]*pushedItem // cache of items push-synced
	receipts      chan []byte            // channel to receive receipts
	ps            PubSub                 // PubSub interface to send chunks and receive receipts
	logger        log.Logger             // custom logger
	retryInterval time.Duration          // dynamically adjusted time interval between retries
}

const maxMeasurements = 20000

// pushedItem captures the info needed for the pusher about a chunk during the
// push-sync--receipt roundtrip
type pushedItem struct {
	tag         *chunk.Tag       // tag for the chunk
	shortcut    bool             // if the chunk receipt was sent by self
	firstSentAt time.Time        // first sent at time
	lastSentAt  time.Time        // most recently sent at time
	synced      bool             // set when chunk got synced
	span        opentracing.Span // roundtrip span
}

// NewPusher constructs a Pusher and starts up the push sync protocol
// takes
// - a DB interface to subscribe to push sync index to allow iterating over recently stored chunks
// - a pubsub interface to send chunks and receive statements of custody
// - tags that hold the tags
func NewPusher(store DB, ps PubSub, tags *chunk.Tags) *Pusher {
	p := &Pusher{
		store:         store,
		tags:          tags,
		quit:          make(chan struct{}),
		closed:        make(chan struct{}),
		pushed:        make(map[string]*pushedItem),
		receipts:      make(chan []byte),
		ps:            ps,
		logger:        log.New("self", label(ps.BaseAddr())),
		retryInterval: 3 * time.Second,
	}
	go p.sync()
	return p
}

// Close closes the pusher
func (p *Pusher) Close() {
	close(p.quit)
	select {
	case <-p.closed:
	case <-time.After(3 * time.Second):
		log.Error("timeout closing pusher")
	}
}

// sync starts a forever loop that pushes chunks to their neighbourhood
// and receives receipts (statements of custody) for them.
// chunks that are not acknowledged with a receipt are retried
// not earlier than retryInterval after they were last pushed
// the routine also updates counts of states on a tag in order
// to monitor the proportion of saved, sent and synced chunks of
// a file or collection
func (p *Pusher) sync() {
	var chunks <-chan chunk.Chunk
	var unsubscribe func()
	var syncedAddrs []storage.Address
	defer close(p.closed)
	// timer, initially set to 0 to fall through select case on timer.C for initialisation
	timer := time.NewTimer(0)
	defer timer.Stop()

	// register handler for pssReceiptTopic on pss pubsub
	deregister := p.ps.Register(pssReceiptTopic, false, func(msg []byte, _ *p2p.Peer) error {
		return p.handleReceiptMsg(msg)
	})
	defer deregister()

	chunksInBatch := -1
	var batchStartTime time.Time
	ctx := context.Background()

	var average uint64 = 100000 // microseconds
	var measurements uint64

	for {
		select {
		// handle incoming chunks
		case ch, more := <-chunks:
			// if no more, set to nil, reset timer to 0 to finalise batch immediately
			if !more {
				chunks = nil
				var dur time.Duration
				if chunksInBatch == 0 {
					dur = 500 * time.Millisecond
				}
				timer.Reset(dur)
				break
			}

			chunksInBatch++
			metrics.GetOrRegisterCounter("pusher.send-chunk", nil).Inc(1)
			// if no need to sync this chunk then continue
			if !p.needToSync(ch) {
				break
			}

			metrics.GetOrRegisterCounter("pusher.send-chunk.send-to-sync", nil).Inc(1)
			// send the chunk and ignore the error
			go func(ch chunk.Chunk) {
				if err := p.sendChunkMsg(ch); err != nil {
					p.logger.Error("error sending chunk", "addr", ch.Address().Hex(), "err", err)
				}
			}(ch)

		// handle incoming receipts
		case addr := <-p.receipts:
			hexaddr := hex.EncodeToString(addr)
			p.logger.Trace("got receipt", "addr", hexaddr)
			metrics.GetOrRegisterCounter("pusher.receipts.all", nil).Inc(1)
			// ignore if already received receipt
			item, found := p.pushed[hexaddr]
			if !found {
				metrics.GetOrRegisterCounter("pusher.receipts.not-found", nil).Inc(1)
				p.logger.Trace("not wanted or already got... ignore", "addr", hexaddr)
				break
			}
			if item.synced { // already got receipt in this same batch
				metrics.GetOrRegisterCounter("pusher.receipts.already-synced", nil).Inc(1)
				p.logger.Trace("just synced... ignore", "addr", hexaddr)
				break
			}
			// increment synced count for the tag if exists
			tag := item.tag
			if tag != nil {
				tag.Inc(chunk.StateSynced)
				if tag.Done(chunk.StateSynced) {
					p.logger.Debug("closing root span for tag", "taguid", tag.Uid, "tagname", tag.Name)
					tag.FinishRootSpan()
				}
				// finish span for pushsync roundtrip, only have this span if we have a tag
				item.span.Finish()
			}

			totalDuration := time.Since(item.firstSentAt)
			metrics.GetOrRegisterResettingTimer("pusher.chunk.roundtrip", nil).Update(totalDuration)
			metrics.GetOrRegisterCounter("pusher.receipts.synced", nil).Inc(1)

			// calibrate retryInterval based on roundtrip times
			measurements, average = p.updateRetryInterval(item, measurements, average)

			// collect synced addresses and corresponding items to do subsequent batch operations
			syncedAddrs = append(syncedAddrs, addr)
			item.synced = true

			// retry interval timer triggers starting from new
		case <-timer.C:
			// initially timer is set to go off as well as every time we hit the end of push index
			// so no wait for retryInterval needed to set  items synced
			metrics.GetOrRegisterCounter("pusher.subscribe-push", nil).Inc(1)
			// if subscribe was running, stop it
			if unsubscribe != nil {
				unsubscribe()
			}

			// delete from pushed items
			for i := 0; i < len(syncedAddrs); i++ {
				delete(p.pushed, syncedAddrs[i].Hex())
			}
			// set chunk status to synced, insert to db GC index
			if err := p.store.Set(ctx, chunk.ModeSetSync, syncedAddrs...); err != nil {
				log.Error("pushsync: error setting chunks to synced", "err", err)
			}

			// reset synced list
			syncedAddrs = nil

			// we don't want to record the first iteration
			if chunksInBatch != -1 {
				// this measurement is not a timer, but we want a histogram, so it fits the data structure
				metrics.GetOrRegisterResettingTimer("pusher.subscribe-push.chunks-in-batch.hist", nil).Update(time.Duration(chunksInBatch))
				metrics.GetOrRegisterResettingTimer("pusher.subscribe-push.chunks-in-batch.time", nil).UpdateSince(batchStartTime)
				metrics.GetOrRegisterCounter("pusher.subscribe-push.chunks-in-batch", nil).Inc(int64(chunksInBatch))
			}
			chunksInBatch = 0
			batchStartTime = time.Now()

			// and start iterating on Push index from the beginning
			chunks, unsubscribe = p.store.SubscribePush(ctx)
			// reset timer to go off after retryInterval
			timer.Reset(p.retryInterval)

		case <-p.quit:
			if unsubscribe != nil {
				unsubscribe()
			}
			return
		}
	}
}

// handleReceiptMsg is a handler for pssReceiptTopic that
// - deserialises receiptMsg and
// - sends the receipted address on a channel
func (p *Pusher) handleReceiptMsg(msg []byte) error {
	receipt, err := decodeReceiptMsg(msg)
	if err != nil {
		return err
	}
	p.logger.Trace("Handler", "receipt", label(receipt.Addr))
	go p.pushReceipt(receipt.Addr)
	return nil
}

// pushReceipt just inserts the address into the channel
func (p *Pusher) pushReceipt(addr []byte) {
	select {
	case p.receipts <- addr:
	case <-p.quit:
	}
}

// sendChunkMsg sends chunks to their destination
// using the PubSub interface Send method (e.g., pss neighbourhood addressing)
func (p *Pusher) sendChunkMsg(ch chunk.Chunk) error {
	rlpTimer := time.Now()

	cmsg := &chunkMsg{
		Origin: p.ps.BaseAddr(),
		Addr:   ch.Address(),
		Data:   ch.Data(),
		Nonce:  newNonce(),
	}
	msg, err := rlp.EncodeToBytes(cmsg)
	if err != nil {
		return err
	}
	p.logger.Trace("send chunk", "addr", label(ch.Address()))

	metrics.GetOrRegisterResettingTimer("pusher.send.chunk.rlp", nil).UpdateSince(rlpTimer)

	defer metrics.GetOrRegisterResettingTimer("pusher.send.chunk.pss", nil).UpdateSince(time.Now())
	return p.ps.Send(ch.Address()[:], pssChunkTopic, msg)
}

// needToSync checks if a chunk needs to be push-synced:
// * if not sent yet OR
// * if sent but more than retryInterval ago, so need resend OR
// * if self is closest node to chunk TODO: and not light node
//   in this case send receipt to self to trigger synced state on chunk
func (p *Pusher) needToSync(ch chunk.Chunk) bool {
	item, found := p.pushed[ch.Address().Hex()]
	now := time.Now()
	// has been pushed already
	if found {
		// has synced already since subscribe called
		if item.synced {
			return false
		}
		item.lastSentAt = now
	} else {
		// first time encountered
		addr := ch.Address()
		hexaddr := addr.Hex()
		// remember item
		tag, _ := p.tags.Get(ch.TagID())
		item = &pushedItem{
			tag:         tag,
			firstSentAt: now,
			lastSentAt:  now,
		}

		// increment SENT count on tag  if it exists
		if tag != nil {
			tag.Inc(chunk.StateSent)
			// opentracing for chunk roundtrip
			_, span := spancontext.StartSpan(tag.Context(), "chunk.sent")
			span.LogFields(olog.String("ref", hexaddr))
			span.SetTag("addr", hexaddr)
			item.span = span
		}

		// remember the item
		p.pushed[hexaddr] = item
		if p.ps.IsClosestTo(addr) {
			p.logger.Trace("self is closest to ref: push receipt locally", "ref", hexaddr)
			item.shortcut = true
			go p.pushReceipt(addr)
			return false
		}
		p.logger.Trace("self is not the closest to ref: send chunk to neighbourhood", "ref", hexaddr)
	}
	return true
}

// updateRetryInterval calibrates the period after which push index iterator restart from the beginning
func (p *Pusher) updateRetryInterval(item *pushedItem, measurements uint64, average uint64) (uint64, uint64) {
	if !item.shortcut { // only real network roundtrips counted, no shortcuts
		roundtripDuration := time.Since(item.lastSentAt)
		measurement := uint64(roundtripDuration) / 1000 // in microseconds
		// recalculate average
		average = (measurements*average + measurement) / (measurements + 1)
		if measurement < maxMeasurements {
			measurements++
		}
		p.retryInterval = time.Duration(average*2) * time.Microsecond
	}
	return measurements, average
}
