/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusTimelineRepository interface {
	SaveTimelines(timelines []*PipelineStatusTimeline) error
	SaveTimelinesWithTxn(timelines []*PipelineStatusTimeline, tx *pg.Tx) error
	UpdateTimelines(timelines []*PipelineStatusTimeline) error
	UpdateTimelinesWithTxn(timelines []*PipelineStatusTimeline, tx *pg.Tx) error
	FetchTimelinesByPipelineId(pipelineId int) ([]*PipelineStatusTimeline, error)
	// FetchTimelinesByWfrId - Gets the exposed timelines for Helm Applications,
	// ignoring internalTimelineStatusList in sql query as it is not handled at FE
	FetchTimelinesByWfrId(wfrId int) ([]*PipelineStatusTimeline, error)
	FetchTimelineByWfrIdAndStatus(wfrId int, status timelineStatus.TimelineStatus) (*PipelineStatusTimeline, error)
	FetchTimelineByInstalledAppVersionHistoryIdAndStatus(installedAppVersionHistoryId int, status timelineStatus.TimelineStatus) (*PipelineStatusTimeline, error)
	FetchTimelineByWfrIdAndStatuses(wfrId int, statuses []timelineStatus.TimelineStatus) ([]*PipelineStatusTimeline, error)
	FetchTimelineByInstalledAppVersionHistoryIdAndPipelineStatuses(installedAppVersionHistoryId int, statuses []timelineStatus.TimelineStatus) ([]*PipelineStatusTimeline, error)
	GetLastStatusPublishedTimeForWfrId(wfrId int) (time.Time, error)
	FetchTimelinesForWfrIdExcludingStatuses(wfrId int, statuses ...timelineStatus.TimelineStatus) ([]*PipelineStatusTimeline, error)
	CheckIfTerminalStatusTimelinePresentByWfrId(wfrId int) (bool, error)
	CheckIfTimelineStatusPresentByWfrId(wfrId int, status timelineStatus.TimelineStatus) (bool, error)
	CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) (bool, error)
	CheckIfTimelineStatusPresentByInstalledAppVersionHistoryId(installedAppVersionHistoryId int, status timelineStatus.TimelineStatus) (bool, error)
	DeleteByCdWfrIdAndTimelineStatuses(cdWfrId int, status []timelineStatus.TimelineStatus) error
	DeleteByCdWfrIdAndTimelineStatusesWithTxn(cdWfrId int, status []timelineStatus.TimelineStatus, tx *pg.Tx) error
	// FetchTimelinesByInstalledAppVersionHistoryId - Gets the exposed timelines for Helm Applications,
	// ignoring internalTimelineStatusList in sql query as it is not handled at FE
	FetchTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) ([]*PipelineStatusTimeline, error)
	FetchLatestTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) (*PipelineStatusTimeline, error)
	GetConnection() *pg.DB
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
	tableName                    struct{}                      `sql:"pipeline_status_timeline" pg:",discard_unknown_columns"`
	Id                           int                           `sql:"id,pk"`
	InstalledAppVersionHistoryId int                           `sql:"installed_app_version_history_id,type:integer"`
	CdWorkflowRunnerId           int                           `sql:"cd_workflow_runner_id,type:integer"`
	Status                       timelineStatus.TimelineStatus `sql:"status"`
	StatusDetail                 string                        `sql:"status_detail"`
	StatusTime                   time.Time                     `sql:"status_time"`
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
		Where("status NOT IN (?)", pg.In(timelineStatus.InternalTimelineStatusList)).
		Order("status_time ASC").Select()
	if err != nil {
		impl.logger.Errorw("error in getting timelines by wfrId", "err", err, "wfrId", wfrId)
		return nil, err
	}
	return timelines, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByWfrIdAndStatus(wfrId int, status timelineStatus.TimelineStatus) (*PipelineStatusTimeline, error) {
	timeline := &PipelineStatusTimeline{}
	err := impl.dbConnection.Model(timeline).
		Where("cd_workflow_runner_id = ?", wfrId).
		Where("status = ?", status).
		Limit(1).Select()
	return timeline, err
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByInstalledAppVersionHistoryIdAndStatus(installedAppVersionHistoryId int, status timelineStatus.TimelineStatus) (*PipelineStatusTimeline, error) {
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

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByWfrIdAndStatuses(wfrId int, statuses []timelineStatus.TimelineStatus) ([]*PipelineStatusTimeline, error) {
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

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelineByInstalledAppVersionHistoryIdAndPipelineStatuses(installedAppVersionHistoryId int, statuses []timelineStatus.TimelineStatus) ([]*PipelineStatusTimeline, error) {
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

func (impl *PipelineStatusTimelineRepositoryImpl) GetLastStatusPublishedTimeForWfrId(wfrId int) (time.Time, error) {
	timeline := &PipelineStatusTimeline{}
	err := impl.dbConnection.Model(timeline).
		Column("status_time").
		Where("cd_workflow_runner_id = ?", wfrId).
		Where("status NOT IN (?)", pg.In(timelineStatus.InternalTimelineStatusList)).
		Order("status_time DESC").
		Limit(1).Select()
	return timeline.StatusTime, err
}

func (impl *PipelineStatusTimelineRepositoryImpl) FetchTimelinesForWfrIdExcludingStatuses(wfrId int, statuses ...timelineStatus.TimelineStatus) ([]*PipelineStatusTimeline, error) {
	var timelines []*PipelineStatusTimeline
	query := impl.dbConnection.Model(&timelines).
		Where("cd_workflow_runner_id = ?", wfrId)
	if len(statuses) > 0 {
		query = query.Where("status NOT in (?)", pg.In(statuses))
	}
	err := query.Order("status_time DESC").
		Limit(1).
		Select()
	return timelines, err
}

func (impl *PipelineStatusTimelineRepositoryImpl) CheckIfTerminalStatusTimelinePresentByWfrId(wfrId int) (bool, error) {
	timeline := &PipelineStatusTimeline{}
	exists, err := impl.dbConnection.Model(timeline).
		Where("cd_workflow_runner_id = ?", wfrId).
		Where("status in (?)", pg.In(timelineStatus.TerminalTimelineStatusList)).Exists()
	if err != nil {
		impl.logger.Errorw("error in checking if terminal timeline of latest wf by pipelineId and status", "err", err, "wfrId", wfrId)
		return false, err
	}
	return exists, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) CheckIfTimelineStatusPresentByWfrId(wfrId int, status timelineStatus.TimelineStatus) (bool, error) {
	timeline := &PipelineStatusTimeline{}
	exists, err := impl.dbConnection.Model(timeline).
		Where("cd_workflow_runner_id = ?", wfrId).
		Where("status = ?", status).
		Exists()
	if err != nil {
		impl.logger.Errorw("error in checking if timeline status exists for wfrId", "err", err, "wfrId", wfrId, "status", status)
		return false, err
	}
	return exists, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(installedAppVersionHistoryId int) (bool, error) {
	timeline := &PipelineStatusTimeline{}
	exists, err := impl.dbConnection.Model(timeline).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Where("status in (?)", pg.In(timelineStatus.TerminalTimelineStatusList)).Exists()
	if err != nil {
		impl.logger.Errorw("error in checking if terminal timeline of latest installed app by installedAppVersionHistoryId and status", "err", err, "wfrId", installedAppVersionHistoryId)
		return false, err
	}
	return exists, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) CheckIfTimelineStatusPresentByInstalledAppVersionHistoryId(installedAppVersionHistoryId int, status timelineStatus.TimelineStatus) (bool, error) {
	timeline := &PipelineStatusTimeline{}
	exists, err := impl.dbConnection.Model(timeline).
		Where("installed_app_version_history_id = ?", installedAppVersionHistoryId).
		Where("status = ?", status).
		Exists()
	if err != nil {
		impl.logger.Errorw("error in checking if timeline status exists for wfrId", "err", err, "installedAppVersionHistoryId", installedAppVersionHistoryId, "status", status)
		return false, err
	}
	return exists, nil
}

func (impl *PipelineStatusTimelineRepositoryImpl) DeleteByCdWfrIdAndTimelineStatuses(cdWfrId int, status []timelineStatus.TimelineStatus) error {
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

func (impl *PipelineStatusTimelineRepositoryImpl) DeleteByCdWfrIdAndTimelineStatusesWithTxn(cdWfrId int, status []timelineStatus.TimelineStatus, tx *pg.Tx) error {
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
		Where("status NOT IN (?)", pg.In(timelineStatus.InternalTimelineStatusList)).
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

func (impl *PipelineStatusTimelineRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}
