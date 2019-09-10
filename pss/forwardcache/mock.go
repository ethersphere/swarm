package forwardcache

import (
	"github.com/ethersphere/swarm/pss/message"
	"time"
)

const (
	defaultCacheTTL = time.Second * 10
)

func NewMockForwardCache(config *Config) *forwardCache {
	if config == nil {
		config = &Config{
			CacheTTL: defaultCacheTTL,
			QuitC:    nil,
		}
	}

	if &config.CacheTTL == nil {
		config.CacheTTL = defaultCacheTTL
	}
	fc := &forwardCache{
		fwdCache: make(map[message.Digest]cacheEntry),
		Config:   *config,
	}
	return fc

}
