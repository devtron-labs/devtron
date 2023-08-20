package tests

import (
	"errors"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestResourceProtectConfig(t *testing.T) {
	sugardLogger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	t.Run("enable resource protection", func(t *testing.T) {
		resourceProtectionRepository, protectionServiceImpl := getServices(t, sugardLogger)
		request := getResourceProtectModel(protect.EnabledProtectionState)
		resourceProtectionRepository.On("ConfigureResourceProtection", request.AppId, request.EnvId, request.ProtectionState, request.UserId).Return(nil)
		err = protectionServiceImpl.ConfigureResourceProtection(request)
		resourceProtectionRepository.AssertCalled(t, "ConfigureResourceProtection", request.AppId, request.EnvId, request.ProtectionState, request.UserId)
		assert.NoError(t, err)
	})
	t.Run("disable resource protection", func(t *testing.T) {
		resourceProtectionRepository, protectionServiceImpl := getServices(t, sugardLogger)
		request := getResourceProtectModel(protect.DisabledProtectionState)
		resourceProtectionRepository.On("ConfigureResourceProtection", request.AppId, request.EnvId, request.ProtectionState, request.UserId).Return(nil)
		resourceProtectionUpdateListener := mocks.NewResourceProtectionUpdateListener(t)
		resourceProtectionUpdateListener.On("OnStateChange", request.AppId, request.EnvId, request.ProtectionState, request.UserId).Return()
		protectionServiceImpl.RegisterListener(resourceProtectionUpdateListener)
		err = protectionServiceImpl.ConfigureResourceProtection(request)
		resourceProtectionUpdateListener.AssertCalled(t, "OnStateChange", request.AppId, request.EnvId, request.ProtectionState, request.UserId)
		assert.NoError(t, err)
	})

	t.Run("disable resource protection with error", func(t *testing.T) {
		resourceProtectionRepository, protectionServiceImpl := getServices(t, sugardLogger)
		request := getResourceProtectModel(protect.DisabledProtectionState)
		customErr := errors.New("failed to save")
		resourceProtectionRepository.On("ConfigureResourceProtection", request.AppId, request.EnvId, request.ProtectionState, request.UserId).Return(customErr)
		resourceProtectionUpdateListener := mocks.NewResourceProtectionUpdateListener(t)
		protectionServiceImpl.RegisterListener(resourceProtectionUpdateListener)
		err = protectionServiceImpl.ConfigureResourceProtection(request)
		resourceProtectionUpdateListener.AssertNotCalled(t, "OnStateChange", request.AppId, request.EnvId, request.ProtectionState, request.UserId)
		assert.Error(t, err, customErr.Error())
	})

	t.Run("get resource protection metadata", func(t *testing.T) {
		resourceProtectionRepository, protectionServiceImpl := getServices(t, sugardLogger)
		appId := 1
		resourceProtectionDtos := getTestResourceProtectionDto()
		resourceProtectionRepository.On("GetResourceProtectMetadata", appId).Return(resourceProtectionDtos, nil)
		protectMetadata, err := protectionServiceImpl.GetResourceProtectMetadata(appId)
		assert.NoError(t, err)
		verifyProtectMetadataWithDtos(t, resourceProtectionDtos, protectMetadata)
	})

	t.Run("get resource protection metadata with error", func(t *testing.T) {
		resourceProtectionRepository, protectionServiceImpl := getServices(t, sugardLogger)
		appId := 1
		customErr := errors.New("failed to fetch")
		resourceProtectionRepository.On("GetResourceProtectMetadata", appId).Return(nil, customErr)
		protectMetadata, err := protectionServiceImpl.GetResourceProtectMetadata(appId)
		assert.Nil(t, protectMetadata)
		assert.Error(t, err, customErr.Error())
	})
}

func verifyProtectMetadataWithDtos(t *testing.T, resourceProtectionDtos []*protect.ResourceProtectionDto, resourceProtectModels []*protect.ResourceProtectModel) {
	assert.Equal(t, len(resourceProtectModels), len(resourceProtectionDtos))
	for index, resourceProtectionDto := range resourceProtectionDtos {
		resourceProtectModel := resourceProtectModels[index]
		assert.Equal(t, resourceProtectModel.AppId, resourceProtectionDto.AppId)
		assert.Equal(t, resourceProtectModel.EnvId, resourceProtectionDto.EnvId)
		assert.Equal(t, resourceProtectModel.ProtectionState, resourceProtectionDto.State)
		assert.Equal(t, resourceProtectModel.UserId, int32(0))
	}
}

func getTestResourceProtectionDto() []*protect.ResourceProtectionDto {
	var protectionDtos []*protect.ResourceProtectionDto
	protectionDtos = append(protectionDtos, &protect.ResourceProtectionDto{AppId: 1, EnvId: 1, State: protect.EnabledProtectionState, AuditLog: sql.AuditLog{CreatedBy: 1}})
	protectionDtos = append(protectionDtos, &protect.ResourceProtectionDto{AppId: 2, EnvId: 1, State: protect.DisabledProtectionState, AuditLog: sql.AuditLog{CreatedBy: 2}})
	return protectionDtos
}

func getResourceProtectModel(state protect.ProtectionState) *protect.ResourceProtectModel {
	request := &protect.ResourceProtectModel{EnvId: 1, AppId: 1, ProtectionState: state, UserId: 1}
	return request
}

func getServices(t *testing.T, sugardLogger *zap.SugaredLogger) (*mocks.ResourceProtectionRepository, *protect.ResourceProtectionServiceImpl) {
	resourceProtectionRepository := mocks.NewResourceProtectionRepository(t)
	protectionServiceImpl := protect.NewResourceProtectionServiceImpl(sugardLogger, resourceProtectionRepository)
	return resourceProtectionRepository, protectionServiceImpl
}
