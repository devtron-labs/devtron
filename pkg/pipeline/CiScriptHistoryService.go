package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CiScriptHistoryService interface {
	CreateCiScriptHistory(ciPipelineScript *pipelineConfig.CiPipelineScript, tx *pg.Tx, built bool, builtBy int32, builtOn time.Time) (historyModel *history.CiScriptHistory, err error)
}

type CiScriptHistoryServiceImpl struct {
	logger                    *zap.SugaredLogger
	ciScriptHistoryRepository history.CiScriptHistoryRepository
}

func NewCiScriptHistoryServiceImpl(logger *zap.SugaredLogger, ciScriptHistoryRepository history.CiScriptHistoryRepository) *CiScriptHistoryServiceImpl {
	return &CiScriptHistoryServiceImpl{
		logger:                    logger,
		ciScriptHistoryRepository: ciScriptHistoryRepository,
	}
}

func (impl CiScriptHistoryServiceImpl) CreateCiScriptHistory(ciPipelineScript *pipelineConfig.CiPipelineScript, tx *pg.Tx, built bool, builtBy int32, builtOn time.Time) (historyModel *history.CiScriptHistory, err error) {
	//creating new entry
	historyModel = &history.CiScriptHistory{
		CiPipelineScriptsId: ciPipelineScript.Id,
		Script:              ciPipelineScript.Script,
		Stage:               ciPipelineScript.Stage,
		Built:               built,
		BuiltBy:             builtBy,
		BuiltOn:             builtOn,
		AuditLog: sql.AuditLog{
			CreatedOn: ciPipelineScript.CreatedOn,
			CreatedBy: ciPipelineScript.CreatedBy,
			UpdatedOn: ciPipelineScript.UpdatedOn,
			UpdatedBy: ciPipelineScript.UpdatedBy,
		},
	}
	if tx != nil {
		_, err = impl.ciScriptHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.ciScriptHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for ci script", "err", err)
		return nil, err
	}
	return historyModel, err
}
