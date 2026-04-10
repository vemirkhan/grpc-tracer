// Package cache provides a simple TTL-based in-memory cache for trace span lookups.
package cache

import (
	"sync"
	"time"
)

// entry holds a cached value along with its expiration time.
type entry struct {
	value     interface{}
	expiresAt time.Time
}

// Cache is a thread-safe TTL cache.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]entry
	ttl     time.Duration
	stopGC  chan struct{}
}

// New creates a new Cache with the given TTL and a background GC interval.
// Pass gcInterval == 0 to disable automatic eviction.
func New(ttl, gcInterval time.Duration) *Cache {
	c := &Cache{
		items:  make(map[string]entry),
		ttl:    ttl,
		stopGC: make(chan struct{}),
	}
	if gcInterval > 0 {
		go c.runGC(gcInterval)
	}
	return c
}

// Set stores a value under key, overwriting any existing entry.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Get retrieves a value by key. Returns (value, true) if found and not expired,
// or (nil, false) otherwise.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len returns the number of items currently in the cache (including expired).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stop halts the background GC goroutine.
func (c *Cache) Stop() {
	close(c.stopGC)
}

// runGC periodically removes expired entries.
func (c *Cache) runGC(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.evict()
		case <-c.stopGC:
			return
		}
	}
}

// evict removes all expired entries.
func (c *Cache) evict() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
}
