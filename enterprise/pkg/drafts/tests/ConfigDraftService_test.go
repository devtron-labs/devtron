package tests

import (
	"context"
	"encoding/json"
	"errors"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts"
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts/mocks"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	mocks4 "github.com/devtron-labs/devtron/enterprise/pkg/protect/mocks"
	mocks6 "github.com/devtron-labs/devtron/internal/sql/repository/app/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chart"
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
		configDraftRepository, configDraftServiceImpl, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId + 1}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.LastVersionOutdated)
	})
	t.Run("approval request for draft in terminal state", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftVersionUserId := userId + 1
		appId := 1
		envId := protect.BASE_CONFIG_ENV_ID
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CMDraftResource}
		sampleCMData, sampleCm := getSampleCMCS(appId, envId, 0, "cm2", true)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftVersionUserId := userId + 1
		appId := 1
		envId := 1
		cmId := 1
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CMDraftResource}
		sampleCMData, sampleCm := getSampleCMCS(appId, envId, cmId, "cm2", true)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 4
		envId := protect.BASE_CONFIG_ENV_ID
		cmId := 6
		resourceName := "cm3"
		_, sampleCm := getSampleCMCS(appId, envId, cmId, resourceName, true)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := 6
		cmId := 7
		resourceName := "cm3"
		_, sampleCm := getSampleCMCS(appId, envId, cmId, resourceName, true)
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

	t.Run("approval request for CS draft with add action at base level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := protect.BASE_CONFIG_ENV_ID
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CSDraftResource}
		name := "random-secret"
		sampleCMData, sampleCm := getSampleCMCS(appId, envId, 0, name, false)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.AddResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CSGlobalAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest")).
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
		configMapService.AssertCalled(t, "CSGlobalAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest"))
	})

	t.Run("approval request for CS draft with update action at Env level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := 6
		resourceId := 7
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CSDraftResource}
		sampleCMData, sampleCm := getSampleCMCS(appId, envId, resourceId, "cs2", false)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.UpdateResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CSEnvironmentAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest")).
			Return(func(configMapRequest *bean.ConfigDataRequest) *bean.ConfigDataRequest {
				assert.NotNil(t, configMapRequest)
				assert.Equal(t, resourceId, configMapRequest.Id)
				assert.Equal(t, sampleCMData.AppId, configMapRequest.AppId)
				assert.Equal(t, sampleCMData.EnvironmentId, configMapRequest.EnvironmentId)
				assert.Equal(t, draftVersionUserId, configMapRequest.UserId)
				assert.Equal(t, sampleCMData.ConfigData, configMapRequest.ConfigData)
				return configMapRequest
			}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
		configMapService.AssertCalled(t, "CSEnvironmentAddUpdate", mock.AnythingOfType("*bean.ConfigDataRequest"))
	})

	t.Run("approval request for CS draft with delete action at Base level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 4
		envId := protect.BASE_CONFIG_ENV_ID
		csId := 6
		resourceName := "cs3"
		_, sampleCm := getSampleCMCS(appId, envId, csId, resourceName, false)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CSDraftResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.DeleteResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CSGlobalDelete", resourceName, csId, draftVersionUserId).Return(true, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
	})

	t.Run("approval request for CS draft with delete action at env level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, configMapService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := 6
		cmId := 7
		resourceName := "cs3"
		_, sampleCm := getSampleCMCS(appId, envId, cmId, resourceName, false)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.CSDraftResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleCm, Action: drafts.DeleteResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		configMapService.On("CSEnvironmentDelete", resourceName, cmId, draftVersionUserId).Return(true, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
	})

	t.Run("approval request for deployment template with UPDATE action at BASE level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := protect.BASE_CONFIG_ENV_ID
		resourceName := "Base-DT"
		sampleDeploymentTemplate, templateRequest := getSampleDT(t)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleDeploymentTemplate, Action: drafts.UpdateResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		ctx := context.Background()
		chartService.On("DeploymentTemplateValidate", ctx, templateRequest.ValuesOverride, templateRequest.ChartRefId).Return(true, nil)
		chartService.On("UpdateAppOverride", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, template *chart.TemplateRequest) *chart.TemplateRequest {
				assert.NotNil(t, template)
				assert.Equal(t, draftVersionUserId, template.UserId)
				return template
			}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
	})

	t.Run("approval request for deployment template with UPDATE action at BASE level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := protect.BASE_CONFIG_ENV_ID
		resourceName := "Base-DT"
		sampleDeploymentTemplate, templateRequest := getSampleDT(t)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleDeploymentTemplate, Action: drafts.UpdateResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		ctx := context.Background()
		chartService.On("DeploymentTemplateValidate", ctx, templateRequest.ValuesOverride, templateRequest.ChartRefId).Return(true, nil)
		chartService.On("UpdateAppOverride", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, template *chart.TemplateRequest) *chart.TemplateRequest {
				assert.NotNil(t, template)
				assert.Equal(t, draftVersionUserId, template.UserId)
				return template
			}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NoError(t, err)
	})

	t.Run("approval request for deployment template with UPDATE action at BASE level with invalid template", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := protect.BASE_CONFIG_ENV_ID
		resourceName := "Base-DT"
		sampleDeploymentTemplate, templateRequest := getSampleDT(t)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: sampleDeploymentTemplate, Action: drafts.UpdateResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		ctx := context.Background()
		chartService.On("DeploymentTemplateValidate", ctx, templateRequest.ValuesOverride, templateRequest.ChartRefId).Return(false, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.TemplateOutdated)
	})

	t.Run("approval request for deployment template with ADD action at ENV Level with outdated template", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := 6
		resourceName := "Base-DT"
		envPropsJson, environmentProperties := getEnvDT(t)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: envPropsJson, Action: drafts.AddResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		ctx := context.Background()
		chartService.On("DeploymentTemplateValidate", ctx, environmentProperties.EnvOverrideValues, environmentProperties.ChartRefId).Return(false, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.TemplateOutdated)
	})

	t.Run("approval request for deployment template with ADD action at ENV Level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, propertiesConfigService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appIdVal := 5
		envId := 6
		resourceName := "Base-DT"
		envPropsJson, environmentProperties := getEnvDT(t)
		draftDto := &drafts.DraftDto{AppId: appIdVal, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: envPropsJson, Action: drafts.AddResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		ctx := context.Background()
		chartService.On("DeploymentTemplateValidate", ctx, environmentProperties.EnvOverrideValues, environmentProperties.ChartRefId).Return(true, nil)
		propertiesConfigService.On("CreateEnvironmentProperties", appIdVal, mock.Anything).Return(func(appId int, propertiesRequest *bean.EnvironmentProperties) *bean.EnvironmentProperties {
			assert.Equal(t, appIdVal, appId)
			assert.NotNil(t, propertiesRequest)
			return propertiesRequest
		}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Nil(t, err)
	})

	t.Run("approval request for deployment template with ADD action at ENV Level with no-chart exist error", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, propertiesConfigService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appIdVal := 5
		envId := 6
		resourceName := "Base-DT"
		envPropsJson, environmentProperties := getEnvDT(t)
		draftDto := &drafts.DraftDto{AppId: appIdVal, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: envPropsJson, Action: drafts.AddResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		ctx := context.Background()
		times := 0
		chartService.On("DeploymentTemplateValidate", ctx, environmentProperties.EnvOverrideValues, environmentProperties.ChartRefId).Return(true, nil)
		propertiesConfigService.On("CreateEnvironmentProperties", appIdVal, mock.Anything).Return(func(appId int, propertiesRequest *bean.EnvironmentProperties) *bean.EnvironmentProperties {
			assert.Equal(t, appIdVal, appId)
			assert.NotNil(t, propertiesRequest)
			return propertiesRequest
		}, func(appId int, propertiesRequest *bean.EnvironmentProperties) error {
			var err2 error
			if times == 0 {
				err2 = errors.New(bean2.NOCHARTEXIST)
				times++
			}
			return err2
		})
		chartService.On("CreateChartFromEnvOverride", mock.Anything, mock.Anything).Return(func(templateRequest chart.TemplateRequest, ctx context.Context) (chart *chart.TemplateRequest) {
			return &templateRequest
		}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Nil(t, err)
	})

	t.Run("approval request for deployment template with UPDATE action at ENV Level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, propertiesConfigService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appIdVal := 5
		envId := 6
		resourceName := "Base-DT"
		envPropsJson, environmentProperties := getEnvDT(t)
		draftDto := &drafts.DraftDto{AppId: appIdVal, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: envPropsJson, Action: drafts.UpdateResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		ctx := context.Background()
		chartService.On("DeploymentTemplateValidate", ctx, environmentProperties.EnvOverrideValues, environmentProperties.ChartRefId).Return(true, nil)
		propertiesConfigService.On("UpdateEnvironmentProperties", appIdVal, mock.Anything, draftVersionUserId).Return(func(appId int, propertiesRequest *bean.EnvironmentProperties, userId int32) *bean.EnvironmentProperties {
			assert.NotNil(t, propertiesRequest)
			assert.Equal(t, envId, propertiesRequest.EnvironmentId)
			return propertiesRequest
		}, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Nil(t, err)
	})

	t.Run("approval request for deployment template with DELETE action at ENV Level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, propertiesConfigService := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appIdVal := 5
		envId := 6
		resourceName := "Base-DT"
		//envPropsJson, environmentProperties := getEnvDT(t)
		draftDto := &drafts.DraftDto{AppId: appIdVal, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: "{\"id\":1}", Action: drafts.DeleteResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		propertiesConfigService.On("ResetEnvironmentProperties", 1).Return(true, nil)
		configDraftRepository.On("UpdateDraftState", draftId, drafts.PublishedDraftState, userId).Return(nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Nil(t, err)
	})
}

func getEnvDT(t *testing.T) (string, *bean.EnvironmentProperties) {
	envPropsJson := "{\"environmentId\":1,\"envOverrideValues\":{\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":null,\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"\\/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"\\/bin\\/sh\",\"-c\",\"touch \\/tmp\\/healthy; sleep 30; rm -rf \\/tmp\\/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"\\/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io\\/devtron\\/envoy:v1.14.1\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"51m\",\"memory\":\"51Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"flaggerCanary\":{\"addOtherGateways\":[],\"addOtherHosts\":[],\"analysis\":{\"interval\":\"15s\",\"maxWeight\":50,\"stepWeight\":5,\"threshold\":5},\"annotations\":{},\"appProtocol\":\"http\",\"corsPolicy\":null,\"createIstioGateway\":{\"annotations\":{},\"enabled\":false,\"host\":null,\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":null}},\"enabled\":false,\"gatewayRefs\":null,\"headers\":null,\"labels\":{},\"loadtest\":{\"enabled\":true,\"url\":\"http:\\/\\/flagger-loadtester.istio-system\\/\"},\"match\":[{\"uri\":{\"prefix\":\"\\/\"}}],\"portDiscovery\":true,\"retries\":null,\"rewriteUri\":\"\\/\",\"serviceport\":8080,\"targetPort\":8080,\"thresholds\":{\"latency\":500,\"successRate\":90},\"timeout\":null},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/example2\",\"\\/example2\\/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"example-secret\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":null,\"headers\":null,\"match\":[{\"uri\":{\"prefix\":\"\\/v1\"}},{\"uri\":{\"prefix\":\"\\/v2\"}}],\"retries\":{\"attempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"\\/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseForSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"service\":{\"annotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30},\"chartRefId\":28,\"IsOverride\":true,\"isAppMetricsEnabled\":false,\"currentViewEditor\":\"BASIC\",\"isBasicViewLocked\":false,\"id\":28,\"status\":1,\"manualReviewed\":true,\"active\":true,\"namespace\":\"devtron-demo\"}"
	envConfigProperties := &bean.EnvironmentProperties{}
	err := json.Unmarshal([]byte(envPropsJson), envConfigProperties)
	assert.Nil(t, err)
	return envPropsJson, envConfigProperties
}

func getSampleDT(t *testing.T) (string, *chart.TemplateRequest) {
	dtJson := "{\"id\":2,\"refChartTemplate\":\"reference-chart_4-17-0\",\"refChartTemplateVersion\":\"4.17.0\",\"chartRefId\":29,\"appId\":2,\"valuesOverride\":{\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":null,\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"\\/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"\\/bin\\/sh\",\"-c\",\"touch \\/tmp\\/healthy; sleep 30; rm -rf \\/tmp\\/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"\\/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io\\/devtron\\/envoy:v1.14.1\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{\"nginx.ingress.kubernetes.io\\/ssl-redirect\":\"false\",\"nginx.ingress.kubernetes.io\\/force-ssl-redirect\":\"true\"},\"className\":\"nginx\",\"enabled\":false,\"hosts\":[{\"host\":\"qa.devtron.info\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/orchestrator\"]},{\"host\":\"qa.devtron.info\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/dashboard\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/example2\",\"\\/example2\\/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"secret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"\\/v1\"}},{\"uri\":{\"prefix\":\"\\/v2\"}}],\"retries\":{\"attempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"\\/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseForSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podLabels\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[{\"apiVersion\":\"networking.k8s.io\\/v1\",\"kind\":\"Ingress\",\"metadata\":{\"annotations\":{\"kubernetes.io\\/ingress.class\":\"nginx\"},\"name\":\"argocd-server-ingress\",\"namespace\":\"devtroncd\"},\"spec\":{\"rules\":[{\"host\":\"argocd-qa.devtron.info\",\"http\":{\"paths\":[{\"backend\":{\"service\":{\"name\":\"argocd-server\",\"port\":{\"name\":\"https\"}}},\"path\":\"\\/\",\"pathType\":\"Prefix\"}]}}]}}],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"2\",\"memory\":\"1500Mi\"},\"requests\":{\"cpu\":\"2\",\"memory\":\"1500Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"containerSecurityContext\":{\"allowPrivilegeEscalation\":false,\"runAsUser\":1000,\"runAsNonRoot\":true},\"service\":{\"annotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"NodePort\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"serviceAccountName\":\"devtron\",\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30,\"podSecurityContext\":{\"fsGroup\":1000,\"runAsGroup\":1000,\"runAsUser\":1000}},\"defaultAppOverride\":{\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":null,\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"\\/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"\\/bin\\/sh\",\"-c\",\"touch \\/tmp\\/healthy; sleep 30; rm -rf \\/tmp\\/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"\\/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io\\/devtron\\/envoy:v1.14.1\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{\"nginx.ingress.kubernetes.io\\/ssl-redirect\":\"false\",\"nginx.ingress.kubernetes.io\\/force-ssl-redirect\":\"true\"},\"className\":\"nginx\",\"enabled\":false,\"hosts\":[{\"host\":\"qa.devtron.info\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/orchestrator\"]},{\"host\":\"qa.devtron.info\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/dashboard\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"\\/example2\",\"\\/example2\\/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"secret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"\\/v1\"}},{\"uri\":{\"prefix\":\"\\/v2\"}}],\"retries\":{\"attempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"\\/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseForSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podLabels\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[{\"apiVersion\":\"networking.k8s.io\\/v1\",\"kind\":\"Ingress\",\"metadata\":{\"annotations\":{\"kubernetes.io\\/ingress.class\":\"nginx\"},\"name\":\"argocd-server-ingress\",\"namespace\":\"devtroncd\"},\"spec\":{\"rules\":[{\"host\":\"argocd-qa.devtron.info\",\"http\":{\"paths\":[{\"backend\":{\"service\":{\"name\":\"argocd-server\",\"port\":{\"name\":\"https\"}}},\"path\":\"\\/\",\"pathType\":\"Prefix\"}]}}]}}],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"2\",\"memory\":\"1500Mi\"},\"requests\":{\"cpu\":\"2\",\"memory\":\"1500Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"containerSecurityContext\":{\"allowPrivilegeEscalation\":false,\"runAsUser\":1000,\"runAsNonRoot\":true},\"service\":{\"annotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"NodePort\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"serviceAccountName\":\"devtron\",\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30,\"podSecurityContext\":{\"fsGroup\":1000,\"runAsGroup\":1000,\"runAsUser\":1000}},\"isAppMetricsEnabled\":false,\"isBasicViewLocked\":false,\"currentViewEditor\":\"ADVANCED\"}"
	templateRequest := &chart.TemplateRequest{}
	err := json.Unmarshal([]byte(dtJson), templateRequest)
	assert.Nil(t, err)
	return dtJson, templateRequest
}

func getSampleCMCS(appId, envId, resourceId int, name string, isCM bool) (bean.ConfigDataRequest, string) {
	//{"appId":2,"configData":[{"name":"random-secret","type":"volume","external":false,"roleARN":"","externalType":"","data":{"k1":"djI="},"mountPath":"/mount","subPath":false}]}
	var cmData []byte
	if isCM {
		dataMap := make(map[string]string)
		dataMap["k1"] = "v1"
		cmData, _ = json.Marshal(dataMap)
	} else {
		dataMap := make(map[string][]byte)
		dataMap["k1"] = []byte("v1")
		cmData, _ = json.Marshal(dataMap)
	}
	configData := &bean.ConfigData{Name: name, Type: "environment", External: false, Data: cmData}
	configDataRequest := bean.ConfigDataRequest{Id: resourceId, AppId: appId, EnvironmentId: envId, UserId: 0, ConfigData: []*bean.ConfigData{configData}}
	configDataRequestJson, _ := json.Marshal(configDataRequest)
	return configDataRequest, string(configDataRequestJson)
}

func getMockedConfigDraftServices(t *testing.T, sugardLogger *zap.SugaredLogger) (*mocks.ConfigDraftRepository, *drafts.ConfigDraftServiceImpl, *mocks2.ConfigMapService, *mocks3.ChartService, *mocks2.PropertiesConfigService) {
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
	return configDraftRepository, configDraftServiceImpl, configMapService, chartService, propertiesConfigService
}