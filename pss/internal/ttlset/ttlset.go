package ttlset

import (
	"sync"
	"time"

	"github.com/tilinna/clock"
)

// Config defines the TTLSet configuration
type Config struct {
	EntryTTL time.Duration // time after which items are removed
	Clock    clock.Clock   // time reference
}

// TTLSet implements a Set that automatically removes expired keys
// after a predefined expiration time
type TTLSet struct {
	Config
	set  map[interface{}]setEntry
	lock sync.RWMutex
}

type setEntry struct {
	expiresAt time.Time
}

// New instances a TTLSet
func New(config *Config) *TTLSet {
	ts := &TTLSet{
		set:    make(map[interface{}]setEntry),
		Config: *config,
	}
	return ts
}

// Add adds a new key to the set
func (ts *TTLSet) Add(key interface{}) error {
	var entry setEntry
	var ok bool

	ts.lock.Lock()
	defer ts.lock.Unlock()

	if entry, ok = ts.set[key]; !ok {
		entry = setEntry{}
	}
	entry.expiresAt = ts.Clock.Now().Add(ts.EntryTTL)
	ts.set[key] = entry
	return nil
}

// Has returns whether or not a key is already/still in the set
func (ts *TTLSet) Has(key interface{}) bool {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	entry, ok := ts.set[key]
	if ok {
		if entry.expiresAt.After(ts.Clock.Now()) {
			return true
		}
		delete(ts.set, key) // since we're holding the lock, take the chance to delete a expired record
	}
	return false
}

// GC will remove expired entries from the set
func (ts *TTLSet) GC() {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	for k, v := range ts.set {
		if v.expiresAt.Before(ts.Clock.Now()) {
			delete(ts.set, k)
		}
	}
}

// Count returns the number of entries in the set
func (ts *TTLSet) Count() int {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	return len(ts.set)
}
