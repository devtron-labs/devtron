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
	"github.com/devtron-labs/devtron/pkg/infraConfig/constants"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	util2 "github.com/devtron-labs/devtron/util"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// GetUnitSuffix loosely typed method to get the unit suffix using the unitKey type
func GetUnitSuffix(unitKey constants.ConfigKeyStr, unitStr string) units.UnitSuffix {
	switch unitKey {
	case constants.CPU_LIMIT, constants.CPU_REQUEST:
		return units.CPUUnitStr(unitStr).GetCPUUnit()
	case constants.MEMORY_LIMIT, constants.MEMORY_REQUEST:
		return units.MemoryUnitStr(unitStr).GetMemoryUnit()
	}
	return units.TimeUnitStr(unitStr).GetTimeUnit()
}

// GetUnitSuffixStr loosely typed method to get the unit suffix using the unitKey type
func GetUnitSuffixStr(unitKey constants.ConfigKey, unit units.UnitSuffix) string {
	switch unitKey {
	case constants.CPULimitKey, constants.CPURequestKey:
		return string(unit.GetCPUUnitStr())
	case constants.MemoryLimitKey, constants.MemoryRequestKey:
		return string(unit.GetMemoryUnitStr())
	}
	return string(unit.GetTimeUnitStr())
}

// GetDefaultConfigKeysMap returns a map of default config keys
func GetDefaultConfigKeysMap() map[constants.ConfigKeyStr]bool {
	return map[constants.ConfigKeyStr]bool{
		constants.CPU_LIMIT:      true,
		constants.CPU_REQUEST:    true,
		constants.MEMORY_LIMIT:   true,
		constants.MEMORY_REQUEST: true,
		constants.TIME_OUT:       true,
	}
}

func GetConfigKeyStr(configKey constants.ConfigKey) constants.ConfigKeyStr {
	switch configKey {
	case constants.CPULimitKey:
		return constants.CPU_LIMIT
	case constants.CPURequestKey:
		return constants.CPU_REQUEST
	case constants.MemoryLimitKey:
		return constants.MEMORY_LIMIT
	case constants.MemoryRequestKey:
		return constants.MEMORY_REQUEST
	case constants.TimeOutKey:
		return constants.TIME_OUT
	}
	return ""
}

func GetConfigKey(configKeyStr constants.ConfigKeyStr) constants.ConfigKey {
	switch configKeyStr {
	case constants.CPU_LIMIT:
		return constants.CPULimitKey
	case constants.CPU_REQUEST:
		return constants.CPURequestKey
	case constants.MEMORY_LIMIT:
		return constants.MemoryLimitKey
	case constants.MEMORY_REQUEST:
		return constants.MemoryRequestKey
	case constants.TIME_OUT:
		return constants.TimeOutKey
	}
	return 0
}

func GetTypedValue(configKey constants.ConfigKeyStr, value interface{}) (interface{}, error) {
	switch configKey {
	case constants.CPU_LIMIT, constants.CPU_REQUEST, constants.MEMORY_LIMIT, constants.MEMORY_REQUEST:
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
	case constants.TIME_OUT:
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
func validateConfigItems(propertyConfigs []*bean.ConfigurationBean, defaultKeyMap map[constants.ConfigKeyStr]bool) error {
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
