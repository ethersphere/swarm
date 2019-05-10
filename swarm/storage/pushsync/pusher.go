// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package pushsync

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// DB interface implemented by localstore
type DB interface {
	// subscribe to chunk to be push synced - iterates from earliest to newest
	SubscribePush(context.Context) (<-chan storage.Chunk, func())
	// called to set a chunk as synced - and allow it to be garbage collected
	// TODO this should take ... last argument to delete many in one batch
	Set(context.Context, chunk.ModeSet, storage.Address) error
}

// Pusher takes care of the push syncing
type Pusher struct {
	store    DB                     // localstore DB
	tags     *chunk.Tags            // tags to update counts
	quit     chan struct{}          // channel to signal quitting on all loops
	pushed   map[string]*pushedItem // cache of items push-synced
	receipts chan chunk.Address     // channel to receive receipts
	ps       PubSub                 // PubSub interface to send chunks and receive receipts
}

var (
	retryInterval = 1 * time.Second // seconds to wait before retry sync
)

// pushedItem captures the info needed for the pusher about a chunk during the
// push-sync--receipt roundtrip
type pushedItem struct {
	tag    *chunk.Tag // tag for the chunk
	sentAt time.Time  // most recently sent at time
	synced bool       // set when chunk got synced
}

// New contructs a Pusher and starts up the push sync protocol
// takes
// - a DB interface to subscribe to push sync index to allow iterating over recently stored chunks
// - a pubsub interface to send chunks and receive statements of custody
// - tags that hold the several tag
func New(store DB, ps PubSub, tags *chunk.Tags) *Pusher {
	p := &Pusher{
		store:    store,
		tags:     tags,
		quit:     make(chan struct{}),
		pushed:   make(map[string]*pushedItem),
		receipts: make(chan chunk.Address),
		ps:       ps,
	}
	go p.sync()
	return p
}

// Close closes the pusher
func (p *Pusher) Close() {
	close(p.quit)
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
	var cancel, stop func()
	var ctx context.Context
	var synced []storage.Address

	// timer
	timer := time.NewTimer(0)
	defer timer.Stop()

	// register handler for pssReceiptTopic on pss pubsub
	deregister := p.ps.Register(pssReceiptTopic, false, func(msg []byte, _ *p2p.Peer) error {
		return p.handleReceiptMsg(msg)
	})
	defer deregister()

	for {
		select {

		// handle incoming chunks
		case ch, more := <-chunks:
			// if no more, set to nil and wait for timer
			if !more {
				chunks = nil
				continue
			}
			// if no need to sync this chunk then continue
			if !p.needToSync(ch) {
				continue
			}
			// send the chunk and ignore the error
			if err := p.sendChunkMsg(ch); err != nil {
				log.Warn("error sending chunk", "addr", ch.Address(), "err", err)
			}

			// handle incoming receipts
		case addr := <-p.receipts:
			log.Debug("synced", "addr", addr)
			// ignore if already received receipt
			item, found := p.pushed[addr.Hex()]
			if !found {
				log.Debug("not wanted or already got... ignore", "addr", addr)
				continue
			}
			if item.synced {
				log.Debug("just synced... ignore", "addr", addr)
				continue
			}
			// collect synced addresses
			synced = append(synced, addr)
			// set synced flag
			item.synced = true
			// increment synced count for the tag if exists
			if item.tag != nil {
				item.tag.Inc(chunk.StateSynced)
			}

			// retry interval timer triggers starting from new
		case <-timer.C:
			// TODO: implement some smart retry strategy relying on sent/synced ratio change
			// if subscribe was running, stop it
			if stop != nil {
				stop()
			}
			for _, addr := range synced {
				// set chunk status to synced, insert to db GC index
				if err := p.store.Set(context.Background(), chunk.ModeSetSync, addr); err != nil {
					log.Warn("error setting chunk to synced", "addr", addr, "err", err)
					continue
				}
				delete(p.pushed, addr.Hex())
			}
			// reset synced list
			synced = nil

			// and start iterating on Push index from the beginning
			ctx, cancel = context.WithCancel(context.Background())
			chunks, stop = p.store.SubscribePush(ctx)
			// reset timer to go off after retryInterval
			timer.Reset(retryInterval)

		case <-p.quit:
			// if there was a subscription, cancel it
			if cancel != nil {
				cancel()
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
	log.Debug("Handler", "receipt", label(receipt.Addr), "self", label(p.ps.BaseAddr()))
	p.PushReceipt(receipt.Addr)
	return nil
}

// pushReceipt just inserts the address into the channel
// it is also called by the push sync Storer if the originator and storer identical
func (p *Pusher) PushReceipt(addr []byte) {
	select {
	case p.receipts <- addr:
	case <-p.quit:
	}
}

// sendChunkMsg sends chunks to their destination
// using the PubSub interface Send method (e.g., pss neighbourhood addressing)
func (p *Pusher) sendChunkMsg(ch chunk.Chunk) error {
	cmsg := &chunkMsg{
		Origin: p.ps.BaseAddr(),
		Addr:   ch.Address()[:],
		Data:   ch.Data(),
		Nonce:  newNonce(),
	}
	msg, err := rlp.EncodeToBytes(cmsg)
	if err != nil {
		return err
	}
	log.Debug("send chunk", "addr", label(ch.Address()), "self", label(p.ps.BaseAddr()))
	return p.ps.Send(ch.Address()[:], pssChunkTopic, msg)
}

// needToSync checks if a chunk needs to be push-synced:
// * if not sent yet OR
// * if sent but more then retryInterval ago, so need resend
func (p *Pusher) needToSync(ch chunk.Chunk) bool {
	item, found := p.pushed[ch.Address().Hex()]
	// has been pushed already
	if found {
		// has synced already since subscribe called
		if item.synced {
			return false
		}
		// too early to retry
		if item.sentAt.Add(retryInterval).After(time.Now()) {
			return false
		}
		// first time encountered
	} else {
		// remember item
		tag, _ := p.tags.Get(ch.Tag())
		item = &pushedItem{
			tag: tag,
		}
		// increment SENT count on tag  if it exists
		if item.tag != nil {
			item.tag.Inc(chunk.StateSent)
		}
		// remember the item
		p.pushed[ch.Address().Hex()] = item
	}
	item.sentAt = time.Now()
	return true
}
