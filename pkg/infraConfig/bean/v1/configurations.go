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

// Package v1 implements the infra config with interface values.
package v1

type ConfigurationBean struct {
	ConfigurationBeanAbstract
	Value            any             `json:"value,omitempty"`
	Count            int             `json:"count,omitempty"`
	ConfigState      ConfigStateType `json:"configState,omitempty"`
	AppliedConfigIds []int           `json:"-"`
}

func (c *ConfigurationBean) DeepCopy() *ConfigurationBean {
	if c == nil {
		return nil
	}
	config := *c
	return &config
}

func (c *ConfigurationBean) IsEmpty() bool {
	return c == nil
}

func (c *ConfigurationBean) GetStringValue() string {
	return c.Value.(string)
}

// ConfigStateType represents the derived state of the ConfigurationBean.Value
type ConfigStateType string

const (
	// OVERRIDDEN is used when the configuration is overridden in the profile
	OVERRIDDEN ConfigStateType = "OVERRIDDEN"
	// PARTIALLY_INHERITING is used when the configuration is partially inherited from the global profile
	PARTIALLY_INHERITING ConfigStateType = "PARTIALLY_INHERITING"
	// INHERITING_GLOBAL_PROFILE is used when the configuration is inherited from the global profile
	INHERITING_GLOBAL_PROFILE ConfigStateType = "INHERITING_GLOBAL_PROFILE"
)

// GenericConfigurationBean is for internal use only
//   - used for specific handling of configurations and to avoid type assertion
//   - not exposed to the API
//   - derived from ConfigurationBean
type GenericConfigurationBean[T any] struct {
	ConfigurationBeanAbstract
	Value T
}

type ConfigurationBeanAbstract struct {
	Id          int          `json:"id"`
	Key         ConfigKeyStr `json:"key" validate:"required,oneof=cpu_limit cpu_request memory_limit memory_request timeout node_selector tolerations cm cs"`
	Unit        string       `json:"unit"`
	ProfileName string       `json:"profileName,omitempty"`
	ProfileId   int          `json:"profileId,omitempty"`
	Active      bool         `json:"active"`
}

// whenever new constant gets added here,
// we need to add it in util.GetConfigKeysMapForPlatform method as well

// ConfigKey represents the configuration key in the DB model
type ConfigKey int

const (
	CPULimitKey      ConfigKey = 1
	CPURequestKey    ConfigKey = 2
	MemoryLimitKey   ConfigKey = 3
	MemoryRequestKey ConfigKey = 4
	TimeOutKey       ConfigKey = 5

	// enterprise-only keys; kept together to maintain order

	NodeSelectorKey ConfigKey = 6
	TolerationsKey  ConfigKey = 7
	ConfigMapKey    ConfigKey = 8
	SecretKey       ConfigKey = 9
)

// ConfigKeyStr represents the configuration key in the API
type ConfigKeyStr string

const (
	CPU_LIMIT      ConfigKeyStr = "cpu_limit"
	CPU_REQUEST    ConfigKeyStr = "cpu_request"
	MEMORY_LIMIT   ConfigKeyStr = "memory_limit"
	MEMORY_REQUEST ConfigKeyStr = "memory_request"
	TIME_OUT       ConfigKeyStr = "timeout"

	// enterprise-only keys; kept together to maintain order

	NODE_SELECTOR ConfigKeyStr = "node_selector"
	TOLERATIONS   ConfigKeyStr = "tolerations"
	CONFIG_MAP    ConfigKeyStr = "cm"
	SECRET        ConfigKeyStr = "cs"
)

// AllConfigKeysV0 contains the list of supported configuration keys in V0
var AllConfigKeysV0 = []ConfigKeyStr{CPU_LIMIT, CPU_REQUEST, MEMORY_LIMIT, MEMORY_REQUEST, TIME_OUT}

// AllConfigKeysV1 contains the list of supported configuration keys in V1
var AllConfigKeysV1 = append(AllConfigKeysV0, []ConfigKeyStr{NODE_SELECTOR, TOLERATIONS, CONFIG_MAP, SECRET}...)

type InfraConfigKeys map[ConfigKeyStr]bool

func (s InfraConfigKeys) IsSupported(key ConfigKeyStr) bool {
	if s == nil {
		return false
	}
	_, ok := s[key]
	return ok
}

func (s InfraConfigKeys) IsConfigured(key ConfigKeyStr) bool {
	return s.IsSupported(key) && !s[key]
}

func (s InfraConfigKeys) GetAllSupportedKeys() []ConfigKeyStr {
	keys := make([]ConfigKeyStr, 0)
	for key := range s {
		if s.IsSupported(key) {
			keys = append(keys, key)
		}
	}
	return keys
}

func (s InfraConfigKeys) GetUnConfiguredKeys() []ConfigKeyStr {
	keys := make([]ConfigKeyStr, 0)
	for key := range s {
		if !s.IsSupported(key) {
			continue
		}
		if !s.IsConfigured(key) {
			keys = append(keys, key)
		}
	}
	return keys
}

func (s InfraConfigKeys) MarkUnConfigured(key ConfigKeyStr) InfraConfigKeys {
	if s.IsSupported(key) {
		s[key] = true
	}
	return s
}

func (s InfraConfigKeys) MarkConfigured(key ConfigKeyStr) InfraConfigKeys {
	if s.IsSupported(key) {
		s[key] = false
	}
	return s
}
