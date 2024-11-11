package read

import (
	"context"
	"encoding/json"
	"errors"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/bean/configMapBean"
	"github.com/devtron-labs/devtron/pkg/configDiff/adaptor"
	bean3 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/configDiff/utils"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ConfigMapHistoryReadService interface {
	GetHistoryForDeployedCMCSById(ctx context.Context, id, pipelineId int, configType repository.ConfigType, componentName string, userHasAdminAccess bool) (*bean.HistoryDetailDto, error)
	GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType repository.ConfigType) (history *repository.ConfigmapAndSecretHistory, exists bool, cmCsNames []string, err error)
	GetDeployedHistoryList(pipelineId, baseConfigId int, configType repository.ConfigType, componentName string) ([]*bean.DeployedHistoryComponentMetadataDto, error)
	CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (cmId int, csId int, exists bool, err error)
	GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int, configType repository.ConfigType, userHasAdminAccess bool) ([]*bean.ComponentLevelHistoryDetailDto, error)
	ConvertConfigDataToComponentLevelDto(config *bean2.ConfigData, configType repository.ConfigType, userHasAdminAccess bool) (*bean.ComponentLevelHistoryDetailDto, error)
	GetConfigmapHistoryDataByDeployedOnAndPipelineId(ctx context.Context, pipelineId int, deployedOn time.Time, userHasAdminAccess bool) (*bean3.DeploymentAndCmCsConfig, *bean3.DeploymentAndCmCsConfig, error)
}

type ConfigMapHistoryReadServiceImpl struct {
	logger                     *zap.SugaredLogger
	configMapHistoryRepository repository.ConfigMapHistoryRepository
	scopedVariableManager      variables.ScopedVariableCMCSManager
}

func NewConfigMapHistoryReadService(
	logger *zap.SugaredLogger,
	configMapHistoryRepository repository.ConfigMapHistoryRepository,
	scopedVariableManager variables.ScopedVariableCMCSManager,
) *ConfigMapHistoryReadServiceImpl {
	return &ConfigMapHistoryReadServiceImpl{
		logger:                     logger,
		configMapHistoryRepository: configMapHistoryRepository,
		scopedVariableManager:      scopedVariableManager,
	}
}

func (impl *ConfigMapHistoryReadServiceImpl) GetHistoryForDeployedCMCSById(ctx context.Context, id, pipelineId int, configType repository.ConfigType, componentName string, userHasAdminAccess bool) (*bean.HistoryDetailDto, error) {
	history, err := impl.configMapHistoryRepository.GetHistoryForDeployedCMCSById(id, pipelineId, configType)
	if err != nil {
		impl.logger.Errorw("error in getting histories for cm/cs", "err", err, "id", id, "pipelineId", pipelineId)
		return nil, err
	}
	var configData []*bean2.ConfigData
	var configList bean2.ConfigList
	var secretList bean2.SecretList
	if configType == repository.CONFIGMAP_TYPE {
		configList = bean2.ConfigList{}
		if len(history.Data) > 0 {
			err := json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		configData = configList.ConfigData
	} else if configType == repository.SECRET_TYPE {
		secretList = bean2.SecretList{}
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

	config := &bean2.ConfigData{}
	for _, data := range configData {
		if data.Name == componentName {
			config = data
			break
		}
	}
	historyDto := &bean.HistoryDetailDto{
		Type:           config.Type,
		External:       &config.External,
		MountPath:      config.MountPath,
		SubPath:        &config.SubPath,
		FilePermission: config.FilePermission,
		CodeEditorValue: &bean.HistoryDetailConfig{
			DisplayName:      bean.DataDisplayName,
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
			if config.ExternalSecretType == util.KubernetesSecret {
				externalSecretData, err := json.Marshal(config.ExternalSecret)
				if err != nil {
					impl.logger.Errorw("error in marshaling external secret data", "err", err)
				}
				if len(externalSecretData) > 0 {
					historyDto.CodeEditorValue.DisplayName = bean.ExternalSecretDisplayName
					historyDto.CodeEditorValue.Value = string(externalSecretData)
				}
			} else if config.IsESOExternalSecretType() {
				externalSecretDataBytes, jErr := json.Marshal(config.ESOSecretData)
				if jErr != nil {
					impl.logger.Errorw("error in marshaling eso secret data", "esoSecretData", config.ESOSecretData, "err", jErr)
					return nil, jErr
				}
				if len(externalSecretDataBytes) > 0 {
					historyDto.CodeEditorValue.DisplayName = bean.ESOSecretDataDisplayName
					historyDto.CodeEditorValue.Value = string(externalSecretDataBytes)
				}
			}
		}
	}
	return historyDto, nil
}

func (impl *ConfigMapHistoryReadServiceImpl) GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType repository.ConfigType) (history *repository.ConfigmapAndSecretHistory, exists bool, cmCsNames []string, err error) {
	impl.logger.Debugw("received request, CheckIfHistoryExistsForPipelineIdAndWfrId", "pipelineId", pipelineId, "wfrId", wfrId)
	//checking if history exists for pipelineId and wfrId
	history, err = impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId, configType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return history, false, cmCsNames, err
	} else if err == pg.ErrNoRows {
		return history, false, cmCsNames, nil
	}
	var configData []*bean2.ConfigData
	if configType == repository.CONFIGMAP_TYPE {
		configList := bean2.ConfigList{}
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return history, false, cmCsNames, err
			}
		}
		configData = configList.ConfigData
	} else if configType == repository.SECRET_TYPE {
		secretList := bean2.SecretList{}
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

func (impl *ConfigMapHistoryReadServiceImpl) GetDeployedHistoryList(pipelineId, baseConfigId int, configType repository.ConfigType, componentName string) ([]*bean.DeployedHistoryComponentMetadataDto, error) {
	impl.logger.Debugw("received request, GetDeployedHistoryList", "pipelineId", pipelineId, "baseConfigId", baseConfigId)

	//checking if history exists for pipelineId and wfrId
	histories, err := impl.configMapHistoryRepository.GetDeployedHistoryList(pipelineId, baseConfigId, configType, componentName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting history list for pipelineId and baseConfigId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historyList []*bean.DeployedHistoryComponentMetadataDto
	for _, history := range histories {
		historyList = append(historyList, &bean.DeployedHistoryComponentMetadataDto{
			Id:               history.Id,
			DeployedOn:       history.DeployedOn,
			DeployedBy:       history.DeployedByEmailId,
			DeploymentStatus: history.DeploymentStatus,
		})
	}
	return historyList, nil
}

func (impl *ConfigMapHistoryReadServiceImpl) CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (cmId int, csId int, exists bool, err error) {
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

func (impl *ConfigMapHistoryReadServiceImpl) GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int, configType repository.ConfigType, userHasAdminAccess bool) ([]*bean.ComponentLevelHistoryDetailDto, error) {
	history, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId, configType)
	if err != nil {
		impl.logger.Errorw("error in getting histories for cm/cs", "err", err, "wfrId", wfrId, "pipelineId", pipelineId)
		return nil, err
	}
	var configData []*bean2.ConfigData
	var configList bean2.ConfigList
	var secretList bean2.SecretList
	if err != nil {
		return nil, err
	}
	if configType == repository.CONFIGMAP_TYPE {
		configList = bean2.ConfigList{}
		if len(history.Data) > 0 {
			err := json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		configData = configList.ConfigData
	} else if configType == repository.SECRET_TYPE {
		secretList = bean2.SecretList{}
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

	var componentLevelHistoryData []*bean.ComponentLevelHistoryDetailDto
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

func (impl *ConfigMapHistoryReadServiceImpl) ConvertConfigDataToComponentLevelDto(config *bean2.ConfigData, configType repository.ConfigType, userHasAdminAccess bool) (*bean.ComponentLevelHistoryDetailDto, error) {
	historyDto := &bean.HistoryDetailDto{
		Type:           config.Type,
		External:       &config.External,
		MountPath:      config.MountPath,
		SubPath:        &config.SubPath,
		FilePermission: config.FilePermission,
		CodeEditorValue: &bean.HistoryDetailConfig{
			DisplayName: bean.DataDisplayName,
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
				displayName = bean.ESOSecretDataDisplayName
				externalSecretData, err = json.Marshal(config.ESOSecretData)
				if err != nil {
					impl.logger.Errorw("error in marshaling external secret data", "err", err)
					return nil, err
				}
			} else {
				displayName = bean.ExternalSecretDisplayName
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
	componentLevelData := &bean.ComponentLevelHistoryDetailDto{
		ComponentName: config.Name,
		HistoryConfig: historyDto,
	}
	return componentLevelData, nil
}

func (impl *ConfigMapHistoryReadServiceImpl) GetConfigmapHistoryDataByDeployedOnAndPipelineId(ctx context.Context, pipelineId int, deployedOn time.Time, userHasAdminAccess bool) (*bean3.DeploymentAndCmCsConfig, *bean3.DeploymentAndCmCsConfig, error) {
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

func (impl *ConfigMapHistoryReadServiceImpl) getResolvedConfigData(ctx context.Context, pipelineId int, deployedOn time.Time, configType repository.ConfigType, userHasAdminAccess bool) (*bean3.DeploymentAndCmCsConfig, error) {
	configsList := &bean4.ConfigsList{}
	secretsList := &bean4.SecretsList{}
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
	resolvedConfigDataList := make([]*bean4.ConfigData, 0, len(resolvedDataMap))
	for _, resolvedConfigData := range resolvedDataMap {
		resolvedConfigDataList = append(resolvedConfigDataList, adapter.ConvertConfigDataToPipelineConfigData(&resolvedConfigData))
	}
	configDataReq := &bean4.ConfigDataRequest{}
	var resourceType bean4.ResourceType
	if configType == repository.SECRET_TYPE {
		impl.encodeSecretDataFromNonAdminUsers(secretsList.ConfigData, userHasAdminAccess)
		impl.encodeSecretDataFromNonAdminUsers(resolvedConfigDataList, userHasAdminAccess)
		configDataReq.ConfigData = secretsList.ConfigData
		resourceType = bean4.CS
	} else if configType == repository.CONFIGMAP_TYPE {
		configDataReq.ConfigData = configsList.ConfigData
		resourceType = bean4.CM
	}

	configDataJson, err := utils.ConvertToJsonRawMessage(configDataReq)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting config data to json raw message", "pipelineId", pipelineId, "deployedOn", deployedOn, "err", err)
		return nil, err
	}
	resolvedConfigDataReq := &bean4.ConfigDataRequest{ConfigData: resolvedConfigDataList}
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
	return bean3.NewDeploymentAndCmCsConfig().WithConfigData(configDataJson).WithResourceType(resourceType).
		WithVariableSnapshot(variableSnapshotMap).WithResolvedValue(resolvedConfigDataStringJson), nil
}

func (impl *ConfigMapHistoryReadServiceImpl) encodeSecretDataFromNonAdminUsers(configDataList []*bean4.ConfigData, userHasAdminAccess bool) {
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

func (impl *ConfigMapHistoryReadServiceImpl) getConfigDataRequestForHistory(history *repository.ConfigmapAndSecretHistory) (*bean4.ConfigsList, *bean4.SecretsList, error) {

	configsList := &bean4.ConfigsList{}
	secretsList := &bean4.SecretsList{}
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
