package ttlset

import (
	"errors"
	"sync"
	"time"

	"github.com/tilinna/clock"
)

// TTLSet implements a Set that automatically removes expired items
// after a predefined expiration time
type TTLSet interface {
	Add(key interface{}) error // Add adds a new key to the set
	Has(key interface{}) bool  // Check returns whether or not the key is already/still in the set
	Start() error              // Start launches this service
	Stop() error               // Stop will close the service and release all resources
}

// Config defines the TTLSet configuration
type Config struct {
	EntryTTL time.Duration // time after which items are removed
	Clock    clock.Clock   // time reference
}

type setEntry struct {
	expiresAt time.Time
}

type ttlSet struct {
	Config
	quitC chan struct{}
	set   map[interface{}]setEntry
	lock  sync.RWMutex
}

// ErrAlreadyStarted is returned if this service was already started and Start() is called again
var ErrAlreadyStarted = errors.New("Already started")

// ErrAlreadyStopped is returned if this service was already stopped and Stop() is called again
var ErrAlreadyStopped = errors.New("Already stopped")

// New instances a the default ForwardCache implementation
func New(config *Config) TTLSet {
	ts := &ttlSet{
		set:    make(map[interface{}]setEntry),
		Config: *config,
	}
	return ts
}

// Add adds a new key to the set
func (ts *ttlSet) Add(key interface{}) error {
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
func (ts *ttlSet) Has(key interface{}) bool {
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

// clean is used to periodically remove expired entries from the set
func (ts *ttlSet) clean() {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	for k, v := range ts.set {
		if v.expiresAt.Before(ts.Clock.Now()) {
			delete(ts.set, k)
		}
	}
}

func (ts *ttlSet) newTicker(callback func()) {
	ticker := ts.Clock.NewTicker(ts.EntryTTL)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				callback()
			case <-ts.quitC:
				return
			}
		}
	}()
}

// Start launches this service
func (ts *ttlSet) Start() error {
	if ts.quitC != nil {
		return ErrAlreadyStarted
	}
	ts.quitC = make(chan struct{})
	ts.newTicker(func() {
		ts.clean()
	})

	return nil
}

// Stop will close the service and release all resources
func (ts *ttlSet) Stop() error {
	if ts.quitC == nil {
		return ErrAlreadyStopped
	}
	close(ts.quitC)
	ts.quitC = nil
	return nil
}
