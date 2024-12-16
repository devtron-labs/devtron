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

package util

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	util2 "github.com/devtron-labs/devtron/util"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// GetUnitSuffix loosely typed method to get the unit suffix using the unitKey type
func GetUnitSuffix(unitKey bean.ConfigKeyStr, unitStr string) units.UnitSuffix {
	switch unitKey {
	case bean.CPU_LIMIT, bean.CPU_REQUEST:
		return units.CPUUnitStr(unitStr).GetCPUUnit()
	case bean.MEMORY_LIMIT, bean.MEMORY_REQUEST:
		return units.MemoryUnitStr(unitStr).GetMemoryUnit()
	}
	return units.TimeUnitStr(unitStr).GetTimeUnit()
}

// GetUnitSuffixStr loosely typed method to get the unit suffix using the unitKey type
func GetUnitSuffixStr(unitKey bean.ConfigKey, unit units.UnitSuffix) string {
	switch unitKey {
	case bean.CPULimitKey, bean.CPURequestKey:
		return string(unit.GetCPUUnitStr())
	case bean.MemoryLimitKey, bean.MemoryRequestKey:
		return string(unit.GetMemoryUnitStr())
	}
	return string(unit.GetTimeUnitStr())
}

// GetDefaultConfigKeysMap returns a map of default config keys
func GetDefaultConfigKeysMap() map[bean.ConfigKeyStr]bool {
	return map[bean.ConfigKeyStr]bool{
		bean.CPU_LIMIT:      true,
		bean.CPU_REQUEST:    true,
		bean.MEMORY_LIMIT:   true,
		bean.MEMORY_REQUEST: true,
		bean.TIME_OUT:       true,
	}
}

func GetConfigKeyStr(configKey bean.ConfigKey) bean.ConfigKeyStr {
	switch configKey {
	case bean.CPULimitKey:
		return bean.CPU_LIMIT
	case bean.CPURequestKey:
		return bean.CPU_REQUEST
	case bean.MemoryLimitKey:
		return bean.MEMORY_LIMIT
	case bean.MemoryRequestKey:
		return bean.MEMORY_REQUEST
	case bean.TimeOutKey:
		return bean.TIME_OUT
	}
	return ""
}

func GetConfigKey(configKeyStr bean.ConfigKeyStr) bean.ConfigKey {
	switch configKeyStr {
	case bean.CPU_LIMIT:
		return bean.CPULimitKey
	case bean.CPU_REQUEST:
		return bean.CPURequestKey
	case bean.MEMORY_LIMIT:
		return bean.MemoryLimitKey
	case bean.MEMORY_REQUEST:
		return bean.MemoryRequestKey
	case bean.TIME_OUT:
		return bean.TimeOutKey
	}
	return 0
}

func GetTypedValue(configKey bean.ConfigKeyStr, value interface{}) (interface{}, error) {
	switch configKey {
	case bean.CPU_LIMIT, bean.CPU_REQUEST, bean.MEMORY_LIMIT, bean.MEMORY_REQUEST:
		//value is float64 or convertible to it
		switch v := value.(type) {
		case string:
			valueFloat, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse string to float for %s: %w", configKey, err)
			}
			return util2.TruncateFloat(valueFloat, 2), nil
		case float64:
			return util2.TruncateFloat(v, 2), nil
		default:
			return nil, fmt.Errorf("unsupported type for %s: %v", configKey, reflect.TypeOf(value))
		}
	case bean.TIME_OUT:
		switch v := value.(type) {
		case string:
			valueFloat, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse string to float for %s: %w", configKey, err)
			}
			return math.Min(math.Floor(valueFloat), math.MaxInt64), nil
		case float64:
			return math.Min(math.Floor(v), math.MaxInt64), nil
		default:
			return nil, fmt.Errorf("unsupported type for %s: %v", configKey, reflect.TypeOf(value))
		}
	// Default case
	default:
		return nil, fmt.Errorf("unsupported config key: %s", configKey)
	}
}

// todo remove this validation, as it is written additionally due to validator v9.30.0 constraint for map[string]*struct is not handled
func ValidatePayloadConfig(profileToUpdate *bean.ProfileBeanDto) error {
	if len(profileToUpdate.Name) == 0 {
		return errors.New("profile name is required")
	}
	defaultKeyMap := GetDefaultConfigKeysMap()
	for _, config := range profileToUpdate.Configurations {
		err := validateConfigItems(config, defaultKeyMap)
		if err != nil {
			return err
		}
	}
	return nil
}
func validateConfigItems(propertyConfigs []*bean.ConfigurationBean, defaultKeyMap map[bean.ConfigKeyStr]bool) error {
	var validationErrors []string
	for _, config := range propertyConfigs {
		if _, isValidKey := defaultKeyMap[config.Key]; !isValidKey {
			validationErrors = append(validationErrors, fmt.Sprintf("invalid configuration property \"%s\"", config.Key))
			continue
		}
		//_, err := GetTypedValue(config.Key, config.Value)
		//if err != nil {
		//	validationErrors = append(validationErrors, fmt.Sprintf("error in parsing value for key \"%s\": %v", config.Key, err))
		//	continue
		//}
	}
	// If any validation errors were found, return them as a single error
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}
	return nil
}

func IsValidProfileNameRequested(profileName, reqProfileName string) bool {
	return !(profileName == "" || (profileName == bean.GLOBAL_PROFILE_NAME && reqProfileName != bean.GLOBAL_PROFILE_NAME))
}

func IsValidProfileNameRequestedV0(profileName, reqProfileName string) bool {
	return !(profileName == "" || (profileName == bean.DEFAULT_PROFILE_NAME && reqProfileName != bean.DEFAULT_PROFILE_NAME))
}
