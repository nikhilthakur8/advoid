package utils

import (
	"sync"
	"time"
)

type item struct {
	value     any
	expiresAt int64
}

type TTLCache struct {
	data   sync.Map
	ticker *time.Ticker
}

var cacheInstance *TTLCache

func init() {
	cacheInstance = NewTTLCache(1 * time.Minute)
}

func GetCacheInstance() *TTLCache {
	return cacheInstance
}

// NewTTLCache creates a TTL cache with periodic cleanup
func NewTTLCache(cleanUpInterval time.Duration) *TTLCache {
	cache := &TTLCache{
		ticker: time.NewTicker(cleanUpInterval),
	}

	// Cleanup routine
	go func() {
		for range cache.ticker.C {
			cache.cleanup()
		}
	}()

	return cache
}

// Set inserts/updates a key with TTL
func (c *TTLCache) Set(key string, value any, ttl time.Duration) {
	c.data.Store(key, item{
		value:     value,
		expiresAt: time.Now().Add(ttl).UnixNano(),
	})
}

// Get retrieves an item + refreshes TTL (5 minutes) if valid
func (c *TTLCache) Get(key string) (any, bool) {
	v, ok := c.data.Load(key)
	if !ok {
		return nil, false
	}

	it := v.(item)

	// refresh TTL to EXACTLY 1 minute
	it.expiresAt = time.Now().Add(1 * time.Minute).UnixNano()
	c.data.Store(key, it)

	return it.value, true
}

// cleanup removes expired keys
func (c *TTLCache) cleanup() {
	now := time.Now().UnixNano()

	c.data.Range(func(key, value any) bool {
		it := value.(item)
		if now > it.expiresAt {
			c.data.Delete(key)
		}
		return true
	})
}
