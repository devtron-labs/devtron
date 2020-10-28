package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/go-redis/cache/internal"
	"github.com/go-redis/cache/internal/lrucache"
	"github.com/go-redis/cache/internal/singleflight"

	"github.com/go-redis/redis"
)

var ErrCacheMiss = errors.New("cache: key is missing")
var errRedisLocalCacheNil = errors.New("cache: both Redis and LocalCache are nil")

func SetLogger(logger internal.Logger) {
	internal.Log = logger
}

type rediser interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
	Del(keys ...string) *redis.IntCmd
}

type Item struct {
	Ctx context.Context

	Key    string
	Object interface{}

	// Func returns object to be cached.
	Func func() (interface{}, error)

	// Expiration is the cache expiration time.
	// Default expiration is 1 hour.
	Expiration time.Duration
}

func (item *Item) object() (interface{}, error) {
	if item.Object != nil {
		return item.Object, nil
	}
	if item.Func != nil {
		return item.Func()
	}
	return nil, nil
}

func (item *Item) exp() time.Duration {
	if item.Expiration < 0 {
		return 0
	}
	if item.Expiration < time.Second {
		return time.Hour
	}
	return item.Expiration
}

type Codec struct {
	Redis rediser

	hooks      []Hook
	localCache *lrucache.Cache

	Marshal   func(interface{}) ([]byte, error)
	Unmarshal func([]byte, interface{}) error

	group singleflight.Group

	hits        uint64
	misses      uint64
	localHits   uint64
	localMisses uint64
}

// UseLocalCache causes Codec to cache items in local LRU cache.
func (cd *Codec) UseLocalCache(maxLen int, expiration time.Duration) {
	cd.localCache = lrucache.New(maxLen, expiration)
}

// Set caches the item.
func (cd *Codec) Set(item *Item) error {
	if err := cd.beforeSet(item); err != nil {
		return err
	}

	_, err := cd.setItem(item)

	if err := cd.afterSet(item); err != nil {
		return err
	}
	return err
}

func (cd *Codec) setItem(item *Item) ([]byte, error) {
	object, err := item.object()
	if err != nil {
		return nil, err
	}

	b, err := cd.Marshal(object)
	if err != nil {
		internal.Log.Printf("cache: Marshal key=%q failed: %s", item.Key, err)
		return nil, err
	}

	if cd.localCache != nil {
		cd.localCache.Set(item.Key, b)
	}

	if cd.Redis == nil {
		if cd.localCache == nil {
			return nil, errRedisLocalCacheNil
		}
		return b, nil
	}

	err = cd.Redis.Set(item.Key, b, item.exp()).Err()
	if err != nil {
		internal.Log.Printf("cache: Set key=%q failed: %s", item.Key, err)
	}
	return b, err
}

// Exists reports whether object for the given key exists.
func (cd *Codec) Exists(key string) bool {
	return cd.Get(key, nil) == nil
}

// Get gets the object for the given key.
func (cd *Codec) Get(key string, object interface{}) error {
	return cd.get(nil, key, object)
}

func (cd *Codec) GetContext(c context.Context, key string, object interface{}) error {
	return cd.get(c, key, object)
}

func (cd *Codec) get(c context.Context, key string, object interface{}) error {
	c, err := cd.beforeGet(c, key, object)
	if err != nil {
		return err
	}

	err = cd._get(key, object, false)

	if _, err := cd.afterGet(c, key, object); err != nil {
		return err
	}

	return err
}

func (cd *Codec) _get(key string, object interface{}, onlyLocalCache bool) error {
	b, err := cd.getBytes(key, onlyLocalCache)
	if err != nil {
		return err
	}

	if object == nil || len(b) == 0 {
		return nil
	}

	err = cd.Unmarshal(b, object)
	if err != nil {
		internal.Log.Printf("cache: key=%q Unmarshal(%T) failed: %s", key, object, err)
		return err
	}

	return nil
}

func (cd *Codec) getBytes(key string, onlyLocalCache bool) ([]byte, error) {
	if cd.localCache != nil {
		b, ok := cd.localCache.Get(key)
		if ok {
			atomic.AddUint64(&cd.localHits, 1)
			return b, nil
		}
		atomic.AddUint64(&cd.localMisses, 1)
	}

	if onlyLocalCache {
		return nil, ErrCacheMiss
	}
	if cd.Redis == nil {
		if cd.localCache == nil {
			return nil, errRedisLocalCacheNil
		}
		return nil, ErrCacheMiss
	}

	b, err := cd.Redis.Get(key).Bytes()
	if err != nil {
		atomic.AddUint64(&cd.misses, 1)
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		internal.Log.Printf("cache: Get key=%q failed: %s", key, err)
		return nil, err
	}
	atomic.AddUint64(&cd.hits, 1)

	if cd.localCache != nil {
		cd.localCache.Set(key, b)
	}
	return b, nil
}

// Once gets the item.Object for the given item.Key from the cache or
// executes, caches, and returns the results of the given item.Func,
// making sure that only one execution is in-flight for a given item.Key
// at a time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (cd *Codec) Once(item *Item) error {
	if err := cd.beforeOnce(item); err != nil {
		return err
	}

	err := cd.once(item)

	if err := cd.afterOnce(item); err != nil {
		return err
	}

	return err
}

func (cd *Codec) once(item *Item) error {
	b, cached, err := cd.getSetItemBytesOnce(item)
	if err != nil {
		return err
	}

	if item.Object == nil || len(b) == 0 {
		return nil
	}

	err = cd.Unmarshal(b, item.Object)
	if err != nil {
		internal.Log.Printf("cache: key=%q Unmarshal(%T) failed: %s", item.Key, item.Object, err)
		if cached {
			_ = cd._delete(item.Key)
			return cd.once(item)
		}
		return err
	}

	return nil
}

func (cd *Codec) getSetItemBytesOnce(item *Item) (b []byte, cached bool, err error) {
	if cd.localCache != nil {
		b, err := cd.getItemBytesFast(item)
		if err == nil {
			return b, true, nil
		}
	}

	obj, err := cd.group.Do(item.Key, func() (interface{}, error) {
		b, err := cd.getItemBytes(item)
		if err == nil {
			cached = true
			return b, nil
		}

		obj, err := item.Func()
		if err != nil {
			return nil, err
		}

		b, err = cd.setItem(&Item{
			Key:        item.Key,
			Object:     obj,
			Expiration: item.Expiration,
		})
		if b != nil {
			// Ignore error if we have the result.
			return b, nil
		}
		return nil, err
	})
	if err != nil {
		return nil, false, err
	}
	return obj.([]byte), cached, nil
}

func (cd *Codec) getItemBytes(item *Item) ([]byte, error) {
	return cd.getBytes(item.Key, false)
}

func (cd *Codec) getItemBytesFast(item *Item) ([]byte, error) {
	return cd.getBytes(item.Key, true)
}

func (cd *Codec) Delete(key string) error {
	return cd.delete(nil, key)
}

func (cd *Codec) DeleteContext(c context.Context, key string) error {
	return cd.delete(c, key)
}

func (cd *Codec) delete(c context.Context, key string) error {
	c, err := cd.beforeDelete(c, key)
	if err != nil {
		return err
	}

	err = cd._delete(key)

	if _, err := cd.afterDelete(c, key); err != nil {
		return err
	}

	return err
}

func (cd *Codec) _delete(key string) error {
	if cd.localCache != nil {
		cd.localCache.Delete(key)
	}

	if cd.Redis == nil {
		if cd.localCache == nil {
			return errRedisLocalCacheNil
		}
		return nil
	}

	deleted, err := cd.Redis.Del(key).Result()
	if err != nil {
		internal.Log.Printf("cache: Del key=%q failed: %s", key, err)
		return err
	}
	if deleted == 0 {
		return ErrCacheMiss
	}
	return nil
}

//------------------------------------------------------------------------------

type Hook interface {
	BeforeSet(item *Item) error
	AfterSet(item *Item) error

	BeforeGet(c context.Context, key string, object interface{}) (context.Context, error)
	AfterGet(c context.Context, key string, object interface{}) (context.Context, error)

	BeforeDelete(c context.Context, key string) (context.Context, error)
	AfterDelete(c context.Context, key string) (context.Context, error)

	BeforeOnce(item *Item) error
	AfterOnce(item *Item) error
}

func (cd *Codec) AddHook(h Hook) {
	cd.hooks = append(cd.hooks, h)
}

func (cd *Codec) beforeSet(item *Item) error {
	for _, h := range cd.hooks {
		err := h.BeforeSet(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cd *Codec) afterSet(item *Item) error {
	for _, h := range cd.hooks {
		err := h.AfterSet(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cd *Codec) beforeGet(c context.Context, key string, object interface{}) (context.Context, error) {
	for _, h := range cd.hooks {
		var err error
		c, err = h.BeforeGet(c, key, object)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (cd *Codec) afterGet(c context.Context, key string, object interface{}) (context.Context, error) {
	for _, h := range cd.hooks {
		var err error
		c, err = h.AfterGet(c, key, object)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (cd *Codec) beforeDelete(c context.Context, key string) (context.Context, error) {
	for _, h := range cd.hooks {
		var err error
		c, err = h.BeforeDelete(c, key)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (cd *Codec) afterDelete(c context.Context, key string) (context.Context, error) {
	for _, h := range cd.hooks {
		var err error
		c, err = h.AfterDelete(c, key)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (cd *Codec) beforeOnce(item *Item) error {
	for _, h := range cd.hooks {
		err := h.BeforeOnce(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cd *Codec) afterOnce(item *Item) error {
	for _, h := range cd.hooks {
		err := h.AfterOnce(item)
		if err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type Stats struct {
	Hits        uint64
	Misses      uint64
	LocalHits   uint64
	LocalMisses uint64
}

// Stats returns cache statistics.
func (cd *Codec) Stats() *Stats {
	stats := Stats{
		Hits:   atomic.LoadUint64(&cd.hits),
		Misses: atomic.LoadUint64(&cd.misses),
	}
	if cd.localCache != nil {
		stats.LocalHits = atomic.LoadUint64(&cd.localHits)
		stats.LocalMisses = atomic.LoadUint64(&cd.localMisses)
	}
	return &stats
}
