// This package provides a simple LRU cache. It is based on the
// LRU implementation in groupcache:
// https://github.com/golang/groupcache/tree/master/lru
package lru

import (
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
)

// Cache is a thread-safe fixed size LRU cache.
type Cache struct {
	lru  *simplelru.LRU
	lock sync.RWMutex
}

// New creates an LRU of the given size
func New(size int) (*Cache, error) {
	return NewWithEvict(size, nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewWithEvict(size int, onEvicted func(key interface{}, value interface{})) (*Cache, error) {
	return NewLargeWithEvict(uint64(size), onEvicted)
}

// NewLargeWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewLargeWithEvict(size uint64, onEvicted func(key interface{}, value interface{})) (*Cache, error) {
	lru, err := simplelru.NewLRULarge(size, simplelru.EvictCallback(onEvicted))
	if err != nil {
		return nil, err
	}
	c := &Cache{
		lru: lru,
	}
	return c, nil
}

// Purge is used to completely clear the cache
func (c *Cache) Purge() {
	c.lock.Lock()
	c.lru.Purge()
	c.lock.Unlock()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *Cache) Add(key, value interface{}) bool {
	return c.AddSize(key, value, 1)
}

// AddSize adds a value to the cache.  Returns true if an eviction occurred.
func (c *Cache) AddSize(key, value interface{}, size uint64) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.AddSize(key, value, size)
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Get(key)
}

// Check if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *Cache) Contains(key interface{}) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Contains(key)
}

// Returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *Cache) Peek(key interface{}) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Peek(key)
}

// ContainsOrAdd checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *Cache) ContainsOrAdd(key, value interface{}) (ok, evict bool) {
	return c.ContainsOrAddSize(key, value, 1)
}

// ContainsOrAddSize checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *Cache) ContainsOrAddSize(key, value interface{}, size uint64) (ok, evict bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.lru.Contains(key) {
		return true, false
	} else {
		evict := c.lru.AddSize(key, value, size)
		return false, evict
	}
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key interface{}) {
	c.lock.Lock()
	c.lru.Remove(key)
	c.lock.Unlock()
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	c.lock.Lock()
	c.lru.RemoveOldest()
	c.lock.Unlock()
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *Cache) Keys() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Keys()
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Len()
}

// Size return the size of the cache.
func (c *Cache) Size() uint64 {
        c.lock.RLock()
        defer c.lock.RUnlock()
        return c.lru.Size()
}
