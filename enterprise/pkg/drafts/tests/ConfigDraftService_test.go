package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts"
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts/mocks"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	mocks4 "github.com/devtron-labs/devtron/enterprise/pkg/protect/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	mocks6 "github.com/devtron-labs/devtron/internal/sql/repository/app/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chart"
	mocks3 "github.com/devtron-labs/devtron/pkg/chart/mocks"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	mocks7 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	mocks2 "github.com/devtron-labs/devtron/pkg/pipeline/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	mocks5 "github.com/devtron-labs/devtron/pkg/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestConfigDraftService(t *testing.T) {
	sugardLogger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	t.Run("approval request with outdated version id", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 1
		userId := int32(1)
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId + 1}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, drafts.LastVersionOutdated)
	})
	t.Run("approval request for draft in terminal state", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, configMapService, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, chartService, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, chartService, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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

	t.Run("approval request for deployment template with UPDATE action at BASE level with invalid json template", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := protect.BASE_CONFIG_ENV_ID
		resourceName := "Base-DT"
		//sampleDeploymentTemplate, templateRequest := getSampleDT(t)
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		invalidJson := "{\"}"
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: invalidJson, Action: drafts.UpdateResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NotNil(t, err)
	})

	t.Run("approval request for deployment template with UPDATE action at BASE level with invalid template", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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

	t.Run("approval request for deployment template with UPDATE action at BASE level with error occurred during template validation", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		validationErrorMsg := "error during validating template"
		chartService.On("DeploymentTemplateValidate", ctx, templateRequest.ValuesOverride, templateRequest.ChartRefId).Return(false, errors.New(validationErrorMsg))
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, validationErrorMsg)
	})

	t.Run("approval request for deployment template with ADD action at ENV Level with outdated template", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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

	t.Run("approval request for deployment template with ADD action at ENV Level with error occurred during template validation", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		validateErrMsg := "error during template validation"
		chartService.On("DeploymentTemplateValidate", ctx, environmentProperties.EnvOverrideValues, environmentProperties.ChartRefId).Return(false, errors.New(validateErrMsg))
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, validateErrMsg)
	})

	t.Run("approval request for deployment template with ADD action at ENV Level with invalid json", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		draftVersionUserId := userId + 1
		appId := 5
		envId := 6
		resourceName := "Base-DT"
		draftDto := &drafts.DraftDto{AppId: appId, EnvId: envId, ResourceName: resourceName, DraftState: drafts.AwaitApprovalDraftState, Resource: drafts.DeploymentTemplateResource}
		invalidJson := "{\"}"
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId, UserId: draftVersionUserId, Data: invalidJson, Action: drafts.AddResourceAction}
		draftLatestVersion.Draft = draftDto
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{draftLatestVersion}, nil)
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.NotNil(t, err)
	})

	t.Run("approval request for deployment template with ADD action at ENV Level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, chartService, propertiesConfigService, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, chartService, propertiesConfigService, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, chartService, propertiesConfigService, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		configDraftRepository, configDraftServiceImpl, _, _, propertiesConfigService, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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

	t.Run("approval request for deployment template with error occurred during DELETE action at ENV Level", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, propertiesConfigService, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
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
		errMsg := "failed to reset props"
		propertiesConfigService.On("ResetEnvironmentProperties", 1).Return(false, errors.New(errMsg))
		err = configDraftServiceImpl.ApproveDraft(draftId, draftVersionId, userId)
		assert.Error(t, err, errMsg)
	})

	t.Run("add draft version with outdated last version id", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		lastDraftVersionId := 2
		request := drafts.ConfigDraftVersionRequest{
			DraftId:            draftId,
			LastDraftVersionId: lastDraftVersionId,
		}
		configDraftRepository.On("GetLatestDraftVersionId", draftId).Return(3, nil)
		draftVersionId, err := configDraftServiceImpl.AddDraftVersion(request)
		assert.Zero(t, draftVersionId)
		assert.Error(t, err, drafts.LastVersionOutdated)
	})

	t.Run("add draft version with data only", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		lastDraftVersionId := 2
		toBeVersionId := lastDraftVersionId + 1
		userId := int32(4)
		request := drafts.ConfigDraftVersionRequest{
			DraftId:            draftId,
			LastDraftVersionId: lastDraftVersionId,
			Data:               "random-data",
			UserId:             userId,
		}
		configDraftRepository.On("GetLatestDraftVersionId", draftId).Return(lastDraftVersionId, nil)
		configDraftRepository.On("SaveDraftVersion", mock.AnythingOfType("*drafts.DraftVersion")).Return(func(draftVersionDto *drafts.DraftVersion) int {
			assert.NotNil(t, draftVersionDto)
			assert.Equal(t, draftId, draftVersionDto.DraftsId)
			assert.Equal(t, userId, draftVersionDto.UserId)
			assert.Equal(t, request.Data, draftVersionDto.Data)
			return toBeVersionId
		}, nil)
		latestDraftVersionId, err := configDraftServiceImpl.AddDraftVersion(request)
		assert.NoError(t, err)
		assert.Equal(t, toBeVersionId, latestDraftVersionId)
	})

	t.Run("add draft version with comment only and also propose changes", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		lastDraftVersionId := 2
		userId := int32(4)
		request := drafts.ConfigDraftVersionRequest{
			DraftId:            draftId,
			LastDraftVersionId: lastDraftVersionId,
			Data:               "",
			UserComment:        "user-comment",
			UserId:             userId,
			ChangeProposed:     true,
		}
		configDraftRepository.On("GetLatestDraftVersionId", draftId).Return(lastDraftVersionId, nil)
		configDraftRepository.On("SaveDraftVersionComment", mock.AnythingOfType("*drafts.DraftVersionComment")).Return(func(versionComment *drafts.DraftVersionComment) error {
			assert.NotNil(t, versionComment)
			assert.Equal(t, lastDraftVersionId, versionComment.DraftVersionId)
			assert.Equal(t, draftId, versionComment.DraftId)
			assert.Equal(t, draftId, versionComment.DraftId)
			assert.Equal(t, request.UserComment, versionComment.Comment)
			assert.True(t, versionComment.Active)
			assert.Equal(t, userId, versionComment.CreatedBy)
			assert.Equal(t, userId, versionComment.UpdatedBy)
			return nil
		})
		configDraftRepository.On("UpdateDraftState", draftId, drafts.AwaitApprovalDraftState, userId).Return(nil)
		latestDraftVersionId, err := configDraftServiceImpl.AddDraftVersion(request)
		assert.NoError(t, err)
		assert.Equal(t, lastDraftVersionId, latestDraftVersionId)
	})

	t.Run("get draft comments", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, userService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftComments, userIds := getSampleComments(draftId)
		userInfos := getSampleUserInfos(userIds)
		configDraftRepository.On("GetDraftVersionComments", draftId).Return(draftComments, nil)
		userService.On("GetByIds", userIds).Return(userInfos, nil)
		commentsResponse, err := configDraftServiceImpl.GetDraftComments(draftId)
		assert.NoError(t, err)
		assert.Equal(t, draftId, commentsResponse.DraftId)
		verifyUserComments(t, draftComments, commentsResponse.DraftVersionComments, userInfos)
	})

	t.Run("get draft version metadata", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, userService, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionDtos, userIds := getSampleDraftVersionMetadata(draftId)
		userInfos := getSampleUserInfos(userIds)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return(draftVersionDtos, nil)
		userService.On("GetByIds", userIds).Return(userInfos, nil)
		metadataResponse, err := configDraftServiceImpl.GetDraftVersionMetadata(draftId)
		assert.NoError(t, err)
		assert.Equal(t, draftId, metadataResponse.DraftId)
		verifyVersionMetadata(t, draftVersionDtos, metadataResponse.DraftVersions, userInfos)
	})

	t.Run("get draft version metadata with error cases", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		//draftVersionDtos, userIds := getSampleDraftVersionMetadata(draftId)
		//userInfos := getSampleUserInfos(userIds)
		errorMsg := "error from db"
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return([]*drafts.DraftVersion{}, errors.New(errorMsg))
		//userService.On("GetByIds", userIds).Return(userInfos, nil)
		metadataResponse, err := configDraftServiceImpl.GetDraftVersionMetadata(draftId)
		assert.Error(t, err, errorMsg)
		assert.Nil(t, metadataResponse)
		//verifyVersionMetadata(t, draftVersionDtos, metadataResponse.DraftVersions, userInfos)
	})

	t.Run("get draft by id", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, userService, appRepo, envRepo := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		appId := 2
		envId := 3
		draftVersionId := 4
		mockedDraftMetadataArray, userIds := getSampleDraftVersionMetadata(draftId)
		userId := userIds[0]
		resourceType := drafts.CMDraftResource
		draftState := drafts.AwaitApprovalDraftState
		draftVersionData := "sample-cm"
		draftVersion := &drafts.DraftVersion{
			Draft: &drafts.DraftDto{
				Id:           draftId,
				AppId:        appId,
				EnvId:        envId,
				ResourceName: "resource-name",
				Resource:     resourceType,
				DraftState:   draftState,
			},
			Id:       draftVersionId,
			DraftsId: draftId,
			Action:   drafts.AddResourceAction,
			Data:     draftVersionData,
			UserId:   userId,
		}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftVersion, nil)
		appName := "appName"
		appRepo.On("FindById", appId).Return(&app.App{Id: appId, AppName: appName}, nil)
		environmentIdentifier := "env-identifier"
		envRepo.On("FindById", envId).Return(&repository.Environment{Id: envId, EnvironmentIdentifier: environmentIdentifier}, nil)
		sampleEmailIds := getSampleEmailIds()
		userService.On("GetConfigApprovalUsersByEnv", appName, environmentIdentifier).Return(sampleEmailIds, nil)
		configDraftRepository.On("GetDraftVersionsMetadata", draftId).Return(mockedDraftMetadataArray, nil)
		draftResponse, err := configDraftServiceImpl.GetDraftById(draftId, userId)
		assert.NoError(t, err)
		assert.Equal(t, draftId, draftResponse.DraftId)
		assert.Equal(t, appId, draftResponse.AppId)
		assert.Equal(t, envId, draftResponse.EnvId)
		assert.Equal(t, resourceType, draftResponse.Resource)
		assert.Equal(t, draftState, draftResponse.DraftState)
		assert.Equal(t, draftVersionId, draftResponse.DraftVersionId)
		assert.Equal(t, draftVersionData, draftResponse.Data)
		assert.False(t, *draftResponse.CanApprove)
		assert.Equal(t, sampleEmailIds, draftResponse.Approvers)
	})

	t.Run("get draft count", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		appId := 1
		envIds := []int{2, 3}
		draftDtos, draftCount := getSampleDraftDtos(appId, envIds)
		configDraftRepository.On("GetDraftMetadataForAppAndEnv", appId, envIds).Return(draftDtos, nil)
		draftsCountResponse, err := configDraftServiceImpl.GetDraftsCount(appId, envIds)
		assert.NoError(t, err)
		assert.Equal(t, len(envIds), len(draftsCountResponse))
		for _, draftCountResponse := range draftsCountResponse {
			assert.Equal(t, appId, draftCountResponse.AppId)
			assert.Equal(t, draftCount, draftCountResponse.DraftsCount)
		}
	})

	t.Run("update draft state", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		toUpdateDraftState := drafts.AwaitApprovalDraftState
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId}
		draftDto := &drafts.DraftDto{DraftState: drafts.InitDraftState}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		configDraftRepository.On("UpdateDraftState", draftId, toUpdateDraftState, userId).Return(nil)
		latestDraftVersion, err := configDraftServiceImpl.UpdateDraftState(draftId, draftVersionId, toUpdateDraftState, userId)
		assert.NoError(t, err)
		assert.Equal(t, draftLatestVersion, latestDraftVersion)
	})

	t.Run("error handling cases while updating draft state", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		toUpdateDraftState := drafts.AwaitApprovalDraftState
		//draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId}
		//draftDto := &drafts.DraftDto{DraftState: drafts.InitDraftState}
		errMsg := "no data found"
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(nil, errors.New(errMsg))
		//configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		//configDraftRepository.On("UpdateDraftState", draftId, toUpdateDraftState, userId).Return(nil)
		latestDraftVersion, err := configDraftServiceImpl.UpdateDraftState(draftId, draftVersionId, toUpdateDraftState, userId)
		assert.Error(t, err, errMsg)
		assert.Nil(t, nil, latestDraftVersion)
	})

	t.Run("error handling cases while updating draft state", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftVersionId := 2
		userId := int32(3)
		toUpdateDraftState := drafts.AwaitApprovalDraftState
		draftLatestVersion := &drafts.DraftVersion{Id: draftVersionId}
		draftDto := &drafts.DraftDto{DraftState: drafts.InitDraftState}
		configDraftRepository.On("GetLatestConfigDraft", draftId).Return(draftLatestVersion, nil)
		configDraftRepository.On("GetDraftMetadataById", draftId).Return(draftDto, nil)
		errMsg := "error from db"
		configDraftRepository.On("UpdateDraftState", draftId, toUpdateDraftState, userId).Return(errors.New(errMsg))
		latestDraftVersion, err := configDraftServiceImpl.UpdateDraftState(draftId, draftVersionId, toUpdateDraftState, userId)
		assert.Error(t, err, errMsg)
		assert.Equal(t, draftLatestVersion, latestDraftVersion)
	})

	t.Run("get Drafts of particular resource", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		appId := 1
		envId := 2
		userId := int32(3)
		resourceType := drafts.CMDraftResource
		mockedDraftDtos := getSampleDraftDtosForResource(appId, envId, resourceType)
		configDraftRepository.On("GetDraftMetadata", appId, envId, resourceType).Return(mockedDraftDtos, nil)
		appConfigDrafts, err := configDraftServiceImpl.GetDrafts(appId, envId, resourceType, userId)
		assert.NoError(t, err)
		verifyAppConfigDrafts(t, mockedDraftDtos, appConfigDrafts)
	})

	t.Run("delete comment with failure case", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftCommentId := 2
		userId := int32(3)
		configDraftRepository.On("DeleteComment", draftId, draftCommentId, userId).Return(0, nil)
		err := configDraftServiceImpl.DeleteComment(draftId, draftCommentId, userId)
		assert.Error(t, err, drafts.FailedToDeleteComment)
	})

	t.Run("delete comment", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		draftId := 1
		draftCommentId := 2
		userId := int32(3)
		configDraftRepository.On("DeleteComment", draftId, draftCommentId, userId).Return(1, nil)
		err := configDraftServiceImpl.DeleteComment(draftId, draftCommentId, userId)
		assert.NoError(t, err)
	})

	t.Run("config draft state change", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		appId := 1
		envId := 2
		protectionState := protect.DisabledProtectionState
		userId := int32(3)
		configDraftRepository.On("DiscardDrafts", appId, envId, userId).Return(nil)
		configDraftServiceImpl.OnStateChange(appId, envId, protectionState, userId)
		configDraftRepository.AssertCalled(t, "DiscardDrafts", appId, envId, userId)
	})

	t.Run("create draft cases", func(t *testing.T) {
		configDraftRepository, configDraftServiceImpl, _, _, _, _, _, _ := getMockedConfigDraftServices(t, sugardLogger)
		mockedRequest := drafts.ConfigDraftRequest{
			AppId:          1,
			EnvId:          2,
			Resource:       drafts.CSDraftResource,
			ResourceName:   "resource-name",
			Action:         drafts.AddResourceAction,
			Data:           "draft-data",
			UserComment:    "userComment",
			ChangeProposed: true,
			UserId:         3,
		}
		mockedConfigDraftResponse := &drafts.ConfigDraftResponse{}
		mockedConfigDraftResponse.ConfigDraftRequest = mockedRequest
		configDraftRepository.On("CreateConfigDraft", mock.AnythingOfType("drafts.ConfigDraftRequest")).Return(func(request drafts.ConfigDraftRequest) *drafts.ConfigDraftResponse {
			assert.Equal(t, mockedRequest.AppId, request.AppId)
			assert.Equal(t, mockedRequest.EnvId, request.EnvId)
			assert.Equal(t, mockedRequest.Resource, request.Resource)
			assert.Equal(t, mockedRequest.ResourceName, request.ResourceName)
			assert.Equal(t, mockedRequest.Action, request.Action)
			assert.Equal(t, mockedRequest.Data, request.Data)
			return mockedConfigDraftResponse
		}, nil)
		configDraftResponse, err := configDraftServiceImpl.CreateDraft(mockedRequest)
		assert.Nil(t, err)
		assert.Equal(t, mockedConfigDraftResponse, configDraftResponse)
	})
}

func verifyAppConfigDrafts(t *testing.T, mockedDraftDtos []*drafts.DraftDto, appConfigDrafts []drafts.AppConfigDraft) {
	for index, appConfigDraft := range appConfigDrafts {
		mockedDraftDto := mockedDraftDtos[index]
		assert.Equal(t, mockedDraftDto.Id, appConfigDraft.DraftId)
		assert.Equal(t, mockedDraftDto.Resource, appConfigDraft.Resource)
		assert.Equal(t, mockedDraftDto.ResourceName, appConfigDraft.ResourceName)
		assert.Equal(t, mockedDraftDto.DraftState, appConfigDraft.DraftState)
	}
}

func getSampleDraftDtosForResource(appId int, envId int, resourceType drafts.DraftResourceType) []*drafts.DraftDto {
	var draftDtos []*drafts.DraftDto
	sampleDraftDtos, _ := getSampleDraftDtos(appId, []int{envId})
	for _, sampleDraftDto := range sampleDraftDtos {
		if sampleDraftDto.Resource == resourceType {
			draftDtos = append(draftDtos, sampleDraftDto)
		}
	}
	return draftDtos
}

func getSampleDraftDtos(appId int, envIds []int) ([]*drafts.DraftDto, int) {
	var sampleDrafts []*drafts.DraftDto
	for index, envId := range envIds {
		sampleDrafts = append(sampleDrafts, &drafts.DraftDto{
			AppId:        appId,
			EnvId:        envId,
			Resource:     drafts.CMDraftResource,
			ResourceName: fmt.Sprintf("cm-%d-resource-name", index),
			DraftState:   drafts.InitDraftState,
		})
		sampleDrafts = append(sampleDrafts, &drafts.DraftDto{
			AppId:        appId,
			EnvId:        envId,
			Resource:     drafts.CSDraftResource,
			ResourceName: fmt.Sprintf("cs-%d-resource-name", index),
			DraftState:   drafts.AwaitApprovalDraftState,
		})
	}
	return sampleDrafts, 2
}

func getSampleEmailIds() []string {
	var sampleEmailIds []string
	for i := 0; i < 10; i++ {
		sampleEmailIds = append(sampleEmailIds, fmt.Sprintf("%d@gmail.com", i))
	}
	return sampleEmailIds
}

func verifyVersionMetadata(t *testing.T, mockedVersionDtos []*drafts.DraftVersion, responseVersionMetadataArray []*drafts.DraftVersionMetadata, userInfos []bean2.UserInfo) {
	userInfoMap := getUserInfoMap(userInfos)
	mockedVersionDtoMap := make(map[int]*drafts.DraftVersion)
	for _, mockedVersionDto := range mockedVersionDtos {
		mockedVersionDtoMap[mockedVersionDto.Id] = mockedVersionDto
	}
	for _, responseVersionMetadata := range responseVersionMetadataArray {
		draftVersionId := responseVersionMetadata.DraftVersionId
		mockedVersion := mockedVersionDtoMap[draftVersionId]
		userId := mockedVersion.UserId
		assert.Equal(t, userId, responseVersionMetadata.UserId)
		assert.Equal(t, userInfoMap[userId].EmailId, responseVersionMetadata.UserEmail)
		assert.Equal(t, mockedVersion.CreatedOn, responseVersionMetadata.ActivityTime)
	}
}

func getSampleDraftVersionMetadata(draftId int) ([]*drafts.DraftVersion, []int32) {
	var draftVersionMetadataList []*drafts.DraftVersion
	var userIds []int32
	for i := 0; i < 10; i++ {
		userId := int32(i)
		userIds = append(userIds, userId)
		draftVersionMetadata := &drafts.DraftVersion{
			Id:        i,
			UserId:    userId,
			CreatedOn: time.Now(),
			DraftsId:  draftId,
		}
		draftVersionMetadataList = append(draftVersionMetadataList, draftVersionMetadata)
	}
	return draftVersionMetadataList, userIds
}

func verifyUserComments(t *testing.T, mockedComments []*drafts.DraftVersionComment, responseComments []drafts.DraftVersionCommentBean, userInfos []bean2.UserInfo) {
	assert.Equal(t, len(mockedComments), len(responseComments))
	userInfoMap := getUserInfoMap(userInfos)
	commentIdMap := make(map[int]*drafts.DraftVersionComment)
	for _, mockedComment := range mockedComments {
		commentIdMap[mockedComment.Id] = mockedComment
	}
	for _, responseComment := range responseComments {
		userComments := responseComment.UserComments
		for _, userComment := range userComments {
			draftVersionComment := commentIdMap[userComment.CommentId]
			verifyDraftVersionComment(t, draftVersionComment, userComment, userInfoMap[userComment.UserId])
		}
	}
}

func getUserInfoMap(userInfos []bean2.UserInfo) map[int32]bean2.UserInfo {
	userInfoMap := make(map[int32]bean2.UserInfo)
	for _, userInfo := range userInfos {
		userInfoMap[userInfo.Id] = userInfo
	}
	return userInfoMap
}

func verifyDraftVersionComment(t *testing.T, mockedComment *drafts.DraftVersionComment, responseComment drafts.UserCommentMetadata, userInfo bean2.UserInfo) {
	assert.Equal(t, mockedComment.Id, responseComment.CommentId)
	assert.Equal(t, mockedComment.Comment, responseComment.Comment)
	assert.Equal(t, mockedComment.CreatedBy, responseComment.UserId)
	assert.Equal(t, userInfo.EmailId, responseComment.UserEmail)
}

func getSampleUserInfos(userIds []int32) []bean2.UserInfo {
	var userInfos []bean2.UserInfo
	for _, userId := range userIds {
		userInfos = append(userInfos, bean2.UserInfo{Id: userId, EmailId: fmt.Sprintf("%d@gmail.com", userId)})
	}
	return userInfos
}

func getSampleComments(draftId int) ([]*drafts.DraftVersionComment, []int32) {
	var draftComments []*drafts.DraftVersionComment
	var userIds []int32
	for i := 0; i < 10; i++ {
		userId := int32(i)
		draftVersionComment := &drafts.DraftVersionComment{
			Id:             i,
			DraftId:        draftId,
			DraftVersionId: i,
			Comment:        fmt.Sprintf("random-comment-%d", i),
			Active:         true,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				UpdatedBy: userId,
			},
		}
		draftComments = append(draftComments, draftVersionComment)
		userIds = append(userIds, userId)
	}
	return draftComments, userIds
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

func getMockedConfigDraftServices(t *testing.T, sugardLogger *zap.SugaredLogger) (*mocks.ConfigDraftRepository, *drafts.ConfigDraftServiceImpl, *mocks2.ConfigMapService, *mocks3.ChartService, *mocks2.PropertiesConfigService, *mocks5.UserService, *mocks6.AppRepository, *mocks7.EnvironmentRepository) {
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
	return configDraftRepository, configDraftServiceImpl, configMapService, chartService, propertiesConfigService, userService, appRepository, environmentRepository
}