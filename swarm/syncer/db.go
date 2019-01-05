package syncer

import (
	"context"
	"encoding/binary"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/syndtr/goleveldb/leveldb" // "github.com/syndtr/goleveldb/leveldb/iterator"
)

// DB

const (
	batchSize  = 256 // chunk hashes to keep in memory,
	bufferSize = 16  // items to send while batch is retrieved
)

// State is the enum type for chunk states
type State = uint32

const (
	SPLIT  State = iota // chunk has been processed by filehasher/swarm safe call
	STORED              // chunk stored locally
	SENT                // chunk sent to neighbourhood
	SYNCED              // proof is received; chunk removed from sync db; chunk is available everywhere
)

var (
	retryInterval = 30 * time.Second // seconds to wait before retry sync
)

var (
	indexKey = []byte{1} // fixed key to store db storage index
	sizeKey  = []byte{2} // fixed key to store db size
)

// item holds info about a chunk, Address and Tag are exported fields so that they are
// rlp serialised for disk storage in the DB
// the rest of the fields are used for in memory cache
type item struct {
	Addr   storage.Address // chunk address
	Tag    string          // tag to track batches of chunks (allows for sync ETA)
	PO     uint            // distance
	key    []byte          // key to look up the item in the DB
	chunk  storage.Chunk   // chunk is retrieved when item is popped from batch and pushed to buffer
	sentAt time.Time       // time chunk is sent to a peer for syncing
	state  uint32          // state
}

// newkey increments the db storage index and creates a byte slice key from it
// by prefixing the binary serialisation with 0 byte
func (db *DB) newKey() []byte {
	index := atomic.AddInt64(&db.index, 1)
	key := make([]byte, 9)
	binary.BigEndian.PutUint64(key[1:], uint64(index))
	return key
}

// bytes serialises item to []byte using rlp
func (pi *item) bytes() []byte {
	buf, err := rlp.EncodeToBytes(pi)
	if err != nil {
		panic(err.Error())
	}
	return buf
}

// DB implements a persisted FIFO queue for
// - scheduling one repeatable task on each element in order of insertion
// - it iterates over items put in the db in order of storage and
// - calls a task function on them
// - it listens on a channel for asynchronous task completion signals
// - deletes completed items from storage
// - retries the task on items after a delay not shorter than retryInterval
// - it persists across sessions
// - call tags to update state counts for a tag
type DB struct {
	db         *storage.LDBDatabase // the underlying database
	tags       *tags                // tags info on processing rates
	chunkStore storage.ChunkStore   // chunkstore to get the chunks from and put into
	batchC     chan *storeBatch     // channel to pass current batch from listen loop to batch write loop
	requestC   chan struct{}        // channel to indicate request for new items for iteration
	itemsC     chan []*item         // channel to pass batch of iteration between feed loop and batch write loop
	waiting    sync.Map             // stores items in memory while waiting for proof response
	putC       chan *item           // channel to pass items to the listen loop from the Put API call
	receiptsC  chan storage.Address // channel to receive completed items
	chunkC     chan *item           // buffer fed from db iteration and consumed by the main loop
	quit       chan struct{}        // channel to signal quitting on all loops
	dbquit     chan struct{}        // channel to signal batch was written and db can be closed
	index      int64                // ever incrementing storage index
	size       int64                // number of items
	depthFunc  func() uint          // call to changed depth
	depthC     chan uint            // kademlia neighbourhood depth
}

type storeBatch struct {
	*leveldb.Batch
	toDelete []string
	new      int
}

// NewDB constructs a DB
func NewDB(dbpath string, store storage.ChunkStore, f func(storage.Chunk) error, receiptsC chan storage.Address, depthFunc func() uint) (*DB, error) {

	ldb, err := storage.NewLDBDatabase(dbpath)
	if err != nil {
		return nil, err
	}
	db := &DB{
		db:         ldb,
		chunkStore: store,
		tags:       newTags(),
		batchC:     make(chan *storeBatch, 1),
		requestC:   make(chan struct{}, 1),
		itemsC:     make(chan []*item),
		putC:       make(chan *item),
		depthC:     make(chan uint),
		depthFunc:  depthFunc,
		receiptsC:  receiptsC,
		chunkC:     make(chan *item, bufferSize),
		quit:       make(chan struct{}),
		dbquit:     make(chan struct{}),
	}
	db.index = db.getInt(indexKey)
	db.size = db.getInt(sizeKey)
	go db.listen()
	go db.feedBuffer()
	go db.writeBatches()
	go db.iter(f)
	return db, nil
}

// Put queues the item for batch insertion
func (db *DB) Put(i *item) {
	db.putC <- i
	db.tags.Inc(i.Tag, STORED)
}

// Close terminates loops by closing the quit channel
func (db *DB) Close() {
	close(db.quit)
	<-db.dbquit
	db.db.Close()
}

// Size returns the number of items written out in the DB
func (db *DB) Size() int64 {
	return db.getInt(sizeKey)
}

// listen listens until quit to put and delete events and writes them in a batch
func (db *DB) listen() {
	var depth uint = 256
	batch := &storeBatch{Batch: new(leveldb.Batch)}
	for {
		select {
		case <-db.quit:
			// make sure batch is saved to disk so as not to lose chunks
			db.batchC <- batch
			close(db.batchC)
			return

		case depth = <-db.depthC:

		case item := <-db.putC:
			// consume putC for insertion
			// we can assume no duplicates are sent

			key := db.newKey()
			batch.Put(key, item.bytes())
			batch.Put(indexKey, key[1:])
			db.size++
			batch.new++
			if item.PO >= depth {
				go func() {
					db.waiting.Store(item.Addr.Hex(), item)
					db.receiptsC <- item.Addr
				}()
				// continue
			}

		case addr := <-db.receiptsC:
			log.Warn("received receipt", "addr", label(addr[:]))
			// consume receiptsC for removal
			// potential duplicates
			v, ok := db.waiting.Load(addr.Hex())
			if !ok {
				// already deleted
				continue
			}
			//
			it := v.(*item)
			// in case we receive twice within one batch
			if it.state == SYNCED {
				continue
			}
			log.Warn("Synced", "addr", label(addr[:]))
			it.state = SYNCED
			db.tags.Inc(it.Tag, SYNCED)
			batch.Delete(it.key)
			db.size--
			batch.toDelete = append(batch.toDelete, addr.Hex())
		}
		batch.Put(sizeKey, int64ToBytes(db.size))

		// if batchwriter loop is idle, hand over the batch and creates a new one
		// if batchwriter loop is busy, keep adding to the same batch
		select {
		case db.batchC <- batch:
			batch = &storeBatch{Batch: new(leveldb.Batch)}
		default:
		}
	}
}

// writeBatches is a forever loop that updates the database in batches
// containing insertions and deletions
// whenever needed, after the update, an iteration is performed to gather items
func (db *DB) writeBatches() {
	var from uint64 // start cursor for db iteration
	// retryInterval time after retrieval, the batch start offset is reset to 0
	// if reset for all iterations then items would get retrieved spuriously
	// multiple times before retryInterval
	timer := time.NewTimer(retryInterval)
	defer timer.Stop()
	var timerC <-chan time.Time
	var requestC chan struct{}
	items := make([]*item, batchSize)

	for {
		select {
		case batch := <-db.batchC:
			if batch == nil {
				close(db.dbquit)
				return
			}
			// consume batches passed by listener if there were new updates
			// write out the batch of updates collected
			err := db.db.Write(batch.Batch)
			if err != nil {
				panic(err.Error())
			}
			// delete items from the waiting list
			for addr := range batch.toDelete {
				db.waiting.Delete(addr)
			}
			// if there are new items, allow iter batch requests
			if batch.new > 0 {
				requestC = db.requestC
			}
			continue
		case <-requestC:
			// accept requests to gather items

		case <-timerC:
			// start index from is reset to 0 if retryInterval time passed
			from = 0
			timerC = nil
			requestC = db.requestC
			continue
		}

		origfrom := from
		size := 0
		// retrieve maximum batchSize items from db in order of storage
		// starting from index from
		from = db.iterate(from, func(val *item) bool {
			// every deserialised item is put in the batch
			items[size] = val
			// entry is created in the waiting map
			db.waiting.Store(val.Addr.Hex(), val)
			// increment size and return when the batch is filled
			size++
			return size < batchSize
		})
		// signal done with effective number of items retrieved
		if size == 0 {
			db.requestC <- struct{}{} // this cannot block
			requestC = nil
			if origfrom == 0 {
				timerC = nil
			} else {
				timerC = timer.C
			}
			continue
		}
		if origfrom == 0 {
			timer.Reset(retryInterval)
			timerC = timer.C
		}
		db.itemsC <- items[:size]
		from++
	}
}

// forever loop feeds into a buffered channel through batches of DB retrievals
// - batchSize determines the maximum amount of hashes read from the database in one go
// - bufferSize determines the maximum number of chunks to consume while a batch is being retrieved
// - retryInterval is the period we wait for the storage proof of a chunk before we retry syncing it
func (db *DB) feedBuffer() {
	var items []*item
	// feed buffer from DB
	for {
		// if read batch is fully written to the buffer, i.e., loop var i reaches size
		// reread into iter batch
		db.requestC <- struct{}{} // does not block
		items = <-db.itemsC       // blocks until there are items to be read
		log.Trace("reading batch to buffer", "size", len(items))

		for _, next := range items {
			if err := db.getChunk(next); err != nil {
				log.Warn("chunk not found ... skipping and removing")
				db.receiptsC <- next.Addr
				continue
			}

			// feed the item to the chunk buffer of the db
			select {
			case db.chunkC <- next:
			case <-db.quit:
				// closing the buffer so that iter loop is terminated
				close(db.chunkC)
				return
			}
		}
	}
}

// getChunk fills in the chunk field of the item using chunkStore to retrieve by address
func (db *DB) getChunk(next *item) error {
	if next.chunk != nil {
		return nil
	}
	// retrieve the corresponding chunk if chunkStore is given
	if db.chunkStore != nil {
		chunk, err := db.chunkStore.Get(context.TODO(), next.Addr)
		if err != nil {
			return err
		}
		next.chunk = chunk
		return nil
	}
	// otherwise create a fake chunk with only an address
	next.chunk = storage.NewChunk(next.Addr, nil)
	return nil
}

// iter consumes chunks from the buffer and calls f on the chunk
// if last called more than retryInterval time ago
func (db *DB) iter(f func(storage.Chunk) error) {
	for c := range db.chunkC {
		addr := c.Addr
		val, ok := db.waiting.Load(addr.Hex())
		// skip if deleted
		if !ok {
			continue
		}
		c = val.(*item)
		state := State(atomic.LoadUint32(&c.state))
		// if deleted since retrieved or already asked not too long ago
		if state == SYNCED || state == SENT && c.sentAt.Add(retryInterval).After(time.Now()) {
			continue
		}
		if c.chunk == nil {
			err := db.getChunk(c)
			if err != nil {
				continue
			}
		}
		if state != SENT {
			c.state = SENT
			db.tags.Inc(c.Tag, SENT)
		}
		f(c.chunk)
		c.sentAt = time.Now()
		// not to waste memory
		c.chunk = nil
	}
}

// iterate iterates through the keys in order of f
func (db *DB) iterate(since uint64, f func(val *item) bool) uint64 {
	it := db.db.NewIterator()
	defer it.Release()
	sincekey := make([]byte, 9)
	binary.BigEndian.PutUint64(sincekey[1:], since)
	for ok := it.Seek(sincekey); ok; ok = it.Next() {
		key := it.Key()
		if key[0] != byte(0) {
			break
		}
		// deserialise the stored value as an item
		var val item
		err := rlp.DecodeBytes(it.Value(), &val)
		if err != nil {
			panic(err.Error())
		}
		// remember the key
		val.key = make([]byte, 9)
		copy(val.key, key)
		val.state = STORED
		since = binary.BigEndian.Uint64(key[1:])
		// call the function on the value, continue if it returns true
		if !f(&val) {
			break
		}
	}
	return since
}

// getInt retrieves a counter from the db, and deserialises it as int64
// used for storage index and entry count
func (db *DB) getInt(key []byte) int64 {
	b, err := db.db.Get(key)
	if err != nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(b))
}

// int64ToBytes serialises an int64 to bytes using bigendian
func int64ToBytes(n int64) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, uint64(n))
	return key
}
