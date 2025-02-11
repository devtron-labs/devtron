// Copyright (c) 2014-present José Carlos Nieto, https://menteslibres.net/xiam
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cache

import (
	"container/list"
	"errors"
	"sync"
)

const defaultCapacity = 128

// Cache holds a map of volatile key -> values.
type Cache struct {
	keys     *list.List
	items    map[uint64]*list.Element
	mu       sync.RWMutex
	capacity int
}

type cacheItem struct {
	key   uint64
	value interface{}
}

// NewCacheWithCapacity initializes a new caching space with the given
// capacity.
func NewCacheWithCapacity(capacity int) (*Cache, error) {
	if capacity < 1 {
		return nil, errors.New("Capacity must be greater than zero.")
	}
	c := &Cache{
		capacity: capacity,
	}
	c.init()
	return c, nil
}

// NewCache initializes a new caching space with default settings.
func NewCache() *Cache {
	c, err := NewCacheWithCapacity(defaultCapacity)
	if err != nil {
		panic(err.Error()) // Should never happen as we're not providing a negative defaultCapacity.
	}
	return c
}

func (c *Cache) init() {
	c.items = make(map[uint64]*list.Element)
	c.keys = list.New()
}

// Read attempts to retrieve a cached value as a string, if the value does not
// exists returns an empty string and false.
func (c *Cache) Read(h Hashable) (string, bool) {
	if v, ok := c.ReadRaw(h); ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// ReadRaw attempts to retrieve a cached value as an interface{}, if the value
// does not exists returns nil and false.
func (c *Cache) ReadRaw(h Hashable) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[h.Hash()]
	if ok {
		return item.Value.(*cacheItem).value, true
	}

	return nil, false
}

// Write stores a value in memory. If the value already exists its overwritten.
func (c *Cache) Write(h Hashable, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := h.Hash()

	if item, ok := c.items[key]; ok {
		item.Value.(*cacheItem).value = value
		c.keys.MoveToFront(item)
		return
	}

	c.items[key] = c.keys.PushFront(&cacheItem{key, value})

	for c.keys.Len() > c.capacity {
		item := c.keys.Remove(c.keys.Back()).(*cacheItem)
		delete(c.items, item.key)
		if p, ok := item.value.(HasOnEvict); ok {
			p.OnEvict()
		}
	}
}

// Clear generates a new memory space, leaving the old memory unreferenced, so
// it can be claimed by the garbage collector.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, item := range c.items {
		if p, ok := item.Value.(*cacheItem).value.(HasOnEvict); ok {
			p.OnEvict()
		}
	}

	c.init()
}
