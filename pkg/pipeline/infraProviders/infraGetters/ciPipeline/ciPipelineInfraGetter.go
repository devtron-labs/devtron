package ciPipeline

import "github.com/devtron-labs/devtron/pkg/infraConfig"

// CiInfraGetter gets infra config for ci workflows
type CiInfraGetter struct {
	infraConfigService infraConfig.InfraConfigService
}

func NewCiInfraGetter(infraConfigService infraConfig.InfraConfigService) *CiInfraGetter {
	return &CiInfraGetter{infraConfigService: infraConfigService}
}

// GetInfraConfigurationsByScope gets infra config for ci workflows using the scope
func (ciInfraGetter CiInfraGetter) GetInfraConfigurationsByScope(scope *infraConfig.Scope) (*infraConfig.InfraConfig, error) {
	return ciInfraGetter.infraConfigService.GetInfraConfigurationsByScope(*scope)
}
