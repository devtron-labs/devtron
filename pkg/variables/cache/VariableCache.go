package cache

import "github.com/devtron-labs/devtron/pkg/variables/repository"

type VariableCacheObj struct {
	definitions []*repository.VariableDefinition
	loaded      bool
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
