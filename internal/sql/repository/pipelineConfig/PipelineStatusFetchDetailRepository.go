package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusFetchDetail struct {
	tableName                    struct{}  `sql:"pipeline_status_timeline_fetch_detail" pg:",discard_unknown_columns"`
	Id                           int       `sql:"id,pk"`
	InstalledAppVersionHistoryId int       `sql:"installed_app_version_history_id"`
	CdWorkflowRunnerId           int       `sql:"cd_workflow_runner_id"`
	LastFetchedAt                time.Time `sql:"last_fetched_at"`
	FetchCount                   int       `sql:"fetch_count"`
	sql.AuditLog
}

type PipelineStatusFetchDetailRepository interface {
	Save(model *PipelineStatusFetchDetail) error
	Update(model *PipelineStatusFetchDetail) error
	GetByCdWfrId(cdWfrId int) (*PipelineStatusFetchDetail, error)
}

type PipelineStatusFetchDetailRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineStatusFetchDetailRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *PipelineStatusFetchDetailRepositoryImpl {
	return &PipelineStatusFetchDetailRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *PipelineStatusFetchDetailRepositoryImpl) Save(model *PipelineStatusFetchDetail) error {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("error in saving cd pipeline status fetch detail", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *PipelineStatusFetchDetailRepositoryImpl) Update(model *PipelineStatusFetchDetail) error {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Errorw("error in updating cd pipeline status fetch detail", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *PipelineStatusFetchDetailRepositoryImpl) GetByCdWfrId(cdWfrId int) (*PipelineStatusFetchDetail, error) {
	var model PipelineStatusFetchDetail
	_, err := impl.dbConnection.Model(&model).Where("cd_workflow_runner_id = ?", cdWfrId).Exists()
	if err != nil {
		impl.logger.Errorw("error in getting cd pipeline status fetch detail by cdWfrId", "err", err, "cdWfrId", cdWfrId)
		return nil, err
	}
	return &model, nil
}
