package history

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ConfigMapHistoryService interface {
	CreateHistoryFromAppLevelConfig(appLevelConfig *chartConfig.ConfigMapAppModel, configType repository.ConfigType) error
	CreateHistoryFromEnvLevelConfig(envLevelConfig *chartConfig.ConfigMapEnvModel, configType repository.ConfigType) error
	CreateConfigMapHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, deployedOn time.Time, deployedBy int32) error
	MergeAppLevelAndEnvLevelConfigs(appLevelConfig *chartConfig.ConfigMapAppModel, envLevelConfig *chartConfig.ConfigMapEnvModel, configType repository.ConfigType, configMapSecretNames []string) (string, error)
	GetHistoryForDeployedCMCS(pipelineId int, configType repository.ConfigType) ([]*ConfigMapAndSecretHistoryDto, error)
}

type ConfigMapHistoryServiceImpl struct {
	logger                     *zap.SugaredLogger
	configMapHistoryRepository repository.ConfigMapHistoryRepository
	pipelineRepository         pipelineConfig.PipelineRepository
	configMapRepository        chartConfig.ConfigMapRepository
}

func NewConfigMapHistoryServiceImpl(logger *zap.SugaredLogger,
	configMapHistoryRepository repository.ConfigMapHistoryRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	configMapRepository chartConfig.ConfigMapRepository) *ConfigMapHistoryServiceImpl {
	return &ConfigMapHistoryServiceImpl{
		logger:                     logger,
		configMapHistoryRepository: configMapHistoryRepository,
		pipelineRepository:         pipelineRepository,
		configMapRepository:        configMapRepository,
	}
}

func (impl ConfigMapHistoryServiceImpl) CreateHistoryFromAppLevelConfig(appLevelConfig *chartConfig.ConfigMapAppModel, configType repository.ConfigType) error {
	pipelines, err := impl.pipelineRepository.FindActiveByAppId(appLevelConfig.AppId)
	if err != nil {
		impl.logger.Errorw("err in getting pipelines, CreateHistoryFromAppLevelConfig", "err", err, "appLevelConfig", appLevelConfig)
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
		_, err = impl.configMapHistoryRepository.CreateHistory(historyModel)
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
		impl.logger.Errorw("err in getting pipelines, CreateHistoryFromAppLevelConfig", "err", err, "envLevelConfig", envLevelConfig)
		return err
	}
	appLevelConfig, err := impl.configMapRepository.GetByAppIdAppLevel(envLevelConfig.AppId)
	if err != nil {
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
		_, err = impl.configMapHistoryRepository.CreateHistory(historyModel)
		if err != nil {
			impl.logger.Errorw("error in creating new entry for CM/CS history", "historyModel", historyModel)
			return err
		}
	}
	return nil
}

func (impl ConfigMapHistoryServiceImpl) CreateConfigMapHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, deployedOn time.Time, deployedBy int32) error {
	//creating history for configmaps, secrets(if any)
	appLevelConfig, err := impl.configMapRepository.GetByAppIdAppLevel(pipeline.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting app level config", "err", err, "appId", pipeline.AppId)
		return err
	}
	envLevelConfig, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting env level config", "err", err, "appId", pipeline.AppId)
		return err
	}
	configMapData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, repository.CONFIGMAP_TYPE, nil)
	if err != nil {
		impl.logger.Errorw("err in merging app and env level configs", "err", err)
		return err
	}
	historyModel := &repository.ConfigmapAndSecretHistory{
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
	_, err = impl.configMapHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for cm history", "historyModel", historyModel)
		return err
	}
	secretData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, repository.SECRET_TYPE, nil)
	if err != nil {
		impl.logger.Errorw("err in merging app and env level configs", "err", err)
		return err
	}
	//using old model, updating secret data
	historyModel.DataType = repository.SECRET_TYPE
	historyModel.Id = 0
	historyModel.Data = secretData
	_, err = impl.configMapHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for secret history", "historyModel", historyModel)
		return err
	}
	return nil
}

func (impl ConfigMapHistoryServiceImpl) MergeAppLevelAndEnvLevelConfigs(appLevelConfig *chartConfig.ConfigMapAppModel, envLevelConfig *chartConfig.ConfigMapEnvModel, configType repository.ConfigType, configMapSecretNames []string) (string, error) {
	var configDataAppLevel string
	var configDataEnvLevel string
	if configType == repository.CONFIGMAP_TYPE {
		configDataAppLevel = appLevelConfig.ConfigMapData
		configDataEnvLevel = envLevelConfig.ConfigMapData
	} else if configType == repository.SECRET_TYPE {
		configDataAppLevel = appLevelConfig.SecretData
		configDataEnvLevel = envLevelConfig.SecretData
	}
	configListAppLevel := &ConfigList{}
	if len(configDataAppLevel) > 0 {
		err := json.Unmarshal([]byte(configDataAppLevel), configListAppLevel)
		if err != nil {
			impl.logger.Debugw("error while Unmarshal", "err", err)
			return "", err
		}
	}
	configListEnvLevel := &ConfigList{}
	if len(configDataEnvLevel) > 0 {
		err := json.Unmarshal([]byte(configDataEnvLevel), configListEnvLevel)
		if err != nil {
			impl.logger.Debugw("error while Unmarshal", "err", err)
			return "", err
		}
	}
	var finalConfigs []*ConfigData
	envLevelConfigs := make(map[string]bool)
	var filterNameMap map[string]bool
	for _, name := range configMapSecretNames {
		filterNameMap[name] = true
	}
	//if filter name map is not empty, to add configs by filtering names
	//if filter name map is empty, adding all env level configs to final configs
	for _, item := range configListEnvLevel.ConfigData {
		if _, ok := filterNameMap[item.Name]; ok || len(filterNameMap) == 0 {
			//adding all env configs whose name is in filter name map
			envLevelConfigs[item.Name] = true
			finalConfigs = append(finalConfigs, item)
		}
	}
	for _, item := range configListAppLevel.ConfigData {
		//if filter name map is not empty, adding all global configs which are not present in env level and are present in filter name map to final configs
		//if filter name map is empty,adding all global configs which are not present in env level to final configs
		if _, ok := envLevelConfigs[item.Name]; !ok {
			if _, ok = filterNameMap[item.Name]; ok || len(filterNameMap) == 0 {
				finalConfigs = append(finalConfigs, item)
			}
		}
	}
	var finalConfigList ConfigList
	finalConfigList.ConfigData = finalConfigs
	finalConfigDataByte, err := json.Marshal(finalConfigList)
	if err != nil {
		impl.logger.Errorw("error in marshaling config", "err", err)
		return "", err
	}
	return string(finalConfigDataByte), err
}

func (impl ConfigMapHistoryServiceImpl) GetHistoryForDeployedCMCS(pipelineId int, configType repository.ConfigType) ([]*ConfigMapAndSecretHistoryDto, error) {
	histories, err := impl.configMapHistoryRepository.GetHistoryForDeployedCMCS(pipelineId, configType)
	if err != nil {
		impl.logger.Errorw("error in getting histories for cm/cs", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historiesDto []*ConfigMapAndSecretHistoryDto
	for _, history := range histories {
		configList := ConfigList{}
		if len(history.Data) > 0 {
			err := json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		historyDto := &ConfigMapAndSecretHistoryDto{
			Id:         history.Id,
			PipelineId: history.PipelineId,
			DataType:   string(history.DataType),
			ConfigData: configList.ConfigData,
			Deployed:   history.Deployed,
			DeployedOn: history.DeployedOn,
			DeployedBy: history.DeployedBy,
			AuditLog: sql.AuditLog{
				CreatedBy: history.CreatedBy,
				CreatedOn: history.CreatedOn,
				UpdatedBy: history.UpdatedBy,
				UpdatedOn: history.UpdatedOn,
			},
		}
		historiesDto = append(historiesDto, historyDto)
	}
	return historiesDto, nil
}
