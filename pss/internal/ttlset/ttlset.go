package ttlset

import (
	"sync"
	"time"

	"github.com/ethersphere/swarm/pss/internal/ttlset/ticker"
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
	set    map[interface{}]setEntry
	lock   sync.RWMutex
	ticker *ticker.Ticker
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

	ticker := ticker.New(&ticker.Config{
		Interval: config.EntryTTL,
		Clock:    config.Clock,
		Callback: func() {
			ts.clean()
		},
	})
	ts.ticker = ticker

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

// clean is used internally to periodically remove expired entries from the set
func (ts *TTLSet) clean() {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	for k, v := range ts.set {
		if v.expiresAt.Before(ts.Clock.Now()) {
			delete(ts.set, k)
		}
	}
}

// Stop will close the service and release all resources
func (ts *TTLSet) Stop() error {
	return ts.ticker.Stop()
}
