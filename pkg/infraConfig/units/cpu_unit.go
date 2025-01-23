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

type CPUUnitFactory struct {
	logger   *zap.SugaredLogger
	cpuUnits map[unitsBean.CPUUnitStr]v1.Unit
}

func NewCPUUnitFactory(logger *zap.SugaredLogger) *CPUUnitFactory {
	return &CPUUnitFactory{
		logger:   logger,
		cpuUnits: unitsBean.GetCPUUnit(),
	}
}

func (c *CPUUnitFactory) GetAllUnits() map[string]v1.Unit {
	cpuUnits := c.cpuUnits
	units := make(map[string]v1.Unit)
	for key, value := range cpuUnits {
		units[string(key)] = value
	}
	return units
}

func (c *CPUUnitFactory) GetDefaultUnitSuffix() string {
	var defaultUnit unitsBean.UnitType
	return defaultUnit.GetCPUUnitStr().String()
}

func (c *CPUUnitFactory) ParseValAndUnit(val float64, unitType unitsBean.UnitType) (*unitsBean.ParsedValue, error) {
	return unitsBean.NewParsedValue().
		WithValueString(strconv.FormatFloat(val, 'f', -1, 64)).
		WithUnit(unitType), nil
}

func (c *CPUUnitFactory) GetValue(valueString string) (float64, error) {
	// Convert string to float64 and truncate to 2 decimal places
	valueFloat, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0, err
	}
	truncateValue := globalUtil.TruncateFloat(valueFloat, 2)
	return truncateValue, nil
}

func (c *CPUUnitFactory) Validate(cpuConfig *v1.GenericConfigurationBean[float64]) (*unitsBean.ConfigValue[float64], error) {
	cpuUnitSuffix := unitsBean.CPUUnitStr(cpuConfig.Unit)
	cpuConfigUnit, ok := cpuUnitSuffix.GetUnit()
	if !ok {
		errMsg := errors.InvalidUnitFound(cpuConfig.Unit, cpuConfig.Key)
		return nil, util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return unitsBean.NewConfigValue(cpuConfigUnit, cpuConfig.Value), nil
}
