/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
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
	"go.uber.org/zap"
	"math"
	"net/http"
	"strconv"
)

type TimeUnitFactory struct {
	logger    *zap.SugaredLogger
	timeUnits map[unitsBean.TimeUnitStr]v1.Unit
}

func NewTimeUnitFactory(logger *zap.SugaredLogger) *TimeUnitFactory {
	return &TimeUnitFactory{
		logger:    logger,
		timeUnits: unitsBean.GetTimeUnit(),
	}
}

func (t *TimeUnitFactory) GetAllUnits() map[string]v1.Unit {
	timeUnits := t.timeUnits
	units := make(map[string]v1.Unit)
	for key, value := range timeUnits {
		units[string(key)] = value
	}
	return units
}

func (t *TimeUnitFactory) GetDefaultUnitSuffix() string {
	var defaultUnit unitsBean.UnitType
	return defaultUnit.GetTimeUnitStr().String()
}

func (t *TimeUnitFactory) ParseValAndUnit(val float64, unitType unitsBean.UnitType) (*unitsBean.ParsedValue, error) {
	modifiedValue := math.Min(math.Floor(val), math.MaxInt64)
	return unitsBean.NewParsedValue().
		WithValueString(strconv.FormatInt(int64(modifiedValue), 10)).
		WithUnit(unitType), nil
}

func (t *TimeUnitFactory) GetValue(valueString string) (float64, error) {
	// Convert string to float64 and ensure it's within integer range
	valueFloat, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0, err
	}
	modifiedValue := math.Min(math.Floor(valueFloat), math.MaxInt64)
	return modifiedValue, nil
}

func (t *TimeUnitFactory) Validate(timeConfig *v1.GenericConfigurationBean[float64]) (*unitsBean.ConfigValue[float64], error) {
	if timeConfig == nil {
		return &unitsBean.ConfigValue[float64]{}, nil
	}
	timeoutUnitSuffix := unitsBean.TimeUnitStr(timeConfig.Unit)
	timeoutUnit, ok := timeoutUnitSuffix.GetUnit()
	if !ok {
		errMsg := errors.InvalidUnitFound(timeConfig.Unit, timeConfig.Key)
		return nil, util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return unitsBean.NewConfigValue(timeoutUnit, timeConfig.Value), nil
}
