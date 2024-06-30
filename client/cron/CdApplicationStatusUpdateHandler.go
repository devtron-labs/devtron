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

package cron

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	"github.com/devtron-labs/devtron/pkg/workflow/status"
	"github.com/devtron-labs/devtron/util"
	cron2 "github.com/devtron-labs/devtron/util/cron"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type CdApplicationStatusUpdateHandler interface {
	HelmApplicationStatusUpdate()
	ArgoApplicationStatusUpdate()
	ArgoPipelineTimelineUpdate()
	SyncPipelineStatusForResourceTreeCall(pipeline *pipelineConfig.Pipeline) error
	SyncPipelineStatusForAppStoreForResourceTreeCall(installedAppVersion *repository2.InstalledAppVersions) error
	ManualSyncPipelineStatus(appId, envId int, userId int32) error
}

type CdApplicationStatusUpdateHandlerImpl struct {
	logger                               *zap.SugaredLogger
	cron                                 *cron.Cron
	appService                           app.AppService
	workflowDagExecutor                  dag.WorkflowDagExecutor
	installedAppService                  EAMode.InstalledAppDBService
	AppStatusConfig                      *app.AppServiceConfig
	pipelineStatusTimelineRepository     pipelineConfig.PipelineStatusTimelineRepository
	eventClient                          client2.EventClient
	appListingRepository                 repository.AppListingRepository
	cdWorkflowRepository                 pipelineConfig.CdWorkflowRepository
	pipelineRepository                   pipelineConfig.PipelineRepository
	installedAppVersionHistoryRepository repository2.InstalledAppVersionHistoryRepository
	installedAppVersionRepository        repository2.InstalledAppRepository
	cdWorkflowCommonService              cd.CdWorkflowCommonService
	workflowStatusService                status.WorkflowStatusService
}

func NewCdApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, appService app.AppService,
	workflowDagExecutor dag.WorkflowDagExecutor, installedAppService EAMode.InstalledAppDBService,
	AppStatusConfig *app.AppServiceConfig,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	eventClient client2.EventClient, appListingRepository repository.AppListingRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineRepository pipelineConfig.PipelineRepository, installedAppVersionHistoryRepository repository2.InstalledAppVersionHistoryRepository,
	installedAppVersionRepository repository2.InstalledAppRepository, cronLogger *cron2.CronLoggerImpl,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	workflowStatusService status.WorkflowStatusService) *CdApplicationStatusUpdateHandlerImpl {

	cron := cron.New(
		cron.WithChain(cron.SkipIfStillRunning(cronLogger), cron.Recover(cronLogger)))
	cron.Start()
	impl := &CdApplicationStatusUpdateHandlerImpl{
		logger:                               logger,
		cron:                                 cron,
		appService:                           appService,
		workflowDagExecutor:                  workflowDagExecutor,
		installedAppService:                  installedAppService,
		AppStatusConfig:                      AppStatusConfig,
		pipelineStatusTimelineRepository:     pipelineStatusTimelineRepository,
		eventClient:                          eventClient,
		appListingRepository:                 appListingRepository,
		cdWorkflowRepository:                 cdWorkflowRepository,
		pipelineRepository:                   pipelineRepository,
		installedAppVersionHistoryRepository: installedAppVersionHistoryRepository,
		installedAppVersionRepository:        installedAppVersionRepository,
		cdWorkflowCommonService:              cdWorkflowCommonService,
		workflowStatusService:                workflowStatusService,
	}
	_, err := cron.AddFunc(AppStatusConfig.CdHelmPipelineStatusCronTime, impl.HelmApplicationStatusUpdate)
	if err != nil {
		logger.Errorw("error in starting helm application status update cron job", "err", err)
		return nil
	}
	_, err = cron.AddFunc(AppStatusConfig.CdPipelineStatusCronTime, impl.ArgoApplicationStatusUpdate)
	if err != nil {
		logger.Errorw("error in starting argo application status update cron job", "err", err)
		return nil
	}
	_, err = cron.AddFunc("@every 1m", impl.ArgoPipelineTimelineUpdate)
	if err != nil {
		logger.Errorw("error in starting argo application status update cron job", "err", err)
		return nil
	}
	return impl
}

func (impl *CdApplicationStatusUpdateHandlerImpl) HelmApplicationStatusUpdate() {
	cronProcessStartTime := time.Now()
	defer func() {
		middleware.DeploymentStatusCronDuration.WithLabelValues(pipeline.DEVTRON_APP_HELM_PIPELINE_STATUS_UPDATE_CRON).Observe(time.Since(cronProcessStartTime).Seconds())
	}()
	HelmPipelineStatusCheckEligibleTime, err := strconv.Atoi(impl.AppStatusConfig.HelmPipelineStatusCheckEligibleTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	getPipelineDeployedWithinHours := impl.AppStatusConfig.GetPipelineDeployedWithinHours
	err = impl.workflowStatusService.CheckHelmAppStatusPeriodicallyAndUpdateInDb(HelmPipelineStatusCheckEligibleTime, getPipelineDeployedWithinHours)
	if err != nil {
		impl.logger.Errorw("error helm app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ArgoApplicationStatusUpdate() {
	cronProcessStartTime := time.Now()
	defer func() {
		middleware.DeploymentStatusCronDuration.WithLabelValues(pipeline.DEVTRON_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON).Observe(time.Since(cronProcessStartTime).Seconds())
	}()
	// TODO: remove below cron with division of cron for argo pipelines of devtron-apps and helm-apps
	defer func() {
		middleware.DeploymentStatusCronDuration.WithLabelValues(pipeline.HELM_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON).Observe(time.Since(cronProcessStartTime).Seconds())
	}()
	getPipelineDeployedBeforeMinutes, err := strconv.Atoi(impl.AppStatusConfig.PipelineDegradedTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	getPipelineDeployedWithinHours := impl.AppStatusConfig.GetPipelineDeployedWithinHours
	err = impl.workflowStatusService.CheckArgoAppStatusPeriodicallyAndUpdateInDb(getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours)
	if err != nil {
		impl.logger.Errorw("error argo app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ArgoPipelineTimelineUpdate() {
	degradedTime, err := strconv.Atoi(impl.AppStatusConfig.PipelineDegradedTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	err = impl.workflowStatusService.CheckArgoPipelineTimelineStatusPeriodicallyAndUpdateInDb(30, degradedTime)
	if err != nil {
		impl.logger.Errorw("error argo app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) SyncPipelineStatusForResourceTreeCall(pipeline *pipelineConfig.Pipeline) error {
	cdWfr, err := impl.cdWorkflowRepository.FindLatestByPipelineIdAndRunnerType(pipeline.Id, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest cdWfr by cdPipelineId", "err", err, "pipelineId", pipeline.Id)
		return nil
	}
	if !util.IsTerminalRunnerStatus(cdWfr.Status) {
		impl.workflowStatusService.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipeline.Id, 0, 1, false)
	}
	return nil
}

func (impl *CdApplicationStatusUpdateHandlerImpl) SyncPipelineStatusForAppStoreForResourceTreeCall(installedAppVersion *repository2.InstalledAppVersions) error {
	// find installed app version history using parameter obj
	installedAppVersionHistory, err := impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistory(installedAppVersion.Id)
	if err != nil {
		impl.logger.Errorw("error in getting latest installedAppVersionHistory by installedAppVersionId", "err", err, "installedAppVersionId", installedAppVersion.Id)
		return nil
	}
	if !util.IsTerminalRunnerStatus(installedAppVersionHistory.Status) {
		impl.workflowStatusService.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(0, installedAppVersion.Id, 1, true)
	}
	return nil
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ManualSyncPipelineStatus(appId, envId int, userId int32) error {
	var cdPipeline *pipelineConfig.Pipeline
	var installedApp repository2.InstalledApps
	var err error

	if envId == 0 {
		installedApp, err = impl.installedAppVersionRepository.GetInstalledAppByAppIdAndDeploymentType(appId, util2.PIPELINE_DEPLOYMENT_TYPE_ACD)
		if err != nil {
			impl.logger.Errorw("error in getting installed app by appId", "err", err, "appid", appId)
			return nil
		}

	} else {
		cdPipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting cdPipeline by appId and envId", "err", err, "appid", appId, "envId", envId)
			return nil
		}
		if len(cdPipelines) != 1 {
			return fmt.Errorf("invalid number of cd pipelines found")
		}
		cdPipeline = cdPipelines[0]
	}

	err, isTimelineUpdated := impl.workflowStatusService.UpdatePipelineTimelineAndStatusByLiveApplicationFetch(bean2.TriggerContext{}, cdPipeline, installedApp, userId)
	if err != nil {
		impl.logger.Errorw("error on argo pipeline status update", "err", err)
		return nil
	}
	if !isTimelineUpdated {
		return fmt.Errorf("timeline unchanged")
	}

	return nil
}
