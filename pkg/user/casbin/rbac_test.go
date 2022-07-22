package casbin

import (
	"github.com/patrickmn/go-cache"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEnforcerCache(t *testing.T) {

	t.SkipNow()

	lock := make(map[string]*CacheData)
	cache123 := cache.New(-1, 5*time.Minute)

	t.Run("requesterAndWriter", func(t *testing.T) {
		for i := 0; i < 100_000; i++ {
			emailId := GetRandomStringOfGivenLength(rand.Intn(1000)) + "@yopmail.com"
			getAndSet(lock, emailId, cache123)
		}
	})
	t.Run("CacheInvalidate", func(t *testing.T) {
		invalidateCache_123(lock, cache123)
	})

	t.Run("cache123-maintainer", func(t *testing.T) {
		//for true {
		//	fmt.Println("hello-world")
		//}
	})
}

func GetRandomStringOfGivenLength(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func invalidateCache_123(lock map[string]*CacheData, cache *cache.Cache) {
	for emailId := range lock {
		cache.Delete(emailId)
		cacheLock123 := getEnforcerCacheLock_123(lock, emailId)
		cacheLock123.lock.Lock()
		cacheLock123.cacheCleaningFlag = true
		cacheLock123.lock.Unlock()
	}
}

func getAndSet(lock map[string]*CacheData, emailId string, cache *cache.Cache) bool {
	cacheLock := getEnforcerCacheLock_123(lock, emailId)
	cacheLock.lock.RLock()
	atomic.AddInt64(&cacheLock.enforceReqCounter, 1)
	_, found := cache.Get(emailId)
	cacheLock.lock.RUnlock()
	if found {
		// do nothing
		cacheLock.lock.Lock()
		defer cacheLock.lock.Unlock()
		returnVal := atomic.AddInt64(&cacheLock.enforceReqCounter, -1)
		if cacheLock.cacheCleaningFlag {
			if returnVal == 0 {
				cacheLock.cacheCleaningFlag = false
			}
		}
		return true
	}

	resultVal := enforce(emailId)
	cacheLock.lock.Lock()
	if !cacheLock.cacheCleaningFlag {
		cache.Set(emailId, resultVal, -1)
	}
	returnVal := atomic.AddInt64(&cacheLock.enforceReqCounter, -1)
	if cacheLock.cacheCleaningFlag {
		if returnVal == 0 {
			cacheLock.cacheCleaningFlag = false
		}
	}
	cacheLock.lock.Unlock()
	return false
}

func getEnforcerCacheLock_123(lock map[string]*CacheData, emailId string) *CacheData {
	enforcerCacheMutex, found := lock[getLockKey(emailId)]
	if !found {
		enforcerCacheMutex =
			&CacheData{
				lock:              &sync.RWMutex{},
				enforceReqCounter: int64(0),
				cacheCleaningFlag: false,
			}
		lock[getLockKey(emailId)] = enforcerCacheMutex
	}
	return enforcerCacheMutex
}

func enforce(randomeKey string) bool {
	return len(randomeKey)%2 == 0
}
