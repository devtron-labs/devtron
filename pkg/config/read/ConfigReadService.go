/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package read

import (
	"encoding/json"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	configMapRepository "github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	internalUtil "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/commonService"
	configMapBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"go.uber.org/zap"
)

type ConfigReadService interface {
	GetCmCsForPrePostStageTrigger(scope resourceQualifiers.Scope, appId int, envId int, isJob bool) (*apiBean.ConfigMapJson, *apiBean.ConfigSecretJson, error)
}

type ConfigReadServiceImpl struct {
	logger                *zap.SugaredLogger
	commonService         commonService.CommonService
	configMapRepository   configMapRepository.ConfigMapRepository
	mergeUtil             internalUtil.MergeUtil
	scopedVariableManager variables.ScopedVariableCMCSManager
}

func NewConfigReadServiceImpl(logger *zap.SugaredLogger, commonService commonService.CommonService,
	configMapRepository configMapRepository.ConfigMapRepository, mergeUtil internalUtil.MergeUtil,
	scopedVariableManager variables.ScopedVariableCMCSManager) *ConfigReadServiceImpl {
	return &ConfigReadServiceImpl{
		logger:                logger,
		commonService:         commonService,
		configMapRepository:   configMapRepository,
		mergeUtil:             mergeUtil,
		scopedVariableManager: scopedVariableManager,
	}
}

func (impl *ConfigReadServiceImpl) GetCmCsForPrePostStageTrigger(scope resourceQualifiers.Scope, appId int, envId int, isJob bool) (*apiBean.ConfigMapJson, *apiBean.ConfigSecretJson, error) {
	appLevelConfig, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && !internalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching app level config", "appId", appId, "error", err)
		return nil, nil, err
	}
	envLevelConfig, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && !internalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching env level config", "appId", appId, "envId", envId, "error", err)
		return nil, nil, err
	}
	request := configMapBean.NewResolvedCmCsRequest(scope).
		WithAppId(appId).WithEnvId(envId).ForJob(isJob)
	return impl.getResolvedCmCsForPrePostStageTrigger(request, appLevelConfig, envLevelConfig)
}

func (impl *ConfigReadServiceImpl) getResolvedCmCsForPrePostStageTrigger(request *configMapBean.ResolvedCmCsRequest,
	appLevelConfig *configMapRepository.ConfigMapAppModel, envLevelConfig *configMapRepository.ConfigMapEnvModel) (*apiBean.ConfigMapJson, *apiBean.ConfigSecretJson, error) {
	var secretDataJsonApp string
	var configMapJsonApp string
	if appLevelConfig != nil && appLevelConfig.Id > 0 {
		configMapJsonApp = appLevelConfig.ConfigMapData
		secretDataJsonApp = appLevelConfig.SecretData
	}
	var secretDataJsonEnv string
	var configMapJsonEnv string
	if envLevelConfig != nil && envLevelConfig.Id > 0 {
		configMapJsonEnv = envLevelConfig.ConfigMapData
		secretDataJsonEnv = envLevelConfig.SecretData
	}
	configMapJson, err := impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
	if err != nil {
		return nil, nil, err
	}
	var secretDataJson string
	if request.IsJob {
		secretDataJson, err = impl.mergeUtil.ConfigSecretMergeForJob(secretDataJsonApp, secretDataJsonEnv)
		if err != nil {
			return nil, nil, err
		}
	} else {
		chartVersion, err := impl.commonService.FetchLatestChartVersion(request.AppId, request.EnvId)
		if err != nil {
			return nil, nil, err
		}
		secretDataJson, err = impl.mergeUtil.ConfigSecretMergeForCDStages(secretDataJsonApp, secretDataJsonEnv, chartVersion)
		if err != nil {
			return nil, nil, err
		}
	}
	configResponse := apiBean.ConfigMapJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(configMapJson), &configResponse)
		if err != nil {
			return nil, nil, err
		}
	}
	secretResponse := apiBean.ConfigSecretJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
		if err != nil {
			return nil, nil, err
		}
	}
	var appLevelConfigId, envLevelConfigId int
	if appLevelConfig != nil {
		appLevelConfigId = appLevelConfig.Id
	}
	if envLevelConfig != nil {
		envLevelConfigId = envLevelConfig.Id
	}
	resolvedConfigResponse, resolvedSecretResponse, err := impl.scopedVariableManager.ResolveForPrePostStageTrigger(request.Scope, configResponse, secretResponse, appLevelConfigId, envLevelConfigId)
	if err != nil {
		return nil, nil, err
	}
	return resolvedConfigResponse, resolvedSecretResponse, nil
}
