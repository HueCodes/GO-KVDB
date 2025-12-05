package kvcache_test

import (
	"fmt"
	"time"

	"github.com/HueCodes/Fast-Cache/kvcache"
)

func ExampleNewKVCache() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	cache.Set("user:1", "Alice")
	value, exists := cache.Get("user:1")
	if exists {
		fmt.Println(value)
	}
	// Output: Alice
}

func ExampleKVCache_Set() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	// Set with default TTL
	cache.Set("key1", "value1")

	// Set with custom TTL
	cache.Set("key2", "value2", 30*time.Second)

	fmt.Println("Values set successfully")
	// Output: Values set successfully
}

func ExampleKVCache_Get() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	cache.Set("greeting", "Hello, World!")

	value, exists := cache.Get("greeting")
	if exists {
		fmt.Println(value)
	}
	// Output: Hello, World!
}

func ExampleKVCache_Stats() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Get("key1")
	cache.Get("key2")

	stats := cache.Stats()
	fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)
	// Output: Hits: 1, Misses: 1
}

func ExampleNewKVCacheWithCapacity() {
	// Create cache with capacity limit of 100 entries per shard
	cache := kvcache.NewKVCacheWithCapacity(5*time.Minute, 100)
	defer cache.Close()

	// LRU eviction kicks in when capacity is exceeded
	for i := 0; i < 30000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	stats := cache.Stats()
	fmt.Printf("Evictions occurred: %t\n", stats.Evictions > 0)
	// Output: Evictions occurred: true
}

func ExampleKVCache_SetMulti() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	entries := map[string]interface{}{
		"user:1": "Alice",
		"user:2": "Bob",
		"user:3": "Charlie",
	}

	cache.SetMulti(entries)
	fmt.Printf("Size: %d\n", cache.Size())
	// Output: Size: 3
}

func ExampleKVCache_GetMulti() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	results := cache.GetMulti([]string{"a", "b", "c", "d"})
	fmt.Printf("Retrieved: %d values\n", len(results))
	// Output: Retrieved: 3 values
}

func ExampleKVCache_Clear() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	fmt.Printf("Before clear: %d\n", cache.Size())
	cache.Clear()
	fmt.Printf("After clear: %d\n", cache.Size())
	// Output: Before clear: 2
	// After clear: 0
}

func ExampleCacheStats_HitRate() {
	cache := kvcache.NewKVCache(5 * time.Minute)
	defer cache.Close()

	cache.Set("key1", "value1")

	// Generate some hits and misses
	for i := 0; i < 80; i++ {
		cache.Get("key1") // Hit
	}
	for i := 0; i < 20; i++ {
		cache.Get("missing") // Miss
	}

	stats := cache.Stats()
	fmt.Printf("Hit rate: %.0f%%\n", stats.HitRate())
	// Output: Hit rate: 80%
}
