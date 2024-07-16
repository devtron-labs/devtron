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
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	"github.com/devtron-labs/devtron/pkg/app/status/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PipelineStatusTimelineService interface {
	SaveTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error
	FetchTimelines(appId, envId, wfrId int, showTimeline bool) (*PipelineTimelineDetailDto, error)
	FetchTimelinesForAppStore(installedAppId, envId, installedAppVersionHistoryId int, showTimeline bool) (*PipelineTimelineDetailDto, error)
	NewDevtronAppPipelineStatusTimelineDbObject(cdWorkflowRunnerId int, timelineStatus timelineStatus.TimelineStatus, timelineDescription string, userId int32) *pipelineConfig.PipelineStatusTimeline
	NewHelmAppDeploymentStatusTimelineDbObject(installedAppVersionHistoryId int, timelineStatus timelineStatus.TimelineStatus, timelineDescription string, userId int32) *pipelineConfig.PipelineStatusTimeline
	SaveTimelineIfNotAlreadyPresent(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) (isTimelineUpdated bool, err error)
	GetArgoAppSyncStatus(cdWfrId int) bool
	GetTimelineStatusesFor(request *bean.TimelineGetRequest) ([]timelineStatus.TimelineStatus, error)
	GetArgoAppSyncStatusForAppStore(installedAppVersionHistoryId int) bool
	SaveMultipleTimelinesIfNotAlreadyPresent(timelines []*pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error

	MarkPipelineStatusTimelineFailed(cdWfrId int, statusDetailMessage string) error
	MarkPipelineStatusTimelineSuperseded(cdWfrId int) error
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
	deploymentConfigService                common.DeploymentConfigService
}

func NewPipelineStatusTimelineServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	userService user.UserService,
	pipelineStatusTimelineResourcesService PipelineStatusTimelineResourcesService,
	pipelineStatusSyncDetailService PipelineStatusSyncDetailService,
	installedAppRepository repository.InstalledAppRepository,
	installedAppVersionHistory repository.InstalledAppVersionHistoryRepository,
	deploymentConfigService common.DeploymentConfigService,
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
		deploymentConfigService:                deploymentConfigService,
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
	Status                       timelineStatus.TimelineStatus `json:"status"`
	StatusDetail                 string                        `json:"statusDetail"`
	StatusTime                   time.Time                     `json:"statusTime"`
	ResourceDetails              []*SyncStageResourceDetailDto `json:"resourceDetails,omitempty"`
}

func (impl *PipelineStatusTimelineServiceImpl) SaveTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error {
	var err error
	var redundantTimelines []*pipelineConfig.PipelineStatusTimeline
	if timeline.InstalledAppVersionHistoryId != 0 {
		//get unableToFetch or timedOut timeline
		redundantTimelines, err = impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndPipelineStatuses(timeline.InstalledAppVersionHistoryId, []timelineStatus.TimelineStatus{timelineStatus.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS, timelineStatus.TIMELINE_STATUS_FETCH_TIMED_OUT})
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in getting unableToFetch/timedOut timelines", "err", err, "installedAppVersionHistoryId", timeline.InstalledAppVersionHistoryId)
			return err
		}
	} else if timeline.CdWorkflowRunnerId != 0 {
		redundantTimelines, err = impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatuses(timeline.CdWorkflowRunnerId, []timelineStatus.TimelineStatus{timelineStatus.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS, timelineStatus.TIMELINE_STATUS_FETCH_TIMED_OUT})
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in getting unableToFetch/timedOut timelines", "err", err, "cdWfrId", timeline.CdWorkflowRunnerId)
			return err
		}
	} else {
		return fmt.Errorf("invalide timeline object recieved to be saved")
	}
	if len(redundantTimelines) > 1 {
		return fmt.Errorf("multiple unableToFetch/TimedOut timelines found")
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

func (impl *PipelineStatusTimelineServiceImpl) NewDevtronAppPipelineStatusTimelineDbObject(cdWorkflowRunnerId int, timelineStatus timelineStatus.TimelineStatus, timelineDescription string, userId int32) *pipelineConfig.PipelineStatusTimeline {
	timeline := &pipelineConfig.PipelineStatusTimeline{
		CdWorkflowRunnerId: cdWorkflowRunnerId,
		Status:             timelineStatus,
		StatusDetail:       timelineDescription,
		StatusTime:         time.Now(),
	}
	timeline.CreateAuditLog(userId)
	return timeline
}

func (impl *PipelineStatusTimelineServiceImpl) NewHelmAppDeploymentStatusTimelineDbObject(installedAppVersionHistoryId int, timelineStatus timelineStatus.TimelineStatus, timelineDescription string, userId int32) *pipelineConfig.PipelineStatusTimeline {
	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installedAppVersionHistoryId,
		Status:                       timelineStatus,
		StatusDetail:                 timelineDescription,
		StatusTime:                   time.Now(),
	}
	timeline.CreateAuditLog(userId)
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

	envDeploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching environment deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	deploymentAppType = envDeploymentConfig.DeploymentAppType
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
			if timeline.Status == timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_STARTED {
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
	deploymentConfig, err := impl.deploymentConfigService.GetConfigForHelmApps(installedAppVersion.InstalledApp.AppId, installedAppVersion.InstalledApp.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", installedAppVersion.InstalledApp.AppId, "envId", installedAppVersion.InstalledApp.EnvironmentId, "err", err)
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
	deploymentAppType = deploymentConfig.DeploymentAppType
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
			if timeline.Status == timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_STARTED {
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

func (impl *PipelineStatusTimelineServiceImpl) SaveTimelineIfNotAlreadyPresent(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) (isTimelineUpdated bool, err error) {
	isTimelineUpdated = false
	if timeline.InstalledAppVersionHistoryId != 0 {
		terminalStatusExists, timelineErr := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(timeline.InstalledAppVersionHistoryId)
		if timelineErr != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by installedAppVersionHistoryId", "err", timelineErr, "installedAppVersionHistoryId", timeline.InstalledAppVersionHistoryId)
			return isTimelineUpdated, timelineErr
		}
		if terminalStatusExists {
			return isTimelineUpdated, nil
		}
		_, err := impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(timeline.InstalledAppVersionHistoryId, timeline.Status)
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in getting latest timeline", "err", err)
			return isTimelineUpdated, err
		} else if errors.Is(err, pg.ErrNoRows) {
			err = impl.SaveTimeline(timeline, tx)
			if err != nil {
				impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
				return isTimelineUpdated, err
			}
			isTimelineUpdated = true
		}
	} else if timeline.CdWorkflowRunnerId != 0 {
		terminalStatusExists, timelineErr := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByWfrId(timeline.CdWorkflowRunnerId)
		if timelineErr != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", timelineErr, "wfrId", timeline.CdWorkflowRunnerId)
			return isTimelineUpdated, timelineErr
		}
		if terminalStatusExists {
			return isTimelineUpdated, nil
		}
		_, err := impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatus(timeline.CdWorkflowRunnerId, timeline.Status)
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in getting latest timeline", "err", err)
			return isTimelineUpdated, err
		} else if errors.Is(err, pg.ErrNoRows) {
			err = impl.SaveTimeline(timeline, tx)
			if err != nil {
				impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
				return isTimelineUpdated, err
			}
			isTimelineUpdated = true
		}
	}
	return isTimelineUpdated, nil
}

func (impl *PipelineStatusTimelineServiceImpl) GetArgoAppSyncStatus(cdWfrId int) bool {
	timeline, err := impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatus(cdWfrId, timelineStatus.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED)
	if err != nil {
		impl.logger.Errorw("error in fetching argocd sync status", "err", err)
		return false
	}
	if timeline != nil && timeline.Id == 0 {
		return false
	}
	return true
}

func (impl *PipelineStatusTimelineServiceImpl) GetTimelineStatusesFor(request *bean.TimelineGetRequest) ([]timelineStatus.TimelineStatus, error) {
	timelines, err := impl.pipelineStatusTimelineRepository.FetchTimelinesForWfrIdExcludingStatuses(request.GetCdWfrId(), request.GetExcludingStatuses()...)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ArgoCd sync status", "err", err)
		return []timelineStatus.TimelineStatus{}, err
	} else if util.IsErrNoRows(err) {
		return []timelineStatus.TimelineStatus{}, nil
	}
	timelineStatuses := make([]timelineStatus.TimelineStatus, 0, len(timelines))
	for _, timeline := range timelines {
		timelineStatuses = append(timelineStatuses, timeline.Status)
	}
	return timelineStatuses, nil
}

func (impl *PipelineStatusTimelineServiceImpl) GetArgoAppSyncStatusForAppStore(installedAppVersionHistoryId int) bool {
	timeline, err := impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(installedAppVersionHistoryId, timelineStatus.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED)
	if err != nil {
		impl.logger.Errorw("error in fetching argocd sync status", "err", err)
		return false
	}
	if timeline != nil && timeline.Id == 0 {
		return false
	}
	return true
}

func (impl *PipelineStatusTimelineServiceImpl) SaveMultipleTimelinesIfNotAlreadyPresent(timelines []*pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error {
	for _, timeline := range timelines {
		if timeline.InstalledAppVersionHistoryId != 0 {
			_, err := impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(timeline.InstalledAppVersionHistoryId, timeline.Status)
			if err != nil && !errors.Is(err, pg.ErrNoRows) {
				impl.logger.Errorw("error in getting latest timeline", "err", err)
				return err
			} else if errors.Is(err, pg.ErrNoRows) {
				err = impl.SaveTimeline(timeline, tx)
				if err != nil {
					return err
				}
			}
		} else if timeline.CdWorkflowRunnerId != 0 {
			_, err := impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatus(timeline.CdWorkflowRunnerId, timeline.Status)
			if err != nil && !errors.Is(err, pg.ErrNoRows) {
				impl.logger.Errorw("error in getting latest timeline", "err", err)
				return err
			} else if errors.Is(err, pg.ErrNoRows) {
				err = impl.SaveTimeline(timeline, tx)
				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("invalide timeline object recieved to be saved")
		}
	}
	return nil
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
			Status:             timelineStatus.TIMELINE_STATUS_DEPLOYMENT_FAILED,
			StatusDetail:       statusDetailMessage,
			StatusTime:         time.Now(),
		}
		timeline.CreateAuditLog(1)
		impl.logger.Infow("marking pipeline deployment failed", "cdWfrId", cdWfrId, "statusDetail", statusDetailMessage)
		timelineErr = impl.SaveTimeline(timeline, nil)
		if timelineErr != nil {
			impl.logger.Errorw("error in creating timeline status for deployment fail", "err", timelineErr, "timeline", timeline)
		}
	}
	return nil
}

func (impl *PipelineStatusTimelineServiceImpl) MarkPipelineStatusTimelineSuperseded(cdWfrId int) error {
	//creating cd pipeline status timeline for deployment failed
	terminalStatusExists, timelineErr := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByWfrId(cdWfrId)
	if timelineErr != nil {
		impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", timelineErr, "wfrId", cdWfrId)
		return timelineErr
	}
	if !terminalStatusExists {
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: cdWfrId,
			Status:             timelineStatus.TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
			StatusDetail:       timelineStatus.TIMELINE_DESCRIPTION_DEPLOYMENT_SUPERSEDED,
			StatusTime:         time.Now(),
		}
		timeline.CreateAuditLog(1)
		impl.logger.Infow("marking pipeline deployment superseded", "cdWfrId", cdWfrId)
		timelineErr = impl.SaveTimeline(timeline, nil)
		if timelineErr != nil {
			impl.logger.Errorw("error in creating timeline status for deployment superseded", "err", timelineErr, "timeline", timeline)
		}
	}
	return nil
}
