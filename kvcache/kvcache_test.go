package kvcache

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestBasicOperations tests core functionality
func TestBasicOperations(t *testing.T) {
	cache := NewKVCache(5 * time.Minute)

	// Test Set and Get
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}

	// Test Delete
	cache.Delete("key1")
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Key should be deleted")
	}
}

// TestExpiration tests TTL functionality
func TestExpiration(t *testing.T) {
	cache := NewKVCache(100 * time.Millisecond)

	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Key should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Key should be expired")
	}
}

// TestCustomTTL tests per-key TTL
func TestCustomTTL(t *testing.T) {
	cache := NewKVCache(5 * time.Minute)

	cache.Set("short", "value", 50*time.Millisecond)
	cache.Set("long", "value", 5*time.Minute)

	time.Sleep(100 * time.Millisecond)

	_, ok := cache.Get("short")
	if ok {
		t.Error("Short TTL key should be expired")
	}

	_, ok = cache.Get("long")
	if !ok {
		t.Error("Long TTL key should still exist")
	}
}

// TestCapacityLimit tests LRU eviction
func TestCapacityLimit(t *testing.T) {
	cache := NewKVCacheWithCapacity(5*time.Minute, 10) // 10 items per shard

	// Fill cache beyond capacity
	for i := 0; i < 3000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	stats := cache.Stats()
	if stats.Evictions == 0 {
		t.Error("Expected evictions due to capacity limit")
	}

	// Most recent keys should still exist
	_, ok := cache.Get("key2999")
	if !ok {
		t.Error("Recent key should exist")
	}
}

// TestConcurrency tests thread safety
func TestConcurrency(t *testing.T) {
	cache := NewKVCache(5 * time.Minute)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Set(fmt.Sprintf("key%d", n), n)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Get(fmt.Sprintf("key%d", n))
		}(i)
	}

	wg.Wait()

	// Verify all keys exist
	for i := 0; i < 100; i++ {
		_, ok := cache.Get(fmt.Sprintf("key%d", i))
		if !ok {
			t.Errorf("Key %d should exist", i)
		}
	}
}

// TestStats tests metrics collection
func TestStats(t *testing.T) {
	cache := NewKVCache(5 * time.Minute)

	cache.Set("key1", "value1")
	cache.Get("key1") // Hit
	cache.Get("key2") // Miss

	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.HitRate() != 50.0 {
		t.Errorf("Expected 50%% hit rate, got %.2f%%", stats.HitRate())
	}
}

// TestBatchOperations tests SetMulti and GetMulti
func TestBatchOperations(t *testing.T) {
	cache := NewKVCache(5 * time.Minute)

	entries := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	cache.SetMulti(entries)

	keys := []string{"key1", "key2", "key3", "key4"}
	results := cache.GetMulti(keys)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results["key1"] != "value1" {
		t.Error("Incorrect value for key1")
	}
}

// BenchmarkSet measures write performance
func BenchmarkSet(b *testing.B) {
	cache := NewKVCache(5 * time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}
}

// BenchmarkGet measures read performance
func BenchmarkGet(b *testing.B) {
	cache := NewKVCache(5 * time.Minute)

	// Pre-populate
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key%d", i%10000))
	}
}

// BenchmarkConcurrentReads tests parallel read performance
func BenchmarkConcurrentReads(b *testing.B) {
	cache := NewKVCache(5 * time.Minute)

	// Pre-populate
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get(fmt.Sprintf("key%d", i%10000))
			i++
		}
	})
}

// BenchmarkConcurrentWrites tests parallel write performance
func BenchmarkConcurrentWrites(b *testing.B) {
	cache := NewKVCache(5 * time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Set(fmt.Sprintf("key%d", i), i)
			i++
		}
	})
}

// BenchmarkMixedWorkload tests realistic mixed read/write scenario
func BenchmarkMixedWorkload(b *testing.B) {
	cache := NewKVCache(5 * time.Minute)

	// Pre-populate
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if rand.Intn(100) < 80 { // 80% reads, 20% writes
				cache.Get(fmt.Sprintf("key%d", i%10000))
			} else {
				cache.Set(fmt.Sprintf("key%d", i%10000), i)
			}
			i++
		}
	})
}

// BenchmarkMemoryAllocation tests GC pressure
func BenchmarkMemoryAllocation(b *testing.B) {
	cache := NewKVCache(5 * time.Minute)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key%d", i%1000), i)
		cache.Get(fmt.Sprintf("key%d", i%1000))
	}

	runtime.ReadMemStats(&m2)
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "B/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}
