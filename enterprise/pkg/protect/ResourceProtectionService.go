package protect

import "go.uber.org/zap"

type ConfigProtectionService interface {
	ConfigureConfigProtection(configProtectionState ProtectionState, appId int, envId int)
}

type ConfigProtectionServiceImpl struct {
	logger                     *zap.SugaredLogger
	configProtectionRepository ResourceProtectionRepository
}

func NewConfigProtectionServiceImpl(logger *zap.SugaredLogger, configProtectionRepository ResourceProtectionRepository) *ConfigProtectionServiceImpl {
	return &ConfigProtectionServiceImpl{
		logger:                     logger,
		configProtectionRepository: configProtectionRepository,
	}
}

func (impl ConfigProtectionServiceImpl) ConfigureConfigProtection(configProtectionState ProtectionState, appId int, envId int) {

}
