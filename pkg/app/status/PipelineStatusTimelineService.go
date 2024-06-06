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

package status

import (
	"fmt"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PipelineStatusTimelineService interface {
	SaveTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx, isAppStore bool) error
	FetchTimelines(appId, envId, wfrId int, showTimeline bool) (*PipelineTimelineDetailDto, error)
	FetchTimelinesForAppStore(installedAppId, envId, installedAppVersionHistoryId int, showTimeline bool) (*PipelineTimelineDetailDto, error)
	GetTimelineDbObjectByTimelineStatusAndTimelineDescription(cdWorkflowRunnerId int, installedAppVersionHistoryId int, timelineStatus pipelineConfig.TimelineStatus, timelineDescription string, userId int32, statusTime time.Time) *pipelineConfig.PipelineStatusTimeline
	SavePipelineStatusTimelineIfNotAlreadyPresent(pipelineId int, timelineStatus pipelineConfig.TimelineStatus, timeline *pipelineConfig.PipelineStatusTimeline, isAppStore bool) (latestTimeline *pipelineConfig.PipelineStatusTimeline, err error, isTimelineUpdated bool)
	GetArgoAppSyncStatus(cdWfrId int) bool
	GetArgoAppSyncStatusForAppStore(installedAppVersionHistoryId int) bool
	SaveTimelines(timeline []*pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error

	MarkPipelineStatusTimelineFailed(cdWfrId int, statusDetailMessage string) error
}

type PipelineStatusTimelineServiceImpl struct {
	logger                                 *zap.SugaredLogger
	pipelineStatusTimelineRepository       pipelineConfig.PipelineStatusTimelineRepository
	cdWorkflowRepository                   pipelineConfig.CdWorkflowRepository
	userService                            user.UserService
	pipelineStatusTimelineResourcesService PipelineStatusTimelineResourcesService
	pipelineStatusSyncDetailService        PipelineStatusSyncDetailService
	installedAppRepository                 repository.InstalledAppRepository
	installedAppVersionHistory             repository.InstalledAppVersionHistoryRepository
}

func NewPipelineStatusTimelineServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	userService user.UserService,
	pipelineStatusTimelineResourcesService PipelineStatusTimelineResourcesService,
	pipelineStatusSyncDetailService PipelineStatusSyncDetailService,
	installedAppRepository repository.InstalledAppRepository,
	installedAppVersionHistory repository.InstalledAppVersionHistoryRepository,
) *PipelineStatusTimelineServiceImpl {
	return &PipelineStatusTimelineServiceImpl{
		logger:                                 logger,
		pipelineStatusTimelineRepository:       pipelineStatusTimelineRepository,
		cdWorkflowRepository:                   cdWorkflowRepository,
		userService:                            userService,
		pipelineStatusTimelineResourcesService: pipelineStatusTimelineResourcesService,
		pipelineStatusSyncDetailService:        pipelineStatusSyncDetailService,
		installedAppRepository:                 installedAppRepository,
		installedAppVersionHistory:             installedAppVersionHistory,
	}
}

type PipelineTimelineDetailDto struct {
	DeploymentStartedOn        time.Time                    `json:"deploymentStartedOn"`
	DeploymentFinishedOn       time.Time                    `json:"deploymentFinishedOn"`
	TriggeredBy                string                       `json:"triggeredBy"`
	Timelines                  []*PipelineStatusTimelineDto `json:"timelines"`
	StatusLastFetchedAt        time.Time                    `json:"statusLastFetchedAt"`
	StatusFetchCount           int                          `json:"statusFetchCount"`
	WfrStatus                  string                       `json:"wfrStatus"`
	DeploymentAppDeleteRequest bool                         `json:"deploymentAppDeleteRequest"`
}

type PipelineStatusTimelineDto struct {
	Id                           int                           `json:"id"`
	InstalledAppVersionHistoryId int                           `json:"InstalledAppVersionHistoryId,omitempty"`
	CdWorkflowRunnerId           int                           `json:"cdWorkflowRunnerId"`
	Status                       pipelineConfig.TimelineStatus `json:"status"`
	StatusDetail                 string                        `json:"statusDetail"`
	StatusTime                   time.Time                     `json:"statusTime"`
	ResourceDetails              []*SyncStageResourceDetailDto `json:"resourceDetails,omitempty"`
}

func (impl *PipelineStatusTimelineServiceImpl) SaveTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx, isAppStore bool) error {
	var err error
	var redundantTimelines []*pipelineConfig.PipelineStatusTimeline
	if isAppStore {
		//get unableToFetch or timedOut timeline
		redundantTimelines, err = impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndPipelineStatuses(timeline.InstalledAppVersionHistoryId, []pipelineConfig.TimelineStatus{pipelineConfig.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS, pipelineConfig.TIMELINE_STATUS_FETCH_TIMED_OUT})
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting unableToFetch/timedOut timelines", "err", err, "installedAppVersionHistoryId", timeline.InstalledAppVersionHistoryId)
			return err
		}
	} else {
		redundantTimelines, err = impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatuses(timeline.CdWorkflowRunnerId, []pipelineConfig.TimelineStatus{pipelineConfig.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS, pipelineConfig.TIMELINE_STATUS_FETCH_TIMED_OUT})
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting unableToFetch/timedOut timelines", "err", err, "cdWfrId", timeline.CdWorkflowRunnerId)
			return err
		}
	}
	if len(redundantTimelines) > 1 {
		return fmt.Errorf("multiple unableToFetch/timedOut timelines found")
	} else if len(redundantTimelines) == 1 && redundantTimelines[0].Id > 0 {
		timeline.Id = redundantTimelines[0].Id
	} else {
		// do nothing
	}

	//saving/updating timeline
	err = impl.saveOrUpdateTimeline(timeline, tx)
	if err != nil {
		impl.logger.Errorw("error in saving/updating timeline", "err", err, "timeline", timeline)
		return err
	}
	return nil
}
func (impl *PipelineStatusTimelineServiceImpl) GetTimelineDbObjectByTimelineStatusAndTimelineDescription(cdWorkflowRunnerId int, installedAppVersionHistoryId int, timelineStatus pipelineConfig.TimelineStatus, timelineDescription string, userId int32, statusTime time.Time) *pipelineConfig.PipelineStatusTimeline {
	timeline := &pipelineConfig.PipelineStatusTimeline{
		CdWorkflowRunnerId:           cdWorkflowRunnerId,
		InstalledAppVersionHistoryId: installedAppVersionHistoryId,
		Status:                       timelineStatus,
		StatusDetail:                 timelineDescription,
		StatusTime:                   time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: userId,
			CreatedOn: time.Now(),
			UpdatedBy: userId,
			UpdatedOn: time.Now(),
		},
	}
	return timeline
}

func (impl *PipelineStatusTimelineServiceImpl) saveOrUpdateTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error {
	if tx == nil {
		if timeline.Id > 0 {
			err := impl.pipelineStatusTimelineRepository.UpdateTimelines([]*pipelineConfig.PipelineStatusTimeline{timeline})
			if err != nil {
				impl.logger.Errorw("error in updating timeline", "err", err, "timeline", timeline)
				return err
			}
		} else {
			err := impl.pipelineStatusTimelineRepository.SaveTimelines([]*pipelineConfig.PipelineStatusTimeline{timeline})
			if err != nil {
				impl.logger.Errorw("error in saving timeline", "err", err, "timeline", timeline)
				return err
			}
		}
	} else {
		if timeline.Id > 0 {
			err := impl.pipelineStatusTimelineRepository.UpdateTimelinesWithTxn([]*pipelineConfig.PipelineStatusTimeline{timeline}, tx)
			if err != nil {
				impl.logger.Errorw("error in updating timeline", "err", err, "timeline", timeline)
				return err
			}
		} else {
			err := impl.pipelineStatusTimelineRepository.SaveTimelinesWithTxn([]*pipelineConfig.PipelineStatusTimeline{timeline}, tx)
			if err != nil {
				impl.logger.Errorw("error in saving timeline", "err", err, "timeline", timeline)
				return err
			}
		}
	}
	return nil
}

func (impl *PipelineStatusTimelineServiceImpl) FetchTimelines(appId, envId, wfrId int, showTimeline bool) (*PipelineTimelineDetailDto, error) {
	var triggeredBy int32
	var deploymentStartedOn time.Time
	var deploymentFinishedOn time.Time
	var wfrStatus string
	var deploymentAppType string
	var err error
	wfr := &pipelineConfig.CdWorkflowRunner{}
	if wfrId == 0 {
		//fetch latest wfr by app and env
		wfr, err = impl.cdWorkflowRepository.FindLatestWfrByAppIdAndEnvironmentId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
			return nil, err
		}
		wfrId = wfr.Id
	} else {
		//fetch latest wfr by id
		wfr, err = impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
		if err != nil {
			impl.logger.Errorw("error in getting wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
			return nil, err
		}
	}
	deploymentStartedOn = wfr.StartedOn
	deploymentFinishedOn = wfr.FinishedOn
	triggeredBy = wfr.TriggeredBy
	wfrStatus = wfr.Status
	deploymentAppType = wfr.CdWorkflow.Pipeline.DeploymentAppType
	triggeredByUserEmailId, err := impl.userService.GetEmailById(triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in getting user email by id", "err", err, "userId", triggeredBy)
		return nil, err
	}
	var timelineDtos []*PipelineStatusTimelineDto
	var statusLastFetchedAt time.Time
	var statusFetchCount int
	if util.IsAcdApp(deploymentAppType) && showTimeline {
		timelines, err := impl.pipelineStatusTimelineRepository.FetchTimelinesByWfrId(wfrId)
		if err != nil {
			impl.logger.Errorw("error in getting timelines by wfrId", "err", err, "wfrId", wfrId)
			return nil, err
		}
		var cdWorkflowRunnerIds []int
		for _, timeline := range timelines {
			cdWorkflowRunnerIds = append(cdWorkflowRunnerIds, timeline.CdWorkflowRunnerId)
		}
		if len(cdWorkflowRunnerIds) == 0 {
			return nil, err
		}
		timelineResourceMap, err := impl.pipelineStatusTimelineResourcesService.GetTimelineResourcesForATimeline(cdWorkflowRunnerIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting timeline resources details", "err", err)
			return nil, err
		}
		for _, timeline := range timelines {
			timelineResourceDetails := make([]*SyncStageResourceDetailDto, 0)
			if timeline.Status == pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_STARTED {
				timelineResourceDetails = timelineResourceMap[timeline.CdWorkflowRunnerId]
			}
			timelineDto := &PipelineStatusTimelineDto{
				Id:                 timeline.Id,
				CdWorkflowRunnerId: timeline.CdWorkflowRunnerId,
				Status:             timeline.Status,
				StatusTime:         timeline.StatusTime,
				StatusDetail:       timeline.StatusDetail,
				ResourceDetails:    timelineResourceDetails,
			}
			timelineDtos = append(timelineDtos, timelineDto)
		}
		statusLastFetchedAt, statusFetchCount, err = impl.pipelineStatusSyncDetailService.GetSyncTimeAndCountByCdWfrId(wfrId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline status fetchTime and fetchCount by cdWfrId", "err", err, "cdWfrId", wfrId)
		}
	}
	timelineDetail := &PipelineTimelineDetailDto{
		TriggeredBy:                triggeredByUserEmailId,
		DeploymentStartedOn:        deploymentStartedOn,
		DeploymentFinishedOn:       deploymentFinishedOn,
		Timelines:                  timelineDtos,
		StatusLastFetchedAt:        statusLastFetchedAt,
		StatusFetchCount:           statusFetchCount,
		WfrStatus:                  wfrStatus,
		DeploymentAppDeleteRequest: wfr.CdWorkflow.Pipeline.DeploymentAppDeleteRequest,
	}
	return timelineDetail, nil
}

func (impl *PipelineStatusTimelineServiceImpl) FetchTimelinesForAppStore(installedAppId, envId, installedAppVersionHistoryId int, showTimeline bool) (*PipelineTimelineDetailDto, error) {
	var deploymentStartedOn time.Time
	var deploymentFinishedOn time.Time
	var installedAppVersionHistoryStatus string
	var deploymentAppType string
	var err error
	installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting installed_app_version by appId and envId", "err", err, "appId", installedAppId, "envId", envId)
		return nil, err
	}
	installedAppVersionHistory := &repository.InstalledAppVersionHistory{}
	if installedAppVersionHistoryId == 0 {
		//fetching latest installed_app_version_history from installed_app_version_id
		installedAppVersionHistory, err = impl.installedAppVersionHistory.GetLatestInstalledAppVersionHistory(installedAppVersion.Id)
		if err != nil {
			impl.logger.Errorw("error in getting installed_app_version_history by installAppVersionId", "err", err, "appId", installedAppVersionHistoryId, "envId", envId)
			return nil, err
		}
		installedAppVersionHistoryId = installedAppVersionHistory.Id
	} else {
		//fetching installed_app_version_history directly from installedAppVersionHistoryId
		installedAppVersionHistory, err = impl.installedAppVersionHistory.GetInstalledAppVersionHistory(installedAppVersionHistoryId)
		if err != nil {
			impl.logger.Errorw("error in getting installed_app_version_history by installAppVersionHistoryId", "err", err, "appId", installedAppVersionHistoryId, "envId", envId)
			return nil, err
		}
	}
	if installedAppVersionHistory.StartedOn.IsZero() && installedAppVersionHistory.FinishedOn.IsZero() {
		deploymentStartedOn = installedAppVersionHistory.CreatedOn
		deploymentFinishedOn = installedAppVersionHistory.UpdatedOn
	} else {
		deploymentStartedOn = installedAppVersionHistory.StartedOn
		deploymentFinishedOn = installedAppVersionHistory.FinishedOn
	}
	installedAppVersionHistoryStatus = installedAppVersionHistory.Status
	deploymentAppType = installedAppVersion.InstalledApp.DeploymentAppType
	triggeredByUserEmailId, err := impl.userService.GetEmailById(installedAppVersionHistory.CreatedBy)
	if err != nil {
		impl.logger.Errorw("error in getting user email by id", "err", err, "userId", installedAppVersionHistory.CreatedBy)
		return nil, err
	}
	var timelineDtos []*PipelineStatusTimelineDto
	var statusLastFetchedAt time.Time
	var statusFetchCount int
	if util.IsAcdApp(deploymentAppType) && showTimeline {
		timelines, err := impl.pipelineStatusTimelineRepository.FetchTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId)
		if err != nil {
			impl.logger.Errorw("error in getting timelines by installedAppVersionHistoryId", "err", err, "wfrId", installedAppVersionHistoryId)
			return nil, err
		}
		for _, timeline := range timelines {
			var timelineResourceDetails []*SyncStageResourceDetailDto
			if timeline.Status == pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_STARTED {
				timelineResourceDetails, err = impl.pipelineStatusTimelineResourcesService.GetTimelineResourcesForATimelineForAppStore(timeline.InstalledAppVersionHistoryId)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in getting timeline resources details", "err", err, "installedAppVersionHistoryId", timeline.InstalledAppVersionHistoryId)
					return nil, err
				}
			}
			timelineDto := &PipelineStatusTimelineDto{
				Id:                           timeline.Id,
				InstalledAppVersionHistoryId: timeline.InstalledAppVersionHistoryId,
				Status:                       timeline.Status,
				StatusTime:                   timeline.StatusTime,
				StatusDetail:                 timeline.StatusDetail,
				ResourceDetails:              timelineResourceDetails,
			}
			timelineDtos = append(timelineDtos, timelineDto)
		}
		statusLastFetchedAt, statusFetchCount, err = impl.pipelineStatusSyncDetailService.GetSyncTimeAndCountByInstalledAppVersionHistoryId(installedAppVersionHistoryId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline status fetchTime and fetchCount by installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installedAppVersionHistoryId)
		}
	}
	timelineDetail := &PipelineTimelineDetailDto{
		TriggeredBy:                triggeredByUserEmailId,
		DeploymentStartedOn:        deploymentStartedOn,
		DeploymentFinishedOn:       deploymentFinishedOn,
		Timelines:                  timelineDtos,
		StatusLastFetchedAt:        statusLastFetchedAt,
		StatusFetchCount:           statusFetchCount,
		WfrStatus:                  installedAppVersionHistoryStatus,
		DeploymentAppDeleteRequest: false,
	}
	return timelineDetail, nil
}

func (impl *PipelineStatusTimelineServiceImpl) SavePipelineStatusTimelineIfNotAlreadyPresent(pipelineId int, timelineStatus pipelineConfig.TimelineStatus, timeline *pipelineConfig.PipelineStatusTimeline, isAppStore bool) (latestTimeline *pipelineConfig.PipelineStatusTimeline, err error, isTimelineUpdated bool) {
	isTimelineUpdated = false
	if isAppStore {
		latestTimeline, err = impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(pipelineId, timelineStatus)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline", "err", err)
			return nil, err, isTimelineUpdated
		} else if err == pg.ErrNoRows {
			err = impl.SaveTimeline(timeline, nil, true)
			if err != nil {
				impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
				return nil, err, isTimelineUpdated
			}
			isTimelineUpdated = true
			latestTimeline = timeline
		}
	} else {
		latestTimeline, err = impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatus(pipelineId, timelineStatus)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline", "err", err)
			return nil, err, isTimelineUpdated
		} else if err == pg.ErrNoRows {
			err = impl.SaveTimeline(timeline, nil, false)
			if err != nil {
				impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
				return nil, err, isTimelineUpdated
			}
			isTimelineUpdated = true
			latestTimeline = timeline
		}
	}
	return latestTimeline, nil, isTimelineUpdated
}

func (impl *PipelineStatusTimelineServiceImpl) GetArgoAppSyncStatus(cdWfrId int) bool {
	timeline, err := impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatus(cdWfrId, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED)
	if err != nil {
		impl.logger.Errorw("error in fetching argocd sync status", "err", err)
		return false
	}
	if timeline != nil && timeline.Id == 0 {
		return false
	}
	return true
}

func (impl *PipelineStatusTimelineServiceImpl) GetArgoAppSyncStatusForAppStore(installedAppVersionHistoryId int) bool {
	timeline, err := impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(installedAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED)
	if err != nil {
		impl.logger.Errorw("error in fetching argocd sync status", "err", err)
		return false
	}
	if timeline != nil && timeline.Id == 0 {
		return false
	}
	return true
}

func (impl *PipelineStatusTimelineServiceImpl) SaveTimelines(timeline []*pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error {
	_, err := tx.Model(&timeline).Insert()
	if err != nil {
		return err
	}
	return err
}

func (impl *PipelineStatusTimelineServiceImpl) MarkPipelineStatusTimelineFailed(cdWfrId int, statusDetailMessage string) error {
	//creating cd pipeline status timeline for deployment failed
	terminalStatusExists, timelineErr := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByWfrId(cdWfrId)
	if timelineErr != nil {
		impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", timelineErr, "wfrId", cdWfrId)
		return timelineErr
	}
	if !terminalStatusExists {
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: cdWfrId,
			Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED,
			StatusDetail:       statusDetailMessage,
			StatusTime:         time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		impl.logger.Infow("marking pipeline deployment failed", "cdWfrId", cdWfrId, "statusDetail", statusDetailMessage)
		timelineErr = impl.SaveTimeline(timeline, nil, false)
		if timelineErr != nil {
			impl.logger.Errorw("error in creating timeline status for deployment fail", "err", timelineErr, "timeline", timeline)
		}
	}
	return nil
}
