package types

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

const (
	DEFAULT_NAMESPACE = "default"
	EVENT_K8S_KIND    = "Event"
	LIST_VERB         = "list"
	Delete            = "delete"
)

type ResourceIdentifier struct {
	Name             string                  `json:"name"` // pod name for logs request
	Namespace        string                  `json:"namespace"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind"`
}

type K8sRequestBean struct {
	ResourceIdentifier ResourceIdentifier `json:"resourceIdentifier"`
	Filter             string             `json:"filter,omitempty"`
	LabelSelector      []string           `json:"labelSelector,omitempty"`
	FieldSelector      []string           `json:"fieldSelector,omitempty"`
}

type Config struct {
	ClusterCacheResyncDuration      time.Duration             `env:"CLUSTER_CACHE_RESYNC_DURATION" envDefault:"12h"`
	ClusterCacheWatchResyncDuration time.Duration             `env:"CLUSTER_CACHE_WATCH_RESYNC_DURATION" envDefault:"10m"`
	ClusterSyncRetryTimeoutDuration time.Duration             `env:"CLUSTER_SYNC_RETRY_TIMEOUT_DURATION" envDefault:"10s"`
	ClusterCacheListSemaphoreSize   int64                     `env:"CLUSTER_CACHE_LIST_SEMAPHORE_SIZE" envDefault:"5"`
	ClusterCacheListPageSize        int64                     `env:"CLUSTER_CACHE_LIST_PAGE_SIZE" envDefault:"500"`
	ClusterCacheListPageBufferSize  int32                     `env:"CLUSTER_CACHE_LIST_PAGE_BUFFER_SIZE" envDefault:"10"`
	ClusterCacheAttemptLimit        int32                     `env:"CLUSTER_CACHE_ATTEMPT_LIMIT" envDefault:"1"`
	ClusterCacheRetryUseBackoff     bool                      `env:"CLUSTER_CACHE_RETRY_USE_BACKOFF"`
	NamespacesToCache               []string                  `json:"namespacesToCache" env:"CACHED_NAMESPACES" envDefault:"gireesh-ns" envSeparator:","` // empty means all
	GVKToCacheJson                  string                    `env:"CACHED_GVKs" envDefault:"[]"`                                                         // empty means all, in case no gvk then pass anything
	GVKToCache                      []schema.GroupVersionKind `json:"gvkToCacheJson"`
}

func (config *Config) NamespaceCached(requestingNamespace string) bool {
	namespacesToCache := config.NamespacesToCache
	if len(requestingNamespace) == 0 && len(namespacesToCache) == 0 {
		return true
	}
	if len(requestingNamespace) == 0 {
		return false
	} else if len(namespacesToCache) == 0 {
		return true
	}
	for _, configuredNamespace := range namespacesToCache {
		if configuredNamespace == requestingNamespace {
			return true
		}
	}
	return false
}

func (config *Config) GVKCached(gvk schema.GroupVersionKind) bool {
	if len(config.GVKToCache) == 0 {
		return true
	}
	for _, groupVersionKind := range config.GVKToCache {
		if groupVersionKind == gvk {
			return true
		}
	}
	return false
}
