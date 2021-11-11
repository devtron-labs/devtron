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
	"encoding/json"
	"errors"
	appBean "github.com/devtron-labs/devtron/api/appbean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type AppRestHandler interface {
	GetAppAllDetail(w http.ResponseWriter, r *http.Request)
	//CreateApp(w http.ResponseWriter, r *http.Request)
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
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService, validator *validator.Validate, enforcerUtil rbac.EnforcerUtil,
	enforcer rbac.Enforcer, appLabelService app.AppLabelService, pipelineBuilder pipeline.PipelineBuilder, gitRegistryService pipeline.GitRegistryConfig,
	chartService pipeline.ChartService, configMapService pipeline.ConfigMapService, appListingService app.AppListingService, propertiesConfigService pipeline.PropertiesConfigService, appWorkflowService appWorkflow.AppWorkflowService, materialRepository pipelineConfig.MaterialRepository) *AppRestHandlerImpl {
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
	}
	return handler
}

const (
	CIPIPELINE string = "CI_PIPELINE"
	CDPIPELINE string = "CD_PIPELINE"
)

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

	//rback implementation for app
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback implementation ends here for app

	// get/build app metadata starts
	appMetadataResp, done := handler.validateAndBuildAppMetadata(w, appId)
	if done {
		return
	}
	// get/build app metadata ends

	// get/build git materials starts
	gitMaterialsResp, done := handler.validateAndBuildAppGitMaterials(w, appId)
	if done {
		return
	}
	// get/build git materials ends

	// get/build global deployment template starts
	globalDeploymentTemplateResp, done := handler.validateAndBuildAppDeploymentTemplate(w, appId, 0)
	if done {
		return
	}
	// get/build global deployment template ends

	// get/build global config maps starts
	globalConfigMapsResp, done := handler.validateAndBuildAppGlobalConfigMaps(w, appId)
	if done {
		return
	}
	// get/build global config maps ends

	// get/build global secrets starts
	globalSecretsResp, done := handler.validateAndBuildAppGlobalSecrets(w, appId)
	if done {
		return
	}
	// get/build global secrets ends

	// get/build environment override starts
	environmentOverrides, done := handler.validateAndBuildEnvironmentOverrides(w, appId)
	if done {
		return
	}

	// get/build environment override ends

	// get/build app workflows starts
	appWorkflows, done := handler.validateAndBuildAppWorkflows(w, appId)
	if done {
		return
	}
	// get/build app workflows ends

	// get/build docker config starts
	dockerConfig, done := handler.validateAndBuildDockerConfig(w, appId)
	if done {
		return
	}
	// get/build docker config ends

	// build full object for response
	appDetail := &appBean.AppDetail{
		Metadata:                 appMetadataResp,
		GitMaterials:             gitMaterialsResp,
		GlobalDeploymentTemplate: globalDeploymentTemplateResp,
		GlobalConfigMaps:         globalConfigMapsResp,
		GlobalSecrets:            globalSecretsResp,
		EnvironmentOverrides:     environmentOverrides,
		AppWorkflows:             appWorkflows,
		DockerConfig:             dockerConfig,
	}
	// end

	writeJsonResp(w, nil, appDetail, http.StatusOK)
}

//get/build app metadata
func (handler AppRestHandlerImpl) validateAndBuildAppMetadata(w http.ResponseWriter, appId int) (*appBean.AppMetadata, bool) {
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
				GitUrl:          gitMaterial.Url,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				GitAccountUrl:   gitRegistry.Url,
			})
		}
	}
	return gitMaterialsResp, false
}

//get/build deployment template
func (handler AppRestHandlerImpl) validateAndBuildAppDeploymentTemplate(w http.ResponseWriter, appId int, envId int) (*appBean.DeploymentTemplate, bool) {
	chartRefData, err := handler.chartService.ChartRefAutocompleteForAppOrEnv(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, ChartRefAutocompleteForApp in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	if chartRefData == nil {
		err = errors.New("invalid appId - chartRefData is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, true
	}

	deploymentTemplate, err := handler.chartService.FindLatestChartForAppByAppId(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentTemplate in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	if deploymentTemplate == nil {
		err = errors.New("invalid appId - deploymentTemplate is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, true
	}

	// set deployment template & showAppMetrics
	var deploymentTemplateObj map[string]interface{}
	var showAppMetrics bool
	var deploymentTemplateRaw json.RawMessage
	if envId > 0 {
		env, err := handler.propertiesConfigService.GetEnvironmentProperties(appId, envId, chartRefData.LatestEnvChartRef)
		if err != nil {
			handler.logger.Errorw("service err, GetEnvConfOverride", "err", err, "payload", appId, envId, chartRefData.LatestEnvChartRef)
		}
		deploymentTemplateRaw = env.EnvironmentConfig.EnvOverrideValues
		if *env.AppMetrics != showAppMetrics {
			showAppMetrics = true
		}
	} else {
		showAppMetrics = deploymentTemplate.IsAppMetricsEnabled
		deploymentTemplateRaw = deploymentTemplate.DefaultAppOverride
	}

	if deploymentTemplateRaw != nil {
		err = json.Unmarshal([]byte(deploymentTemplateRaw), &deploymentTemplateObj)
		if err != nil {
			handler.logger.Errorw("service err, un-marshling fail in deploymentTemplate", "err", err, "appId", appId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return nil, true
		}
	}

	// set chartRefId
	var chartRefId int
	if envId == 0 {
		chartRefId = chartRefData.LatestAppChartRef
	} else {
		chartRefId = chartRefData.LatestEnvChartRef
	}

	deploymentTemplateResp := &appBean.DeploymentTemplate{
		ChartRefId:     chartRefId,
		Template:       deploymentTemplateObj,
		ShowAppMetrics: showAppMetrics,
	}

	return deploymentTemplateResp, false
}

// get/build global config maps
func (handler AppRestHandlerImpl) validateAndBuildAppGlobalConfigMaps(w http.ResponseWriter, appId int) ([]*appBean.ConfigMap, bool) {
	configMapData, err := handler.configMapService.CMGlobalFetch(appId)
	if err != nil {
		handler.logger.Errorw("service err, CMGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppConfigMaps(w, appId, configMapData)
}

// get/build environment config maps
func (handler AppRestHandlerImpl) validateAndBuildAppEnvironmentConfigMaps(w http.ResponseWriter, appId int, envId int) ([]*appBean.ConfigMap, bool) {
	configMapData, err := handler.configMapService.CMEnvironmentFetch(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CMGlobalFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppConfigMaps(w, appId, configMapData)
}

// get/build config maps
func (handler AppRestHandlerImpl) validateAndBuildAppConfigMaps(w http.ResponseWriter, appId int, configMapData *pipeline.ConfigDataRequest) ([]*appBean.ConfigMap, bool) {
	var configMapsResp []*appBean.ConfigMap
	if configMapData != nil && len(configMapData.ConfigData) > 0 {
		for _, configMap := range configMapData.ConfigData {

			// initialise
			globalConfigMap := &appBean.ConfigMap{
				Name:       configMap.Name,
				IsExternal: configMap.External,
				UsageType:  configMap.Type,
			}

			// set data
			var dataObj map[string]interface{}
			if configMap.Data != nil {
				err := json.Unmarshal([]byte(configMap.Data), &dataObj)
				if err != nil {
					handler.logger.Errorw("service err, un-marshling fail in config map", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, true
				}
			}
			globalConfigMap.Data = dataObj

			// set data volume usage type
			if configMap.Type == util.ConfigMapSecretUsageTypeVolume {
				globalConfigMap.DataVolumeUsageConfig = &appBean.ConfigMapSecretDataVolumeUsageConfig{
					MountPath:      configMap.MountPath,
					SubPath:        configMap.SubPath,
					FilePermission: configMap.FilePermission,
				}
			}

			configMapsResp = append(configMapsResp, globalConfigMap)
		}
	}
	return configMapsResp, false
}

// get/build global secrets
func (handler AppRestHandlerImpl) validateAndBuildAppGlobalSecrets(w http.ResponseWriter, appId int) ([]*appBean.Secret, bool) {
	secretData, err := handler.configMapService.CSGlobalFetchWithSecretValues(appId)
	if err != nil {
		handler.logger.Errorw("service err, CSGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppSecrets(w, appId, secretData)
}

// get/build environment secrets
func (handler AppRestHandlerImpl) validateAndBuildAppEnvironmentSecrets(w http.ResponseWriter, appId int, envId int) ([]*appBean.Secret, bool) {
	secretData, err := handler.configMapService.CSEnvironmentFetchWithSecretValues(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CSEnvironmentFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppSecrets(w, appId, secretData)
}

// get/build secrets
func (handler AppRestHandlerImpl) validateAndBuildAppSecrets(w http.ResponseWriter, appId int, secretData *pipeline.ConfigDataRequest) ([]*appBean.Secret, bool) {
	var secretsResp []*appBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {
		for _, secret := range secretData.ConfigData {
			// initialise
			globalSecret := &appBean.Secret{
				Name:         secret.Name,
				RoleArn:      secret.RoleARN,
				IsExternal:   secret.External,
				UsageType:    secret.Type,
				ExternalType: secret.ExternalSecretType,
			}

			// set data
			var dataObj map[string]interface{}
			if secret.Data != nil {
				err := json.Unmarshal([]byte(secret.Data), &dataObj)
				if err != nil {
					handler.logger.Errorw("service err, un-marshling fail in secret", "err", err, "appId", appId)
					writeJsonResp(w, err, nil, http.StatusInternalServerError)
					return nil, true
				}
			}
			globalSecret.Data = dataObj

			// set external data
			var externalSecretsResp []*appBean.ExternalSecret
			if len(secret.ExternalSecret) > 0 {
				for _, externalSecret := range secret.ExternalSecret {
					externalSecretsResp = append(externalSecretsResp, &appBean.ExternalSecret{
						Name:     externalSecret.Name,
						Key:      externalSecret.Key,
						Property: externalSecret.Property,
						IsBinary: externalSecret.IsBinary,
					})
				}
			}
			globalSecret.ExternalSecretData = externalSecretsResp

			// set data volume usage type
			if secret.Type == util.ConfigMapSecretUsageTypeVolume {
				globalSecret.DataVolumeUsageConfig = &appBean.ConfigMapSecretDataVolumeUsageConfig{
					MountPath:      secret.MountPath,
					SubPath:        secret.SubPath,
					FilePermission: secret.FilePermission,
				}
			}

			secretsResp = append(secretsResp, globalSecret)
		}
	}
	return secretsResp, false
}

func (handler AppRestHandlerImpl) validateAndBuildEnvironmentOverrides(w http.ResponseWriter, appId int) (map[string]*appBean.EnvironmentOverride, bool) {
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
func (handler AppRestHandlerImpl) validateAndBuildDockerConfig(w http.ResponseWriter, appId int) (*appBean.DockerConfig, bool) {
	ciConfig, err := handler.pipelineBuilder.GetCiPipeline(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetCiPipeline in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	dockerConfig := &appBean.DockerConfig{
		DockerRegistry:   ciConfig.DockerRegistry,
		DockerRepository: ciConfig.DockerRepository,
	}

	//getting gitMaterialUrl by id
	gitMaterial, err := handler.materialRepository.FindById(ciConfig.DockerBuildConfig.GitMaterialId)
	if err != nil {
		handler.logger.Errorw("error in fetching materialUrl by ID in GetAppAllDetail", "err", err, "gitMaterialId", ciConfig.DockerBuildConfig.GitMaterialId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}
	dockerConfig.BuildConfig = &appBean.DockerBuildConfig{
		Args:           ciConfig.DockerBuildConfig.Args,
		DockerfilePath: ciConfig.DockerBuildConfig.DockerfilePath,
		GitMaterialUrl: gitMaterial.Url,
	}
	return dockerConfig, false
}
func (handler AppRestHandlerImpl) validateAndBuildAppWorkflows(w http.ResponseWriter, appId int) ([]*appBean.AppWorkflow, bool) {
	var appWorkflowsResp []*appBean.AppWorkflow
	workflowsList, err := handler.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		handler.logger.Errorw("error in fetching workflows for app in GetAppAllDetail", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}
	ciConfig, err := handler.pipelineBuilder.GetCiPipeline(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetCiPipeline in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}
	cdConfig, err := handler.pipelineBuilder.GetCdPipelinesForApp(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetCdPipelines in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	for _, workflow := range workflowsList {
		workflowResp := &appBean.AppWorkflow{
			Name: workflow.Name,
		}
		var HasInternalCiPipeline bool
		if len(workflow.AppWorkflowMappingDto) == 0 {
			handler.logger.Infow("no pipeline found in workflow in GetAppAllDetail", "workflow", workflow.Name, "appId", appId)
		} else {
			for _, appWorkflowMap := range workflow.AppWorkflowMappingDto {
				if appWorkflowMap.Type == CIPIPELINE {
					for _, ciPipeline := range ciConfig.CiPipelines {
						if ciPipeline.Id == appWorkflowMap.ComponentId {
							//checking if ci pipeline is external or not; if external, will not include this workflow
							if !ciPipeline.IsExternal {
								HasInternalCiPipeline = true
								workflowResp.CiPipeline = BuildCiPipelineResp(ciPipeline)
							}
						}
					}
				} else if appWorkflowMap.Type == CDPIPELINE && HasInternalCiPipeline {
					for _, cdPipeline := range cdConfig.Pipelines {
						if cdPipeline.Id == appWorkflowMap.ComponentId {
							workflowResp.CdPipeline = append(workflowResp.CdPipeline, BuildCdPipelineResp(cdPipeline))
						}
					}
				}
			}
		}
		appWorkflowsResp = append(appWorkflowsResp, workflowResp)
	}
	return appWorkflowsResp, false
}

func BuildCiPipelineResp(ciPipeline *bean.CiPipeline) *appBean.CiPipelineDetails {
	var ciMaterialsResp []*appBean.CiMaterial
	var beforeDockerBuildScriptsResp []*appBean.CiScript
	var afterDockerBuildScriptsResp []*appBean.CiScript
	var beforeDockerBuildTasks []*appBean.Task
	var afterDockerBuildTasks []*appBean.Task
	for _, beforeDockerBuildScript := range ciPipeline.BeforeDockerBuildScripts {
		beforeDockerBuildScriptResp := &appBean.CiScript{
			Name:           beforeDockerBuildScript.Name,
			Index:          beforeDockerBuildScript.Index,
			Script:         beforeDockerBuildScript.Script,
			OutputLocation: beforeDockerBuildScript.OutputLocation,
		}
		beforeDockerBuildScriptsResp = append(beforeDockerBuildScriptsResp, beforeDockerBuildScriptResp)
	}
	for _, afterDockerBuildScript := range ciPipeline.AfterDockerBuildScripts {
		afterDockerBuildScriptResp := &appBean.CiScript{
			Name:           afterDockerBuildScript.Name,
			Index:          afterDockerBuildScript.Index,
			Script:         afterDockerBuildScript.Script,
			OutputLocation: afterDockerBuildScript.OutputLocation,
		}
		afterDockerBuildScriptsResp = append(afterDockerBuildScriptsResp, afterDockerBuildScriptResp)
	}
	for _, ciMaterial := range ciPipeline.CiMaterial {
		ciMaterialResp := &appBean.CiMaterial{
			Path:            ciMaterial.Path,
			CheckoutPath:    ciMaterial.CheckoutPath,
			GitMaterialName: ciMaterial.GitMaterialName,
		}
		ciMaterialResp.Source = &appBean.SourceTypeConfig{
			Type:  ciMaterial.Source.Type,
			Value: ciMaterial.Source.Value,
		}
		ciMaterialsResp = append(ciMaterialsResp, ciMaterialResp)
	}
	for _, beforeDockerBuild := range ciPipeline.BeforeDockerBuild {
		beforeDockerBuildTask := &appBean.Task{
			Name: beforeDockerBuild.Name,
			Type: beforeDockerBuild.Type,
			Cmd:  beforeDockerBuild.Cmd,
			Args: beforeDockerBuild.Args,
		}
		beforeDockerBuildTasks = append(beforeDockerBuildTasks, beforeDockerBuildTask)
	}
	for _, afterDockerBuild := range ciPipeline.AfterDockerBuild {
		afterDockerBuildTask := &appBean.Task{
			Name: afterDockerBuild.Name,
			Type: afterDockerBuild.Type,
			Cmd:  afterDockerBuild.Cmd,
			Args: afterDockerBuild.Args,
		}
		afterDockerBuildTasks = append(afterDockerBuildTasks, afterDockerBuildTask)
	}

	ciPipelineResp := &appBean.CiPipelineDetails{
		Name:                     ciPipeline.Name,
		IsManual:                 ciPipeline.IsManual,
		DockerArgs:               ciPipeline.DockerArgs,
		LinkedCount:              ciPipeline.LinkedCount,
		ScanEnabled:              ciPipeline.ScanEnabled,
		CiMaterials:              ciMaterialsResp,
		BeforeDockerBuildScripts: beforeDockerBuildScriptsResp,
		AfterDockerBuildScripts:  afterDockerBuildScriptsResp,
		BeforeDockerBuild:        beforeDockerBuildTasks,
		AfterDockerBuild:         afterDockerBuildTasks,
	}
	return ciPipelineResp
}

func BuildCdPipelineResp(cdPipeline *bean.CDPipelineConfigObject) *appBean.CdPipelineDetails {
	var strategiesResp []appBean.Strategy
	for _, strategy := range cdPipeline.Strategies {
		strategyResp := appBean.Strategy{
			DeploymentTemplate: strategy.DeploymentTemplate,
			Config:             strategy.Config,
			Default:            strategy.Default,
		}
		strategiesResp = append(strategiesResp, strategyResp)
	}
	preStagesResp := appBean.CdStage{
		TriggerType: cdPipeline.PreStage.TriggerType,
		Name:        cdPipeline.PreStage.Name,
		Config:      cdPipeline.PreStage.Config,
		Status:      cdPipeline.PreStage.Status,
	}
	postStagesResp := appBean.CdStage{
		TriggerType: cdPipeline.PostStage.TriggerType,
		Name:        cdPipeline.PostStage.Name,
		Config:      cdPipeline.PostStage.Config,
		Status:      cdPipeline.PostStage.Status,
	}
	preStageConfigMapSecretNamesResp := appBean.PreStageConfigMapSecretNames{
		ConfigMaps: cdPipeline.PreStageConfigMapSecretNames.ConfigMaps,
		Secrets:    cdPipeline.PreStageConfigMapSecretNames.Secrets,
	}
	postStageConfigMapSecretNamesResp := appBean.PostStageConfigMapSecretNames{
		ConfigMaps: cdPipeline.PostStageConfigMapSecretNames.ConfigMaps,
		Secrets:    cdPipeline.PostStageConfigMapSecretNames.Secrets,
	}

	cdPipelineResp := &appBean.CdPipelineDetails{
		EnvironmentName:               cdPipeline.EnvironmentName,
		TriggerType:                   cdPipeline.TriggerType,
		Name:                          cdPipeline.Name,
		Strategies:                    strategiesResp,
		Namespace:                     cdPipeline.Namespace,
		DeploymentTemplate:            cdPipeline.DeploymentTemplate,
		PreStage:                      preStagesResp,
		PostStage:                     postStagesResp,
		PreStageConfigMapSecretNames:  preStageConfigMapSecretNamesResp,
		PostStageConfigMapSecretNames: postStageConfigMapSecretNamesResp,
		RunPreStageInEnv:              cdPipeline.RunPreStageInEnv,
		RunPostStageInEnv:             cdPipeline.RunPostStageInEnv,
		CdArgoSetup:                   cdPipeline.CdArgoSetup,
	}
	return cdPipelineResp
}
