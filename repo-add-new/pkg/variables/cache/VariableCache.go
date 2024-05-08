package cache

import (
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"sync"
)

type VariableCacheObj struct {
	definitions []*repository.VariableDefinition
	loaded      bool
	CacheLock   *sync.Mutex
}

func (cache *VariableCacheObj) TakeLock() {
	cache.CacheLock.Lock()
}

func (cache *VariableCacheObj) ReleaseLock() {
	cache.CacheLock.Unlock()
}

func (cache *VariableCacheObj) IsLoaded() bool {
	return cache.loaded
}

func (cache *VariableCacheObj) ResetCache() {
	cache.loaded = false
}

func (cache *VariableCacheObj) SetData(defns []*repository.VariableDefinition) {
	cache.definitions = defns
	cache.loaded = true
}

func (cache *VariableCacheObj) GetData() []*repository.VariableDefinition {
	if cache.loaded {
		return cache.definitions
	}
	return nil
}
