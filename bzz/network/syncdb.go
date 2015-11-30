package network

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/bzz/storage"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

/*
syncDb is a queueing service for outgoing deliveries

once its in-memory buffer is full it reverts to persisting in db
when in db mode a db iterator iterates through the items keeping their order
once the db read catches up (there is no more items in the db), then
it switches back to im-memory buffer.
*/
type syncDb struct {
	start          []byte               // this syncdb starting index in requestdb
	key            storage.Key          // remote peers address key
	counterKey     []byte               // db key to persist counter
	priority       uint                 // priotity High|Medium|Low
	counter        uint64               // incrementing index to enforce order
	buffer         chan interface{}     // incoming request channel
	db             *storage.LDBDatabase // underlying db (should be interface)
	done           chan bool            // chan to signal  quit for goroutines
	quit           chan bool            // chan to signal  quit for goroutines
	total, dbTotal int                  // counts for one session
	save           chan *syncDbEntry
	batch          chan chan error
}

const dbBatchSize = 1000

var (
	counterKeyPrefix = byte(1)
)

// constructor needs a shared request db (leveldb)
// priority is used in the index key
// uses a buffer and a leveldb for persistent storage
func newSyncDb(db *storage.LDBDatabase, key storage.Key, priority uint, bufferSize uint, deliver func(interface{}, chan bool) bool) *syncDb {
	counterKey := make([]byte, 33)
	counterKey[0] = counterKeyPrefix
	copy(counterKey[1:], key)

	data, err := db.Get(counterKey)
	var counter uint64
	if err == nil {
		counter = binary.LittleEndian.Uint64(data)
	}
	start := make([]byte, 42)
	start[1] = byte(priorities - priority)
	copy(start[2:], key)

	syncdb := &syncDb{
		start:      start,
		key:        key,
		priority:   priority,
		counter:    counter,
		counterKey: counterKey,
		buffer:     make(chan interface{}, bufferSize),
		db:         db,
		quit:       make(chan bool),
		done:       make(chan bool),
		save:       make(chan *syncDbEntry),
		batch:      make(chan chan error),
	}
	glog.V(logger.Debug).Infof("[BZZ] syncDb[%v] - initialised", priority)

	go syncdb.handle(deliver)
	go syncdb.batchServer()
	return syncdb
}

// handle is a forever iterator loop that reads from incoming buffer
// and switches to persistent db storage upon buffer contention
// upon clearing db backlog it switches back to in-memory buffer
// automatically started when syncdb is initialised
// saves the buffer to db upon receiving quit signal
func (self *syncDb) handle(deliver func(interface{}, chan bool) bool) {
	var usedb bool
	var n, t, b, d int
	var queue chan interface{}
	quit := self.quit
	done := self.done
	var more bool
LOOP:
	for {
		var req interface{}
		// waiting for item on the relevant queue

		select {
		// this select case reads from the buffer and calls deliver if in buffering mode
		// or saves the request to db if not
		// quits the loop if queue is closed
		case req, more = <-queue:
			if !more {
				break LOOP
			}
			t++
			// if in db persisting mode, then saves the request
			// if first item saved (b==0) and not stopping, launches a db iterator
			// with the deliver function in a parallel go-routine
			if usedb {
				self.put(req)
				if b == 0 {
					go self.iterate(deliver)
				}
				b++
				d++
				continue
			}
			// deliver request : this is blocking on network write so
			// it is passed the quit channel as argument, so that it returns
			// if syncdb is stopped
			more = deliver(req, quit)
			if !more {
				// received quit signal, save request currently waiting delivery
				self.put(req)
				// so that it is only called once
				quit = nil
				// close buffer so loop knows when to stop (saving buffer to db)
				close(self.buffer)
				// use the db to persist all the backlog of requests currently in the buffer
				usedb = true
				// count the latest db batch
				b = 1
				continue
			}
			self.total++
			// glog.V(logger.Ridiculousness).Infof("[BZZ] syncDb[%v] - deliver from buffer: %v/%v", self.priority, n, t)
			// if draining n items from the queue
			if n > 0 {
				n--
				if n == 0 {
					glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] db mode", self.priority)
					usedb = true
					b = 0
				}
			}
			// if buffer is full and we are delivering from it,
			// start draining and after n items will start
			// saving the requests to db
			l := len(queue)
			if n == 0 && l >= cap(queue)-1 {
				n = len(queue)
				glog.V(logger.Debug).Infof("[BZZ] syncDb[%v] buffer full: switching to db. session tally (db/total): %v/%v", self.priority, self.dbTotal, self.total)
			}

			// signal that db items have been consumed ready to switch back to buffer
		case <-done:
			glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] buffer mode", self.priority)
			usedb = false
			if queue == nil {
				// deblock buffer, allow loop to read it
				queue = self.buffer
			} else {
				// clean up once more items that got there between iterator call and signal
				queue = nil // block reading from buffer
				go self.iterate(deliver)
			}
			// order counter reset to 0
			self.counter = 0

		case <-quit:
			close(self.buffer)
			// so that this is only called once
			quit = nil
			done = nil
			if !usedb {
				usedb = true
				b = 0
			}
			if queue == nil {
				queue = self.buffer
			}
		}
	}
	errc := make(chan error)
	// write last batch to db
	self.db.Put(self.counterKey, storage.U64ToBytes(self.counter))
	self.batch <- errc
	close(self.save)
	<-errc
	glog.V(logger.Info).Infof("[BZZ] syncDb[%v:%v]: saved %v keys (saved counter at %v)", self.key.Log(), self.priority, b, self.counter)
	close(self.done)
}

// batchServer loop handles db writes, batches are written only when
// a new iterator cycle is started
func (self *syncDb) batchServer() {
	batch := new(leveldb.Batch)
	var n int
	var err error
	var entry *syncDbEntry
	var more bool

	for {
		select {
		// incoming entry to put into db
		case entry, more = <-self.save:
			if !more {
				return
			}
			batch.Put(entry.key, entry.val)
			n++
			// need to save the batch if it gets too large
			if n%dbBatchSize == 0 {
				err = self.db.Write(batch)
				if err != nil {
					continue
				}
				batch = new(leveldb.Batch)
			}

			// explicit request for batch
		case c := <-self.batch:
			if n == 0 {
				close(c)
				continue
			}
			c <- self.db.Write(batch)
			batch = new(leveldb.Batch)
			n = 0
		}
	}
}

type syncDbEntry struct {
	key, val []byte
}

func (self syncDbEntry) String() string {
	return fmt.Sprintf("key: %x, value: %x", self.key, self.val)
}

/*
	iterate is iterating over store requests to be sent over to the peer
	this is mainly to prevent crashes due to network output buffer contention (???)
	as well as to make syncronisation resilient to disconnects
	the messages are supposed to be sent in the p2p priority queue.

	the request DB is shared between peers, but domains for each syncdb
	are disjoint. dbkeys are structured:
	0: 0x00 (0x01 reserved for counter)
	1: priorities - priority (so that high priority can be replayed first)
	2-34: peers address
*/
func (self *syncDb) iterate(fun func(interface{}, chan bool) bool) {
	key := make([]byte, 42)
	copy(key, self.start)
	var more bool
	var it iterator.Iterator
	var n, total int
	var entryKey []byte
	var err error
	errc := make(chan error)

	for round := 0; round == 0 || n > 0; round++ {
		// request a new batch be written
		self.batch <- errc
		// wait for the write to finish
		err, more = <-errc
		if round > 0 && !more {
			// the errc is closed if we got no more
			break
		}
		if !more {
			errc = make(chan error)
		}
		_ = err
		it = self.db.NewIterator()
		del := new(leveldb.Batch)
		it.Seek(key)
		glog.V(logger.Detail).Infof("[BZZ] syncDb[%v]: seek iterator: %x (round %v)", self.priority, key, round)
		key = it.Key()
		n = 0
		for len(key) >= 34 &&
			key[0] == 0 &&
			key[1] == byte(priorities-self.priority) &&
			bytes.Equal(key[2:34], self.key) {

			entry := &syncDbEntry{key, it.Value()}
			entryKey = entry.key[:]
			more = fun(entry, self.quit)
			if !more {
				// delivery cut short, quit the iteration process
				return
			}

			n++
			total++
			self.total++
			self.dbTotal++

			// glog.V(logger.Ridiculousness).Infof("[BZZ] syncDb[%v] - %v delivered chunk from db. round: %v, n: %v, total: %v, session total from db: %v/%v", self.priority, entry, round, n, total, self.dbTotal, self.total)
			// this should be batch delete at the end
			del.Delete(key)

			it.Next()
			key = it.Key()
		} //iterate
		glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] - end of db range for %v, rounds: %v, total: %v, session total from db: %v/%v", self.priority, self.key.Log(), round, total, self.dbTotal, self.total)

		key = entryKey
		it.Release()
		self.db.Write(del)
	} //rounds
	self.done <- true
}

//
func (self *syncDb) stop() {
	close(self.quit)
	for _ = range self.done {
	}
}

// saves one req to the database
func (self *syncDb) put(req interface{}) {
	entry, err := self.newSyncDbEntry(req)
	if err != nil {
		glog.V(logger.Warn).Infof("[BZZ] syncDb[%v] db.put %v failed: %v", self.priority, req, err)
		return
	}
	// glog.V(logger.Ridiculousness).Infof("[BZZ] syncDb[%v] db.put %v", self.priority, req)
	self.save <- entry
}

// calculate a dbkey for the request, for the db to work
// * one byte right after peer key encodes the priority
// * order is guaranteed by adding / recoring counts
func (self *syncDb) newSyncDbEntry(req interface{}) (entry *syncDbEntry, err error) {
	var key storage.Key
	var chunk *storage.Chunk
	var id uint64
	var ok bool
	var sreq *storeRequestMsgData

	if key, ok = req.(storage.Key); ok {
		id = generateId()
	} else if chunk, ok = req.(*storage.Chunk); ok {
		key = chunk.Key
		id = generateId()
	} else if sreq, ok = req.(*storeRequestMsgData); ok {
		key = sreq.Key
		id = sreq.Id
	} else if entry, ok = req.(*syncDbEntry); !ok {
		return nil, fmt.Errorf("type not allowed: %v (%T)", req, req)
	}

	// order by peer > priority > seqid
	// value is request id if exists
	if entry == nil {
		dbkey := make([]byte, 42)
		dbval := make([]byte, 40)

		// encode key
		dbkey[0] = 0
		dbkey[1] = byte(priorities - self.priority)
		copy(dbkey[2:], self.key) //db  peer
		binary.BigEndian.PutUint64(dbkey[34:], self.counter)
		self.counter++
		// encode value
		copy(dbval, key[:])
		binary.BigEndian.PutUint64(dbval[32:], id)

		entry = &syncDbEntry{dbkey, dbval}
	}
	return
}
