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

type Mode = int

var (
	errDBClosed  = errors.New("DB closed")
	errCancelled = errors.New("iteration cancelled")
)

// batch wraps leveldb batch extending it with a waitgroup and a done channel
type Batch struct {
	*leveldb.Batch
	wg   sync.WaitGroup // to signal and wait for parallel writes to batch
	Done chan struct{}  // to signal when batch is written
	Err  error          // error resulting from write
}

// newBatch constructs a new batch
func newBatch() *Batch {
	return &Batch{
		Batch: new(leveldb.Batch),
		Done:  make(chan struct{}),
	}
}

// DB extends shed DB with batch execution, garbage collection and iteration support with subscriptions
type DB struct {
	*shed.DB                                           // underlying shed.DB
	update   func(*Batch, Mode, *shed.IndexItem) error // mode-dependent update method
	access   func(Mode, *shed.IndexItem) error         // mode dependent access method
	batch    chan *Batch                               // channel to obtain current batch
	quit     chan struct{}                             // channel to be closed when DB quits
}

// New constructs a new DB
func New(sdb *shed.DB, update func(*Batch, Mode, *shed.IndexItem) error, access func(Mode, *shed.IndexItem) error) *DB {
	db := &DB{
		DB:     sdb,
		update: update,
		access: access,
		batch:  make(chan *Batch),
		quit:   make(chan struct{}),
	}
	go db.listen()
	return db
}

// Close terminates loops by closing the quit channel
func (db *DB) Close() {
	// signal quit to listen loop
	close(db.quit)
	// wait till batch channel is closed and last batch is written
	for b := range db.batch {
		b.wg.Done()
		<-b.Done
	}
	// close shed db
	db.DB.Close()
}

// Accessor is a wrapper around DB where Put/Get is overwritten to apply the
// update/access method for the mode
// using With(mode) the DB implements the ChunkStore interface
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
	b := <-db.batch
	log.Debug("obtained batch")
	if b == nil {
		return errDBClosed
	}
	// call the update with the  access mode
	err := db.update(b, mode, item)
	if err != nil {
		return err
	}
	// signal to listen loop that the update to batch is complete
	b.wg.Done()
	// wait for batch to be written and return batch error
	// this is in order for Put calls to be synchronous
	select {
	case <-b.Done:
	case <-ctx.Done():
		return ctx.Err()
	}
	return b.Err
}

// listen is a forever loop handing out the current batch to updaters
// and apply the batch when the db is free
// if the db is quit, the last batch is written out and batch channel is closed
func (db *DB) listen() {
	b := newBatch()        // current batch
	var done chan struct{} //
	wasdone := make(chan struct{})
	close(wasdone)
	for {
		b.wg.Add(1)
		select {
		case db.batch <- b:
			// allow
			done = wasdone
		case <-done:
			b.wg.Done()
			// if batchwriter is idle, hand over the batch and creates a new one
			// if batchwriter loop is busy, keep adding to the same batch
			go db.writeBatch(b)
			wasdone = b.Done
			// disable case until further ops happen
			done = nil
			b = newBatch()
		case <-db.quit:
			// make sure batch is saved to disk so as not to lose chunks
			if done != nil {
				b.wg.Done()
				db.writeBatch(b)
				<-b.Done
			}
			close(db.batch)
			return
		}
	}
}

// writeBatch writes out the batch, sets the error and closes the done channel
func (db *DB) writeBatch(b *Batch) {
	// wait for all updaters to finish writing to this batch
	b.wg.Wait()
	// apply the batch
	b.Err = db.DB.WriteBatch(b.Batch)
	// signal batch write to callers
	close(b.Done)
}

/*
	Address         []byte
	Data            []byte
	AccessTimestamp int64
	StoreTimestamp  int64
*/
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
