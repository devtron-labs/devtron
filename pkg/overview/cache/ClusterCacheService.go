/*
 * Copyright (c) 2024. Devtron Inc.
 */

package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"go.uber.org/zap"
)

// ClusterCacheService provides caching functionality for cluster overview data
type ClusterCacheService interface {
	GetClusterOverview() (*bean.ClusterOverviewResponse, bool)
	SetClusterOverview(data *bean.ClusterOverviewResponse) error
	InvalidateClusterOverview()
	InvalidateAll()
	IsRefreshing() bool
	SetRefreshing(refreshing bool)
	GetCacheAge() time.Duration
}

// ClusterCacheServiceImpl implements ClusterCacheService using in-memory cache
type ClusterCacheServiceImpl struct {
	logger        *zap.SugaredLogger
	overviewCache *cacheEntry
	cacheMutex    sync.RWMutex
}

// cacheEntry represents a cached item with timestamp
type cacheEntry struct {
	data         interface{}
	lastUpdated  time.Time
	isRefreshing bool
}

// NewClusterCacheServiceImpl creates a new instance of ClusterCacheServiceImpl
func NewClusterCacheServiceImpl(logger *zap.SugaredLogger) *ClusterCacheServiceImpl {
	return &ClusterCacheServiceImpl{
		logger: logger,
	}
}

// GetClusterOverview retrieves cluster overview data from cache
func (impl *ClusterCacheServiceImpl) GetClusterOverview() (*bean.ClusterOverviewResponse, bool) {
	impl.cacheMutex.RLock()
	defer impl.cacheMutex.RUnlock()

	if impl.overviewCache == nil {
		return nil, false
	}

	if data, ok := impl.overviewCache.data.(*bean.ClusterOverviewResponse); ok {
		age := time.Since(impl.overviewCache.lastUpdated)
		impl.logger.Infow("cluster overview cache hit", "cacheAge", age)
		return data, true
	}

	impl.logger.Errorw("cluster overview cache data type mismatch")
	return nil, false
}

// SetClusterOverview stores cluster overview data in cache with timestamp
func (impl *ClusterCacheServiceImpl) SetClusterOverview(data *bean.ClusterOverviewResponse) error {
	if data == nil {
		return fmt.Errorf("cannot cache nil cluster overview data")
	}

	impl.cacheMutex.Lock()
	defer impl.cacheMutex.Unlock()

	impl.overviewCache = &cacheEntry{
		data:        data,
		lastUpdated: time.Now(),
	}

	impl.logger.Debugw("cluster overview data cached", "timestamp", impl.overviewCache.lastUpdated)
	return nil
}

// InvalidateClusterOverview removes cluster overview data from cache
func (impl *ClusterCacheServiceImpl) InvalidateClusterOverview() {
	impl.cacheMutex.Lock()
	defer impl.cacheMutex.Unlock()

	impl.overviewCache = nil
	impl.logger.Debugw("cluster overview cache invalidated")
}

// InvalidateAll removes all cached data
func (impl *ClusterCacheServiceImpl) InvalidateAll() {
	impl.cacheMutex.Lock()
	defer impl.cacheMutex.Unlock()

	impl.overviewCache = nil
	impl.logger.Debugw("all cluster cache invalidated")
}

// IsRefreshing checks if cache is currently being refreshed
func (impl *ClusterCacheServiceImpl) IsRefreshing() bool {
	impl.cacheMutex.RLock()
	defer impl.cacheMutex.RUnlock()

	if impl.overviewCache == nil {
		return false
	}
	return impl.overviewCache.isRefreshing
}

// SetRefreshing marks cache as being refreshed
func (impl *ClusterCacheServiceImpl) SetRefreshing(refreshing bool) {
	impl.cacheMutex.Lock()
	defer impl.cacheMutex.Unlock()

	if impl.overviewCache != nil {
		impl.overviewCache.isRefreshing = refreshing
	} else if refreshing {
		// Initialize cache entry if setting refreshing to true
		impl.overviewCache = &cacheEntry{
			isRefreshing: true,
			lastUpdated:  time.Now(),
		}
	}
}

// GetCacheAge returns how old the cached data is
func (impl *ClusterCacheServiceImpl) GetCacheAge() time.Duration {
	impl.cacheMutex.RLock()
	defer impl.cacheMutex.RUnlock()

	if impl.overviewCache == nil {
		return 0
	}
	return time.Since(impl.overviewCache.lastUpdated)
}
