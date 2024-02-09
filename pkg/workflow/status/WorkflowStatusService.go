package status

import (
	"context"
	"fmt"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	pubub "github.com/devtron-labs/common-lib/pubsub-lib"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	client2 "github.com/devtron-labs/devtron/client/events"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/app/status"
	app_status "github.com/devtron-labs/devtron/pkg/appStatus"
	repository3 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	"github.com/devtron-labs/devtron/pkg/workflow/status/bean"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/utils/strings/slices"
	"time"
)

type WorkflowStatusService interface {
	CheckHelmAppStatusPeriodicallyAndUpdateInDb(helmPipelineStatusCheckEligibleTime int,
		getPipelineDeployedWithinHours int) error

	UpdatePipelineTimelineAndStatusByLiveApplicationFetch(triggerContext bean3.TriggerContext,
		pipeline *pipelineConfig.Pipeline, installedApp repository3.InstalledApps, userId int32) (error, bool)

	CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipelineId int, installedAppVersionId int,
		userId int32, isAppStoreApplication bool)

	CheckArgoPipelineTimelineStatusPeriodicallyAndUpdateInDb(pendingSinceSeconds int, timeForDegradation int) error

	CheckArgoAppStatusPeriodicallyAndUpdateInDb(getPipelineDeployedBeforeMinutes int, getPipelineDeployedWithinHours int) error
}

type WorkflowStatusServiceImpl struct {
	logger                          *zap.SugaredLogger
	workflowDagExecutor             dag.WorkflowDagExecutor
	pipelineStatusTimelineService   status.PipelineStatusTimelineService
	appService                      app.AppService
	config                          *types.CdConfig
	appStatusService                app_status.AppStatusService
	acdConfig                       *argocdServer.ACDConfig
	AppConfig                       *app.AppServiceConfig
	argoUserService                 argo.ArgoUserService
	pipelineStatusSyncDetailService status.PipelineStatusSyncDetailService
	argocdClientWrapperService      argocdServer.ArgoClientWrapperService

	cdWorkflowRepository                 pipelineConfig.CdWorkflowRepository
	pipelineOverrideRepository           chartConfig.PipelineOverrideRepository
	installedAppVersionHistoryRepository repository3.InstalledAppVersionHistoryRepository
	appRepository                        appRepository.AppRepository
	envRepository                        repository2.EnvironmentRepository
	installedAppRepository               repository3.InstalledAppRepository
	pipelineStatusTimelineRepository     pipelineConfig.PipelineStatusTimelineRepository
	pipelineRepository                   pipelineConfig.PipelineRepository

	application application.ServiceClient
	eventClient client2.EventClient
}

func NewWorkflowStatusServiceImpl(logger *zap.SugaredLogger,
	workflowDagExecutor dag.WorkflowDagExecutor,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	appService app.AppService, appStatusService app_status.AppStatusService,
	acdConfig *argocdServer.ACDConfig, AppConfig *app.AppServiceConfig,
	argoUserService argo.ArgoUserService,
	pipelineStatusSyncDetailService status.PipelineStatusSyncDetailService,
	argocdClientWrapperService argocdServer.ArgoClientWrapperService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	installedAppVersionHistoryRepository repository3.InstalledAppVersionHistoryRepository,
	appRepository appRepository.AppRepository,
	envRepository repository2.EnvironmentRepository,
	installedAppRepository repository3.InstalledAppRepository,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	application application.ServiceClient,
	eventClient client2.EventClient) (*WorkflowStatusServiceImpl, error) {
	impl := &WorkflowStatusServiceImpl{
		logger:                               logger,
		workflowDagExecutor:                  workflowDagExecutor,
		pipelineStatusTimelineService:        pipelineStatusTimelineService,
		appService:                           appService,
		appStatusService:                     appStatusService,
		acdConfig:                            acdConfig,
		AppConfig:                            AppConfig,
		argoUserService:                      argoUserService,
		pipelineStatusSyncDetailService:      pipelineStatusSyncDetailService,
		argocdClientWrapperService:           argocdClientWrapperService,
		cdWorkflowRepository:                 cdWorkflowRepository,
		pipelineOverrideRepository:           pipelineOverrideRepository,
		installedAppVersionHistoryRepository: installedAppVersionHistoryRepository,
		appRepository:                        appRepository,
		envRepository:                        envRepository,
		installedAppRepository:               installedAppRepository,
		pipelineStatusTimelineRepository:     pipelineStatusTimelineRepository,
		pipelineRepository:                   pipelineRepository,
		application:                          application,
		eventClient:                          eventClient,
	}
	config, err := types.GetCdConfig()
	if err != nil {
		return nil, err
	}
	impl.config = config
	return impl, nil
}

func (impl *WorkflowStatusServiceImpl) CheckHelmAppStatusPeriodicallyAndUpdateInDb(helmPipelineStatusCheckEligibleTime int,
	getPipelineDeployedWithinHours int) error {
	wfrList, err := impl.cdWorkflowRepository.GetLatestTriggersOfHelmPipelinesStuckInNonTerminalStatuses(getPipelineDeployedWithinHours)
	if err != nil {
		impl.logger.Errorw("error in getting latest triggers of helm pipelines which are stuck in non terminal statuses", "err", err)
		return err
	}
	impl.logger.Debugw("checking helm app status for non terminal deployment triggers", "wfrList", wfrList, "number of wfr", len(wfrList))
	for _, wfr := range wfrList {
		if time.Now().Sub(wfr.StartedOn) <= time.Duration(helmPipelineStatusCheckEligibleTime)*time.Second {
			// if wfr is updated within configured time then do not include for this cron cycle
			continue
		}
		appIdentifier := &client.AppIdentifier{
			ClusterId:   wfr.CdWorkflow.Pipeline.Environment.ClusterId,
			Namespace:   wfr.CdWorkflow.Pipeline.Environment.Namespace,
			ReleaseName: wfr.CdWorkflow.Pipeline.DeploymentAppName,
		}
		if isWfrUpdated := impl.workflowDagExecutor.UpdateWorkflowRunnerStatusForDeployment(appIdentifier, wfr, true); !isWfrUpdated {
			continue
		}
		wfr.UpdatedBy = 1
		wfr.UpdatedOn = time.Now()
		if wfr.Status == pipelineConfig.WorkflowFailed {
			err = impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(wfr.RefCdWorkflowRunnerId, pipelineConfig.NEW_DEPLOYMENT_INITIATED)
			if err != nil {
				impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", err)
				return err
			}
		}
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(wfr)
		if err != nil {
			impl.logger.Errorw("error on update cd workflow runner", "wfr", wfr, "err", err)
			return err
		}
		if slices.Contains(pipelineConfig.WfrTerminalStatusList, wfr.Status) {
			util3.TriggerCDMetrics(pipelineConfig.GetTriggerMetricsFromRunnerObj(wfr), impl.config.ExposeCDMetrics)
		}

		impl.logger.Infow("updated workflow runner status for helm app", "wfr", wfr)
		if wfr.Status == pipelineConfig.WorkflowSucceeded {
			pipelineOverride, err := impl.pipelineOverrideRepository.FindLatestByCdWorkflowId(wfr.CdWorkflowId)
			if err != nil {
				impl.logger.Errorw("error in getting latest pipeline override by cdWorkflowId", "err", err, "cdWorkflowId", wfr.CdWorkflowId)
				return err
			}
			go impl.appService.WriteCDSuccessEvent(pipelineOverride.Pipeline.AppId, pipelineOverride.Pipeline.EnvironmentId, wfr, pipelineOverride)
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent(bean3.TriggerContext{}, pipelineOverride)
			if err != nil {
				impl.logger.Errorw("error on handling deployment success event", "wfr", wfr, "err", err)
				return err
			}
		}
	}
	return nil
}

func (impl *WorkflowStatusServiceImpl) UpdatePipelineTimelineAndStatusByLiveApplicationFetch(triggerContext bean3.TriggerContext,
	pipeline *pipelineConfig.Pipeline, installedApp repository3.InstalledApps, userId int32) (error, bool) {
	isTimelineUpdated := false
	isSucceeded := false
	var pipelineOverride *chartConfig.PipelineOverride
	if pipeline != nil {
		isAppStore := false
		cdWfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipeline.Id, bean2.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil {
			impl.logger.Errorw("error in getting latest cdWfr by cdPipelineId", "err", err, "pipelineId", pipeline.Id)
			return nil, isTimelineUpdated
		}
		impl.logger.Debugw("ARGO_PIPELINE_STATUS_UPDATE_REQ", "stage", "checkingDeploymentStatus", "argoAppName", pipeline, "cdWfr", cdWfr)
		if util3.IsTerminalStatus(cdWfr.Status) {
			// drop event
			return nil, isTimelineUpdated
		}

		if !impl.acdConfig.ArgoCDAutoSyncEnabled {
			// if manual sync check for application sync status
			isArgoAppSynced := impl.pipelineStatusTimelineService.GetArgoAppSyncStatus(cdWfr.Id)
			if !isArgoAppSynced {
				return nil, isTimelineUpdated
			}
		}
		// this should only be called when we have git-ops configured
		// try fetching status from argo cd
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
		}
		ctx := context.WithValue(context.Background(), "token", acdToken)
		query := &application2.ApplicationQuery{
			Name: &pipeline.DeploymentAppName,
		}
		app, err := impl.application.Get(ctx, query)
		if err != nil {
			impl.logger.Errorw("error in getting acd application", "err", err, "argoAppName", pipeline)
			// updating cdWfr status
			cdWfr.Status = pipelineConfig.WorkflowUnableToFetchState
			cdWfr.UpdatedOn = time.Now()
			cdWfr.UpdatedBy = 1
			err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&cdWfr)
			if err != nil {
				impl.logger.Errorw("error on update cd workflow runner", "cdWfr", cdWfr, "err", err)
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
				impl.logger.Errorw("error in creating timeline status for app", "err", err, "timeline", timeline)
				return err, isTimelineUpdated
			}
		} else {
			if app == nil {
				impl.logger.Errorw("found empty argo application object", "appName", pipeline.DeploymentAppName)
				return fmt.Errorf("found empty argo application object"), isTimelineUpdated
			}
			isSucceeded, isTimelineUpdated, pipelineOverride, err = impl.appService.UpdateDeploymentStatusForGitOpsPipelines(app, time.Now(), isAppStore)
			if err != nil {
				impl.logger.Errorw("error in updating deployment status for gitOps cd pipelines", "app", app)
				return err, isTimelineUpdated
			}
			appStatus := app.Status.Health.Status
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(pipeline.AppId, pipeline.EnvironmentId, string(appStatus))
			if err != nil {
				impl.logger.Errorw("error occurred while updating app-status for cd pipeline", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
				impl.logger.Debugw("ignoring the error, UpdateStatusWithAppIdEnvId", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
			}
		}
		if isSucceeded {
			// handling deployment success event
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent(triggerContext, pipelineOverride)
			if err != nil {
				impl.logger.Errorw("error in handling deployment success event", "pipelineOverride", pipelineOverride, "err", err)
				return err, isTimelineUpdated
			}
		}
	} else {
		isAppStore := true
		installedAppVersionHistory, err := impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistoryByInstalledAppId(installedApp.Id)
		if err != nil {
			impl.logger.Errorw("error in getting latest installedAppVersionHistory by installedAppId", "err", err, "installedAppId", installedApp.Id)
			return nil, isTimelineUpdated
		}
		impl.logger.Debugw("ARGO_PIPELINE_STATUS_UPDATE_REQ", "stage", "checkingDeploymentStatus", "argoAppName", installedApp, "installedAppVersionHistory", installedAppVersionHistory)
		if util3.IsTerminalStatus(installedAppVersionHistory.Status) {
			// drop event
			return nil, isTimelineUpdated
		}
		if !impl.acdConfig.ArgoCDAutoSyncEnabled {
			isArgoAppSynced := impl.pipelineStatusTimelineService.GetArgoAppSyncStatusForAppStore(installedAppVersionHistory.Id)
			if !isArgoAppSynced {
				return nil, isTimelineUpdated
			}
		}
		appDetails, err := impl.appRepository.FindActiveById(installedApp.AppId)
		if err != nil {
			impl.logger.Errorw("error in getting appDetails from appId", "err", err)
			return nil, isTimelineUpdated
		}
		// TODO if Environment object in installedApp is nil then fetch envDetails also from envRepository
		envDetail, err := impl.envRepository.FindById(installedApp.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error in getting envDetails from environment id", "err", err)
			return nil, isTimelineUpdated
		}
		var acdAppName string
		if len(installedApp.Environment.Name) != 0 {
			acdAppName = appDetails.AppName + installedApp.Environment.Name
		} else {
			acdAppName = appDetails.AppName + "-" + envDetail.Name
		}

		// this should only be called when we have git-ops configured
		// try fetching status from argo cd
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
		}

		ctx := context.WithValue(context.Background(), "token", acdToken)
		query := &application2.ApplicationQuery{
			Name: &acdAppName,
		}
		app, err := impl.application.Get(ctx, query)
		if err != nil {
			impl.logger.Errorw("error in getting acd application", "err", err, "argoAppName", installedApp)
			// updating cdWfr status
			installedAppVersionHistory.Status = pipelineConfig.WorkflowUnableToFetchState
			installedAppVersionHistory.UpdatedOn = time.Now()
			installedAppVersionHistory.UpdatedBy = 1
			installedAppVersionHistory, err = impl.installedAppVersionHistoryRepository.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
			if err != nil {
				impl.logger.Errorw("error on update installedAppVersionHistory", "installedAppVersionHistory", installedAppVersionHistory, "err", err)
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
				impl.logger.Errorw("error in creating timeline status for app", "err", err, "timeline", timeline)
				return err, isTimelineUpdated
			}
		} else {
			if app == nil {
				impl.logger.Errorw("found empty argo application object", "appName", acdAppName)
				return fmt.Errorf("found empty argo application object"), isTimelineUpdated
			}
			isSucceeded, isTimelineUpdated, pipelineOverride, err = impl.appService.UpdateDeploymentStatusForGitOpsPipelines(app, time.Now(), isAppStore)
			if err != nil {
				impl.logger.Errorw("error in updating deployment status for gitOps cd pipelines", "app", app)
				return err, isTimelineUpdated
			}
			appStatus := app.Status.Health.Status
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(installedApp.AppId, installedApp.EnvironmentId, string(appStatus))
			if err != nil {
				impl.logger.Errorw("error occurred while updating app-status for installed app", "err", err, "appId", installedApp.AppId, "envId", installedApp.EnvironmentId)
				impl.logger.Debugw("ignoring the error, UpdateStatusWithAppIdEnvId", "err", err, "appId", installedApp.AppId, "envId", installedApp.EnvironmentId)
			}
		}
		if isSucceeded {
			// handling deployment success event
			// updating cdWfr status
			installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
			installedAppVersionHistory.FinishedOn = time.Now()
			installedAppVersionHistory.UpdatedOn = time.Now()
			installedAppVersionHistory.UpdatedBy = 1
			installedAppVersionHistory, err = impl.installedAppVersionHistoryRepository.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
			if err != nil {
				impl.logger.Errorw("error on update installedAppVersionHistory", "installedAppVersionHistory", installedAppVersionHistory, "err", err)
				return err, isTimelineUpdated
			}

		}
	}

	return nil, isTimelineUpdated
}

func (impl *WorkflowStatusServiceImpl) CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipelineId int, installedAppVersionId int,
	userId int32, isAppStoreApplication bool) {
	var lastSyncTime time.Time
	var err error
	if isAppStoreApplication {
		lastSyncTime, err = impl.pipelineStatusSyncDetailService.GetLastSyncTimeForLatestInstalledAppVersionHistoryByInstalledAppVersionId(installedAppVersionId)
	} else {
		lastSyncTime, err = impl.pipelineStatusSyncDetailService.GetLastSyncTimeForLatestCdWfrByCdPipelineId(pipelineId)
	}
	if err != nil {
		impl.logger.Errorw("error in getting last sync time by pipelineId", "err", err, "pipelineId", pipelineId, "installedAppVersionHistoryId", installedAppVersionId)
		return
	}

	// sync argocd app
	if pipelineId != 0 {
		err := impl.syncACDDevtronApps(impl.AppConfig.ArgocdManualSyncCronPipelineDeployedBefore, pipelineId)
		if err != nil {
			impl.logger.Errorw("error in syncing devtron apps deployed via argoCD", "err", err)
			return
		}
	}
	if installedAppVersionId != 0 {
		err := impl.syncACDHelmApps(impl.AppConfig.ArgocdManualSyncCronPipelineDeployedBefore, installedAppVersionId)
		if err != nil {
			impl.logger.Errorw("error in syncing Helm apps deployed via argoCD", "err", err)
			return
		}
	}

	// pipelineId can be cdPipelineId or installedAppVersionId, using isAppStoreApplication flag to identify between them
	if lastSyncTime.IsZero() || (!lastSyncTime.IsZero() && time.Since(lastSyncTime) > 5*time.Second) { // create new nats event
		statusUpdateEvent := bean.ArgoPipelineStatusSyncEvent{
			PipelineId:            pipelineId,
			InstalledAppVersionId: installedAppVersionId,
			UserId:                userId,
			IsAppStoreApplication: isAppStoreApplication,
		}
		// write event
		err = impl.eventClient.WriteNatsEvent(pubub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, statusUpdateEvent)
		if err != nil {
			impl.logger.Errorw("error in writing nats event", "topic", pubub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, "payload", statusUpdateEvent)
		}
	}
}

func (impl *WorkflowStatusServiceImpl) CheckArgoAppStatusPeriodicallyAndUpdateInDb(getPipelineDeployedBeforeMinutes int, getPipelineDeployedWithinHours int) error {
	pipelines, err := impl.pipelineRepository.GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatuses(getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours)
	if err != nil {
		impl.logger.Errorw("error in getting pipelines having latest trigger stuck in non terminal statuses", "err", err)
		return err
	}
	impl.logger.Debugw("received stuck argo cd pipelines", "pipelines", pipelines, "number of pipelines", len(pipelines))

	for _, pipeline := range pipelines {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipeline.Id, 0, 1, false)
	}

	installedAppVersions, err := impl.installedAppRepository.GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatusesForAppStore(getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours)
	if err != nil {
		impl.logger.Errorw("error in getting installedAppVersions having latest trigger stuck in non terminal statuses", "err", err)
		return err
	}
	impl.logger.Debugw("received stuck argo installed appStore app", "installedAppVersions", installedAppVersions, "number of triggers", len(installedAppVersions))

	for _, installedAppVersion := range installedAppVersions {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(0, installedAppVersion.Id, 1, true)
	}
	return nil
}

func (impl *WorkflowStatusServiceImpl) CheckArgoPipelineTimelineStatusPeriodicallyAndUpdateInDb(pendingSinceSeconds int, timeForDegradation int) error {
	// getting all the progressing status that are stuck since some time after kubectl apply success sync stage
	// and are not eligible for CheckArgoAppStatusPeriodicallyAndUpdateInDb
	pipelines, err := impl.pipelineRepository.GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelines(pendingSinceSeconds, timeForDegradation)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelines", "err", err)
		return err
	}
	impl.logger.Debugw("received argo cd pipelines stuck at kubectl apply synced stage", "pipelines", pipelines)

	installedAppVersions, err := impl.installedAppRepository.GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelinesForAppStore(pendingSinceSeconds, timeForDegradation)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelinesForAppStore", "err", err)
		return err
	}

	impl.logger.Debugw("received argo appStore application stuck at kubectl apply synced stage", "pipelines", installedAppVersions)
	for _, pipeline := range pipelines {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipeline.Id, 0, 1, false)
	}

	for _, installedAppVersion := range installedAppVersions {
		impl.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(0, installedAppVersion.Id, 1, true)
	}
	return nil
}

func (impl *WorkflowStatusServiceImpl) syncACDDevtronApps(deployedBeforeMinutes int, pipelineId int) error {
	if impl.acdConfig.ArgoCDAutoSyncEnabled {
		// don't check for apps if auto sync is enabled
		return nil
	}
	cdWfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest cdWfr by cdPipelineId", "err", err, "pipelineId", pipelineId)
		return err
	}
	if util3.IsTerminalStatus(cdWfr.Status) {
		return nil
	}
	pipelineStatusTimeline, err := impl.pipelineStatusTimelineRepository.FetchLatestTimelineByWfrId(cdWfr.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching latest pipeline status by cdWfrId", "err", err)
		return err
	}
	if pipelineStatusTimeline.Status == pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED && time.Since(pipelineStatusTimeline.StatusTime) >= time.Minute*time.Duration(deployedBeforeMinutes) {
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
			return err
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "token", acdToken)
		syncTime := time.Now()
		syncErr := impl.argocdClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, cdWfr.CdWorkflow.Pipeline.DeploymentAppName)
		if syncErr != nil {
			impl.logger.Errorw("error in syncing argoCD app", "err", syncErr)
			timelineObject := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(cdWfr.Id, 0, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED, fmt.Sprintf("error occured in syncing argocd application. err: %s", syncErr.Error()), 1, time.Now())
			_ = impl.pipelineStatusTimelineService.SaveTimeline(timelineObject, nil, false)
			cdWfr.Status = pipelineConfig.WorkflowFailed
			cdWfr.UpdatedBy = 1
			cdWfr.UpdatedOn = time.Now()
			cdWfrUpdateErr := impl.cdWorkflowRepository.UpdateWorkFlowRunner(&cdWfr)
			if cdWfrUpdateErr != nil {
				impl.logger.Errorw("error in updating cd workflow runner as failed in argocd app sync cron", "err", err)
				return err
			}
			return nil
		}
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: cdWfr.Id,
			StatusTime:         syncTime,
			Status:             pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED,
			StatusDetail:       "argocd sync completed",
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		_, err, _ = impl.pipelineStatusTimelineService.SavePipelineStatusTimelineIfNotAlreadyPresent(timeline.CdWorkflowRunnerId, timeline.Status, timeline, false)
	}
	return nil
}

func (impl *WorkflowStatusServiceImpl) syncACDHelmApps(deployedBeforeMinutes int, installedAppVersionId int) error {
	if impl.acdConfig.ArgoCDAutoSyncEnabled {
		// don't check for apps if auto sync is enabled
		return nil
	}
	installedAppVersionHistory, err := impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistory(installedAppVersionId)
	if err != nil {
		impl.logger.Errorw("error in getting latest cdWfr by cdPipelineId", "err", err, "installedAppVersionId", installedAppVersionId)
		return err
	}
	if util3.IsTerminalStatus(installedAppVersionHistory.Status) {
		return nil
	}
	installedAppVersionHistoryId := installedAppVersionHistory.Id
	pipelineStatusTimeline, err := impl.pipelineStatusTimelineRepository.FetchLatestTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistoryId)
	if err != nil {
		impl.logger.Errorw("error in fetching latest pipeline status by cdWfrId", "err", err)
		return err
	}
	if pipelineStatusTimeline.Status == pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED && time.Since(pipelineStatusTimeline.StatusTime) >= time.Minute*time.Duration(deployedBeforeMinutes) {
		installedApp, err := impl.installedAppRepository.GetInstalledAppByInstalledAppVersionId(installedAppVersionHistory.InstalledAppVersionId)
		if err != nil {
			impl.logger.Errorw("error in fetching installed_app by installedAppVersionId", "err", err)
			return err
		}
		appDetails, err := impl.appRepository.FindActiveById(installedApp.AppId)
		if err != nil {
			impl.logger.Errorw("error in getting appDetails from appId", "err", err)
			return err
		}
		envDetails, err := impl.envRepository.FindById(installedApp.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error in fetching environment by envId", "err", err)
		}
		argoAppName := fmt.Sprintf("%s-%s", appDetails.AppName, envDetails.Name)
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
			return err
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "token", acdToken)
		syncTime := time.Now()
		syncErr := impl.argocdClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, argoAppName)
		if syncErr != nil {
			impl.logger.Errorw("error in syncing argoCD app", "err", syncErr)
			timelineObject := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED, fmt.Sprintf("error occured in syncing argocd application. err: %s", syncErr.Error()), 1, time.Now())
			_ = impl.pipelineStatusTimelineService.SaveTimeline(timelineObject, nil, false)
			installedAppVersionHistory.Status = pipelineConfig.WorkflowFailed
			installedAppVersionHistory.UpdatedBy = 1
			installedAppVersionHistory.UpdatedOn = time.Now()
			_, installedAppUpdateErr := impl.installedAppVersionHistoryRepository.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
			if installedAppUpdateErr != nil {
				impl.logger.Errorw("error in updating cd workflow runner as failed in argocd app sync cron", "err", err)
				return err
			}
			return nil
		}
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: installedAppVersionHistoryId,
			StatusTime:                   syncTime,
			Status:                       pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED,
			StatusDetail:                 "argocd sync completed",
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		_, err, _ = impl.pipelineStatusTimelineService.SavePipelineStatusTimelineIfNotAlreadyPresent(timeline.CdWorkflowRunnerId, timeline.Status, timeline, false)
	}
	return nil
}
