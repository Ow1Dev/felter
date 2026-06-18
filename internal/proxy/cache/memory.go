package cache

import (
	"context"
	"sync"
	"time"
)

// entry holds a cached user ID and its expiration time.
type entry struct {
	userID int64
	expiry time.Time
}

// MemoryCache is an in-memory UserCache implementation with per-entry TTL.
type MemoryCache struct {
	mu      sync.RWMutex
	data    map[string]entry
	ttl     time.Duration
	cleaner *time.Ticker
	done    chan struct{}
}

// NewMemoryCache creates a new in-memory cache with the given TTL.
// It starts a background goroutine that cleans expired entries.
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	c := &MemoryCache{
		data:    make(map[string]entry),
		ttl:     ttl,
		cleaner: time.NewTicker(5 * time.Minute),
		done:    make(chan struct{}),
	}
	go c.cleanupLoop()
	return c
}

// Close stops the background cleanup goroutine.
func (c *MemoryCache) Close() {
	close(c.done)
	c.cleaner.Stop()
}

func (c *MemoryCache) key(provider, providerID string) string {
	return provider + ":" + providerID
}

// Get implements UserCache.
func (c *MemoryCache) Get(_ context.Context, provider, providerID string) (int64, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.data[c.key(provider, providerID)]
	if !ok || time.Now().After(e.expiry) {
		return 0, false, nil
	}
	return e.userID, true, nil
}

// Set implements UserCache.
func (c *MemoryCache) Set(_ context.Context, provider, providerID string, userID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[c.key(provider, providerID)] = entry{
		userID: userID,
		expiry: time.Now().Add(c.ttl),
	}
	return nil
}

func (c *MemoryCache) cleanupLoop() {
	for {
		select {
		case <-c.cleaner.C:
			c.purgeExpired()
		case <-c.done:
			return
		}
	}
}

func (c *MemoryCache) purgeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, v := range c.data {
		if now.After(v.expiry) {
			delete(c.data, k)
		}
	}
}
