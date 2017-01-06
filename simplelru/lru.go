package simplelru

import (
	"container/list"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// LRU implements a non-thread safe fixed size LRU cache
type LRU struct {
	size, total uint64
	evictList   *list.List
	items       map[interface{}]*list.Element
	onEvict     EvictCallback
}

// entry is used to hold a value in the evictList
type entry struct {
	key   interface{}
	value interface{}
	size  uint64
}

// NewLRU constructs a count constrained LRU of the given size
func NewLRU(size int, onEvict EvictCallback) (*LRU, error) {
	return NewLRULarge(uint64(size), onEvict)
}

// NewLRUofType constructs an LRU of the given type and size
func NewLRULarge(size uint64, onEvict EvictCallback) (*LRU, error) {
	c := &LRU{
		size:      size,
		total:     0,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache
func (c *LRU) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry).value)
		}
		delete(c.items, k)
	}
	c.total = 0
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRU) Add(key, value interface{}) bool {
	return c.AddSize(key, value, 1)
}

// AddSize adds a value to the cache with an associated size.  Returns true if an eviction occurred.
func (c *LRU) AddSize(key, value interface{}, size uint64) bool {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		return false
	}

	// Add new item
	ent := &entry{key, value, size}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry
	c.total += size

	evict := c.total > c.size
	// Verify size not exceeded
	for c.total > c.size {
		c.removeOldest()
	}
	return evict
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(key interface{}) (value interface{}, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return
}

// Check if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(key interface{}) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU) Peek(key interface{}) (value interface{}, ok bool) {
	if ent, ok := c.items[key]; ok {
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) Remove(key interface{}) bool {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		c.total -= ent.Value.(*entry).size
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU) RemoveOldest() (interface{}, interface{}, bool) {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// GetOldest returns the oldest entry
func (c *LRU) GetOldest() (interface{}, interface{}, bool) {
	ent := c.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	return c.evictList.Len()
}

// Size returns the total bytes of all items in the cache.
func (c *LRU) Size() uint64 {
	return c.total
}

// removeOldest removes the oldest item from the cache.
func (c *LRU) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *LRU) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	c.total -= kv.size
	delete(c.items, kv.key)
	if c.onEvict != nil {
		c.onEvict(kv.key, kv.value)
	}
}
