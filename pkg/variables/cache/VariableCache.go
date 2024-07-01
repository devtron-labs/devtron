/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
