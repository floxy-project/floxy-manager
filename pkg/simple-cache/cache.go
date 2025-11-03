package simplecache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration time.
type CacheItem[T any] struct {
	Value     T
	ExpiresAt time.Time
}

// IsExpired checks if the cache item has expired.
func (item *CacheItem[T]) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// Cache is a simple in-memory cache with expiration.
type Cache[K comparable, V any] struct {
	items map[K]*CacheItem[V]
	mutex sync.RWMutex
}

// New creates a new cache instance.
func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]*CacheItem[V]),
	}
}

// Get retrieves a value from the cache
// Returns the value and a boolean indicating if the key was found and not expired.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists || item.IsExpired() {
		var zero V

		return zero, false
	}

	return item.Value, true
}

// Set stores a value in the cache with the given expiration duration.
func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem[V]{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache.
func (c *Cache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[K]*CacheItem[V])
}

// Cleanup removes all expired items from the cache.
func (c *Cache[K, V]) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}

// Size returns the number of items in the cache (including expired ones).
func (c *Cache[K, V]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// StartCleanup starts a background goroutine that periodically cleans up expired items.
func (c *Cache[K, V]) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			c.Cleanup()
		}
	}()
}
