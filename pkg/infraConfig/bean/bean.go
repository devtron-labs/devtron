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

package bean

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"time"
)

// service layer structs

type ProfileBeanAbstract struct {
	Id          int         `json:"id"`
	Name        string      `json:"name" validate:"required,min=1,max=50"`
	Description string      `json:"description" validate:"max=300"`
	Active      bool        `json:"active"`
	Type        ProfileType `json:"type"`
	AppCount    int         `json:"appCount"`
	CreatedBy   int32       `json:"createdBy"`
	CreatedOn   time.Time   `json:"createdOn"`
	UpdatedBy   int32       `json:"updatedBy"`
	UpdatedOn   time.Time   `json:"updatedOn"`
}

type ProfileBeanDto struct {
	ProfileBeanAbstract
	Configurations map[string][]*ConfigurationBean `json:"configurations"`
}

// Deprecated
type ProfileBeanV0 struct {
	ProfileBeanAbstract
	Configurations []ConfigurationBeanV0 `json:"configurations" validate:"dive"`
}

type ConfigurationBean struct {
	ConfigurationBeanAbstract
	Value interface{} `json:"value"`
}

// Deprecated
type ConfigurationBeanV0 struct {
	ConfigurationBeanAbstract
	Value float64 `json:"value" validate:"required,gt=0"`
}

type InfraConfigMetaData struct {
	DefaultConfigurations map[string][]*ConfigurationBean        `json:"defaultConfigurations"`
	ConfigurationUnits    map[ConfigKeyStr]map[string]units.Unit `json:"configurationUnits"`
}

// Deprecated
type InfraConfigMetaDataV0 struct {
	DefaultConfigurations []ConfigurationBeanV0                  `json:"defaultConfigurations"`
	ConfigurationUnits    map[ConfigKeyStr]map[string]units.Unit `json:"configurationUnits"`
}

type ProfileResponse struct {
	Profile ProfileBeanDto `json:"profile"`
	InfraConfigMetaData
}

// Deprecated
type ProfileResponseV0 struct {
	Profile ProfileBeanV0 `json:"profile"`
	InfraConfigMetaDataV0
}

type Scope struct {
	AppId int
}

type ConfigurationBeanAbstract struct {
	Id          int          `json:"id"`
	Key         ConfigKeyStr `json:"key"`
	Unit        string       `json:"unit" validate:"required,gt=0"`
	ProfileName string       `json:"profileName"`
	ProfileId   int          `json:"profileId"`
	Active      bool         `json:"active"`
}

// InfraConfig is used for read only purpose outside this package
type InfraConfig struct {
	// currently only for ci
	CiLimitCpu       string `env:"LIMIT_CI_CPU" envDefault:"0.5"`
	CiLimitMem       string `env:"LIMIT_CI_MEM" envDefault:"3G"`
	CiReqCpu         string `env:"REQ_CI_CPU" envDefault:"0.5"`
	CiReqMem         string `env:"REQ_CI_MEM" envDefault:"3G"`
	CiDefaultTimeout int64  `env:"DEFAULT_TIMEOUT" envDefault:"3600"`
}

func (infraConfig InfraConfig) GetCiLimitCpu() string {
	return infraConfig.CiLimitCpu
}

func (infraConfig *InfraConfig) SetCiLimitCpu(cpu string) {
	infraConfig.CiLimitCpu = cpu
}

func (infraConfig InfraConfig) GetCiLimitMem() string {
	return infraConfig.CiLimitMem
}

func (infraConfig *InfraConfig) SetCiLimitMem(mem string) {
	infraConfig.CiLimitMem = mem
}

func (infraConfig InfraConfig) GetCiReqCpu() string {
	return infraConfig.CiReqCpu
}

func (infraConfig *InfraConfig) SetCiReqCpu(cpu string) {
	infraConfig.CiReqCpu = cpu
}

func (infraConfig InfraConfig) GetCiReqMem() string {
	return infraConfig.CiReqMem
}

func (infraConfig *InfraConfig) SetCiReqMem(mem string) {
	infraConfig.CiReqMem = mem
}

func (infraConfig InfraConfig) GetCiDefaultTimeout() int64 {
	return infraConfig.CiDefaultTimeout
}

func (infraConfig *InfraConfig) SetCiDefaultTimeout(timeout int64) {
	infraConfig.CiDefaultTimeout = timeout
}
