package infraGetters

import "github.com/devtron-labs/devtron/pkg/infraConfig"

type InfraGetter interface {
	GetInfraConfigurationsByScope(scope *infraConfig.Scope) (*infraConfig.InfraConfig, error)
}
