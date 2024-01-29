package infraConfig

type InfraGetter interface {
	GetInfraConfigurationsByScope(scope Scope) (*InfraConfig, error)
}

// CiInfraGetter gets infra config for ci workflows
type CiInfraGetter struct {
	infraConfigServiceImpl *InfraConfigServiceImpl
}

// create similar infra Getter for pre/post cd workflows

// CdInfraGetter gets infra config for ci workflows
type CdInfraGetter struct {
	infraConfigServiceImpl *InfraConfigServiceImpl
}

func NewCiInfraGetter(infraConfigServiceImpl *InfraConfigServiceImpl) *CiInfraGetter {
	return &CiInfraGetter{infraConfigServiceImpl: infraConfigServiceImpl}
}

// GetInfraConfigurationsByScope gets infra config for ci workflows using the scope
func (ciInfraGetter CiInfraGetter) GetInfraConfigurationsByScope(scope Scope) (*InfraConfig, error) {
	return ciInfraGetter.infraConfigServiceImpl.getInfraConfigurationsByScope(scope)
}

// GetInfraConfigurationsByScope gets infra config for cd workflows using the scope
func (cdInfraGetter CdInfraGetter) GetInfraConfigurationsByScope(scope Scope) (*InfraConfig, error) {
	// currently unimplemented
	return nil, nil
}
