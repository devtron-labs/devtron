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
	util2 "github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"strconv"
	"time"
)

// repo structs

type InfraProfileEntity struct {
	tableName   struct{} `sql:"infra_profile" pg:",discard_unknown_columns"`
	Id          int      `sql:"id"`
	Name        string   `sql:"name"`
	Description string   `sql:"description"`
	Active      bool     `sql:"active"`
	sql.AuditLog
}

func (infraProfile *InfraProfileEntity) ConvertToProfileBean() ProfileBean {
	profileType := util2.DEFAULT
	if infraProfile.Name != util2.DEFAULT_PROFILE_NAME {
		profileType = util2.NORMAL
	}
	return ProfileBean{
		Id:          infraProfile.Id,
		Name:        infraProfile.Name,
		Type:        profileType,
		Description: infraProfile.Description,
		Active:      infraProfile.Active,
		CreatedBy:   infraProfile.CreatedBy,
		CreatedOn:   infraProfile.CreatedOn,
		UpdatedBy:   infraProfile.UpdatedBy,
		UpdatedOn:   infraProfile.UpdatedOn,
	}
}

type InfraProfileConfigurationEntity struct {
	tableName struct{}         `sql:"infra_profile_configuration" pg:",discard_unknown_columns"`
	Id        int              `sql:"id"`
	Key       util2.ConfigKey  `sql:"key"`
	Value     string           `sql:"value"`
	Unit      units.UnitSuffix `sql:"unit"`
	ProfileId int              `sql:"profile_id"`
	Platform  string           `sql:"platform"`
	Active    bool             `sql:"active"`
	sql.AuditLog
}

// service layer structs

type ProfileBean struct {
	Id             int                             `json:"id"`
	Name           string                          `json:"name" validate:"required,min=1,max=50"`
	Description    string                          `json:"description" validate:"max=300"`
	Active         bool                            `json:"active"`
	Configurations map[string][]*ConfigurationBean `json:"configurations" validate:"dive"`
	Type           util2.ProfileType               `json:"type"`
	AppCount       int                             `json:"appCount"`
	CreatedBy      int32                           `json:"createdBy"`
	CreatedOn      time.Time                       `json:"createdOn"`
	UpdatedBy      int32                           `json:"updatedBy"`
	UpdatedOn      time.Time                       `json:"updatedOn"`
}

func (profileBean *ProfileBean) ConvertToInfraProfileEntity() *InfraProfileEntity {
	return &InfraProfileEntity{
		Id:          profileBean.Id,
		Name:        profileBean.Name,
		Description: profileBean.Description,
	}
}

type ConfigurationBean struct {
	Id          int                `json:"id"`
	Key         util2.ConfigKeyStr `json:"key"`
	Value       string             `json:"value" validate:"required,gt=0"`
	Unit        string             `json:"unit" validate:"required,gt=0"`
	ProfileName string             `json:"profileName"`
	ProfileId   int                `json:"profileId"`
	Active      bool               `json:"active"`
}

type InfraConfigMetaData struct {
	DefaultConfigurations []ConfigurationBean                          `json:"defaultConfigurations"`
	ConfigurationUnits    map[util2.ConfigKeyStr]map[string]units.Unit `json:"configurationUnits"`
}
type ProfileResponse struct {
	Profile ProfileBean `json:"profile"`
	InfraConfigMetaData
}

type Scope struct {
	AppId int
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

func (infraConfig InfraConfig) LoadCiLimitCpu() (*InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiLimitCpu)
	if err != nil {
		return nil, err
	}
	return &InfraProfileConfigurationEntity{
		Key:   util2.CPULimit,
		Value: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:  units.CPUUnitStr(suffix).GetCPUUnit(),
	}, nil

}

func (infraConfig InfraConfig) LoadCiLimitMem() (*InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiLimitMem)
	if err != nil {
		return nil, err
	}
	return &InfraProfileConfigurationEntity{
		Key:   util2.MemoryLimit,
		Value: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:  units.MemoryUnitStr(suffix).GetMemoryUnit(),
	}, nil

}

func (infraConfig InfraConfig) LoadCiReqCpu() (*InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiReqCpu)
	if err != nil {
		return nil, err
	}
	return &InfraProfileConfigurationEntity{
		Key:   util2.CPURequest,
		Value: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:  units.CPUUnitStr(suffix).GetCPUUnit(),
	}, nil
}

func (infraConfig InfraConfig) LoadCiReqMem() (*InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiReqMem)
	if err != nil {
		return nil, err
	}

	return &InfraProfileConfigurationEntity{
		Key:   util2.MemoryRequest,
		Value: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:  units.MemoryUnitStr(suffix).GetMemoryUnit(),
	}, nil
}

func (infraConfig InfraConfig) LoadDefaultTimeout() (*InfraProfileConfigurationEntity, error) {
	return &InfraProfileConfigurationEntity{
		Key:   util2.TimeOut,
		Value: strconv.FormatInt(infraConfig.CiDefaultTimeout, 10),
		Unit:  units.SecondStr.GetTimeUnit(),
	}, nil
}

func (infraConfig InfraConfig) LoadInfraConfigInEntities() ([]*InfraProfileConfigurationEntity, error) {
	cpuLimit, err := infraConfig.LoadCiLimitCpu()
	if err != nil {
		return nil, err
	}
	memLimit, err := infraConfig.LoadCiLimitMem()
	if err != nil {
		return nil, err
	}
	cpuReq, err := infraConfig.LoadCiReqCpu()
	if err != nil {
		return nil, err
	}
	memReq, err := infraConfig.LoadCiReqMem()
	if err != nil {
		return nil, err
	}
	timeout, err := infraConfig.LoadDefaultTimeout()
	if err != nil {
		return nil, err
	}

	defaultConfigurations := []*InfraProfileConfigurationEntity{cpuLimit, memLimit, cpuReq, memReq, timeout}
	return defaultConfigurations, nil
}
