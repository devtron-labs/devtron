/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pipeline

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	pubub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/app/status"
	appGroup2 "github.com/devtron-labs/devtron/pkg/appGroup"
	app_status "github.com/devtron-labs/devtron/pkg/appStatus"
	repository3 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CdHandler interface {
	UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, string, error)
	GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]pipelineConfig.CdWorkflowWithArtifact, error)
	GetRunningWorkflowLogs(environmentId int, pipelineId int, workflowId int) (*bufio.Reader, func() error, error)
	FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int) (WorkflowResponse, error)
	DownloadCdWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error)
	FetchCdPrePostStageStatus(pipelineId int) ([]pipelineConfig.CdWorkflowWithArtifact, error)
	CancelStage(workflowRunnerId int, userId int32) (int, error)
	FetchAppWorkflowStatusForTriggerView(appId int) ([]*pipelineConfig.CdWorkflowStatus, error)
	CheckHelmAppStatusPeriodicallyAndUpdateInDb(helmPipelineStatusCheckEligibleTime int) error
	CheckArgoAppStatusPeriodicallyAndUpdateInDb(getPipelineDeployedBeforeMinutes int, getPipelineDeployedWithinHours int) error
	CheckArgoPipelineTimelineStatusPeriodicallyAndUpdateInDb(pendingSinceSeconds int, timeForDegradation int) error
	UpdatePipelineTimelineAndStatusByLiveApplicationFetch(pipeline *pipelineConfig.Pipeline, installedApp repository3.InstalledApps, userId int32) (err error, isTimelineUpdated bool)
	CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipelineId int, userId int32, isAppStoreApplication bool)
	FetchAppWorkflowStatusForTriggerViewForEnvironment(request appGroup2.AppGroupingRequest) ([]*pipelineConfig.CdWorkflowStatus, error)
	FetchAppDeploymentStatusForEnvironments(request appGroup2.AppGroupingRequest) ([]*pipelineConfig.AppDeploymentStatus, error)
}

type CdHandlerImpl struct {
	Logger                                 *zap.SugaredLogger
	cdService                              CdWorkflowService
	cdConfig                               *CdConfig
	ciConfig                               *CiConfig
	userService                            user.UserService
	ciLogService                           CiLogService
	ciArtifactRepository                   repository.CiArtifactRepository
	ciPipelineMaterialRepository           pipelineConfig.CiPipelineMaterialRepository
	cdWorkflowRepository                   pipelineConfig.CdWorkflowRepository
	envRepository                          repository2.EnvironmentRepository
	pipelineRepository                     pipelineConfig.PipelineRepository
	ciWorkflowRepository                   pipelineConfig.CiWorkflowRepository
	helmAppService                         client.HelmAppService
	pipelineOverrideRepository             chartConfig.PipelineOverrideRepository
	workflowDagExecutor                    WorkflowDagExecutor
	appListingService                      app.AppListingService
	appListingRepository                   repository.AppListingRepository
	pipelineStatusTimelineRepository       pipelineConfig.PipelineStatusTimelineRepository
	application                            application.ServiceClient
	argoUserService                        argo.ArgoUserService
	deploymentEventHandler                 app.DeploymentEventHandler
	eventClient                            client2.EventClient
	pipelineStatusTimelineResourcesService status.PipelineStatusTimelineResourcesService
	pipelineStatusSyncDetailService        status.PipelineStatusSyncDetailService
	pipelineStatusTimelineService          status.PipelineStatusTimelineService
	appService                             app.AppService
	appStatusService                       app_status.AppStatusService
	enforcerUtil                           rbac.EnforcerUtil
	installedAppRepository                 repository3.InstalledAppRepository
	installedAppVersionHistoryRepository   repository3.InstalledAppVersionHistoryRepository
	appRepository                          app2.AppRepository
	appGroupService                        appGroup2.AppGroupService
}

func NewCdHandlerImpl(Logger *zap.SugaredLogger, cdConfig *CdConfig, userService user.UserService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	cdWorkflowService CdWorkflowService,
	ciLogService CiLogService,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	envRepository repository2.EnvironmentRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciConfig *CiConfig, helmAppService client.HelmAppService,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, workflowDagExecutor WorkflowDagExecutor,
	appListingService app.AppListingService, appListingRepository repository.AppListingRepository,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	application application.ServiceClient, argoUserService argo.ArgoUserService,
	deploymentEventHandler app.DeploymentEventHandler,
	eventClient client2.EventClient,
	pipelineStatusTimelineResourcesService status.PipelineStatusTimelineResourcesService,
	pipelineStatusSyncDetailService status.PipelineStatusSyncDetailService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	appService app.AppService,
	appStatusService app_status.AppStatusService, enforcerUtil rbac.EnforcerUtil,
	installedAppRepository repository3.InstalledAppRepository,
	installedAppVersionHistoryRepository repository3.InstalledAppVersionHistoryRepository, appRepository app2.AppRepository,
	appGroupService appGroup2.AppGroupService) *CdHandlerImpl {
	return &CdHandlerImpl{
		Logger:                                 Logger,
		cdConfig:                               cdConfig,
		userService:                            userService,
		cdService:                              cdWorkflowService,
		ciLogService:                           ciLogService,
		cdWorkflowRepository:                   cdWorkflowRepository,
		ciArtifactRepository:                   ciArtifactRepository,
		ciPipelineMaterialRepository:           ciPipelineMaterialRepository,
		envRepository:                          envRepository,
		pipelineRepository:                     pipelineRepository,
		ciWorkflowRepository:                   ciWorkflowRepository,
		ciConfig:                               ciConfig,
		helmAppService:                         helmAppService,
		pipelineOverrideRepository:             pipelineOverrideRepository,
		workflowDagExecutor:                    workflowDagExecutor,
		appListingService:                      appListingService,
		appListingRepository:                   appListingRepository,
		pipelineStatusTimelineRepository:       pipelineStatusTimelineRepository,
		application:                            application,
		argoUserService:                        argoUserService,
		deploymentEventHandler:                 deploymentEventHandler,
		eventClient:                            eventClient,
		pipelineStatusTimelineResourcesService: pipelineStatusTimelineResourcesService,
		pipelineStatusSyncDetailService:        pipelineStatusSyncDetailService,
		pipelineStatusTimelineService:          pipelineStatusTimelineService,
		appService:                             appService,
		appStatusService:                       appStatusService,
		enforcerUtil:                           enforcerUtil,
		installedAppRepository:                 installedAppRepository,
		installedAppVersionHistoryRepository:   installedAppVersionHistoryRepository,
		appRepository:                          appRepository,
		appGroupService:                        appGroupService,
	}
}

type ArgoPipelineStatusSyncEvent struct {
	PipelineId            int   `json:"pipelineId"`
	UserId                int32 `json:"userId"`
	IsAppStoreApplication bool  `json:"isAppStoreApplication"`
}

const NotTriggered string = "Not Triggered"
const NotDeployed = "Not Deployed"
const WorklowTypeDeploy = "DEPLOY"
const WorklowTypePre = "PRE"
const WorklowTypePost = "POST"

func (impl *CdHandlerImpl) CheckArgoAppStatusPeriodicallyAndUpdateInDb(getPipelineDeployedBeforeMinutes int, getPipelineDeployedWithinHours int) error {
	pipelines, err := impl.pipelineRepository.GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatuses(getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours)
	if err != nil {
		impl.Logger.Errorw("error in getting pipelines having latest trigger stuck in non terminal statuses", "err", err)
		return err
	}
	impl.Logger.Debugw("received stuck argo cd pipelines", "pipelines", pipelines)

	installedAppVersions, err := impl.installedAppRepository.GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatusesForAppStore(getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours)
	if err != nil {
		impl.Logger.Errorw("error in getting installedAppVersions having latest trigger stuck in non terminal statuses", "err", err)
		return err
	}
	impl.Logger.Debugw("received stuck argo installed appStore app", "pipelines", installedAppVersions)

	for _, pipeline := range pipelines {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipeline.Id, 1, false)
	}
	for _, installedAppVersion := range installedAppVersions {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(installedAppVersion.Id, 1, true)
	}
	return nil
}

func (impl *CdHandlerImpl) CheckArgoPipelineTimelineStatusPeriodicallyAndUpdateInDb(pendingSinceSeconds int, timeForDegradation int) error {
	//getting all the progressing status that are stuck since some time after kubectl apply success sync stage
	//and are not eligible for CheckArgoAppStatusPeriodicallyAndUpdateInDb
	pipelines, err := impl.pipelineRepository.GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelines(pendingSinceSeconds, timeForDegradation)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err in GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelines", "err", err)
		return err
	}
	impl.Logger.Debugw("received argo cd pipelines stuck at kubectl apply synced stage", "pipelines", pipelines)

	installedAppVersions, err := impl.installedAppRepository.GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelinesForAppStore(pendingSinceSeconds, timeForDegradation)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err in GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelinesForAppStore", "err", err)
		return err
	}

	impl.Logger.Debugw("received argo appStore application stuck at kubectl apply synced stage", "pipelines", installedAppVersions)
	for _, pipeline := range pipelines {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipeline.Id, 1, false)
	}

	for _, installedAppVersion := range installedAppVersions {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(installedAppVersion.Id, 1, true)
	}
	return nil
}

func (impl *CdHandlerImpl) CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipelineId int, userId int32, isAppStoreApplication bool) {
	var lastSyncTime time.Time
	var err error
	if isAppStoreApplication {
		lastSyncTime, err = impl.pipelineStatusSyncDetailService.GetLastSyncTimeForLatestInstalledAppVersionHistoryByInstalledAppVersionId(pipelineId)
	} else {
		lastSyncTime, err = impl.pipelineStatusSyncDetailService.GetLastSyncTimeForLatestCdWfrByCdPipelineId(pipelineId)
	}
	if err != nil {
		impl.Logger.Errorw("error in getting last sync time by pipelineId", "err", err, "pipelineId", pipelineId)
		return
	}
	//TODO: remove hard coding
	//pipelineId can be cdPipelineId or installedAppVersionId, using isAppStoreApplication flag to identify between them
	if lastSyncTime.IsZero() || (!lastSyncTime.IsZero() && time.Since(lastSyncTime) > 5*time.Second) { //create new nats event
		statusUpdateEvent := ArgoPipelineStatusSyncEvent{
			PipelineId:            pipelineId,
			UserId:                userId,
			IsAppStoreApplication: isAppStoreApplication,
		}
		//write event
		err = impl.eventClient.WriteNatsEvent(pubub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, statusUpdateEvent)
		if err != nil {
			impl.Logger.Errorw("error in writing nats event", "topic", pubub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, "payload", statusUpdateEvent)
		}
	}
}

func (impl *CdHandlerImpl) UpdatePipelineTimelineAndStatusByLiveApplicationFetch(pipeline *pipelineConfig.Pipeline, installedApp repository3.InstalledApps, userId int32) (error, bool) {
	isTimelineUpdated := false
	isSucceeded := false
	if pipeline != nil {
		isAppStore := false
		cdWfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipeline.Id, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil {
			impl.Logger.Errorw("error in getting latest cdWfr by cdPipelineId", "err", err, "pipelineId", pipeline.Id)
			return nil, isTimelineUpdated
		}
		impl.Logger.Debugw("ARGO_PIPELINE_STATUS_UPDATE_REQ", "stage", "checkingDeploymentStatus", "argoAppName", pipeline, "cdWfr", cdWfr)
		if util3.IsTerminalStatus(cdWfr.Status) {
			//drop event
			return nil, isTimelineUpdated
		}
		//this should only be called when we have git-ops configured
		//try fetching status from argo cd
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.Logger.Errorw("error in getting acd token", "err", err)
		}
		ctx := context.WithValue(context.Background(), "token", acdToken)
		query := &application2.ApplicationQuery{
			Name: &pipeline.DeploymentAppName,
		}
		app, err := impl.application.Get(ctx, query)
		if err != nil {
			impl.Logger.Errorw("error in getting acd application", "err", err, "argoAppName", pipeline)
			//updating cdWfr status
			cdWfr.Status = pipelineConfig.WorkflowUnableToFetchState
			cdWfr.UpdatedOn = time.Now()
			cdWfr.UpdatedBy = 1
			err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&cdWfr)
			if err != nil {
				impl.Logger.Errorw("error on update cd workflow runner", "cdWfr", cdWfr, "err", err)
				return err, isTimelineUpdated
			}
			// creating cd pipeline status timeline
			timeline := &pipelineConfig.PipelineStatusTimeline{
				CdWorkflowRunnerId: cdWfr.Id,
				Status:             pipelineConfig.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS,
				StatusDetail:       "Failed to connect to Argo CD to fetch deployment status.",
				StatusTime:         time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: userId,
					CreatedOn: time.Now(),
					UpdatedBy: userId,
					UpdatedOn: time.Now(),
				},
			}
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, isAppStore)
			if err != nil {
				impl.Logger.Errorw("error in creating timeline status for app", "err", err, "timeline", timeline)
				return err, isTimelineUpdated
			}
		} else {
			if app == nil {
				impl.Logger.Errorw("found empty argo application object", "appName", pipeline.DeploymentAppName)
				return fmt.Errorf("found empty argo application object"), isTimelineUpdated
			}
			isSucceeded, isTimelineUpdated, err = impl.appService.UpdateDeploymentStatusForGitOpsPipelines(app, time.Now(), isAppStore)
			if err != nil {
				impl.Logger.Errorw("error in updating deployment status for gitOps cd pipelines", "app", app)
				return err, isTimelineUpdated
			}
			appStatus := app.Status.Health.Status
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(pipeline.AppId, pipeline.EnvironmentId, string(appStatus))
			if err != nil {
				impl.Logger.Errorw("error occurred while updating app-status for cd pipeline", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
				impl.Logger.Debugw("ignoring the error, UpdateStatusWithAppIdEnvId", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
			}
		}
		if isSucceeded {
			//handling deployment success event
			gitHash := ""
			if app != nil {
				gitHash = app.Status.Sync.Revision
			}
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent(gitHash, 0)
			if err != nil {
				impl.Logger.Errorw("error in handling deployment success event", "err", err)
				return err, isTimelineUpdated
			}
		}
	} else {
		isAppStore := true
		installedAppVersionHistory, err := impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistoryByInstalledAppId(installedApp.Id)
		if err != nil {
			impl.Logger.Errorw("error in getting latest installedAppVersionHistory by installedAppId", "err", err, "installedAppId", installedApp.Id)
			return nil, isTimelineUpdated
		}
		impl.Logger.Debugw("ARGO_PIPELINE_STATUS_UPDATE_REQ", "stage", "checkingDeploymentStatus", "argoAppName", installedApp, "installedAppVersionHistory", installedAppVersionHistory)
		if util3.IsTerminalStatus(installedAppVersionHistory.Status) {
			//drop event
			return nil, isTimelineUpdated
		}
		appDetails, err := impl.appRepository.FindActiveById(installedApp.AppId)
		if err != nil {
			impl.Logger.Errorw("error in getting appDetails from appId", "err", err)
			return nil, isTimelineUpdated
		}

		//TODO if Environment object in installedApp is nil then fetch envDetails also from envRepository
		envDetail, err := impl.envRepository.FindById(installedApp.EnvironmentId)
		if err != nil {
			impl.Logger.Errorw("error in getting envDetails from environment id", "err", err)
			return nil, isTimelineUpdated
		}
		var acdAppName string
		if len(installedApp.Environment.Name) != 0 {
			acdAppName = appDetails.AppName + installedApp.Environment.Name
		} else {
			acdAppName = appDetails.AppName + "-" + envDetail.Name
		}

		//this should only be called when we have git-ops configured
		//try fetching status from argo cd
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.Logger.Errorw("error in getting acd token", "err", err)
		}

		ctx := context.WithValue(context.Background(), "token", acdToken)
		query := &application2.ApplicationQuery{
			Name: &acdAppName,
		}
		app, err := impl.application.Get(ctx, query)
		if err != nil {
			impl.Logger.Errorw("error in getting acd application", "err", err, "argoAppName", installedApp)
			//updating cdWfr status
			installedAppVersionHistory.Status = pipelineConfig.WorkflowUnableToFetchState
			installedAppVersionHistory.UpdatedOn = time.Now()
			installedAppVersionHistory.UpdatedBy = 1
			installedAppVersionHistory, err = impl.installedAppVersionHistoryRepository.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
			if err != nil {
				impl.Logger.Errorw("error on update installedAppVersionHistory", "installedAppVersionHistory", installedAppVersionHistory, "err", err)
				return err, isTimelineUpdated
			}
			// creating installedApp pipeline status timeline
			timeline := &pipelineConfig.PipelineStatusTimeline{
				InstalledAppVersionHistoryId: installedAppVersionHistory.Id,
				Status:                       pipelineConfig.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS,
				StatusDetail:                 "Failed to connect to Argo CD to fetch deployment status.",
				StatusTime:                   time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: userId,
					CreatedOn: time.Now(),
					UpdatedBy: userId,
					UpdatedOn: time.Now(),
				},
			}
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, isAppStore)
			if err != nil {
				impl.Logger.Errorw("error in creating timeline status for app", "err", err, "timeline", timeline)
				return err, isTimelineUpdated
			}
		} else {
			if app == nil {
				impl.Logger.Errorw("found empty argo application object", "appName", acdAppName)
				return fmt.Errorf("found empty argo application object"), isTimelineUpdated
			}
			isSucceeded, isTimelineUpdated, err = impl.appService.UpdateDeploymentStatusForGitOpsPipelines(app, time.Now(), isAppStore)
			if err != nil {
				impl.Logger.Errorw("error in updating deployment status for gitOps cd pipelines", "app", app)
				return err, isTimelineUpdated
			}
			appStatus := app.Status.Health.Status
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(installedApp.AppId, installedApp.EnvironmentId, string(appStatus))
			if err != nil {
				impl.Logger.Errorw("error occurred while updating app-status for installed app", "err", err, "appId", installedApp.AppId, "envId", installedApp.EnvironmentId)
				impl.Logger.Debugw("ignoring the error, UpdateStatusWithAppIdEnvId", "err", err, "appId", installedApp.AppId, "envId", installedApp.EnvironmentId)
			}
		}
		if isSucceeded {
			//handling deployment success event
			//updating cdWfr status
			installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
			installedAppVersionHistory.FinishedOn = time.Now()
			installedAppVersionHistory.UpdatedOn = time.Now()
			installedAppVersionHistory.UpdatedBy = 1
			installedAppVersionHistory, err = impl.installedAppVersionHistoryRepository.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
			if err != nil {
				impl.Logger.Errorw("error on update installedAppVersionHistory", "installedAppVersionHistory", installedAppVersionHistory, "err", err)
				return err, isTimelineUpdated
			}

		}
	}

	return nil, isTimelineUpdated
}

func (impl *CdHandlerImpl) CheckHelmAppStatusPeriodicallyAndUpdateInDb(helmPipelineStatusCheckEligibleTime int) error {
	wfrList, err := impl.cdWorkflowRepository.GetLatestTriggersOfHelmPipelinesStuckInNonTerminalStatuses()
	if err != nil {
		impl.Logger.Errorw("error in getting latest triggers of helm pipelines which are stuck in non terminal statuses", "err", err)
		return err
	}
	impl.Logger.Infow("checking helm app status for non terminal deployment triggers", "wfrList", wfrList)
	for _, wfr := range wfrList {
		if time.Now().Sub(wfr.UpdatedOn) <= time.Duration(helmPipelineStatusCheckEligibleTime)*time.Second {
			//if wfr is updated within configured time then do not include for this cron cycle
			continue
		}
		appIdentifier := &client.AppIdentifier{
			ClusterId:   wfr.CdWorkflow.Pipeline.Environment.ClusterId,
			Namespace:   wfr.CdWorkflow.Pipeline.Environment.Namespace,
			ReleaseName: wfr.CdWorkflow.Pipeline.DeploymentAppName,
		}
		helmAppStatus, err := impl.helmAppService.GetApplicationStatus(context.Background(), appIdentifier)
		if err != nil {
			impl.Logger.Errorw("error in getting helm app release status ", "appIdentifier", appIdentifier, "err", err)
			//return err
			//skip this error and continue for next workflow status
			impl.Logger.Warnw("found error, skipping helm apps status update for this trigger", "appIdentifier", appIdentifier, "err", err)

			// Handle release not found errors
			if !strings.Contains(err.Error(), "release: not found") {
				continue
			}
		}
		if helmAppStatus == application.Healthy {
			wfr.Status = pipelineConfig.WorkflowSucceeded
			wfr.FinishedOn = time.Now()

		} else if err != nil && strings.Contains(err.Error(), "release: not found") {
			// If release not found, mark the deployment as failure
			wfr.Status = pipelineConfig.WorkflowFailed

		} else {
			wfr.Status = pipelineConfig.WorkflowInProgress
		}
		wfr.UpdatedBy = 1
		wfr.UpdatedOn = time.Now()
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(wfr)
		if err != nil {
			impl.Logger.Errorw("error on update cd workflow runner", "wfr", wfr, "err", err)
			return err
		}
		cdMetrics := util3.CDMetrics{
			AppName:         wfr.CdWorkflow.Pipeline.DeploymentAppName,
			Status:          wfr.Status,
			DeploymentType:  wfr.CdWorkflow.Pipeline.DeploymentAppType,
			EnvironmentName: wfr.CdWorkflow.Pipeline.Environment.Name,
			Time:            time.Since(wfr.StartedOn).Seconds() - time.Since(wfr.FinishedOn).Seconds(),
		}
		util3.TriggerCDMetrics(cdMetrics, impl.cdConfig.ExposeCDMetrics)
		impl.Logger.Infow("updated workflow runner status for helm app", "wfr", wfr)
		if helmAppStatus == application.Healthy {
			pipelineOverride, err := impl.pipelineOverrideRepository.FindLatestByCdWorkflowId(wfr.CdWorkflowId)
			if err != nil {
				impl.Logger.Errorw("error in getting latest pipeline override by cdWorkflowId", "err", err, "cdWorkflowId", wfr.CdWorkflowId)
				return err
			}
			go impl.appService.WriteCDSuccessEvent(pipelineOverride.Pipeline.AppId, pipelineOverride.Pipeline.EnvironmentId, pipelineOverride)
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent("", pipelineOverride.Id)
			if err != nil {
				impl.Logger.Errorw("error on handling deployment success event", "wfr", wfr, "err", err)
				return err
			}
		}
	}
	return nil
}

func (impl *CdHandlerImpl) CancelStage(workflowRunnerId int, userId int32) (int, error) {
	workflowRunner, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(workflowRunnerId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return 0, err
	}
	if !(string(v1alpha1.NodePending) == workflowRunner.Status || string(v1alpha1.NodeRunning) == workflowRunner.Status) {
		impl.Logger.Info("cannot cancel stage, stage not in progress")
		return 0, errors.New("cannot cancel stage, stage not in progress")
	}
	pipeline, err := impl.pipelineRepository.FindById(workflowRunner.CdWorkflow.PipelineId)
	if err != nil {
		impl.Logger.Errorw("error while fetching cd pipeline", "err", err)
		return 0, err
	}

	env, err := impl.envRepository.FindById(pipeline.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("could not fetch stage env", "err", err)
		return 0, err
	}

	serverUrl := env.Cluster.ServerUrl
	configMap := env.Cluster.Config
	bearerToken := configMap["bearer_token"]

	var isExtCluster bool
	if workflowRunner.WorkflowType == PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if workflowRunner.WorkflowType == POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}

	runningWf, err := impl.cdService.GetWorkflow(workflowRunner.Name, workflowRunner.Namespace, serverUrl, bearerToken, isExtCluster)
	if err != nil {
		impl.Logger.Errorw("cannot find workflow ", "name", workflowRunner.Name)
		return 0, errors.New("cannot find workflow " + workflowRunner.Name)
	}

	// Terminate workflow
	err = impl.cdService.TerminateWorkflow(runningWf.Name, runningWf.Namespace, serverUrl, bearerToken, isExtCluster)
	if err != nil {
		impl.Logger.Error("cannot terminate wf runner", "err", err)
		return 0, err
	}

	workflowRunner.Status = WorkflowCancel
	workflowRunner.UpdatedOn = time.Now()
	workflowRunner.UpdatedBy = userId
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(workflowRunner)
	if err != nil {
		impl.Logger.Error("cannot update deleted workflow runner status, but wf deleted", "err", err)
		return 0, err
	}
	return workflowRunner.Id, nil
}

func (impl *CdHandlerImpl) UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, string, error) {
	wfStatusRs := impl.extractWorkfowStatus(workflowStatus)
	workflowName, status, podStatus, message, podName := wfStatusRs.WorkflowName, wfStatusRs.Status, wfStatusRs.PodStatus, wfStatusRs.Message, wfStatusRs.PodName
	impl.Logger.Debugw("cd update for ", "wf ", workflowName, "status", status)
	if workflowName == "" {
		return 0, "", errors.New("invalid wf name")
	}
	workflowId, err := strconv.Atoi(workflowName[:strings.Index(workflowName, "-")])
	if err != nil {
		impl.Logger.Error("invalid wf status update req", "err", err)
		return 0, "", err
	}

	savedWorkflow, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(workflowId)
	if err != nil {
		impl.Logger.Error("cannot get saved wf", "err", err)
		return 0, "", err
	}

	ciWorkflowConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(savedWorkflow.CdWorkflow.PipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciWorkflowConfig", "err", err)
		return 0, "", err
	}

	ciArtifactLocationFormat := ciWorkflowConfig.CdArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.cdConfig.CdArtifactLocationFormat
	}

	if impl.stateChanged(status, podStatus, message, workflowStatus.FinishedAt.Time, savedWorkflow) {
		if savedWorkflow.Status != WorkflowCancel {
			savedWorkflow.Status = status
		}
		savedWorkflow.PodStatus = podStatus
		savedWorkflow.Message = message
		savedWorkflow.FinishedOn = workflowStatus.FinishedAt.Time
		savedWorkflow.Name = workflowName
		savedWorkflow.LogLocation = wfStatusRs.LogLocation
		savedWorkflow.PodName = podName
		savedWorkflow.UpdatedOn = time.Now()
		savedWorkflow.UpdatedBy = 1
		impl.Logger.Debugw("updating workflow ", "workflow", savedWorkflow)
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(savedWorkflow)
		if err != nil {
			impl.Logger.Error("update wf failed for id " + strconv.Itoa(savedWorkflow.Id))
			return 0, "", err
		}
		cdMetrics := util3.CDMetrics{
			AppName:         savedWorkflow.CdWorkflow.Pipeline.DeploymentAppName,
			Status:          savedWorkflow.Status,
			DeploymentType:  savedWorkflow.CdWorkflow.Pipeline.DeploymentAppType,
			EnvironmentName: savedWorkflow.CdWorkflow.Pipeline.Environment.Name,
			Time:            time.Since(savedWorkflow.StartedOn).Seconds() - time.Since(savedWorkflow.FinishedOn).Seconds(),
		}
		util3.TriggerCDMetrics(cdMetrics, impl.cdConfig.ExposeCDMetrics)
		if string(v1alpha1.NodeError) == savedWorkflow.Status || string(v1alpha1.NodeFailed) == savedWorkflow.Status {
			impl.Logger.Warnw("cd stage failed for workflow: ", "wfId", savedWorkflow.Id)
		}
	}
	return savedWorkflow.Id, savedWorkflow.Status, nil
}

func (impl *CdHandlerImpl) extractWorkfowStatus(workflowStatus v1alpha1.WorkflowStatus) *WorkflowStatus {
	workflowName := ""
	status := string(workflowStatus.Phase)
	podStatus := "Pending"
	message := ""
	logLocation := ""
	podName := ""
	for k, v := range workflowStatus.Nodes {
		impl.Logger.Debugw("extractWorkflowStatus", "workflowName", k, "v", v)
		if v.TemplateName == CD_WORKFLOW_NAME {
			if v.BoundaryID == "" {
				workflowName = k
			} else {
				workflowName = v.BoundaryID
			}
			podName = k
			podStatus = string(v.Phase)
			message = v.Message
			if v.Outputs != nil && len(v.Outputs.Artifacts) > 0 {
				if v.Outputs.Artifacts[0].S3 != nil {
					logLocation = v.Outputs.Artifacts[0].S3.Key
				} else if v.Outputs.Artifacts[0].GCS != nil {
					logLocation = v.Outputs.Artifacts[0].GCS.Key
				}
			}
			break
		}
	}
	workflowStatusRes := &WorkflowStatus{
		WorkflowName: workflowName,
		Status:       status,
		PodStatus:    podStatus,
		Message:      message,
		LogLocation:  logLocation,
		PodName:      podName,
	}
	return workflowStatusRes
}

type WorkflowStatus struct {
	WorkflowName, Status, PodStatus, Message, LogLocation, PodName string
}

func (impl *CdHandlerImpl) stateChanged(status string, podStatus string, msg string,
	finishedAt time.Time, savedWorkflow *pipelineConfig.CdWorkflowRunner) bool {
	return savedWorkflow.Status != status || savedWorkflow.PodStatus != podStatus || savedWorkflow.Message != msg || savedWorkflow.FinishedOn != finishedAt
}

func (impl *CdHandlerImpl) GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]pipelineConfig.CdWorkflowWithArtifact, error) {

	var cdWorkflowArtifact []pipelineConfig.CdWorkflowWithArtifact
	if pipelineId == 0 {
		wfrList, err := impl.cdWorkflowRepository.FindCdWorkflowMetaByEnvironmentId(appId, environmentId, offset, size)
		if err != nil && err != pg.ErrNoRows {
			return cdWorkflowArtifact, err
		}
		cdWorkflowArtifact = impl.converterWFRList(wfrList)
	} else {
		wfrList, err := impl.cdWorkflowRepository.FindCdWorkflowMetaByPipelineId(pipelineId, offset, size)
		if err != nil && err != pg.ErrNoRows {
			return cdWorkflowArtifact, err
		}
		cdWorkflowArtifact = impl.converterWFRList(wfrList)
	}

	return cdWorkflowArtifact, nil
}

func (impl *CdHandlerImpl) GetRunningWorkflowLogs(environmentId int, pipelineId int, wfrId int) (*bufio.Reader, func() error, error) {
	cdWorkflow, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
	if err != nil {
		impl.Logger.Errorw("error on fetch wf runner", "err", err)
		return nil, nil, err
	}

	env, err := impl.envRepository.FindById(environmentId)
	if err != nil {
		impl.Logger.Errorw("could not fetch stage env", "err", err)
		return nil, nil, err
	}

	pipeline, err := impl.pipelineRepository.FindById(cdWorkflow.CdWorkflow.PipelineId)
	if err != nil {
		impl.Logger.Errorw("error while fetching cd pipeline", "err", err)
		return nil, nil, err
	}

	serverUrl := env.Cluster.ServerUrl
	configMap := env.Cluster.Config
	bearerToken := configMap["bearer_token"]

	var isExtCluster bool
	if cdWorkflow.WorkflowType == PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if cdWorkflow.WorkflowType == POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}
	return impl.getWorkflowLogs(pipelineId, cdWorkflow, bearerToken, serverUrl, isExtCluster)
}

func (impl *CdHandlerImpl) getWorkflowLogs(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner, token string, host string, runStageInEnv bool) (*bufio.Reader, func() error, error) {
	cdLogRequest := BuildLogRequest{
		PodName:   cdWorkflow.PodName,
		Namespace: cdWorkflow.Namespace,
	}

	logStream, cleanUp, err := impl.ciLogService.FetchRunningWorkflowLogs(cdLogRequest, token, host, runStageInEnv)
	if logStream == nil || err != nil {
		if !cdWorkflow.BlobStorageEnabled {
			return nil, nil, errors.New("logs-not-stored-in-repository")
		} else if string(v1alpha1.NodeSucceeded) == cdWorkflow.Status || string(v1alpha1.NodeError) == cdWorkflow.Status || string(v1alpha1.NodeFailed) == cdWorkflow.Status || cdWorkflow.Status == WorkflowCancel {
			impl.Logger.Debugw("pod is not live ", "err", err)
			return impl.getLogsFromRepository(pipelineId, cdWorkflow)
		}
		impl.Logger.Errorw("err on fetch workflow logs", "err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(logStream)
	return logReader, cleanUp, err
}

func (impl *CdHandlerImpl) getLogsFromRepository(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner) (*bufio.Reader, func() error, error) {
	impl.Logger.Debug("getting historic logs")

	cdConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return nil, nil, err
	}

	if cdConfig.LogsBucket == "" {
		cdConfig.LogsBucket = impl.cdConfig.DefaultBuildLogsBucket //TODO -fixme
	}
	if cdConfig.CdCacheRegion == "" {
		cdConfig.CdCacheRegion = impl.cdConfig.DefaultCdLogsBucketRegion
	}

	cdLogRequest := BuildLogRequest{
		PipelineId:    cdWorkflow.CdWorkflow.PipelineId,
		WorkflowId:    cdWorkflow.Id,
		PodName:       cdWorkflow.PodName,
		LogsFilePath:  cdWorkflow.LogLocation, // impl.cdConfig.DefaultBuildLogsKeyPrefix + "/" + cdWorkflow.Name + "/main.log", //TODO - fixme
		CloudProvider: impl.ciConfig.CloudProvider,
		AzureBlobConfig: &blob_storage.AzureBlobBaseConfig{
			Enabled:           impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
			AccountName:       impl.ciConfig.AzureAccountName,
			BlobContainerName: impl.ciConfig.AzureBlobContainerCiLog,
			AccountKey:        impl.ciConfig.AzureAccountKey,
		},
		AwsS3BaseConfig: &blob_storage.AwsS3BaseConfig{
			AccessKey:         impl.ciConfig.BlobStorageS3AccessKey,
			Passkey:           impl.ciConfig.BlobStorageS3SecretKey,
			EndpointUrl:       impl.ciConfig.BlobStorageS3Endpoint,
			IsInSecure:        impl.ciConfig.BlobStorageS3EndpointInsecure,
			BucketName:        cdConfig.LogsBucket,
			Region:            cdConfig.CdCacheRegion,
			VersioningEnabled: impl.ciConfig.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             cdConfig.LogsBucket,
			CredentialFileJsonData: impl.ciConfig.BlobStorageGcpCredentialJson,
		},
	}
	impl.Logger.Infow("s3 log req ", "req", cdLogRequest)
	oldLogsStream, cleanUp, err := impl.ciLogService.FetchLogs(impl.ciConfig.BaseLogLocationPath, cdLogRequest)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(oldLogsStream)
	return logReader, cleanUp, err
}

func (impl *CdHandlerImpl) FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int) (WorkflowResponse, error) {
	workflowR, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	} else if err == pg.ErrNoRows {
		return WorkflowResponse{}, nil
	}
	workflow := impl.converterWFR(*workflowR)

	triggeredByUser, err := impl.userService.GetById(workflow.TriggeredBy)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	}
	if triggeredByUser == nil {
		triggeredByUser = &bean.UserInfo{EmailId: "anonymous"}
	}
	ciArtifactId := workflow.CiArtifactId
	if ciArtifactId > 0 {
		ciArtifact, err := impl.ciArtifactRepository.Get(ciArtifactId)
		if err != nil {
			impl.Logger.Errorw("error fetching artifact data", "err", err)
			return WorkflowResponse{}, err
		}

		// handling linked ci pipeline
		if ciArtifact.ParentCiArtifact > 0 && ciArtifact.WorkflowId == nil {
			ciArtifactId = ciArtifact.ParentCiArtifact
		}
	}
	ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifactId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching ci wf", "artifactId", workflow.CiArtifactId, "err", err)
		return WorkflowResponse{}, err
	}
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdForRegexAndFixed(ciWf.CiPipelineId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	}

	var ciMaterialsArr []CiPipelineMaterialResponse
	for _, m := range ciMaterials {
		res := CiPipelineMaterialResponse{
			Id:              m.Id,
			GitMaterialId:   m.GitMaterialId,
			GitMaterialName: m.GitMaterial.Name[strings.Index(m.GitMaterial.Name, "-")+1:],
			Type:            string(m.Type),
			Value:           m.Value,
			Active:          m.Active,
			Url:             m.GitMaterial.Url,
		}
		ciMaterialsArr = append(ciMaterialsArr, res)
	}
	gitTriggers := make(map[int]pipelineConfig.GitCommit)
	if ciWf.GitTriggers != nil {
		gitTriggers = ciWf.GitTriggers
	}

	workflowResponse := WorkflowResponse{
		Id:                 workflow.Id,
		Name:               workflow.Name,
		Status:             workflow.Status,
		PodStatus:          workflow.PodStatus,
		Message:            workflow.Message,
		StartedOn:          workflow.StartedOn,
		FinishedOn:         workflow.FinishedOn,
		Namespace:          workflow.Namespace,
		CiMaterials:        ciMaterialsArr,
		TriggeredBy:        workflow.TriggeredBy,
		TriggeredByEmail:   triggeredByUser.EmailId,
		Artifact:           workflow.Image,
		Stage:              workflow.WorkflowType,
		GitTriggers:        gitTriggers,
		BlobStorageEnabled: workflow.BlobStorageEnabled,
	}
	return workflowResponse, nil

}

func (impl *CdHandlerImpl) DownloadCdWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error) {
	wfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil {
		impl.Logger.Errorw("unable to fetch ciWorkflow", "err", err)
		return nil, err
	}

	if !wfr.BlobStorageEnabled {
		return nil, errors.New("logs-not-stored-in-repository")
	}

	cdConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciConfig", "err", err)
		return nil, err
	}

	if cdConfig.LogsBucket == "" {
		cdConfig.LogsBucket = impl.cdConfig.DefaultBuildLogsBucket
	}
	if cdConfig.CdCacheRegion == "" {
		cdConfig.CdCacheRegion = impl.cdConfig.DefaultCdLogsBucketRegion
	}

	item := strconv.Itoa(wfr.Id)
	awsS3BaseConfig := &blob_storage.AwsS3BaseConfig{
		AccessKey:         impl.ciConfig.BlobStorageS3AccessKey,
		Passkey:           impl.ciConfig.BlobStorageS3SecretKey,
		EndpointUrl:       impl.ciConfig.BlobStorageS3Endpoint,
		IsInSecure:        impl.ciConfig.BlobStorageS3EndpointInsecure,
		BucketName:        cdConfig.LogsBucket,
		Region:            cdConfig.CdCacheRegion,
		VersioningEnabled: impl.ciConfig.BlobStorageS3BucketVersioned,
	}
	azureBlobBaseConfig := &blob_storage.AzureBlobBaseConfig{
		Enabled:           impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
		AccountKey:        impl.ciConfig.AzureAccountKey,
		AccountName:       impl.ciConfig.AzureAccountName,
		BlobContainerName: impl.ciConfig.AzureBlobContainerCiLog,
	}
	gcpBlobBaseConfig := &blob_storage.GcpBlobBaseConfig{
		BucketName:             cdConfig.LogsBucket,
		CredentialFileJsonData: impl.ciConfig.BlobStorageGcpCredentialJson,
	}
	key := fmt.Sprintf("%s/"+impl.cdConfig.CdArtifactLocationFormat, impl.cdConfig.DefaultArtifactKeyPrefix, wfr.CdWorkflow.Id, wfr.Id)
	baseLogLocationPathConfig := impl.cdConfig.BaseLogLocationPath
	blobStorageService := blob_storage.NewBlobStorageServiceImpl(nil)
	destinationKey := filepath.Clean(filepath.Join(baseLogLocationPathConfig, item))
	request := &blob_storage.BlobStorageRequest{
		StorageType:         impl.ciConfig.CloudProvider,
		SourceKey:           key,
		DestinationKey:      destinationKey,
		AzureBlobBaseConfig: azureBlobBaseConfig,
		AwsS3BaseConfig:     awsS3BaseConfig,
		GcpBlobBaseConfig:   gcpBlobBaseConfig,
	}
	_, numBytes, err := blobStorageService.Get(request)
	if err != nil {
		impl.Logger.Errorw("error occurred while downloading file", "request", request, "error", err)
		return nil, errors.New("failed to download resource")
	}

	file, err := os.Open(destinationKey)
	if err != nil {
		impl.Logger.Errorw("unable to open file", "file", item, "err", err)
		return nil, errors.New("unable to open file")
	}

	impl.Logger.Infow("Downloaded ", "name", file.Name(), "bytes", numBytes)
	return file, nil
}

func (impl *CdHandlerImpl) converterWFR(wfr pipelineConfig.CdWorkflowRunner) pipelineConfig.CdWorkflowWithArtifact {
	workflow := pipelineConfig.CdWorkflowWithArtifact{}
	if wfr.Id > 0 {
		workflow.Name = wfr.Name
		workflow.Id = wfr.Id
		workflow.Namespace = wfr.Namespace
		workflow.Status = wfr.Status
		workflow.Message = wfr.Message
		workflow.PodStatus = wfr.PodStatus
		workflow.FinishedOn = wfr.FinishedOn
		workflow.TriggeredBy = wfr.TriggeredBy
		workflow.StartedOn = wfr.StartedOn
		workflow.WorkflowType = string(wfr.WorkflowType)
		workflow.CdWorkflowId = wfr.CdWorkflowId
		workflow.Image = wfr.CdWorkflow.CiArtifact.Image
		workflow.PipelineId = wfr.CdWorkflow.PipelineId
		workflow.CiArtifactId = wfr.CdWorkflow.CiArtifactId
		workflow.BlobStorageEnabled = wfr.BlobStorageEnabled

	}
	return workflow
}

func (impl *CdHandlerImpl) converterWFRList(wfrList []pipelineConfig.CdWorkflowRunner) []pipelineConfig.CdWorkflowWithArtifact {
	var workflowList []pipelineConfig.CdWorkflowWithArtifact
	var results []pipelineConfig.CdWorkflowWithArtifact
	var ids []int32
	for _, item := range wfrList {
		ids = append(ids, item.TriggeredBy)
		workflowList = append(workflowList, impl.converterWFR(item))
	}
	userEmails := make(map[int32]string)
	users, err := impl.userService.GetByIds(ids)
	if err != nil {
		impl.Logger.Errorw("unable to find user", "err", err)
	}
	for _, item := range users {
		userEmails[item.Id] = item.EmailId
	}
	for _, item := range workflowList {
		item.EmailId = userEmails[item.TriggeredBy]
		results = append(results, item)
	}
	return results
}

func (impl *CdHandlerImpl) FetchCdPrePostStageStatus(pipelineId int) ([]pipelineConfig.CdWorkflowWithArtifact, error) {
	var results []pipelineConfig.CdWorkflowWithArtifact
	wfrPre, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_PRE)
	if err != nil && err != pg.ErrNoRows {
		return results, err
	}
	if wfrPre.Id > 0 {
		workflowPre := impl.converterWFR(wfrPre)
		results = append(results, workflowPre)
	} else {
		workflowPre := pipelineConfig.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_PRE), PipelineId: pipelineId}
		results = append(results, workflowPre)
	}

	wfrPost, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_POST)
	if err != nil && err != pg.ErrNoRows {
		return results, err
	}
	if wfrPost.Id > 0 {
		workflowPost := impl.converterWFR(wfrPost)
		results = append(results, workflowPost)
	} else {
		workflowPost := pipelineConfig.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_POST), PipelineId: pipelineId}
		results = append(results, workflowPost)
	}
	return results, nil

}

func (impl *CdHandlerImpl) FetchAppWorkflowStatusForTriggerView(appId int) ([]*pipelineConfig.CdWorkflowStatus, error) {
	var cdWorkflowStatus []*pipelineConfig.CdWorkflowStatus

	pipelines, err := impl.pipelineRepository.FindActiveByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		return cdWorkflowStatus, err
	}
	pipelineIds := make([]int, 0)
	partialDeletedPipelines := make(map[int]bool)
	//pipelineIdsMap := make(map[int]int)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
		partialDeletedPipelines[pipeline.Id] = pipeline.DeploymentAppDeleteRequest
	}

	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}

	cdMap := make(map[int]*pipelineConfig.CdWorkflowStatus)
	result, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelineIds)
	if err != nil {
		return cdWorkflowStatus, err
	}
	var wfrIds []int
	for _, item := range result {
		wfrIds = append(wfrIds, item.WfrId)
	}

	statusMap := make(map[int]string)
	if len(wfrIds) > 0 {
		wfrList, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntityStatus(wfrIds)
		if err != nil && !util.IsErrNoRows(err) {
			return cdWorkflowStatus, err
		}
		for _, item := range wfrList {
			statusMap[item.Id] = item.Status
		}
	}

	for _, item := range result {
		if _, ok := cdMap[item.PipelineId]; !ok {
			cdWorkflowStatus := &pipelineConfig.CdWorkflowStatus{}
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		} else {
			cdWorkflowStatus := cdMap[item.PipelineId]
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		}
		cdMap[item.PipelineId].DeploymentAppDeleteRequest = partialDeletedPipelines[item.PipelineId]
	}

	for _, item := range cdMap {
		if item.PreStatus == "" {
			item.PreStatus = NotTriggered
		}
		if item.DeployStatus == "" {
			item.DeployStatus = NotDeployed
		}
		if item.PostStatus == "" {
			item.PostStatus = NotTriggered
		}
		cdWorkflowStatus = append(cdWorkflowStatus, item)
	}

	if len(cdWorkflowStatus) == 0 {
		for _, item := range pipelineIds {
			cdWs := &pipelineConfig.CdWorkflowStatus{}
			cdWs.PipelineId = item
			cdWs.PreStatus = NotTriggered
			cdWs.DeployStatus = NotDeployed
			cdWs.PostStatus = NotTriggered
			cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
		}
	} else {
		for _, item := range pipelineIds {
			if _, ok := cdMap[item]; !ok {
				cdWs := &pipelineConfig.CdWorkflowStatus{}
				cdWs.PipelineId = item
				cdWs.PreStatus = NotTriggered
				cdWs.DeployStatus = NotDeployed
				cdWs.PostStatus = NotTriggered
				cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
			}
		}
	}

	return cdWorkflowStatus, err
}

func (impl *CdHandlerImpl) FetchAppWorkflowStatusForTriggerViewForEnvironment(request appGroup2.AppGroupingRequest) ([]*pipelineConfig.CdWorkflowStatus, error) {
	cdWorkflowStatus := make([]*pipelineConfig.CdWorkflowStatus, 0)
	var pipelines []*pipelineConfig.Pipeline
	var err error
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	if len(request.AppIds) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}

	var appIds []int
	for _, pipeline := range pipelines {
		appIds = append(appIds, pipeline.AppId)
	}
	if len(appIds) == 0 {
		impl.Logger.Warnw("there is no app id found for fetching cd pipelines", "request", request)
		return cdWorkflowStatus, nil
	}
	pipelines, err = impl.pipelineRepository.FindActiveByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		return cdWorkflowStatus, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	for _, pipeline := range pipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	//authorization block ends here
	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}
	cdMap := make(map[int]*pipelineConfig.CdWorkflowStatus)
	wfrStatus, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelineIds)
	if err != nil {
		return cdWorkflowStatus, err
	}
	var wfrIds []int
	for _, item := range wfrStatus {
		wfrIds = append(wfrIds, item.WfrId)
	}

	statusMap := make(map[int]string)
	if len(wfrIds) > 0 {
		cdWorkflowRunners, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntityStatus(wfrIds)
		if err != nil && !util.IsErrNoRows(err) {
			return cdWorkflowStatus, err
		}
		for _, item := range cdWorkflowRunners {
			statusMap[item.Id] = item.Status
		}
	}

	for _, item := range wfrStatus {
		if _, ok := cdMap[item.PipelineId]; !ok {
			cdWorkflowStatus := &pipelineConfig.CdWorkflowStatus{}
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		} else {
			cdWorkflowStatus := cdMap[item.PipelineId]
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		}
	}

	for _, item := range cdMap {
		if item.PreStatus == "" {
			item.PreStatus = NotTriggered
		}
		if item.DeployStatus == "" {
			item.DeployStatus = NotDeployed
		}
		if item.PostStatus == "" {
			item.PostStatus = NotTriggered
		}
		cdWorkflowStatus = append(cdWorkflowStatus, item)
	}

	if len(cdWorkflowStatus) == 0 {
		for _, item := range pipelineIds {
			cdWs := &pipelineConfig.CdWorkflowStatus{}
			cdWs.PipelineId = item
			cdWs.PreStatus = NotTriggered
			cdWs.DeployStatus = NotDeployed
			cdWs.PostStatus = NotTriggered
			cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
		}
	} else {
		for _, item := range pipelineIds {
			if _, ok := cdMap[item]; !ok {
				cdWs := &pipelineConfig.CdWorkflowStatus{}
				cdWs.PipelineId = item
				cdWs.PreStatus = NotTriggered
				cdWs.DeployStatus = NotDeployed
				cdWs.PostStatus = NotTriggered
				cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
			}
		}
	}

	return cdWorkflowStatus, err
}

func (impl *CdHandlerImpl) FetchAppDeploymentStatusForEnvironments(request appGroup2.AppGroupingRequest) ([]*pipelineConfig.AppDeploymentStatus, error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "pipelineBuilder.authorizationDeploymentStatusForAppGrouping")
	deploymentStatuses := make([]*pipelineConfig.AppDeploymentStatus, 0)
	deploymentStatusesMap := make(map[int]*pipelineConfig.AppDeploymentStatus)
	pipelineAppMap := make(map[int]int)
	statusMap := make(map[int]string)
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	if len(request.AppIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return nil, err
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pipelineIds = append(pipelineIds, pipeline.Id)
		pipelineAppMap[pipeline.Id] = pipeline.AppId
	}
	span.End()
	//authorization block ends here

	if len(pipelineIds) == 0 {
		return deploymentStatuses, nil
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "pipelineBuilder.FetchAllCdStagesLatestEntity")
	result, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelineIds)
	span.End()
	if err != nil {
		return deploymentStatuses, err
	}
	var wfrIds []int
	for _, item := range result {
		wfrIds = append(wfrIds, item.WfrId)
	}
	if len(wfrIds) > 0 {
		_, span = otel.Tracer("orchestrator").Start(request.Ctx, "pipelineBuilder.FetchAllCdStagesLatestEntityStatus")
		wfrList, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntityStatus(wfrIds)
		span.End()
		if err != nil && !util.IsErrNoRows(err) {
			return deploymentStatuses, err
		}
		for _, item := range wfrList {
			if item.Status == "" {
				statusMap[item.Id] = NotDeployed
			} else {
				statusMap[item.Id] = item.Status
			}
		}
	}

	for _, item := range result {
		if _, ok := deploymentStatusesMap[item.PipelineId]; !ok {
			deploymentStatus := &pipelineConfig.AppDeploymentStatus{}
			deploymentStatus.PipelineId = item.PipelineId
			if item.WorkflowType == WorklowTypeDeploy {
				deploymentStatus.DeployStatus = statusMap[item.WfrId]
				deploymentStatus.AppId = pipelineAppMap[deploymentStatus.PipelineId]
				deploymentStatusesMap[item.PipelineId] = deploymentStatus
			}
		}
	}
	//in case there is no workflow found for pipeline, set all the pipeline status - Not Deployed
	for _, pipelineId := range pipelineIds {
		if _, ok := deploymentStatusesMap[pipelineId]; !ok {
			deploymentStatus := &pipelineConfig.AppDeploymentStatus{}
			deploymentStatus.PipelineId = pipelineId
			deploymentStatus.DeployStatus = NotDeployed
			deploymentStatus.AppId = pipelineAppMap[deploymentStatus.PipelineId]
			deploymentStatusesMap[pipelineId] = deploymentStatus
		}
	}
	for _, deploymentStatus := range deploymentStatusesMap {
		deploymentStatuses = append(deploymentStatuses, deploymentStatus)
	}

	return deploymentStatuses, err
}
