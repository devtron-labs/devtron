package tests

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts"
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts/mocks"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	mocks4 "github.com/devtron-labs/devtron/enterprise/pkg/protect/mocks"
	mocks6 "github.com/devtron-labs/devtron/internal/sql/repository/app/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	mocks3 "github.com/devtron-labs/devtron/pkg/chart/mocks"
	mocks7 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	mocks2 "github.com/devtron-labs/devtron/pkg/pipeline/mocks"
	mocks5 "github.com/devtron-labs/devtron/pkg/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

func TestConfigDraftService(t *testing.T) {
	sugardLogger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	t.Run("approval request with outdated version id", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId + 1}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.LastVersionOutdated)
	})
	t.Run("approval request for draft in terminal state", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId}
		draftDto := &drafts.DraftDto{DraftState: drafts.PublishedDraftState}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.DraftAlreadyInTerminalState)
	})
	t.Run("approval request for draft in init state", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId}
		draftDto := &drafts.DraftDto{DraftState: drafts.InitDraftState}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.ApprovalRequestNotRaised)
	})

	t.Run("approval request for self contributed draft", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: userId}
		draftDto := &drafts.DraftDto{DraftState: drafts.AwaitApprovalDraftState}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.UserContributedToDraft)
	})

	t.Run("approval request for cm draft with add action at base level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftVersionUserId := userId + 1
		appId := 1
		envId := protect.BASE_CONFIG_ENV_ID
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CMDraftResource}
		sampleCMData, sampleCm := getSampleCm(appId, envId, 0, "cm2")
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.AddResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CMGlobalAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest")).
			Return(func(configMapRequest *bean.ConfigDataRequest) *bean.ConfigDataRequest {
				assert.NotNil(t, configMapRequest)
				assert.Equal(t, sampleCMData.AppId, configMapRequest.AppId)
				assert.Equal(t, sampleCMData.EnvironmentId, configMapRequest.EnvironmentId)
				assert.Equal(t, draftVersionUserId, configMapRequest.UserId)
				assert.Equal(t, sampleCMData.ConfigData, configMapRequest.ConfigData)
				return configMapRequest
			}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
		configMapService.AssertCalled(t, "CMGlobalAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest"))
	})

	t.Run("approval request for cm draft with update action at Env level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftVersionUserId := userId + 1
		appId := 1
		envId := 1
		cmId := 1
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CMDraftResource}
		sampleCMData, sampleCm := getSampleCm(appId, envId, cmId, "cm2")
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.UpdateResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CMEnvironmentAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest")).
			Return(func(configMapRequest *bean.ConfigDataRequest) *bean.ConfigDataRequest {
				assert.NotNil(t, configMapRequest)
				assert.Equal(t, cmId, configMapRequest.Id)
				assert.Equal(t, sampleCMData.AppId, configMapRequest.AppId)
				assert.Equal(t, sampleCMData.EnvironmentId, configMapRequest.EnvironmentId)
				assert.Equal(t, draftVersionUserId, configMapRequest.UserId)
				assert.Equal(t, sampleCMData.ConfigData, configMapRequest.ConfigData)
				return configMapRequest
			}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
		configMapService.AssertCalled(t, "CMEnvironmentAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest"))
	})

	t.Run("approval request for cm draft with delete action at Base level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 4
		envId := protect.BASE_CONFIG_ENV_ID
		cmId := 6
		resourceName := "cm3"
		_, sampleCm := getSampleCm(appId, envId, cmId, resourceName)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CMDraftResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.DeleteResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CMGlobalDelete", resourceName, cmId, draftVersionUserId).Return(true, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
	})

	t.Run("approval request for cm draft with delete action at env level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := 6
		cmId := 7
		resourceName := "cm3"
		_, sampleCm := getSampleCm(appId, envId, cmId, resourceName)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CMDraftResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.DeleteResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CMEnvironmentDelete", resourceName, cmId, draftVersionUserId).Return(true, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
	})
}

func getMockedConfigDraftServices(t *testing.T, sugardLogger *zap.SugaredLogger) (*mocks.ConfigDraftRepository, *drafts.ConfigDraftServiceImpl, *mocks2.ConfigMapService) {
	configDraftRepository := mocks.NewConfigDraftRepository(t)
	configMapService := mocks2.NewConfigMapService(t)
	chartService := mocks3.NewChartService(t)
	propertiesConfigService := mocks2.NewPropertiesConfigService(t)
	resourceProtectionService := mocks4.NewResourceProtectionService(t)
	userService := mocks5.NewUserService(t)
	appRepository := mocks6.NewAppRepository(t)
	environmentRepository := mocks7.NewEnvironmentRepository(t)
	resourceProtectionService.On("RegisterListener", mock.AnythingOfType("*drafts.ConfigDraftServiceImpl")).Return()
	configDraftServiceImpl := drafts.NewConfigDraftServiceImpl(sugardLogger, configDraftRepository, configMapService, chartService, propertiesConfigService, resourceProtectionService, userService, appRepository, environmentRepository)
	return configDraftRepository, configDraftServiceImpl, configMapService
}

func getSampleCm(appId, envId, cmId int, name string) (bean.ConfigDataRequest, string) {
	dataMap := make(map[string]string)
	dataMap["k1"] = "v1"
	cmData, _ := json.Marshal(dataMap)
	configData := &bean.ConfigData{Name: name, Type: "environment", External: false, Data: cmData}
	configDataRequest := bean.ConfigDataRequest{Id: cmId, AppId: appId, EnvironmentId: envId, UserId: 0, ConfigData: []*bean.ConfigData{configData}}
	configDataRequestJson, _ := json.Marshal(configDataRequest)
	return configDataRequest, string(configDataRequestJson)
}
