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
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/storage"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 3, "verbosity of logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

// loopback implements PubSub as a central subscription engine,
// ie a msg sent is received by all handlers registered for the topic
type loopBack struct {
	async    bool
	addr     []byte
	handlers map[string][]func(msg []byte, p *p2p.Peer) error
}

func newLoopBack(async bool) *loopBack {
	return &loopBack{
		async:    async,
		addr:     make([]byte, 32),
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
		go lb.send(to, topic, msg)
		return nil
	}
	return lb.send(to, topic, msg)
}

func (lb *loopBack) send(to []byte, topic string, msg []byte) error {
	p := p2p.NewPeer(enode.ID{}, "", nil)
	for _, handler := range lb.handlers[topic] {
		if err := handler(msg, p); err != nil {
			return err
		}
	}
	return nil
}

// BaseAddr needed to implement PubSub interface
func (lb *loopBack) BaseAddr() []byte {
	return lb.addr
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
func (t *testPushSyncIndex) SubscribePush(context.Context) (<-chan storage.Chunk, func()) {
	chunks := make(chan storage.Chunk)
	tagCnt := len(t.tagIDs)
	quit := make(chan struct{})
	stop := func() { close(quit) }
	go func() {
		// feed fake chunks into the db, hashes encode the order so that
		// it can be traced
		feed := func(i int) bool {
			// generate fake hashes that encode the chunk order
			addr := make([]byte, 32)
			binary.BigEndian.PutUint64(addr, uint64(i))
			// remember when the chunk was put
			// if sent again, dont modify the time
			t.sent.Store(i, time.Now())
			// increment stored count on tag
			tagID := t.tagIDs[i%tagCnt]
			if tag, _ := t.tags.Get(tagID); tag != nil {
				tag.Inc(chunk.StateStored)
			}
			select {
			// chunks have no data and belong to tag i%tagCount
			case chunks <- storage.NewChunk(addr, nil).WithTagID(tagID):
				return true
			case <-quit:
				return false
			}
		}
		// push the chunks already pushed but not yet synced
		t.sent.Range(func(k, _ interface{}) bool {
			log.Debug("resending", "cur", k)
			return feed(k.(int))
		})
		// generate the new chunks from t.i
		for t.i < t.total && feed(t.i) {
			t.i++
		}

		log.Debug("sent all chunks", "total", t.total)
		close(chunks)
	}()
	return chunks, stop
}

func (t *testPushSyncIndex) Set(ctx context.Context, _ chunk.ModeSet, addrs ...storage.Address) error {
	for _, addr := range addrs {
		cur := int(binary.BigEndian.Uint64(addr[:8]))
		t.sent.Delete(cur)
		t.synced <- cur
		log.Debug("set chunk synced", "cur", cur, "addr", addr)
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
// - repeated sending is attempted only if retryInterval time passed
// - already synced chunks are not resynced
// - if no more data inserted, the db is emptied shortly
func TestPusher(t *testing.T) {

	timeout := 10 * time.Second
	chunkCnt := 200
	tagCnt := 4

	errc := make(chan error)
	sent := &sync.Map{}
	sendTimes := make(map[int]time.Time)
	synced := make(map[int]int)
	quit := make(chan struct{})
	defer close(quit)

	errf := func(s string, vals ...interface{}) {
		select {
		case errc <- fmt.Errorf(s, vals...):
		case <-quit:
		}
	}

	ps := newLoopBack(false)

	max := 0 // the highest index sent so far
	respond := func(msg []byte, _ *p2p.Peer) error {
		chmsg, err := decodeChunkMsg(msg)
		if err != nil {
			errf("error decoding chunk message: %v", err)
			return nil
		}
		// check outgoing chunk messages
		cur := int(binary.BigEndian.Uint64(chmsg.Addr[:8]))
		if cur > max {
			errf("incorrect order of chunks from db chunk #%d before #%d", cur, max)
			return nil
		}
		v, found := sent.Load(cur)
		previouslySentAt, repeated := sendTimes[cur]
		if !found {
			if !repeated {
				errf("chunk #%d not sent but received", cur)
			}
			return nil
		}
		sentAt := v.(time.Time)
		if repeated {
			// expect at least retryInterval since previous push
			if expectedAt := previouslySentAt.Add(retryInterval); expectedAt.After(sentAt) {
				errf("resync chunk #%d too early. previously sent at %v, next at %v < expected at %v", cur, previouslySentAt, sentAt, expectedAt)
				return nil
			}
		}
		// remember the latest time sent
		sendTimes[cur] = sentAt
		max++
		// respond ~ mock storer protocol
		go func() {
			receipt := &receiptMsg{Addr: chmsg.Addr}
			rmsg, err := rlp.EncodeToBytes(receipt)
			if err != nil {
				errf("error encoding receipt message: %v", err)
			}
			log.Debug("chunk sent", "addr", hex.EncodeToString(receipt.Addr))
			// random delay to allow retries
			if !delayResponse() {
				log.Debug("chunk/receipt lost", "addr", hex.EncodeToString(receipt.Addr))
				return
			}
			log.Debug("store chunk,  send receipt", "addr", hex.EncodeToString(receipt.Addr))
			err = ps.Send(chmsg.Origin, pssReceiptTopic, rmsg)
			if err != nil {
				errf("error sending receipt message: %v", err)
			}
		}()
		return nil
	}
	// register the respond function
	ps.Register(pssChunkTopic, false, respond)
	tags, tagIDs := setupTags(chunkCnt, tagCnt)
	// construct the mock push sync index iterator
	tp := newTestPushSyncIndex(chunkCnt, tagIDs, tags, sent)
	// start push syncing in a go routine
	kad := network.NewKademlia(nil, network.NewKadParams())
	p := NewPusher(tp, ps, tags, kad)
	defer p.Close()
	// collect synced chunks until all chunks synced
	// wait on errc for errors on any thread
	// otherwise time out
	for {
		select {
		case i := <-tp.synced:
			sent.Delete(i)
			n := synced[i]
			synced[i] = n + 1
			if len(synced) == chunkCnt {
				expTotal := chunkCnt / tagCnt
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
		tags.Create("", int64(chunkCnt/tagCnt))
	}
	// extract tag ids
	tags.Range(func(k, _ interface{}) bool {
		tagIDs = append(tagIDs, k.(uint32))
		return true
	})
	// add an extra for which no tag exists
	return tags, append(tagIDs, 0)
}

func checkTags(t *testing.T, expTotal int, tagIDs []uint32, tags *chunk.Tags) {
	for _, tagID := range tagIDs {
		tag, err := tags.Get(tagID)
		if err != nil {
			t.Fatalf("expected no error getting tag '%v', got %v", tagID, err)
		}
		n, total, err := tag.Status(chunk.StateSent)
		if err != nil {
			t.Fatalf("getting status for tag '%v', expected no error, got %v", tagID, err)
		}
		if int(n) != expTotal {
			t.Fatalf("expected Sent count on tag '%v' to be %v, got %v", tagID, expTotal, n)
		}
		if int(total) != expTotal {
			t.Fatalf("expected Sent count on tag '%v' to be %v, got %v", tagID, expTotal, n)
		}
		n, total, err = tag.Status(chunk.StateSynced)
		if err != nil {
			t.Fatalf("getting status for tag '%v', expected no error, got %v", tagID, err)
		}
		if int(n) != expTotal {
			t.Fatalf("expected Sent count on tag '%v' to be %v, got %v", tagID, expTotal, n)
		}
		if int(total) != expTotal {
			t.Fatalf("expected Sent count on tag '%v' to be %v, got %v", tagID, expTotal, n)
		}
	}
}

type testStore struct {
	store *sync.Map
}

func (t *testStore) Put(_ context.Context, _ chunk.ModePut, ch chunk.Chunk) (bool, error) {
	cur := binary.BigEndian.Uint64(ch.Address()[:8])
	var storedCnt uint32 = 1
	v, loaded := t.store.LoadOrStore(cur, &storedCnt)
	if loaded {
		atomic.AddUint32(v.(*uint32), 1)
	}
	return false, nil
}

// TestPushSyncAndStoreWithLoopbackPubSub tests the push sync protocol
// push syncer node communicate with storers via mock PubSub
func TestPushSyncAndStoreWithLoopbackPubSub(t *testing.T) {
	timeout := 10 * time.Second
	chunkCnt := 2000
	tagCnt := 4
	storerCnt := 3
	sent := &sync.Map{}
	store := &sync.Map{}
	// mock pubsub messenger
	ps := newLoopBack(true)

	tags, tagIDs := setupTags(chunkCnt, tagCnt)
	// construct the mock push sync index iterator
	tp := newTestPushSyncIndex(chunkCnt, tagIDs, tags, sent)
	// neighbourhood function mocked
	nnf := func(storage.Address) bool { return true }
	// start push syncing in a go routine
	p := NewPusher(tp, ps, tags, nnf)
	defer p.Close()

	// set up a number of storers
	storers := make([]*Storer, storerCnt)
	for i := 0; i < storerCnt; i++ {
		storers[i] = NewStorer(&testStore{store}, ps, nnf, p.pushReceipt)
	}

	synced := 0
	for {
		select {
		case i := <-tp.synced:
			synced++
			sent.Delete(i)
			if synced == chunkCnt {
				expTotal := chunkCnt / tagCnt
				checkTags(t, expTotal, tagIDs[:tagCnt-1], tags)
				for i := uint64(0); i < uint64(chunkCnt); i++ {
					v, ok := store.Load(i)
					if !ok {
						t.Fatalf("chunk %v not stored", i)
					}
					if cnt := *(v.(*uint32)); cnt != uint32(storerCnt) {
						t.Fatalf("chunk %v expected to be saved %v times, got %v", i, storerCnt, cnt)
					}
				}
				return
			}
		case <-time.After(timeout):
			t.Fatalf("timeout waiting for all chunks to be synced")
		}
	}

}
