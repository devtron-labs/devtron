/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package appClone

import (
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	appWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type AppCloneService interface {
	CloneApp(createReq *bean.CreateAppDTO, context context.Context) (*bean.CreateAppDTO, error)
}
type AppCloneServiceImpl struct {
	logger                  *zap.SugaredLogger
	pipelineBuilder         pipeline.PipelineBuilder
	materialRepository      pipelineConfig.MaterialRepository
	chartService            chart.ChartService
	configMapService        pipeline.ConfigMapService
	appWorkflowService      appWorkflow.AppWorkflowService
	appListingService       app.AppListingService
	propertiesConfigService pipeline.PropertiesConfigService
	pipelineStageService    pipeline.PipelineStageService
	ciTemplateService       pipeline.CiTemplateService
	appRepository           app2.AppRepository
	ciPipelineRepository    pipelineConfig.CiPipelineRepository
	pipelineRepository      pipelineConfig.PipelineRepository
	appWorkflowRepository   appWorkflow2.AppWorkflowRepository
	ciPipelineConfigService pipeline.CiPipelineConfigService
}

func NewAppCloneServiceImpl(logger *zap.SugaredLogger,
	pipelineBuilder pipeline.PipelineBuilder,
	materialRepository pipelineConfig.MaterialRepository,
	chartService chart.ChartService,
	configMapService pipeline.ConfigMapService,
	appWorkflowService appWorkflow.AppWorkflowService,
	appListingService app.AppListingService,
	propertiesConfigService pipeline.PropertiesConfigService,
	ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository,
	pipelineStageService pipeline.PipelineStageService, ciTemplateService pipeline.CiTemplateService,
	appRepository app2.AppRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository, appWorkflowRepository appWorkflow2.AppWorkflowRepository,
	ciPipelineConfigService pipeline.CiPipelineConfigService) *AppCloneServiceImpl {
	return &AppCloneServiceImpl{
		logger:                  logger,
		pipelineBuilder:         pipelineBuilder,
		materialRepository:      materialRepository,
		chartService:            chartService,
		configMapService:        configMapService,
		appWorkflowService:      appWorkflowService,
		appListingService:       appListingService,
		propertiesConfigService: propertiesConfigService,
		pipelineStageService:    pipelineStageService,
		ciTemplateService:       ciTemplateService,
		appRepository:           appRepository,
		ciPipelineRepository:    ciPipelineRepository,
		pipelineRepository:      pipelineRepository,
		appWorkflowRepository:   appWorkflowRepository,
		ciPipelineConfigService: ciPipelineConfigService,
	}
}

type CloneRequest struct {
	RefAppId    int                            `json:"refAppId"`
	Name        string                         `json:"name"`
	ProjectId   int                            `json:"projectId"`
	AppLabels   []*bean.Label                  `json:"labels,omitempty" validate:"dive"`
	Description *bean2.GenericNoteResponseBean `json:"description"`
	AppType     helper.AppType                 `json:"appType"`
}

type CreateWorkflowMappingDto struct {
	oldAppId             int
	newAppId             int
	userId               int32
	newWfId              int
	gitMaterialMapping   map[int]int
	externalCiPipelineId int
	oldToNewCDPipelineId map[int]int
}

func (impl *AppCloneServiceImpl) CloneApp(createReq *bean.CreateAppDTO, context context.Context) (*bean.CreateAppDTO, error) {
	//validate template app
	templateApp, err := impl.appRepository.FindById(createReq.TemplateId)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	//If the template does not exist then don't clone
	//If the template app-type is chart-store app then don't clone
	//If the template app-type and create request app-type is not same then don't clone
	if (templateApp == nil && templateApp.Id == 0) || (templateApp.AppType == helper.ChartStoreApp) || (templateApp.AppType != createReq.AppType) {
		impl.logger.Warnw("template app does not exist", "id", createReq.TemplateId)
		err = &util.ApiError{
			Code:            constants.AppDoesNotExist.Code,
			InternalMessage: "app does not exist",
			UserMessage:     constants.AppAlreadyExists.UserMessage(createReq.TemplateId),
		}
		return nil, err
	}
	//create new app
	cloneReq := &CloneRequest{
		RefAppId:    createReq.TemplateId,
		Name:        createReq.AppName,
		ProjectId:   createReq.TeamId,
		AppLabels:   createReq.AppLabels,
		AppType:     createReq.AppType,
		Description: createReq.GenericNote,
	}
	userId := createReq.UserId
	appStatus, err := impl.appListingService.FetchAppStageStatus(cloneReq.RefAppId, int(cloneReq.AppType))
	if err != nil {
		return nil, err
	}

	refAppStatus := make(map[string]bool)
	for _, as := range appStatus {
		refAppStatus[as.StageName] = as.Status
	}

	//TODO check stage of current app
	if createReq.AppType != helper.Job {
		if !refAppStatus["APP"] {
			impl.logger.Warnw("status not", "APP", cloneReq.RefAppId)
			return nil, nil
		}
	}
	app, err := impl.CreateApp(cloneReq, userId)
	if err != nil {
		impl.logger.Errorw("error in creating app", "req", cloneReq, "err", err)
		return nil, err
	}
	newAppId := app.Id
	if !refAppStatus["MATERIAL"] {
		impl.logger.Errorw("status not", "MATERIAL", cloneReq.RefAppId)
		return app, nil
	}
	_, gitMaerialMap, err := impl.CloneGitRepo(cloneReq.RefAppId, newAppId, userId)
	if err != nil {
		impl.logger.Errorw("error in cloning git", "ref", cloneReq.RefAppId, "new", newAppId, "err", err)
		return nil, err
	}

	_, err = impl.CreateCiTemplate(cloneReq.RefAppId, newAppId, userId, gitMaerialMap)
	if err != nil {
		impl.logger.Errorw("error in cloning docker template", "ref", cloneReq.RefAppId, "new", newAppId, "err", err)
		return nil, err
	}
	if createReq.AppType != helper.Job {
		if !refAppStatus["TEMPLATE"] {
			impl.logger.Errorw("status not", "TEMPLATE", cloneReq.RefAppId)
			return app, nil
		}
		if !refAppStatus["CHART"] {
			impl.logger.Errorw("status not", "CHART", cloneReq.RefAppId)
			return app, nil
		}
		_, err = impl.CreateDeploymentTemplate(cloneReq.RefAppId, newAppId, userId, context)
		if err != nil {
			impl.logger.Errorw("error in creating deployment template", "ref", cloneReq.RefAppId, "new", newAppId, "err", err)
			return nil, err
		}
	}
	_, err = impl.CreateGlobalCM(cloneReq.RefAppId, newAppId, userId)

	if err != nil {
		impl.logger.Errorw("error in creating global cm", "ref", cloneReq.RefAppId, "new", newAppId, "err", err)
		return nil, err
	}
	_, err = impl.CreateGlobalSecret(cloneReq.RefAppId, newAppId, userId)
	if err != nil {
		impl.logger.Errorw("error in creating global secret", "ref", cloneReq.RefAppId, "new", newAppId, "err", err)
		return nil, err
	}

	if createReq.AppType != helper.Job {
		_, err = impl.CreateEnvCm(context, cloneReq.RefAppId, newAppId, userId)
		if err != nil {
			impl.logger.Errorw("error in creating env cm", "err", err)
			return nil, err
		}
		_, err = impl.CreateEnvSecret(context, cloneReq.RefAppId, newAppId, userId)
		if err != nil {
			impl.logger.Errorw("error in creating env secret", "err", err)
			return nil, err
		}
		_, err = impl.createEnvOverride(cloneReq.RefAppId, newAppId, userId, context)
		if err != nil {
			impl.logger.Errorw("error in cloning  env override", "err", err)
			return nil, err
		}
	} else {
		_, err := impl.configMapService.ConfigSecretEnvironmentClone(cloneReq.RefAppId, newAppId, userId)
		if err != nil {
			impl.logger.Errorw("error in cloning cm cs env override", "err", err)
			return nil, err
		}
	}
	_, err = impl.CreateWf(cloneReq.RefAppId, newAppId, userId, gitMaerialMap, context)
	if err != nil {
		impl.logger.Errorw("error in creating wf", "ref", cloneReq.RefAppId, "new", newAppId, "err", err)
		return nil, err
	}

	return app, nil
}

func (impl *AppCloneServiceImpl) CreateApp(cloneReq *CloneRequest, userId int32) (*bean.CreateAppDTO, error) {
	createAppReq := &bean.CreateAppDTO{
		AppName:     cloneReq.Name,
		UserId:      userId,
		TeamId:      cloneReq.ProjectId,
		AppLabels:   cloneReq.AppLabels,
		AppType:     cloneReq.AppType,
		GenericNote: cloneReq.Description,
	}
	createRes, err := impl.pipelineBuilder.CreateApp(createAppReq)
	return createRes, err
}

func (impl *AppCloneServiceImpl) CloneGitRepo(oldAppId, newAppId int, userId int32) (*bean.CreateMaterialDTO, map[int]int, error) {
	originalApp, err := impl.pipelineBuilder.GetApp(oldAppId)
	if err != nil {
		return nil, nil, err
	}
	createMaterial := &bean.CreateMaterialDTO{
		AppId:  newAppId,
		UserId: userId,
	}
	var savedGitMaterials []*bean.GitMaterial
	gitMaterialsMap := make(map[int]int)
	for _, material := range originalApp.Material {
		gitMaterial := &bean.GitMaterial{
			Name:          material.Name,
			Url:           material.Url,
			Id:            0,
			GitProviderId: material.GitProviderId,
			CheckoutPath:  material.CheckoutPath,
			FilterPattern: material.FilterPattern,
		}
		createMaterial.Material = []*bean.GitMaterial{gitMaterial} // append(createMaterial.Material, gitMaterial)
		createMaterialres, err := impl.pipelineBuilder.CreateMaterialsForApp(createMaterial)
		if err != nil {
			return nil, nil, err
		}
		savedGitMaterials = append(savedGitMaterials, createMaterialres.Material...)
		gitMaterialsMap[material.Id] = createMaterial.Material[0].Id
	}
	createMaterial.Material = savedGitMaterials
	//impl.logger.Infof()
	return createMaterial, gitMaterialsMap, err
}

func (impl *AppCloneServiceImpl) CreateCiTemplate(oldAppId, newAppId int, userId int32, gitMaterialMap map[int]int) (*bean.PipelineCreateResponse, error) {
	refCiConf, err := impl.pipelineBuilder.GetCiPipeline(oldAppId)
	if err != nil {
		return nil, err
	}
	if gitMaterialMap == nil || len(gitMaterialMap) == 0 {
		return nil, fmt.Errorf("no git for %d", newAppId)
	}

	//gitMaterialMap contains the mappings for old app git-material-id -> new app git-material-id
	dockerfileGitMaterial := gitMaterialMap[refCiConf.CiBuildConfig.GitMaterialId]
	buildContextGitMaterial := gitMaterialMap[refCiConf.CiBuildConfig.BuildContextGitMaterialId]
	//this might be possible if build-configuration is not set in the old app.
	if dockerfileGitMaterial == 0 {
		//set the dockerfileGitMaterial to first material in the map
		for _, newAppMaterialId := range gitMaterialMap {
			dockerfileGitMaterial = newAppMaterialId
			break
		}
	}
	//if buildContextGitMaterial not found set to build context repo to dockerfile git repo
	if buildContextGitMaterial == 0 {
		buildContextGitMaterial = dockerfileGitMaterial
	}

	ciBuildConfig := refCiConf.CiBuildConfig
	ciBuildConfig.GitMaterialId = dockerfileGitMaterial
	ciBuildConfig.BuildContextGitMaterialId = buildContextGitMaterial
	ciConfRequest := &bean.CiConfigRequest{
		Id:                0,
		AppId:             newAppId,
		DockerRegistry:    refCiConf.DockerRegistry,
		DockerRepository:  refCiConf.DockerRepository,
		CiBuildConfig:     ciBuildConfig,
		DockerRegistryUrl: refCiConf.DockerRegistry,
		CiTemplateName:    refCiConf.CiTemplateName,
		UserId:            userId,
		BeforeDockerBuild: refCiConf.BeforeDockerBuild,
		AfterDockerBuild:  refCiConf.AfterDockerBuild,
		ScanEnabled:       refCiConf.ScanEnabled,
	}

	res, err := impl.pipelineBuilder.CreateCiPipeline(ciConfRequest)
	return res, err
}

func (impl *AppCloneServiceImpl) CreateDeploymentTemplate(oldAppId, newAppId int, userId int32, context context.Context) (*chart.TemplateRequest, error) {
	refTemplate, err := impl.chartService.FindLatestChartForAppByAppId(oldAppId)
	if err != nil {
		impl.logger.Errorw("error in fetching ref app chart ", "app", oldAppId, "err", err)
		return nil, err
	}
	templateReq := chart.TemplateRequest{
		Id:                0,
		AppId:             newAppId,
		Latest:            refTemplate.Latest,
		ValuesOverride:    refTemplate.DefaultAppOverride,
		ChartRefId:        refTemplate.ChartRefId,
		UserId:            userId,
		IsBasicViewLocked: refTemplate.IsBasicViewLocked,
		CurrentViewEditor: refTemplate.CurrentViewEditor,
	}
	templateRes, err := impl.chartService.Create(templateReq, context)
	if err != nil {
		impl.logger.Errorw("template clone err", "req", templateReq, "err", templateReq)
	}
	return templateRes, err
}

func (impl *AppCloneServiceImpl) CreateAppMetrics(oldAppId, newAppId int, userId int32) {

}

func (impl *AppCloneServiceImpl) CreateGlobalCM(oldAppId, newAppId int, userId int32) (*bean3.ConfigDataRequest, error) {
	refCM, err := impl.configMapService.CMGlobalFetch(oldAppId)
	if err != nil {
		return nil, err
	}
	thisCm, err := impl.configMapService.CMGlobalFetch(newAppId)
	if err != nil {
		return nil, err
	}

	cfgDatas := impl.configDataClone(refCM.ConfigData)
	for _, cfgData := range cfgDatas {
		newCm := &bean3.ConfigDataRequest{
			AppId:         newAppId,
			EnvironmentId: refCM.EnvironmentId,
			ConfigData:    []*bean3.ConfigData{cfgData},
			UserId:        userId,
			Id:            thisCm.Id,
		}
		thisCm, err = impl.configMapService.CMGlobalAddUpdate(newCm)
		if err != nil {
			return nil, err
		}
	}
	return thisCm, err

}

func (impl *AppCloneServiceImpl) CreateEnvCm(ctx context.Context, oldAppId, newAppId int, userId int32) (interface{}, error) {
	refEnvs, err := impl.appListingService.FetchOtherEnvironment(ctx, oldAppId)
	if err != nil {
		return nil, err
	}
	for _, refEnv := range refEnvs {
		impl.logger.Debugw("cloning cfg for env", "env", refEnv)
		refCm, err := impl.configMapService.CMEnvironmentFetch(oldAppId, refEnv.EnvironmentId)
		if err != nil {
			return nil, err
		}
		thisCm, err := impl.configMapService.CMEnvironmentFetch(newAppId, refEnv.EnvironmentId)
		if err != nil {
			return nil, err
		}

		var refEnvCm []*bean3.ConfigData
		for _, refCmData := range refCm.ConfigData {
			if !refCmData.Global || refCmData.Data != nil {
				refEnvCm = append(refEnvCm, refCmData)
			}
		}
		if len(refEnvCm) == 0 {
			impl.logger.Debug("no env cm")
			continue
		}
		cfgDatas := impl.configDataClone(refEnvCm)
		for _, cfgData := range cfgDatas {
			newCm := &bean3.ConfigDataRequest{
				AppId:         newAppId,
				EnvironmentId: refEnv.EnvironmentId,
				ConfigData:    []*bean3.ConfigData{cfgData},
				UserId:        userId,
				Id:            thisCm.Id,
			}
			thisCm, err = impl.configMapService.CMEnvironmentAddUpdate(newCm)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func (impl *AppCloneServiceImpl) CreateEnvSecret(ctx context.Context, oldAppId, newAppId int, userId int32) (interface{}, error) {
	refEnvs, err := impl.appListingService.FetchOtherEnvironment(ctx, oldAppId)
	if err != nil {
		return nil, err
	}
	for _, refEnv := range refEnvs {
		impl.logger.Debugw("cloning cfg for env", "env", refEnv)
		refCm, err := impl.configMapService.CSEnvironmentFetch(oldAppId, refEnv.EnvironmentId)
		if err != nil {
			return nil, err
		}
		thisCm, err := impl.configMapService.CSEnvironmentFetch(newAppId, refEnv.EnvironmentId)
		if err != nil {
			return nil, err
		}

		var refEnvCm []*bean3.ConfigData
		for _, refCmData := range refCm.ConfigData {
			if !refCmData.Global || refCmData.Data != nil {
				refEnvCm = append(refEnvCm, refCmData)
			}
		}
		if len(refEnvCm) == 0 {
			impl.logger.Debug("no env cm")
			continue
		}
		cfgDatas := impl.configDataClone(refEnvCm)
		for _, cfgData := range cfgDatas {
			var configData []*bean3.ConfigData
			configData = append(configData, cfgData)
			newCm := &bean3.ConfigDataRequest{
				AppId:         newAppId,
				EnvironmentId: refEnv.EnvironmentId,
				ConfigData:    configData,
				UserId:        userId,
				Id:            thisCm.Id,
			}
			thisCm, err = impl.configMapService.CSEnvironmentAddUpdate(newCm)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func (impl *AppCloneServiceImpl) createEnvOverride(oldAppId, newAppId int, userId int32, ctx context.Context) (interface{}, error) {
	refEnvs, err := impl.appListingService.FetchOtherEnvironment(ctx, oldAppId)
	if err != nil {
		return nil, err
	}
	for _, refEnv := range refEnvs {

		chartRefRes, err := impl.chartService.ChartRefAutocompleteForAppOrEnv(oldAppId, refEnv.EnvironmentId)
		if err != nil {
			return nil, err
		}
		refEnvProperties, err := impl.propertiesConfigService.GetEnvironmentProperties(oldAppId, refEnv.EnvironmentId, chartRefRes.LatestEnvChartRef)
		if err != nil {
			return nil, err
		}
		if !refEnvProperties.IsOverride {
			impl.logger.Debugw("no env override", "env", refEnv.EnvironmentId)
			continue
		}
		thisEnvProperties, err := impl.propertiesConfigService.GetEnvironmentProperties(newAppId, refEnv.EnvironmentId, chartRefRes.LatestEnvChartRef)
		if err != nil {
			return nil, err
		}
		envPropertiesReq := &bean3.EnvironmentProperties{
			Id:                thisEnvProperties.EnvironmentConfig.Id,
			EnvOverrideValues: refEnvProperties.EnvironmentConfig.EnvOverrideValues,
			Status:            refEnvProperties.EnvironmentConfig.Status,
			ManualReviewed:    refEnvProperties.EnvironmentConfig.ManualReviewed,
			Active:            refEnvProperties.EnvironmentConfig.Active,
			Namespace:         refEnvProperties.EnvironmentConfig.Namespace,
			EnvironmentId:     refEnvProperties.EnvironmentConfig.EnvironmentId,
			EnvironmentName:   refEnvProperties.EnvironmentConfig.EnvironmentName,
			Latest:            refEnvProperties.EnvironmentConfig.Latest,
			UserId:            userId,
			AppMetrics:        refEnvProperties.EnvironmentConfig.AppMetrics,
			ChartRefId:        refEnvProperties.EnvironmentConfig.ChartRefId,
			IsOverride:        refEnvProperties.EnvironmentConfig.IsOverride,
			IsBasicViewLocked: refEnvProperties.EnvironmentConfig.IsBasicViewLocked,
			CurrentViewEditor: refEnvProperties.EnvironmentConfig.CurrentViewEditor,
		}
		createResp, err := impl.propertiesConfigService.CreateEnvironmentProperties(newAppId, envPropertiesReq)
		if err != nil {
			if err.Error() == bean2.NOCHARTEXIST {
				templateRequest := chart.TemplateRequest{
					AppId:             newAppId,
					ChartRefId:        envPropertiesReq.ChartRefId,
					ValuesOverride:    []byte("{}"),
					UserId:            userId,
					IsBasicViewLocked: envPropertiesReq.IsBasicViewLocked,
					CurrentViewEditor: envPropertiesReq.CurrentViewEditor,
				}
				_, err = impl.chartService.CreateChartFromEnvOverride(templateRequest, ctx)
				if err != nil {
					impl.logger.Error(err)
					return nil, nil
				}
				createResp, err = impl.propertiesConfigService.CreateEnvironmentProperties(newAppId, envPropertiesReq)

			}
		}
		impl.logger.Debugw("env override create res", "createRes", createResp)
		//create object
		//save object

	}
	return nil, nil
}

func (impl *AppCloneServiceImpl) configDataClone(cfData []*bean3.ConfigData) []*bean3.ConfigData {
	var copiedData []*bean3.ConfigData
	for _, refdata := range cfData {
		data := &bean3.ConfigData{
			Name:               refdata.Name,
			Type:               refdata.Type,
			External:           refdata.External,
			MountPath:          refdata.MountPath,
			Data:               refdata.Data,
			DefaultData:        refdata.DefaultData,
			DefaultMountPath:   refdata.DefaultMountPath,
			Global:             refdata.Global,
			ExternalSecretType: refdata.ExternalSecretType,
		}
		copiedData = append(copiedData, data)
	}
	return copiedData
}

func (impl *AppCloneServiceImpl) CreateGlobalSecret(oldAppId, newAppId int, userId int32) (*bean3.ConfigDataRequest, error) {

	refCs, err := impl.configMapService.CSGlobalFetch(oldAppId)
	if err != nil {
		return nil, err
	}
	thisCm, err := impl.configMapService.CMGlobalFetch(newAppId)
	if err != nil {
		return nil, err
	}

	cfgDatas := impl.configDataClone(refCs.ConfigData)
	for _, cfgData := range cfgDatas {
		var configData []*bean3.ConfigData
		configData = append(configData, cfgData)
		newCm := &bean3.ConfigDataRequest{
			AppId:         newAppId,
			EnvironmentId: refCs.EnvironmentId,
			ConfigData:    configData,
			UserId:        userId,
			Id:            thisCm.Id,
		}
		thisCm, err = impl.configMapService.CSGlobalAddUpdate(newCm)
		if err != nil {
			return nil, err
		}
	}
	return thisCm, err
}

func (impl *AppCloneServiceImpl) CreateWf(oldAppId, newAppId int, userId int32, gitMaterialMapping map[int]int, ctx context.Context) (interface{}, error) {
	refAppWFs, err := impl.appWorkflowService.FindAppWorkflows(oldAppId)
	if err != nil {
		return nil, err
	}

	impl.logger.Debugw("workflow found", "wf", refAppWFs)

	createWorkflowMappingDtoResp := CreateWorkflowMappingDto{
		oldToNewCDPipelineId: make(map[int]int),
	}
	for _, refAppWF := range refAppWFs {
		thisWf := appWorkflow.AppWorkflowDto{
			Id:                    0,
			Name:                  refAppWF.Name,
			AppId:                 newAppId,
			AppWorkflowMappingDto: nil, //first create new mapping then add it
			UserId:                userId,
		}
		thisWf, err = impl.appWorkflowService.CreateAppWorkflow(thisWf)
		if err != nil {
			impl.logger.Errorw("error in creating workflow without external-ci", "err", err)
			return nil, err
		}

		isExternalCiPresent := false
		for _, awm := range refAppWF.AppWorkflowMappingDto {
			if awm.Type == appWorkflow2.WEBHOOK {
				isExternalCiPresent = true
				break
			}
		}
		createWorkflowMappingDto := CreateWorkflowMappingDto{
			newAppId:             newAppId,
			oldAppId:             oldAppId,
			newWfId:              thisWf.Id,
			userId:               userId,
			oldToNewCDPipelineId: createWorkflowMappingDtoResp.oldToNewCDPipelineId,
		}
		var externalCiPipelineId int
		if isExternalCiPresent {
			externalCiPipelineId, err = impl.createExternalCiAndAppWorkflowMapping(createWorkflowMappingDto)
			if err != nil {
				impl.logger.Errorw("error in createExternalCiAndAppWorkflowMapping", "err", err)
				return nil, err
			}
		}
		createWorkflowMappingDto.gitMaterialMapping = gitMaterialMapping
		createWorkflowMappingDto.externalCiPipelineId = externalCiPipelineId

		createWorkflowMappingDtoResp, err = impl.createWfInstances(refAppWF.AppWorkflowMappingDto, createWorkflowMappingDto, ctx)
		if err != nil {
			impl.logger.Errorw("error in creating workflow mapping", "err", err)
			return nil, err
		}
	}
	return nil, nil
}

func (impl *AppCloneServiceImpl) createExternalCiAndAppWorkflowMapping(createWorkflowMappingDto CreateWorkflowMappingDto) (int, error) {
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in beginning transaction", "err", err)
		return 0, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	externalCiPipelineId, _, err := impl.ciPipelineConfigService.CreateExternalCiAndAppWorkflowMapping(createWorkflowMappingDto.newAppId, createWorkflowMappingDto.newWfId, createWorkflowMappingDto.userId, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new external ci pipeline and new app workflow mapping", "refAppId", createWorkflowMappingDto.oldAppId, "newAppId", createWorkflowMappingDto.newAppId, "err", err)
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return externalCiPipelineId, nil
}

func (impl *AppCloneServiceImpl) createWfInstances(refWfMappings []appWorkflow.AppWorkflowMappingDto, createWorkflowMappingDto CreateWorkflowMappingDto, ctx context.Context) (CreateWorkflowMappingDto, error) {
	impl.logger.Debugw("wf mapping cloning", "refWfMappings", refWfMappings)
	var ciMapping []appWorkflow.AppWorkflowMappingDto
	var cdMappings []appWorkflow.AppWorkflowMappingDto
	var webhookMappings []appWorkflow.AppWorkflowMappingDto

	refWfMappings = appWorkflow.LevelWiseSort(refWfMappings)

	for _, appWf := range refWfMappings {
		if appWf.Type == appWorkflow2.CIPIPELINE {
			ciMapping = append(ciMapping, appWf)
		} else if appWf.Type == appWorkflow2.CDPIPELINE {
			cdMappings = append(cdMappings, appWf)
		} else if appWf.Type == appWorkflow2.WEBHOOK {
			webhookMappings = append(webhookMappings, appWf)
		} else {
			return createWorkflowMappingDto, fmt.Errorf("unsupported wf type: %s", appWf.Type)
		}
	}
	sourceToNewPipelineIdMapping := make(map[int]int)
	refApp, err := impl.pipelineBuilder.GetApp(createWorkflowMappingDto.oldAppId)
	if err != nil {
		impl.logger.Errorw("error in getting app from refAppId", "refAppId", createWorkflowMappingDto.oldAppId)
		return createWorkflowMappingDto, err
	}
	if len(webhookMappings) > 0 {
		for _, refwebhookMappings := range cdMappings {
			cdCloneReq := &cloneCdPipelineRequest{
				refCdPipelineId:       refwebhookMappings.ComponentId,
				refAppId:              createWorkflowMappingDto.oldAppId,
				appId:                 createWorkflowMappingDto.newAppId,
				userId:                createWorkflowMappingDto.userId,
				ciPipelineId:          0,
				appWfId:               createWorkflowMappingDto.newWfId,
				refAppName:            refApp.AppName,
				sourceToNewPipelineId: sourceToNewPipelineIdMapping,
				externalCiPipelineId:  createWorkflowMappingDto.externalCiPipelineId,
			}
			pipeline, err := impl.CreateCdPipeline(cdCloneReq, ctx)
			impl.logger.Debugw("cd pipeline created", "pipeline", pipeline)
			if err != nil {
				impl.logger.Errorw("error in getting cd-pipeline", "refAppId", createWorkflowMappingDto.oldAppId, "newAppId", createWorkflowMappingDto.newAppId, "err", err)
				return createWorkflowMappingDto, err
			}
		}
		return createWorkflowMappingDto, nil
	}

	if len(ciMapping) == 0 {
		impl.logger.Warn("no ci pipeline found")
		return createWorkflowMappingDto, nil
	} else if len(ciMapping) != 1 {
		impl.logger.Warn("more than one ci pipeline not supported")
		return createWorkflowMappingDto, nil
	}

	if err != nil {
		return createWorkflowMappingDto, err
	}
	var ci *bean.CiConfigRequest
	for _, refCiMapping := range ciMapping {
		impl.logger.Debugw("creating ci", "ref", refCiMapping)

		cloneCiPipelineRequest := &cloneCiPipelineRequest{
			refAppId:              createWorkflowMappingDto.oldAppId,
			refCiPipelineId:       refCiMapping.ComponentId,
			userId:                createWorkflowMappingDto.userId,
			appId:                 createWorkflowMappingDto.newAppId,
			wfId:                  createWorkflowMappingDto.newWfId,
			gitMaterialMapping:    createWorkflowMappingDto.gitMaterialMapping,
			refAppName:            refApp.AppName,
			oldToNewIdForLinkedCD: createWorkflowMappingDto.oldToNewCDPipelineId,
		}
		ci, err = impl.CreateCiPipeline(cloneCiPipelineRequest)
		if err != nil {
			impl.logger.Errorw("error in creating ci pipeline, app clone", "err", err)
			return createWorkflowMappingDto, err
		}
		impl.logger.Debugw("ci created", "ci", ci)
	}

	for _, refCdMapping := range cdMappings {
		cdCloneReq := &cloneCdPipelineRequest{
			refCdPipelineId:       refCdMapping.ComponentId,
			refAppId:              createWorkflowMappingDto.oldAppId,
			appId:                 createWorkflowMappingDto.newAppId,
			userId:                createWorkflowMappingDto.userId,
			ciPipelineId:          ci.CiPipelines[0].Id,
			appWfId:               createWorkflowMappingDto.newWfId,
			refAppName:            refApp.AppName,
			sourceToNewPipelineId: sourceToNewPipelineIdMapping,
		}
		pipeline, err := impl.CreateCdPipeline(cdCloneReq, ctx)
		if err != nil {
			impl.logger.Errorw("error in creating cd pipeline, app clone", "err", err)
			return createWorkflowMappingDto, err
		}
		createWorkflowMappingDto.oldToNewCDPipelineId[refCdMapping.ComponentId] = pipeline.Pipelines[0].Id
		impl.logger.Debugw("cd pipeline created", "pipeline", pipeline)
	}

	//find ci
	//save ci
	//find cd
	//save cd
	//save mappings
	return createWorkflowMappingDto, nil
}

type cloneCiPipelineRequest struct {
	refAppId              int
	refCiPipelineId       int
	userId                int32
	appId                 int
	wfId                  int
	gitMaterialMapping    map[int]int
	refAppName            string
	oldToNewIdForLinkedCD map[int]int
}

func (impl *AppCloneServiceImpl) CreateCiPipeline(req *cloneCiPipelineRequest) (*bean.CiConfigRequest, error) {
	refCiConfig, err := impl.pipelineBuilder.GetCiPipeline(req.refAppId)
	if err != nil {
		return nil, err
	}

	var refCiPipeline *bean.CiPipeline
	var uniqueId int
	for id, reqCiPipeline := range refCiConfig.CiPipelines {
		if reqCiPipeline.Id == req.refCiPipelineId {
			refCiPipeline = reqCiPipeline
			uniqueId = id
			break
		}
	}
	if refCiPipeline == nil {
		return nil, nil
	}
	pipelineName := refCiPipeline.Name
	if strings.HasPrefix(pipelineName, req.refAppName) {
		pipelineName = strings.Replace(pipelineName, req.refAppName+"-ci-", "", 1)
	}

	pipelineExists, err := impl.ciPipelineRepository.CheckIfPipelineExistsByNameAndAppId(pipelineName, req.appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching pipeline by name, FindByName", "err", err, "patch cipipeline name", pipelineName)
		return nil, err
	}
	if pipelineExists {
		pipelineName = fmt.Sprintf("%s-%d", pipelineName, uniqueId) // making pipeline name unique
	}
	var ciMaterilas []*bean.CiMaterial
	for _, refCiMaterial := range refCiPipeline.CiMaterial {
		//FIXME
		gitMaterialId := req.gitMaterialMapping[refCiMaterial.GitMaterialId]
		if refCiPipeline.ParentCiPipeline != 0 {
			gitMaterialId = refCiMaterial.GitMaterialId
		}
		ciMaterial := &bean.CiMaterial{
			GitMaterialId: gitMaterialId,
			Id:            0,
			Source: &bean.SourceTypeConfig{
				Type:  refCiMaterial.Source.Type,
				Value: refCiMaterial.Source.Value,
				Regex: refCiMaterial.Source.Regex,
			},
		}
		ciMaterilas = append(ciMaterilas, ciMaterial)
	}
	var beforeDockerBuildScripts []*bean.CiScript
	var afterDockerBuildScripts []*bean.CiScript

	for _, script := range refCiPipeline.BeforeDockerBuildScripts {
		ciScript := &bean.CiScript{
			Id:             0,
			Index:          script.Index,
			Name:           script.Name,
			Script:         script.Script,
			OutputLocation: script.OutputLocation,
		}
		beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciScript)
	}
	for _, script := range refCiPipeline.AfterDockerBuildScripts {
		ciScript := &bean.CiScript{
			Id:             0,
			Index:          script.Index,
			Name:           script.Name,
			Script:         script.Script,
			OutputLocation: script.OutputLocation,
		}
		afterDockerBuildScripts = append(afterDockerBuildScripts, ciScript)
	}

	//getting pre stage and post stage details
	preStageDetail, postStageDetail, err := impl.pipelineStageService.GetCiPipelineStageDataDeepCopy(refCiPipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in getting pre & post stage detail by ciPipelineId", "err", err, "ciPipelineId", refCiPipeline.Id)
		return nil, err
	}
	ciPatchReq := &bean.CiPatchRequest{
		CiPipeline: &bean.CiPipeline{
			IsManual:                 refCiPipeline.IsManual,
			DockerArgs:               refCiPipeline.DockerArgs,
			IsExternal:               refCiPipeline.IsExternal,
			ExternalCiConfig:         bean.ExternalCiConfig{},
			CiMaterial:               ciMaterilas,
			Name:                     pipelineName,
			Id:                       0,
			Version:                  refCiPipeline.Version,
			Active:                   refCiPipeline.Active,
			Deleted:                  refCiPipeline.Deleted,
			BeforeDockerBuild:        refCiPipeline.BeforeDockerBuild,
			AfterDockerBuild:         refCiPipeline.AfterDockerBuild,
			BeforeDockerBuildScripts: beforeDockerBuildScripts,
			AfterDockerBuildScripts:  afterDockerBuildScripts,
			ParentCiPipeline:         refCiPipeline.ParentCiPipeline,
			IsDockerConfigOverridden: refCiPipeline.IsDockerConfigOverridden,
			PreBuildStage:            preStageDetail,
			PostBuildStage:           postStageDetail,
			EnvironmentId:            refCiPipeline.EnvironmentId,
			ScanEnabled:              refCiPipeline.ScanEnabled,
			PipelineType:             refCiPipeline.PipelineType,
		},
		AppId:         req.appId,
		Action:        bean.CREATE,
		AppWorkflowId: req.wfId,
		UserId:        req.userId,
		IsCloneJob:    true,
	}
	if refCiPipeline.EnvironmentId != 0 {
		ciPatchReq.IsJob = true
	}
	if !refCiPipeline.IsExternal && refCiPipeline.IsDockerConfigOverridden {
		//get template override
		templateOverrideBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(refCiPipeline.Id)
		if err != nil {
			return nil, err
		}
		templateOverride := templateOverrideBean.CiTemplateOverride
		ciBuildConfig := templateOverrideBean.CiBuildConfig
		//getting new git material for this app
		//gitMaterial, err := impl.materialRepository.FindByAppIdAndCheckoutPath(req.appId, templateOverride.GitMaterial.CheckoutPath)
		if len(req.gitMaterialMapping) == 0 {
			impl.logger.Errorw("no git materials found for the app", "appId", req.appId)
			return nil, fmt.Errorf("no git materials found for the app, %d", req.appId)
		}
		gitMaterialId := req.gitMaterialMapping[ciBuildConfig.GitMaterialId]
		buildContextGitMaterialId := req.gitMaterialMapping[ciBuildConfig.BuildContextGitMaterialId]
		if gitMaterialId == 0 {
			for _, id := range req.gitMaterialMapping {
				gitMaterialId = id
				break
			}
		}
		if buildContextGitMaterialId == 0 {
			buildContextGitMaterialId = gitMaterialId
		}
		ciBuildConfig.GitMaterialId = gitMaterialId
		ciBuildConfig.BuildContextGitMaterialId = buildContextGitMaterialId
		templateOverride.GitMaterialId = gitMaterialId
		ciBuildConfig.Id = 0
		ciPatchReq.CiPipeline.DockerConfigOverride = bean.DockerConfigOverride{
			DockerRegistry:   templateOverride.DockerRegistryId,
			DockerRepository: templateOverride.DockerRepository,
			CiBuildConfig:    ciBuildConfig,
		}
	} else if refCiPipeline.IsExternal {
		ciPatchReq.CiPipeline.IsDockerConfigOverridden = false
	}
	return impl.pipelineBuilder.PatchCiPipeline(ciPatchReq)
}

type cloneCdPipelineRequest struct {
	refCdPipelineId       int
	refAppId              int
	appId                 int
	userId                int32
	ciPipelineId          int
	appWfId               int
	refAppName            string
	sourceToNewPipelineId map[int]int
	externalCiPipelineId  int
}

func (impl *AppCloneServiceImpl) CreateCdPipeline(req *cloneCdPipelineRequest, ctx context.Context) (*bean.CdPipelines, error) {
	refPipelines, err := impl.pipelineBuilder.GetCdPipelinesForApp(req.refAppId)
	if err != nil {
		return nil, err
	}
	var refCdPipeline *bean.CDPipelineConfigObject
	for _, refPipeline := range refPipelines.Pipelines {
		if refPipeline.Id == req.refCdPipelineId {
			refCdPipeline = refPipeline
			break
		}
	}
	if refCdPipeline == nil {
		return nil, fmt.Errorf("no cd pipeline found")
	}
	refCdPipeline.SourceToNewPipelineId = req.sourceToNewPipelineId
	pipelineName := refCdPipeline.Name
	if strings.HasPrefix(pipelineName, req.refAppName) {
		pipelineName = strings.Replace(pipelineName, req.refAppName+"-", "", 1)
	}
	// by default all deployment types are allowed
	AllowedDeploymentAppTypes := map[string]bool{
		util.PIPELINE_DEPLOYMENT_TYPE_ACD:  true,
		util.PIPELINE_DEPLOYMENT_TYPE_HELM: true,
	}
	DeploymentAppConfigForEnvironment, err := impl.pipelineBuilder.GetDeploymentConfigMap(refCdPipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment config for environment", "err", err)
	}
	for deploymentType, allowed := range DeploymentAppConfigForEnvironment {
		AllowedDeploymentAppTypes[deploymentType] = allowed
	}
	isGitopsConfigured, err := impl.pipelineBuilder.IsGitopsConfigured()
	if err != nil {
		impl.logger.Errorw("error in checking if gitOps configured", "err", err)
		return nil, err
	}
	var deploymentAppType string
	if AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_ACD] && AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_HELM] {
		deploymentAppType = refCdPipeline.DeploymentAppType
	} else if AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_ACD] && isGitopsConfigured {
		deploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
	} else if AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_HELM] {
		deploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	}

	cdPipeline := &bean.CDPipelineConfigObject{
		Id:                            0,
		EnvironmentId:                 refCdPipeline.EnvironmentId,
		CiPipelineId:                  req.ciPipelineId,
		TriggerType:                   refCdPipeline.TriggerType,
		Name:                          pipelineName,
		Strategies:                    refCdPipeline.Strategies,
		Namespace:                     refCdPipeline.Namespace,
		AppWorkflowId:                 req.appWfId,
		DeploymentTemplate:            refCdPipeline.DeploymentTemplate,
		PreStage:                      refCdPipeline.PreStage, //FIXME
		PostStage:                     refCdPipeline.PostStage,
		PreStageConfigMapSecretNames:  refCdPipeline.PreStageConfigMapSecretNames,
		PostStageConfigMapSecretNames: refCdPipeline.PostStageConfigMapSecretNames,
		RunPostStageInEnv:             refCdPipeline.RunPostStageInEnv,
		RunPreStageInEnv:              refCdPipeline.RunPreStageInEnv,
		DeploymentAppType:             deploymentAppType,
		PreDeployStage:                refCdPipeline.PreDeployStage,
		PostDeployStage:               refCdPipeline.PostDeployStage,
		SourceToNewPipelineId:         refCdPipeline.SourceToNewPipelineId,
		RefPipelineId:                 refCdPipeline.Id,
		ParentPipelineType:            refCdPipeline.ParentPipelineType,
		IsDigestEnforcedForPipeline:   refCdPipeline.IsDigestEnforcedForPipeline,
	}
	if refCdPipeline.ParentPipelineType == "WEBHOOK" {
		cdPipeline.CiPipelineId = 0
		cdPipeline.ParentPipelineId = req.externalCiPipelineId
	} else if refCdPipeline.ParentPipelineType != appWorkflow.CI_PIPELINE_TYPE {
		cdPipeline.ParentPipelineId = refCdPipeline.ParentPipelineId
	}
	cdPipelineReq := &bean.CdPipelines{
		Pipelines: []*bean.CDPipelineConfigObject{cdPipeline},
		AppId:     req.appId,
		UserId:    req.userId,
	}
	cdPipelineRes, err := impl.pipelineBuilder.CreateCdPipelines(cdPipelineReq, ctx)
	return cdPipelineRes, err

}
