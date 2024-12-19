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
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
	"go.uber.org/zap"
	"math"
	"strconv"
)

type TimeUnitFactory struct {
	logger    *zap.SugaredLogger
	timeUnits map[bean2.TimeUnitStr]bean.Unit
}

func NewTimeUnitFactory(logger *zap.SugaredLogger) *TimeUnitFactory {
	return &TimeUnitFactory{
		logger:    logger,
		timeUnits: bean2.GetTimeUnit(),
	}
}

func (t *TimeUnitFactory) GetAllUnits() map[string]bean.Unit {
	timeUnits := t.timeUnits
	units := make(map[string]bean.Unit)
	for key, value := range timeUnits {
		units[string(key)] = value
	}
	return units
}

func (t *TimeUnitFactory) ParseValAndUnit(val string) (*bean2.ParsedValue, error) {
	valueFloat, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	modifiedValue := math.Min(math.Floor(valueFloat), math.MaxInt64)
	return bean2.NewParsedValue().
		WithValueString(strconv.FormatInt(int64(modifiedValue), 10)).
		WithUnit(bean2.SecondStr.String()), nil
}

func (t *TimeUnitFactory) Validate(profileBean, defaultProfile *bean.ProfileBeanDto) error {
	// currently validating cpu and memory limits and reqs only
	var (
		timeout *bean.ConfigurationBean
	)

	for _, platformConfigurations := range profileBean.Configurations {
		for _, configuration := range platformConfigurations {
			// get cpu limit and req
			switch configuration.Key {
			case bean.TIME_OUT:
				timeout = configuration
			}
		}
	}

	// validate timeout
	err := validateTimeOut(timeout)
	if err != nil {
		return err
	}
	return nil
}

func validateTimeOut(timeOut *bean.ConfigurationBean) error {
	if timeOut == nil {
		return nil
	}
	timeoutUnitSuffix := bean2.TimeUnitStr(timeOut.Unit)
	_, ok := timeoutUnitSuffix.GetUnit()
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, timeOut.Unit, timeOut.Key))
	}
	timeout, err := adapter.GetTypedValue(timeOut.Key, timeOut.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, timeOut.Key, timeOut.Value))
	}
	_, ok = timeout.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, timeOut.Key, timeOut.Value))
	}
	return nil
}
