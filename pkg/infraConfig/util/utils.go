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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/constants"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	util2 "github.com/devtron-labs/devtron/util"
	"math"
	"reflect"
	"strconv"
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
	case constants.CPULimit, constants.CPURequest:
		return string(unit.GetCPUUnitStr())
	case constants.MemoryLimit, constants.MemoryRequest:
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
	case constants.CPULimit:
		return constants.CPU_LIMIT
	case constants.CPURequest:
		return constants.CPU_REQUEST
	case constants.MemoryLimit:
		return constants.MEMORY_LIMIT
	case constants.MemoryRequest:
		return constants.MEMORY_REQUEST
	case constants.TimeOut:
		return constants.TIME_OUT
	}
	return ""
}

func GetConfigKey(configKeyStr constants.ConfigKeyStr) constants.ConfigKey {
	switch configKeyStr {
	case constants.CPU_LIMIT:
		return constants.CPULimit
	case constants.CPU_REQUEST:
		return constants.CPURequest
	case constants.MEMORY_LIMIT:
		return constants.MemoryLimit
	case constants.MEMORY_REQUEST:
		return constants.MemoryRequest
	case constants.TIME_OUT:
		return constants.TimeOut
	}
	return 0
}

func GetTypedValue(configKey constants.ConfigKeyStr, value interface{}) interface{} {
	switch configKey {
	case constants.CPU_LIMIT, constants.CPU_REQUEST, constants.MEMORY_LIMIT, constants.MEMORY_REQUEST:
		// Assume value is float64 or convertible to it
		switch v := value.(type) {
		case string:
			valueFloat, _ := strconv.ParseFloat(v, 64)
			return util2.TruncateFloat(valueFloat, 2)
		case float64:
			return util2.TruncateFloat(v, 2)
		default:
			panic(fmt.Sprintf("Unsupported type for %s: %v", configKey, reflect.TypeOf(value)))
		}

	case constants.TIME_OUT:
		// Ensure the value is a float64 within int64 bounds
		switch v := value.(type) {
		case string:
			valueFloat, _ := strconv.ParseFloat(v, 64)
			return math.Min(math.Floor(valueFloat), math.MaxInt64)
		case float64:
			return math.Min(math.Floor(v), math.MaxInt64)
		default:
			panic(fmt.Sprintf("Unsupported type for %s: %v", configKey, reflect.TypeOf(value)))
		}
	// Default case
	default:
		// Return value as-is
		return value
	}
}
