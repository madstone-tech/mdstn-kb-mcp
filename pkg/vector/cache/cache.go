// Package cache provides an in-memory LRU cache for vector embeddings.
package cache

import (
	"container/list"
	"sync"
	"time"
)

// Entry represents a cache entry with TTL
type Entry struct {
	Value     []float64
	ExpiresAt time.Time
}

// IsExpired checks if the entry has expired
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache is a thread-safe LRU cache with TTL support
type Cache struct {
	mu      sync.RWMutex
	maxSize int
	ttl     time.Duration
	items   map[string]*list.Element
	list    *list.List
	hits    int64
	misses  int64
	evicts  int64
}

// cacheItem represents an item in the LRU list
type cacheItem struct {
	key   string
	value *Entry
}

// NewCache creates a new LRU cache with the specified max size and TTL
func NewCache(maxSize int, ttl time.Duration) *Cache {
	if maxSize <= 0 {
		maxSize = 100
	}
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}

	return &Cache{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*list.Element),
		list:    list.New(),
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) ([]float64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	item := elem.Value.(*cacheItem)
	if item.value.IsExpired() {
		c.list.Remove(elem)
		delete(c.items, key)
		c.misses++
		return nil, false
	}

	c.list.MoveToFront(elem)
	c.hits++
	return item.value.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.list.MoveToFront(elem)
		item := elem.Value.(*cacheItem)
		item.value = &Entry{
			Value:     value,
			ExpiresAt: time.Now().Add(c.ttl),
		}
		return
	}

	item := &cacheItem{
		key: key,
		value: &Entry{
			Value:     value,
			ExpiresAt: time.Now().Add(c.ttl),
		},
	}

	elem := c.list.PushFront(item)
	c.items[key] = elem

	if c.list.Len() > c.maxSize {
		c.evictOldest()
	}
}

// evictOldest removes the least recently used item
func (c *Cache) evictOldest() {
	elem := c.list.Back()
	if elem != nil {
		c.list.Remove(elem)
		item := elem.Value.(*cacheItem)
		delete(c.items, item.key)
		c.evicts++
	}
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.list = list.New()
	c.hits = 0
	c.misses = 0
	c.evicts = 0
}

// Size returns the current number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats returns cache statistics
func (c *Cache) Stats() (hits, misses, evicts int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, c.evicts
}

// HitRate returns the cache hit rate (0.0 to 1.0)
func (c *Cache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0.0
	}
	return float64(c.hits) / float64(total)
}

// Delete removes a specific key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.list.Remove(elem)
		delete(c.items, key)
	}
}

// Has checks if a key exists in the cache and is not expired
func (c *Cache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, exists := c.items[key]
	if !exists {
		return false
	}

	item := elem.Value.(*cacheItem)
	return !item.value.IsExpired()
}

// SetTTL updates the TTL for new entries
func (c *Cache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}

// CleanupExpired removes all expired entries
func (c *Cache) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var toDelete []string
	for key, elem := range c.items {
		item := elem.Value.(*cacheItem)
		if item.value.IsExpired() {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		elem := c.items[key]
		c.list.Remove(elem)
		delete(c.items, key)
	}
}
