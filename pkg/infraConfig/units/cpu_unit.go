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
)

type CPUUnitFactory struct {
	logger   *zap.SugaredLogger
	cpuUnits map[bean2.CPUUnitStr]bean.Unit
}

func NewCPUUnitFactory(logger *zap.SugaredLogger) *CPUUnitFactory {
	return &CPUUnitFactory{
		logger:   logger,
		cpuUnits: bean2.GetCPUUnit(),
	}
}

func (c *CPUUnitFactory) GetAllUnits() map[string]bean.Unit {
	cpuUnits := c.cpuUnits
	units := make(map[string]bean.Unit)
	for key, value := range cpuUnits {
		units[string(key)] = value
	}
	return units
}

func (c *CPUUnitFactory) ParseValAndUnit(val string) (*bean2.ParsedValue, error) {
	return ParseCPUorMemoryValue(val)
}

func (c *CPUUnitFactory) Validate(profileBean, defaultProfile *bean.ProfileBeanDto) error {
	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *bean.ConfigurationBean
		cpuReq   *bean.ConfigurationBean
	)

	for _, platformConfigurations := range profileBean.Configurations {
		for _, configuration := range platformConfigurations {
			// get cpu limit and req
			switch configuration.Key {
			case bean.CPU_LIMIT:
				cpuLimit = configuration
			case bean.CPU_REQUEST:
				cpuReq = configuration
			}
		}
	}
	// validate cpu
	err := validateCPU(cpuLimit, cpuReq)
	if err != nil {
		return err
	}
	return nil
}

func validateCPU(cpuLimit, cpuReq *bean.ConfigurationBean) error {
	cpuLimitUnitSuffix := bean2.CPUUnitStr(cpuLimit.Unit)
	cpuLimitUnit, ok := cpuLimitUnitSuffix.GetUnit()
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, cpuLimit.Unit, cpuLimit.Key))
	}

	cpuLimitInterfaceVal, err := adapter.GetTypedValue(cpuLimit.Key, cpuLimit.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuLimit.Key, cpuLimit.Value))
	}

	cpuLimitVal, ok := cpuLimitInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuLimit.Key, cpuLimit.Value))
	}

	cpuReqUnitSuffix := bean2.CPUUnitStr(cpuReq.Unit)
	cpuReqUnit, ok := cpuReqUnitSuffix.GetUnit()
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, cpuReq.Unit, cpuReq.Key))
	}

	cpuReqInterfaceVal, err := adapter.GetTypedValue(cpuReq.Key, cpuReq.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuReq.Key, cpuReq.Value))
	}

	cpuReqVal, ok := cpuReqInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuReq.Key, cpuReq.Value))
	}

	// validate cpu limit and req
	if !validLimReq(cpuLimitVal, cpuLimitUnit.ConversionFactor, cpuReqVal, cpuReqUnit.ConversionFactor) {
		return errors.New(bean.CPULimReqErrorCompErr)
	}
	return nil
}

func validLimReq(lim, limFactor, req, reqFactor float64) bool {
	// this condition should be true for the valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRatio := lim / req
	convFactor := limFactor / reqFactor
	return limitToReqRatio*convFactor >= 1
}
