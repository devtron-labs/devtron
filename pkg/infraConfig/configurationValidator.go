package infraConfig

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/pkg/errors"
)

type Validator interface {
	ValidateCpuMem(profileBean *ProfileBean, defaultProfile *ProfileBean) error
}

type ValidatorImpl struct {
	units *units.Units
}

func NewValidatorImpl(units *units.Units) *ValidatorImpl {
	return &ValidatorImpl{
		units: units,
	}
}

func (impl *ValidatorImpl) ValidateCpuMem(profileBean *ProfileBean, defaultProfile *ProfileBean) error {
	var (
		cpuLimit *ConfigurationBean
		cpuReq   *ConfigurationBean
		memLimit *ConfigurationBean
		memReq   *ConfigurationBean
	)

	for i, _ := range profileBean.Configurations {
		// get cpu limit and req
		switch profileBean.Configurations[i].Key {
		case CPU_LIMIT:
			cpuLimit = &profileBean.Configurations[i]
		case CPU_REQUEST:
			cpuReq = &profileBean.Configurations[i]
		case MEMORY_LIMIT:
			memLimit = &profileBean.Configurations[i]
		case MEMORY_REQUEST:
			memReq = &profileBean.Configurations[i]
		}
	}

	for i, _ := range defaultProfile.Configurations {
		// get cpu limit and req
		switch defaultProfile.Configurations[i].Key {
		case CPU_LIMIT:
			if cpuLimit == nil {
				cpuLimit = &defaultProfile.Configurations[i]
			}
		case CPU_REQUEST:
			if cpuReq == nil {
				cpuReq = &defaultProfile.Configurations[i]
			}
		case MEMORY_LIMIT:
			if memLimit == nil {
				memLimit = &defaultProfile.Configurations[i]
			}
		case MEMORY_REQUEST:
			if memReq == nil {
				memReq = &defaultProfile.Configurations[i]
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
	return nil
}

func (impl *ValidatorImpl) validateCPU(cpuLimit, cpuReq *ConfigurationBean) error {
	cpuLimitUnitSuffix := units.CPUUnitStr(cpuLimit.Unit)
	cpuReqUnitSuffix := units.CPUUnitStr(cpuReq.Unit)
	cpuUnits := impl.units.GetCpuUnits()
	cpuLimitUnit, ok := cpuUnits[cpuLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, cpuLimit.Unit, cpuLimit.Key))
	}
	cpuReqUnit, ok := cpuUnits[cpuReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, cpuReq.Unit, cpuReq.Key))
	}

	if !validLimReq(cpuLimit.Value, cpuLimitUnit.ConversionFactor, cpuReq.Value, cpuReqUnit.ConversionFactor) {
		return errors.New(CPULimReqErrorCompErr)
	}
	return nil
}

func (impl *ValidatorImpl) validateMEM(memLimit, memReq *ConfigurationBean) error {
	memLimitUnitSuffix := units.MemoryUnitStr(memLimit.Unit)
	memReqUnitSuffix := units.MemoryUnitStr(memReq.Unit)
	memUnits := impl.units.GetMemoryUnits()
	memLimitUnit, ok := memUnits[memLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, memLimit.Unit, memLimit.Key))
	}
	memReqUnit, ok := memUnits[memReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, memReq.Unit, memReq.Key))
	}

	if !validLimReq(memLimit.Value, memLimitUnit.ConversionFactor, memReq.Value, memReqUnit.ConversionFactor) {
		return errors.New(MEMLimReqErrorCompErr)
	}
	return nil
}

func validLimReq(lim, limFactor, req, reqFactor float64) bool {
	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRatio := lim / req
	convFactor := limFactor / reqFactor
	return limitToReqRatio*convFactor >= 1
}
