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

package bean

import (
	serviceBean "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
)

type CPUUnitStr string

const (
	CORE  CPUUnitStr = "Core"
	MILLI CPUUnitStr = "m"
)

func (cpuUnitStr CPUUnitStr) GetUnitSuffix() UnitType {
	switch cpuUnitStr {
	case CORE:
		return Core
	case MILLI:
		return Milli
	default:
		return Core
	}
}

func (cpuUnitStr CPUUnitStr) GetUnit() (serviceBean.Unit, bool) {
	cpuUnits := GetCPUUnit()
	cpuUnit, exists := cpuUnits[cpuUnitStr]
	return cpuUnit, exists
}

func (cpuUnitStr CPUUnitStr) String() string {
	return string(cpuUnitStr)
}

func GetCPUUnit() map[CPUUnitStr]serviceBean.Unit {
	return map[CPUUnitStr]serviceBean.Unit{
		MILLI: {
			Name:             string(MILLI),
			ConversionFactor: 1e-3,
		},
		CORE: {
			Name:             string(CORE),
			ConversionFactor: 1,
		},
	}
}
