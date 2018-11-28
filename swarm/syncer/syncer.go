package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// Syncer binds together
// - syncdb
// - protocol (dispatcher)
// - pubsub transport layer
type Syncer struct {
	db *DB // sync db
}

// New constructs a Syncer
func New(dbpath string, baseAddr storage.Address, store storage.ChunkStore, ps PubSub) (*Syncer, error) {
	d := newDispatcher(baseAddr).withPubSub(ps)
	receiptsC := make(chan storage.Address)
	db, err := NewDB(dbpath, store, d.sendChunk, receiptsC)
	if err != nil {
		return nil, err
	}
	d.processReceipt = func(addr storage.Address) error {
		receiptsC <- addr
		return nil
	}
	return &Syncer{db: db}, nil
}

// Close closes the syncer
func (s *Syncer) Close() {
	s.db.Close()
}

// Put puts the chunk to storage and inserts into sync index
// currently chunkstore call is syncronous so it needs to
// wrap dbstore
func (s *Syncer) Put(tagname string, chunk storage.Chunk) {
	it := &item{
		Addr:  chunk.Address(),
		Tag:   tagname,
		chunk: chunk,
		state: SPLIT, // can be left explicit
	}
	s.db.tags.Inc(tagname, SPLIT)
	// this put returns with error if this is a duplicate
	err := s.db.chunkStore.Put(context.TODO(), chunk)
	if err == errExists {
		return
	}
	if err != nil {
		log.Error("syncer: error storing chunk: %v", err)
		return
	}
	s.db.Put(it)
}

// NewTag creates a new info object for a file/collection of chunks
func (s *Syncer) NewTag(name string, total int) (*Tag, error) {
	return s.db.tags.New(name, total)
}

// Status returns the number of chunks in a state tagged with tag
func (s *Syncer) Status(name string, state State) (int, int) {
	v, _ := s.db.tags.tags.Load(name)
	return v.(*Tag).Status(state)
}
