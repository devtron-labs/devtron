package v1

import "github.com/devtron-labs/devtron/api/bean"

type InfraConfigEnt struct {
	// cm and cs
	ConfigMaps []bean.ConfigSecretMap `env:"-"`
	Secrets    []bean.ConfigSecretMap `env:"-"`
}

func (infraConfig *InfraConfig) SetCiVariableSnapshot(value map[string]string) *InfraConfig {
	return infraConfig
}
