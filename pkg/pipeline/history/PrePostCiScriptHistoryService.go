package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PrePostCiScriptHistoryService interface {
	CreatePrePostCiScriptHistory(ciPipelineScript *pipelineConfig.CiPipelineScript, tx *pg.Tx, built bool, builtBy int32, builtOn time.Time) (historyModel *repository.PrePostCiScriptHistory, err error)
}

type PrePostCiScriptHistoryServiceImpl struct {
	logger                           *zap.SugaredLogger
	prePostCiScriptHistoryRepository repository.PrePostCiScriptHistoryRepository
}

func NewPrePostCiScriptHistoryServiceImpl(logger *zap.SugaredLogger, prePostCiScriptHistoryRepository repository.PrePostCiScriptHistoryRepository) *PrePostCiScriptHistoryServiceImpl {
	return &PrePostCiScriptHistoryServiceImpl{
		logger:                           logger,
		prePostCiScriptHistoryRepository: prePostCiScriptHistoryRepository,
	}
}

func (impl PrePostCiScriptHistoryServiceImpl) CreatePrePostCiScriptHistory(ciPipelineScript *pipelineConfig.CiPipelineScript, tx *pg.Tx, built bool, builtBy int32, builtOn time.Time) (historyModel *repository.PrePostCiScriptHistory, err error) {
	//creating new entry
	historyModel = &repository.PrePostCiScriptHistory{
		CiPipelineScriptsId: ciPipelineScript.Id,
		Script:              ciPipelineScript.Script,
		Stage:               ciPipelineScript.Stage,
		Name:                ciPipelineScript.Name,
		OutputLocation:      ciPipelineScript.OutputLocation,
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
		_, err = impl.prePostCiScriptHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.prePostCiScriptHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for ci script", "err", err)
		return nil, err
	}
	return historyModel, err
}
