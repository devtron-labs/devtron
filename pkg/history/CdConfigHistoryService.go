package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CdConfigHistoryService interface {
	CreateCdConfigHistory(pipeline *pipelineConfig.Pipeline, tx *pg.Tx, stage history.CdStageType, deployed bool, deployedBy int32, deployedOn time.Time) error
	GetHistoryForDeployedCdConfig(pipelineId int, stage history.CdStageType) ([]*history.CdConfigHistory, error)
}

type CdConfigHistoryServiceImpl struct {
	logger                    *zap.SugaredLogger
	cdConfigHistoryRepository history.CdConfigHistoryRepository
}

func NewCdConfigHistoryServiceImpl(logger *zap.SugaredLogger, cdConfigHistoryRepository history.CdConfigHistoryRepository) *CdConfigHistoryServiceImpl {
	return &CdConfigHistoryServiceImpl{
		logger:                    logger,
		cdConfigHistoryRepository: cdConfigHistoryRepository,
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

func (impl CdConfigHistoryServiceImpl) GetHistoryForDeployedCdConfig(pipelineId int, stage history.CdStageType) ([]*history.CdConfigHistory, error) {
	histories, err := impl.cdConfigHistoryRepository.GetHistoryForDeployedCdConfigByStage(pipelineId, stage)
	if err != nil {
		impl.logger.Errorw("error in getting cd config history", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return histories, nil
}