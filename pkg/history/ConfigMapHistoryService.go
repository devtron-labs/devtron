package history

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ConfigMapHistoryService interface {
	CreateHistoryFromAppLevelConfig(appLevelConfig *chartConfig.ConfigMapAppModel, configType history.ConfigType) error
	CreateHistoryFromEnvLevelConfig(envLevelConfig *chartConfig.ConfigMapEnvModel, configType history.ConfigType) error
	CreateConfigMapHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, deployedOn time.Time, deployedBy int32) error
	MergeAppLevelAndEnvLevelConfigs(appLevelConfig *chartConfig.ConfigMapAppModel, envLevelConfig *chartConfig.ConfigMapEnvModel, configType history.ConfigType) (string, error)
}

type ConfigMapHistoryServiceImpl struct {
	logger                     *zap.SugaredLogger
	configMapHistoryRepository history.ConfigMapHistoryRepository
	pipelineRepository         pipelineConfig.PipelineRepository
	configMapRepository        chartConfig.ConfigMapRepository
}

func NewConfigMapHistoryServiceImpl(logger *zap.SugaredLogger,
	configMapHistoryRepository history.ConfigMapHistoryRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	configMapRepository chartConfig.ConfigMapRepository) *ConfigMapHistoryServiceImpl {
	return &ConfigMapHistoryServiceImpl{
		logger:                     logger,
		configMapHistoryRepository: configMapHistoryRepository,
		pipelineRepository:         pipelineRepository,
		configMapRepository:        configMapRepository,
	}
}

func (impl ConfigMapHistoryServiceImpl) CreateHistoryFromAppLevelConfig(appLevelConfig *chartConfig.ConfigMapAppModel, configType history.ConfigType) error {
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
		configData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, configType)
		if err != nil {
			impl.logger.Errorw("err in merging app and env level configs", "err", err)
			return err
		}
		historyModel := &history.ConfigmapAndSecretHistory{
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

func (impl ConfigMapHistoryServiceImpl) CreateHistoryFromEnvLevelConfig(envLevelConfig *chartConfig.ConfigMapEnvModel, configType history.ConfigType) error {
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
		configData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, configType)
		if err != nil {
			impl.logger.Errorw("err in merging app and env level configs", "err", err)
			return err
		}
		historyModel := &history.ConfigmapAndSecretHistory{
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
	configMapData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, history.CONFIGMAP_TYPE)
	if err != nil {
		impl.logger.Errorw("err in merging app and env level configs", "err", err)
		return err
	}
	historyModel := &history.ConfigmapAndSecretHistory{
		PipelineId: pipeline.Id,
		DataType:   history.CONFIGMAP_TYPE,
		Deployed:   true,
		DeployedBy: deployedBy,
		DeployedOn: deployedOn,
		Data:       configMapData,
	}
	if appLevelConfig.UpdatedOn.After(envLevelConfig.UpdatedOn) {
		historyModel.AuditLog = sql.AuditLog{
			CreatedBy: appLevelConfig.CreatedBy,
			CreatedOn: appLevelConfig.CreatedOn,
			UpdatedBy: appLevelConfig.UpdatedBy,
			UpdatedOn: appLevelConfig.UpdatedOn,
		}
	} else {
		historyModel.AuditLog = sql.AuditLog{
			CreatedBy: envLevelConfig.CreatedBy,
			CreatedOn: envLevelConfig.CreatedOn,
			UpdatedBy: envLevelConfig.UpdatedBy,
			UpdatedOn: envLevelConfig.UpdatedOn,
		}
	}
	_, err = impl.configMapHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for cm history", "historyModel", historyModel)
		return err
	}
	secretData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, history.SECRET_TYPE)
	if err != nil {
		impl.logger.Errorw("err in merging app and env level configs", "err", err)
		return err
	}
	//using old model, updating secret data
	historyModel.DataType = history.SECRET_TYPE
	historyModel.Id = 0
	historyModel.Data = secretData
	_, err = impl.configMapHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for secret history", "historyModel", historyModel)
		return err
	}
	return nil
}

func (impl ConfigMapHistoryServiceImpl) MergeAppLevelAndEnvLevelConfigs(appLevelConfig *chartConfig.ConfigMapAppModel, envLevelConfig *chartConfig.ConfigMapEnvModel, configType history.ConfigType) (string, error) {
	var configDataAppLevel string
	var configDataEnvLevel string
	if configType == history.CONFIGMAP_TYPE {
		configDataAppLevel = appLevelConfig.ConfigMapData
		configDataEnvLevel = envLevelConfig.ConfigMapData
	} else if configType == history.SECRET_TYPE {
		configDataAppLevel = appLevelConfig.SecretData
		configDataEnvLevel = envLevelConfig.SecretData
	}
	configsListAppLevel := &ConfigsList{}
	if len(configDataAppLevel) > 0 {
		err := json.Unmarshal([]byte(configDataAppLevel), configsListAppLevel)
		if err != nil {
			impl.logger.Debugw("error while Unmarshal", "err", err)
			return "", err
		}
	}
	configsListEnvLevel := &ConfigsList{}
	if len(configDataEnvLevel) > 0 {
		err := json.Unmarshal([]byte(configDataEnvLevel), configsListEnvLevel)
		if err != nil {
			impl.logger.Debugw("error while Unmarshal", "err", err)
			return "", err
		}
	}
	var finalConfigs []*ConfigData
	envLevelConfigs := make(map[string]json.RawMessage)
	//adding all env level configs to final configs as these won't get affected by global changes
	for _, item := range configsListEnvLevel.ConfigData {
		envLevelConfigs[item.Name] = item.Data
		finalConfigs = append(finalConfigs, item)
	}
	for _, item := range configsListAppLevel.ConfigData {
		//adding all global configs which are not present in env level to final configs
		if _, ok := envLevelConfigs[item.Name]; !ok {
			finalConfigs = append(finalConfigs, item)
		}
	}
	var finalConfigsList ConfigsList
	finalConfigsList.ConfigData = finalConfigs
	finalConfigDataByte, err := json.Marshal(finalConfigsList)
	if err != nil {
		impl.logger.Errorw("error in marshaling config", "err", err)
		return "", err
	}
	return string(finalConfigDataByte), err
}
