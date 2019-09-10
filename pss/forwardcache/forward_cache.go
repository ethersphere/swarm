package forwardcache

import (
	"fmt"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pss/message"
	"sync"
	"time"
)

// ForwardCache is used for preventing backwards routing
// will also be instrumental in flood guard mechanism
// and mailbox implementation
type ForwardCache interface {
	AddFwdCache(msg *message.Message) error
	CheckFwdCache(msg *message.Message) bool
}

type Config struct {
	CacheTTL time.Duration
	QuitC    chan struct{}
}

type forwardCache struct {
	Config
	fwdCache   map[message.Digest]cacheEntry // checksum of unique fields from pssmsg mapped to expiry, cache to determine whether to drop msg
	fwdCacheMu sync.RWMutex
}

type cacheEntry struct {
	expiresAt time.Time
}

func NewForwardCache(config *Config) *forwardCache {
	fc := &forwardCache{
		fwdCache: make(map[message.Digest]cacheEntry),
		Config:   *config,
	}
	return fc

}

// add a message to the cache
func (fc *forwardCache) AddFwdCache(msg *message.Message) error {
	defer metrics.GetOrRegisterResettingTimer("pss.addfwdcache", nil).UpdateSince(time.Now())

	var entry cacheEntry
	var ok bool

	fc.fwdCacheMu.Lock()
	defer fc.fwdCacheMu.Unlock()

	digest := msg.Digest()
	if entry, ok = fc.fwdCache[digest]; !ok {
		entry = cacheEntry{}
	}
	entry.expiresAt = time.Now().Add(fc.CacheTTL)
	fc.fwdCache[digest] = entry
	return nil
}

// check if message is in the cache
func (fc *forwardCache) CheckFwdCache(msg *message.Message) bool {
	fc.fwdCacheMu.Lock()
	defer fc.fwdCacheMu.Unlock()

	digest := msg.Digest()
	entry, ok := fc.fwdCache[digest]
	if ok {
		if entry.expiresAt.After(time.Now()) {
			log.Trace("unexpired cache", "digest", fmt.Sprintf("%x", digest))
			metrics.GetOrRegisterCounter("pss.checkfwdcache.unexpired", nil).Inc(1)
			return true
		}
		metrics.GetOrRegisterCounter("pss.checkfwdcache.expired", nil).Inc(1)
	}
	return false
}

// cleanFwdCache is used to periodically remove expired entries from the forward cache
func (fc *forwardCache) cleanFwdCache() {
	metrics.GetOrRegisterCounter("pss.cleanfwdcache", nil).Inc(1)
	fc.fwdCacheMu.Lock()
	defer fc.fwdCacheMu.Unlock()
	for k, v := range fc.fwdCache {
		if v.expiresAt.Before(time.Now()) {
			delete(fc.fwdCache, k)
		}
	}
}

func (fc *forwardCache) startCacheCleaner() {
	go func() {
		cacheTicker := time.NewTicker(fc.CacheTTL)
		defer cacheTicker.Stop()
		for {
			select {
			case <-cacheTicker.C:
				fc.cleanFwdCache()
			case <-fc.QuitC:
				return
			}
		}
	}()
}
