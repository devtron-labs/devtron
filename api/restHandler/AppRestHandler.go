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
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/app"
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
	logger             *zap.SugaredLogger
	userAuthService    user.UserService
	validator          *validator.Validate
	enforcerUtil       rbac.EnforcerUtil
	enforcer           rbac.Enforcer
	appLabelService    app.AppLabelService
	pipelineBuilder    pipeline.PipelineBuilder
	gitRegistryService pipeline.GitRegistryConfig
	chartService       pipeline.ChartService
	configMapService   pipeline.ConfigMapService
	appListingService  app.AppListingService
	propertiesConfigService pipeline.PropertiesConfigService
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService, validator *validator.Validate, enforcerUtil rbac.EnforcerUtil,
	enforcer rbac.Enforcer, appLabelService app.AppLabelService, pipelineBuilder pipeline.PipelineBuilder, gitRegistryService pipeline.GitRegistryConfig,
	chartService pipeline.ChartService, configMapService pipeline.ConfigMapService, appListingService app.AppListingService, propertiesConfigService pipeline.PropertiesConfigService) *AppRestHandlerImpl {
	handler := &AppRestHandlerImpl{
		logger:             logger,
		userAuthService:    userAuthService,
		validator:          validator,
		enforcerUtil:       enforcerUtil,
		enforcer:           enforcer,
		appLabelService:    appLabelService,
		pipelineBuilder:    pipelineBuilder,
		gitRegistryService: gitRegistryService,
		chartService:       chartService,
		configMapService:   configMapService,
		appListingService:  appListingService,
		propertiesConfigService: propertiesConfigService,
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

	// build full object for response
	appDetail := &apiBean.AppDetail{
		Metadata:                 appMetadataResp,
		GitMaterials:             gitMaterialsResp,
		GlobalDeploymentTemplate: globalDeploymentTemplateResp,
		GlobalConfigMaps:         globalConfigMapsResp,
		GlobalSecrets:            globalSecretsResp,
		EnvironmentOverrides:     environmentOverrides,
	}
	// end

	writeJsonResp(w, nil, appDetail, http.StatusOK)
}

//get/build app metadata
func (handler AppRestHandlerImpl) validateAndBuildAppMetadata(w http.ResponseWriter, appId int) (*apiBean.AppMetadata, bool) {
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

	var appLabelsRes []*apiBean.AppLabel
	if len(appMetaInfo.Labels) > 0 {
		for _, label := range appMetaInfo.Labels {
			appLabelsRes = append(appLabelsRes, &apiBean.AppLabel{
				Key:   label.Key,
				Value: label.Value,
			})
		}
	}
	appMetadataResp := &apiBean.AppMetadata{
		AppName:     appMetaInfo.AppName,
		ProjectName: appMetaInfo.ProjectName,
		Labels:      appLabelsRes,
	}

	return appMetadataResp, false
}

//get/build git materials
func (handler AppRestHandlerImpl) validateAndBuildAppGitMaterials(w http.ResponseWriter, appId int) ([]*apiBean.GitMaterial, bool) {
	gitMaterials := handler.pipelineBuilder.GetMaterialsForAppId(appId)
	var gitMaterialsResp []*apiBean.GitMaterial
	if len(gitMaterials) > 0 {
		for _, gitMaterial := range gitMaterials {
			gitRegistry, err := handler.gitRegistryService.FetchOneGitProvider(strconv.Itoa(gitMaterial.GitProviderId))
			if err != nil {
				handler.logger.Errorw("service err, getGitProvider in GetAppAllDetail", "err", err, "appId", appId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return nil, true
			}

			gitMaterialsResp = append(gitMaterialsResp, &apiBean.GitMaterial{
				GitUrl:          gitMaterial.Url,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				GitAccountName:  gitRegistry.Name,
			})
		}
	}
	return gitMaterialsResp, false
}

//get/build deployment template
func (handler AppRestHandlerImpl) validateAndBuildAppDeploymentTemplate(w http.ResponseWriter, appId int, envId int) (*apiBean.DeploymentTemplate, bool) {
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
		env, err:= handler.propertiesConfigService.GetEnvironmentProperties(appId,envId,chartRefData.LatestEnvChartRef)
		if err!=nil{
			handler.logger.Errorw("service err, GetEnvConfOverride", "err", err, "payload", appId, envId, chartRefData.LatestEnvChartRef)
		}
		deploymentTemplateRaw = env.EnvironmentConfig.EnvOverrideValues
		if *env.AppMetrics != showAppMetrics{
			showAppMetrics = true
		}
	}else{
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

	deploymentTemplateResp := &apiBean.DeploymentTemplate{
		ChartRefId:     chartRefId,
		Template:       deploymentTemplateObj,
		ShowAppMetrics: showAppMetrics,
	}

	return deploymentTemplateResp, false
}

// get/build global config maps
func (handler AppRestHandlerImpl) validateAndBuildAppGlobalConfigMaps(w http.ResponseWriter, appId int) ([]*apiBean.ConfigMap, bool) {
	configMapData, err := handler.configMapService.CMGlobalFetch(appId)
	if err != nil {
		handler.logger.Errorw("service err, CMGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppConfigMaps(w, appId, configMapData)
}

// get/build environment config maps
func (handler AppRestHandlerImpl) validateAndBuildAppEnvironmentConfigMaps(w http.ResponseWriter, appId int, envId int) ([]*apiBean.ConfigMap, bool) {
	configMapData, err := handler.configMapService.CMEnvironmentFetch(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CMGlobalFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppConfigMaps(w, appId, configMapData)
}

// get/build config maps
func (handler AppRestHandlerImpl) validateAndBuildAppConfigMaps(w http.ResponseWriter, appId int, configMapData *pipeline.ConfigDataRequest) ([]*apiBean.ConfigMap, bool) {
	var configMapsResp []*apiBean.ConfigMap
	if configMapData != nil && len(configMapData.ConfigData) > 0 {
		for _, configMap := range configMapData.ConfigData {

			// initialise
			globalConfigMap := &apiBean.ConfigMap{
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
				globalConfigMap.DataVolumeUsageConfig = &apiBean.ConfigMapSecretDataVolumeUsageConfig{
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
func (handler AppRestHandlerImpl) validateAndBuildAppGlobalSecrets(w http.ResponseWriter, appId int) ([]*apiBean.Secret, bool) {
	secretData, err := handler.configMapService.CSGlobalFetchWithSecretValues(appId)
	if err != nil {
		handler.logger.Errorw("service err, CSGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppSecrets(w, appId, secretData)
}

// get/build environment secrets
func (handler AppRestHandlerImpl) validateAndBuildAppEnvironmentSecrets(w http.ResponseWriter, appId int, envId int) ([]*apiBean.Secret, bool) {
	secretData, err := handler.configMapService.CSEnvironmentFetchWithSecretValues(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CSEnvironmentFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	return handler.validateAndBuildAppSecrets(w, appId, secretData)
}

// get/build secrets
func (handler AppRestHandlerImpl) validateAndBuildAppSecrets(w http.ResponseWriter, appId int, secretData *pipeline.ConfigDataRequest) ([]*apiBean.Secret, bool) {
	var secretsResp []*apiBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {
		for _, secret := range secretData.ConfigData {
			// initialise
			globalSecret := &apiBean.Secret{
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
			var externalSecretsResp []*apiBean.ExternalSecret
			if len(secret.ExternalSecret) > 0 {
				for _, externalSecret := range secret.ExternalSecret {
					externalSecretsResp = append(externalSecretsResp, &apiBean.ExternalSecret{
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
				globalSecret.DataVolumeUsageConfig = &apiBean.ConfigMapSecretDataVolumeUsageConfig{
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

func (handler AppRestHandlerImpl) validateAndBuildEnvironmentOverrides(w http.ResponseWriter, appId int) (map[string]*apiBean.EnvironmentOverride, bool) {
	appEnvironments, err := handler.appListingService.FetchOtherEnvironment(appId)
	if err != nil {
		handler.logger.Errorw("service err, Fetch app environments in GetAppAllDetail", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	environmentOverrides := make(map[string]*apiBean.EnvironmentOverride)
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

			environmentOverrides[appEnvironment.EnvironmentName] = &apiBean.EnvironmentOverride{
				Secrets:            envSecretsResp,
				ConfigMaps:         envConfigMapsResp,
				DeploymentTemplate: envDeploymentTemplateResp,
			}
		}
	}
	return environmentOverrides, false
}
