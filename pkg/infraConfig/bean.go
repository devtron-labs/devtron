package infraConfig

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/pkg/errors"
	"time"
)

type ConfigKey int

const CPULimit ConfigKey = 1
const CPURequest ConfigKey = 2
const MemoryLimit ConfigKey = 3
const MemoryRequest ConfigKey = 4
const TimeOut ConfigKey = 5

type ConfigKeyStr string

const CPU_LIMIT ConfigKeyStr = "cpu_limit"
const CPU_REQUEST ConfigKeyStr = "cpu_request"
const MEMORY_LIMIT ConfigKeyStr = "memory_limit"
const MEMORY_REQUEST ConfigKeyStr = "memory_request"
const TIME_OUT ConfigKeyStr = "timeout"

func GetConfigKeyStr(configKey ConfigKey) ConfigKeyStr {
	switch configKey {
	case CPULimit:
		return CPU_LIMIT
	case CPURequest:
		return CPU_REQUEST
	case MemoryLimit:
		return MEMORY_LIMIT
	case MemoryRequest:
		return MEMORY_REQUEST
	case TimeOut:
		return TIME_OUT
	}
	return ""
}

func GetConfigKey(configKeyStr ConfigKeyStr) ConfigKey {
	switch configKeyStr {
	case CPU_LIMIT:
		return CPULimit
	case CPU_REQUEST:
		return CPURequest
	case MEMORY_LIMIT:
		return MemoryLimit
	case MEMORY_REQUEST:
		return MemoryRequest
	case TIME_OUT:
		return TimeOut
	}
	return 0
}

// repo structs

type InfraProfile struct {
	tableName   struct{} `sql:"infra_profile"`
	Id          int      `sql:"id"`
	Name        string   `sql:"name"`
	Description string   `sql:"description"`
	Active      bool     `sql:"active"`
	sql.AuditLog
}

func (infraProfile *InfraProfile) ConvertToProfileBean() ProfileBean {
	return ProfileBean{
		Id:          infraProfile.Id,
		Name:        infraProfile.Name,
		Description: infraProfile.Description,
		Active:      infraProfile.Active,
		CreatedBy:   infraProfile.CreatedBy,
		CreatedOn:   infraProfile.CreatedOn,
		UpdatedBy:   infraProfile.UpdatedBy,
		UpdatedOn:   infraProfile.UpdatedOn,
	}
}

type InfraProfileConfiguration struct {
	tableName    struct{}         `sql:"infra_profile_configuration"`
	Id           int              `sql:"id"`
	Key          ConfigKey        `sql:"name"`
	Value        string           `sql:"description"`
	Unit         units.UnitSuffix `sql:"unit"`
	ProfileId    int              `sql:"profile_id"`
	Active       bool             `sql:"active"`
	InfraProfile InfraProfile
	sql.AuditLog
}

func (infraProfileConfiguration *InfraProfileConfiguration) ConvertToConfigurationBean() ConfigurationBean {
	return ConfigurationBean{
		Id:    infraProfileConfiguration.Id,
		Key:   GetConfigKeyStr(infraProfileConfiguration.Key),
		Value: infraProfileConfiguration.Value,
		// Unit:
		ProfileId: infraProfileConfiguration.ProfileId,
		Active:    infraProfileConfiguration.Active,
	}
}

// service layer structs

type ProfileBean struct {
	Id             int                 `json:"id"`
	Name           string              `json:"name" validate:"required"`
	Description    string              `json:"description"`
	Active         bool                `json:"active"`
	Configurations []ConfigurationBean `json:"configuration"`
	AppCount       int                 `json:"appCount"`
	CreatedBy      int32               `json:"createdBy"`
	CreatedOn      time.Time           `json:"createdOn"`
	UpdatedBy      int32               `json:"updatedBy"`
	UpdatedOn      time.Time           `json:"updatedOn"`
}

func (profileBean *ProfileBean) ConvertToInfraProfile() *InfraProfile {
	return &InfraProfile{
		Id:          profileBean.Id,
		Name:        profileBean.Name,
		Description: profileBean.Description,
	}
}

type ConfigurationBean struct {
	Id          int          `json:"id"`
	Key         ConfigKeyStr `json:"key" validate:"required"`
	Value       string       `json:"value" validate:"required"`
	Unit        string       `json:"unit"`
	ProfileName string       `json:"profileName"`
	ProfileId   int          `json:"profileId"`
	Active      bool         `json:"active"`
}

func (configurationBean *ConfigurationBean) ConvertToInfraProfileConfiguration() *InfraProfileConfiguration {
	return &InfraProfileConfiguration{
		Id:    configurationBean.Id,
		Key:   GetConfigKey(configurationBean.Key),
		Value: configurationBean.Value,
		// Unit:      units.GetUnitSuffix(configurationBean.Unit),
		ProfileId: configurationBean.ProfileId,
		Active:    configurationBean.Active,
	}
}

type InfraConfigMetaData struct {
	DefaultConfigurations []ConfigurationBean           `json:"defaultConfigurations"`
	ConfigurationUnits    map[ConfigKeyStr][]units.Unit `json:"configurationUnits"`
}
type ProfileResponse struct {
	Profile ProfileBean `json:"profile"`
	InfraConfigMetaData
}

type ProfilesResponse struct {
	Profiles []ProfileBean `json:"profiles"`
	InfraConfigMetaData
}

type InfraConfig struct {
	// currently only for ci
	CiLimitCpu       string `env:"LIMIT_CI_CPU" envDefault:"0.5"`
	CiLimitMem       string `env:"LIMIT_CI_MEM" envDefault:"3G"`
	CiReqCpu         string `env:"REQ_CI_CPU" envDefault:"0.5"`
	CiReqMem         string `env:"REQ_CI_MEM" envDefault:"3G"`
	CiDefaultTimeout int64  `env:"DEFAULT_TIMEOUT" envDefault:"3600"`
	// TODO: add cd config in future
}

func (infraConfig InfraConfig) GetCiLimitCpu() (*InfraProfileConfiguration, error) {
	positive, _, num, _, suffix, err := units.ParseQuantityString(infraConfig.CiLimitCpu)
	if err != nil {
		return nil, err
	}
	if !positive {
		return nil, errors.New("negative value not allowed for cpu limits")
	}

	return &InfraProfileConfiguration{
		Key:   CPULimit,
		Value: num,
		Unit:  units.GetCPUUnit(units.CPUUnitStr(suffix)),
	}, nil

}

func (infraConfig InfraConfig) GetCiLimitMem() (*InfraProfileConfiguration, error) {
	positive, _, num, _, suffix, err := units.ParseQuantityString(infraConfig.CiLimitMem)
	if err != nil {
		return nil, err
	}
	if !positive {
		return nil, errors.New("negative value not allowed for memory limits")
	}

	return &InfraProfileConfiguration{
		Key:   MemoryLimit,
		Value: num,
		Unit:  units.GetMemoryUnit(units.MemoryUnitStr(suffix)),
	}, nil

}

func (infraConfig InfraConfig) GetCiReqCpu() (*InfraProfileConfiguration, error) {
	positive, _, num, _, suffix, err := units.ParseQuantityString(infraConfig.CiReqCpu)
	if err != nil {
		return nil, err
	}
	if !positive {
		return nil, errors.New("negative value not allowed for cpu requests")
	}

	return &InfraProfileConfiguration{
		Key:   CPURequest,
		Value: num,
		Unit:  units.GetCPUUnit(units.CPUUnitStr(suffix)),
	}, nil
}

func (infraConfig InfraConfig) GetCiReqMem() (*InfraProfileConfiguration, error) {
	positive, _, num, _, suffix, err := units.ParseQuantityString(infraConfig.CiReqMem)
	if err != nil {
		return nil, err
	}
	if !positive {
		return nil, errors.New("negative value not allowed for memory requests")
	}

	return &InfraProfileConfiguration{
		Key:   MemoryRequest,
		Value: num,
		Unit:  units.GetMemoryUnit(units.MemoryUnitStr(suffix)),
	}, nil
}

func (infraConfig InfraConfig) GetDefaultTimeout() (*InfraProfileConfiguration, error) {
	return &InfraProfileConfiguration{
		Key:   TimeOut,
		Value: fmt.Sprintf("%d", infraConfig.CiDefaultTimeout),
		Unit:  units.GetTimeUnit(units.SecondStr),
	}, nil
}

// todo: delete this function
// Transform will iterate through elements of input slice and apply transform function on each object
// and returns the transformed slice
func Transform[T any, K any](input []T, transform func(inp T) K) []K {

	res := make([]K, len(input))
	for i, _ := range input {
		res[i] = transform(input[i])
	}
	return res

}
