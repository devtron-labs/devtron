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

package deployment

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"time"
)

type DeploymentStatusService interface {
	// TODO refactoring: Move to DB service
	SaveTimelineForHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status timelineStatus.TimelineStatus, statusDetail string, statusTime time.Time, tx *pg.Tx) error
	// UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus updates failed status in pipelineConfig.PipelineStatusTimeline table
	UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error
}

func (impl *FullModeDeploymentServiceImpl) SaveTimelineForHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status timelineStatus.TimelineStatus, statusDetail string, statusTime time.Time, tx *pg.Tx) error {

	if !util.IsAcdApp(installAppVersionRequest.DeploymentAppType) && !util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		return nil
	}

	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
		Status:                       status,
		StatusDetail:                 statusDetail,
		StatusTime:                   statusTime,
		AuditLog: sql.AuditLog{
			CreatedBy: installAppVersionRequest.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: installAppVersionRequest.UserId,
			UpdatedOn: time.Now(),
		},
	}
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx)
	if timelineErr != nil {
		impl.Logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
	}
	return timelineErr
}

func (impl *FullModeDeploymentServiceImpl) UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error {
	if err != nil {
		terminalStatusExists, timelineErr := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(installAppVersionRequest.InstalledAppVersionHistoryId)
		if timelineErr != nil {
			impl.Logger.Errorw("error in checking if terminal status timeline exists by installedAppVersionHistoryId", "err", timelineErr, "installedAppVersionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId)
			return timelineErr
		}
		if !terminalStatusExists {
			impl.Logger.Infow("marking pipeline deployment failed", "err", err)
			timeline := &pipelineConfig.PipelineStatusTimeline{
				InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
				Status:                       timelineStatus.TIMELINE_STATUS_DEPLOYMENT_FAILED,
				StatusDetail:                 fmt.Sprintf("Deployment failed: %v", err),
				StatusTime:                   time.Now(),
			}
			timeline.CreateAuditLog(1)
			timelineErr = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil)
			if timelineErr != nil {
				impl.Logger.Errorw("error in creating timeline status for deployment fail", "err", timelineErr, "timeline", timeline)
			}
		}
		impl.Logger.Errorw("error in triggering installed application deployment, setting status as fail ", "versionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId, "err", err)

		installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installAppVersionRequest.InstalledAppVersionHistoryId)
		if err != nil {
			impl.Logger.Errorw("error in getting installedAppVersionHistory by installedAppVersionHistoryId", "installedAppVersionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId, "err", err)
			return err
		}
		installedAppVersionHistory.SetStatus(pipelineConfig.WorkflowFailed)
		installedAppVersionHistory.FinishedOn = triggeredAt
		installedAppVersionHistory.UpdateAuditLog(installAppVersionRequest.UserId)
		_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
		if err != nil {
			impl.Logger.Errorw("error updating installed app version history status", "err", err, "installedAppVersionHistory", installedAppVersionHistory)
			return err
		}

	} else {
		//update [n,n-1] statuses as failed if not terminal
		terminalStatus := []string{string(health.HealthStatusHealthy), pipelineConfig.WorkflowAborted, pipelineConfig.WorkflowFailed, pipelineConfig.WorkflowSucceeded}
		previousNonTerminalHistory, err := impl.installedAppRepositoryHistory.FindPreviousInstalledAppVersionHistoryByStatus(installAppVersionRequest.Id, installAppVersionRequest.InstalledAppVersionHistoryId, terminalStatus)
		if err != nil {
			impl.Logger.Errorw("error fetching previous installed app version history, updating installed app version history status,", "err", err, "installAppVersionRequest", installAppVersionRequest)
			return err
		} else if len(previousNonTerminalHistory) == 0 {
			impl.Logger.Errorw("no previous history found in updating installedAppVersionHistory status,", "err", err, "installAppVersionRequest", installAppVersionRequest)
			return nil
		}
		dbConnection := impl.installedAppRepositoryHistory.GetConnection()
		tx, err := dbConnection.Begin()
		if err != nil {
			impl.Logger.Errorw("error on update status, txn begin failed", "err", err)
			return err
		}
		// Rollback tx on error.
		defer tx.Rollback()
		var timelines []*pipelineConfig.PipelineStatusTimeline
		for _, previousHistory := range previousNonTerminalHistory {
			if previousHistory.Status == string(health.HealthStatusHealthy) ||
				previousHistory.Status == pipelineConfig.WorkflowSucceeded ||
				previousHistory.Status == pipelineConfig.WorkflowAborted ||
				previousHistory.Status == pipelineConfig.WorkflowFailed {
				//terminal status return
				impl.Logger.Infow("skip updating installedAppVersionHistory status as previous history status is", "status", previousHistory.Status)
				continue
			}
			impl.Logger.Infow("updating installedAppVersionHistory status as previous runner status is", "status", previousHistory.Status)
			previousHistory.FinishedOn = triggeredAt
			previousHistory.Status = pipelineConfig.WorkflowFailed
			previousHistory.UpdatedOn = time.Now()
			previousHistory.UpdatedBy = installAppVersionRequest.UserId
			timeline := &pipelineConfig.PipelineStatusTimeline{
				InstalledAppVersionHistoryId: previousHistory.Id,
				Status:                       timelineStatus.TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
				StatusDetail:                 "This deployment is superseded.",
				StatusTime:                   time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: 1,
					CreatedOn: time.Now(),
					UpdatedBy: 1,
					UpdatedOn: time.Now(),
				},
			}
			timelines = append(timelines, timeline)
		}

		err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistoryWithTxn(previousNonTerminalHistory, tx)
		if err != nil {
			impl.Logger.Errorw("error updating cd wf runner status", "err", err, "previousNonTerminalHistory", previousNonTerminalHistory)
			return err
		}
		err = impl.pipelineStatusTimelineRepository.SaveTimelinesWithTxn(timelines, tx)
		if err != nil {
			impl.Logger.Errorw("error updating pipeline status timelines", "err", err, "timelines", timelines)
			return err
		}
		err = tx.Commit()
		if err != nil {
			impl.Logger.Errorw("error in db transaction commit", "err", err)
			return err
		}
	}
	return nil
}
