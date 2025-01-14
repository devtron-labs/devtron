package audit

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository/audit"
	"strconv"
)

func GetInfraConfigTriggerAudit(config *v1.InfraConfig) ([]*audit.InfraConfigTriggerHistory, error) {
	infraConfigTriggerHistories := make([]*audit.InfraConfigTriggerHistory, 0)
	infraConfigTriggerHistories = append(infraConfigTriggerHistories, GetCpuLimit(config))
	infraConfigTriggerHistories = append(infraConfigTriggerHistories, GetCpuRequest(config))
	infraConfigTriggerHistories = append(infraConfigTriggerHistories, GetMemoryLimit(config))
	infraConfigTriggerHistories = append(infraConfigTriggerHistories, GetMemoryRequest(config))
	infraConfigTriggerHistories = append(infraConfigTriggerHistories, GetCiDefaultTimeout(config))
	infraConfigEntTriggerHistories, err := getInfraConfigEntTriggerAudit(config)
	if err != nil {
		return infraConfigTriggerHistories, err
	}
	infraConfigTriggerHistories = append(infraConfigTriggerHistories, infraConfigEntTriggerHistories...)
	return infraConfigTriggerHistories, nil
}

func GetCpuLimit(config *v1.InfraConfig) *audit.InfraConfigTriggerHistory {
	return &audit.InfraConfigTriggerHistory{
		ValueString: config.CiLimitCpu,
		Key:         v1.CPULimitKey,
	}
}

func GetCpuRequest(config *v1.InfraConfig) *audit.InfraConfigTriggerHistory {
	return &audit.InfraConfigTriggerHistory{
		ValueString: config.CiReqCpu,
		Key:         v1.CPURequestKey,
	}
}

func GetMemoryLimit(config *v1.InfraConfig) *audit.InfraConfigTriggerHistory {
	return &audit.InfraConfigTriggerHistory{
		ValueString: config.CiLimitMem,
		Key:         v1.MemoryLimitKey,
	}
}

func GetMemoryRequest(config *v1.InfraConfig) *audit.InfraConfigTriggerHistory {
	return &audit.InfraConfigTriggerHistory{
		ValueString: config.CiReqMem,
		Key:         v1.MemoryRequestKey,
	}
}

func GetCiDefaultTimeout(config *v1.InfraConfig) *audit.InfraConfigTriggerHistory {
	return &audit.InfraConfigTriggerHistory{
		ValueString: strconv.FormatFloat(config.CiDefaultTimeout, 'f', -1, 64),
		Key:         v1.TimeOutKey,
	}
}
