package bzz

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

/*
syncDb is a cache and queueing service for outgoing deliveries

once its cache is full it reverts to persisting in db
and serves the
*/
type syncDb struct {
	key            Key              // remote peers address key
	priority       uint             // priotity High|Medium|Low
	counter        uint64           // incrementing index to enforce order
	cache          chan interface{} // incoming request channel
	db             *LDBDatabase     // underlying db (should be interface)
	done           chan bool        // chan to signal  quit for goroutines
	quit           chan bool        // chan to signal  quit for goroutines
	total, dbTotal int              // counts for one session
}

var (
	counterKey = []byte{1}
)

// constructor needs a shared request db (leveldb)
// priority is used in the index key
// uses a cache and a leveldb for persistent storage
func newSyncDb(db *LDBDatabase, key Key, priority uint, bufferSize uint, deliver func(interface{})) *syncDb {
	data, err := db.Get(counterKey)
	var counter uint64
	if err == nil {
		counter = binary.LittleEndian.Uint64(data)
	}
	syncdb := &syncDb{
		key:      key,
		priority: priority,
		counter:  counter,
		cache:    make(chan interface{}, bufferSize),
		db:       db,
		quit:     make(chan bool),
		done:     make(chan bool),
	}
	glog.V(logger.Debug).Infof("[BZZ] syncDb[%v] - initialised", priority)

	go syncdb.handle(deliver)
	return syncdb
}

// handle is the loop that takes care of caching, and persistent
// storage
func (self *syncDb) handle(deliver func(interface{})) {
	var caching bool
	var n, t, d int
	var queue chan interface{}
	var err error
LOOP:
	for {
		var req interface{}
		// if cache is full and we are delivering from,
		// start draining and after n items will start
		// putting in db
		if caching && n == 0 && len(queue) == cap(queue) {
			n = cap(queue)
			glog.V(logger.Debug).Infof("[BZZ] syncDb[%v] cache full: switching to db. session tally (db/total): %v/%v", self.priority, self.dbTotal, self.total)
		}
		// waiting for item on the relevant queue

		select {
		case req = <-queue:
			t++
			if caching {
				deliver(req)
				// glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] - deliver from cache: %v/%v", self.priority, n, t)
				// if draining n items from the queue
				if n > 0 {
					n--
					if n == 0 {
						glog.V(logger.Debug).Infof("[BZZ] syncDb[%v] cache -> db mode", self.priority)
						caching = false
						go self.iterate(deliver)
					}
				}
			} else {
				err = self.put(req)
				d++
				if err != nil {
					glog.V(logger.Warn).Infof("[BZZ] syncDb[%v] db.put %v failed: %v", self.priority, req, err)
				} else {
					glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] db.put %v : %v/%v", self.priority, req, d, t)
				}
			}

			// receives a signal
		case <-self.done:
			// (otherwise new items are put to the db
			glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] db -> cache mode", self.priority)
			caching = true
			if queue == nil {
				queue = self.cache
			}
		case <-self.quit:
			glog.V(logger.Debug).Infof("[BZZ] syncDb[%v] quit: saving cache", self.priority)
			self.save()
			break LOOP
		}
	}
}

type syncDbEntry struct {
	key, val []byte
}

func (self syncDbEntry) String() string {
	return fmt.Sprintf("key: %x, value: %x", self.key, self.val)
}

// iterate is iterating over st ore requests to be sent over to the peer
// this is to prevent crashes due to network output buffer contention (???)
// the messages are supposed to be sent in the p2p priority queue.
// TODO: as soon as there is API for that feature, adjust.
// TODO: when peer drops the iterator position is not persisted
// the request DB is shared between peers, keys are prefixed by the peers address
// and the iterator (id BE32)
func (self *syncDb) iterate(fun func(interface{})) {
	start := make([]byte, 42)
	start[1] = byte(priorities - self.priority)
	copy(start[2:], self.key)
	key := make([]byte, 42)

	copy(key, start)
	var n, t, r int
	var it iterator.Iterator
	var entry *syncDbEntry
LOOP:
	for {
		if n == 0 {
			r++
			it = self.db.NewIterator()
			glog.V(logger.Debug).Infof("[BZZ] syncDb[%v]: seek iterator: %x", self.priority, key)
			it.Seek(key)
			if !it.Valid() {
				break LOOP
			}
			key = it.Key()
			entry = &syncDbEntry{key, it.Value()}
			n = 10000
		}

		// reached the end of this peers range
		if key[0] != 0 || key[1] != byte(priorities-self.priority) || !bytes.Equal(key[2:34], self.key) {
			glog.V(logger.Debug).Infof("[BZZ] syncDb[%v] - end of db range for %v %v/%v", self.priority, self.key, t, r)
			n = 0
			break LOOP
		}

		// apply func
		fun(entry)
		glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] - %v delivered chunk from db: %v/%v", self.priority, entry, self.total, r)
		self.db.Delete(key)

		n--
		self.total++
		self.dbTotal++

		it.Next()
		key = it.Key()
		if len(key) == 0 {
			key = start
			if n == 0 {
				break LOOP
			}
			n = 0
		}

		select {
		case <-self.quit:
			break LOOP
		default:
		}
	}
	// signal to caller finish with db:
	self.done <- true
	it.Release()
	glog.V(logger.Detail).Infof("[BZZ] syncDb[%v] - deliver %v from db: %v/%v", self.priority, self.key, t, r)
	// order counter reset to 0
	self.counter = 0
}

//
func (self *syncDb) stop() {
	self.save()
	close(self.quit)
}

// saves the contents of the queue to db
func (self *syncDb) save() error {
	var i, e int
	var err, glerr error
	size := len(self.cache)
	close(self.cache)

	for req := range self.cache {
		err = self.put(req)
		if err != nil {
			e++
			glog.V(logger.Warn).Infof("[BZZ] syncDb[%v:%v] save failed: %v", self.key, self.priority, err)
			glerr = err
		}
		if i == size {
			break
		}
		i++
	}
	glog.V(logger.Info).Infof("[BZZ] syncDb[%v:%v]: saved %v/%v keys (counter at %v)", self.key, self.priority, i-e, i, self.counter)
	// save db counter
	self.db.Put(counterKey, u64ToBytes(self.counter))
	return glerr
}

// saves one req to the database
func (self *syncDb) put(req interface{}) error {
	entry, err := self.newSyncDbEntry(req)
	if err != nil {
		return fmt.Errorf("syncDb.put: %v", err)
	}
	self.db.Put(entry.key, entry.val)
	return nil
}

// calculate a dbkey for the request, for the db to work
// * one byte right after peer key encodes the priority
// * order is guaranteed by adding / recoring counts
func (self *syncDb) newSyncDbEntry(req interface{}) (entry *syncDbEntry, err error) {
	var key Key
	var chunk *Chunk
	var id uint64
	var ok bool
	var sreq *storeRequestMsgData

	if key, ok = req.(Key); ok {
		id = generateId()
	} else if chunk, ok = req.(*Chunk); ok {
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
		self.counter++
		binary.BigEndian.PutUint64(dbkey[34:], self.counter)
		// encode value
		copy(dbval, key[:])
		binary.BigEndian.PutUint64(dbval[32:], id)

		entry = &syncDbEntry{dbkey, dbval}
	}
	return
}
