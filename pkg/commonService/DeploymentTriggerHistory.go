package commonService

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type DeploymentTriggerHistoryService interface {
	CreateHistoriesForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, strategy *chartConfig.PipelineStrategy, envOverride *chartConfig.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) error
	CreateChartsHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, envOverride *chartConfig.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) error
	CreateConfigMapHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, deployedOn time.Time, deployedBy int32) error
	CreateStrategyHistoryForDeploymentTrigger(strategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32) error
}

type DeploymentTriggerHistoryServiceImpl struct {
	logger                            *zap.SugaredLogger
	chartHistoryRepository            history.ChartHistoryRepository
	configMapHistoryRepository        history.ConfigMapHistoryRepository
	pipelineStrategyHistoryRepository history.PipelineStrategyHistoryRepository
	chartRepository                   chartConfig.ChartRepository
	configMapRepository               chartConfig.ConfigMapRepository
}

func NewDeploymentTriggerHistoryServiceImpl(logger *zap.SugaredLogger,
	chartHistoryRepository history.ChartHistoryRepository,
	configMapHistoryRepository history.ConfigMapHistoryRepository,
	pipelineStrategyHistoryRepository history.PipelineStrategyHistoryRepository,
	chartRepository chartConfig.ChartRepository,
	configMapRepository chartConfig.ConfigMapRepository) *DeploymentTriggerHistoryServiceImpl {
	return &DeploymentTriggerHistoryServiceImpl{
		logger:                            logger,
		chartHistoryRepository:            chartHistoryRepository,
		configMapHistoryRepository:        configMapHistoryRepository,
		pipelineStrategyHistoryRepository: pipelineStrategyHistoryRepository,
		chartRepository:                   chartRepository,
		configMapRepository:               configMapRepository,
	}
}

func (impl DeploymentTriggerHistoryServiceImpl) CreateHistoriesForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, strategy *chartConfig.PipelineStrategy, envOverride *chartConfig.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) error {
	//creating history for deployment template
	err := impl.CreateChartsHistoryForDeploymentTrigger(pipeline, envOverride, renderedImageTemplate, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating charts history for deployment trigger", "err", err)
		return err
	}
	err = impl.CreateConfigMapHistoryForDeploymentTrigger(pipeline, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating CM/CS history for deployment trigger", "err", err)
		return err
	}
	err = impl.CreateStrategyHistoryForDeploymentTrigger(strategy, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating strategy history for deployment trigger", "err", err)
		return err
	}
	return nil
}

func (impl DeploymentTriggerHistoryServiceImpl) CreateChartsHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, envOverride *chartConfig.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) error {
	historyModel := &history.ChartsHistory{
		PipelineId:              pipeline.Id,
		ImageDescriptorTemplate: renderedImageTemplate,
		Deployed:                true,
		DeployedBy:              deployedBy,
		DeployedOn:              deployedOn,
	}
	if envOverride.IsOverride {
		historyModel.Template = envOverride.EnvOverrideValues
		historyModel.AuditLog = sql.AuditLog{
			CreatedOn: envOverride.CreatedOn,
			CreatedBy: envOverride.CreatedBy,
			UpdatedOn: envOverride.UpdatedOn,
			UpdatedBy: envOverride.UpdatedBy,
		}
	} else {
		historyModel.Template = envOverride.Chart.GlobalOverride
		historyModel.AuditLog = sql.AuditLog{
			CreatedOn: envOverride.Chart.CreatedOn,
			CreatedBy: envOverride.Chart.CreatedBy,
			UpdatedOn: envOverride.Chart.UpdatedOn,
			UpdatedBy: envOverride.Chart.UpdatedBy,
		}
	}
	//creating new entry
	_, err := impl.chartHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("err in creating history entry for charts", "err", err, "history", historyModel)
		return err
	}
	return nil
}

func (impl DeploymentTriggerHistoryServiceImpl) CreateConfigMapHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, deployedOn time.Time, deployedBy int32) error {
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
	//configMapData, err := impl.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, history.CONFIGMAP_TYPE)
	//if err != nil {
	//	impl.logger.Errorw("err in merging app and env level configs", "err", err)
	//	return err
	//}
	historyModel := &history.ConfigmapAndSecretHistory{
		PipelineId: pipeline.Id,
		DataType:   history.CONFIGMAP_TYPE,
		Deployed:   true,
		DeployedBy: deployedBy,
		DeployedOn: deployedOn,
		//Data:       configMapData,
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
	//secretData, err := impl.configMapHistoryService.MergeAppLevelAndEnvLevelConfigs(appLevelConfig, envLevelConfig, history.SECRET_TYPE)
	//if err != nil {
	//	impl.logger.Errorw("err in merging app and env level configs", "err", err)
	//	return err
	//}
	//using old model, updating secret data
	historyModel.DataType = history.SECRET_TYPE
	historyModel.Id = 0
	//historyModel.Data = secretData
	_, err = impl.configMapHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("error in creating new entry for secret history", "historyModel", historyModel)
		return err
	}
	return nil
}

func (impl DeploymentTriggerHistoryServiceImpl) CreateStrategyHistoryForDeploymentTrigger(pipelineStrategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32) error {
	//creating new entry
	historyModel := &history.PipelineStrategyHistory{
		PipelineId: pipelineStrategy.PipelineId,
		Strategy:   pipelineStrategy.Strategy,
		Config:     pipelineStrategy.Config,
		Default:    pipelineStrategy.Default,
		Deployed:   true,
		DeployedBy: deployedBy,
		DeployedOn: deployedOn,
		AuditLog: sql.AuditLog{
			CreatedOn: pipelineStrategy.CreatedOn,
			CreatedBy: pipelineStrategy.CreatedBy,
			UpdatedOn: pipelineStrategy.UpdatedOn,
			UpdatedBy: pipelineStrategy.UpdatedBy,
		},
	}
	_, err := impl.pipelineStrategyHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("err in creating history entry for ci script", "err", err)
		return err
	}
	return err
}
