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

type TimeUnitStr string

const (
	SecondStr TimeUnitStr = "Seconds"
	MinuteStr TimeUnitStr = "Minutes"
	HourStr   TimeUnitStr = "Hours"
)

func (timeUnitStr TimeUnitStr) GetUnitSuffix() UnitType {
	switch timeUnitStr {
	case SecondStr:
		return Second
	case MinuteStr:
		return Minute
	case HourStr:
		return Hour
	default:
		return Second
	}
}

func (timeUnitStr TimeUnitStr) GetUnit() (serviceBean.Unit, bool) {
	timeUnits := GetTimeUnit()
	timeUnit, exists := timeUnits[timeUnitStr]
	return timeUnit, exists
}

func (timeUnitStr TimeUnitStr) String() string {
	return string(timeUnitStr)
}

func GetTimeUnit() map[TimeUnitStr]serviceBean.Unit {
	return map[TimeUnitStr]serviceBean.Unit{
		SecondStr: {
			Name:             string(SecondStr),
			ConversionFactor: 1,
		},
		MinuteStr: {
			Name:             string(MinuteStr),
			ConversionFactor: 60,
		},
		HourStr: {
			Name:             string(HourStr),
			ConversionFactor: 3600,
		},
	}
}
