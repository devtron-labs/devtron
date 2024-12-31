// Package tinylfu is an implementation of the TinyLFU caching algorithm
/*
   http://arxiv.org/abs/1512.00727
*/
package tinylfu

import (
	"container/list"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
)

type Item struct {
	Key      string
	Value    interface{}
	ExpireAt time.Time
	OnEvict  func()

	listid int
	keyh   uint64
}

func (item Item) expired() bool {
	return !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)
}

type T struct {
	w       int
	samples int

	countSketch *cm4
	bouncer     *doorkeeper

	data map[string]*list.Element

	lru  *lruCache
	slru *slruCache
}

func New(size int, samples int) *T {
	const lruPct = 1

	lruSize := (lruPct * size) / 100
	if lruSize < 1 {
		lruSize = 1
	}
	slruSize := int(float64(size) * ((100.0 - lruPct) / 100.0))
	if slruSize < 1 {
		slruSize = 1
	}
	slru20 := int(0.2 * float64(slruSize))
	if slru20 < 1 {
		slru20 = 1
	}

	data := make(map[string]*list.Element, size)

	return &T{
		w:       0,
		samples: samples,

		countSketch: newCM4(size),
		bouncer:     newDoorkeeper(samples, 0.01),

		data: data,

		lru:  newLRU(lruSize, data),
		slru: newSLRU(slru20, slruSize-slru20, data),
	}
}

func (t *T) onEvict(item *Item) {
	if item.OnEvict != nil {
		item.OnEvict()
	}
}

func (t *T) Get(key string) (interface{}, bool) {
	t.w++
	if t.w == t.samples {
		t.countSketch.reset()
		t.bouncer.reset()
		t.w = 0
	}

	keyh := xxhash.Sum64String(key)
	t.countSketch.add(keyh)

	val, ok := t.data[key]
	if !ok {
		return nil, false
	}

	item := val.Value.(*Item)
	if item.expired() {
		t.del(val)
		return nil, false
	}

	// Save the value since it is overwritten below.
	value := item.Value

	if item.listid == 0 {
		t.lru.get(val)
	} else {
		t.slru.get(val)
	}

	return value, true
}

func (t *T) Set(newItem *Item) {
	newItem.keyh = xxhash.Sum64String(newItem.Key)

	oldItem, evicted := t.lru.add(newItem)
	if !evicted {
		return
	}

	// estimate count of what will be evicted from slru
	victim := t.slru.victim()
	if victim == nil {
		t.slru.add(oldItem)
		return
	}

	if !t.bouncer.allow(oldItem.keyh) {
		t.onEvict(oldItem)
		return
	}

	victimCount := t.countSketch.estimate(victim.keyh)
	itemCount := t.countSketch.estimate(oldItem.keyh)

	if itemCount > victimCount {
		t.slru.add(oldItem)
	} else {
		t.onEvict(oldItem)
	}
}

func (t *T) Del(key string) {
	if val, ok := t.data[key]; ok {
		t.del(val)
	}
}

func (t *T) del(val *list.Element) {
	item := val.Value.(*Item)
	delete(t.data, item.Key)

	if item.listid == 0 {
		t.lru.Remove(val)
	} else {
		t.slru.Remove(val)
	}

	t.onEvict(item)
}

//------------------------------------------------------------------------------

type SyncT struct {
	mu sync.Mutex
	t  *T
}

func NewSync(size int, samples int) *SyncT {
	return &SyncT{
		t: New(size, samples),
	}
}

func (t *SyncT) Get(key string) (interface{}, bool) {
	t.mu.Lock()
	val, ok := t.t.Get(key)
	t.mu.Unlock()
	return val, ok
}

func (t *SyncT) Set(item *Item) {
	t.mu.Lock()
	t.t.Set(item)
	t.mu.Unlock()
}

func (t *SyncT) Del(key string) {
	t.mu.Lock()
	t.t.Del(key)
	t.mu.Unlock()
}
