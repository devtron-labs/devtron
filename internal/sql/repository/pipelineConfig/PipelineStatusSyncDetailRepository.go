package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusSyncDetail struct {
	tableName                    struct{}  `sql:"pipeline_status_timeline_sync_detail" pg:",discard_unknown_columns"`
	Id                           int       `sql:"id,pk"`
	InstalledAppVersionHistoryId int       `sql:"installed_app_version_history_id"`
	CdWorkflowRunnerId           int       `sql:"cd_workflow_runner_id"`
	LastSyncedAt                 time.Time `sql:"last_synced_at"`
	SyncCount                    int       `sql:"sync_count"`
	sql.AuditLog
}

type PipelineStatusSyncDetailRepository interface {
	Save(model *PipelineStatusSyncDetail) error
	Update(model *PipelineStatusSyncDetail) error
	GetByCdWfrId(cdWfrId int) (*PipelineStatusSyncDetail, error)
	GetOfLatestCdWfrByArgoAppName(argoAppName string) (*PipelineStatusSyncDetail, error)
}

type PipelineStatusSyncDetailRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineStatusSyncDetailRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *PipelineStatusSyncDetailRepositoryImpl {
	return &PipelineStatusSyncDetailRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *PipelineStatusSyncDetailRepositoryImpl) Save(model *PipelineStatusSyncDetail) error {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("error in saving cd pipeline status sync detail", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *PipelineStatusSyncDetailRepositoryImpl) Update(model *PipelineStatusSyncDetail) error {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Errorw("error in updating cd pipeline status sync detail", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *PipelineStatusSyncDetailRepositoryImpl) GetByCdWfrId(cdWfrId int) (*PipelineStatusSyncDetail, error) {
	var model PipelineStatusSyncDetail
	err := impl.dbConnection.Model(&model).Where("cd_workflow_runner_id = ?", cdWfrId).Select()
	if err != nil {
		impl.logger.Errorw("error in getting cd pipeline status sync detail by cdWfrId", "err", err, "cdWfrId", cdWfrId)
		return nil, err
	}
	return &model, nil
}

func (impl *PipelineStatusSyncDetailRepositoryImpl) GetOfLatestCdWfrByArgoAppName(argoAppName string) (*PipelineStatusSyncDetail, error) {
	var model PipelineStatusSyncDetail
	query := `select * from pipeline_status_timeline_sync_detail 
              	where cd_workflow_runner_id = (select cwr.id from cd_workflow_runner cwr inner join cd_workflow cw on cw.is=cwr.cd_workflow_id 
                	inner join pipeline p on p.id=cw.pipeline_id where (p.app_id,p.environment_id) in
                	    (select app_id, env_id from deployment_status where app_name =? order by id desc limit ?) and p.active=? order by cwr.id desc limit ?);`
	_, err := impl.dbConnection.Query(&model, query, argoAppName, 1, true, 1)
	if err != nil {
		impl.logger.Errorw("error in getting cd pipeline status sync detail of latest cdWfr by argoAppName", "err", err, "argoAppName", argoAppName)
		return nil, err
	}
	return &model, nil
}
