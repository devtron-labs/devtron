package bean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
)

type LockConfigRequest struct {
	Allowed bool     `json:"allowed, notnull"`
	Config  []string `json:"config, notnull"`
}

type LockConfigResponse struct {
	Id      int      `json:"id,pk"`
	Allowed bool     `json:"allowed, notnull"`
	Config  []string `json:"config, notnull"`
}

type LockConfiguration struct {
	tableName struct{} `sql:"lock_configuration"`
	Id        int      `sql:"id,pk"`
	Config    string   `sql:"config"`
	Active    bool     `sql:"active"`
	sql.AuditLog
}

type LockConfig struct {
	Path    string
	Allowed bool
}

func (impl *LockConfiguration) ConvertDBDtoToResponse() *LockConfigResponse {
	config, allowed := getConfigAndStatus(impl.Config)
	return &LockConfigResponse{
		Id:      impl.Id,
		Config:  config,
		Allowed: allowed,
	}
}

func (impl *LockConfigRequest) ConvertRequestToDBDto() *LockConfiguration {
	configByte := impl.getLockConfig()
	return &LockConfiguration{
		Config: string(configByte),
		Active: true,
	}
}

func getConfigAndStatus(config string) ([]string, bool) {
	var configs []string
	allowed := true
	var lockConfigs []LockConfig
	_ = json.Unmarshal([]byte(config), &lockConfigs)
	for _, lockConfig := range lockConfigs {
		configs = append(configs, lockConfig.Path)
		allowed = allowed && lockConfig.Allowed
	}
	return configs, allowed

}

func (impl *LockConfigRequest) getLockConfig() []byte {
	var lockConfigs []LockConfig
	for _, config := range impl.Config {
		lockConfig := LockConfig{
			Path:    config,
			Allowed: impl.Allowed,
		}
		lockConfigs = append(lockConfigs, lockConfig)
	}
	marsh, _ := json.Marshal(lockConfigs)
	return marsh
}
