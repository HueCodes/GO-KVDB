package kvcache

import (
	"sync"
	"time"
)

// CacheEntry represents a single key-value pair with expiration
type CacheEntry struct {
	Value      interface{}
	Expiration int64 // UnixNano timestamp for expiration
}

// KVCache is the main key-value cache structure
type KVCache struct {
	store map[string]CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewKVCache creates a new key-value cache with specified TTL
func NewKVCache(defaultTTL time.Duration) *KVCache {
	cache := &KVCache{
		store: make(map[string]CacheEntry),
		ttl:   defaultTTL,
	}
	// Start cleanup routine
	go cache.cleanup()
	return cache
}

// Set adds or updates a key-value pair with optional custom TTL
func (c *KVCache) Set(key string, value interface{}, ttl ...time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(c.ttl).UnixNano()
	if len(ttl) > 0 && ttl[0] > 0 {
		expiration = time.Now().Add(ttl[0]).UnixNano()
	}

	c.store[key] = CacheEntry{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves a value by key, returning nil if not found or expired
func (c *KVCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.store[key]
	if !exists {
		return nil, false
	}

	if entry.Expiration > 0 && time.Now().UnixNano() > entry.Expiration {
		// Entry expired, remove it
		go c.Delete(key)
		return nil, false
	}

	return entry.Value, true
}

// Delete removes a key-value pair
func (c *KVCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.store, key)
}

// cleanup periodically removes expired entries
func (c *KVCache) cleanup() {
	for {
		time.Sleep(time.Minute)
		c.mutex.Lock()
		for key, entry := range c.store {
			if entry.Expiration > 0 && time.Now().UnixNano() > entry.Expiration {
				delete(c.store, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Size returns the number of entries in the cache
func (c *KVCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.store)
}
