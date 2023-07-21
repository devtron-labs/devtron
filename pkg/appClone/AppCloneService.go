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
	appRepository app2.AppRepository) *AppCloneServiceImpl {
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
	}
}

type CloneRequest struct {
	RefAppId    int                           `json:"refAppId"`
	Name        string                        `json:"name"`
	ProjectId   int                           `json:"projectId"`
	AppLabels   []*bean.Label                 `json:"labels,omitempty" validate:"dive"`
	Description bean2.GenericNoteResponseBean `json:"description"`
	AppType     helper.AppType                `json:"appType"`
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
		Description: createReq.Description,
	}
	userId := createReq.UserId
	appStatus, err := impl.appListingService.FetchAppStageStatus(cloneReq.RefAppId, int(cloneReq.AppType))
	if err != nil {
		return nil, err
	}
	refApp, err := impl.pipelineBuilder.GetApp(cloneReq.RefAppId)
	if err != nil {
		return nil, err
	}
	isSameProject := refApp.TeamId == cloneReq.ProjectId
	/*	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(0, "APP", stages.AppId))
		appStageStatus = append(appStageStatus, impl.makeAppStageStatus(1, "MATERIAL", materialExists))
		appStageStatus = append(appStageStatus, impl.makeAppStageStatus(2, "TEMPLATE", stages.CiTemplateId))
		appStageStatus = append(appStageStatus, impl.makeAppStageStatus(3, "CI_PIPELINE", stages.CiPipelineId))
		appStageStatus = append(appStageStatus, impl.makeAppStageStatus(4, "CHART", stages.ChartId))
		appStageStatus = append(appStageStatus, impl.makeAppStageStatus(5, "CD_PIPELINE", stages.PipelineId))
	*/
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
	if isSameProject {
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
	}
	_, err = impl.CreateWf(cloneReq.RefAppId, newAppId, userId, gitMaerialMap, context, isSameProject)
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
		Description: cloneReq.Description,
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

func (impl *AppCloneServiceImpl) CreateGlobalCM(oldAppId, newAppId int, userId int32) (*pipeline.ConfigDataRequest, error) {
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
		newCm := &pipeline.ConfigDataRequest{
			AppId:         newAppId,
			EnvironmentId: refCM.EnvironmentId,
			ConfigData:    []*pipeline.ConfigData{cfgData},
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

		var refEnvCm []*pipeline.ConfigData
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
			newCm := &pipeline.ConfigDataRequest{
				AppId:         newAppId,
				EnvironmentId: refEnv.EnvironmentId,
				ConfigData:    []*pipeline.ConfigData{cfgData},
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

		var refEnvCm []*pipeline.ConfigData
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
			var configData []*pipeline.ConfigData
			configData = append(configData, cfgData)
			newCm := &pipeline.ConfigDataRequest{
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
		envPropertiesReq := &pipeline.EnvironmentProperties{
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

func (impl *AppCloneServiceImpl) configDataClone(cfData []*pipeline.ConfigData) []*pipeline.ConfigData {
	var copiedData []*pipeline.ConfigData
	for _, refdata := range cfData {
		data := &pipeline.ConfigData{
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

func (impl *AppCloneServiceImpl) CreateGlobalSecret(oldAppId, newAppId int, userId int32) (*pipeline.ConfigDataRequest, error) {

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
		var configData []*pipeline.ConfigData
		configData = append(configData, cfgData)
		newCm := &pipeline.ConfigDataRequest{
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

func (impl *AppCloneServiceImpl) CreateWf(oldAppId, newAppId int, userId int32, gitMaterialMapping map[int]int, ctx context.Context, isSameProject bool) (interface{}, error) {
	refAppWFs, err := impl.appWorkflowService.FindAppWorkflows(oldAppId)
	if err != nil {
		return nil, err
	}
	impl.logger.Debugw("workflow found", "wf", refAppWFs)
	for _, refAppWF := range refAppWFs {
		thisWf := appWorkflow.AppWorkflowDto{
			Id:                    0,
			Name:                  refAppWF.Name,
			AppId:                 newAppId,
			AppWorkflowMappingDto: nil, //first create new mapping then add it
			UserId:                userId,
		}

		isExternalCiPresent := false
		for _, awm := range refAppWF.AppWorkflowMappingDto {
			if awm.Type == appWorkflow2.WEBHOOK {
				isExternalCiPresent = true
			}
		}

		if !isExternalCiPresent {
			thisWf, err = impl.appWorkflowService.CreateAppWorkflow(thisWf)
			impl.logger.Debugw("workflow found", thisWf)
			if err != nil {
				impl.logger.Errorw("errir in creating workflow without extenal-ci", "err", err)
				return nil, err
			}
		}

		err = impl.createWfMappings(refAppWF.AppWorkflowMappingDto, oldAppId, newAppId, userId, thisWf.Id, gitMaterialMapping, ctx, isSameProject)
		if err != nil {
			impl.logger.Errorw("error in creating workflow mapping", "err", err)
			return nil, err
		}
	}
	return nil, nil
}

func (impl *AppCloneServiceImpl) createWfMappings(refWfMappings []appWorkflow.AppWorkflowMappingDto, oldAppId, newAppId int, userId int32, thisWfId int, gitMaterialMapping map[int]int, ctx context.Context, isSameProject bool) error {
	impl.logger.Debugw("wf mapping cloning", "refWfMappings", refWfMappings)
	var ciMapping []appWorkflow.AppWorkflowMappingDto
	var cdMappings []appWorkflow.AppWorkflowMappingDto
	var webhookMappings []appWorkflow.AppWorkflowMappingDto
	for _, appWf := range refWfMappings {
		if appWf.Type == appWorkflow2.CIPIPELINE {
			ciMapping = append(ciMapping, appWf)
		} else if appWf.Type == appWorkflow2.CDPIPELINE {
			cdMappings = append(cdMappings, appWf)
		} else if appWf.Type == appWorkflow2.WEBHOOK {
			webhookMappings = append(webhookMappings, appWf)
		} else {
			return fmt.Errorf("unsupported wf type: %s", appWf.Type)
		}
	}
	refApp, err := impl.pipelineBuilder.GetApp(oldAppId)
	if len(webhookMappings) > 0 {
		if isSameProject {
			for _, refwebhookMappings := range cdMappings {
				cdCloneReq := &cloneCdPipelineRequest{
					refCdPipelineId: refwebhookMappings.ComponentId,
					refAppId:        oldAppId,
					appId:           newAppId,
					userId:          userId,
					ciPipelineId:    0,
					appWfId:         thisWfId,
					refAppName:      refApp.AppName,
				}
				pipeline, err := impl.CreateCdPipeline(cdCloneReq, ctx)
				impl.logger.Debugw("cd pipeline created", "pipeline", pipeline)
				if err != nil {
					impl.logger.Errorw("error in getting cd-pipeling", "err", err)
					return err
				}
			}
		} else {
			impl.logger.Debug("not the same project, skipping cd pipeline creation")
		}
		return nil
	}

	if len(ciMapping) == 0 {
		impl.logger.Warn("no ci pipeline found")
		return nil
	} else if len(ciMapping) != 1 {
		impl.logger.Warn("more than one cd pipeline not supported")
		return nil
	}

	if err != nil {
		return err
	}
	var ci *bean.CiConfigRequest
	for _, refCiMapping := range ciMapping {
		impl.logger.Debugw("creating ci", "ref", refCiMapping)

		cloneCiPipelineRequest := &cloneCiPipelineRequest{
			refAppId:           oldAppId,
			refCiPipelineId:    refCiMapping.ComponentId,
			userId:             userId,
			appId:              newAppId,
			wfId:               thisWfId,
			gitMaterialMapping: gitMaterialMapping,
			refAppName:         refApp.AppName,
		}
		ci, err = impl.CreateCiPipeline(cloneCiPipelineRequest)
		if err != nil {
			impl.logger.Errorw("error in creating ci pipeline, app clone", "err", err)
			return err
		}
		impl.logger.Debugw("ci created", "ci", ci)
	}
	if isSameProject {
		for _, refCdMapping := range cdMappings {
			cdCloneReq := &cloneCdPipelineRequest{
				refCdPipelineId: refCdMapping.ComponentId,
				refAppId:        oldAppId,
				appId:           newAppId,
				userId:          userId,
				ciPipelineId:    ci.CiPipelines[0].Id,
				appWfId:         thisWfId,
				refAppName:      refApp.AppName,
			}
			pipeline, err := impl.CreateCdPipeline(cdCloneReq, ctx)
			if err != nil {
				impl.logger.Errorw("error in creating cd pipeline, app clone", "err", err)
				return err
			}
			impl.logger.Debugw("cd pipeline created", "pipeline", pipeline)
		}
	} else {
		impl.logger.Debug("not the same project, skipping cd pipeline creation")
	}

	//find ci
	//save ci
	//find cd
	//save cd
	//save mappings
	return nil
}

type cloneCiPipelineRequest struct {
	refAppId           int
	refCiPipelineId    int
	userId             int32
	appId              int
	wfId               int
	gitMaterialMapping map[int]int
	refAppName         string
}

func (impl *AppCloneServiceImpl) CreateCiPipeline(req *cloneCiPipelineRequest) (*bean.CiConfigRequest, error) {
	refCiConfig, err := impl.pipelineBuilder.GetCiPipeline(req.refAppId)
	if err != nil {
		return nil, err
	}
	for _, refCiPipeline := range refCiConfig.CiPipelines {
		if refCiPipeline.Id == req.refCiPipelineId {
			pipelineName := refCiPipeline.Name
			if strings.HasPrefix(pipelineName, req.refAppName) {
				pipelineName = strings.Replace(pipelineName, req.refAppName+"-ci-", "", 1)
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
			}

			return impl.pipelineBuilder.PatchCiPipeline(ciPatchReq)
		}
	}
	return nil, fmt.Errorf("ci pipeline not found ")
}

type cloneCdPipelineRequest struct {
	refCdPipelineId int
	refAppId        int
	appId           int
	userId          int32
	ciPipelineId    int
	appWfId         int
	refAppName      string
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

	if refCdPipeline.ParentPipelineType == "WEBHOOK" {
		cdPipeline := &bean.CDPipelineConfigObject{
			Id:                            0,
			EnvironmentId:                 refCdPipeline.EnvironmentId,
			CiPipelineId:                  0,
			TriggerType:                   refCdPipeline.TriggerType,
			Name:                          pipelineName,
			Strategies:                    refCdPipeline.Strategies,
			Namespace:                     refCdPipeline.Namespace,
			AppWorkflowId:                 0,
			DeploymentTemplate:            refCdPipeline.DeploymentTemplate,
			PreStage:                      refCdPipeline.PreStage, //FIXME
			PostStage:                     refCdPipeline.PostStage,
			PreStageConfigMapSecretNames:  refCdPipeline.PreStageConfigMapSecretNames,
			PostStageConfigMapSecretNames: refCdPipeline.PostStageConfigMapSecretNames,
			RunPostStageInEnv:             refCdPipeline.RunPostStageInEnv,
			RunPreStageInEnv:              refCdPipeline.RunPreStageInEnv,
			DeploymentAppType:             refCdPipeline.DeploymentAppType,
			ParentPipelineId:              0,
			ParentPipelineType:            refCdPipeline.ParentPipelineType,
		}
		cdPipelineReq := &bean.CdPipelines{
			Pipelines: []*bean.CDPipelineConfigObject{cdPipeline},
			AppId:     req.appId,
			UserId:    req.userId,
		}
		cdPipelineRes, err := impl.pipelineBuilder.CreateCdPipelines(cdPipelineReq, ctx)
		return cdPipelineRes, err
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
	}
	cdPipelineReq := &bean.CdPipelines{
		Pipelines: []*bean.CDPipelineConfigObject{cdPipeline},
		AppId:     req.appId,
		UserId:    req.userId,
	}
	cdPipelineRes, err := impl.pipelineBuilder.CreateCdPipelines(cdPipelineReq, ctx)
	return cdPipelineRes, err

}
