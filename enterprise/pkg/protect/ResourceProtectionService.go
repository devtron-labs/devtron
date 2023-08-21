package protect

import (
	"go.uber.org/zap"
	"reflect"
)

type ResourceProtectionService interface {
	ConfigureResourceProtection(request *ResourceProtectModel) error
	GetResourceProtectMetadata(appId int) ([]*ResourceProtectModel, error)
	ResourceProtectionEnabled(appId, envId int) bool
	ResourceProtectionEnabledForEnv(envId int) map[int]bool
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

func (impl *ResourceProtectionServiceImpl) RegisterListener(listener ResourceProtectionUpdateListener) {
	impl.logger.Infof("registering listener %s", reflect.TypeOf(listener))
	impl.listeners = append(impl.listeners, listener)
}

func (impl *ResourceProtectionServiceImpl) ConfigureResourceProtection(request *ResourceProtectModel) error {
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

func (impl *ResourceProtectionServiceImpl) GetResourceProtectMetadata(appId int) ([]*ResourceProtectModel, error) {
	protectionDtos, err := impl.resourceProtectionRepository.GetResourceProtectMetadata(appId)
	if err != nil {
		return nil, err
	}
	var resourceProtectModels []*ResourceProtectModel
	for _, protectionDto := range protectionDtos {
		resourceProtectModel := impl.convertToResourceProtectModel(protectionDto)
		resourceProtectModels = append(resourceProtectModels, resourceProtectModel)
	}
	return resourceProtectModels, nil
}

func (impl *ResourceProtectionServiceImpl) ResourceProtectionEnabled(appId, envId int) bool {
	resourceProtectionDto, err := impl.resourceProtectionRepository.GetResourceProtectionState(appId, envId)
	if err != nil {
		return false
	}
	protectionState := DisabledProtectionState
	if resourceProtectionDto != nil {
		protectionState = resourceProtectionDto.State
	}
	return protectionState == EnabledProtectionState
}

func (impl *ResourceProtectionServiceImpl) convertToResourceProtectModel(protectionDto *ResourceProtectionDto) *ResourceProtectModel {
	resourceProtectModel := &ResourceProtectModel{
		AppId:           protectionDto.AppId,
		EnvId:           protectionDto.EnvId,
		ProtectionState: protectionDto.State,
	}
	return resourceProtectModel
}

func (impl *ResourceProtectionServiceImpl) ResourceProtectionEnabledForEnv(envId int) map[int]bool {
	appVsState := make(map[int]bool)
	protectionDtos, err := impl.resourceProtectionRepository.GetResourceProtectionStateForEnv(envId)
	if err == nil {
		for _, protectionDto := range protectionDtos {
			appVsState[protectionDto.AppId] = protectionDto.State == EnabledProtectionState
		}
	}
	return appVsState
}


