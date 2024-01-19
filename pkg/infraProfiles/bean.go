package infraProfiles

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraProfiles/units"
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

type InfraProfileConfiguration struct {
	tableName struct{}         `sql:"infra_profile_configuration"`
	Id        int              `sql:"id"`
	Key       ConfigKey        `sql:"name"`
	Value     string           `sql:"description"`
	Unit      units.UnitSuffix `sql:"unit"`
	ProfileId int              `sql:"profile_id"`
	Active    bool             `sql:"active"`
	sql.AuditLog
}

// service layer structs

type Profile struct {
	Id            int             `json:"id"`
	Name          string          `json:"name" validate:"required"`
	Description   string          `json:"description"`
	Active        bool            `json:"active"`
	Configuration []Configuration `json:"configuration"`
	AppCount      int             `json:"appCount"`
	CreatedBy     int32           `json:"createdBy"`
	CreatedOn     time.Time       `json:"createdOn"`
	UpdatedBy     int32           `json:"updatedBy"`
	UpdatedOn     time.Time       `json:"updatedOn"`
}

type Configuration struct {
	Id          int          `json:"id"`
	Key         ConfigKeyStr `json:"key" validate:"required"`
	Value       string       `json:"value" validate:"required"`
	Unit        string       `json:"unit"`
	ProfileName int          `json:"profileName"`
	Active      bool         `json:"active"`
}

type ProfileResponse struct {
	Profile               Profile                 `json:"profile"`
	DefaultConfigurations []Configuration         `json:"defaultConfigurations"`
	ConfigurationUnits    map[string][]units.Unit `json:"configurationUnits"`
}

type ProfilesResponse struct {
	Profiles              []Profile               `json:"profiles"`
	DefaultConfigurations []Configuration         `json:"defaultConfigurations"`
	ConfigurationUnits    map[string][]units.Unit `json:"configurationUnits"`
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
