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

import serviceBean "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"

type NoUnitStr string

const (
	NoUnit NoUnitStr = ""
)

// GetUnitSuffix returns the UnitSuffix for NoUnit (just return 20 as it represents no unit)
func (noUnitStr NoUnitStr) GetUnitSuffix() UnitType {
	switch noUnitStr {
	case NoUnit:
		return NoSuffix
	default:
		return NoSuffix
	}
}

func (noUnitStr NoUnitStr) GetUnit() (serviceBean.Unit, bool) {
	noUnits := GetNoUnit()
	noUnit, exists := noUnits[noUnitStr]
	return noUnit, exists
}

func (noUnitStr NoUnitStr) String() string {
	return string(noUnitStr)
}

func GetNoUnit() map[NoUnitStr]serviceBean.Unit {
	return map[NoUnitStr]serviceBean.Unit{
		NoUnit: {
			Name:             "NoUnit",
			ConversionFactor: 0,
		},
	}
}
