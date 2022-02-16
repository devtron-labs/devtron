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

type CdConfigHistoryService interface {
	CreateCdConfigHistory(pipeline *pipelineConfig.Pipeline, tx *pg.Tx, stage history.CdStageType, deployed bool, deployedBy int32, deployedOn time.Time) error
	GetHistoryForDeployedCdConfig(pipelineId int, stage history.CdStageType) ([]*CdConfigHistoryDto, error)
}

type CdConfigHistoryServiceImpl struct {
	logger                    *zap.SugaredLogger
	cdConfigHistoryRepository history.CdConfigHistoryRepository
	configMapRepository       chartConfig.ConfigMapRepository
	configMapHistoryService   ConfigMapHistoryService
}

func NewCdConfigHistoryServiceImpl(logger *zap.SugaredLogger, cdConfigHistoryRepository history.CdConfigHistoryRepository,
	configMapRepository chartConfig.ConfigMapRepository, configMapHistoryService ConfigMapHistoryService) *CdConfigHistoryServiceImpl {
	return &CdConfigHistoryServiceImpl{
		logger:                    logger,
		cdConfigHistoryRepository: cdConfigHistoryRepository,
		configMapRepository:       configMapRepository,
		configMapHistoryService:   configMapHistoryService,
	}
}

func (impl CdConfigHistoryServiceImpl) CreateCdConfigHistory(pipeline *pipelineConfig.Pipeline, tx *pg.Tx, stage history.CdStageType, deployed bool, deployedBy int32, deployedOn time.Time) (err error) {
	historyModel := &history.CdConfigHistory{
		PipelineId: pipeline.Id,
		Deployed:   deployed,
		DeployedBy: deployedBy,
		DeployedOn: deployedOn,
		AuditLog: sql.AuditLog{
			CreatedOn: pipeline.CreatedOn,
			CreatedBy: pipeline.CreatedBy,
			UpdatedOn: pipeline.UpdatedOn,
			UpdatedBy: pipeline.UpdatedBy,
		},
	}
	if stage == history.PRE_CD_TYPE {
		historyModel.Stage = history.PRE_CD_TYPE
		historyModel.Config = pipeline.PreStageConfig
		historyModel.ConfigMapSecretNames = pipeline.PreStageConfigMapSecretNames
		historyModel.ExecInEnv = pipeline.RunPreStageInEnv
	} else if stage == history.POST_CD_TYPE {
		historyModel.Stage = history.POST_CD_TYPE
		historyModel.Config = pipeline.PostStageConfig
		historyModel.ConfigMapSecretNames = pipeline.PostStageConfigMapSecretNames
		historyModel.ExecInEnv = pipeline.RunPostStageInEnv
	}
	configMapData, secretData, err := impl.GetConfigMapSecretData(pipeline, stage)
	if err != nil {
		impl.logger.Errorw("err in getting cm and cs data for cd config history entry", "err", err)
		return err
	}
	historyModel.ConfigMapData = configMapData
	historyModel.SecretData = secretData
	if tx != nil {
		_, err = impl.cdConfigHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.cdConfigHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for cd config", "err", err)
		return err
	}
	return nil
}

func (impl CdConfigHistoryServiceImpl) GetHistoryForDeployedCdConfig(pipelineId int, stage history.CdStageType) ([]*CdConfigHistoryDto, error) {
	histories, err := impl.cdConfigHistoryRepository.GetHistoryForDeployedCdConfigByStage(pipelineId, stage)
	if err != nil {
		impl.logger.Errorw("error in getting cd config history", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historiesDto []*CdConfigHistoryDto
	for _, history := range histories {
		configMapList := ConfigsList{}
		if len(history.ConfigMapData) > 0 {
			err := json.Unmarshal([]byte(history.ConfigMapData), &configMapList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		secretList := ConfigsList{}
		if len(history.SecretData) > 0 {
			err := json.Unmarshal([]byte(history.SecretData), &secretList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		var configMapSecretNames PrePostStageConfigMapSecretNames
		if history.ConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(history.ConfigMapSecretNames), &configMapSecretNames)
			if err != nil {
				impl.logger.Error("error in un-marshaling config map secret names", "err", err)
				return nil, err
			}
		}

		historyDto := &CdConfigHistoryDto{
			Id:                   history.Id,
			PipelineId:           history.PipelineId,
			Config:               history.Config,
			Stage:                string(history.Stage),
			ConfigMapSecretNames: configMapSecretNames,
			ConfigMapData:        configMapList.ConfigData,
			SecretData:           secretList.ConfigData,
			ExecInEnv:            history.ExecInEnv,
			Deployed:             history.Deployed,
			DeployedOn:           history.DeployedOn,
			DeployedBy:           history.DeployedBy,
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

func (impl CdConfigHistoryServiceImpl) GetConfigMapSecretData(pipeline *pipelineConfig.Pipeline, stage history.CdStageType) (configMapData, secretData string, err error) {
	var configMapSecretNames PrePostStageConfigMapSecretNames
	if stage == history.PRE_CD_TYPE {
		if pipeline.PreStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(pipeline.PreStageConfigMapSecretNames), &configMapSecretNames)
			if err != nil {
				impl.logger.Error("error in un-marshaling pre stage config map secret names", "err", err)
				return "", "", err
			}
		}
	} else if stage == history.POST_CD_TYPE {
		if pipeline.PostStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(pipeline.PostStageConfigMapSecretNames), &configMapSecretNames)
			if err != nil {
				impl.logger.Error("error in un-marshaling post stage config map secret names", "err", err)
				return "", "", err
			}
		}
	}
	var configmapNameMap map[string]bool
	var secretNameMap map[string]bool
	for _, name := range configMapSecretNames.ConfigMaps {
		configmapNameMap[name] = true
	}
	for _, name := range configMapSecretNames.Secrets {
		secretNameMap[name] = true
	}
	appLevelConfig, err := impl.configMapRepository.GetByAppIdAppLevel(pipeline.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting app level config", "err", err, "appId", pipeline.AppId)
		return "", "", err
	}
	envLevelConfig, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting env level config", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
		return "", "", err
	}
	if len(configmapNameMap) > 0 {

		configMapData, err = impl.configMapHistoryService.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, history.CONFIGMAP_TYPE, configmapNameMap)
		if err != nil {
			impl.logger.Errorw("error in getting filtered config map data", "err", err)
			return "", "", err
		}
	}
	if len(secretNameMap) > 0 {
		secretData, err = impl.configMapHistoryService.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, history.SECRET_TYPE, secretNameMap)
		if err != nil {
			impl.logger.Errorw("error in getting filtered secret data", "err", err)
			return "", "", err
		}
	}
	return configMapData, secretData, nil
}
