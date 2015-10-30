package bzz

import (
	"encoding/binary"
	"fmt"
	"time"

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
	DbSyncState //
	// Start      Key    // lower limit of address space
	// Stop       Key    // upper limit of address space
	// First      uint64 // counter taken from last sync state
	// Last       uint64 // counter of remote peer dbStore at the time of last connection
	SessionAt  uint64    // set at the time of connection
	LastSeenAt uint64    // set at the time of connection
	Latest     Key       // cursor of dbstore when last (continuously set by syncer)
	Synced     bool      // true iff Sync is done up to the last disconnect
	synced     chan bool // signal that sync stage finished
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
	KeyBufferSize  uint   // size of key buffer
	SyncBatchSize  uint   // maximum batchsize for outgoing requests
	SyncBufferSize uint   // size of buffer for
	SyncCacheSize  uint   // cache capacity to store request queue in memory
	SyncPriorities []uint // list of priority levels for req types 0-3
	SyncModes      []bool // list of sync modes for  for req types 0-3
}

// constructor with default values
func newSyncParams() *SyncParams {
	return &SyncParams{
		KeyBufferSize:  keyBufferSize,
		SyncBufferSize: syncBufferSize,
		SyncBatchSize:  syncBatchSize,
		SyncCacheSize:  syncCacheSize,
		SyncPriorities: []uint{High, High, Medium, Low, Low},
		SyncModes:      []bool{true, true, true, true, false},
	}
}

// implemented by dbStoreSyncIterator
type keyIterator interface {
	Next() Key
}

// syncer is the agent that manages content distribution/storage replication/chunk storeRequest forwarding
type syncer struct {
	key                 Key             // remote peers address key
	*SyncParams                         // sync parameters
	state               syncState       // sync state for our dbStore
	syncStates          chan syncState  // different stages of sync
	unsyncedKeysRequest chan bool       // trigger to send unsynced keys
	keyCounts           [priorities]int // counts
	deliveryCounts      [priorities]int // counts
	keyCount            int
	deliveryCount       int

	counter  func() uint64                 // db counter
	localGet func(Key) (*Chunk, error)     // get access toww local store
	kitf     func(s syncState) keyIterator // chunk db iterator function
	db       *LDBDatabase                  // delivery msg db
	quit     chan bool                     // signal to quit loops

	queues        [priorities]*syncDb                   // in-memory cache / queues for sync reqs
	keys          [priorities]chan interface{}          // buffer for unsynced keys
	deliveries    [priorities]chan *storeRequestMsgData // delivery
	deliveryInput chan uint                             // wake up trigger delivery request
	history       chan interface{}                      // db iterator channel

	unsyncedKeys    func([]*syncRequest, syncState) error // send unsyncedKeysMsg
	deliveryRequest func([]*syncRequest) error            //
	store           func(*storeRequestMsgData) error      // send storeRequestMsg
}

// a syncer instance is linked to each peer connection
// constructor is called from protocol,
func newSyncer(
	db *LDBDatabase, remotekey Key,
	counter func() uint64, kitf func(s syncState) keyIterator, get func(Key) (*Chunk, error),
	unsyncedKeys func([]*syncRequest, syncState) error,
	deliveryRequest func([]*syncRequest) error,
	store func(*storeRequestMsgData) error,
	params *SyncParams,
	state syncState,
) (*syncer, error) {

	// syncCacheSize := params.SyncCacheSize
	// syncBatchSize := params.SyncBatchSize
	syncBufferSize := params.SyncBufferSize
	keyBufferSize := params.KeyBufferSize

	self := &syncer{
		key:                 remotekey,
		counter:             counter,
		syncStates:          make(chan syncState, 20),
		history:             make(chan interface{}),
		unsyncedKeysRequest: make(chan bool, 1),
		kitf:                kitf,
		localGet:            get,
		SyncParams:          params,
		state:               state,
		quit:                make(chan bool),
		unsyncedKeys:        unsyncedKeys,
		deliveryRequest:     deliveryRequest,
		store:               store,
	}

	// initialising
	done := make(chan bool)
	for i := 0; i < priorities; i++ {
		self.keys[i] = make(chan interface{}, keyBufferSize)
		self.deliveries[i] = make(chan *storeRequestMsgData)
		self.queues[i] = newSyncDb(db, remotekey, uint(i), syncBufferSize, self.deliver(uint(i)), done)
	}
	self.state = state
	glog.V(logger.Info).Infof("[BZZ] syncer started: %v", state)
	// launch chunk delivery service
	go self.handleDeliveries()
	go self.handleHistory()
	go self.handleUnsyncedKeys()
	go self.sync(done)

	return self, nil
}

func newSyncState(localaddr, remoteaddr kademlia.Address, count uint64) *syncState {
	start := kademlia.CommonBitsAddrByte(localaddr, remoteaddr, byte(0))
	stop := kademlia.CommonBitsAddrByte(localaddr, remoteaddr, byte(255))

	return &syncState{
		DbSyncState: DbSyncState{
			Start: Key(start[:]),
			Stop:  Key(stop[:]),
			First: 0,
			Last:  count,
		},
		Latest:    Key(start[:]),
		SessionAt: count,
		synced:    make(chan bool),
	}
}

/*
 sync implements the syncing script
 * first all items left in the request Db are replayed
   * type = StaleSync
   * Mode: by default once again via confirmation roundtrip
   * Priority: the items are replayed as the proirity specified for StaleSync
   * but within the order respects earlier priority level of request
 * after all items are consumed for a priority level, the the respective
  queue for delivery requests is open (this way new reqs not writter to db)
 * the token is reset to current
*/
func (self *syncer) sync(done chan bool) {

	// trigger an unsyncedKeys msg
	self.handleDeliveryRequestMsg(nil)

	// first replay stale requests from request db
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: start replaying stale requests from request db", self.key)
	for p := priorities - 1; p >= 0; p-- {
		self.queues[p].iterate(self.replay(), nil)
	}
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: done replaying stale requests from request db", self.key)
	// start syncdb on each priority level
	close(done)
	// only called once

	// unless peer is synced sync unfinished history beginning on
	if !self.state.Synced && self.state.Last != self.state.SessionAt {

		glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: syncing history between %v - %v for chunk addresses %v - %v", self.key, self.state.First, self.state.Last, self.state.Start, self.state.Stop)
		self.syncStates <- self.state
		<-self.state.synced

		// history all 	the way to last disconnect
		glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: syncing history between %v - %v for chunk addresses %v - %v", self.key, self.state.First, self.state.Last, self.state.Start, self.state.Stop)
		self.state.First = self.state.Last
		self.state.Last = self.state.LastSeenAt
		self.syncStates <- self.state
		<-self.state.synced
	}

	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: syncing history between %v - %v for chunk addresses %v - %v", self.key, self.state.First, self.state.Last, self.state.Start, self.state.Stop)
	self.state.First = self.state.LastSeenAt
	self.state.Last = self.state.SessionAt
	// syncing since last disconnect to current session
	self.syncStates <- self.state
	<-self.state.synced

	// sync finished
	close(self.syncStates)
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: syncing complete", self.key)

}

// stop quits both request processor and saves the request cache to disk
func (self *syncer) stop() {
	close(self.quit)
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: save sync request db backlog", self.key)
	for _, db := range self.queues {
		db.stop()
	}
}

// rlp serialisable sync request
type syncRequest struct {
	Key      Key
	Priority uint
}

func (self *syncRequest) String() string {
	return fmt.Sprintf("<Key: %v, Priority: %v>", self.Key, self.Priority)
}

func (self *syncer) newSyncRequest(req interface{}, p int) *syncRequest {
	key, _, _, _, err := self.parseRequest(req)
	// TODO: if req has chunk, it should be put in a cache
	// create
	if err != nil {
		return nil
	}
	return &syncRequest{key, uint(p)}
}

// serves historical items from the DB
// * read is on demand, blocking unless history channel is read
// * accepts sync requests (syncStates) to create new db iterator
// * signals back if one iteration finishes
// * quits if all sync request srved and syncStates channel is close
// * complete sync is signaled by the closed history channel
func (self *syncer) handleHistory() {
	var t uint
LOOP:
	for state := range self.syncStates {
		var n uint
		it := self.kitf(state)
	IT:
		for {
			key := it.Next()

			if key != nil {
				select {
				case self.history <- Key(key):
					n++
					t++
					glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: history: %v. %v/%v unsynced keys", self.key, key, n, t)
				case <-self.quit:
					break LOOP
				}
			} else {
				glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: history sync iteration finished: %v/%v unsynced keys", self.key, n, t)
				state.synced <- true
				break IT
			}
		}

	}
	close(self.history)
}

// handles UnsyncedKeysMsg after msg decoding - unsynced hashes upto sync state
// * the remote sync state is just stored and handled in protocol
// * filters through the new syncRequests and send the ones missing
// * back immediately as a deliveryRequest message
// * empty message just pings back for more (is this needed?)
// * strict signed sync states may be needed.
func (self *syncer) handleUnsyncedKeysMsg(unsynced []*syncRequest) error {
	var missing []*syncRequest
	var err error
	for _, req := range unsynced {
		// skip keys that are found,
		_, err = self.localGet(Key(req.Key[:]))
		if err != nil {
			missing = append(missing, req)
		}
	}
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: received %v unsynced keys: %v missing", self.key, len(unsynced), len(missing))
	// send delivery request with missing keys
	err = self.deliveryRequest(missing)
	if err != nil {
		return err
	}
	return nil
}

// handles deliveryRequestMsg
// * serves actual chunks asked by the remote peer
// by pushing to the delivery queue (sync db) of the correct priority
// (remote peer is free to reprioritize)
// * the message implies remote peer wants more, so trigger for
// * new outgoing unsynced keys message is fired
func (self *syncer) handleDeliveryRequestMsg(deliver []*syncRequest) error {
	// queue the actual delivery of a chunk ()
	glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: received %v delivery requests", self.key, len(deliver))
	for _, sreq := range deliver {
		// TODO: look up in cache here or in deliveries
		// r = self.pullCached(sreq.Key) // pulls and deletes from cache
		self.addDelivery(sreq.Key, sreq.Priority)
	}

	// sends it out as unsyncedKeysMsg
	self.triggerUnsyncedKeys()
	return nil
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
func (self *syncer) handleUnsyncedKeys() {
	// send out new
	var unsynced []*syncRequest
	var more bool
	var total uint
	histPrior := self.SyncPriorities[SyncReq]
	p := High
	keys := self.keys[p] // starts out sending right away?
	var synced, wasSynced bool
	synced = self.state.Synced
	var timer <-chan time.Time

LOOP:
	// loop structure needs to be like this so that while draining
	// lower priorities we do not miss higher ones coming in.
	for {
		var req interface{}

		select {
		// idle if keys is nil
		case req = <-keys:
			glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: picking top delivery request %v (%v)", self.key, req, p)
		default:
			// all keys of priority queue p consumed
			// if the priority queue is the one for historical chunks
			// and the level has been drained we fill the batch with history
			if !synced && uint(p) == histPrior {
				// pull new history item
				req, more = <-self.history
				if !more {
					synced = true
					self.state.Synced = true
				}
				glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: (priority %v): %v (synced = %v) historical", self.key, p, req, synced)
			}
		}
		// still need reqs to fill the batch
		if sreq := self.newSyncRequest(req, p); sreq != nil { // if histPrior = 1 and no history, then req can be nil
			// extract key from req
			glog.V(logger.Detail).Infof("[BZZ] syncer[%v] (priority %v): parsing request %v (synced = %v)", self.key, p, req, synced)
			total++
			unsynced = append(unsynced, sreq)
		}
		// send msg iff
		// * all queues are depleted and no more syncing OR
		// * batch full OR
		if p == Low && synced || len(unsynced) == int(self.SyncBatchSize) {

			// set sync to current counter
			// (all nonhistorical outgoing traffic sheduled and persisted
			self.state.LastSeenAt = self.counter()
			// if there are new keys unsynced
			// or if the state just changed to synced (history is synced)
			if len(unsynced) > 0 || !wasSynced && self.state.Synced {
				glog.V(logger.Debug).Infof("[BZZ] syncer[%v] sending out msg with %v unsynced keys: %v, state: %v", self.key, len(unsynced), unsynced, self.state)
				glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: session tally: keys: <%v> (%v)", self.key, self.keyCounts, self.keyCount)
				glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: session tally: deliveries: <%v> (%v)", self.key, self.deliveryCounts, self.deliveryCount)
				self.unsyncedKeys(unsynced, self.state)
				unsynced = nil
			} else {
				timer = time.NewTimer(500 * time.Millisecond).C
				select {
				case <-timer:
					go self.triggerUnsyncedKeys()
				case <-self.quit:
					break LOOP
				case <-self.unsyncedKeysRequest:
					timer = nil
					// triggers listening to new keys
					if keys == nil {
						p = High
						keys = self.keys[High]
					}
				}
			}
			keys = nil
			wasSynced = synced
		} else {
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
func (self *syncer) handleDeliveries() {
	var req *storeRequestMsgData
	p := High
	var deliveries chan *storeRequestMsgData
	var msg *storeRequestMsgData
	var err error
	var c = [priorities]int{}
	var n = [priorities]int{}
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
				}
				p = High
			} else {
				p--
				continue
			}
		}
		msg, err = self.newStoreRequestMsgData(req)
		if err != nil {
			glog.V(logger.Warn).Infof("[BZZ] syncer[%v]: failed to deliver %v: %v", self.key, req, err)
		} else {
			glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: deliver high %v/%v - %v", self.key, c[High], n[High])
			glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: deliver medium %v/%v", self.key, c[Medium], n[Medium])
			glog.V(logger.Debug).Infof("[BZZ] syncer[%v]: deliver low %v/%v", self.key, c[Low], n[Low])
			err = self.store(msg)
			if err != nil {
				glog.V(logger.Warn).Infof("[BZZ] syncer[%v]: failed to deliver %v: %v", self.key, req, err)
			} else {
				glog.V(logger.Detail).Infof("[BZZ] syncer[%v]: %v successfully delivered", self.key)
			}
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
		self.addKey(req, priority)
	} else {
		self.addDelivery(req, priority)
	}
}

// addSyncRequest queues sync request for sync confirmation with given priority
// ie the key will go out in an unsyncedKeys message
func (self *syncer) addKey(req interface{}, priority uint) {
	self.keys[priority] <- req
	self.keyCounts[priority]++
	self.keyCount++
}

// addDeliveryRequest queues delivery request for with given priority
// ie the chunk will be delivered ASAP mod priorities
// requests are persisted across sessions for correct sync
func (self *syncer) addDelivery(req interface{}, priority uint) {
	self.queues[priority].cache <- req
}

// doDeliveryRequest delivers the chunk for the request with given priority
func (self *syncer) doDelivery(req interface{}, priority uint) {
	msgdata, err := self.newStoreRequestMsgData(req)
	if err != nil {
		glog.V(logger.Warn).Infof("unable to deliver request %v: %v", msgdata, err)
		return
	}
	self.deliveries[priority] <- msgdata
	self.deliveryCounts[priority]++
	self.deliveryCount++
}

// returns the delivery function for given priority
// passed on to syncDb
func (self *syncer) deliver(priority uint) func(req interface{}) {
	return func(req interface{}) {
		self.doDelivery(req, priority)
	}
}

// returns the replay function passed on to syncDb
// depending on sync mode settings for StaleSyncReq,
// re	play of request db backlog sends items via confirmation
// or directly delivers
func (self *syncer) replay() func(req interface{}) {
	sync := self.SyncModes[StaleSyncReq]
	priority := self.SyncPriorities[StaleSyncReq]
	return func(req interface{}) {
		// sync mode for this type ON
		if sync {
			self.addKey(req, priority)
		} else {
			self.doDelivery(req, priority)
		}
	}
}

// given a request, extends it to a full storeRequestMsgData
// see types accepted
func (self *syncer) newStoreRequestMsgData(req interface{}) (*storeRequestMsgData, error) {

	key, id, chunk, sreq, err := self.parseRequest(req)
	if err != nil {
		return nil, err
	}

	if sreq == nil {
		if chunk == nil {
			var err error
			chunk, err = self.localGet(key)
			if err != nil {
				return nil, err
			}
		}

		sreq = &storeRequestMsgData{
			Id:    id,
			Key:   key,
			SData: chunk.SData,
		}
	}

	return sreq, nil
}

// parse request types and extracts, key, id, chunk, request if available
// does not do chunk lookup !
func (self *syncer) parseRequest(req interface{}) (Key, uint64, *Chunk, *storeRequestMsgData, error) {
	var key Key
	var entry *syncDbEntry
	var chunk *Chunk
	var id uint64
	var ok bool
	var sreq *storeRequestMsgData
	var err error

	if key, ok = req.(Key); ok {
		id = generateId()
	} else if entry, ok = req.(*syncDbEntry); ok {
		id = binary.BigEndian.Uint64(entry.val[32:])
		key = Key(entry.val[:32])

	} else if chunk, ok = req.(*Chunk); ok {
		key = chunk.Key
		id = generateId()
	} else if sreq, ok = req.(*storeRequestMsgData); !ok {
		err = fmt.Errorf("type not allowed: %v (%T)", req, req)
	}
	return key, id, chunk, sreq, err
}
