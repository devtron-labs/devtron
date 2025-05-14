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
