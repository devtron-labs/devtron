package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type TimelineStatus = string

const (
	TIMELINE_STATUS_DEPLOYMENT_INITIATED                TimelineStatus = "DEPLOYMENT_INITIATED"
	TIMELINE_STATUS_GIT_COMMIT                          TimelineStatus = "GIT_COMMIT"
	TIMELINE_STATUS_GIT_COMMIT_FAILED                   TimelineStatus = "GIT_COMMIT_FAILED"
	TIMELINE_STATUS_KUBECTL_APPLY_STARTED               TimelineStatus = "KUBECTL_APPLY_STARTED"
	TIMELINE_STATUS_KUBECTL_APPLY_SYNCED                TimelineStatus = "KUBECTL_APPLY_SYNCED"
	TIMELINE_STATUS_APP_HEALTHY                         TimelineStatus = "HEALTHY"
	TIMELINE_STATUS_DEPLOYMENT_FAILED                   TimelineStatus = "FAILED"
	TIMELINE_STATUS_FETCH_TIMED_OUT                     TimelineStatus = "TIMED_OUT"
	TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS              TimelineStatus = "UNABLE_TO_FETCH_STATUS"
	TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED               TimelineStatus = "DEPLOYMENT_SUPERSEDED"
	TIMELINE_STATUS_MANIFEST_GENERATED                  TimelineStatus = "MANIFEST_GENERATED"
	TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO        TimelineStatus = "HELM_MANIFEST_PUSHED_TO_HELM_REPO"
	TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO_FAILED TimelineStatus = "HELM_MANIFEST_PUSHED_TO_HELM_REPO_FAILED"
)

const (
	TIMELINE_DESCRIPTION_DEPLOYMENT_INITIATED       string = "Deployment initiated successfully."
	TIMELINE_DESCRIPTION_VULNERABLE_IMAGE           string = "Deployment failed: Vulnerability policy violated."
	TIMELINE_DESCRIPTION_MANIFEST_GENERATED         string = "HELM_PACKAGE_GENERATED"
	TIMELINE_DESCRIPTION_MANIFEST_GENERATION_FAILED string = "HELM_PACKAGE_GENERATION_FAILED"
)

type PipelineStatusTimelineRepository interface {
	SaveTimelines(timelines []*PipelineStatusTimeline) error
	SaveTimelinesWithTxn(timelines []*PipelineStatusTimeline, tx *pg.Tx) error
	UpdateTimelines(timelines []*PipelineStatusTimeline) error
	UpdateTimelinesWithTxn(timelines []*PipelineStatusTimeline, tx *pg.Tx) error
	FetchTimelinesByPipelineId(pipelineId int) ([]*PipelineStatusTimeline, error)
	FetchTimelinesByWfrId(wfrId int) ([]*PipelineStatusTimeline, error)
	FetchTimelineByWfrIdAndStatus(wfrId int, status TimelineStatus) (*PipelineStatusTimeline, error)
	FetchTimelineByInstalledAppVersionHistoryIdAndStatus(installedAppVersionHistoryId int, status TimelineStatus) (*PipelineStatusTimeline, error)
	FetchTimelineByWfrIdAndStatuses(wfrId int, statuses []TimelineStatus) ([]*PipelineStatusTimeline, error)
	FetchTimelineByInstalledAppVersionHistoryIdAndPipelineStatuses(installedAppVersionHistoryId int, statuses []TimelineStatus) ([]*PipelineStatusTimeline, error)
	FetchLatestTimelineByWfrId(wfrId int) (*PipelineStatusTimeline, error)
	CheckIfTerminalStatusTimelinePresentByWfrId(wfrId int) (bool, error)
	CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) (bool, error)
	FetchLatestTimelineByAppIdAndEnvId(appId, envId int) (*PipelineStatusTimeline, error)
	DeleteByCdWfrIdAndTimelineStatuses(cdWfrId int, status []TimelineStatus) error
	DeleteByCdWfrIdAndTimelineStatusesWithTxn(cdWfrId int, status []TimelineStatus, tx *pg.Tx) error
	FetchTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) ([]*PipelineStatusTimeline, error)
	FetchLatestTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) (*PipelineStatusTimeline, error)
}

type PipelineStatusTimelineRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineStatusTimelineRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *PipelineStatusTimelineRepositoryImpl {
	return &PipelineStatusTimelineRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type PipelineStatusTimeline struct {
	tableName                    struct{}       `sql:"pipeline_status_timeline" pg:",discard_unknown_columns"`
	Id                           int            `sql:"id,pk"`
	InstalledAppVersionHistoryId int            `sql:"installed_app_version_history_id,type:integer"`
	CdWorkflowRunnerId           int            `sql:"cd_workflow_runner_id,type:integer"`
	Status                       TimelineStatus `sql:"status"`
	StatusDetail                 string         `sql:"status_detail"`
	StatusTime                   time.Time      `sql:"status_time"`
	sql.AuditLog
}

func (impl *PipelineStatusTimelineRepositoryImpl) SaveTimelines(timelines []*PipelineStatusTimeline) error {
	err := impl.dbConnection.Insert(&timelines)
	if err != nil {
		impl.logger.Errorw("error in saving timeline of cd pipeline status", "err", err, "timeline", timelines)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) SaveTimelinesWithTxn(timelines []*PipelineStatusTimeline, tx *pg.Tx) error {
	err := tx.Insert(&timelines)
	if err != nil {
		impl.logger.Errorw("error in saving timelines of cd pipeline status", "err", err, "timelines", timelines)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) UpdateTimelines(timelines []*PipelineStatusTimeline) error {
	_, err := impl.dbConnection.Model(&timelines).Update()
	if err != nil {
		impl.logger.Errorw("error in updating timeline of cd pipeline status", "err", err, "timeline", timelines)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) UpdateTimelinesWithTxn(timelines []*PipelineStatusTimeline, tx *pg.Tx) error {
	_, err := tx.Model(&timelines).Update()
	if err != nil {
		impl.logger.Errorw("error in updating timelines of cd pipeline status", "err", err, "timelines", timelines)
		return err
	}
	return nil
}
func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelinesByPipelineId(pipelineId int) ([]*PipelineStatusTimeline, error) {
	var timelines []*PipelineStatusTimeline
	err := impl.dbConnection.Model(&timelines).
		Join("INNER JOIN cd_workflow_runner wfr ON wfr.id = pipeline_status_timeline.cd_workflow_runner_id").
		Join("INNER JOIN cd_workflow cw ON cw.id=wfr.cd_workflow_id").
		Where("cw.pipelineId = ?", pipelineId).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timelines by pipelineId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return timelines, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelinesByWfrId(wfrId int) ([]*PipelineStatusTimeline, error) {
	var timelines []*PipelineStatusTimeline
	err := impl.dbConnection.Model(&timelines).
		Where("cd_workflow_runner_id = ?", wfrId).
		Order("status_time ASC").Select()
	if err != nil {
		impl.logger.Errorw("error in getting timelines by wfrId", "err", err, "wfrId", wfrId)
		return nil, err
	}
	return timelines, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByWfrIdAndStatus(wfrId int, status TimelineStatus) (*PipelineStatusTimeline, error) {
	timeline := &PipelineStatusTimeline{}
	err := impl.dbConnection.Model(timeline).
		Where("cd_workflow_runner_id = ?", wfrId).
		Where("status = ?", status).
		Limit(1).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest wf by wfrId and status", "err", err, "wfrId", wfrId, "status", status)
		return nil, err
	}
	return timeline, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByInstalledAppVersionHistoryIdAndStatus(installedAppVersionHistoryId int, status TimelineStatus) (*PipelineStatusTimeline, error) {
	timeline := &PipelineStatusTimeline{}
	err := impl.dbConnection.Model(timeline).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Where("status = ?", status).
		Limit(1).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest installed app version history by installedAppVersionHistoryId and status", "err", err, "installedAppVersionHistoryId", installedAppVersionHistoryId, "status", status)
		return nil, err
	}
	return timeline, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByWfrIdAndStatuses(wfrId int, statuses []TimelineStatus) ([]*PipelineStatusTimeline, error) {
	var timelines []*PipelineStatusTimeline
	err := impl.dbConnection.Model(&timelines).
		Where("cd_workflow_runner_id = ?", wfrId).
		Where("status in (?)", pg.In(statuses)).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest wf by wfrId and statuses", "err", err, "wfrId", wfrId, "statuses", statuses)
		return nil, err
	}
	return timelines, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByInstalledAppVersionHistoryIdAndPipelineStatuses(installedAppVersionHistoryId int, statuses []TimelineStatus) ([]*PipelineStatusTimeline, error) {
	var timelines []*PipelineStatusTimeline
	err := impl.dbConnection.Model(&timelines).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Where("status in (?)", pg.In(statuses)).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest wf by installedAppVersionHistoryId and statuses", "err", err, "installedAppVersionHistoryId", installedAppVersionHistoryId, "statuses", statuses)
		return nil, err
	}
	return timelines, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchLatestTimelineByWfrId(wfrId int) (*PipelineStatusTimeline, error) {
	timeline := &PipelineStatusTimeline{}
	err := impl.dbConnection.Model(timeline).
		Where("cd_workflow_runner_id = ?", wfrId).
		Order("status_time DESC").
		Limit(1).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest wf by wfrId", "err", err, "wfrId", wfrId)
		return nil, err
	}
	return timeline, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) CheckIfTerminalStatusTimelinePresentByWfrId(wfrId int) (bool, error) {
	terminalStatus := []string{TIMELINE_STATUS_APP_HEALTHY, TIMELINE_STATUS_DEPLOYMENT_FAILED, TIMELINE_STATUS_GIT_COMMIT_FAILED, TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED, TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO_FAILED}
	timeline := &PipelineStatusTimeline{}
	exists, err := impl.dbConnection.Model(timeline).
		Where("cd_workflow_runner_id = ?", wfrId).
		Where("status in (?)", pg.In(terminalStatus)).Exists()
	if err != nil {
		impl.logger.Errorw("error in checking if terminal timeline of latest wf by pipelineId and status", "err", err, "wfrId", wfrId)
		return false, err
	}
	return exists, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) (bool, error) {
	terminalStatus := []string{string(TIMELINE_STATUS_APP_HEALTHY), string(TIMELINE_STATUS_DEPLOYMENT_FAILED), string(TIMELINE_STATUS_GIT_COMMIT_FAILED), string(TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED)}
	timeline := &PipelineStatusTimeline{}
	exists, err := impl.dbConnection.Model(timeline).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Where("status in (?)", pg.In(terminalStatus)).Exists()
	if err != nil {
		impl.logger.Errorw("error in checking if terminal timeline of latest installed app by installedAppVersionHistoryId and status", "err", err, "wfrId", installedAppVersionHistoryId)
		return false, err
	}
	return exists, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchLatestTimelineByAppIdAndEnvId(appId, envId int) (*PipelineStatusTimeline, error) {
	var timeline PipelineStatusTimeline
	err := impl.dbConnection.Model(&timeline).
		Column("pipeline_status_timeline.*").
		Join("INNER JOIN cd_workflow_runner wfr ON wfr.id = pipeline_status_timeline.cd_workflow_runner_id").
		Join("INNER JOIN cd_workflow cw ON cw.id=wfr.cd_workflow_id").
		Join("INNER JOIN pipeline p ON p.id=cw.pipeline_id").
		Where("p.app_id = ?", appId).
		Where("p.environment_id = ?", envId).
		Where("p.deleted = false").
		Order("pipeline_status_timeline.status_time DESC").
		Limit(1).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting timelines by pipelineId", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}
	return &timeline, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) DeleteByCdWfrIdAndTimelineStatuses(cdWfrId int, status []TimelineStatus) error {
	var timeline PipelineStatusTimeline
	_, err := impl.dbConnection.Model(&timeline).
		Where("cd_workflow_runner_id = ?", cdWfrId).
		Where("status in (?)", pg.In(status)).Delete()
	if err != nil {
		impl.logger.Errorw("error in deleting pipeline status timeline by cdWfrId and status", "err", err, "cdWfrId", cdWfrId, "status", status)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) DeleteByCdWfrIdAndTimelineStatusesWithTxn(cdWfrId int, status []TimelineStatus, tx *pg.Tx) error {
	var timeline PipelineStatusTimeline
	_, err := tx.Model(&timeline).
		Where("cd_workflow_runner_id = ?", cdWfrId).
		Where("status in (?)", pg.In(status)).Delete()
	if err != nil {
		impl.logger.Errorw("error in deleting pipeline status timeline by cdWfrId and status", "err", err, "cdWfrId", cdWfrId, "status", status)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) ([]*PipelineStatusTimeline, error) {
	var timelines []*PipelineStatusTimeline
	err := impl.dbConnection.Model(&timelines).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Order("status_time ASC").Select()
	if err != nil {
		impl.logger.Errorw("error in getting timelines by installAppVersionHistoryId", "err", err, "wfrId", installedAppVersionHistoryId)
		return nil, err
	}
	return timelines, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchLatestTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) (*PipelineStatusTimeline, error) {
	timeline := &PipelineStatusTimeline{}
	err := impl.dbConnection.Model(timeline).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Order("status_time DESC").
		Limit(1).Select()
	if err != nil {
		impl.logger.Errorw("error in getting timeline of latest installed_app_version_history by installed_app_version_history_id", "err", err, "installed_app_version_history_id", installedAppVersionHistoryId)
		return nil, err
	}
	return timeline, nil
}
