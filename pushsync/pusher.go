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
	"bytes"
	"context"
	"encoding/hex"
	"sync"
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
	retryInterval time.Duration          // time interval between retries
}

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
		pushed:        make(map[string]*pushedItem),
		receipts:      make(chan []byte),
		closed:        make(chan struct{}),
		ps:            ps,
		logger:        log.New("self", label(ps.BaseAddr())),
		retryInterval: 100 * time.Millisecond,
	}
	go p.run()
	return p
}

// Close closes the pusher
func (p *Pusher) Close() {
	close(p.quit)
	select {
	case <-p.closed:
	case <-time.After(3 * time.Second):
		p.logger.Error("timeout closing pusher")
	}
}

// sync starts a forever loop that pushes chunks to their neighbourhood
// and receives receipts (statements of custody) for them.
// chunks that are not acknowledged with a receipt are retried
// not earlier than retryInterval after they were last pushed
// the routine also updates counts of states on a tag in order
// to monitor the proportion of saved, sent and synced chunks of
// a file or collection
func (p *Pusher) run() {
	var syncedAddrs []storage.Address
	// timer, initially set to 0 to fall through select case on timer.C for initialisation
	timer := time.NewTimer(p.retryInterval)
	defer timer.Stop()

	// register handler for pssReceiptTopic on pss pubsub
	deregister := p.ps.Register(pssReceiptTopic, false, func(msg []byte, _ *p2p.Peer) error {
		return p.handleReceiptMsg(msg)
	})
	defer deregister()

	// start a forever subscription to chunks to push
	var batchStartTime time.Time
	ctx := context.Background()
	newChunks, unsubscribe := p.store.SubscribePush(ctx)
	// channel for retried chunks
	var retryChunks <-chan chunk.Chunk
	var retryUnsubscribe func()
	// cursors to control retried chunks
	var first, last storage.Address

	// channel for memory items to be removed
	deleteC := make(chan []storage.Address)
	// channel used as lock for syncing
	rsyncC := make(chan struct{}, 1)
	syncC := rsyncC
	rsyncC <- struct{}{}
	// wait for go routines calling db before close
	wg := sync.WaitGroup{}
	var nextAt time.Time

	for {
		// disable case when no synced adds
		if len(syncedAddrs) == 0 {
			syncC = nil
		} else {
			syncC = rsyncC
		}

		select {
		// handle incoming new chunks
		case ch, more := <-newChunks:
			if !more {
				p.logger.Error("pushsync: chunk channel closed")
				return
			}
			metrics.GetOrRegisterCounter("pusher.send-chunk", nil).Inc(1)
			// if no need to sync this chunk then send
			if p.needToSync(ch) {
				// send the chunk and ignore the error
				go func(ch chunk.Chunk) {
					if err := p.sendChunkMsg(ch); err != nil {
						p.logger.Error("error sending chunk", "addr", ch.Address().Hex(), "err", err)
					}
				}(ch)
			}
			// remember item so that retry  iteration knows where to stop
			// should be counter on item since last can be deleted
			last = ch.Address()

			// handle incoming chunks to resend
		case ch, more := <-retryChunks:
			if !more {
				p.logger.Error("pushsync: retry chunk channel closed")
				return
			}
			waitFunc := func() (bool, time.Duration) {
				item, found := p.pushed[ch.Address().Hex()]
				if !found {
					return true, 0
				}
				if item.synced {
					return true, 0
				}
				next := item.lastSentAt.Add(p.retryInterval)
				ago := time.Since(next)
				if ago < 0 {
					return true, -ago
				}
				item.lastSentAt = time.Now()

				// send the chunk and ignore the error
				go func(ch chunk.Chunk) {
					if err := p.sendChunkMsg(ch); err != nil {
						p.logger.Error("error sending chunk", "addr", ch.Address().Hex(), "err", err)
					}
				}(ch)

				if bytes.Equal(ch.Address(), last) {
					var wait time.Duration
					if first != nil {
						ago := time.Since(nextAt)
						if ago > 0 {
							wait = -ago
						}
					}
					return true, wait
				}

				if first == nil {
					first = ch.Address()
					nextAt = time.Now().Add(p.retryInterval)
					timer.Reset(p.retryInterval)
				}
				return false, 0
			}

			if wait, duration := waitFunc(); wait {
				// if retry subscribe was running, stop it
				if retryUnsubscribe != nil {
					retryUnsubscribe()
				}
				retryChunks = nil
				retryUnsubscribe = nil
				timer.Reset(duration)
			}

		// handle incoming receipts
		case addr := <-p.receipts:
			hexaddr := hex.EncodeToString(addr)
			metrics.GetOrRegisterCounter("pusher.receipts.all", nil).Inc(1)
			// ignore if already received receipt
			item, found := p.pushed[hexaddr]
			if !found {
				metrics.GetOrRegisterCounter("pusher.receipts.not-found", nil).Inc(1)
				break
			}
			if item.synced { // already got receipt in this same batch
				metrics.GetOrRegisterCounter("pusher.receipts.already-synced", nil).Inc(1)
				break
			}
			// increment synced count for the tag if exists
			tag := item.tag
			if tag != nil {
				tag.Inc(chunk.StateSynced)
				if tag.Done(chunk.StateSynced) {
					tag.FinishRootSpan()
				}
				// finish span for pushsync roundtrip, only have this span if we have a tag
				item.span.Finish()
			}

			totalDuration := time.Since(item.firstSentAt)
			metrics.GetOrRegisterResettingTimer("pusher.chunk.roundtrip", nil).Update(totalDuration)
			metrics.GetOrRegisterCounter("pusher.receipts.synced", nil).Inc(1)
			// collect synced addresses and corresponding items to do subsequent batch operations
			syncedAddrs = append(syncedAddrs, addr)
			item.synced = true

			// retry interval timer triggers starting SubscribePush
		case <-timer.C:
			// initially timer is set to go off as well as every time we hit the end of push index
			// so no wait for retryInterval needed to set  items synced
			metrics.GetOrRegisterCounter("pusher.subscribe-push", nil).Inc(1)
			// if retry subscribe was running, stop it
			if retryUnsubscribe != nil {
				retryUnsubscribe()
			}
			retryChunks, retryUnsubscribe = p.store.SubscribePush(ctx)
			first = nil

		case <-syncC:
			wg.Add(1)
			go func(addrs []storage.Address) {
				// set chunk status to synced, insert to db GC index
				if err := p.store.Set(ctx, chunk.ModeSetSync, addrs...); err != nil {
					log.Error("pushsync: error setting chunks to synced", "err", err)
					return
				}
				wg.Done()
				rsyncC <- struct{}{}
				select {
				case deleteC <- addrs:
				case <-p.quit:
				}
			}(syncedAddrs)

			// this measurement is not a timer, but we want a histogram, so it fits the data structure
			metrics.GetOrRegisterResettingTimer("pusher.subscribe-push.chunks-in-batch.hist", nil).Update(time.Duration(len(syncedAddrs)))
			metrics.GetOrRegisterResettingTimer("pusher.subscribe-push.chunks-in-batch.time", nil).UpdateSince(batchStartTime)
			metrics.GetOrRegisterCounter("pusher.subscribe-push.chunks-in-batch", nil).Inc(int64(len(syncedAddrs)))

			// reset synced list
			syncedAddrs = nil
			batchStartTime = time.Now()

		case addrs := <-deleteC:
			// delete from pushed items
			for i := 0; i < len(addrs); i++ {
				delete(p.pushed, addrs[i].Hex())
			}

		case <-p.quit:
			wg.Wait()
			close(p.closed)
			if unsubscribe != nil {
				unsubscribe()
			}
			if retryUnsubscribe != nil {
				retryUnsubscribe()
			}
			return
		}
	}
}

// handleReceiptMsg is a handler for pssReceiptTopic that
// - deserialises receiptMsg and
// - sends the receipted address on a channel
// since message handling is asynchronous no need to pushReceipt in go routine
func (p *Pusher) handleReceiptMsg(msg []byte) error {
	receipt, err := decodeReceiptMsg(msg)
	if err != nil {
		return err
	}
	p.logger.Trace("handler", "receipt", label(receipt.Addr))
	p.pushReceipt(receipt.Addr)
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
// if self is closest node to chunk TODO: and not light node
//   in this case send receipt to self to trigger synced state on chunk
func (p *Pusher) needToSync(ch chunk.Chunk) bool {
	item := p.addPushed(ch)
	addr := ch.Address()
	if p.ps.IsClosestTo(addr) {
		p.logger.Trace("self is closest to ref: push receipt locally", "ref", addr.Hex())
		item.shortcut = true
		go p.pushReceipt(addr)
		return false
	}
	return true
}

// addPushed remembers a new chunk sent
func (p *Pusher) addPushed(ch chunk.Chunk) *pushedItem {
	addr := ch.Address()
	hexaddr := addr.Hex()
	tag, _ := p.tags.Get(ch.TagID())
	now := time.Now()
	item := &pushedItem{
		tag:         tag,
		firstSentAt: now,
		lastSentAt:  now,
	}
	// increment SENT count on tag if it exists
	if tag != nil {
		tag.Inc(chunk.StateSent)
		// opentracing for pushsync roundtrip as span within tag context
		_, span := spancontext.StartSpan(tag.Context(), "chunk.sent")
		span.LogFields(olog.String("ref", hexaddr))
		span.SetTag("addr", hexaddr)
		item.span = span
	}
	p.pushed[hexaddr] = item
	return item
}
