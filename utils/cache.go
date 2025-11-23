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
	data   map[string]item
	mu     sync.RWMutex
	ticker *time.Ticker
}

var cacheInstance *TTLCache

func init() {
	cacheInstance = NewTTLCache(1 * time.Minute)
}

func GetCacheInstance() *TTLCache {
	return cacheInstance
}

// It will create a new TTL cache with the specified cleanup interval.
func NewTTLCache(cleanUpInterval time.Duration) *TTLCache {
	cache := &TTLCache{
		data:   make(map[string]item),
		ticker: time.NewTicker(cleanUpInterval),
	}

	// This goroutine will periodically clean up expired items.
	go func() {
		for range cache.ticker.C {
			cache.cleanup()
		}
	}()

	return cache
}

func (c *TTLCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = item{
		value:     value,
		expiresAt: time.Now().Add(ttl).UnixNano(),
	}
}

// Get retrieves an item from the cache and refreshes its TTL if found and not expired.
func (c *TTLCache) Get(key string) (any, bool) {
	c.mu.RLock()
	it, found := c.data[key]
	c.mu.RUnlock()

	if !found || time.Now().UnixNano() > it.expiresAt {
		return nil, false
	}

	// Refresh TTL optimistically under read lock.
	c.mu.Lock()
	// double-check if still valid after lock switch
	it2, stillFound := c.data[key]
	if stillFound && time.Now().UnixNano() <= it2.expiresAt {
		it2.expiresAt = time.Now().Add(5 * time.Minute).UnixNano()
		c.data[key] = it2
	}
	c.mu.Unlock()

	return it.value, true
}

func (c *TTLCache) cleanup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, it := range c.data {
		if now > it.expiresAt {
			delete(c.data, key)
		}
	}
}
