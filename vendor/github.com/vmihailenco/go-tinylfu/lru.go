package tinylfu

import "container/list"

// Cache is an LRU cache.  It is not safe for concurrent access.
type lruCache struct {
	data map[string]*list.Element
	cap  int
	ll   *list.List
}

func newLRU(cap int, data map[string]*list.Element) *lruCache {
	return &lruCache{
		data: data,
		cap:  cap,
		ll:   list.New(),
	}
}

// Get returns a value from the cache
func (lru *lruCache) get(v *list.Element) {
	lru.ll.MoveToFront(v)
}

// Set sets a value in the cache
func (lru *lruCache) add(newItem *Item) (_ *Item, evicted bool) {
	if lru.ll.Len() < lru.cap {
		lru.data[newItem.Key] = lru.ll.PushFront(newItem)
		return nil, false
	}

	// reuse the tail item
	val := lru.ll.Back()
	item := val.Value.(*Item)

	delete(lru.data, item.Key)

	oldItem := *item
	*item = *newItem

	lru.data[item.Key] = val
	lru.ll.MoveToFront(val)

	return &oldItem, true
}

// Len returns the total number of items in the cache
func (lru *lruCache) Len() int {
	return len(lru.data)
}

// Remove removes an item from the cache, returning the item and a boolean indicating if it was found
func (lru *lruCache) Remove(v *list.Element) {
	lru.ll.Remove(v)
}
