package protect

import "go.uber.org/zap"

type ResourceProtectionService interface {
	ConfigureResourceProtection(request *ResourceProtectRequest) error
}

type ResourceProtectionServiceImpl struct {
	logger                       *zap.SugaredLogger
	resourceProtectionRepository ResourceProtectionRepository
}

func NewResourceProtectionServiceImpl(logger *zap.SugaredLogger, resourceProtectionRepository ResourceProtectionRepository) *ResourceProtectionServiceImpl {
	return &ResourceProtectionServiceImpl{
		logger:                       logger,
		resourceProtectionRepository: resourceProtectionRepository,
	}
}

func (impl ResourceProtectionServiceImpl) ConfigureResourceProtection(request *ResourceProtectRequest) error {
	impl.logger.Infow("configuring resource protection", "request", request)
	err := impl.resourceProtectionRepository.ConfigureResourceProtection(request.AppId, request.EnvId, request.ProtectionState, request.UserId)
	if err != nil {
		return err
	}
	// need to inform listeners

	return nil
}
