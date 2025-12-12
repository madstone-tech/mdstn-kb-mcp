package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheBasicOperations(t *testing.T) {
	cache := NewCache(3, 1*time.Hour)

	// Test Set and Get
	vec1 := []float64{1.0, 2.0, 3.0}
	cache.Set("key1", vec1)

	val, ok := cache.Get("key1")
	require.True(t, ok)
	assert.Equal(t, vec1, val)

	// Test non-existent key
	_, ok = cache.Get("nonexistent")
	assert.False(t, ok)
}

func TestCacheHitMiss(t *testing.T) {
	cache := NewCache(10, 1*time.Hour)

	// Test miss
	_, ok := cache.Get("missing")
	assert.False(t, ok)

	hits, misses, _ := cache.Stats()
	assert.Equal(t, int64(0), hits)
	assert.Equal(t, int64(1), misses)

	// Test hit
	cache.Set("key1", []float64{1.0})
	_, ok = cache.Get("key1")
	assert.True(t, ok)

	hits, misses, _ = cache.Stats()
	assert.Equal(t, int64(1), hits)
	assert.Equal(t, int64(1), misses)
}

func TestCacheLRUEviction(t *testing.T) {
	cache := NewCache(3, 1*time.Hour)

	// Fill cache
	cache.Set("key1", []float64{1.0})
	cache.Set("key2", []float64{2.0})
	cache.Set("key3", []float64{3.0})

	assert.Equal(t, 3, cache.Size())

	// Add one more (should evict key1 - least recently used)
	cache.Set("key4", []float64{4.0})

	assert.Equal(t, 3, cache.Size())
	_, ok := cache.Get("key1")
	assert.False(t, ok, "least recently used should be evicted")

	_, ok = cache.Get("key4")
	assert.True(t, ok, "newly added key should exist")

	_, _, evicts := cache.Stats()
	assert.Greater(t, evicts, int64(0))
}

func TestCacheLRUReordering(t *testing.T) {
	cache := NewCache(3, 1*time.Hour)

	cache.Set("key1", []float64{1.0})
	cache.Set("key2", []float64{2.0})
	cache.Set("key3", []float64{3.0})

	// Access key1 (moves it to front)
	cache.Get("key1")

	// Add key4 (should evict key2, not key1)
	cache.Set("key4", []float64{4.0})

	_, ok1 := cache.Get("key1")
	_, ok2 := cache.Get("key2")
	_, ok4 := cache.Get("key4")

	assert.True(t, ok1, "accessed key should still exist")
	assert.False(t, ok2, "least recently used (key2) should be evicted")
	assert.True(t, ok4, "newly added key should exist")
}

func TestCacheTTLExpiration(t *testing.T) {
	cache := NewCache(10, 100*time.Millisecond)

	cache.Set("key1", []float64{1.0})

	val, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, []float64{1.0}, val)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	_, ok = cache.Get("key1")
	assert.False(t, ok, "expired key should return false")

	assert.Equal(t, 0, cache.Size(), "expired entry should be removed from map when accessed")
}

func TestCacheClear(t *testing.T) {
	cache := NewCache(10, 1*time.Hour)

	cache.Set("key1", []float64{1.0})
	cache.Set("key2", []float64{2.0})

	assert.Equal(t, 2, cache.Size())

	cache.Clear()

	assert.Equal(t, 0, cache.Size())
	_, ok := cache.Get("key1")
	assert.False(t, ok)

	hits, misses, evicts := cache.Stats()
	assert.Equal(t, int64(0), hits)
	assert.Equal(t, int64(1), misses) // Get("key1") after clear is a miss
	assert.Equal(t, int64(0), evicts)
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache(10, 1*time.Hour)

	cache.Set("key1", []float64{1.0})
	cache.Set("key2", []float64{2.0})

	assert.Equal(t, 2, cache.Size())

	cache.Delete("key1")

	assert.Equal(t, 1, cache.Size())
	_, ok := cache.Get("key1")
	assert.False(t, ok)

	_, ok = cache.Get("key2")
	assert.True(t, ok)
}

func TestCacheHas(t *testing.T) {
	cache := NewCache(10, 1*time.Hour)

	assert.False(t, cache.Has("key1"))

	cache.Set("key1", []float64{1.0})
	assert.True(t, cache.Has("key1"))

	cache.Delete("key1")
	assert.False(t, cache.Has("key1"))
}

func TestCacheHitRate(t *testing.T) {
	cache := NewCache(10, 1*time.Hour)

	assert.Equal(t, 0.0, cache.HitRate())

	cache.Set("key1", []float64{1.0})

	// 2 hits
	cache.Get("key1")
	cache.Get("key1")

	// 1 miss
	cache.Get("nonexistent")

	hitRate := cache.HitRate()
	assert.Equal(t, 2.0/3.0, hitRate)
}

func TestCacheSetTTL(t *testing.T) {
	cache := NewCache(10, 100*time.Millisecond)

	cache.Set("key1", []float64{1.0})
	cache.SetTTL(1 * time.Hour)

	cache.Set("key2", []float64{2.0})

	// key1 should expire after 100ms
	time.Sleep(150 * time.Millisecond)

	_, ok1 := cache.Get("key1")
	assert.False(t, ok1, "key1 should be expired")

	// key2 should still be valid (1 hour TTL)
	_, ok2 := cache.Get("key2")
	assert.True(t, ok2, "key2 should still be valid")
}

func TestCacheCleanupExpired(t *testing.T) {
	cache := NewCache(10, 100*time.Millisecond)

	cache.Set("key1", []float64{1.0})
	cache.Set("key2", []float64{2.0})

	time.Sleep(150 * time.Millisecond)

	cache.SetTTL(1 * time.Hour)
	cache.Set("key3", []float64{3.0})

	assert.Equal(t, 3, cache.Size(), "should have 3 items before cleanup")

	cache.CleanupExpired()

	assert.Equal(t, 1, cache.Size(), "should have 1 item after cleanup")
	_, ok := cache.Get("key3")
	assert.True(t, ok, "non-expired key should exist")
}

func TestCacheThreadSafety(t *testing.T) {
	cache := NewCache(100, 1*time.Hour)
	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				key := "key_" + string(rune(id*10+j))
				val := []float64{float64(id*10 + j)}
				cache.Set(key, val)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Cache should have all entries
	assert.Greater(t, cache.Size(), 0)
}

func TestCacheDefaultSize(t *testing.T) {
	cache := NewCache(0, 1*time.Hour)
	assert.Greater(t, cache.maxSize, 0, "should have default size")
}

func TestCacheDefaultTTL(t *testing.T) {
	cache := NewCache(10, 0)
	require.Greater(t, cache.ttl.Nanoseconds(), int64(0), "should have default TTL")
}

func TestCacheOverwrite(t *testing.T) {
	cache := NewCache(10, 1*time.Hour)

	cache.Set("key1", []float64{1.0})
	val1, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, []float64{1.0}, val1)

	cache.Set("key1", []float64{2.0})
	val2, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, []float64{2.0}, val2)
	assert.Equal(t, 1, cache.Size(), "overwriting should not increase size")
}

func TestCacheEmptyVector(t *testing.T) {
	cache := NewCache(10, 1*time.Hour)

	emptyVec := []float64{}
	cache.Set("empty", emptyVec)

	val, ok := cache.Get("empty")
	assert.True(t, ok)
	assert.Equal(t, emptyVec, val)
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewCache(1000, 1*time.Hour)
	cache.Set("key", []float64{1.0, 2.0, 3.0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkCacheSet(b *testing.B) {
	cache := NewCache(1000, 1*time.Hour)
	vec := []float64{1.0, 2.0, 3.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", vec)
	}
}
