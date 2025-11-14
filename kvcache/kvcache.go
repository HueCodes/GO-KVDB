package kvcache

import (
	"hash"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
)

// CacheEntry represents a single key-value pair with expiration
type CacheEntry struct {
	Value      interface{}
	Expiration int64 // UnixNano timestamp for expiration (use atomic operations)
	lastAccess int64 // For LRU tracking
}

// KVCache is the main key-value cache structure
type KVCache struct {
	shards      []*shard
	numShards   int
	ttl         time.Duration
	maxCapacity int // Max entries per shard, 0 = unlimited
	entryPool   sync.Pool
	hashPool    sync.Pool

	// Metrics
	hits      atomic.Uint64
	misses    atomic.Uint64
	evictions atomic.Uint64
}

type shard struct {
	store map[string]*CacheEntry
	mutex sync.RWMutex
	size  int // Track size to avoid map iterations
}

// NewKVCache creates a new key-value cache with specified TTL
func NewKVCache(defaultTTL time.Duration) *KVCache {
	return NewKVCacheWithCapacity(defaultTTL, 0)
}

// NewKVCacheWithCapacity creates a cache with TTL and max capacity per shard
func NewKVCacheWithCapacity(defaultTTL time.Duration, maxCapacityPerShard int) *KVCache {
	numShards := 256 // Increased from 16 for better concurrency
	shards := make([]*shard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = &shard{
			store: make(map[string]*CacheEntry),
		}
	}
	cache := &KVCache{
		shards:      shards,
		numShards:   numShards,
		ttl:         defaultTTL,
		maxCapacity: maxCapacityPerShard,
		entryPool: sync.Pool{
			New: func() interface{} {
				return &CacheEntry{}
			},
		},
		hashPool: sync.Pool{
			New: func() interface{} {
				return fnv.New32a()
			},
		},
	}
	// Start cleanup routine
	go cache.cleanup()
	return cache
}

// getShard returns the shard for a given key using pooled hash
func (c *KVCache) getShard(key string) *shard {
	h := c.hashPool.Get().(hash.Hash32)
	h.Reset()
	h.Write([]byte(key))
	idx := h.Sum32() % uint32(c.numShards)
	c.hashPool.Put(h)
	return c.shards[idx]
}

// Set adds or updates a key-value pair with optional custom TTL
func (c *KVCache) Set(key string, value interface{}, ttl ...time.Duration) {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	expiration := time.Now().Add(c.ttl).UnixNano()
	if len(ttl) > 0 && ttl[0] > 0 {
		expiration = time.Now().Add(ttl[0]).UnixNano()
	}

	// Check if we need to evict (LRU) before adding
	if c.maxCapacity > 0 && shard.size >= c.maxCapacity {
		if _, exists := shard.store[key]; !exists {
			// Need to evict - find oldest entry
			c.evictOldest(shard)
		}
	}

	// Get entry from pool or create new
	entry, ok := shard.store[key]
	if !ok {
		entry = c.entryPool.Get().(*CacheEntry)
		shard.size++
	}

	entry.Value = value
	atomic.StoreInt64(&entry.Expiration, expiration)
	atomic.StoreInt64(&entry.lastAccess, time.Now().UnixNano())

	shard.store[key] = entry
}

// Get retrieves a value by key, returning nil if not found or expired
func (c *KVCache) Get(key string) (interface{}, bool) {
	shard := c.getShard(key)
	shard.mutex.RLock()

	entry, exists := shard.store[key]
	if !exists {
		shard.mutex.RUnlock()
		c.misses.Add(1)
		return nil, false
	}

	// Fast path: atomic expiration check without lock
	expiration := atomic.LoadInt64(&entry.Expiration)
	now := time.Now().UnixNano()

	if expiration > 0 && now > expiration {
		// Entry expired - upgrade to write lock for lazy deletion
		shard.mutex.RUnlock()
		shard.mutex.Lock()
		// Double-check after acquiring write lock
		if entry, exists := shard.store[key]; exists {
			exp := atomic.LoadInt64(&entry.Expiration)
			if exp > 0 && time.Now().UnixNano() > exp {
				delete(shard.store, key)
				shard.size--
				c.entryPool.Put(entry)
			}
		}
		shard.mutex.Unlock()
		c.misses.Add(1)
		return nil, false
	}

	// Update last access time for LRU (atomic, no lock needed)
	atomic.StoreInt64(&entry.lastAccess, now)
	value := entry.Value
	shard.mutex.RUnlock()

	c.hits.Add(1)
	return value, true
}

// Delete removes a key-value pair
func (c *KVCache) Delete(key string) {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if entry, exists := shard.store[key]; exists {
		delete(shard.store, key)
		shard.size--
		c.entryPool.Put(entry)
	}
}

// evictOldest removes the least recently used entry from a shard
// Must be called with shard.mutex held
func (c *KVCache) evictOldest(s *shard) {
	var oldestKey string
	var oldestTime int64 = 1<<63 - 1 // Max int64

	for key, entry := range s.store {
		lastAccess := atomic.LoadInt64(&entry.lastAccess)
		if lastAccess < oldestTime {
			oldestTime = lastAccess
			oldestKey = key
		}
	}

	if oldestKey != "" {
		if entry := s.store[oldestKey]; entry != nil {
			delete(s.store, oldestKey)
			s.size--
			c.entryPool.Put(entry)
			c.evictions.Add(1)
		}
	}
}

// cleanup periodically removes expired entries with minimal blocking
func (c *KVCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().UnixNano()

		for _, shard := range c.shards {
			// Collect expired keys with read lock first
			shard.mutex.RLock()
			expiredKeys := make([]string, 0, 16)

			for key, entry := range shard.store {
				expiration := atomic.LoadInt64(&entry.Expiration)
				if expiration > 0 && now > expiration {
					expiredKeys = append(expiredKeys, key)
				}
			}
			shard.mutex.RUnlock()

			// Delete expired entries with write lock (if any found)
			if len(expiredKeys) > 0 {
				shard.mutex.Lock()
				for _, key := range expiredKeys {
					// Double-check expiration after acquiring lock
					if entry, exists := shard.store[key]; exists {
						exp := atomic.LoadInt64(&entry.Expiration)
						if exp > 0 && now > exp {
							delete(shard.store, key)
							shard.size--
							c.entryPool.Put(entry)
						}
					}
				}
				shard.mutex.Unlock()
			}
		}
	}
}

// Size returns the number of entries in the cache
func (c *KVCache) Size() int {
	total := 0
	for _, shard := range c.shards {
		shard.mutex.RLock()
		total += shard.size
		shard.mutex.RUnlock()
	}
	return total
}

// Stats returns cache statistics
func (c *KVCache) Stats() CacheStats {
	return CacheStats{
		Hits:      c.hits.Load(),
		Misses:    c.misses.Load(),
		Evictions: c.evictions.Load(),
		Size:      uint64(c.Size()),
	}
}

// CacheStats holds cache performance metrics
type CacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	Size      uint64
}

// HitRate returns the cache hit rate as a percentage
func (s CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// SetMulti sets multiple key-value pairs in a single operation
func (c *KVCache) SetMulti(entries map[string]interface{}, ttl ...time.Duration) {
	for key, value := range entries {
		c.Set(key, value, ttl...)
	}
}

// GetMulti retrieves multiple values by keys
func (c *KVCache) GetMulti(keys []string) map[string]interface{} {
	result := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		if value, ok := c.Get(key); ok {
			result[key] = value
		}
	}
	return result
}

// Clear removes all entries from the cache
func (c *KVCache) Clear() {
	for _, shard := range c.shards {
		shard.mutex.Lock()
		for key, entry := range shard.store {
			delete(shard.store, key)
			c.entryPool.Put(entry)
		}
		shard.size = 0
		shard.mutex.Unlock()
	}
}
