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
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/errors"
	unitsBean "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
	globalUtil "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type MemoryUnitFactory struct {
	logger      *zap.SugaredLogger
	memoryUnits map[unitsBean.MemoryUnitStr]v1.Unit
}

func NewMemoryUnitFactory(logger *zap.SugaredLogger) *MemoryUnitFactory {
	return &MemoryUnitFactory{
		logger:      logger,
		memoryUnits: unitsBean.GetMemoryUnit(),
	}
}

func (m *MemoryUnitFactory) GetAllUnits() map[string]v1.Unit {
	memoryUnits := m.memoryUnits
	units := make(map[string]v1.Unit)
	for key, value := range memoryUnits {
		units[string(key)] = value
	}
	return units
}

func (m *MemoryUnitFactory) GetDefaultUnitSuffix() string {
	var defaultUnit unitsBean.UnitType
	return defaultUnit.GetMemoryUnitStr().String()
}

func (m *MemoryUnitFactory) ParseValAndUnit(val float64, unitType unitsBean.UnitType) (*unitsBean.ParsedValue, error) {
	return unitsBean.NewParsedValue().
		WithValueString(strconv.FormatFloat(val, 'f', -1, 64)).
		WithUnit(unitType), nil
}

func (m *MemoryUnitFactory) GetValue(valueString string) (float64, error) {
	// Convert string to float64 and truncate to 2 decimal places
	valueFloat, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0, err
	}
	truncateValue := globalUtil.TruncateFloat(valueFloat, 2)
	return truncateValue, nil
}

func (m *MemoryUnitFactory) Validate(memConfig *v1.GenericConfigurationBean[float64]) (*unitsBean.ConfigValue[float64], error) {
	memConfigUnitSuffix := unitsBean.MemoryUnitStr(memConfig.Unit)
	memConfigUnit, ok := memConfigUnitSuffix.GetUnit()
	if !ok {
		errMsg := errors.InvalidUnitFound(memConfig.Unit, memConfig.Key)
		return nil, util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return unitsBean.NewConfigValue(memConfigUnit, memConfig.Value), nil
}
