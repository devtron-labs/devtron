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

package configMapAndSecret

import (
	"encoding/json"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ConfigMapHistoryService interface {
	CreateHistoryFromAppLevelConfig(appLevelConfig *chartConfig.ConfigMapAppModel, configType repository.ConfigType) error
	CreateHistoryFromEnvLevelConfig(envLevelConfig *chartConfig.ConfigMapEnvModel, configType repository.ConfigType) error
	CreateCMCSHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, deployedOn time.Time, deployedBy int32) (int, int, error)
	MergeAppLevelAndEnvLevelConfigs(appLevelConfig *chartConfig.ConfigMapAppModel, envLevelConfig *chartConfig.ConfigMapEnvModel, configType repository.ConfigType, configMapSecretNames []string) (string, error)
}

type ConfigMapHistoryServiceImpl struct {
	logger                     *zap.SugaredLogger
	configMapHistoryRepository repository.ConfigMapHistoryRepository
	pipelineRepository         pipelineConfig.PipelineRepository
	configMapRepository        chartConfig.ConfigMapRepository
	userService                user.UserService
	scopedVariableManager      variables.ScopedVariableCMCSManager
}

func NewConfigMapHistoryServiceImpl(logger *zap.SugaredLogger,
	configMapHistoryRepository repository.ConfigMapHistoryRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	userService user.UserService,
	scopedVariableManager variables.ScopedVariableCMCSManager,
) *ConfigMapHistoryServiceImpl {
	return &ConfigMapHistoryServiceImpl{
		logger:                     logger,
		configMapHistoryRepository: configMapHistoryRepository,
		pipelineRepository:         pipelineRepository,
		configMapRepository:        configMapRepository,
		userService:                userService,
		scopedVariableManager:      scopedVariableManager,
	}
}

func (impl ConfigMapHistoryServiceImpl) CreateHistoryFromAppLevelConfig(appLevelConfig *chartConfig.ConfigMapAppModel, configType repository.ConfigType) error {
	pipelines, err := impl.pipelineRepository.FindActiveByAppId(appLevelConfig.AppId)
	if err != nil {
		impl.logger.Errorw("err in getting pipelines, CreateHistoryFromAppLevelConfig", "err", err, "appLevelConfig", appLevelConfig)
		return err
	}
	//creating history for global
	configData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, nil, configType, nil)
	if err != nil {
		impl.logger.Errorw("err in merging app and env level configs", "err", err)
		return err
	}
	historyModel := &repository.ConfigmapAndSecretHistory{
		AppId:    appLevelConfig.AppId,
		DataType: configType,
		Deployed: false,
		Data:     configData,
		AuditLog: sql.AuditLog{
			CreatedBy: appLevelConfig.CreatedBy,
			CreatedOn: appLevelConfig.CreatedOn,
			UpdatedBy: appLevelConfig.UpdatedBy,
			UpdatedOn: appLevelConfig.UpdatedOn,
		},
	}
	_, err = impl.configMapHistoryRepository.CreateHistory(nil, historyModel)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for CM/CS history", "historyModel", historyModel)
		return err
	}
	for _, pipeline := range pipelines {
		envLevelConfig, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(pipeline.AppId, pipeline.EnvironmentId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("err in getting env level config", "err", err, "appId", appLevelConfig.AppId)
			return err
		}
		configData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, configType, nil)
		if err != nil {
			impl.logger.Errorw("err in merging app and env level configs", "err", err)
			return err
		}
		historyModel := &repository.ConfigmapAndSecretHistory{
			AppId:      appLevelConfig.AppId,
			PipelineId: pipeline.Id,
			DataType:   configType,
			Deployed:   false,
			Data:       configData,
			AuditLog: sql.AuditLog{
				CreatedBy: appLevelConfig.CreatedBy,
				CreatedOn: appLevelConfig.CreatedOn,
				UpdatedBy: appLevelConfig.UpdatedBy,
				UpdatedOn: appLevelConfig.UpdatedOn,
			},
		}
		_, err = impl.configMapHistoryRepository.CreateHistory(nil, historyModel)
		if err != nil {
			impl.logger.Errorw("error in creating new entry for CM/CS history", "historyModel", historyModel)
			return err
		}
	}

	return nil
}

func (impl ConfigMapHistoryServiceImpl) CreateHistoryFromEnvLevelConfig(envLevelConfig *chartConfig.ConfigMapEnvModel, configType repository.ConfigType) error {
	pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(envLevelConfig.AppId, envLevelConfig.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("err in getting pipelines, CreateHistoryFromEnvLevelConfig", "err", err, "envLevelConfig", envLevelConfig)
		return err
	}
	appLevelConfig, err := impl.configMapRepository.GetByAppIdAppLevel(envLevelConfig.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting app level config", "err", err, "appId", envLevelConfig.AppId)
		return err
	}
	for _, pipeline := range pipelines {
		configData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, configType, nil)
		if err != nil {
			impl.logger.Errorw("err in merging app and env level configs", "err", err)
			return err
		}
		historyModel := &repository.ConfigmapAndSecretHistory{
			AppId:      envLevelConfig.AppId,
			PipelineId: pipeline.Id,
			DataType:   configType,
			Deployed:   false,
			Data:       configData,
			AuditLog: sql.AuditLog{
				CreatedBy: envLevelConfig.CreatedBy,
				CreatedOn: envLevelConfig.CreatedOn,
				UpdatedBy: envLevelConfig.UpdatedBy,
				UpdatedOn: envLevelConfig.UpdatedOn,
			},
		}
		_, err = impl.configMapHistoryRepository.CreateHistory(nil, historyModel)
		if err != nil {
			impl.logger.Errorw("error in creating new entry for CM/CS history", "historyModel", historyModel)
			return err
		}
	}
	return nil
}

func (impl ConfigMapHistoryServiceImpl) CreateCMCSHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, deployedOn time.Time, deployedBy int32) (int, int, error) {
	//creating history for configmaps, secrets(if any)
	appLevelConfig, err := impl.configMapRepository.GetByAppIdAppLevel(pipeline.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting app level config", "err", err, "appId", pipeline.AppId)
		return 0, 0, err
	}
	envLevelConfig, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting env level config", "err", err, "appId", pipeline.AppId)
		return 0, 0, err
	}
	configMapData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, repository.CONFIGMAP_TYPE, nil)
	if err != nil {
		impl.logger.Errorw("err in merging app and env level configs", "err", err)
		return 0, 0, err
	}
	historyModelForCM := repository.ConfigmapAndSecretHistory{
		AppId:      pipeline.AppId,
		PipelineId: pipeline.Id,
		DataType:   repository.CONFIGMAP_TYPE,
		Deployed:   true,
		DeployedBy: deployedBy,
		DeployedOn: deployedOn,
		Data:       configMapData,
		AuditLog: sql.AuditLog{
			CreatedBy: deployedBy,
			CreatedOn: deployedOn,
			UpdatedBy: deployedBy,
			UpdatedOn: deployedOn,
		},
	}

	tx, err := impl.configMapHistoryRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to create new cm/cs history", "error", err)
		return 0, 0, err
	}
	defer impl.configMapHistoryRepository.RollbackTx(tx)
	cmHistory, err := impl.configMapHistoryRepository.CreateHistory(tx, &historyModelForCM)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for cm history", "historyModel", historyModelForCM)
		return 0, 0, err
	}
	secretData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, repository.SECRET_TYPE, nil)
	if err != nil {
		impl.logger.Errorw("err in merging app and env level configs", "err", err)
		return 0, 0, err
	}
	historyModelForCS := historyModelForCM
	historyModelForCS.DataType = repository.SECRET_TYPE
	historyModelForCS.Data = secretData
	historyModelForCS.Id = 0
	csHistory, err := impl.configMapHistoryRepository.CreateHistory(tx, &historyModelForCS)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for secret history", "historyModel", historyModelForCS)
		return 0, 0, err
	}
	err = impl.configMapHistoryRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to create new cm/cs history", "error", err)
		return 0, 0, err
	}
	return cmHistory.Id, csHistory.Id, nil
}

func (impl ConfigMapHistoryServiceImpl) MergeAppLevelAndEnvLevelConfigs(appLevelConfig *chartConfig.ConfigMapAppModel, envLevelConfig *chartConfig.ConfigMapEnvModel, configType repository.ConfigType, configMapSecretNames []string) (string, error) {
	var err error
	var appLevelConfigData []*bean.ConfigData
	var envLevelConfigData []*bean.ConfigData
	if configType == repository.CONFIGMAP_TYPE {
		var configDataAppLevel string
		var configDataEnvLevel string
		if appLevelConfig != nil {
			configDataAppLevel = appLevelConfig.ConfigMapData
		}
		if envLevelConfig != nil {
			configDataEnvLevel = envLevelConfig.ConfigMapData
		}
		configListAppLevel := &bean.ConfigList{}
		if len(configDataAppLevel) > 0 {
			err = json.Unmarshal([]byte(configDataAppLevel), configListAppLevel)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return "", err
			}
		}
		configListEnvLevel := &bean.ConfigList{}
		if len(configDataEnvLevel) > 0 {
			err = json.Unmarshal([]byte(configDataEnvLevel), configListEnvLevel)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return "", err
			}
		}
		appLevelConfigData = configListAppLevel.ConfigData
		envLevelConfigData = configListEnvLevel.ConfigData
	} else if configType == repository.SECRET_TYPE {
		var secretDataAppLevel string
		var secretDataEnvLevel string
		if appLevelConfig != nil {
			secretDataAppLevel = appLevelConfig.SecretData
		}
		if envLevelConfig != nil {
			secretDataEnvLevel = envLevelConfig.SecretData
		}
		secretListAppLevel := &bean.SecretList{}
		if len(secretDataAppLevel) > 0 {
			err = json.Unmarshal([]byte(secretDataAppLevel), secretListAppLevel)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return "", err
			}
		}
		secretListEnvLevel := &bean.SecretList{}
		if len(secretDataEnvLevel) > 0 {
			err = json.Unmarshal([]byte(secretDataEnvLevel), secretListEnvLevel)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return "", err
			}
		}
		appLevelConfigData = secretListAppLevel.ConfigData
		envLevelConfigData = secretListEnvLevel.ConfigData
	}

	var finalConfigs []*bean.ConfigData
	envLevelConfigs := make(map[string]bool)
	filterNameMap := make(map[string]bool)
	for _, name := range configMapSecretNames {
		filterNameMap[name] = true
	}
	//if filter name map is not empty, to add configs by filtering names
	//if filter name map is empty, adding all env level configs to final configs
	for _, item := range envLevelConfigData {
		if _, ok := filterNameMap[item.Name]; ok || len(filterNameMap) == 0 {
			//adding all env configs whose name is in filter name map
			envLevelConfigs[item.Name] = true
			finalConfigs = append(finalConfigs, item)
		}
	}
	for _, item := range appLevelConfigData {
		//if filter name map is not empty, adding all global configs which are not present in env level and are present in filter name map to final configs
		//if filter name map is empty,adding all global configs which are not present in env level to final configs
		if _, ok := envLevelConfigs[item.Name]; !ok {
			if _, ok = filterNameMap[item.Name]; ok || len(filterNameMap) == 0 {
				finalConfigs = append(finalConfigs, item)
			}
		}
	}
	var finalConfigDataByte []byte
	if configType == repository.CONFIGMAP_TYPE {
		var finalConfigList bean.ConfigList
		finalConfigList.ConfigData = finalConfigs
		finalConfigDataByte, err = json.Marshal(finalConfigList)
		if err != nil {
			impl.logger.Errorw("error in marshaling config", "err", err)
			return "", err
		}
	} else if configType == repository.SECRET_TYPE {
		var finalConfigList bean.SecretList
		finalConfigList.ConfigData = finalConfigs
		finalConfigDataByte, err = json.Marshal(finalConfigList)
		if err != nil {
			impl.logger.Errorw("error in marshaling config", "err", err)
			return "", err
		}
	}
	return string(finalConfigDataByte), err
}
