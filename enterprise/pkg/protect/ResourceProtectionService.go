package protect

import (
	"go.uber.org/zap"
	"reflect"
)

type ResourceProtectionService interface {
	ConfigureResourceProtection(request *ResourceProtectRequest) error
	RegisterListener(listener ResourceProtectionUpdateListener)
}

type ResourceProtectionUpdateListener interface {
	OnStateChange(appId int, envId int, state ProtectionState, userId int32)
}

type ResourceProtectionServiceImpl struct {
	logger                       *zap.SugaredLogger
	resourceProtectionRepository ResourceProtectionRepository
	listeners                    []ResourceProtectionUpdateListener
}

func NewResourceProtectionServiceImpl(logger *zap.SugaredLogger, resourceProtectionRepository ResourceProtectionRepository) *ResourceProtectionServiceImpl {
	return &ResourceProtectionServiceImpl{
		logger:                       logger,
		resourceProtectionRepository: resourceProtectionRepository,
	}
}

func (impl ResourceProtectionServiceImpl) RegisterListener(listener ResourceProtectionUpdateListener) {
	impl.logger.Infof("registering listener %s", reflect.TypeOf(listener))
	impl.listeners = append(impl.listeners, listener)
}

func (impl ResourceProtectionServiceImpl) ConfigureResourceProtection(request *ResourceProtectRequest) error {
	impl.logger.Infow("configuring resource protection", "request", request)
	err := impl.resourceProtectionRepository.ConfigureResourceProtection(request.AppId, request.EnvId, request.ProtectionState, request.UserId)
	if err != nil {
		return err
	}
	for _, protectionUpdateListener := range impl.listeners {
		protectionUpdateListener.OnStateChange(request.AppId, request.EnvId, request.ProtectionState, request.UserId)
	}
	return nil
}
