/*
 * Copyright (c) 2024. Devtron Inc.
 */

package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
)

// ClusterOverviewConfig represents configuration for cluster overview functionality
type ClusterOverviewConfig struct {
	// CacheEnabled enables or disables caching for cluster overview data
	CacheEnabled bool `env:"CLUSTER_OVERVIEW_CACHE_ENABLED" envDefault:"true" description:"Enable caching for cluster overview data"`

	// BackgroundRefreshEnabled enables proactive background cache refresh
	BackgroundRefreshEnabled bool `env:"CLUSTER_OVERVIEW_BACKGROUND_REFRESH_ENABLED" envDefault:"true" description:"Enable background refresh of cluster overview cache"`

	// RefreshIntervalSeconds defines how often to refresh cache in background
	RefreshIntervalSeconds int `env:"CLUSTER_OVERVIEW_REFRESH_INTERVAL_SECONDS" envDefault:"15" description:"Background cache refresh interval in seconds"`

	// MaxParallelClusters limits concurrent cluster API calls during refresh
	MaxParallelClusters int `env:"CLUSTER_OVERVIEW_MAX_PARALLEL_CLUSTERS" envDefault:"15" description:"Maximum number of clusters to fetch in parallel during refresh"`

	// MaxStaleDataSeconds maximum age of cache before considering it too stale
	MaxStaleDataSeconds int `env:"CLUSTER_OVERVIEW_MAX_STALE_DATA_SECONDS" envDefault:"30" description:"Maximum age of cached data in seconds before warning"`
}

// GetRefreshInterval returns the refresh interval as a time.Duration
func (c *ClusterOverviewConfig) GetRefreshInterval() time.Duration {
	return time.Duration(c.RefreshIntervalSeconds) * time.Second
}

// GetMaxStaleDataDuration returns the max stale data duration as a time.Duration
func (c *ClusterOverviewConfig) GetMaxStaleDataDuration() time.Duration {
	return time.Duration(c.MaxStaleDataSeconds) * time.Second
}

func GetClusterOverviewConfig() (*ClusterOverviewConfig, error) {
	cfg := &ClusterOverviewConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse infra overview config: %w", err)
	}

	return cfg, nil
}
