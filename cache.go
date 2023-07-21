package traefik_geoip //nolint:revive,stylecheck

import (
	"sync"

	"github.com/dgryski/go-tinylfu" //nolint:depguard
)

// Cache is a cache.
type Cache struct {
	mu  sync.Mutex
	lfu *tinylfu.T[GeoIPResult]
}

// NewCache creates a new cache.
func NewCache(size int) Cache {
	return Cache{
		lfu: tinylfu.New[GeoIPResult](size, size*10),
	}
}

// Add adds a new key/value pair to the cache.
func (c *Cache) Add(key string, value GeoIPResult) {
	c.mu.Lock()
	c.lfu.Add(key, value)
	c.mu.Unlock()
}

// Get gets a value from the cache.
func (c *Cache) Get(key string) (GeoIPResult, bool) {
	c.mu.Lock()
	value, ok := c.lfu.Get(key)
	c.mu.Unlock()
	return value, ok
}
