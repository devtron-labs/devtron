package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CdTimelineStatus string

const (
	TIMELINE_STATUS_GIT_COMMIT            CdTimelineStatus = "GIT COMMIT"
	TIMELINE_STATUS_KUBECTL_APPLY_STARTED CdTimelineStatus = "KUBECTL APPLY STARTED"
	TIMELINE_STATUS_KUBECTL_APPLY_SYNCED  CdTimelineStatus = "KUBECTL APPLY SYNCED"
	TIMELINE_STATUS_APP_HEALTHY           CdTimelineStatus = "HEALTHY"
	TIMELINE_STATUS_APP_DEGRADED          CdTimelineStatus = "DEGRADED"
)

type CdPipelineStatusTimelineRepository interface {
	SaveTimeline(timeline *CdPipelineStatusTimeline) error
	FetchTimelinesByPipelineId(pipelineId int) ([]*CdPipelineStatusTimeline, error)
	FetchTimelinesByWfrId(wfrId int) ([]*CdPipelineStatusTimeline, error)
	FetchTimelineOfLatestWfByCdWorkflowIdAndStatus(pipelineId int, status CdTimelineStatus) (*CdPipelineStatusTimeline, error)
	CheckTimelineExistsOfLatestWfByCdWorkflowIdAndStatus(cdWorkflowId int, status CdTimelineStatus) (bool, error)
}

type CdPipelineStatusTimelineRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCdPipelineStatusTimelineRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *CdPipelineStatusTimelineRepositoryImpl {
	return &CdPipelineStatusTimelineRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type CdPipelineStatusTimeline struct {
	tableName          struct{}         `sql:"cd_pipeline_status_timeline" pg:",discard_unknown_columns"`
	Id                 int              `sql:"id,pk"`
	CdWorkflowRunnerId int              `sql:"cd_workflow_runner_id"`
	Status             CdTimelineStatus `sql:"status"`
	StatusDetail       string           `sql:"status_detail"`
	StatusTime         time.Time        `sql:"status_time"`
	sql.AuditLog
}

func (impl *CdPipelineStatusTimelineRepositoryImpl) SaveTimeline(timeline *CdPipelineStatusTimeline) error {
	err := impl.dbConnection.Insert(timeline)
	if err != nil {
		impl.logger.Errorw("error in saving timeline of cd pipeline status", "err", err, "timeline", timeline)
		return err
	}
	return nil
}

func (impl *CdPipelineStatusTimelineRepositoryImpl) FetchTimelinesByPipelineId(pipelineId int) ([]*CdPipelineStatusTimeline, error) {
	var timelines []*CdPipelineStatusTimeline
	err := impl.dbConnection.Model(&timelines).
		Join("INNER JOIN cd_workflow_runner wfr ON wfr.id = cd_pipeline_status_timeline.cd_workflow_runner_id").
		Join("INNER JOIN cd_workflow cw ON cw.id=wfr.cd_workflow_id").
		Where("cd_workflow.pipelineId = ?", pipelineId).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timelines by pipelineId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return timelines, nil
}

func (impl *CdPipelineStatusTimelineRepositoryImpl) FetchTimelinesByWfrId(wfrId int) ([]*CdPipelineStatusTimeline, error) {
	var timelines []*CdPipelineStatusTimeline
	err := impl.dbConnection.Model(&timelines).
		Where("cd_workflow_runner_id = ?", wfrId).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timelines by wfrId", "err", err, "wfrId", wfrId)
		return nil, err
	}
	return timelines, nil
}

func (impl *CdPipelineStatusTimelineRepositoryImpl) FetchTimelineOfLatestWfByCdWorkflowIdAndStatus(cdWorkflowId int, status CdTimelineStatus) (*CdPipelineStatusTimeline, error) {
	timeline := &CdPipelineStatusTimeline{}
	err := impl.dbConnection.Model(timeline).
		Join("INNER JOIN cd_workflow_runner wfr ON wfr.id = cd_pipeline_status_timeline.cd_workflow_runner_id").
		Join("INNER JOIN cd_workflow cw ON cw.id=wfr.cd_workflow_id").
		Where("cd_workflow.id = ?", cdWorkflowId).
		Where("cd_pipeline_status_timeline.status = ?", status).
		Order("cw.id DESC").Limit(1).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest wf by pipelineId and status", "err", err, "cdWorkflowId", cdWorkflowId)
		return nil, err
	}
	return timeline, nil
}

func (impl *CdPipelineStatusTimelineRepositoryImpl) CheckTimelineExistsOfLatestWfByCdWorkflowIdAndStatus(cdWorkflowId int, status CdTimelineStatus) (bool, error) {
	timeline := &CdPipelineStatusTimeline{}
	exists, err := impl.dbConnection.Model(timeline).
		Join("INNER JOIN cd_workflow_runner wfr ON wfr.id = cd_pipeline_status_timeline.cd_workflow_runner_id").
		Join("INNER JOIN cd_workflow cw ON cw.id=wfr.cd_workflow_id").
		Where("cd_workflow.id = ?", cdWorkflowId).
		Where("cd_pipeline_status_timeline.status = ?", status).
		Order("cw.id DESC").Limit(1).Exists()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest wf by pipelineId and status", "err", err, "cdWorkflowId", cdWorkflowId)
		return false, err
	}
	return exists, nil
}
