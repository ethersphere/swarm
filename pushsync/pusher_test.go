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
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 0, "verbosity of logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

type testPubSub struct {
	*loopBack
	isClosestTo func([]byte) bool
}

var testBaseAddr = make([]byte, 32)

// BaseAddr needed to implement PubSub interface
// in the testPubSub, this address has no relevant and is given only for logging
func (tps *testPubSub) BaseAddr() []byte {
	return testBaseAddr
}

// IsClosestTo needed to implement PubSub interface
func (tps *testPubSub) IsClosestTo(addr []byte) bool {
	return tps.isClosestTo(addr)
}

// loopback implements PubSub as a central subscription engine,
// ie a msg sent is received by all handlers registered for the topic
type loopBack struct {
	async    bool
	handlers map[string][]func(msg []byte, p *p2p.Peer) error
}

func newLoopBack(async bool) *loopBack {
	return &loopBack{
		async:    async,
		handlers: make(map[string][]func(msg []byte, p *p2p.Peer) error),
	}
}

// Register subscribes to a topic with a handler
func (lb *loopBack) Register(topic string, _ bool, handler func(msg []byte, p *p2p.Peer) error) func() {
	lb.handlers[topic] = append(lb.handlers[topic], handler)
	return func() {}
}

// Send publishes a msg with a topic and directly calls registered handlers with
// that topic
func (lb *loopBack) Send(to []byte, topic string, msg []byte) error {
	if lb.async {
		go func() {
			if !delayResponse() {
				return
			}
			lb.send(to, topic, msg)
		}()
		return nil
	}
	return lb.send(to, topic, msg)
}

func (lb *loopBack) send(to []byte, topic string, msg []byte) error {
	p := p2p.NewPeer(enode.ID{}, "", nil)
	for _, handler := range lb.handlers[topic] {
		log.Debug("handling message", "topic", topic, "to", hex.EncodeToString(to))
		if err := handler(msg, p); err != nil {
			log.Error("error handling message", "topic", topic, "to", hex.EncodeToString(to))
			return err
		}
	}
	return nil
}

// testPushSyncIndex mocks localstore and provides subscription and setting synced status
// it implements the DB interface
type testPushSyncIndex struct {
	i, total int
	tagIDs   []uint32 //
	tags     *chunk.Tags
	sent     *sync.Map // to store time of send for retry
	synced   chan int  // to check if right amount of chunks
}

func newTestPushSyncIndex(chunkCnt int, tagIDs []uint32, tags *chunk.Tags, sent *sync.Map) *testPushSyncIndex {
	return &testPushSyncIndex{
		i:      0,
		total:  chunkCnt,
		tagIDs: tagIDs,
		tags:   tags,
		sent:   sent,
		synced: make(chan int),
	}
}

// SubscribePush allows iteration on the hashes and mocks the behaviour of localstore
// push index
// we keep track of an index so that each call to SubscribePush knows where to start
// generating the new fake hashes
// Before the new fake hashes it iterates over hashes not synced yet
func (tp *testPushSyncIndex) SubscribePush(context.Context) (<-chan storage.Chunk, func()) {
	chunks := make(chan storage.Chunk)
	tagCnt := len(tp.tagIDs)
	quit := make(chan struct{})
	stop := func() { close(quit) }
	go func() {
		// feed fake chunks into the db, hashes encode the order so that
		// it can be traced
		feed := func(i int) bool {
			// generate fake hashes that encode the chunk order
			addr := make([]byte, 32)
			binary.BigEndian.PutUint64(addr, uint64(i))
			tagID := tp.tagIDs[i%tagCnt]
			// remember when the chunk was put
			// if sent again, dont modify the time
			_, loaded := tp.sent.LoadOrStore(i, time.Now())
			if !loaded {
				// increment stored count on tag
				if tag, _ := tp.tags.Get(tagID); tag != nil {
					tag.Inc(chunk.StateStored)
				}
			}
			tp.sent.Store(i, time.Now())
			select {
			// chunks have no data and belong to tag i%tagCount
			case chunks <- storage.NewChunk(addr, nil).WithTagID(tagID):
				return true
			case <-quit:
				return false
			}
		}
		// push the chunks already pushed but not yet synced
		tp.sent.Range(func(k, _ interface{}) bool {
			log.Debug("resending", "idx", k)
			return feed(k.(int))
		})
		// generate the new chunks from tp.i
		for tp.i < tp.total && feed(tp.i) {
			tp.i++
		}
		log.Debug("sent chunks", "sent", tp.i, "total", tp.total)
		close(chunks)
	}()
	return chunks, stop
}

func (tp *testPushSyncIndex) Set(ctx context.Context, _ chunk.ModeSet, addrs ...storage.Address) error {
	for _, addr := range addrs {
		idx := int(binary.BigEndian.Uint64(addr[:8]))
		tp.sent.Delete(idx)
		tp.synced <- idx
		log.Debug("set chunk synced", "idx", idx, "addr", addr)
	}
	return nil
}

var (
	maxDelay       = 210 // max delay in millisecond
	minDelay       = 1   // min delay in millisecond
	retentionLimit = 200 // ~5% of msg lost
)

// delayResponse when called mock connection/throughput
func delayResponse() bool {
	delay := rand.Intn(maxDelay) + minDelay
	time.Sleep(time.Duration(delay) * time.Millisecond)
	return delay < retentionLimit
}

// TestPusher tests the correct behaviour of Pusher
// in the context of inserting n chunks
// receipt response model: the pushed chunk's receipt is sent back
// after a random delay
// The test checks:
// - if sync function is called on chunks in order of insertion (FIFO)
// - already synced chunks are not resynced
// - if no more data inserted, the db is emptied shortly
func TestPusher(t *testing.T) {
	timeout := 10 * time.Second
	chunkCnt := 1024
	tagCnt := 4

	errc := make(chan error)
	sent := &sync.Map{}
	synced := make(map[int]int)
	quit := make(chan struct{})
	defer close(quit)

	errf := func(s string, vals ...interface{}) {
		select {
		case errc <- fmt.Errorf(s, vals...):
		case <-quit:
		}
	}

	lb := newLoopBack(false)

	max := 0 // the highest index sent so far
	respond := func(msg []byte, _ *p2p.Peer) error {
		chmsg, err := decodeChunkMsg(msg)
		if err != nil {
			errf("error decoding chunk message: %v", err)
			return nil
		}
		// check outgoing chunk messages
		idx := int(binary.BigEndian.Uint64(chmsg.Addr[:8]))
		if idx > max {
			errf("incorrect order of chunks from db chunk #%d before #%d", idx, max)
			return nil
		}
		max++
		// respond ~ mock storer protocol
		go func() {
			receipt := &receiptMsg{Addr: chmsg.Addr}
			rmsg, err := rlp.EncodeToBytes(receipt)
			if err != nil {
				errf("error encoding receipt message: %v", err)
			}
			log.Debug("chunk sent", "idx", idx)
			// random delay to allow retries
			if !delayResponse() {
				log.Debug("chunk/receipt lost", "idx", idx)
				return
			}
			log.Debug("store chunk,  send receipt", "idx", idx)
			err = lb.Send(chmsg.Origin, pssReceiptTopic, rmsg)
			if err != nil {
				errf("error sending receipt message: %v", err)
			}
		}()
		return nil
	}
	// register the respond function
	lb.Register(pssChunkTopic, false, respond)
	tags, tagIDs := setupTags(chunkCnt, tagCnt)
	// construct the mock push sync index iterator
	tp := newTestPushSyncIndex(chunkCnt, tagIDs, tags, sent)
	// start push syncing in a go routine
	p := NewPusher(tp, &testPubSub{lb, func([]byte) bool { return false }}, tags)
	defer p.Close()
	// collect synced chunks until all chunks synced
	// wait on errc for errors on any thread
	// otherwise time out
	for {
		select {
		case i := <-tp.synced:
			n := synced[i]
			synced[i] = n + 1
			if len(synced) == chunkCnt {
				expTotal := int64(chunkCnt / tagCnt)
				checkTags(t, expTotal, tagIDs[:tagCnt-1], tags)
				return
			}
		case err := <-errc:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(timeout):
			t.Fatalf("timeout waiting for all chunks to be synced")
		}
	}

}

// setupTags constructs tags object create tagCnt - 1 tags
// the sequential fake chunk i will be tagged with i%tagCnt
func setupTags(chunkCnt, tagCnt int) (tags *chunk.Tags, tagIDs []uint32) {
	// construct tags object
	tags = chunk.NewTags()
	// all but one tag is created
	for i := 0; i < tagCnt-1; i++ {
		tags.Create(context.Background(), "", int64(chunkCnt/tagCnt))
	}
	// extract tag ids
	tags.Range(func(k, _ interface{}) bool {
		tagIDs = append(tagIDs, k.(uint32))
		return true
	})
	// add an extra for which no tag exists
	return tags, append(tagIDs, 0)
}

func checkTags(t *testing.T, expTotal int64, tagIDs []uint32, tags *chunk.Tags) {
	t.Helper()
	for _, tagID := range tagIDs {
		tag, err := tags.Get(tagID)
		if err != nil {
			t.Fatal(err)
		}
		// the tag is adjusted after the store.Set calls show
		err = tag.WaitTillDone(context.Background(), chunk.StateSynced)
		if err != nil {
			t.Fatalf("error waiting for syncing on tag %v: %v", tag.Uid, err)
		}

		testutil.CheckTag(t, tag, 0, expTotal, 0, expTotal, expTotal, expTotal)
	}
}
