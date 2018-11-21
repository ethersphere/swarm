package rushed

import (
	"context"
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/shed"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// Mode is an enum for modes of access/update
type Mode = int

var (
	errDBClosed = errors.New("DB closed")
)

// Batch wraps leveldb.Batch extending it with a waitgroup and a done channel
type Batch struct {
	*leveldb.Batch
	Done chan struct{} // to signal when batch is written
	Err  error         // error resulting from write
}

// NewBatch constructs a new batch
func NewBatch() *Batch {
	return &Batch{
		Batch: new(leveldb.Batch),
		Done:  make(chan struct{}),
	}
}

// DB extends shed DB with batch execution, garbage collection and iteration support with subscriptions
type DB struct {
	*shed.DB                                           // underlying shed.DB
	update   func(*Batch, Mode, *shed.IndexItem) error // mode-dependent update method
	access   func(Mode, *shed.IndexItem) error         // mode-dependent access method
	batch    *Batch                                    // current batch
	mu       sync.RWMutex                              // mutex for accessing current batch
	c        chan struct{}                             // channel to signal writes on
	closed   chan struct{}                             // closed when writeBatches loop quits
}

// New constructs a new DB
func New(sdb *shed.DB, update func(*Batch, Mode, *shed.IndexItem) error, access func(Mode, *shed.IndexItem) error) *DB {
	db := &DB{
		DB:     sdb,
		update: update,
		access: access,
		batch:  NewBatch(),
		c:      make(chan struct{}, 1),
		closed: make(chan struct{}),
	}
	go db.writeBatches()
	return db
}

// Close terminates loops by closing the quit channel
func (db *DB) Close() {
	// signal quit to writeBatches loop
	close(db.c)
	// wait for last batch to be written
	<-db.closed
	db.DB.Close()
}

// Accessor is a wrapper around DB where Put/Get is overwritten to apply the
// update/access method for the mode
// using Mode(mode) the DB implements the ChunkStore interface
type Accessor struct {
	mode Mode
	*DB
}

// Mode returns the ChunkStore interface for the mode of update on a multimode update DB
func (db *DB) Mode(mode Mode) *Accessor {
	return &Accessor{
		mode: mode,
		DB:   db,
	}
}

// Put overwrites the underlying DB Put method for the specific mode of update
func (u *Accessor) Put(ctx context.Context, ch storage.Chunk) error {
	return u.Update(ctx, u.mode, newItemFromChunk(ch))
}

// Get overwrites the underlying DB Get method for the specific mode of access
func (u *Accessor) Get(_ context.Context, addr storage.Address) (storage.Chunk, error) {
	item := newItemFromAddress(addr)
	if err := u.access(u.mode, item); err != nil {
		return nil, err
	}
	return storage.NewChunk(item.Address, item.Data), nil
}

// Update calls the update method for the specific mode with items
func (db *DB) Update(ctx context.Context, mode Mode, item *shed.IndexItem) error {
	// obtain the current batch
	// b := <-db.batch
	db.mu.RLock()
	b := db.batch
	db.mu.RUnlock()
	log.Debug("obtained batch")
	if b == nil {
		return errDBClosed
	}
	// call the update with the  access mode
	err := db.update(b, mode, item)
	if err != nil {
		return err
	}
	// wait for batch to be written and return batch error
	// this is in order for Put calls to be synchronous
	select {
	case db.c <- struct{}{}:
	default:
	}
	select {
	case <-b.Done:
	case <-ctx.Done():
		return ctx.Err()
	}
	return b.Err
}

// writeBatches is a forever loop handing out the current batch to updaters
// and apply the batch when the db is free
// if the db is quit, the last batch is written out and batch channel is closed
func (db *DB) writeBatches() {
	defer close(db.closed)
	for range db.c {
		db.mu.Lock()
		b := db.batch
		db.batch = NewBatch()
		db.mu.Unlock()
		db.writeBatch(b)
	}
}

// writeBatch writes out the batch, sets the error and closes the done channel
func (db *DB) writeBatch(b *Batch) {
	// apply the batch
	b.Err = db.DB.WriteBatch(b.Batch)
	// signal batch write to callers
	close(b.Done)
}

func newItemFromChunk(ch storage.Chunk) *shed.IndexItem {
	return &shed.IndexItem{
		Address: ch.Address(),
		Data:    ch.Data(),
	}
}

func newItemFromAddress(addr storage.Address) *shed.IndexItem {
	return &shed.IndexItem{
		Address: addr,
	}
}
