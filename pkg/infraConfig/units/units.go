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

package units

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	unitsBean "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
)

type UnitService[T any] interface {
	GetAllUnits() map[string]v1.Unit
	GetDefaultUnitSuffix() string
	ParseValAndUnit(val T, unitType unitsBean.UnitType) (*unitsBean.ParsedValue, error)
	Validate(configuration *v1.GenericConfigurationBean[T]) (*unitsBean.ConfigValue[T], error)
}

type UnitStrService interface {
	GetUnitSuffix() unitsBean.UnitType
	GetUnit() (v1.Unit, bool)
	String() string
	unitsBean.CPUUnitStr | unitsBean.MemoryUnitStr | unitsBean.TimeUnitStr | unitsBean.NoUnitStr
}

func parseJsonValueToString[T any](customValue T) (string, error) {
	jsonValue, err := json.Marshal(customValue)
	if err != nil {
		return "", fmt.Errorf("failed to marshal: %w", err)
	}
	return string(jsonValue), nil
}
