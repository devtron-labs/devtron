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

package history

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/pkg/configDiff/adaptor"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/configDiff/utils"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	globalUtil "github.com/devtron-labs/devtron/util"
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
	GetDeploymentDetailsForDeployedCMCSHistory(pipelineId int, configType repository.ConfigType) ([]*ConfigMapAndSecretHistoryDto, error)

	GetHistoryForDeployedCMCSById(ctx context.Context, id, pipelineId int, configType repository.ConfigType, componentName string, userHasAdminAccess bool) (*HistoryDetailDto, error)
	GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType repository.ConfigType) (history *repository.ConfigmapAndSecretHistory, exists bool, cmCsNames []string, err error)
	GetDeployedHistoryList(pipelineId, baseConfigId int, configType repository.ConfigType, componentName string) ([]*DeployedHistoryComponentMetadataDto, error)

	CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (cmId int, csId int, exists bool, err error)
	GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int, configType repository.ConfigType, userHasAdminAccess bool) ([]*ComponentLevelHistoryDetailDto, error)
	ConvertConfigDataToComponentLevelDto(config *bean.ConfigData, configType repository.ConfigType, userHasAdminAccess bool) (*ComponentLevelHistoryDetailDto, error)

	GetConfigmapHistoryDataByDeployedOnAndPipelineId(ctx context.Context, pipelineId int, deployedOn time.Time, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfig, *bean2.DeploymentAndCmCsConfig, error)
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

func (impl ConfigMapHistoryServiceImpl) GetDeploymentDetailsForDeployedCMCSHistory(pipelineId int, configType repository.ConfigType) ([]*ConfigMapAndSecretHistoryDto, error) {
	histories, err := impl.configMapHistoryRepository.GetDeploymentDetailsForDeployedCMCSHistory(pipelineId, configType)
	if err != nil {
		impl.logger.Errorw("error in getting histories for cm/cs", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historiesDto []*ConfigMapAndSecretHistoryDto
	for _, history := range histories {
		userEmailId, err := impl.userService.GetActiveEmailById(history.DeployedBy)
		if err != nil {
			impl.logger.Errorw("unable to find user email by id", "err", err, "id", history.DeployedBy)
			return nil, err
		}
		historyDto := &ConfigMapAndSecretHistoryDto{
			Id:         history.Id,
			AppId:      history.AppId,
			PipelineId: history.PipelineId,
			Deployed:   history.Deployed,
			DeployedOn: history.DeployedOn,
			DeployedBy: history.DeployedBy,
			EmailId:    userEmailId,
		}
		historiesDto = append(historiesDto, historyDto)
	}
	return historiesDto, nil
}

func (impl ConfigMapHistoryServiceImpl) GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType repository.ConfigType) (history *repository.ConfigmapAndSecretHistory, exists bool, cmCsNames []string, err error) {
	impl.logger.Debugw("received request, CheckIfHistoryExistsForPipelineIdAndWfrId", "pipelineId", pipelineId, "wfrId", wfrId)
	//checking if history exists for pipelineId and wfrId
	history, err = impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId, configType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return history, false, cmCsNames, err
	} else if err == pg.ErrNoRows {
		return history, false, cmCsNames, nil
	}
	var configData []*bean.ConfigData
	if configType == repository.CONFIGMAP_TYPE {
		configList := bean.ConfigList{}
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return history, false, cmCsNames, err
			}
		}
		configData = configList.ConfigData
	} else if configType == repository.SECRET_TYPE {
		secretList := bean.SecretList{}
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &secretList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return history, false, cmCsNames, err
			}
		}
		configData = secretList.ConfigData
	}
	for _, data := range configData {
		cmCsNames = append(cmCsNames, data.Name)
	}
	if len(configData) == 0 {
		return history, false, cmCsNames, nil
	}

	return history, true, cmCsNames, nil
}

func (impl ConfigMapHistoryServiceImpl) GetDeployedHistoryList(pipelineId, baseConfigId int, configType repository.ConfigType, componentName string) ([]*DeployedHistoryComponentMetadataDto, error) {
	impl.logger.Debugw("received request, GetDeployedHistoryList", "pipelineId", pipelineId, "baseConfigId", baseConfigId)

	//checking if history exists for pipelineId and wfrId
	histories, err := impl.configMapHistoryRepository.GetDeployedHistoryList(pipelineId, baseConfigId, configType, componentName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting history list for pipelineId and baseConfigId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historyList []*DeployedHistoryComponentMetadataDto
	for _, history := range histories {
		historyList = append(historyList, &DeployedHistoryComponentMetadataDto{
			Id:               history.Id,
			DeployedOn:       history.DeployedOn,
			DeployedBy:       history.DeployedByEmailId,
			DeploymentStatus: history.DeploymentStatus,
		})
	}
	return historyList, nil
}

func (impl ConfigMapHistoryServiceImpl) GetHistoryForDeployedCMCSById(ctx context.Context, id, pipelineId int, configType repository.ConfigType, componentName string, userHasAdminAccess bool) (*HistoryDetailDto, error) {
	history, err := impl.configMapHistoryRepository.GetHistoryForDeployedCMCSById(id, pipelineId, configType)
	if err != nil {
		impl.logger.Errorw("error in getting histories for cm/cs", "err", err, "id", id, "pipelineId", pipelineId)
		return nil, err
	}
	var configData []*bean.ConfigData
	var configList bean.ConfigList
	var secretList bean.SecretList
	if configType == repository.CONFIGMAP_TYPE {
		configList = bean.ConfigList{}
		if len(history.Data) > 0 {
			err := json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		configData = configList.ConfigData
	} else if configType == repository.SECRET_TYPE {
		secretList = bean.SecretList{}
		if len(history.Data) > 0 {
			err := json.Unmarshal([]byte(history.Data), &secretList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		configData = secretList.ConfigData

	}

	variableSnapshotMap, resolvedTemplate, err := impl.scopedVariableManager.ResolveCMCSHistoryDto(ctx, configType, configList, history, componentName, secretList)
	if err != nil {
		return nil, err
	}

	config := &bean.ConfigData{}
	for _, data := range configData {
		if data.Name == componentName {
			config = data
			break
		}
	}
	historyDto := &HistoryDetailDto{
		Type:           config.Type,
		External:       &config.External,
		MountPath:      config.MountPath,
		SubPath:        &config.SubPath,
		FilePermission: config.FilePermission,
		CodeEditorValue: &HistoryDetailConfig{
			DisplayName:      DataDisplayName,
			Value:            string(config.Data),
			VariableSnapshot: variableSnapshotMap,
			ResolvedValue:    resolvedTemplate,
		},
		SecretViewAccess: userHasAdminAccess,
	}
	if configType == repository.SECRET_TYPE {
		if config.Data != nil {
			if !userHasAdminAccess {
				//removing keys and sending
				resultMap := make(map[string]string)
				resultMapFinal := make(map[string]string)
				err = json.Unmarshal(config.Data, &resultMap)
				if err != nil {
					impl.logger.Warnw("unmarshal failed", "error", err)
				}
				for key, _ := range resultMap {
					//hard-coding values to show them as hidden to user
					resultMapFinal[key] = "*****"
				}
				resultByte, err := json.Marshal(resultMapFinal)
				if err != nil {
					impl.logger.Errorw("error while marshaling request", "err", err)
					return nil, err
				}
				historyDto.CodeEditorValue.Value = string(resultByte)
			}
		}
		historyDto.ExternalSecretType = config.ExternalSecretType
		historyDto.RoleARN = config.RoleARN
		historyDto.ESOSubPath = config.ESOSubPath
		if config.External {
			if config.ExternalSecretType == globalUtil.KubernetesSecret {
				externalSecretData, err := json.Marshal(config.ExternalSecret)
				if err != nil {
					impl.logger.Errorw("error in marshaling external secret data", "err", err)
				}
				if len(externalSecretData) > 0 {
					historyDto.CodeEditorValue.DisplayName = ExternalSecretDisplayName
					historyDto.CodeEditorValue.Value = string(externalSecretData)
				}
			} else if config.IsESOExternalSecretType() {
				externalSecretDataBytes, jErr := json.Marshal(config.ESOSecretData)
				if jErr != nil {
					impl.logger.Errorw("error in marshaling eso secret data", "esoSecretData", config.ESOSecretData, "err", jErr)
					return nil, jErr
				}
				if len(externalSecretDataBytes) > 0 {
					historyDto.CodeEditorValue.DisplayName = ESOSecretDataDisplayName
					historyDto.CodeEditorValue.Value = string(externalSecretDataBytes)
				}
			}
		}
	}
	return historyDto, nil
}

func (impl ConfigMapHistoryServiceImpl) GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int, configType repository.ConfigType, userHasAdminAccess bool) ([]*ComponentLevelHistoryDetailDto, error) {
	history, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId, configType)
	if err != nil {
		impl.logger.Errorw("error in getting histories for cm/cs", "err", err, "wfrId", wfrId, "pipelineId", pipelineId)
		return nil, err
	}
	var configData []*bean.ConfigData
	var configList bean.ConfigList
	var secretList bean.SecretList
	if err != nil {
		return nil, err
	}
	if configType == repository.CONFIGMAP_TYPE {
		configList = bean.ConfigList{}
		if len(history.Data) > 0 {
			err := json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		configData = configList.ConfigData
	} else if configType == repository.SECRET_TYPE {
		secretList = bean.SecretList{}
		if len(history.Data) > 0 {
			err := json.Unmarshal([]byte(history.Data), &secretList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		configData = secretList.ConfigData
	}
	resolvedDataMap, variableSnapshotMap, err := impl.scopedVariableManager.GetResolvedCMCSHistoryDtos(ctx, configType, configList, history, secretList)
	if err != nil {
		return nil, err
	}

	var componentLevelHistoryData []*ComponentLevelHistoryDetailDto
	for _, config := range configData {
		componentLevelData, err := impl.ConvertConfigDataToComponentLevelDto(config, configType, userHasAdminAccess)
		if err != nil {
			impl.logger.Errorw("error in converting data to componentLevelData", "err", err)
		}
		componentLevelData.HistoryConfig.CodeEditorValue.VariableSnapshot = variableSnapshotMap[config.Name]
		componentLevelData.HistoryConfig.CodeEditorValue.ResolvedValue = string(resolvedDataMap[config.Name].Data)
		componentLevelHistoryData = append(componentLevelHistoryData, componentLevelData)
	}
	return componentLevelHistoryData, nil
}

func (impl ConfigMapHistoryServiceImpl) ConvertConfigDataToComponentLevelDto(config *bean.ConfigData, configType repository.ConfigType, userHasAdminAccess bool) (*ComponentLevelHistoryDetailDto, error) {
	historyDto := &HistoryDetailDto{
		Type:           config.Type,
		External:       &config.External,
		MountPath:      config.MountPath,
		SubPath:        &config.SubPath,
		FilePermission: config.FilePermission,
		CodeEditorValue: &HistoryDetailConfig{
			DisplayName: DataDisplayName,
			Value:       string(config.Data),
		},
	}
	var err error
	if configType == repository.SECRET_TYPE {
		if config.Data != nil {
			if !userHasAdminAccess {
				//removing keys and sending
				resultMap := make(map[string]string)
				resultMapFinal := make(map[string]string)
				err = json.Unmarshal(config.Data, &resultMap)
				if err != nil {
					impl.logger.Warnw("unmarshal failed", "error", err)
					return nil, err
				}
				for key, _ := range resultMap {
					//hard-coding values to show them as hidden to user
					resultMapFinal[key] = "*****"
				}
				resultByte, err := json.Marshal(resultMapFinal)
				if err != nil {
					impl.logger.Errorw("error while marshaling request", "err", err)
					return nil, err
				}
				historyDto.CodeEditorValue.Value = string(resultByte)
			}
		}
		historyDto.ExternalSecretType = config.ExternalSecretType
		historyDto.RoleARN = config.RoleARN
		historyDto.ESOSubPath = config.ESOSubPath
		if config.External {
			var externalSecretData []byte
			displayName := historyDto.CodeEditorValue.DisplayName
			if config.IsESOExternalSecretType() {
				displayName = ESOSecretDataDisplayName
				externalSecretData, err = json.Marshal(config.ESOSecretData)
				if err != nil {
					impl.logger.Errorw("error in marshaling external secret data", "err", err)
					return nil, err
				}
			} else {
				displayName = ExternalSecretDisplayName
				externalSecretData, err = json.Marshal(config.ExternalSecret)
				if err != nil {
					impl.logger.Errorw("error in marshaling external secret data", "err", err)
					return nil, err
				}
			}
			if len(externalSecretData) > 0 {
				historyDto.CodeEditorValue.DisplayName = displayName
				historyDto.CodeEditorValue.Value = string(externalSecretData)
			}
		}
	}
	componentLevelData := &ComponentLevelHistoryDetailDto{
		ComponentName: config.Name,
		HistoryConfig: historyDto,
	}
	return componentLevelData, nil
}

func (impl ConfigMapHistoryServiceImpl) CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (cmId int, csId int, exists bool, err error) {
	cmHistory, cmErr := impl.configMapHistoryRepository.GetDeployedHistoryForPipelineIdOnTime(pipelineId, deployedOn, repository.SECRET_TYPE)
	if cmErr != nil && !errors.Is(cmErr, pg.ErrNoRows) {
		impl.logger.Errorw("error in checking if config map history exists for pipelineId and deployedOn", "err", cmErr, "pipelineId", pipelineId, "deployedOn", deployedOn)
		return cmId, csId, exists, cmErr
	}
	if cmErr == nil {
		cmId = cmHistory.Id
	}
	csHistory, csErr := impl.configMapHistoryRepository.GetDeployedHistoryForPipelineIdOnTime(pipelineId, deployedOn, repository.SECRET_TYPE)
	if csErr != nil && !errors.Is(csErr, pg.ErrNoRows) {
		impl.logger.Errorw("error in checking if secret history exists for pipelineId and deployedOn", "err", csErr, "pipelineId", pipelineId, "deployedOn", deployedOn)
		return cmId, csId, exists, csErr
	}
	if csErr == nil {
		csId = csHistory.Id
	}
	if cmErr == nil && csErr == nil {
		exists = true
	}
	return cmId, csId, exists, nil
}

func (impl ConfigMapHistoryServiceImpl) GetConfigmapHistoryDataByDeployedOnAndPipelineId(ctx context.Context, pipelineId int, deployedOn time.Time, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfig, *bean2.DeploymentAndCmCsConfig, error) {
	secretConfigData, err := impl.getResolvedConfigData(ctx, pipelineId, deployedOn, repository.SECRET_TYPE, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getting resolved secret config data in case of previous deployments ", "pipelineId", pipelineId, "deployedOn", deployedOn, "err", err)
		return nil, nil, err
	}
	cmConfigData, err := impl.getResolvedConfigData(ctx, pipelineId, deployedOn, repository.CONFIGMAP_TYPE, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getting resolved cm config data in case of previous deployments ", "pipelineId", pipelineId, "deployedOn", deployedOn, "err", err)
		return nil, nil, err
	}

	return secretConfigData, cmConfigData, nil
}

func (impl *ConfigMapHistoryServiceImpl) getResolvedConfigData(ctx context.Context, pipelineId int, deployedOn time.Time, configType repository.ConfigType, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfig, error) {
	configsList := &bean3.ConfigsList{}
	secretsList := &bean3.SecretsList{}
	var err error
	history, err := impl.configMapHistoryRepository.GetDeployedHistoryByPipelineIdAndDeployedOn(pipelineId, deployedOn, configType)
	if err != nil {
		impl.logger.Errorw("error in getting deployed history by pipeline id and deployed on", "pipelineId", pipelineId, "deployedOn", deployedOn, "err", err)
		return nil, err
	}
	if configType == repository.SECRET_TYPE {
		_, secretsList, err = impl.getConfigDataRequestForHistory(history)
		if err != nil {
			impl.logger.Errorw("error in getting config data request for history", "err", err)
			return nil, err
		}
	} else if configType == repository.CONFIGMAP_TYPE {
		configsList, _, err = impl.getConfigDataRequestForHistory(history)
		if err != nil {
			impl.logger.Errorw("error in getting config data request for history", "cmCsHistory", history, "err", err)
			return nil, err
		}
	}

	resolvedDataMap, variableSnapshotMap, err := impl.scopedVariableManager.GetResolvedCMCSHistoryDtos(ctx, configType, adaptor.ReverseConfigListConvertor(*configsList), history, adaptor.ReverseSecretListConvertor(*secretsList))
	if err != nil {
		return nil, err
	}
	resolvedConfigDataList := make([]*bean3.ConfigData, 0, len(resolvedDataMap))
	for _, resolvedConfigData := range resolvedDataMap {
		resolvedConfigDataList = append(resolvedConfigDataList, adapter.ConvertConfigDataToPipelineConfigData(&resolvedConfigData))
	}
	configDataReq := &bean3.ConfigDataRequest{}
	var resourceType bean3.ResourceType
	if configType == repository.SECRET_TYPE {
		impl.encodeSecretDataFromNonAdminUsers(secretsList.ConfigData, userHasAdminAccess)
		impl.encodeSecretDataFromNonAdminUsers(resolvedConfigDataList, userHasAdminAccess)
		configDataReq.ConfigData = secretsList.ConfigData
		resourceType = bean3.CS
	} else if configType == repository.CONFIGMAP_TYPE {
		configDataReq.ConfigData = configsList.ConfigData
		resourceType = bean3.CM
	}

	configDataJson, err := utils.ConvertToJsonRawMessage(configDataReq)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting config data to json raw message", "pipelineId", pipelineId, "deployedOn", deployedOn, "err", err)
		return nil, err
	}
	resolvedConfigDataReq := &bean3.ConfigDataRequest{ConfigData: resolvedConfigDataList}
	resolvedConfigDataString, err := utils.ConvertToString(resolvedConfigDataReq)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting config data to json raw message", "pipelineId", pipelineId, "deployedOn", deployedOn, "err", err)
		return nil, err
	}
	resolvedConfigDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedConfigDataString)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in ConvertToJsonRawMessage for resolvedJson", "resolvedJson", resolvedConfigDataStringJson, "err", err)
		return nil, err
	}
	return bean2.NewDeploymentAndCmCsConfig().WithConfigData(configDataJson).WithResourceType(resourceType).
		WithVariableSnapshot(variableSnapshotMap).WithResolvedValue(resolvedConfigDataStringJson), nil
}

func (impl *ConfigMapHistoryServiceImpl) encodeSecretDataFromNonAdminUsers(configDataList []*bean3.ConfigData, userHasAdminAccess bool) {
	for _, config := range configDataList {
		if config.Data != nil {
			if !userHasAdminAccess {
				//removing keys and sending
				resultMap := make(map[string]string)
				resultMapFinal := make(map[string]string)
				err := json.Unmarshal(config.Data, &resultMap)
				if err != nil {
					impl.logger.Errorw("unmarshal failed", "error", err)
					return
				}
				for key, _ := range resultMap {
					//hard-coding values to show them as hidden to user
					resultMapFinal[key] = "*****"
				}
				config.Data, err = utils.ConvertToJsonRawMessage(resultMapFinal)
				if err != nil {
					impl.logger.Errorw("error while marshaling request", "err", err)
					return
				}
			}
		}
	}
}

func (impl ConfigMapHistoryServiceImpl) getConfigDataRequestForHistory(history *repository.ConfigmapAndSecretHistory) (*bean3.ConfigsList, *bean3.SecretsList, error) {

	configsList := &bean3.ConfigsList{}
	secretsList := &bean3.SecretsList{}
	if history.IsConfigmapHistorySecretType() {
		err := json.Unmarshal([]byte(history.Data), secretsList)
		if err != nil {
			impl.logger.Errorw("error while Unmarshal in secret history data", "error", err)
			return configsList, secretsList, err
		}
		return configsList, secretsList, nil
	} else if history.IsConfigmapHistoryConfigMapType() {
		err := json.Unmarshal([]byte(history.Data), configsList)
		if err != nil {
			impl.logger.Errorw("error while Unmarshal in config history data", "historyData", history.Data, "error", err)
			return configsList, secretsList, err
		}
		return configsList, secretsList, nil
	}
	return configsList, secretsList, nil
}
