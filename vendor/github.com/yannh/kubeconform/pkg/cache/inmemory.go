package cache

import (
	"fmt"
	"sync"
)

// SchemaCache is a cache for downloaded schemas, so each file is only retrieved once
// It is different from pkg/registry/http_cache.go in that:
//   - This cache caches the parsed Schemas
type inMemory struct {
	sync.RWMutex
	schemas map[string]interface{}
}

// New creates a new cache for downloaded schemas
func NewInMemoryCache() Cache {
	return &inMemory{
		schemas: map[string]interface{}{},
	}
}

func key(resourceKind, resourceAPIVersion, k8sVersion string) string {
	return fmt.Sprintf("%s-%s-%s", resourceKind, resourceAPIVersion, k8sVersion)
}

// Get retrieves the JSON schema given a resource signature
func (c *inMemory) Get(resourceKind, resourceAPIVersion, k8sVersion string) (interface{}, error) {
	k := key(resourceKind, resourceAPIVersion, k8sVersion)
	c.RLock()
	defer c.RUnlock()
	schema, ok := c.schemas[k]

	if !ok {
		return nil, fmt.Errorf("schema not found in in-memory cache")
	}

	return schema, nil
}

// Set adds a JSON schema to the schema cache
func (c *inMemory) Set(resourceKind, resourceAPIVersion, k8sVersion string, schema interface{}) error {
	k := key(resourceKind, resourceAPIVersion, k8sVersion)
	c.Lock()
	defer c.Unlock()
	c.schemas[k] = schema

	return nil
}
