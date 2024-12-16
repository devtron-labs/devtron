package service

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/pkg/errors"
)

func (impl *InfraConfigServiceImpl) validateInfraConfig(profileBean *bean.ProfileBeanDto, defaultProfile *bean.ProfileBeanDto) error {

	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *bean.ConfigurationBean
		cpuReq   *bean.ConfigurationBean
		memLimit *bean.ConfigurationBean
		memReq   *bean.ConfigurationBean
		timeout  *bean.ConfigurationBean
	)

	for _, platformConfigurations := range profileBean.Configurations {
		for _, configuration := range platformConfigurations {
			// get cpu limit and req
			switch configuration.Key {
			case bean.CPU_LIMIT:
				cpuLimit = configuration
			case bean.CPU_REQUEST:
				cpuReq = configuration
			case bean.MEMORY_LIMIT:
				memLimit = configuration
			case bean.MEMORY_REQUEST:
				memReq = configuration
			case bean.TIME_OUT:
				timeout = configuration
			}
		}
	}

	// validate cpu
	err := impl.validateCPU(cpuLimit, cpuReq)
	if err != nil {
		return err
	}
	// validate mem
	err = impl.validateMEM(memLimit, memReq)
	if err != nil {
		return err
	}

	err = impl.validateTimeOut(timeout)
	if err != nil {
		return err
	}
	return nil
}
func (impl *InfraConfigServiceImpl) validateCPU(cpuLimit, cpuReq *bean.ConfigurationBean) error {
	cpuLimitUnitSuffix := units.CPUUnitStr(cpuLimit.Unit)
	cpuReqUnitSuffix := units.CPUUnitStr(cpuReq.Unit)
	cpuUnits := impl.units.GetCpuUnits()
	cpuLimitUnit, ok := cpuUnits[cpuLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, cpuLimit.Unit, cpuLimit.Key))
	}
	cpuReqUnit, ok := cpuUnits[cpuReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, cpuReq.Unit, cpuReq.Key))
	}

	cpuLimitInterfaceVal, err := util.GetTypedValue(cpuLimit.Key, cpuLimit.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuLimit.Key, cpuLimit.Value))
	}
	cpuLimitVal, ok := cpuLimitInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuLimit.Key, cpuLimit.Value))
	}

	cpuReqInterfaceVal, err := util.GetTypedValue(cpuReq.Key, cpuReq.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuReq.Key, cpuReq.Value))
	}
	cpuReqVal, ok := cpuReqInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, cpuReq.Key, cpuReq.Value))
	}
	if !validLimReq(cpuLimitVal, cpuLimitUnit.ConversionFactor, cpuReqVal, cpuReqUnit.ConversionFactor) {
		return errors.New(bean.CPULimReqErrorCompErr)
	}
	return nil
}
func (impl *InfraConfigServiceImpl) validateTimeOut(timeOut *bean.ConfigurationBean) error {
	if timeOut == nil {
		return nil
	}
	timeoutUnitSuffix := units.TimeUnitStr(timeOut.Unit)
	timeUnits := impl.units.GetTimeUnits()
	_, ok := timeUnits[timeoutUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, timeOut.Unit, timeOut.Key))
	}
	timeout, err := util.GetTypedValue(timeOut.Key, timeOut.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, timeOut.Key, timeOut.Value))
	}
	_, ok = timeout.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, timeOut.Key, timeOut.Value))
	}
	return nil
}
func (impl *InfraConfigServiceImpl) validateMEM(memLimit, memReq *bean.ConfigurationBean) error {
	memLimitUnitSuffix := units.MemoryUnitStr(memLimit.Unit)
	memReqUnitSuffix := units.MemoryUnitStr(memReq.Unit)
	memUnits := impl.units.GetMemoryUnits()
	memLimitUnit, ok := memUnits[memLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, memLimit.Unit, memLimit.Key))
	}
	memReqUnit, ok := memUnits[memReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidUnit, memReq.Unit, memReq.Key))
	}

	// Use getTypedValue to retrieve appropriate types
	memLimitInterfaceVal, err := util.GetTypedValue(memLimit.Key, memLimit.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, memLimit.Key, memLimit.Value))
	}
	memLimitVal, ok := memLimitInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, memLimit.Key, memLimit.Value))
	}

	memReqInterfaceVal, err := util.GetTypedValue(memReq.Key, memReq.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, memReq.Key, memReq.Value))
	}

	memReqVal, ok := memReqInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(bean.InvalidTypeValue, memReq.Key, memReq.Value))
	}

	if !validLimReq(memLimitVal, memLimitUnit.ConversionFactor, memReqVal, memReqUnit.ConversionFactor) {
		return errors.New(bean.MEMLimReqErrorCompErr)
	}
	return nil
}
func validLimReq(lim, limFactor, req, reqFactor float64) bool {
	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRatio := lim / req
	convFactor := limFactor / reqFactor
	return limitToReqRatio*convFactor >= 1
}
