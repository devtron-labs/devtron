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

type PrePostCdScriptHistoryService interface {
	CreatePrePostCdScriptHistory(pipeline *pipelineConfig.Pipeline, tx *pg.Tx, stage repository.CdStageType, deployed bool, deployedBy int32, deployedOn time.Time) error
	GetHistoryForDeployedPrePostCdScript(pipelineId int, stage repository.CdStageType) ([]*PrePostCdScriptHistoryDto, error)
}

type PrePostCdScriptHistoryServiceImpl struct {
	logger                           *zap.SugaredLogger
	prePostCdScriptHistoryRepository repository.PrePostCdScriptHistoryRepository
	configMapRepository              chartConfig.ConfigMapRepository
	configMapHistoryService          ConfigMapHistoryService
}

func NewPrePostCdScriptHistoryServiceImpl(logger *zap.SugaredLogger, prePostCdScriptHistoryRepository repository.PrePostCdScriptHistoryRepository,
	configMapRepository chartConfig.ConfigMapRepository, configMapHistoryService ConfigMapHistoryService) *PrePostCdScriptHistoryServiceImpl {
	return &PrePostCdScriptHistoryServiceImpl{
		logger:                           logger,
		prePostCdScriptHistoryRepository: prePostCdScriptHistoryRepository,
		configMapRepository:              configMapRepository,
		configMapHistoryService:          configMapHistoryService,
	}
}

func (impl PrePostCdScriptHistoryServiceImpl) CreatePrePostCdScriptHistory(pipeline *pipelineConfig.Pipeline, tx *pg.Tx, stage repository.CdStageType, deployed bool, deployedBy int32, deployedOn time.Time) (err error) {
	historyModel := &repository.PrePostCdScriptHistory{
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
	if stage == repository.PRE_CD_TYPE {
		historyModel.Stage = repository.PRE_CD_TYPE
		historyModel.Script = pipeline.PreStageConfig
		historyModel.ConfigMapSecretNames = pipeline.PreStageConfigMapSecretNames
		historyModel.ExecInEnv = pipeline.RunPreStageInEnv
		historyModel.TriggerType = pipeline.PreTriggerType
	} else if stage == repository.POST_CD_TYPE {
		historyModel.Stage = repository.POST_CD_TYPE
		historyModel.Script = pipeline.PostStageConfig
		historyModel.ConfigMapSecretNames = pipeline.PostStageConfigMapSecretNames
		historyModel.ExecInEnv = pipeline.RunPostStageInEnv
		historyModel.TriggerType = pipeline.PostTriggerType
	}
	configMapData, secretData, err := impl.GetConfigMapSecretData(pipeline, stage)
	if err != nil {
		impl.logger.Errorw("err in getting cm and cs data for cd config history entry", "err", err)
		return err
	}
	historyModel.ConfigMapData = configMapData
	historyModel.SecretData = secretData
	if tx != nil {
		_, err = impl.prePostCdScriptHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.prePostCdScriptHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for pre/post cd script", "err", err)
		return err
	}
	return nil
}

func (impl PrePostCdScriptHistoryServiceImpl) GetHistoryForDeployedPrePostCdScript(pipelineId int, stage repository.CdStageType) ([]*PrePostCdScriptHistoryDto, error) {
	histories, err := impl.prePostCdScriptHistoryRepository.GetHistoryForDeployedPrePostScriptByStage(pipelineId, stage)
	if err != nil {
		impl.logger.Errorw("error in getting pre/post cd script history", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historiesDto []*PrePostCdScriptHistoryDto
	for _, history := range histories {
		configMapList := ConfigList{}
		if len(history.ConfigMapData) > 0 {
			err := json.Unmarshal([]byte(history.ConfigMapData), &configMapList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		secretList := ConfigList{}
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

		historyDto := &PrePostCdScriptHistoryDto{
			Id:                   history.Id,
			PipelineId:           history.PipelineId,
			Script:               history.Script,
			Stage:                string(history.Stage),
			ConfigMapSecretNames: configMapSecretNames,
			ConfigMapData:        configMapList.ConfigData,
			SecretData:           secretList.ConfigData,
			TriggerType:          string(history.TriggerType),
			ExecInEnv:            history.ExecInEnv,
			Deployed:             history.Deployed,
			DeployedOn:           history.DeployedOn,
			DeployedBy:           history.DeployedBy,
		}
		historiesDto = append(historiesDto, historyDto)
	}
	return historiesDto, nil
}

func (impl PrePostCdScriptHistoryServiceImpl) GetConfigMapSecretData(pipeline *pipelineConfig.Pipeline, stage repository.CdStageType) (configMapData, secretData string, err error) {
	var configMapSecretNames PrePostStageConfigMapSecretNames
	if stage == repository.PRE_CD_TYPE {
		if pipeline.PreStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(pipeline.PreStageConfigMapSecretNames), &configMapSecretNames)
			if err != nil {
				impl.logger.Error("error in un-marshaling pre stage config map secret names", "err", err)
				return "", "", err
			}
		}
	} else if stage == repository.POST_CD_TYPE {
		if pipeline.PostStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(pipeline.PostStageConfigMapSecretNames), &configMapSecretNames)
			if err != nil {
				impl.logger.Error("error in un-marshaling post stage config map secret names", "err", err)
				return "", "", err
			}
		}
	}
	if len(configMapSecretNames.ConfigMaps) == 0 && len(configMapSecretNames.Secrets) == 0 {
		return "", "", nil
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
	if len(configMapSecretNames.ConfigMaps) > 0 {
		configMapData, err = impl.configMapHistoryService.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, repository.CONFIGMAP_TYPE, configMapSecretNames.ConfigMaps)
		if err != nil {
			impl.logger.Errorw("error in getting filtered config map data", "err", err)
			return "", "", err
		}
	}
	if len(configMapSecretNames.Secrets) > 0 {
		secretData, err = impl.configMapHistoryService.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, repository.SECRET_TYPE, configMapSecretNames.Secrets)
		if err != nil {
			impl.logger.Errorw("error in getting filtered secret data", "err", err)
			return "", "", err
		}
	}
	return configMapData, secretData, nil
}
