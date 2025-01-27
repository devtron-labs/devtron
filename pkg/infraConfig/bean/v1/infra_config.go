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

// Package v1 implements the infra config with interface values.
package v1

import (
	"github.com/devtron-labs/devtron/api/bean"
	"math"
)

// InfraConfig is used for read-only purpose outside this package
type InfraConfig struct {
	// currently only for ci
	CiLimitCpu string `env:"LIMIT_CI_CPU" envDefault:"0.5"`
	CiLimitMem string `env:"LIMIT_CI_MEM" envDefault:"3G"`
	CiReqCpu   string `env:"REQ_CI_CPU" envDefault:"0.5"`
	CiReqMem   string `env:"REQ_CI_MEM" envDefault:"3G"`
	// CiDefaultTimeout is the default timeout for CI jobs in seconds
	// Earlier it was in int64, but now it is in float64
	CiDefaultTimeout float64 `env:"DEFAULT_TIMEOUT" envDefault:"3600"`

	// cm and cs
	ConfigMaps []bean.ConfigSecretMap `env:"-"`
	Secrets    []bean.ConfigSecretMap `env:"-"`
	InfraConfigEnt
}

func (infraConfig *InfraConfig) GetCiLimitCpu() string {
	if infraConfig == nil {
		return ""
	}
	return infraConfig.CiLimitCpu
}

func (infraConfig *InfraConfig) SetCiLimitCpu(cpu string) *InfraConfig {
	if infraConfig == nil {
		return nil
	}
	infraConfig.CiLimitCpu = cpu
	return infraConfig
}

func (infraConfig *InfraConfig) GetCiLimitMem() string {
	if infraConfig == nil {
		return ""
	}
	return infraConfig.CiLimitMem
}

func (infraConfig *InfraConfig) SetCiLimitMem(mem string) *InfraConfig {
	if infraConfig == nil {
		return nil
	}
	infraConfig.CiLimitMem = mem
	return infraConfig
}

func (infraConfig *InfraConfig) GetCiReqCpu() string {
	if infraConfig == nil {
		return ""
	}
	return infraConfig.CiReqCpu
}

func (infraConfig *InfraConfig) SetCiReqCpu(cpu string) *InfraConfig {
	if infraConfig == nil {
		return nil
	}
	infraConfig.CiReqCpu = cpu
	return infraConfig
}

func (infraConfig *InfraConfig) GetCiReqMem() string {
	if infraConfig == nil {
		return ""
	}
	return infraConfig.CiReqMem
}

func (infraConfig *InfraConfig) SetCiReqMem(mem string) *InfraConfig {
	if infraConfig == nil {
		return nil
	}
	infraConfig.CiReqMem = mem
	return infraConfig
}

func (infraConfig *InfraConfig) GetCiDefaultTimeout() float64 {
	if infraConfig == nil {
		return 0
	}
	return infraConfig.CiDefaultTimeout
}

func (infraConfig *InfraConfig) GetCiTimeoutInt() int64 {
	if infraConfig == nil {
		return 0
	}
	modifiedValue := math.Min(math.Floor(infraConfig.CiDefaultTimeout), math.MaxInt64)
	return int64(modifiedValue)
}

func (infraConfig *InfraConfig) SetCiTimeout(timeout float64) *InfraConfig {
	if infraConfig == nil {
		return nil
	}
	infraConfig.CiDefaultTimeout = timeout
	return infraConfig
}
