package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type onDisk struct {
	sync.RWMutex
	folder string
}

// New creates a new cache for downloaded schemas
func NewOnDiskCache(cache string) Cache {
	return &onDisk{
		folder: cache,
	}
}

func cachePath(folder, resourceKind, resourceAPIVersion, k8sVersion string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s", resourceKind, resourceAPIVersion, k8sVersion)))
	return path.Join(folder, hex.EncodeToString(hash[:]))
}

// Get retrieves the JSON schema given a resource signature
func (c *onDisk) Get(resourceKind, resourceAPIVersion, k8sVersion string) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	f, err := os.Open(cachePath(c.folder, resourceKind, resourceAPIVersion, k8sVersion))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

// Set adds a JSON schema to the schema cache
func (c *onDisk) Set(resourceKind, resourceAPIVersion, k8sVersion string, schema interface{}) error {
	c.Lock()
	defer c.Unlock()
	return ioutil.WriteFile(cachePath(c.folder, resourceKind, resourceAPIVersion, k8sVersion), schema.([]byte), 0644)
}
