# Fast-Cache

A high-performance, thread-safe in-memory key-value cache for Go with TTL support and LRU eviction.

## Features

- Type-safe generic API (Go 1.18+) with zero allocations for value types
- Fast concurrent access: ~40ns reads, ~80ns writes
- Thread-safe with 256-way sharding for minimal lock contention
- TTL support with lazy expiration and background cleanup
- LRU eviction for capacity-limited caches
- Built-in performance metrics (hits, misses, evictions, hit rate)
- Context support for cancellation and timeouts
- Graceful shutdown with Close() method

## Installation

```bash
go get github.com/HueCodes/Fast-Cache/kvcache
```

Requires Go 1.21 or later.

## Quick Start

### Generic API (Recommended)

```go
package main

import (
    "fmt"
    "time"
    "github.com/HueCodes/Fast-Cache/kvcache"
)

func main() {
    // Create type-safe cache
    cache := kvcache.New[string, int](5 * time.Minute)
    defer cache.Close()

    // Set and get values
    cache.Set("counter", 42)
    value, exists := cache.Get("counter")
    if exists {
        fmt.Printf("Counter: %d\n", value)
    }

    // Custom TTL per entry
    cache.Set("session", 12345, 30*time.Second)
}
```

### Legacy API

```go
cache := kvcache.NewKVCache(5 * time.Minute)

cache.Set("user:1", map[string]string{
    "name": "Alice",
    "role": "admin",
})

if user, ok := cache.Get("user:1"); ok {
    fmt.Printf("User: %v\n", user)
}
```

## Advanced Usage

### Capacity-Limited Cache with LRU

```go
// 100 entries per shard × 256 shards = 25,600 total capacity
cache := kvcache.NewWithCapacity[string, User](5*time.Minute, 100)
defer cache.Close()

// Automatically evicts least-recently-used entries when full
for i := 0; i < 100000; i++ {
    cache.Set(fmt.Sprintf("user:%d", i), User{ID: i})
}
```

### Custom Configuration

```go
cache := kvcache.NewWithConfig[string, Data](kvcache.Config{
    DefaultTTL:          10 * time.Minute,
    MaxCapacityPerShard: 1000,
    NumShards:           512,
    CleanupInterval:     30 * time.Second,
})
defer cache.Close()
```

### Performance Metrics

```go
stats := cache.Stats()
fmt.Printf("Hit Rate: %.2f%%\n", stats.HitRate())
fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)
fmt.Printf("Evictions: %d\n", stats.Evictions)
```

### Context Support

```go
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

value, err := cache.GetWithContext(ctx, "key")
if err != nil {
    // Handle timeout or cancellation
}
```

## Performance

Benchmarks on Apple M2, Go 1.25.1:

```
BenchmarkGenericConcurrentReads-8     50,000,000      40 ns/op      0 B/op      0 allocs/op
BenchmarkGenericConcurrentWrites-8    30,000,000      78 ns/op      0 B/op      0 allocs/op
BenchmarkMixedWorkload-8              35,000,000      65 ns/op      0 B/op      0 allocs/op
```

The generic API provides type safety with zero allocations for value types. The legacy interface{} API has 2 allocs/op due to boxing.

## Architecture

Fast-Cache uses a sharded hash map design with 256 independent shards, each protected by its own RWMutex. This enables true parallel access across different keys with minimal lock contention.

Key optimizations:
- sync.Pool for entry and hash object recycling
- Atomic operations for lock-free expiration checks
- Random sampling for O(1) LRU eviction
- Non-blocking background cleanup
- Cache-line padding to prevent false sharing

## API Reference

### Generic API

#### Creation
```go
func New[K comparable, V any](defaultTTL time.Duration) *Cache[K, V]
func NewWithCapacity[K comparable, V any](defaultTTL time.Duration, maxCapacityPerShard int) *Cache[K, V]
func NewWithConfig[K comparable, V any](cfg Config) *Cache[K, V]
```

#### Operations
```go
func (c *Cache[K, V]) Set(key K, value V, ttl ...time.Duration)
func (c *Cache[K, V]) Get(key K) (V, bool)
func (c *Cache[K, V]) GetWithContext(ctx context.Context, key K) (V, error)
func (c *Cache[K, V]) Delete(key K)
func (c *Cache[K, V]) Clear()
func (c *Cache[K, V]) Close() error
```

#### Metrics
```go
func (c *Cache[K, V]) Size() int
func (c *Cache[K, V]) Stats() CacheStats
```

### Legacy API

```go
func NewKVCache(defaultTTL time.Duration) *KVCache
func NewKVCacheWithCapacity(defaultTTL time.Duration, maxCapacityPerShard int) *KVCache

func (c *KVCache) Set(key string, value interface{}, ttl ...time.Duration)
func (c *KVCache) Get(key string) (interface{}, bool)
func (c *KVCache) Delete(key string)
func (c *KVCache) SetMulti(entries map[string]interface{}, ttl ...time.Duration)
func (c *KVCache) GetMulti(keys []string) map[string]interface{}
func (c *KVCache) Clear()
func (c *KVCache) Close() error
func (c *KVCache) Size() int
func (c *KVCache) Stats() CacheStats
```

### Types

```go
type Config struct {
    DefaultTTL          time.Duration
    MaxCapacityPerShard int
    NumShards           int
    CleanupInterval     time.Duration
}

type CacheStats struct {
    Hits      uint64
    Misses    uint64
    Evictions uint64
    Size      uint64
}

func (s CacheStats) HitRate() float64
```

## Testing

```bash
# Run tests
go test -v ./kvcache

# Run with race detector
go test -race ./kvcache

# Run benchmarks
go test -bench=. -benchmem ./kvcache
```

## Use Cases

Ideal for:
- High-throughput web servers (session storage, API response caching)
- Read-heavy workloads
- Rate limiting and request deduplication
- Database query result caching
- Configuration caching
- Microservices with local cache needs

Not suitable for:
- Persistent storage (in-memory only)
- Distributed caching across multiple nodes
- Multi-gigabyte datasets without compression

## Thread Safety

All operations are thread-safe and designed for high concurrency. The cache has been tested with the Go race detector and handles unlimited concurrent goroutines.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License. See LICENSE file for details.

## Acknowledgments

Inspired by groupcache, BigCache, and Ristretto.
