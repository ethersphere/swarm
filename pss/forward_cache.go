package pss

import (
	"fmt"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/log"
	"golang.org/x/crypto/sha3"
	"hash"
	"sync"
	"time"
)

const (
	hasherCount  = 8
	digestLength = 32 // byte length of digest used for pss forward cache (currently same as swarm chunk hash)
)

// ForwadrCache is used for preventing backwards routing
// will also be instrumental in flood guard mechanism
// and mailbox implementation
type ForwardCache struct {
	fwdCache   map[digest]cacheEntry // checksum of unique fields from pssmsg mapped to expiry, cache to determine whether to drop msg
	fwdCacheMu sync.RWMutex
	cacheTTL   time.Duration // how long to keep messages in fwdCache (not implemented)
	hashPool   sync.Pool
}

type cacheEntry struct {
	expiresAt time.Time
}

type digest [digestLength]byte

func newForwardCache(cacheTTL time.Duration) *ForwardCache {
	fc := &ForwardCache{
		fwdCache: make(map[digest]cacheEntry),
		cacheTTL: cacheTTL,
		hashPool: sync.Pool{
			New: func() interface{} {
				return sha3.NewLegacyKeccak256()
			},
		},
	}
	for i := 0; i < hasherCount; i++ {
		hashfunc := sha3.NewLegacyKeccak256()
		fc.hashPool.Put(hashfunc)
	}
	return fc
}

// add a message to the cache
func (fc *ForwardCache) addFwdCache(msg *PssMsg) error {
	defer metrics.GetOrRegisterResettingTimer("pss.addfwdcache", nil).UpdateSince(time.Now())

	var entry cacheEntry
	var ok bool

	fc.fwdCacheMu.Lock()
	defer fc.fwdCacheMu.Unlock()

	digest := fc.msgDigest(msg)
	if entry, ok = fc.fwdCache[digest]; !ok {
		entry = cacheEntry{}
	}
	entry.expiresAt = time.Now().Add(fc.cacheTTL)
	fc.fwdCache[digest] = entry
	return nil
}

// check if message is in the cache
func (fc *ForwardCache) checkFwdCache(msg *PssMsg) bool {
	fc.fwdCacheMu.Lock()
	defer fc.fwdCacheMu.Unlock()

	digest := fc.msgDigest(msg)
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
func (fc *ForwardCache) cleanFwdCache() {
	metrics.GetOrRegisterCounter("pss.cleanfwdcache", nil).Inc(1)
	fc.fwdCacheMu.Lock()
	defer fc.fwdCacheMu.Unlock()
	for k, v := range fc.fwdCache {
		if v.expiresAt.Before(time.Now()) {
			delete(fc.fwdCache, k)
		}
	}
}

// Digest of message
func (fc *ForwardCache) msgDigest(msg *PssMsg) digest {
	return fc.digestBytes(msg.serialize())
}

func (fc *ForwardCache) digestBytes(msg []byte) digest {
	hasher := fc.hashPool.Get().(hash.Hash)
	defer fc.hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(msg)
	d := digest{}
	key := hasher.Sum(nil)
	copy(d[:], key[:digestLength])
	return d
}
