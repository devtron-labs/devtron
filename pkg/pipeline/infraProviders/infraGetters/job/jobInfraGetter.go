package job

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/infraConfig"
)

// JobInfraGetter gets infra config for job workflows
type JobInfraGetter struct {
	jobInfra infraConfig.InfraConfig
}

func NewJobInfraGetter() *JobInfraGetter {
	infra := infraConfig.InfraConfig{}
	env.Parse(&infra)
	return &JobInfraGetter{
		jobInfra: infra,
	}
}

// GetInfraConfigurationsByScope gets infra config for ci workflows using the scope
func (jobInfraGetter JobInfraGetter) GetInfraConfigurationsByScope(scope *infraConfig.Scope) (*infraConfig.InfraConfig, error) {
	infra := jobInfraGetter.jobInfra
	return &infra, nil
}
