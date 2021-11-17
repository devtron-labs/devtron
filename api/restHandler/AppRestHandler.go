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

package restHandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	appBean "github.com/devtron-labs/devtron/api/appbean"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	appWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	teamRepo "github.com/devtron-labs/devtron/internal/sql/repository/team"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	APP_DELETE_FAILED_RESP     = "Devtron was unable to delete the app by itself, please try deleting the app manually."
	APP_CREATE_SUCCESSFUL_RESP = "App created successfully."
)

type AppRestHandler interface {
	GetAppAllDetail(w http.ResponseWriter, r *http.Request)
	CreateApp(w http.ResponseWriter, r *http.Request)
}

type AppRestHandlerImpl struct {
	logger                  *zap.SugaredLogger
	userAuthService         user.UserService
	validator               *validator.Validate
	enforcerUtil            rbac.EnforcerUtil
	enforcer                rbac.Enforcer
	appLabelService         app.AppLabelService
	pipelineBuilder         pipeline.PipelineBuilder
	gitRegistryService      pipeline.GitRegistryConfig
	chartService            pipeline.ChartService
	configMapService        pipeline.ConfigMapService
	appListingService       app.AppListingService
	propertiesConfigService pipeline.PropertiesConfigService
	appWorkflowService      appWorkflow.AppWorkflowService
	materialRepository      pipelineConfig.MaterialRepository
	teamRepository          teamRepo.TeamRepository
	gitProviderRepo         repository.GitProviderRepository
	appWorkflowRepository   appWorkflow2.AppWorkflowRepository
	environmentRepository   cluster.EnvironmentRepository
	configMapRepository     chartConfig.ConfigMapRepository
	envConfigRepo           chartConfig.EnvConfigOverrideRepository
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService, validator *validator.Validate, enforcerUtil rbac.EnforcerUtil,
	enforcer rbac.Enforcer, appLabelService app.AppLabelService, pipelineBuilder pipeline.PipelineBuilder, gitRegistryService pipeline.GitRegistryConfig,
	chartService pipeline.ChartService, configMapService pipeline.ConfigMapService, appListingService app.AppListingService,
	propertiesConfigService pipeline.PropertiesConfigService, appWorkflowService appWorkflow.AppWorkflowService,
	materialRepository pipelineConfig.MaterialRepository, teamRepository teamRepo.TeamRepository, gitProviderRepo repository.GitProviderRepository,
	appWorkflowRepository appWorkflow2.AppWorkflowRepository, environmentRepository cluster.EnvironmentRepository, configMapRepository chartConfig.ConfigMapRepository,
	envConfigRepo chartConfig.EnvConfigOverrideRepository) *AppRestHandlerImpl {
	handler := &AppRestHandlerImpl{
		logger:                  logger,
		userAuthService:         userAuthService,
		validator:               validator,
		enforcerUtil:            enforcerUtil,
		enforcer:                enforcer,
		appLabelService:         appLabelService,
		pipelineBuilder:         pipelineBuilder,
		gitRegistryService:      gitRegistryService,
		chartService:            chartService,
		configMapService:        configMapService,
		appListingService:       appListingService,
		propertiesConfigService: propertiesConfigService,
		appWorkflowService:      appWorkflowService,
		materialRepository:      materialRepository,
		teamRepository:          teamRepository,
		gitProviderRepo:         gitProviderRepo,
		appWorkflowRepository:   appWorkflowRepository,
		environmentRepository:   environmentRepository,
		configMapRepository:     configMapRepository,
		envConfigRepo:           envConfigRepo,
	}
	return handler
}

func (handler AppRestHandlerImpl) GetAppAllDetail(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rbac implementation for app (user should be admin)
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, object); !ok {
		handler.logger.Errorw("Unauthorized User for app update action", "err", err, "appId", appId)
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac implementation ends here for app

	handler.logger.Debugw("Getting app detail v2", "appId", appId)

	//get/build app metadata starts
	appMetadataResp, done := handler.validateAndBuildAppMetadata(w, appId)
	if done {
		return
	}
	//get/build app metadata ends

	//get/build git materials starts
	gitMaterialsResp, done := handler.validateAndBuildAppGitMaterials(w, appId)
	if done {
		return
	}
	//get/build git materials ends

	//get/build docker config starts
	dockerConfig, done := handler.validateAndBuildDockerConfig(w, appId)
	if done {
		return
	}
	//get/build docker config ends

	//get/build global deployment template starts
	globalDeploymentTemplateResp, done := handler.validateAndBuildAppDeploymentTemplate(w, appId, 0)
	if done {
		return
	}
	//get/build global deployment template ends

	//get/build app workflows starts
	appWorkflows, done := handler.validateAndBuildAppWorkflows(w, appId)
	if done {
		return
	}
	//get/build app workflows ends

	//get/build global config maps starts
	globalConfigMapsResp, done := handler.validateAndBuildAppGlobalConfigMaps(w, appId)
	if done {
		return
	}
	//get/build global config maps ends

	//get/build global secrets starts
	globalSecretsResp, done := handler.validateAndBuildAppGlobalSecrets(w, appId)
	if done {
		return
	}
	//get/build global secrets ends

	//get/build environment override starts
	environmentOverrides, done := handler.validateAndBuildEnvironmentOverrides(w, appId, token)
	if done {
		return
	}
	//get/build environment override ends

	//build full object for response
	appDetail := &appBean.AppDetail{
		Metadata:                 appMetadataResp,
		GitMaterials:             gitMaterialsResp,
		DockerConfig:             dockerConfig,
		GlobalDeploymentTemplate: globalDeploymentTemplateResp,
		AppWorkflows:             appWorkflows,
		GlobalConfigMaps:         globalConfigMapsResp,
		GlobalSecrets:            globalSecretsResp,
		EnvironmentOverrides:     environmentOverrides,
	}
	//end

	writeJsonResp(w, nil, appDetail, http.StatusOK)
}

func (handler AppRestHandlerImpl) CreateApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	ctx := context.WithValue(r.Context(), "token", token)
	var createAppRequest appBean.AppDetail
	err = decoder.Decode(&createAppRequest)
	if err != nil {
		handler.logger.Errorw("request err, CreateApp by API", "err", err, "CreateApp", createAppRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//to add more validations here
	handler.logger.Infow("request payload, CreateApp by API", "CreateApp", createAppRequest)
	err = handler.validator.Struct(createAppRequest)
	if err != nil {
		handler.logger.Errorw("validation err, CreateApp by API", "err", err, "CreateApp", createAppRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//TODO implement rbac
	//rbac starts
	//rbac ends

	handler.logger.Infow("creating app v2", "createAppRequest", createAppRequest)

	//creating blank app starts
	createBlankAppResp, done := handler.createBlankApp(w, createAppRequest.Metadata, userId)
	if done {
		return
	}
	//creating blank app ends

	//declaring appId for creating other components of app
	appId := createBlankAppResp.Id

	//creating git material starts
	done = handler.createGitMaterials(w, appId, createAppRequest.GitMaterials, userId)
	if done {
		handler.deleteApp(w, ctx, appId, userId)
		return
	}
	//creating git material ends

	//creating docker config starts
	done = handler.createDockerConfig(w, appId, createAppRequest.DockerConfig, userId)
	if done {
		handler.deleteApp(w, ctx, appId, userId)
		return
	}
	//creating docker config ends

	//creating deployment template starts
	done = handler.createDeploymentTemplate(w, ctx, appId, createAppRequest.GlobalDeploymentTemplate, userId)
	if done {
		handler.deleteApp(w, ctx, appId, userId)
		return
	}
	//creating deployment template ends

	//creating global configMaps starts
	done = handler.createGlobalConfigMaps(w, appId, userId, createAppRequest.GlobalConfigMaps)
	if done {
		handler.deleteApp(w, ctx, appId, userId)
		return
	}
	//creating global configMaps ends

	//creating global secrets starts
	done = handler.createGlobalSecrets(w, appId, userId, createAppRequest.GlobalSecrets)
	if done {
		handler.deleteApp(w, ctx, appId, userId)
		return
	}
	//creating global secrets ends

	//creating workflow starts
	done = handler.createWorkflows(w, ctx, appId, userId, createAppRequest.AppWorkflows)
	if done {
		handler.deleteApp(w, ctx, appId, userId)
		return
	}
	//creating workflow ends

	//creating environment override starts
	done = handler.createEnvOverrides(w, ctx, appId, userId, createAppRequest.EnvironmentOverrides)
	if done {
		handler.deleteApp(w, ctx, appId, userId)
		return
	}
	//creating environment override ends

	writeJsonResp(w, nil, APP_CREATE_SUCCESSFUL_RESP, http.StatusOK)
}

//GetApp related methods starts

//get/build app metadata
func (handler AppRestHandlerImpl) validateAndBuildAppMetadata(w http.ResponseWriter, appId int) (*appBean.AppMetadata, bool) {
	handler.logger.Debugw("Getting app detail - meta data", "appId", appId)

	appMetaInfo, err := handler.appLabelService.GetAppMetaInfo(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	if appMetaInfo == nil {
		err = errors.New("invalid appId - appMetaInfo is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, true
	}

	var appLabelsRes []*appBean.AppLabel
	if len(appMetaInfo.Labels) > 0 {
		for _, label := range appMetaInfo.Labels {
			appLabelsRes = append(appLabelsRes, &appBean.AppLabel{
				Key:   label.Key,
				Value: label.Value,
			})
		}
	}
	appMetadataResp := &appBean.AppMetadata{
		AppName:     appMetaInfo.AppName,
		ProjectName: appMetaInfo.ProjectName,
		Labels:      appLabelsRes,
	}

	return appMetadataResp, false
}

//get/build git materials
func (handler AppRestHandlerImpl) validateAndBuildAppGitMaterials(w http.ResponseWriter, appId int) ([]*appBean.GitMaterial, bool) {
	handler.logger.Debugw("Getting app detail - git materials", "appId", appId)

	gitMaterials := handler.pipelineBuilder.GetMaterialsForAppId(appId)
	var gitMaterialsResp []*appBean.GitMaterial
	if len(gitMaterials) > 0 {
		for _, gitMaterial := range gitMaterials {
			gitRegistry, err := handler.gitRegistryService.FetchOneGitProvider(strconv.Itoa(gitMaterial.GitProviderId))
			if err != nil {
				handler.logger.Errorw("service err, getGitProvider in GetAppAllDetail", "err", err, "appId", appId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return nil, true
			}

			gitMaterialsResp = append(gitMaterialsResp, &appBean.GitMaterial{
				GitRepoUrl:      gitMaterial.Url,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				GitProviderUrl:  gitRegistry.Url,
			})
		}
	}
	return gitMaterialsResp, false
}

//get/build docker build config
func (handler AppRestHandlerImpl) validateAndBuildDockerConfig(w http.ResponseWriter, appId int) (*appBean.DockerConfig, bool) {
	handler.logger.Debugw("Getting app detail - docker build", "appId", appId)

	ciConfig, err := handler.pipelineBuilder.GetCiPipeline(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetCiPipeline in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	//getting gitMaterialUrl by id
	gitMaterial, err := handler.materialRepository.FindById(ciConfig.DockerBuildConfig.GitMaterialId)
	if err != nil {
		handler.logger.Errorw("error in fetching materialUrl by ID in GetAppAllDetail", "err", err, "gitMaterialId", ciConfig.DockerBuildConfig.GitMaterialId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	dockerConfig := &appBean.DockerConfig{
		DockerRegistry:   ciConfig.DockerRegistry,
		DockerRepository: ciConfig.DockerRepository,
		BuildConfig: &appBean.DockerBuildConfig{
			Args:                   ciConfig.DockerBuildConfig.Args,
			DockerfileRelativePath: ciConfig.DockerBuildConfig.DockerfilePath,
			GitCheckoutPath:        gitMaterial.CheckoutPath,
		},
	}

	return dockerConfig, false
}

//get/build deployment template
func (handler AppRestHandlerImpl) validateAndBuildAppDeploymentTemplate(w http.ResponseWriter, appId int, envId int) (*appBean.DeploymentTemplate, bool) {
	handler.logger.Debugw("Getting app detail - deployment template", "appId", appId, "envId", envId)

	chartRefData, err := handler.chartService.ChartRefAutocompleteForAppOrEnv(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, ChartRefAutocompleteForAppOrEnv in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	if chartRefData == nil {
		err = errors.New("invalid appId/envId - chartRefData is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, true
	}

	appDeploymentTemplate, err := handler.chartService.FindLatestChartForAppByAppId(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentTemplate in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	if appDeploymentTemplate == nil {
		err = errors.New("invalid appId - deploymentTemplate is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, true
	}

	//set deployment template & showAppMetrics
	var showAppMetrics bool
	var deploymentTemplateRaw json.RawMessage
	var chartRefId int
	if envId > 0 {
		//on env level
		env, err := handler.propertiesConfigService.GetEnvironmentProperties(appId, envId, chartRefData.LatestEnvChartRef)
		if err != nil {
			handler.logger.Errorw("service err, GetEnvironmentProperties in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return nil, true
		}
		chartRefId = chartRefData.LatestEnvChartRef
		if env.EnvironmentConfig.IsOverride {
			deploymentTemplateRaw = env.EnvironmentConfig.EnvOverrideValues
			showAppMetrics = *env.AppMetrics
		} else {
			showAppMetrics = appDeploymentTemplate.IsAppMetricsEnabled
			deploymentTemplateRaw = appDeploymentTemplate.DefaultAppOverride
		}
	} else {
		//on app level
		showAppMetrics = appDeploymentTemplate.IsAppMetricsEnabled
		deploymentTemplateRaw = appDeploymentTemplate.DefaultAppOverride
		chartRefId = chartRefData.LatestAppChartRef
	}

	var deploymentTemplateObj map[string]interface{}
	if deploymentTemplateRaw != nil {
		err = json.Unmarshal([]byte(deploymentTemplateRaw), &deploymentTemplateObj)
		if err != nil {
			handler.logger.Errorw("service err, un-marshaling fail in deploymentTemplate", "err", err, "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return nil, true
		}
	}

	deploymentTemplateResp := &appBean.DeploymentTemplate{
		ChartRefId:     chartRefId,
		Template:       deploymentTemplateObj,
		ShowAppMetrics: showAppMetrics,
	}

	return deploymentTemplateResp, false
}

//validate and build workflows
func (handler AppRestHandlerImpl) validateAndBuildAppWorkflows(w http.ResponseWriter, appId int) ([]*appBean.AppWorkflow, bool) {
	handler.logger.Debugw("Getting app detail - workflows", "appId", appId)

	workflowsList, err := handler.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		handler.logger.Errorw("error in fetching workflows for app in GetAppAllDetail", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	var appWorkflowsResp []*appBean.AppWorkflow
	for _, workflow := range workflowsList {

		workflowResp := &appBean.AppWorkflow{
			Name: workflow.Name,
		}

		var cdPipelinesResp []*appBean.CdPipelineDetails
		for _, workflowMappingDto := range workflow.AppWorkflowMappingDto {
			if workflowMappingDto.Type == appWorkflow2.CIPIPELINE {
				ciPipeline, err := handler.pipelineBuilder.GetCiPipelineById(workflowMappingDto.ComponentId)
				if err != nil {
					handler.logger.Errorw("service err, GetCiPipelineById in GetAppAllDetail", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, true
				}

				ciPipelineResp, err := handler.validateAndBuildCiPipelineResp(w, appId, ciPipeline)
				if err != nil {
					handler.logger.Errorw("service err, validateAndBuildCiPipelineResp in GetAppAllDetail", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, true
				}
				workflowResp.CiPipeline = ciPipelineResp
			}

			if workflowMappingDto.Type == appWorkflow2.CDPIPELINE {
				cdPipeline, err := handler.pipelineBuilder.GetCdPipelineById(workflowMappingDto.ComponentId)
				if err != nil {
					handler.logger.Errorw("service err, GetCdPipelineById in GetAppAllDetail", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, true
				}

				cdPipelineResp, err := handler.validateAndBuildCdPipelineResp(w, appId, cdPipeline)
				if err != nil {
					handler.logger.Errorw("service err, validateAndBuildCdPipelineResp in GetAppAllDetail", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, true
				}
				cdPipelinesResp = append(cdPipelinesResp, cdPipelineResp)
			}
		}
		workflowResp.CdPipelines = cdPipelinesResp
		appWorkflowsResp = append(appWorkflowsResp, workflowResp)
	}

	return appWorkflowsResp, false
}

//build ci pipeline resp
func (handler AppRestHandlerImpl) validateAndBuildCiPipelineResp(w http.ResponseWriter, appId int, ciPipeline *bean.CiPipeline) (*appBean.CiPipelineDetails, error) {
	handler.logger.Debugw("Getting app detail - build ci pipeline resp", "appId", appId)

	ciPipelineResp := &appBean.CiPipelineDetails{
		Name:                     ciPipeline.Name,
		IsManual:                 ciPipeline.IsManual,
		DockerBuildArgs:          ciPipeline.DockerArgs,
		VulnerabilityScanEnabled: ciPipeline.ScanEnabled,
	}

	//build ciPipelineMaterial resp
	var ciPipelineMaterialsConfig []*appBean.CiPipelineMaterialConfig
	for _, ciMaterial := range ciPipeline.CiMaterial {
		gitMaterial, err := handler.materialRepository.FindById(ciMaterial.GitMaterialId)
		if err != nil {
			handler.logger.Errorw("service err, GitMaterialById in GetAppAllDetail", "err", err, "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return nil, err
		}
		ciPipelineMaterialConfig := &appBean.CiPipelineMaterialConfig{
			Type:         ciMaterial.Source.Type,
			Value:        ciMaterial.Source.Value,
			GitRepoUrl:   gitMaterial.Url,
			CheckoutPath: gitMaterial.CheckoutPath,
		}
		ciPipelineMaterialsConfig = append(ciPipelineMaterialsConfig, ciPipelineMaterialConfig)
	}
	ciPipelineResp.CiPipelineMaterialsConfig = ciPipelineMaterialsConfig

	//build docker pre-build script
	var beforeDockerBuildScriptsResp []*appBean.BuildScript
	for _, beforeDockerBuildScript := range ciPipeline.BeforeDockerBuildScripts {
		beforeDockerBuildScriptResp := &appBean.BuildScript{
			Name:                beforeDockerBuildScript.Name,
			Index:               beforeDockerBuildScript.Index,
			Script:              beforeDockerBuildScript.Script,
			ReportDirectoryPath: beforeDockerBuildScript.OutputLocation,
		}
		beforeDockerBuildScriptsResp = append(beforeDockerBuildScriptsResp, beforeDockerBuildScriptResp)
	}
	ciPipelineResp.BeforeDockerBuildScripts = beforeDockerBuildScriptsResp

	//build docker post build script
	var afterDockerBuildScriptsResp []*appBean.BuildScript
	for _, afterDockerBuildScript := range ciPipeline.AfterDockerBuildScripts {
		afterDockerBuildScriptResp := &appBean.BuildScript{
			Name:                afterDockerBuildScript.Name,
			Index:               afterDockerBuildScript.Index,
			Script:              afterDockerBuildScript.Script,
			ReportDirectoryPath: afterDockerBuildScript.OutputLocation,
		}
		afterDockerBuildScriptsResp = append(afterDockerBuildScriptsResp, afterDockerBuildScriptResp)
	}
	ciPipelineResp.AfterDockerBuildScripts = afterDockerBuildScriptsResp

	return ciPipelineResp, nil
}

//build cd pipeline resp
func (handler AppRestHandlerImpl) validateAndBuildCdPipelineResp(w http.ResponseWriter, appId int, cdPipeline *bean.CDPipelineConfigObject) (*appBean.CdPipelineDetails, error) {
	handler.logger.Debugw("Getting app detail - build cd pipeline resp", "appId", appId)

	cdPipelineResp := &appBean.CdPipelineDetails{
		Name:              cdPipeline.Name,
		EnvironmentName:   cdPipeline.EnvironmentName,
		TriggerType:       cdPipeline.TriggerType,
		DeploymentType:    cdPipeline.DeploymentTemplate,
		RunPreStageInEnv:  cdPipeline.RunPreStageInEnv,
		RunPostStageInEnv: cdPipeline.RunPostStageInEnv,
		IsClusterCdActive: cdPipeline.CdArgoSetup,
	}

	//build DeploymentStrategies resp
	var deploymentTemplateStrategiesResp []*appBean.DeploymentStrategy
	for _, strategy := range cdPipeline.Strategies {
		deploymentTemplateStrategyResp := &appBean.DeploymentStrategy{
			DeploymentType: strategy.DeploymentTemplate,
			IsDefault:      strategy.Default,
		}
		var configObj map[string]interface{}
		if strategy.Config != nil {
			err := json.Unmarshal([]byte(strategy.Config), &configObj)
			if err != nil {
				handler.logger.Errorw("service err, un-marshaling fail in config object in cd", "err", err, "appId", appId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return nil, err
			}
		}
		deploymentTemplateStrategyResp.Config = configObj
		deploymentTemplateStrategiesResp = append(deploymentTemplateStrategiesResp, deploymentTemplateStrategyResp)
	}
	cdPipelineResp.DeploymentStrategies = deploymentTemplateStrategiesResp

	//set pre stage
	preStage := cdPipeline.PreStage
	cdPipelineResp.PreStage = &appBean.CdStage{
		TriggerType: preStage.TriggerType,
		Name:        preStage.Name,
		Config:      preStage.Config,
	}

	//set post stage
	postStage := cdPipeline.PostStage
	cdPipelineResp.PostStage = &appBean.CdStage{
		TriggerType: postStage.TriggerType,
		Name:        postStage.Name,
		Config:      postStage.Config,
	}

	//set pre stage config maps secret names
	preStageConfigMapSecretNames := cdPipeline.PreStageConfigMapSecretNames
	cdPipelineResp.PreStageConfigMapSecretNames = &appBean.CdStageConfigMapSecretNames{
		ConfigMaps: preStageConfigMapSecretNames.ConfigMaps,
		Secrets:    preStageConfigMapSecretNames.Secrets,
	}

	//set post stage config maps secret names
	postStageConfigMapSecretNames := cdPipeline.PostStageConfigMapSecretNames
	cdPipelineResp.PostStageConfigMapSecretNames = &appBean.CdStageConfigMapSecretNames{
		ConfigMaps: postStageConfigMapSecretNames.ConfigMaps,
		Secrets:    postStageConfigMapSecretNames.Secrets,
	}

	return cdPipelineResp, nil
}

//get/build global config maps
func (handler AppRestHandlerImpl) validateAndBuildAppGlobalConfigMaps(w http.ResponseWriter, appId int) ([]*appBean.ConfigMap, bool) {
	handler.logger.Debugw("Getting app detail - global config maps", "appId", appId)

	configMapData, err := handler.configMapService.CMGlobalFetch(appId)
	if err != nil {
		handler.logger.Errorw("service err, CMGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppConfigMaps(w, appId, 0, configMapData)
}

//get/build environment config maps
func (handler AppRestHandlerImpl) validateAndBuildAppEnvironmentConfigMaps(w http.ResponseWriter, appId int, envId int) ([]*appBean.ConfigMap, bool) {
	handler.logger.Debugw("Getting app detail - environment config maps", "appId", appId, "envId", envId)

	configMapData, err := handler.configMapService.CMEnvironmentFetch(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CMGlobalFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppConfigMaps(w, appId, envId, configMapData)
}

//get/build config maps
func (handler AppRestHandlerImpl) validateAndBuildAppConfigMaps(w http.ResponseWriter, appId int, envId int, configMapData *pipeline.ConfigDataRequest) ([]*appBean.ConfigMap, bool) {
	handler.logger.Debugw("Getting app detail - config maps", "appId", appId, "envId", envId)

	var configMapsResp []*appBean.ConfigMap
	if configMapData != nil && len(configMapData.ConfigData) > 0 {
		for _, configMap := range configMapData.ConfigData {

			//initialise
			configMapRes := &appBean.ConfigMap{
				Name:       configMap.Name,
				IsExternal: configMap.External,
				UsageType:  configMap.Type,
			}

			considerGlobalDefaultData := envId > 0 && configMap.Data == nil

			//set data
			var data json.RawMessage
			if considerGlobalDefaultData {
				data = configMap.DefaultData
			} else {
				data = configMap.Data
			}
			var dataObj map[string]interface{}
			if data != nil {
				err := json.Unmarshal([]byte(data), &dataObj)
				if err != nil {
					handler.logger.Errorw("service err, un-marshaling fail in config map", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, true
				}
			}
			configMapRes.Data = dataObj

			//set data volume usage type
			if configMap.Type == util.ConfigMapSecretUsageTypeVolume {
				dataVolumeUsageConfig := &appBean.ConfigMapSecretDataVolumeUsageConfig{
					FilePermission: configMap.FilePermission,
					SubPath:        configMap.SubPath,
				}
				if considerGlobalDefaultData {
					dataVolumeUsageConfig.MountPath = configMap.DefaultMountPath
				} else {
					dataVolumeUsageConfig.MountPath = configMap.MountPath
				}

				configMapRes.DataVolumeUsageConfig = dataVolumeUsageConfig
			}

			configMapsResp = append(configMapsResp, configMapRes)
		}
	}
	return configMapsResp, false
}

//get/build global secrets
func (handler AppRestHandlerImpl) validateAndBuildAppGlobalSecrets(w http.ResponseWriter, appId int) ([]*appBean.Secret, bool) {
	handler.logger.Debugw("Getting app detail - global secret", "appId", appId)

	secretData, err := handler.configMapService.CSGlobalFetch(appId)
	if err != nil {
		handler.logger.Errorw("service err, CSGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	var secretsResp []*appBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {

		for _, secretConfig := range secretData.ConfigData {
			secretDataWithData, err := handler.configMapService.CSGlobalFetchForEdit(secretConfig.Name, secretData.Id)
			if err != nil {
				handler.logger.Errorw("service err, CSGlobalFetch-CSGlobalFetchForEdit in GetAppAllDetail", "err", err, "appId", appId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return nil, true
			}

			secretRes, err := handler.validateAndBuildAppSecrets(w, appId, 0, secretDataWithData)
			if err != nil {
				handler.logger.Errorw("service err, CSGlobalFetch-validateAndBuildAppSecrets in GetAppAllDetail", "err", err, "appId", appId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return nil, true
			}

			for _, secret := range secretRes {
				secretsResp = append(secretsResp, secret)
			}
		}
	}

	return secretsResp, false
}

//get/build environment secrets
func (handler AppRestHandlerImpl) validateAndBuildAppEnvironmentSecrets(w http.ResponseWriter, appId int, envId int) ([]*appBean.Secret, bool) {
	handler.logger.Debugw("Getting app detail - env secrets", "appId", appId, "envId", envId)

	secretData, err := handler.configMapService.CSEnvironmentFetch(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CSEnvironmentFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	var secretsResp []*appBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {

		for _, secretConfig := range secretData.ConfigData {
			secretDataWithData, err := handler.configMapService.CSEnvironmentFetchForEdit(secretConfig.Name, secretData.Id, appId, envId)
			if err != nil {
				handler.logger.Errorw("service err, CSEnvironmentFetchForEdit in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return nil, true
			}

			secretRes, err := handler.validateAndBuildAppSecrets(w, appId, envId, secretDataWithData)
			if err != nil {
				handler.logger.Errorw("service err, CSGlobalFetch-validateAndBuildAppSecrets in GetAppAllDetail", "err", err, "appId", appId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return nil, true
			}

			for _, secret := range secretRes {
				secretsResp = append(secretsResp, secret)
			}
		}
	}

	return secretsResp, false
}

//get/build secrets
func (handler AppRestHandlerImpl) validateAndBuildAppSecrets(w http.ResponseWriter, appId int, envId int, secretData *pipeline.ConfigDataRequest) ([]*appBean.Secret, error) {
	handler.logger.Debugw("Getting app detail - secrets", "appId", appId, "envId", envId)

	var secretsResp []*appBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {
		for _, secret := range secretData.ConfigData {

			//initialise
			globalSecret := &appBean.Secret{
				Name:         secret.Name,
				RoleArn:      secret.RoleARN,
				IsExternal:   secret.External,
				UsageType:    secret.Type,
				ExternalType: secret.ExternalSecretType,
			}

			considerGlobalDefaultData := envId > 0 && secret.Data == nil

			//set data
			var data json.RawMessage
			var externalSecrets []pipeline.ExternalSecret
			if considerGlobalDefaultData {
				data = secret.DefaultData
				externalSecrets = secret.DefaultExternalSecret
			} else {
				data = secret.Data
				externalSecrets = secret.ExternalSecret
			}
			var dataObj map[string]interface{}
			if data != nil {
				err := json.Unmarshal([]byte(data), &dataObj)
				if err != nil {
					handler.logger.Errorw("service err, un-marshaling fail in secret", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, err
				}
			}
			globalSecret.Data = dataObj

			//set external data
			var externalSecretsResp []*appBean.ExternalSecret
			if len(externalSecrets) > 0 {
				for _, externalSecret := range externalSecrets {
					externalSecretsResp = append(externalSecretsResp, &appBean.ExternalSecret{
						Name:     externalSecret.Name,
						Key:      externalSecret.Key,
						Property: externalSecret.Property,
						IsBinary: externalSecret.IsBinary,
					})
				}
			}
			globalSecret.ExternalSecretData = externalSecretsResp

			//set data volume usage type
			if secret.Type == util.ConfigMapSecretUsageTypeVolume {
				globalSecret.DataVolumeUsageConfig = &appBean.ConfigMapSecretDataVolumeUsageConfig{
					SubPath:        secret.SubPath,
					FilePermission: secret.FilePermission,
				}
				if considerGlobalDefaultData {
					globalSecret.DataVolumeUsageConfig.MountPath = secret.DefaultMountPath
				} else {
					globalSecret.DataVolumeUsageConfig.MountPath = secret.MountPath
				}
			}

			secretsResp = append(secretsResp, globalSecret)
		}
	}
	return secretsResp, nil
}

//get/build environment overrides
func (handler AppRestHandlerImpl) validateAndBuildEnvironmentOverrides(w http.ResponseWriter, appId int, token string) (map[string]*appBean.EnvironmentOverride, bool) {
	handler.logger.Debugw("Getting app detail - env override", "appId", appId)

	appEnvironments, err := handler.appListingService.FetchOtherEnvironment(appId)
	if err != nil {
		handler.logger.Errorw("service err, Fetch app environments in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	environmentOverrides := make(map[string]*appBean.EnvironmentOverride)
	if len(appEnvironments) > 0 {
		for _, appEnvironment := range appEnvironments {

			envId := appEnvironment.EnvironmentId

			//check RBAC for environment
			object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
			if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionUpdate, object); !ok {
				handler.logger.Errorw("Unauthorized User for env update action", "err", err, "appId", appId, "envId", envId)
				writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
				return nil, true
			}
			//RBAC end

			envDeploymentTemplateResp, done := handler.validateAndBuildAppDeploymentTemplate(w, appId, envId)
			if done {
				return nil, true
			}

			envSecretsResp, done := handler.validateAndBuildAppEnvironmentSecrets(w, appId, envId)
			if done {
				return nil, true
			}

			envConfigMapsResp, done := handler.validateAndBuildAppEnvironmentConfigMaps(w, appId, envId)
			if done {
				return nil, true
			}

			environmentOverrides[appEnvironment.EnvironmentName] = &appBean.EnvironmentOverride{
				Secrets:            envSecretsResp,
				ConfigMaps:         envConfigMapsResp,
				DeploymentTemplate: envDeploymentTemplateResp,
			}
		}
	}
	return environmentOverrides, false
}

//GetApp related methods starts

//Create App related methods starts

//create a blank app with metadata
func (handler AppRestHandlerImpl) createBlankApp(w http.ResponseWriter, appMetadata *appBean.AppMetadata, userId int32) (*bean.CreateAppDTO, bool) {
	handler.logger.Infow("Create App - creating blank app", "appMetadata", appMetadata)

	//validating app metadata
	err := handler.validator.Struct(appMetadata)
	if err != nil {
		handler.logger.Errorw("validation err, AppMetadata in create app by API", "err", err, "AppMetadata", appMetadata)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, true
	}

	team, err := handler.teamRepository.FindByTeamName(appMetadata.ProjectName)
	if err != nil {
		handler.logger.Infow("no project found by name in CreateApp request by API")
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, true
	}

	handler.logger.Infow("Create App - creating blank app with metadata", "appMetadata", appMetadata)

	createAppRequestDTO := &bean.CreateAppDTO{
		AppName: appMetadata.AppName,
		TeamId:  team.Id,
		UserId:  userId,
	}
	for _, requestLabel := range appMetadata.Labels {
		appLabel := &bean.Label{
			Key:   requestLabel.Key,
			Value: requestLabel.Value,
		}
		createAppRequestDTO.AppLabels = append(createAppRequestDTO.AppLabels, appLabel)
	}

	createAppResp, err := handler.pipelineBuilder.CreateApp(createAppRequestDTO)
	if err != nil {
		handler.logger.Errorw("service err, CreateApp in CreateBlankApp", "err", err, "CreateApp", createAppRequestDTO)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}
	return createAppResp, false
}

//delete app
func (handler AppRestHandlerImpl) deleteApp(w http.ResponseWriter, ctx context.Context, appId int, userId int32) {
	handler.logger.Infow("Delete app", "appid", appId)

	//finding all workflows for app
	workflowsList, err := handler.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		handler.logger.Errorw("error in fetching workflows for app in DeleteApp", "err", err)
		writeJsonResp(w, err, APP_DELETE_FAILED_RESP, http.StatusInternalServerError)
		return
	}
	if len(workflowsList) > 0 {
		//deleting all ci, cd pipelines & workflows before deleting app

		cdPipelines, err := handler.pipelineBuilder.GetCdPipelinesForApp(appId)
		if err != nil {
			handler.logger.Errorw("service err, GetCdPipelines in DeleteApp", "err", err, "appId", appId)
			writeJsonResp(w, err, APP_DELETE_FAILED_RESP, http.StatusInternalServerError)
			return
		}

		for _, cdPipeline := range cdPipelines.Pipelines {
			cdPipelineDeleteRequest := &bean.CDPatchRequest{
				AppId:       appId,
				UserId:      userId,
				Action:      bean.CD_DELETE,
				ForceDelete: true,
				Pipeline:    cdPipeline,
			}
			//TODO update context - token
			_, err = handler.pipelineBuilder.PatchCdPipelines(cdPipelineDeleteRequest, ctx)
			if err != nil {
				handler.logger.Errorw("err in deleting cd pipeline in DeleteApp", "err", err, "payload", cdPipelineDeleteRequest)
				writeJsonResp(w, err, APP_DELETE_FAILED_RESP, http.StatusInternalServerError)
				return
			}
		}

		ciPipelines, err := handler.pipelineBuilder.GetCiPipeline(appId)
		if err != nil {
			handler.logger.Errorw("service err, GetCiPipelines in DeleteApp", "err", err, "appId", appId)
			writeJsonResp(w, err, APP_DELETE_FAILED_RESP, http.StatusInternalServerError)
			return
		}
		for _, ciPipeline := range ciPipelines.CiPipelines {
			//only deleting ci pipeline when it's an internal pipeline(not external)
			if !ciPipeline.IsExternal {
				ciPipelineDeleteRequest := &bean.CiPatchRequest{
					AppId:      appId,
					UserId:     userId,
					Action:     bean.DELETE,
					CiPipeline: ciPipeline,
				}

				_, err := handler.pipelineBuilder.PatchCiPipeline(ciPipelineDeleteRequest)
				if err != nil {
					handler.logger.Errorw("err in deleting ci pipeline in DeleteApp", "err", err, "payload", ciPipelineDeleteRequest)
					writeJsonResp(w, err, APP_DELETE_FAILED_RESP, http.StatusInternalServerError)
					return
				}
			}
		}
		for _, workflow := range workflowsList {
			err = handler.appWorkflowService.DeleteAppWorkflow(appId, workflow.Id, userId)
			if err != nil {
				handler.logger.Errorw("service err, DeleteAppWorkflow ")
				writeJsonResp(w, err, APP_DELETE_FAILED_RESP, http.StatusInternalServerError)
				return
			}
		}
	}
	err = handler.pipelineBuilder.DeleteApp(appId, userId)
	if err != nil {
		handler.logger.Errorw("service error, DeleteApp", "err", err, "appId", appId)
		writeJsonResp(w, err, APP_DELETE_FAILED_RESP, http.StatusInternalServerError)
		return
	}
}

//create git materials
func (handler AppRestHandlerImpl) createGitMaterials(w http.ResponseWriter, appId int, gitMaterials []*appBean.GitMaterial, userId int32) bool {
	handler.logger.Infow("Create App - creating git materials", "appId", appId, "GitMaterials", gitMaterials)

	createMaterialRequestDto := &bean.CreateMaterialDTO{
		AppId:  appId,
		UserId: userId,
	}

	for _, material := range gitMaterials {
		err := handler.validator.Struct(material)
		if err != nil {
			handler.logger.Errorw("validation err, gitMaterial in CreateGitMaterials", "err", err, "GitMaterial", material)
			writeJsonResp(w, err, nil, http.StatusBadRequest)
			return true
		}

		//finding gitProvider to update gitMaterial
		gitProvider, err := handler.gitProviderRepo.FindByUrl(material.GitProviderUrl)
		if err != nil {
			handler.logger.Errorw("service err, FindByUrl in CreateGitMaterials", "err", err, "gitProviderUrl", material.GitProviderUrl)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}

		//validating git material by git provider auth mode
		var hasPrefixResult bool
		if gitProvider.AuthMode == repository.AUTH_MODE_SSH {
			hasPrefixResult = strings.HasPrefix(material.GitRepoUrl, SSH_URL_PREFIX)
		} else {
			hasPrefixResult = strings.HasPrefix(material.GitRepoUrl, HTTPS_URL_PREFIX)
		}
		if !hasPrefixResult {
			handler.logger.Errorw("validation err, CreateGitMaterials : invalid git material url", "err", err, "gitMaterialUrl", material.GitRepoUrl)
			writeJsonResp(w, fmt.Errorf("validation for url failed"), nil, http.StatusBadRequest)
			return true
		}

		gitMaterialRequest := &bean.GitMaterial{
			Url:             material.GitRepoUrl,
			GitProviderId:   gitProvider.Id,
			CheckoutPath:    material.CheckoutPath,
			FetchSubmodules: material.FetchSubmodules,
		}

		createMaterialRequestDto.Material = append(createMaterialRequestDto.Material, gitMaterialRequest)
	}
	_, err := handler.pipelineBuilder.CreateMaterialsForApp(createMaterialRequestDto)
	if err != nil {
		handler.logger.Errorw("service err, CreateMaterialsForApp in CreateGitMaterials", "err", err, "CreateMaterial", createMaterialRequestDto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	return false
}

//create docker config
func (handler AppRestHandlerImpl) createDockerConfig(w http.ResponseWriter, appId int, dockerConfig *appBean.DockerConfig, userId int32) bool {
	handler.logger.Infow("Create App - creating docker config", "appId", appId, "DockerConfig", dockerConfig)

	createDockerConfigRequest := &bean.CiConfigRequest{
		AppId:            appId,
		UserId:           userId,
		DockerRegistry:   dockerConfig.DockerRegistry,
		DockerRepository: dockerConfig.DockerRepository,
	}

	//finding gitMaterial by appId and checkoutPath
	gitMaterial, err := handler.materialRepository.FindByAppIdAndCheckoutPath(appId, dockerConfig.BuildConfig.GitCheckoutPath)
	if err != nil {
		handler.logger.Errorw("service err, FindByAppIdAndCheckoutPath in CreateDockerConfig", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	dockerBuildArgs := make(map[string]string)
	if dockerConfig.BuildConfig.Args != nil {
		dockerBuildArgs = dockerConfig.BuildConfig.Args
	}
	dockerBuildConfigRequest := &bean.DockerBuildConfig{
		GitMaterialId:  gitMaterial.Id,
		DockerfilePath: dockerConfig.BuildConfig.DockerfileRelativePath,
		Args:           dockerBuildArgs,
	}
	createDockerConfigRequest.DockerBuildConfig = dockerBuildConfigRequest
	_, err = handler.pipelineBuilder.CreateCiPipeline(createDockerConfigRequest)
	if err != nil {
		handler.logger.Errorw("service err, CreateCiPipeline in CreateDockerConfig", "err", err, "createRequest", createDockerConfigRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	return false
}

//create global template
func (handler AppRestHandlerImpl) createDeploymentTemplate(w http.ResponseWriter, ctx context.Context, appId int, deploymentTemplate *appBean.DeploymentTemplate, userId int32) bool {
	handler.logger.Infow("Create App - creating deployment template", "appId", appId, "DeploymentTemplate", deploymentTemplate)

	createDeploymentTemplateRequest := pipeline.TemplateRequest{
		AppId:               appId,
		ChartRefId:          deploymentTemplate.ChartRefId,
		IsAppMetricsEnabled: deploymentTemplate.ShowAppMetrics,
		UserId:              userId,
	}

	//marshalling template
	template, err := json.Marshal(deploymentTemplate.Template)
	if err != nil {
		handler.logger.Errorw("service err, could not json marshal template in CreateDeploymentTemplate", "err", err, "appId", appId, "template", deploymentTemplate.Template)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	templateRequest := json.RawMessage(template)
	createDeploymentTemplateRequest.ValuesOverride = templateRequest
	//TODO update context - token
	//creating deployment template
	_, err = handler.chartService.Create(createDeploymentTemplateRequest, ctx)
	if err != nil {
		handler.logger.Errorw("service err, Create in CreateDeploymentTemplate", "err", err, "createRequest", createDeploymentTemplateRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}

	appMetricsRequest := pipeline.AppMetricEnableDisableRequest{
		AppId:               appId,
		UserId:              userId,
		IsAppMetricsEnabled: deploymentTemplate.ShowAppMetrics,
	}
	//updating app metrics
	_, err = handler.chartService.AppMetricsEnableDisable(appMetricsRequest)
	if err != nil {
		handler.logger.Errorw("service err, AppMetricsEnableDisable in createDeploymentTemplate", "err", err, "appId", appId, "payload", appMetricsRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	return false
}

//create global CMs
func (handler AppRestHandlerImpl) createGlobalConfigMaps(w http.ResponseWriter, appId int, userId int32, configMaps []*appBean.ConfigMap) bool {
	handler.logger.Infow("Create App - creating global configMap", "appId", appId)

	for _, configMap := range configMaps {
		//getting app level by app id
		appLevel, err := handler.configMapRepository.GetByAppIdAppLevel(appId)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("error in getting app level by app id in createGlobalConfigMaps", "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		var appLevelId int
		if appLevel != nil {
			appLevelId = appLevel.Id
		}
		configMapRequest := &pipeline.ConfigDataRequest{
			AppId:  appId,
			UserId: userId,
			Id:     appLevelId,
		}
		//marshalling configMap data, i.e. key-value pairs
		configMapKeyValueData, err := json.Marshal(configMap.Data)
		if err != nil {
			handler.logger.Errorw("service err, could not json marshal configMap data in CreateGlobalConfigMap", "err", err, "appId", appId, "configMapData", configMap.Data)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
		var configMapDataRequest []*pipeline.ConfigData
		configMapData := &pipeline.ConfigData{
			Name:     configMap.Name,
			External: configMap.IsExternal,
			Data:     json.RawMessage(configMapKeyValueData),
			Type:     configMap.UsageType,
		}
		dataVolumeUsageConfig := configMap.DataVolumeUsageConfig
		if dataVolumeUsageConfig != nil {
			configMapData.MountPath = dataVolumeUsageConfig.MountPath
			configMapData.SubPath = dataVolumeUsageConfig.SubPath
			configMapData.FilePermission = dataVolumeUsageConfig.FilePermission
		}
		configMapDataRequest = append(configMapDataRequest, configMapData)
		configMapRequest.ConfigData = configMapDataRequest
		//using same var for every request, since appId and userID are same
		_, err = handler.configMapService.CMGlobalAddUpdate(configMapRequest)
		if err != nil {
			handler.logger.Errorw("service err, CMGlobalAddUpdate in CreateGlobalConfigMap", "err", err, "appId", appId, "configMapRequest", configMapRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
	}
	return false

}

//create global secrets
func (handler AppRestHandlerImpl) createGlobalSecrets(w http.ResponseWriter, appId int, userId int32, secrets []*appBean.Secret) bool {
	handler.logger.Infow("Create App - creating global secrets", "appId", appId)

	for _, secret := range secrets {
		//getting app level by app id
		appLevel, err := handler.configMapRepository.GetByAppIdAppLevel(appId)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("error in getting app level by app id in createGlobalSecrets", "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		var appLevelId int
		if appLevel != nil {
			appLevelId = appLevel.Id
		}
		secretRequest := &pipeline.ConfigDataRequest{
			AppId:  appId,
			UserId: userId,
			Id:     appLevelId,
		}
		var secretDataRequest []*pipeline.ConfigData
		secretData := &pipeline.ConfigData{
			Name:               secret.Name,
			External:           secret.IsExternal,
			Type:               secret.UsageType,
			ExternalSecretType: secret.ExternalType,
			RoleARN:            secret.RoleArn,
		}
		dataVolumeUsageConfig := secret.DataVolumeUsageConfig
		if dataVolumeUsageConfig != nil {
			secretData.MountPath = dataVolumeUsageConfig.MountPath
			secretData.SubPath = dataVolumeUsageConfig.SubPath
			secretData.FilePermission = dataVolumeUsageConfig.FilePermission
		}
		if secret.IsExternal {
			var externalDataRequests []pipeline.ExternalSecret
			for _, externalData := range secret.ExternalSecretData {
				externalDataRequest := pipeline.ExternalSecret{
					Name:     externalData.Name,
					IsBinary: externalData.IsBinary,
					Key:      externalData.Key,
					Property: externalData.Property,
				}
				externalDataRequests = append(externalDataRequests, externalDataRequest)
			}
		} else {
			secretKeyValueData, err := json.Marshal(secret.Data)
			if err != nil {
				handler.logger.Errorw("service err, could not json marshal secret data in CreateGlobalSecret", "err", err, "appId", appId, "secretData", secret.Data)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return true
			}
			secretData.Data = secretKeyValueData
		}
		secretDataRequest = append(secretDataRequest, secretData)
		secretRequest.ConfigData = secretDataRequest
		//using same var for every request, since appId and userID are same
		_, err = handler.configMapService.CSGlobalAddUpdate(secretRequest)
		if err != nil {
			handler.logger.Errorw("service err, CSGlobalAddUpdate in CreateGlobalSecret", "err", err, "appId", appId, "secretRequest", secretRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
	}
	return false
}

//create app workflows
func (handler AppRestHandlerImpl) createWorkflows(w http.ResponseWriter, ctx context.Context, appId int, userId int32, workflows []*appBean.AppWorkflow) bool {
	handler.logger.Infow("Create App - creating workflows", "appId", appId, "workflows", workflows)
	for _, workflow := range workflows {
		//Create workflow starts
		wf := &appWorkflow2.AppWorkflow{
			Name:   workflow.Name,
			AppId:  appId,
			Active: true,
			AuditLog: models.AuditLog{
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedBy: userId,
			},
		}
		savedAppWf, err := handler.appWorkflowRepository.SaveAppWorkflow(wf)
		if err != nil {
			handler.logger.Errorw("err in saving new workflow", err, "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
		workflowId := savedAppWf.Id
		//Creating workflow ends

		//Creating CI pipeline starts
		ciPipelineRequest := &bean.CiPatchRequest{
			AppId:         appId,
			UserId:        userId,
			AppWorkflowId: workflowId,
			Action:        bean.CREATE,
		}
		ciPipelineData := workflow.CiPipeline
		var ciMaterialsRequest []*bean.CiMaterial
		for _, ciMaterial := range ciPipelineData.CiPipelineMaterialsConfig {
			//finding gitMaterial by appId and checkoutPath
			gitMaterial, err := handler.materialRepository.FindByAppIdAndCheckoutPath(appId, ciMaterial.CheckoutPath)
			if err != nil {
				handler.logger.Errorw("service err, FindByAppIdAndCheckoutPath in CreateWorkflows", "err", err, "appId", appId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return true
			}
			if ciMaterial.GitRepoUrl != gitMaterial.Url {
				handler.logger.Errorw("error in finding git material for given ciMaterial config", "appId", appId, "ciMaterial", ciMaterial)
				writeJsonResp(w, fmt.Errorf("no git material found"), nil, http.StatusInternalServerError)
				return true
			}
			ciMaterialRequest := &bean.CiMaterial{
				GitMaterialId:   gitMaterial.Id,
				GitMaterialName: gitMaterial.Name,
				Source: &bean.SourceTypeConfig{
					Type:  ciMaterial.Type,
					Value: ciMaterial.Value,
				},
				CheckoutPath: gitMaterial.CheckoutPath,
			}
			ciMaterialsRequest = append(ciMaterialsRequest, ciMaterialRequest)
		}
		ciPipelineRequestData := &bean.CiPipeline{
			Name:                     ciPipelineData.Name,
			IsManual:                 ciPipelineData.IsManual,
			IsExternal:               false, //since app create api only supports internal pipelines currently
			Active:                   true,
			AfterDockerBuildScripts:  convertCiBuildScripts(ciPipelineData.AfterDockerBuildScripts),
			BeforeDockerBuildScripts: convertCiBuildScripts(ciPipelineData.BeforeDockerBuildScripts),
			DockerArgs:               ciPipelineData.DockerBuildArgs,
			ScanEnabled:              ciPipelineData.VulnerabilityScanEnabled,
			CiMaterial:               ciMaterialsRequest,
		}
		ciPipelineRequest.CiPipeline = ciPipelineRequestData
		_, err = handler.pipelineBuilder.PatchCiPipeline(ciPipelineRequest)
		if err != nil {
			handler.logger.Errorw("service err, PatchCiPipelines", "err", err, "PatchCiPipelines", ciPipelineRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
		//Creating CI pipeline ends

		//Creating CD pipeline starts
		//getting app workflow mapping for finding ciPipeline ID
		appWorkflowMapping, err := handler.appWorkflowRepository.FindByWorkflowId(workflowId)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("err in finding app workflow mapping", err, "appId", appId, "workflowId", workflowId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
		var ciPipelineId int
		for _, workflowMapping := range appWorkflowMapping {
			if workflowMapping.Type == appWorkflow2.CIPIPELINE {
				ciPipelineId = workflowMapping.ComponentId
				break
			}
		}

		cdPipelinesRequest := &bean.CdPipelines{
			AppId:  appId,
			UserId: userId,
		}
		var cdPipelineRequestConfigs []*bean.CDPipelineConfigObject
		for _, cdPipeline := range workflow.CdPipelines {
			//getting environment ID by name
			envModel, err := handler.environmentRepository.FindByName(cdPipeline.EnvironmentName)
			if err != nil {
				handler.logger.Errorw("err in fetching environment details by name", "appId", appId, "envName", cdPipeline.EnvironmentName)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return true
			}

			cdPipelineRequestConfig := &bean.CDPipelineConfigObject{
				Name:                          cdPipeline.Name,
				EnvironmentId:                 envModel.Id,
				AppWorkflowId:                 workflowId,
				CiPipelineId:                  ciPipelineId,
				DeploymentTemplate:            cdPipeline.DeploymentType,
				TriggerType:                   cdPipeline.TriggerType,
				CdArgoSetup:                   cdPipeline.IsClusterCdActive,
				RunPreStageInEnv:              cdPipeline.RunPreStageInEnv,
				RunPostStageInEnv:             cdPipeline.RunPostStageInEnv,
				PreStage:                      convertCdStages(cdPipeline.PreStage),
				PostStage:                     convertCdStages(cdPipeline.PostStage),
				PreStageConfigMapSecretNames:  convertCdPreStageCMorCSNames(cdPipeline.PostStageConfigMapSecretNames),
				PostStageConfigMapSecretNames: convertCdPostStageCMorCSNames(cdPipeline.PostStageConfigMapSecretNames),
			}
			convertedDeploymentStrategies, err := convertCdDeploymentStrategies(cdPipeline.DeploymentStrategies)
			if err != nil {
				handler.logger.Errorw("err in converting deployment strategies for creating cd pipeline", "appId", appId, "Strategies", cdPipeline.DeploymentStrategies)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return true
			}
			cdPipelineRequestConfig.Strategies = convertedDeploymentStrategies
			cdPipelineRequestConfigs = append(cdPipelineRequestConfigs, cdPipelineRequestConfig)
		}
		cdPipelinesRequest.Pipelines = cdPipelineRequestConfigs
		//TODO update context - token
		_, err = handler.pipelineBuilder.CreateCdPipelines(cdPipelinesRequest, ctx)
		if err != nil {
			handler.logger.Errorw("service err, CreateCdPipeline", "err", err, "payload", cdPipelinesRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
		//Creating CD pipeline ends
	}
	return false
}

//create environment overrides
func (handler AppRestHandlerImpl) createEnvOverrides(w http.ResponseWriter, ctx context.Context, appId int, userId int32, environmentOverrides map[string]*appBean.EnvironmentOverride) bool {
	handler.logger.Infow("Create App - creating env overrides", "appId", appId, "envOverrides", environmentOverrides)
	for envName, envOverrideValues := range environmentOverrides {
		envModel, err := handler.environmentRepository.FindByName(envName)
		if err != nil {
			handler.logger.Errorw("err in fetching environment details by name in CreateEnvOverrides", "appId", appId, "envName", envName)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}

		//creating deployment template override
		done := handler.createEnvDeploymentTemplate(w, ctx, appId, userId, envOverrideValues.DeploymentTemplate, envModel)
		if done {
			return done
		}

		//creating configMap override
		done = handler.createEnvCM(w, appId, userId, envModel, envOverrideValues.ConfigMaps)
		if done {
			return done
		}

		//creating secrets override
		done = handler.createEnvSecret(w, appId, userId, envModel, envOverrideValues.Secrets)
		if done {
			return done
		}
	}
	return false
}

//create template overrides
func (handler AppRestHandlerImpl) createEnvDeploymentTemplate(w http.ResponseWriter, ctx context.Context, appId int, userId int32, templateOverride *appBean.DeploymentTemplate, envModel *cluster.Environment) bool {
	handler.logger.Infow("Create App - creating template override", "appId", appId, "templateOverride", templateOverride)

	//finding env properties for appId & envId (this get created when cd pipeline is created)
	envProperties, err := handler.propertiesConfigService.GetEnvironmentProperties(appId, envModel.Id, templateOverride.ChartRefId)
	if err != nil {
		handler.logger.Errorw("service err, GetEnvConfOverride in createEnvDeploymentTemplate", "err", err, "appId", appId,"envId", envModel.Id, "chartRefId",templateOverride.ChartRefId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	envConfigPropertiesRequest, err := buildEnvTemplateOverrideRequest(templateOverride, envModel, envProperties, userId)
	if err != nil {
		handler.logger.Errorw("err in converting template config for creating env override", "appId", appId, "templateOverride", templateOverride)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	_, err = handler.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigPropertiesRequest)
	if err != nil {
		if err.Error() == bean2.NOCHARTEXIST {
			templateRequest := pipeline.TemplateRequest{
				AppId:          appId,
				ChartRefId:     envConfigPropertiesRequest.ChartRefId,
				ValuesOverride: []byte("{}"),
				UserId:         userId,
			}
			_, err = handler.chartService.CreateChartFromEnvOverride(templateRequest, ctx)
			if err != nil {
				handler.logger.Errorw("err in creating chart from env override in createEnvDeploymentTemplate","err",err,"payload",templateRequest)
				writeJsonResp(w,err,nil,http.StatusInternalServerError)
				return true
			}
			_, err = handler.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigPropertiesRequest)
			if err!=nil{
				handler.logger.Errorw("err in creating env properties in createEnvDeploymentTemplate","err",err,"payload",templateRequest)
				writeJsonResp(w,err,nil,http.StatusInternalServerError)
				return true
			}
		}
	}
	//_, err = handler.propertiesConfigService.UpdateEnvironmentProperties(appId, envConfigPropertiesRequest, userId)
	//if err != nil {
	//	handler.logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigPropertiesRequest)
	//	writeJsonResp(w, err, nil, http.StatusInternalServerError)
	//	return true
	//}

	//_, err = handler.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigProperties)
	//if err != nil {
	//	if err.Error() == bean2.NOCHARTEXIST {
	//		//TODO update context - token
	//		//ctx, cancel := context.WithCancel(r.Context())
	//		//if cn, ok := w.(http.CloseNotifier); ok {
	//		//	go func(done <-chan struct{}, closed <-chan bool) {
	//		//		select {
	//		//		case <-done:
	//		//		case <-closed:
	//		//			cancel()
	//		//		}
	//		//	}(ctx.Done(), cn.CloseNotify())
	//		//}
	//		//ctx = context.WithValue(r.Context(), "token", token)
	//		templateRequest := pipeline.TemplateRequest{
	//			AppId:          appId,
	//			ChartRefId:     templateOverride.ChartRefId,
	//			ValuesOverride: []byte("{}"),
	//			UserId:         userId,
	//		}
	//
	//		_, err = handler.chartService.CreateChartFromEnvOverride(templateRequest, ctx)
	//		if err != nil {
	//			handler.logger.Errorw("service err, CreateChartFromEnvOverride in CreateEnvDeploymentTemplate", "err", err, "payload", envConfigProperties)
	//			writeJsonResp(w, err, nil, http.StatusInternalServerError)
	//			return true
	//		}
	//		_, err = handler.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigProperties)
	//		if err != nil {
	//			handler.logger.Errorw("service err, CreateChartFromEnvOverride in CreateEnvDeploymentTemplate", "err", err, "payload", envConfigProperties)
	//			writeJsonResp(w, err, nil, http.StatusInternalServerError)
	//			return true
	//		}
	//	} else {
	//		handler.logger.Errorw("service err, CreateEnvDeploymentTemplate", "err", err, "payload", envConfigProperties)
	//		writeJsonResp(w, err, nil, http.StatusInternalServerError)
	//		return true
	//	}
	//}
	//updating app metrics
	appMetricsRequest := &pipeline.AppMetricEnableDisableRequest{
		AppId:               appId,
		UserId:              userId,
		EnvironmentId:       envModel.Id,
		IsAppMetricsEnabled: templateOverride.ShowAppMetrics,
	}
	_, err = handler.propertiesConfigService.EnvMetricsEnableDisable(appMetricsRequest)
	if err != nil {
		handler.logger.Errorw("service err, EnvMetricsEnableDisable", "err", err, "appId", appId, "environmentId", envModel.Id, "payload", appMetricsRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return true
	}
	return false
}

//create CM overrides
func (handler AppRestHandlerImpl) createEnvCM(w http.ResponseWriter, appId int, userId int32, envModel *cluster.Environment, CmOverrides []*appBean.ConfigMap) bool {
	handler.logger.Infow("Create App - creating CM override", "appId", appId, "CmOverrides", CmOverrides)

	for _, cmOverride := range CmOverrides {
		//getting env level by app id
		envLevel, err := handler.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envModel.Id)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("error in getting app level by app id in createEnvCM", "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		var envLevelId int
		if envLevel != nil {
			envLevelId = envLevel.Id
		}
		cmEnvRequest := &pipeline.ConfigDataRequest{
			AppId:         appId,
			UserId:        userId,
			EnvironmentId: envModel.Id,
			Id:            envLevelId,
		}

		cmOverrideData, err := json.Marshal(cmOverride.Data)
		if err != nil {
			handler.logger.Errorw("service err, could not json marshal template in CreateEnvCM", "err", err, "appId", appId, "cmOverrideData", cmOverride.Data)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
		configData := &pipeline.ConfigData{
			Name:     cmOverride.Name,
			External: cmOverride.IsExternal,
			Type:     cmOverride.UsageType,
			Data:     cmOverrideData,
		}
		cmOverrideDataVolumeUsageConfig := cmOverride.DataVolumeUsageConfig
		if cmOverrideDataVolumeUsageConfig != nil {
			configData.MountPath = cmOverrideDataVolumeUsageConfig.MountPath
			configData.SubPath = cmOverrideDataVolumeUsageConfig.SubPath
			configData.FilePermission = cmOverrideDataVolumeUsageConfig.FilePermission
		}
		var configDataRequest []*pipeline.ConfigData
		configDataRequest = append(configDataRequest, configData)
		cmEnvRequest.ConfigData = configDataRequest
		_, err = handler.configMapService.CMEnvironmentAddUpdate(cmEnvRequest)
		if err != nil {
			handler.logger.Errorw("service err, CMEnvironmentAddUpdate in CreateEnvCM", "err", err, "payload", cmEnvRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
	}
	return false
}

//create secret overrides
func (handler AppRestHandlerImpl) createEnvSecret(w http.ResponseWriter, appId int, userId int32, envModel *cluster.Environment, secretOverrides []*appBean.Secret) bool {
	handler.logger.Infow("Create App - creating secret overrides", "appId", appId, "secretOverrides", secretOverrides)

	for _, secretOverride := range secretOverrides {
		//getting env level by app id
		envLevel, err := handler.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envModel.Id)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("error in getting app level by app id in createEnvSecret", "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		var envLevelId int
		if envLevel != nil {
			envLevelId = envLevel.Id
		}
		secretEnvRequest := &pipeline.ConfigDataRequest{
			AppId:         appId,
			UserId:        userId,
			EnvironmentId: envModel.Id,
			Id:            envLevelId,
		}
		secretOverrideData, err := json.Marshal(secretOverride.Data)
		if err != nil {
			handler.logger.Errorw("service err, could not json marshal template in CreateEnvSecret", "err", err, "appId", appId, "secretOverrideData", secretOverride.Data)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
		secretData := &pipeline.ConfigData{
			Name:               secretOverride.Name,
			External:           secretOverride.IsExternal,
			ExternalSecretType: secretOverride.ExternalType,
			Type:               secretOverride.UsageType,
			Data:               secretOverrideData,
			RoleARN:            secretOverride.RoleArn,
			ExternalSecret:     convertCSExternalSecretData(secretOverride.ExternalSecretData),
		}
		secretOverrideDataVolumeUsageConfig := secretOverride.DataVolumeUsageConfig
		if secretOverrideDataVolumeUsageConfig != nil {
			secretData.MountPath = secretOverrideDataVolumeUsageConfig.MountPath
			secretData.SubPath = secretOverrideDataVolumeUsageConfig.SubPath
			secretData.FilePermission = secretOverrideDataVolumeUsageConfig.FilePermission
		}
		var secretDataRequest []*pipeline.ConfigData
		secretDataRequest = append(secretDataRequest, secretData)
		secretEnvRequest.ConfigData = secretDataRequest
		_, err = handler.configMapService.CSEnvironmentAddUpdate(secretEnvRequest)
		if err != nil {
			handler.logger.Errorw("service err, CSEnvironmentAddUpdate", "err", err, "payload", secretEnvRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return true
		}
	}
	return false
}

//Create App related methods ends

//private methods for data conversion below

func convertCSExternalSecretData(externalSecretsData []*appBean.ExternalSecret) []pipeline.ExternalSecret {
	var convertedExternalSecretsData []pipeline.ExternalSecret
	for _, externalSecretData := range externalSecretsData {
		convertedExternalSecret := pipeline.ExternalSecret{
			Key:      externalSecretData.Key,
			Name:     externalSecretData.Name,
			Property: externalSecretData.Property,
			IsBinary: externalSecretData.IsBinary,
		}
		convertedExternalSecretsData = append(convertedExternalSecretsData, convertedExternalSecret)
	}
	return convertedExternalSecretsData
}

func convertCiBuildScripts(buildScripts []*appBean.BuildScript) []*bean.CiScript {
	var convertedBuildScripts []*bean.CiScript
	for _, buildScript := range buildScripts {
		convertedBuildScript := &bean.CiScript{
			Index:          buildScript.Index,
			Name:           buildScript.Name,
			Script:         buildScript.Script,
			OutputLocation: buildScript.ReportDirectoryPath,
		}
		convertedBuildScripts = append(convertedBuildScripts, convertedBuildScript)
	}
	return convertedBuildScripts
}

func convertCdStages(cdStages *appBean.CdStage) bean.CdStage {
	convertedCdStage := bean.CdStage{
		TriggerType: cdStages.TriggerType,
		Name:        cdStages.Name,
		Config:      cdStages.Config,
	}
	return convertedCdStage
}

func convertCdPreStageCMorCSNames(preStageNames *appBean.CdStageConfigMapSecretNames) bean.PreStageConfigMapSecretNames {
	convertPreStageNames := bean.PreStageConfigMapSecretNames{
		ConfigMaps: preStageNames.ConfigMaps,
		Secrets:    preStageNames.Secrets,
	}
	return convertPreStageNames
}

func convertCdPostStageCMorCSNames(postStageNames *appBean.CdStageConfigMapSecretNames) bean.PostStageConfigMapSecretNames {
	convertPostStageNames := bean.PostStageConfigMapSecretNames{
		ConfigMaps: postStageNames.ConfigMaps,
		Secrets:    postStageNames.Secrets,
	}
	return convertPostStageNames
}

func convertCdDeploymentStrategies(deploymentStrategies []*appBean.DeploymentStrategy) ([]bean.Strategy, error) {
	var convertedStrategies []bean.Strategy
	for _, deploymentStrategy := range deploymentStrategies {
		convertedStrategy := bean.Strategy{
			DeploymentTemplate: deploymentStrategy.DeploymentType,
			Default:            deploymentStrategy.IsDefault,
		}
		strategyConfig, err := json.Marshal(deploymentStrategy.Config)
		if err != nil {
			return nil, err
		}
		convertedStrategy.Config = strategyConfig
		convertedStrategies = append(convertedStrategies, convertedStrategy)
	}
	return convertedStrategies, nil
}

func buildEnvTemplateOverrideRequest(templateOverride *appBean.DeploymentTemplate, envModel *cluster.Environment, envProperties *pipeline.EnvironmentPropertiesResponse, userId int32) (*pipeline.EnvironmentProperties, error) {
	template, err := json.Marshal(templateOverride.Template)
	if err != nil {
		return nil, err
	}
	envTemplateOverrideRequest := &pipeline.EnvironmentProperties{
		ChartRefId: templateOverride.ChartRefId,
		AppMetrics: &templateOverride.ShowAppMetrics,
		EnvOverrideValues: template,
		EnvironmentId: envModel.Id,
		EnvironmentName: envModel.Name,
		Id: envProperties.EnvironmentConfig.Id,
		Namespace: envModel.Namespace,
		Status: envProperties.EnvironmentConfig.Status,
		ManualReviewed: envProperties.EnvironmentConfig.ManualReviewed,
		Active: envProperties.EnvironmentConfig.Active,
		IsOverride: true,
		UserId: userId,
	}
	return envTemplateOverrideRequest, nil
}
