package main

import (
	"fmt"
	"time"

	"kv-db-go/kvcache"
)

func main() {
	// Create a cache with 5-minute default TTL
	cache := kvcache.NewKVCache(5 * time.Minute)

	// Set a value with default TTL
	cache.Set("key1", "value1")

	// Set a value with custom TTL
	cache.Set("key2", "value2", 10*time.Second)

	// Get a value
	if value, exists := cache.Get("key1"); exists {
		fmt.Printf("key1: %v\n", value)
	}

	// Check cache size
	fmt.Printf("Cache size: %d\n", cache.Size())

	// Display performance stats
	stats := cache.Stats()
	fmt.Printf("Hit rate: %.2f%% (Hits: %d, Misses: %d)\n",
		stats.HitRate(), stats.Hits, stats.Misses)
}
