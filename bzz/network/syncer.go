package network

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/bzz/storage"
	"github.com/ethereum/go-ethereum/common/kademlia"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

// syncer parameters (global, not peer specific) default values
const (
	keyBufferSize  = 1024 // size of buffer  for unsynced keys
	syncBatchSize  = 128  // maximum batchsize for outgoing requests
	syncBufferSize = 128  // size of buffer  for delivery requests
	syncCacheSize  = 1024 // cache capacity to store request queue in memory
)

// priorities
const (
	Low        = iota // 0
	Medium            // 1
	High              // 2
	priorities        //= 3
)

// request types
const (
	DeliverReq   = iota // 0
	PushReq             // 1
	PropagateReq        // 2
	SyncReq             // 3
	StaleSyncReq        // 4
)

type syncState struct {
	storage.DbSyncState //
	// Start      Key    // lower limit of address space
	// Stop       Key    // upper limit of address space
	// First      uint64 // counter taken from last sync state
	// Last       uint64 // counter of remote peer dbStore at the time of last connection
	SessionAt  uint64      // set at the time of connection
	LastSeenAt uint64      // set at the time of connection
	Latest     storage.Key // cursor of dbstore when last (continuously set by syncer)
	Synced     bool        // true iff Sync is done up to the last disconnect
	synced     chan bool   // signal that sync stage finished
}

type DbAccess struct {
	db *storage.DbStore
}

func NewDbAccess(db *storage.DbStore) *DbAccess {
	return &DbAccess{db}
}

func (self *DbAccess) get(key storage.Key) (*storage.Chunk, error) {
	return self.db.Get(key)
}

func (self *DbAccess) counter() uint64 {
	return self.db.Counter()
}

func (self *DbAccess) iterator(s syncState) keyIterator {
	it, err := self.db.NewSyncIterator(s.DbSyncState)
	if err != nil {
		return nil
	}
	return keyIterator(it)
}

func (self syncState) String() string {
	return fmt.Sprintf(
		"addr: %v-%v, accessCount: %v-%v, session started at: %v, last seen at: %v, latest key: %v, synced: %v",
		self.Start, self.Stop,
		self.First, self.Last,
		self.SessionAt, self.LastSeenAt,
		self.Latest, self.Synced,
	)
}

// syncer parameters (global, not peer specific)
type SyncParams struct {
	RequestDbPath  string // path for request db (leveldb)
	KeyBufferSize  uint   // size of key buffer
	SyncBatchSize  uint   // maximum batchsize for outgoing requests
	SyncBufferSize uint   // size of buffer for
	SyncCacheSize  uint   // cache capacity to store request queue in memory
	SyncPriorities []uint // list of priority levels for req types 0-3
	SyncModes      []bool // list of sync modes for  for req types 0-3
}

// constructor with default values
func NewSyncParams(bzzdir string) *SyncParams {
	return &SyncParams{
		RequestDbPath:  filepath.Join(bzzdir, "requests"),
		KeyBufferSize:  keyBufferSize,
		SyncBufferSize: syncBufferSize,
		SyncBatchSize:  syncBatchSize,
		SyncCacheSize:  syncCacheSize,
		SyncPriorities: []uint{High, Medium, Medium, Low, Low},
		SyncModes:      []bool{true, true, true, true, false},
	}
}

// implemented by dbStoreSyncIterator
type keyIterator interface {
	Next() storage.Key
}

// syncer is the agent that manages content distribution/storage replication/chunk storeRequest forwarding
type syncer struct {
	*SyncParams                               // sync parameters
	key                       storage.Key     // remote peers address key
	state                     syncState       // sync state for our dbStore
	syncStates                chan syncState  // different stages of sync
	unsyncedKeysRequest       chan bool       // trigger to send unsynced keys
	keyCounts, deliveryCounts [priorities]int // counts
	keyCount, deliveryCount   int             //
	quit                      chan bool       // signal to quit loops

	// DB related fields
	dbAccess *DbAccess            // access to dbStore
	db       *storage.LDBDatabase // delivery msg db

	// native fields
	queues        [priorities]*syncDb                   // in-memory cache / queues for sync reqs
	keys          [priorities]chan interface{}          // buffer for unsynced keys
	deliveries    [priorities]chan *storeRequestMsgData // delivery
	deliveryInput chan uint                             // wake up trigger delivery request
	history       chan interface{}                      // db iterator channel

	// bzz protocol instance outgoing message callbacks
	unsyncedKeys func([]*syncRequest, syncState) error // send unsyncedKeysMsg
	store        func(*storeRequestMsgData) error      // send storeRequestMsg
}

// a syncer instance is linked to each peer connection
// constructor is called from protocol,
func newSyncer(
	db *storage.LDBDatabase, remotekey storage.Key,
	dbAccess *DbAccess,
	unsyncedKeys func([]*syncRequest, syncState) error,
	deliveryRequest func([]*syncRequest) error,
	store func(*storeRequestMsgData) error,
	params *SyncParams,
	state syncState,
) (*syncer, error) {

	syncBufferSize := params.SyncBufferSize
	keyBufferSize := params.KeyBufferSize

	self := &syncer{
		key:                 remotekey,
		dbAccess:            dbAccess,
		syncStates:          make(chan syncState, 20),
		history:             make(chan interface{}),
		unsyncedKeysRequest: make(chan bool, 1),
		SyncParams:          params,
		state:               state,
		quit:                make(chan bool),
		unsyncedKeys:        unsyncedKeys,
		store:               store,
	}

	// initialising
	for i := 0; i < priorities; i++ {
		self.keys[i] = make(chan interface{}, keyBufferSize)
		self.deliveries[i] = make(chan *storeRequestMsgData)
		self.queues[i] = newSyncDb(db, remotekey, uint(i), syncBufferSize, self.deliver(uint(i)))
	}
	self.state = state
	glog.V(logger.Info).Infof("[BZZ] syncer started: %v", state)
	// launch chunk delivery service
	go self.syncDeliveries()
	// history feed
	go self.syncHistory()
	// process unsynced keys to broadcast
	go self.syncUnsyncedKeys()
	// launch sync task manager
	go self.sync()

	return self, nil
}

// newSyncState returns a default sync state given local and remote
// addresses and db count
func newSyncState(start, stop kademlia.Address, count uint64) *syncState {
	// inclusive keyrange boundaries for db iterator

	return &syncState{
		DbSyncState: storage.DbSyncState{
			Start: storage.Key(start[:]),
			Stop:  storage.Key(stop[:]),
			First: 0,
			Last:  count - 1,
		},
		SessionAt: count,
		synced:    make(chan bool),
	}
}

func encodeSync(state *syncState) (*json.RawMessage, error) {
	data, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return nil, err
	}
	meta := json.RawMessage(data)
	return &meta, nil
}

func decodeSync(meta *json.RawMessage) (*syncState, error) {
	if meta == nil {
		return nil, fmt.Errorf("unable to deserialise sync state from <nil>")
	}
	state := &syncState{}
	err := json.Unmarshal([]byte(*(meta)), state)
	return state, err
}

/*
 sync implements the syncing script
 * first all items left in the request Db are replayed
   * type = StaleSync
   * Mode: by default once again via confirmation roundtrip
   * Priority: the items are replayed as the proirity specified for StaleSync
   * but within the order respects earlier priority level of request
 * after all items are consumed for a priority level, the the respective
  queue for delivery requests is open (this way new reqs not written to db)
 * the sync state provided by the remote peer is used to sync history
   * all the backlog from earlier (aborted) syncing is completed starting from latest
   * then all backlog upto last disconnect
*/
func (self *syncer) sync() {

	// trigger an unsyncedKeys msg
	self.triggerUnsyncedKeys()

	// 0. first replay stale requests from request db
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: start replaying stale requests from request db", self.key.Log())
	for p := priorities - 1; p >= 0; p-- {
		self.queues[p].iterate(self.replay())
	}
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: done replaying stale requests from request db", self.key.Log())
	// start syncdb on each priority level
	// only called once

	// unless peer is synced sync unfinished history beginning on
	if !self.state.Synced {
		if self.state.Latest != nil && !storage.IsZeroKey(self.state.Latest) {
			// 1. there is unfinished earlier sync
			self.state.Start = self.state.Latest
			self.syncStates <- self.state
			self.wait()
			if self.state.Last < self.state.SessionAt {
				self.state.First = self.state.Last + 1
			}
		}
		// 2. sync up to last disconnect
		if self.state.First < self.state.LastSeenAt {
			self.state.Last = self.state.LastSeenAt - 1
			self.syncStates <- self.state
			self.wait()
			self.state.First = self.state.LastSeenAt
		}
	} else {
		self.state.First = self.state.LastSeenAt
	}
	if self.state.First < self.state.SessionAt {
		self.state.Last = self.state.SessionAt - 1

		// 3. sync up to current session start
		self.syncStates <- self.state
		self.wait()
	}
	// sync finished
	close(self.syncStates)
	self.state.Synced = true
	glog.V(logger.Info).Infof("[BZZ] syncer[%v]: syncing complete", self.key.Log())

}

// wait till syncronised
func (self *syncer) wait() {
	select {
	case <-self.state.synced:
	case <-self.quit:
	}
}

// stop quits both request processor and saves the request cache to disk
func (self *syncer) stop() {
	close(self.quit)
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: save sync request db backlog", self.key.Log())
	for _, db := range self.queues {
		db.stop()
	}
}

// rlp serialisable sync request
type syncRequest struct {
	Key      storage.Key
	Priority uint
}

func (self *syncRequest) String() string {
	return fmt.Sprintf("<Key: %v, Priority: %v>", self.Key.Log(), self.Priority)
}

func (self *syncer) newSyncRequest(req interface{}, p int) (*syncRequest, error) {
	key, _, _, _, err := parseRequest(req)
	// TODO: if req has chunk, it should be put in a cache
	// create
	if err != nil {
		return nil, err
	}
	return &syncRequest{key, uint(p)}, nil
}

// serves historical items from the DB
// * read is on demand, blocking unless history channel is read
// * accepts sync requests (syncStates) to create new db iterator
// * signals back if one iteration finishes
// * quits if all sync request srved and syncStates channel is close
// * complete sync is signaled by the closed history channel
func (self *syncer) syncHistory() {
	var t, r uint
LOOP:
	for state := range self.syncStates {
		r++
		var n uint
		glog.V(logger.Debug).Infof("[BZZ] syncer[%v/%v]: syncing history between %v - %v for chunk addresses %v - %v", self.key.Log(), r, self.state.First, self.state.Last, self.state.Start, self.state.Stop)
		it := self.dbAccess.iterator(state)
		if it != nil {
		IT:
			for {
				key := it.Next()
				if key != nil {
					select {
					// blocking until history channel is read from
					case self.history <- storage.Key(key):
						n++
						t++
						glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: history: %v. %v/%v unsynced keys", self.key.Log(), key.Log(), n, t)
						state.Latest = key
					case <-self.quit:
						break LOOP
					}
				} else {
					break IT
				}
			}
			glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: history sync iteration complete: %v/%v/%v", self.key.Log(), n, t, r)
		}
		// signal end of the iteration ended
		state.synced <- true
		glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: finished syncing history between %v - %v for chunk addresses %v - %v (at %v) (%v/%v/%v)", self.key.Log(), self.state.First, self.state.Last, self.state.Start, self.state.Stop, self.state.Latest, n, t, r)

	}
	// all sync states consumed history channel closed
	close(self.history)
}

func (self *syncer) triggerUnsyncedKeys() {
	select {
	case self.unsyncedKeysRequest <- true:
	default:
	}
}

// assembles a new batch of unsynced keys
// * keys are drawn from the key buffers in order of priority queue
// * if the queues of priority for History (SyncReq) or higher are depleted,
//   historical data is used so historical items are lower priority within
//   their priority group.
// * Order of historical data is unspecified
func (self *syncer) syncUnsyncedKeys() {
	// send out new
	var unsynced []*syncRequest
	var more bool
	var total, history uint
	histPrior := self.SyncPriorities[SyncReq]
	p := High
	keys := self.keys[p] // starts out sending right away?
	var synced, wasSynced bool
	synced = self.state.Synced
	wasSynced = synced
	var timer <-chan time.Time

LOOP:
	// loop structure needs to be like this so that while draining
	// lower priorities we do not miss higher ones coming in.
	for {
		var req interface{}

		select {
		// idle if keys is nil
		case req = <-keys:
			glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: picking top delivery request %v (%v)", self.key.Log(), req, p)
		default:
			// all keys of priority queue p consumed
			// if the priority queue is the one for historical chunks
			// and the level has been drained we fill the batch with history
			if !synced && uint(p) == histPrior {
				// pull new history item
				req, more = <-self.history
				if !more {
					synced = true
				} else {
					self.keyCounts[p]++
					self.keyCount++
					history++
				}
				glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: (priority %v): %v (synced = %v) historical", self.key.Log(), p, req, synced)
			}
		}

		// still need reqs to fill the batch
		if sreq, err := self.newSyncRequest(req, p); err == nil { // if histPrior = 1 and no history, then req can be nil
			// extract key from req
			glog.V(logger.Detail).Infof("[BZZ] syncer[%v] (priority %v): parsing request %v (synced = %v)", self.key.Log(), p, req, synced)
			total++
			unsynced = append(unsynced, sreq)
		}

		// send msg iff
		// * all queues are depleted and no more syncing OR
		// * batch full OR
		if p == Low && synced || len(unsynced) == int(self.SyncBatchSize) {
			// if there are new keys unsynced
			// or if the state just changed to synced (history is synced)
			if len(unsynced) > 0 || !wasSynced && self.state.Synced {
				// set sync to current counter
				// (all nonhistorical outgoing traffic sheduled and persisted
				self.state.LastSeenAt = self.dbAccess.counter()
				glog.V(logger.Debug).Infof("[BZZ] syncer[%v] sending out msg with %v unsynced keys, state: %v", self.key.Log(), len(unsynced), self.state)
				glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: sending %v", self.key.Log(), unsynced)
				glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: session tally: keys: %v (%v), history: %v, total: %v", self.key.Log(), self.keyCounts, self.keyCount, history, total)
				glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: session tally: deliveries: %v (%v)", self.key.Log(), self.deliveryCounts, self.deliveryCount)
				//  send the unsynced keys
				err := self.unsyncedKeys(unsynced, self.state)
				if err != nil {
					glog.V(logger.Warn).Infof("[BZZ] syncer[%v]: unable to send unsynced keys: %v", err)
				}
				unsynced = nil
			} else {
				timer = time.NewTimer(500 * time.Millisecond).C
				select {
				case <-self.quit:
					break LOOP
				case <-timer:
					timer = nil
					go self.triggerUnsyncedKeys()
				case <-self.unsyncedKeysRequest:
					timer = nil
					// triggers listening to new keys
					if keys == nil { // very first initialisation
						p = High
						keys = self.keys[High]
					}
				}
			}
			wasSynced = synced
		} else {
			// hop down one priority level, or start again
			if p == Low {
				p = High
			} else {
				p--
			}
			keys = self.keys[p]
		}
	}
}

// delivery loop
// takes into account priority, send store Requests with chunk (delivery)
// idle blocking if no new deliveries in any of the queues
func (self *syncer) syncDeliveries() {
	var req *storeRequestMsgData
	p := High
	var deliveries chan *storeRequestMsgData
	var msg *storeRequestMsgData
	var err error
	var c = [priorities]int{}
	var n = [priorities]int{}
	var total, success uint
	for {
		deliveries = self.deliveries[p]
		select {
		case req = <-deliveries:
			n[p]++
			c[p]++
		default:
			if p == Low {
				// blocking, depletion on all channels, no preference for priority
				select {
				case req = <-self.deliveries[High]:
					n[High]++
				case req = <-self.deliveries[Medium]:
					n[Medium]++
				case req = <-self.deliveries[Low]:
					n[Low]++
				case <-self.quit:
					return
				}
				p = High
			} else {
				p--
				continue
			}
		}
		total++
		msg, err = self.newStoreRequestMsgData(req)
		if err != nil {
			glog.V(logger.Warn).Infof("[BZZ] syncer[%v]: failed to deliver %v: %v", self.key.Log(), req, err)
		} else {
			err = self.store(msg)
			if err != nil {
				glog.V(logger.Warn).Infof("[BZZ] syncer[%v]: failed to deliver %v: %v", self.key.Log(), req, err)
			} else {
				success++
				glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: %v successfully delivered", self.key.Log(), req)
			}
		}
		if total%self.SyncBatchSize == 0 {
			glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: deliver Total: %v, Success: %v, High: %v/%v, Medium: %v/%v, Low %v/%v", self.key.Log(), total, success, c[High], n[High], c[Medium], n[Medium], c[Low], n[Low])
		}
	}
}

/*
 addRequest handles requests for delivery
 it accepts 4 types:

 * storeRequestMsgData: coming from netstore propagate response
 * chunk: coming from forwarding (questionable: id?)
 * key: from incoming syncRequest
 * syncDbEntry: key,id encoded in db

 Could take simply storeRequestMsgData always and here fill in missing
 chunk, id,

 If sync mode is on for the type of request, then
 it sends the request to the keys queue of the correct priority
 channel buffered with capacity (SyncBufferSize)

 If sync mode is off then, requests are directly sent to requestQueue
*/
func (self *syncer) addRequest(req interface{}, ty int) {
	// retrieve priority for request type
	priority := self.SyncPriorities[ty]
	// sync mode for this type ON
	if self.SyncModes[ty] {
		self.addKey(req, priority, self.quit)
	} else {
		self.addDelivery(req, priority, self.quit)
	}
}

// addKey queues sync request for sync confirmation with given priority
// ie the key will go out in an unsyncedKeys message
func (self *syncer) addKey(req interface{}, priority uint, quit chan bool) bool {
	select {
	case self.keys[priority] <- req:
		self.keyCounts[priority]++
		self.keyCount++
		return true
	case <-quit:
		return false
	}
}

// addDelivery queues delivery request for with given priority
// ie the chunk will be delivered ASAP mod priorities
// requests are persisted across sessions for correct sync
func (self *syncer) addDelivery(req interface{}, priority uint, quit chan bool) bool {
	select {
	case self.queues[priority].buffer <- req:
		return true
	case <-quit:
		return false
	}
}

// doDelivery delivers the chunk for the request with given priority ac
func (self *syncer) doDelivery(req interface{}, priority uint, quit chan bool) bool {
	msgdata, err := self.newStoreRequestMsgData(req)
	if err != nil {
		glog.V(logger.Warn).Infof("unable to deliver request %v: %v", msgdata, err)
		return false
	}
	select {
	case self.deliveries[priority] <- msgdata:
		self.deliveryCounts[priority]++
		self.deliveryCount++
		return true
	case <-quit:
		return false
	}
}

// returns the delivery function for given priority
// passed on to syncDb
func (self *syncer) deliver(priority uint) func(req interface{}, quit chan bool) bool {
	return func(req interface{}, quit chan bool) bool {
		return self.doDelivery(req, priority, quit)
	}
}

// returns the replay function passed on to syncDb
// depending on sync mode settings for StaleSyncReq,
// re	play of request db backlog sends items via confirmation
// or directly delivers
func (self *syncer) replay() func(req interface{}, quit chan bool) bool {
	sync := self.SyncModes[StaleSyncReq]
	priority := self.SyncPriorities[StaleSyncReq]
	// sync mode for this type ON
	if sync {
		return func(req interface{}, quit chan bool) bool {
			return self.addKey(req, priority, quit)
		}
	} else {
		return func(req interface{}, quit chan bool) bool {
			return self.doDelivery(req, priority, quit)
		}

	}
}

// given a request, extends it to a full storeRequestMsgData
// see types accepted
func (self *syncer) newStoreRequestMsgData(req interface{}) (*storeRequestMsgData, error) {

	key, id, chunk, sreq, err := parseRequest(req)
	if err != nil {
		return nil, err
	}

	if sreq == nil {
		if chunk == nil {
			var err error
			chunk, err = self.dbAccess.get(key)
			if err != nil {
				return nil, err
			}
		}

		sreq = &storeRequestMsgData{
			Id:    id,
			Key:   chunk.Key,
			SData: chunk.SData,
		}
	}

	return sreq, nil
}

// parse request types and extracts, key, id, chunk, request if available
// does not do chunk lookup !
func parseRequest(req interface{}) (storage.Key, uint64, *storage.Chunk, *storeRequestMsgData, error) {
	var key storage.Key
	var entry *syncDbEntry
	var chunk *storage.Chunk
	var id uint64
	var ok bool
	var sreq *storeRequestMsgData
	var err error

	if key, ok = req.(storage.Key); ok {
		id = generateId()

	} else if entry, ok = req.(*syncDbEntry); ok {
		id = binary.BigEndian.Uint64(entry.val[32:])
		key = storage.Key(entry.val[:32])

	} else if chunk, ok = req.(*storage.Chunk); ok {
		key = chunk.Key
		id = generateId()

	} else if sreq, ok = req.(*storeRequestMsgData); !ok {
		err = fmt.Errorf("type not allowed: %v (%T)", req, req)
	}

	return key, id, chunk, sreq, err
}
