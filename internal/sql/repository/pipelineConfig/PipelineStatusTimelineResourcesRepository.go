package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ResourceTimelineStage string

const (
	TIMELINE_RESOURCE_STAGE_KUBECTL_APPLY ResourceTimelineStage = "KUBECTL_APPLY"
)

type PipelineStatusTimelineResourcesRepository interface {
	SaveTimelineResources(timelineResources []*PipelineStatusTimelineResources) error
	SaveTimelineResourcesWithTxn(timelineResources []*PipelineStatusTimelineResources, tx *pg.Tx) error
	UpdateTimelineResources(timelineResources []*PipelineStatusTimelineResources) error
	UpdateTimelineResourcesWithTxn(timelineResources []*PipelineStatusTimelineResources, tx *pg.Tx) error
	GetByCdWfrIdAndTimelineStage(cdWfrId int) ([]*PipelineStatusTimelineResources, error)
	GetByInstalledAppVersionHistoryIdAndTimelineStage(installedAppVersionHistoryId int) ([]*PipelineStatusTimelineResources, error)
	GetByCdWfrIds(cdWfrIds []int) ([]*PipelineStatusTimelineResources, error)
}

type PipelineStatusTimelineResourcesRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineStatusTimelineResourcesRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *PipelineStatusTimelineResourcesRepositoryImpl {
	return &PipelineStatusTimelineResourcesRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type PipelineStatusTimelineResources struct {
	tableName                    struct{}              `sql:"pipeline_status_timeline_resources" pg:",discard_unknown_columns"`
	Id                           int                   `sql:"id,pk"`
	InstalledAppVersionHistoryId int                   `sql:"installed_app_version_history_id,type:integer"`
	CdWorkflowRunnerId           int                   `sql:"cd_workflow_runner_id,type:integer"`
	ResourceName                 string                `sql:"resource_name"`
	ResourceKind                 string                `sql:"resource_kind"`
	ResourceGroup                string                `sql:"resource_group"`
	ResourcePhase                string                `sql:"resource_phase"`
	ResourceStatus               string                `sql:"resource_status"`
	StatusMessage                string                `sql:"status_message"`
	TimelineStage                ResourceTimelineStage `sql:"timeline_stage"`
	sql.AuditLog
}

func (impl *PipelineStatusTimelineResourcesRepositoryImpl) SaveTimelineResources(timelineResources []*PipelineStatusTimelineResources) error {
	err := impl.dbConnection.Insert(&timelineResources)
	if err != nil {
		impl.logger.Errorw("error in saving timelineResources resources of cd pipeline status", "err", err, "timelineResources", timelineResources)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineResourcesRepositoryImpl) SaveTimelineResourcesWithTxn(timelineResources []*PipelineStatusTimelineResources, tx *pg.Tx) error {
	err := tx.Insert(&timelineResources)
	if err != nil {
		impl.logger.Errorw("error in saving timeline resources of cd pipeline status", "err", err, "timelineResources", timelineResources)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineResourcesRepositoryImpl) UpdateTimelineResources(timelineResources []*PipelineStatusTimelineResources) error {
	_, err := impl.dbConnection.Model(&timelineResources).Update()
	if err != nil {
		impl.logger.Errorw("error in updating timeline resources of pipeline status", "err", err, "timelineResources", timelineResources)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineResourcesRepositoryImpl) UpdateTimelineResourcesWithTxn(timelineResources []*PipelineStatusTimelineResources, tx *pg.Tx) error {
	_, err := tx.Model(&timelineResources).Update()
	if err != nil {
		impl.logger.Errorw("error in saving timeline resources of cd pipeline status", "err", err, "timelineResources", timelineResources)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineResourcesRepositoryImpl) GetByCdWfrIdAndTimelineStage(cdWfrId int) ([]*PipelineStatusTimelineResources, error) {
	var timelineResources []*PipelineStatusTimelineResources
	err := impl.dbConnection.Model(&timelineResources).
		Where("cd_workflow_runner_id = ?", cdWfrId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline resources by cdWfrId and timeline stage", "err", err, "cdWfrId", cdWfrId, "timelineStage", timelineResources)
		return nil, err
	}
	return timelineResources, nil
}

func (impl *PipelineStatusTimelineResourcesRepositoryImpl) GetByInstalledAppVersionHistoryIdAndTimelineStage(installedAppVersionHistoryId int) ([]*PipelineStatusTimelineResources, error) {
	var timelineResources []*PipelineStatusTimelineResources
	err := impl.dbConnection.Model(&timelineResources).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline resources by installedAppVersionHistoryId and timeline stage", "err", err, "cdWfrId", installedAppVersionHistoryId, "timelineStage", timelineResources)
		return nil, err
	}
	return timelineResources, nil
}

func (impl *PipelineStatusTimelineResourcesRepositoryImpl) GetByCdWfrIds(cdWfrIds []int) ([]*PipelineStatusTimelineResources, error) {
	var timelineResources []*PipelineStatusTimelineResources
	err := impl.dbConnection.Model(&timelineResources).
		Where("cd_workflow_runner_id in (?)", pg.In(cdWfrIds)).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline resources by cdWfrId and timeline stage", "err", err, "cdWfrIds", cdWfrIds, "timelineStage", timelineResources)
		return nil, err
	}
	return timelineResources, nil
}
