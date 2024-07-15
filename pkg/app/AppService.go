/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	health2 "github.com/argoproj/gitops-engine/pkg/health"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/adapter/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	bean4 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"time"

	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/caarlos0/env"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	. "github.com/devtron-labs/devtron/internal/util"
	status2 "github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/commonService"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	_ "github.com/devtron-labs/devtron/pkg/variables/repository"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	util "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppServiceConfig struct {
	CdPipelineStatusCronTime                   string `env:"CD_PIPELINE_STATUS_CRON_TIME" envDefault:"*/2 * * * *"`
	CdHelmPipelineStatusCronTime               string `env:"CD_HELM_PIPELINE_STATUS_CRON_TIME" envDefault:"*/2 * * * *"`
	CdPipelineStatusTimeoutDuration            string `env:"CD_PIPELINE_STATUS_TIMEOUT_DURATION" envDefault:"20"`                   // in minutes
	PipelineDegradedTime                       string `env:"PIPELINE_DEGRADED_TIME" envDefault:"10"`                                // in minutes
	GetPipelineDeployedWithinHours             int    `env:"DEPLOY_STATUS_CRON_GET_PIPELINE_DEPLOYED_WITHIN_HOURS" envDefault:"12"` // in hours
	HelmPipelineStatusCheckEligibleTime        string `env:"HELM_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME" envDefault:"120"`             // in seconds
	ExposeCDMetrics                            bool   `env:"EXPOSE_CD_METRICS" envDefault:"false"`
	DevtronChartHelmInstallRequestTimeout      int    `env:"DEVTRON_CHART_INSTALL_REQUEST_TIMEOUT" envDefault:"6"`         // in minutes
	DevtronChartArgoCdInstallRequestTimeout    int    `env:"DEVTRON_CHART_ARGO_CD_INSTALL_REQUEST_TIMEOUT" envDefault:"1"` // in minutes
	ArgoCdManualSyncCronPipelineDeployedBefore int    `env:"ARGO_APP_MANUAL_SYNC_TIME" envDefault:"3"`                     // in minutes
}

func GetAppServiceConfig() (*AppServiceConfig, error) {
	cfg := &AppServiceConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server app status config: " + err.Error())
		return nil, err
	}
	return cfg, nil
}

type AppServiceImpl struct {
	environmentConfigRepository            chartConfig.EnvConfigOverrideRepository
	pipelineOverrideRepository             chartConfig.PipelineOverrideRepository
	mergeUtil                              *MergeUtil
	logger                                 *zap.SugaredLogger
	pipelineRepository                     pipelineConfig.PipelineRepository
	eventClient                            client.EventClient
	eventFactory                           client.EventFactory
	acdClient                              application.ServiceClient
	appRepository                          app.AppRepository
	configMapRepository                    chartConfig.ConfigMapRepository
	chartRepository                        chartRepoRepository.ChartRepository
	cdWorkflowRepository                   pipelineConfig.CdWorkflowRepository
	commonService                          commonService.CommonService
	chartTemplateService                   ChartTemplateService
	argoUserService                        argo.ArgoUserService
	pipelineStatusTimelineRepository       pipelineConfig.PipelineStatusTimelineRepository
	pipelineStatusTimelineResourcesService status2.PipelineStatusTimelineResourcesService
	pipelineStatusSyncDetailService        status2.PipelineStatusSyncDetailService
	pipelineStatusTimelineService          status2.PipelineStatusTimelineService
	appStatusConfig                        *AppServiceConfig
	appStatusService                       appStatus.AppStatusService
	installedAppRepository                 repository4.InstalledAppRepository
	installedAppVersionHistoryRepository   repository4.InstalledAppVersionHistoryRepository
	scopedVariableManager                  variables.ScopedVariableCMCSManager
	acdConfig                              *argocdServer.ACDConfig
	gitOpsConfigReadService                config.GitOpsConfigReadService
	gitOperationService                    git.GitOperationService
	deploymentTemplateService              deploymentTemplate.DeploymentTemplateService
	appListingService                      AppListingService
}

type AppService interface {
	UpdateReleaseStatus(request *bean.ReleaseStatusUpdateRequest) (bool, error)
	UpdateDeploymentStatusAndCheckIsSucceeded(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, *chartConfig.PipelineOverride, error)
	GetConfigMapAndSecretJson(appId int, envId int, pipelineId int) ([]byte, error)
	UpdateCdWorkflowRunnerByACDObject(app *v1alpha1.Application, cdWfrId int, updateTimedOutStatus bool) error
	GetCmSecretNew(appId int, envId int, isJob bool, scope resourceQualifiers.Scope) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error)
	UpdateDeploymentStatusForGitOpsPipelines(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, bool, *chartConfig.PipelineOverride, error)
	WriteCDSuccessEvent(appId int, envId int, override *chartConfig.PipelineOverride)
	CreateGitOpsRepo(app *app.App, userId int32) (gitopsRepoName string, chartGitAttr *commonBean.ChartGitAttribute, err error)
	GetDeployedManifestByPipelineIdAndCDWorkflowId(appId int, envId int, cdWorkflowId int, ctx context.Context) ([]byte, error)

	// TODO: move inside reader service
	GetActiveCiCdAppsCount() (int, error)
	ComputeAppstatus(appId, envId int, status health2.HealthStatusCode) (string, error)
}

func NewAppService(
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	mergeUtil *MergeUtil, logger *zap.SugaredLogger,
	pipelineRepository pipelineConfig.PipelineRepository,
	eventClient client.EventClient, eventFactory client.EventFactory,
	acdClient application.ServiceClient, appRepository app.AppRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	chartRepository chartRepoRepository.ChartRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	commonService commonService.CommonService,
	chartTemplateService ChartTemplateService, argoUserService argo.ArgoUserService,
	cdPipelineStatusTimelineRepo pipelineConfig.PipelineStatusTimelineRepository,
	pipelineStatusTimelineResourcesService status2.PipelineStatusTimelineResourcesService,
	pipelineStatusSyncDetailService status2.PipelineStatusSyncDetailService,
	pipelineStatusTimelineService status2.PipelineStatusTimelineService,
	appStatusConfig *AppServiceConfig, appStatusService appStatus.AppStatusService,
	installedAppRepository repository4.InstalledAppRepository,
	installedAppVersionHistoryRepository repository4.InstalledAppVersionHistoryRepository,
	scopedVariableManager variables.ScopedVariableCMCSManager, acdConfig *argocdServer.ACDConfig,
	gitOpsConfigReadService config.GitOpsConfigReadService, gitOperationService git.GitOperationService,
	deploymentTemplateService deploymentTemplate.DeploymentTemplateService,
	appListingService AppListingService) *AppServiceImpl {
	appServiceImpl := &AppServiceImpl{
		environmentConfigRepository:            environmentConfigRepository,
		mergeUtil:                              mergeUtil,
		pipelineOverrideRepository:             pipelineOverrideRepository,
		logger:                                 logger,
		pipelineRepository:                     pipelineRepository,
		eventClient:                            eventClient,
		eventFactory:                           eventFactory,
		acdClient:                              acdClient,
		appRepository:                          appRepository,
		configMapRepository:                    configMapRepository,
		chartRepository:                        chartRepository,
		cdWorkflowRepository:                   cdWorkflowRepository,
		commonService:                          commonService,
		chartTemplateService:                   chartTemplateService,
		argoUserService:                        argoUserService,
		pipelineStatusTimelineRepository:       cdPipelineStatusTimelineRepo,
		pipelineStatusTimelineResourcesService: pipelineStatusTimelineResourcesService,
		pipelineStatusSyncDetailService:        pipelineStatusSyncDetailService,
		pipelineStatusTimelineService:          pipelineStatusTimelineService,
		appStatusConfig:                        appStatusConfig,
		appStatusService:                       appStatusService,
		installedAppRepository:                 installedAppRepository,
		installedAppVersionHistoryRepository:   installedAppVersionHistoryRepository,
		scopedVariableManager:                  scopedVariableManager,
		acdConfig:                              acdConfig,
		gitOpsConfigReadService:                gitOpsConfigReadService,
		gitOperationService:                    gitOperationService,
		deploymentTemplateService:              deploymentTemplateService,
		appListingService:                      appListingService,
	}
	return appServiceImpl
}

func (impl *AppServiceImpl) UpdateReleaseStatus(updateStatusRequest *bean.ReleaseStatusUpdateRequest) (bool, error) {
	count, err := impl.pipelineOverrideRepository.UpdateStatusByRequestIdentifier(updateStatusRequest.RequestId, updateStatusRequest.NewStatus)
	if err != nil {
		impl.logger.Errorw("error in updating release status", "request", updateStatusRequest, "error", err)
		return false, err
	}
	return count == 1, nil
}

func (impl *AppServiceImpl) UpdateDeploymentStatusAndCheckIsSucceeded(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, *chartConfig.PipelineOverride, error) {
	isSucceeded := false
	var err error
	var pipelineOverride *chartConfig.PipelineOverride
	if isAppStore {
		var installAppDeleteRequest repository4.InstallAppDeleteRequest
		var gitHash string
		if app.Operation != nil && app.Operation.Sync != nil {
			gitHash = app.Operation.Sync.Revision
		} else if app.Status.OperationState != nil && app.Status.OperationState.Operation.Sync != nil {
			gitHash = app.Status.OperationState.Operation.Sync.Revision
		}
		installAppDeleteRequest, err = impl.installedAppRepository.GetInstalledAppByGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error in fetching installed app by git hash from installed app repository", "err", err)
			return isSucceeded, pipelineOverride, err
		}
		if installAppDeleteRequest.EnvironmentId > 0 {
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(installAppDeleteRequest.AppId, installAppDeleteRequest.EnvironmentId, string(app.Status.Health.Status))
			if err != nil {
				impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", installAppDeleteRequest.AppId, "envId", installAppDeleteRequest.EnvironmentId)
			}
			impl.logger.Debugw("skipping application status update as this app is chart", "appId", installAppDeleteRequest.AppId, "envId", installAppDeleteRequest.EnvironmentId)
		}
	} else {
		repoUrl := app.Spec.Source.RepoURL
		// backward compatibility for updating application status - if unable to find app check it in charts
		chart, err := impl.chartRepository.FindChartByGitRepoUrl(repoUrl)
		if err != nil {
			impl.logger.Errorw("error in fetching chart", "repoUrl", repoUrl, "err", err)
			return isSucceeded, pipelineOverride, err
		}
		if chart == nil {
			impl.logger.Errorw("no git repo found for url", "repoUrl", repoUrl)
			return isSucceeded, pipelineOverride, fmt.Errorf("no git repo found for url %s", repoUrl)
		}
		envId, err := impl.appRepository.FindEnvironmentIdForInstalledApp(chart.AppId)
		if err != nil {
			impl.logger.Errorw("error in fetching app", "err", err, "app", chart.AppId)
			return isSucceeded, pipelineOverride, err
		}
		if envId > 0 {
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(chart.AppId, envId, string(app.Status.Health.Status))
			if err != nil {
				impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", chart.AppId, "envId", envId)
			}
			impl.logger.Debugw("skipping application status update as this app is chart", "appId", chart.AppId, "envId", envId)
		}
	}

	isSucceeded, _, pipelineOverride, err = impl.UpdateDeploymentStatusForGitOpsPipelines(app, statusTime, isAppStore)
	if err != nil {
		impl.logger.Errorw("error in updating deployment status", "argoAppName", app.Name)
		return isSucceeded, pipelineOverride, err
	}
	return isSucceeded, pipelineOverride, nil
}

func (impl *AppServiceImpl) ComputeAppstatus(appId, envId int, status health2.HealthStatusCode) (string, error) {
	appStatusInternal := string(status)

	// get the last accepted deploy type workflow runner, accepted state -> not in (initiated/queued/failed)
	cdWfr, err := impl.cdWorkflowRepository.FindLastUnFailedProcessedRunner(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting latest wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
		return "", err
	}

	override, err := impl.pipelineOverrideRepository.FindLatestByCdWorkflowId(cdWfr.CdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in getting latest wfr by pipelineId", "cdWorkflowId", cdWfr.CdWorkflowId, "err", err)
		return "", err
	}

	if errors.Is(err, pg.ErrNoRows) {
		// not deployed
		return appStatusInternal, nil
	}

	// this is not stop/start type
	if override.DeploymentType != models.DEPLOYMENTTYPE_STOP && override.DeploymentType != models.DEPLOYMENTTYPE_START {
		return appStatusInternal, nil
	}

	// for stop, then user requested for hibernate, then check for hibernation.
	IslastAcceptedReleaseIsStopRequest := models.DEPLOYMENTTYPE_STOP == override.DeploymentType
	// request essentially means that the previous state of the release was hibernation/partial hibernation

	if IslastAcceptedReleaseIsStopRequest {
		appStatusInternal = appStatus.GetHibernationStatus(status)
	}
	return appStatusInternal, nil
}

func (impl *AppServiceImpl) UpdateDeploymentStatusForGitOpsPipelines(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, bool, *chartConfig.PipelineOverride, error) {
	isSucceeded := false
	isTimelineUpdated := false
	isTimelineTimedOut := false
	gitHash := ""
	var err error
	var pipelineOverride *chartConfig.PipelineOverride
	if app != nil {
		gitHash = app.Status.Sync.Revision
	}
	if !isAppStore {
		var isValid bool
		var cdPipeline pipelineConfig.Pipeline
		var cdWfr pipelineConfig.CdWorkflowRunner
		isValid, cdPipeline, cdWfr, pipelineOverride, err = impl.CheckIfPipelineUpdateEventIsValid(app.Name, gitHash)
		if err != nil {
			impl.logger.Errorw("service err, CheckIfPipelineUpdateEventIsValid", "err", err)
			return isSucceeded, isTimelineUpdated, pipelineOverride, err
		}
		if !isValid {
			impl.logger.Infow("deployment status event invalid, skipping", "appName", app.Name)
			return isSucceeded, isTimelineUpdated, pipelineOverride, nil
		}
		timeoutDuration, err := strconv.Atoi(impl.appStatusConfig.CdPipelineStatusTimeoutDuration)
		if err != nil {
			impl.logger.Errorw("error in converting string to int", "err", err)
			return isSucceeded, isTimelineUpdated, pipelineOverride, err
		}
		lastStatusTime, err := impl.pipelineStatusTimelineRepository.GetLastStatusPublishedTimeForWfrId(cdWfr.Id)
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in getting latest timeline before update", "err", err, "cdWfrId", cdWfr.Id)
			return isSucceeded, isTimelineUpdated, pipelineOverride, err
		}

		appStatusInternal, err := impl.ComputeAppstatus(cdPipeline.AppId, cdPipeline.EnvironmentId, app.Status.Health.Status)
		if err != nil {
			impl.logger.Errorw("error in checking if last release is stop type", "err", err, cdPipeline.AppId, "envId", cdPipeline.EnvironmentId)
			return isSucceeded, isTimelineUpdated, pipelineOverride, err
		}
		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(cdPipeline.AppId, cdPipeline.EnvironmentId, appStatusInternal)
		if err != nil {
			impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", cdPipeline.AppId, "envId", cdPipeline.EnvironmentId)
		}
		reconciledAt := &metav1.Time{}
		if app != nil {
			reconciledAt = app.Status.ReconciledAt
		}
		var kubectlSyncedTimeline *pipelineConfig.PipelineStatusTimeline
		// updating cd pipeline status timeline
		isTimelineUpdated, isTimelineTimedOut, kubectlSyncedTimeline, err = impl.UpdatePipelineStatusTimelineForApplicationChanges(app, cdWfr.Id, statusTime, cdWfr.StartedOn, timeoutDuration, lastStatusTime, reconciledAt, false)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline status timeline", "err", err)
		}
		if isTimelineTimedOut {
			// not checking further and directly updating timedOutStatus
			err := impl.UpdateCdWorkflowRunnerByACDObject(app, cdWfr.Id, true)
			if err != nil {
				impl.logger.Errorw("error on update cd workflow runner", "CdWorkflowId", pipelineOverride.CdWorkflowId, "status", pipelineConfig.WorkflowTimedOut, "err", err)
				return isSucceeded, isTimelineUpdated, pipelineOverride, err
			}
			return isSucceeded, isTimelineUpdated, pipelineOverride, nil
		}
		if reconciledAt.IsZero() || (kubectlSyncedTimeline != nil && kubectlSyncedTimeline.Id > 0 && reconciledAt.After(kubectlSyncedTimeline.StatusTime)) {
			releaseCounter, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(pipelineOverride.PipelineId)
			if err != nil {
				impl.logger.Errorw("error on update application status", "releaseCounter", releaseCounter, "gitHash", gitHash, "pipelineOverride", pipelineOverride, "err", err)
				return isSucceeded, isTimelineUpdated, pipelineOverride, err
			}
			if pipelineOverride.PipelineReleaseCounter == releaseCounter {
				isSucceeded, err = impl.UpdateDeploymentStatusForPipeline(app, pipelineOverride, cdWfr.Id)
				if err != nil {
					impl.logger.Errorw("error in updating deployment status for pipeline", "err", err)
					return isSucceeded, isTimelineUpdated, pipelineOverride, err
				}
				if isSucceeded {
					impl.logger.Infow("writing cd success event", "gitHash", gitHash, "pipelineOverride", pipelineOverride)
					go impl.WriteCDSuccessEvent(cdPipeline.AppId, cdPipeline.EnvironmentId, pipelineOverride)
				}
			} else {
				impl.logger.Debugw("event received for older triggered revision", "gitHash", gitHash)
			}
		} else {
			// new revision is not reconciled yet, thus status will not be changes and will remain in progress
		}
	} else {
		isValid, installedAppVersionHistory, appId, envId, err := impl.CheckIfPipelineUpdateEventIsValidForAppStore(app.ObjectMeta.Name, gitHash)
		if err != nil {
			impl.logger.Errorw("service err, CheckIfPipelineUpdateEventIsValidForAppStore", "err", err)
			return isSucceeded, isTimelineUpdated, pipelineOverride, err
		}
		if !isValid {
			impl.logger.Infow("deployment status event invalid, skipping", "appName", app.Name)
			return isSucceeded, isTimelineUpdated, pipelineOverride, nil
		}
		timeoutDuration, err := strconv.Atoi(impl.appStatusConfig.CdPipelineStatusTimeoutDuration)
		if err != nil {
			impl.logger.Errorw("error in converting string to int", "err", err)
			return isSucceeded, isTimelineUpdated, pipelineOverride, err
		}
		latestTimelineBeforeThisEvent, err := impl.pipelineStatusTimelineRepository.FetchLatestTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistory.Id)
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in getting latest timeline before update", "err", err, "installedAppVersionHistoryId", installedAppVersionHistory.Id)
			return isSucceeded, isTimelineUpdated, pipelineOverride, err
		}

		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(appId, envId, string(app.Status.Health.Status))
		if err != nil {
			impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", appId, "envId", envId)
		}
		// reconcile time is how often your applications will sync from Argo CD to the Git repository
		reconciledAt := &metav1.Time{}
		if app != nil {
			reconciledAt = app.Status.ReconciledAt
		}
		var kubectlSyncedTimeline *pipelineConfig.PipelineStatusTimeline
		// updating versionHistory pipeline status timeline
		isTimelineUpdated, isTimelineTimedOut, kubectlSyncedTimeline, err = impl.UpdatePipelineStatusTimelineForApplicationChanges(app, installedAppVersionHistory.Id, statusTime, installedAppVersionHistory.StartedOn, timeoutDuration, latestTimelineBeforeThisEvent.StatusTime, reconciledAt, true)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline status timeline", "err", err)
		}
		if isTimelineTimedOut {
			// not checking further and directly updating timedOutStatus
			err := impl.UpdateInstalledAppVersionHistoryByACDObject(app, installedAppVersionHistory.Id, true)
			if err != nil {
				impl.logger.Errorw("error on update installedAppVersionHistory", "installedAppVersionHistory", installedAppVersionHistory.Id, "status", pipelineConfig.WorkflowTimedOut, "err", err)
				return isSucceeded, isTimelineUpdated, pipelineOverride, err
			}
			return isSucceeded, isTimelineUpdated, pipelineOverride, nil
		}

		if reconciledAt.IsZero() || (kubectlSyncedTimeline != nil && kubectlSyncedTimeline.Id > 0) {
			isSucceeded, err = impl.UpdateDeploymentStatusForAppStore(app, installedAppVersionHistory.Id)
			if err != nil {
				impl.logger.Errorw("error in updating deployment status for pipeline", "err", err)
				return isSucceeded, isTimelineUpdated, pipelineOverride, err
			}
			if isSucceeded {
				impl.logger.Infow("writing installed app success event", "gitHash", gitHash, "installedAppVersionHistory", installedAppVersionHistory)
			}
		} else {
			impl.logger.Debugw("event received for older triggered revision", "gitHash", gitHash)
		}
	}

	return isSucceeded, isTimelineUpdated, pipelineOverride, nil
}

func (impl *AppServiceImpl) CheckIfPipelineUpdateEventIsValidForAppStore(gitOpsAppName string, gitHash string) (bool, *repository4.InstalledAppVersionHistory, int, int, error) {
	isValid := false
	var err error
	installedAppVersionHistory := &repository4.InstalledAppVersionHistory{}
	// checking if the gitOpsAppName is present in installed_apps table, if yes the find installed_app_version_history else return
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppByGitOpsAppName(gitOpsAppName)
	if err != nil {
		impl.logger.Errorw("error in getting all installed apps in GetAllGitOpsAppNameAndInstalledAppMapping", "err", err, "gitOpsAppName", gitOpsAppName)
		return isValid, installedAppVersionHistory, 0, 0, err
	}

	installedAppVersionHistory, err = impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistoryByInstalledAppId(installedAppModel.Id)
	if err != nil {
		impl.logger.Errorw("error in getting latest installedAppVersionHistory by installedAppId", "err", err, "installedAppId", installedAppModel.Id)
		return isValid, installedAppVersionHistory, 0, 0, err
	}
	appId, envId, err := impl.installedAppVersionHistoryRepository.GetAppIdAndEnvIdWithInstalledAppVersionId(installedAppVersionHistory.InstalledAppVersionId)
	if err != nil {
		impl.logger.Errorw("error in getting appId and environmentId using installedAppVersionId", "err", err, "installedAppVersionId", installedAppVersionHistory.InstalledAppVersionId)
		return isValid, installedAppVersionHistory, 0, 0, err
	}
	if gitHash != "" && installedAppVersionHistory.GitHash != gitHash {
		installedAppVersionHistoryByHash, err := impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistoryByGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error on update application status", "gitHash", gitHash, "installedAppVersionHistory", installedAppVersionHistory, "err", err)
			return isValid, installedAppVersionHistory, appId, envId, err
		}
		if installedAppVersionHistoryByHash.StartedOn.Before(installedAppVersionHistory.StartedOn) {
			// we have received trigger hash which is committed before this apps actual gitHash stored by us
			// this means that the hash stored by us will be synced later, so we will drop this event
			return isValid, installedAppVersionHistory, appId, envId, nil
		}
	}
	if util2.IsTerminalRunnerStatus(installedAppVersionHistory.Status) {
		// drop event
		return isValid, installedAppVersionHistory, appId, envId, nil
	}
	if impl.acdConfig.IsManualSyncEnabled() {
		isArgoAppSynced := impl.pipelineStatusTimelineService.GetArgoAppSyncStatusForAppStore(installedAppVersionHistory.Id)
		if !isArgoAppSynced {
			return isValid, installedAppVersionHistory, appId, envId, nil
		}
	}
	isValid = true
	return isValid, installedAppVersionHistory, appId, envId, err
}

func (impl *AppServiceImpl) CheckIfPipelineUpdateEventIsValid(argoAppName, gitHash string) (bool, pipelineConfig.Pipeline, pipelineConfig.CdWorkflowRunner, *chartConfig.PipelineOverride, error) {
	isValid := false
	var err error
	// var deploymentStatus repository.DeploymentStatus
	var pipeline pipelineConfig.Pipeline
	var pipelineOverride *chartConfig.PipelineOverride
	var cdWfr pipelineConfig.CdWorkflowRunner
	pipeline, err = impl.pipelineRepository.GetArgoPipelineByArgoAppName(argoAppName)
	if err != nil {
		impl.logger.Errorw("error in getting cd pipeline by argoAppName", "err", err, "argoAppName", argoAppName)
		return isValid, pipeline, cdWfr, pipelineOverride, err
	}
	// getting latest pipelineOverride for app (by appId and envId)
	pipelineOverride, err = impl.pipelineOverrideRepository.FindLatestByAppIdAndEnvId(pipeline.AppId, pipeline.EnvironmentId, bean4.ArgoCd)
	if err != nil {
		impl.logger.Errorw("error in getting latest pipelineOverride by appId and envId", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
		return isValid, pipeline, cdWfr, pipelineOverride, err
	}
	if gitHash != "" && pipelineOverride.GitHash != gitHash {
		pipelineOverrideByHash, err := impl.pipelineOverrideRepository.FindByPipelineTriggerGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error on update application status", "gitHash", gitHash, "pipelineOverride", pipelineOverride, "err", err)
			return isValid, pipeline, cdWfr, pipelineOverride, err
		}
		if pipelineOverrideByHash.CommitTime.Before(pipelineOverride.CommitTime) {
			// we have received trigger hash which is committed before this apps actual gitHash stored by us
			// this means that the hash stored by us will be synced later, so we will drop this event
			return isValid, pipeline, cdWfr, pipelineOverride, nil
		}
	}
	cdWfr, err = impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), pipelineOverride.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest wfr by pipelineId", "err", err, "pipelineId", pipeline.Id)
		return isValid, pipeline, cdWfr, pipelineOverride, err
	}
	if util2.IsTerminalRunnerStatus(cdWfr.Status) {
		// drop event
		return isValid, pipeline, cdWfr, pipelineOverride, nil
	}
	if impl.acdConfig.IsManualSyncEnabled() {
		// if manual sync, proceed only if ARGOCD_SYNC_COMPLETED timeline is created
		isArgoAppSynced := impl.pipelineStatusTimelineService.GetArgoAppSyncStatus(cdWfr.Id)
		if !isArgoAppSynced {
			return isValid, pipeline, cdWfr, pipelineOverride, nil
		}
	}
	isValid = true
	return isValid, pipeline, cdWfr, pipelineOverride, nil
}

func (impl *AppServiceImpl) UpdateDeploymentStatusForPipeline(app *v1alpha1.Application, pipelineOverride *chartConfig.PipelineOverride, cdWfrId int) (bool, error) {
	impl.logger.Debugw("inserting new app status", "status", app.Status.Health.Status, "argoAppName", app.Name)
	isSucceeded := false
	err := impl.UpdateCdWorkflowRunnerByACDObject(app, cdWfrId, false)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner", "CdWorkflowId", pipelineOverride.CdWorkflowId, "app", app, "err", err)
		return isSucceeded, err
	}
	if argoApplication.Healthy == app.Status.Health.Status {
		isSucceeded = true
	}
	return isSucceeded, nil
}

func (impl *AppServiceImpl) UpdateDeploymentStatusForAppStore(app *v1alpha1.Application, installedVersionHistoryId int) (bool, error) {
	impl.logger.Debugw("inserting new app status", "status", app.Status.Health.Status, "argoAppName", app.Name)
	isSucceeded := false
	err := impl.UpdateInstalledAppVersionHistoryByACDObject(app, installedVersionHistoryId, false)
	if err != nil {
		impl.logger.Errorw("error on update installed version history", "installedVersionHistoryId", installedVersionHistoryId, "app", app, "err", err)
		return isSucceeded, err
	}
	if argoApplication.Healthy == app.Status.Health.Status {
		isSucceeded = true
	}
	return isSucceeded, nil
}

func (impl *AppServiceImpl) UpdatePipelineStatusTimelineForApplicationChanges(app *v1alpha1.Application, runnerHistoryId int, statusTime time.Time, triggeredAt time.Time, statusTimeoutDuration int, lastStatusTime time.Time, reconciledAt *metav1.Time, isAppStore bool) (isTimelineUpdated bool, isTimelineTimedOut bool, kubectlApplySyncedTimeline *pipelineConfig.PipelineStatusTimeline, err error) {
	// runnerHistoryId can be wfrId or installedAppVersionHistoryId
	impl.logger.Debugw("updating pipeline status timeline", "app", app, "pipelineOverride", runnerHistoryId, "APP_TO_UPDATE", app.Name)
	isTimelineUpdated = false
	isTimelineTimedOut = false
	if !isAppStore {
		terminalStatusExists, err := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByWfrId(runnerHistoryId)
		if err != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", err, "wfrId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		if terminalStatusExists {
			impl.logger.Infow("terminal status timeline exists for cdWfr, skipping more timeline changes", "wfrId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, nil
		}
		preRequiredStatusExists, err := impl.pipelineStatusTimelineRepository.CheckIfTimelineStatusPresentByWfrId(runnerHistoryId, timelineStatus.TIMELINE_STATUS_GIT_COMMIT)
		if err != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", err, "wfrId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		if !preRequiredStatusExists {
			impl.logger.Errorw("pre-condition failed: timeline for GIT_COMMIT is missing for wfrId", "wfrId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, fmt.Errorf("pre-condition failed timeline status update")
		}
		err = impl.pipelineStatusSyncDetailService.SaveOrUpdateSyncDetail(runnerHistoryId, 1)
		if err != nil {
			impl.logger.Errorw("error in save/update pipeline status fetch detail", "err", err, "cdWfrId", runnerHistoryId)
		}
		// creating cd pipeline status timeline
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: runnerHistoryId,
			StatusTime:         statusTime,
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		timeline.Status = timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_STARTED
		if app != nil && app.Status.OperationState != nil {
			timeline.StatusDetail = app.Status.OperationState.Message
		}
		// checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
		isTimelineUpdated, err = impl.pipelineStatusTimelineService.SaveTimelineIfNotAlreadyPresent(timeline, nil)
		if err != nil {
			impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		// saving timeline resource details
		err = impl.pipelineStatusTimelineResourcesService.SaveOrUpdatePipelineTimelineResources(runnerHistoryId, app, nil, 1, false)
		if err != nil {
			impl.logger.Errorw("error in saving/updating timeline resources", "err", err, "cdWfrId", runnerHistoryId)
		}
		var kubectlSyncTimelineFetchErr error
		kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr = impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatus(runnerHistoryId, timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED)
		if kubectlSyncTimelineFetchErr != nil && !errors.Is(kubectlSyncTimelineFetchErr, pg.ErrNoRows) {
			impl.logger.Errorw("error in getting latest timeline", "err", kubectlSyncTimelineFetchErr, "cdWfrId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr
		}
		if (kubectlApplySyncedTimeline == nil || kubectlApplySyncedTimeline.Id == 0) && app != nil && app.Status.OperationState != nil && string(app.Status.OperationState.Phase) == string(k8sCommonBean.OperationSucceeded) {
			timeline.Id = 0
			timeline.Status = timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED
			timeline.StatusDetail = app.Status.OperationState.Message
			// checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil)
			if err != nil {
				impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
				return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
			}
			isTimelineUpdated = true
			kubectlApplySyncedTimeline = timeline
			impl.logger.Debugw("APP_STATUS_UPDATE_REQ", "stage", "APPLY_SYNCED", "app", app, "status", timeline.Status)
		}
		if reconciledAt.IsZero() || (kubectlApplySyncedTimeline != nil && kubectlApplySyncedTimeline.Id > 0 && reconciledAt.After(kubectlApplySyncedTimeline.StatusTime)) {
			haveNewTimeline := false
			timeline.Id = 0
			if string(app.Status.Health.Status) == string(health.HealthStatusHealthy) {
				impl.logger.Infow("updating pipeline status timeline for healthy app", "app", app, "APP_TO_UPDATE", app.Name)
				haveNewTimeline = true
				timeline.Status = timelineStatus.TIMELINE_STATUS_APP_HEALTHY
				timeline.StatusDetail = "App status is Healthy."
			}
			if haveNewTimeline {
				// not checking if this status is already present or not because already checked for terminal status existence earlier
				err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil)
				if err != nil {
					impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
					return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
				}
				isTimelineUpdated = true
				impl.logger.Debugw("APP_STATUS_UPDATE_REQ", "stage", "terminal_status", "app", app, "status", timeline.Status)
			}
		}

		if !isTimelineUpdated {
			// no timeline updated since before, in this case we will check for timeout cases
			var lastTimeToCheckForTimeout time.Time
			if lastStatusTime.IsZero() {
				lastTimeToCheckForTimeout = triggeredAt
			} else {
				lastTimeToCheckForTimeout = lastStatusTime
			}
			if time.Since(lastTimeToCheckForTimeout) >= time.Duration(statusTimeoutDuration)*time.Minute {
				// mark as timed out if not already marked
				timeline.Status = timelineStatus.TIMELINE_STATUS_FETCH_TIMED_OUT
				timeline.StatusDetail = "Deployment timed out."
				isTimelineUpdated, err = impl.pipelineStatusTimelineService.SaveTimelineIfNotAlreadyPresent(timeline, nil)
				if err != nil {
					impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
					return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
				}
				isTimelineTimedOut = true
			} else {
				// deployment status will be in progress so leave timeline
			}
		}
	} else {
		terminalStatusExists, err := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(runnerHistoryId)
		if err != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		if terminalStatusExists {
			impl.logger.Infow("terminal status timeline exists for installed App, skipping more timeline changes", "installedAppVersionHistoryId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, nil
		}
		preRequiredStatusExists, err := impl.pipelineStatusTimelineRepository.CheckIfTimelineStatusPresentByInstalledAppVersionHistoryId(runnerHistoryId, timelineStatus.TIMELINE_STATUS_GIT_COMMIT)
		if err != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", err, "installedAppVersionHistoryId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		if !preRequiredStatusExists {
			impl.logger.Errorw("pre-condition failed: timeline for GIT_COMMIT is missing for installedAppVersionHistoryId", "installedAppVersionHistoryId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, fmt.Errorf("pre-condition failed timeline status update")
		}
		err = impl.pipelineStatusSyncDetailService.SaveOrUpdateSyncDetailForAppStore(runnerHistoryId, 1)
		if err != nil {
			impl.logger.Errorw("error in save/update pipeline status fetch detail", "err", err, "installedAppVersionHistoryId", runnerHistoryId)
		}
		// creating installedAppVersionHistory status timeline
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: runnerHistoryId,
			StatusTime:                   statusTime,
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		timeline.Status = timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_STARTED
		if app != nil && app.Status.OperationState != nil {
			timeline.StatusDetail = app.Status.OperationState.Message
		}
		// checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
		isTimelineUpdated, err = impl.pipelineStatusTimelineService.SaveTimelineIfNotAlreadyPresent(timeline, nil)
		if err != nil {
			impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		// saving timeline resource details
		err = impl.pipelineStatusTimelineResourcesService.SaveOrUpdatePipelineTimelineResources(runnerHistoryId, app, nil, 1, true)
		if err != nil {
			impl.logger.Errorw("error in saving/updating timeline resources", "err", err, "installedAppVersionId", runnerHistoryId)
		}
		var kubectlSyncTimelineFetchErr error
		kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr = impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(runnerHistoryId, timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED)
		if kubectlSyncTimelineFetchErr != nil && kubectlSyncTimelineFetchErr != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline", "err", kubectlSyncTimelineFetchErr, "installedAppVersionHistoryId", runnerHistoryId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr
		}
		if (kubectlApplySyncedTimeline == nil || kubectlApplySyncedTimeline.Id == 0) && app != nil && app.Status.OperationState != nil && string(app.Status.OperationState.Phase) == string(k8sCommonBean.OperationSucceeded) {
			timeline.Id = 0
			timeline.Status = timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED
			timeline.StatusDetail = app.Status.OperationState.Message
			// checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil)
			if err != nil {
				impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
				return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
			}
			isTimelineUpdated = true
			kubectlApplySyncedTimeline = timeline
			impl.logger.Debugw("APP_STATUS_UPDATE_REQ", "stage", "APPLY_SYNCED", "app", app, "status", timeline.Status)
		}
		if reconciledAt.IsZero() || (kubectlApplySyncedTimeline != nil && kubectlApplySyncedTimeline.Id > 0) {
			haveNewTimeline := false
			timeline.Id = 0
			if string(app.Status.Health.Status) == string(health.HealthStatusHealthy) {
				impl.logger.Infow("updating pipeline status timeline for healthy app", "app", app, "APP_TO_UPDATE", app.Name)
				haveNewTimeline = true
				timeline.Status = timelineStatus.TIMELINE_STATUS_APP_HEALTHY
				timeline.StatusDetail = "App status is Healthy."
			}
			if haveNewTimeline {
				// not checking if this status is already present or not because already checked for terminal status existence earlier
				err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil)
				if err != nil {
					impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
					return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
				}
				isTimelineUpdated = true
				impl.logger.Debugw("APP_STATUS_UPDATE_REQ", "stage", "terminal_status", "app", app, "status", timeline.Status)
			}
		}

		if !isTimelineUpdated {
			// no timeline updated since before, in this case we will check for timeout cases
			var lastTimeToCheckForTimeout time.Time
			if lastStatusTime.IsZero() {
				lastTimeToCheckForTimeout = triggeredAt
			} else {
				lastTimeToCheckForTimeout = lastStatusTime
			}
			if time.Since(lastTimeToCheckForTimeout) >= time.Duration(statusTimeoutDuration)*time.Minute {
				// mark as timed out if not already marked
				timeline.Status = timelineStatus.TIMELINE_STATUS_FETCH_TIMED_OUT
				timeline.StatusDetail = "Deployment timed out."
				isTimelineUpdated, err = impl.pipelineStatusTimelineService.SaveTimelineIfNotAlreadyPresent(timeline, nil)
				if err != nil {
					impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
					return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
				}
				isTimelineTimedOut = true
			} else {
				// deployment status will be in progress so leave timeline
			}
		}
	}
	return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, nil
}

func (impl *AppServiceImpl) WriteCDSuccessEvent(appId int, envId int, override *chartConfig.PipelineOverride) {
	event := impl.eventFactory.Build(util.Success, &override.PipelineId, appId, &envId, util.CD)
	impl.logger.Debugw("event WriteCDSuccessEvent", "event", event, "override", override)
	event = impl.eventFactory.BuildExtraCDData(event, nil, override.Id, bean.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "event", event, "err", evtErr)
	}
}

func (impl *AppServiceImpl) BuildCDSuccessPayload(appName string, environmentName string) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = appName
	payload.EnvName = environmentName
	return payload
}

type ValuesOverrideResponse struct {
	MergedValues        string
	ReleaseOverrideJSON string
	EnvOverride         *chartConfig.EnvConfigOverride
	PipelineStrategy    *chartConfig.PipelineStrategy
	PipelineOverride    *chartConfig.PipelineOverride
	Artifact            *repository.CiArtifact
	Pipeline            *pipelineConfig.Pipeline
}

func (impl *AppServiceImpl) buildACDContext() (acdContext context.Context, err error) {
	// this method should only call in case of argo-integration and gitops configured
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return nil, err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	return ctx, nil
}

func (impl *AppServiceImpl) GetDeployedManifestByPipelineIdAndCDWorkflowId(appId int, envId int, cdWorkflowId int, ctx context.Context) ([]byte, error) {

	manifestByteArray := make([]byte, 0)

	pipeline, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by appId and envId", "appId", appId, "envId", envId, "err", err)
		return manifestByteArray, err
	}

	pipelineOverride, err := impl.pipelineOverrideRepository.FindLatestByCdWorkflowId(cdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in fetching latest release by appId and envId", "appId", appId, "envId", envId, "err", err)
		return manifestByteArray, err
	}

	envConfigOverride, err := impl.environmentConfigRepository.GetByIdIncludingInactive(pipelineOverride.EnvConfigOverrideId)
	if err != nil {
		impl.logger.Errorw("error in fetching env config repository by appId and envId", "appId", appId, "envId", envId, "err", err)
	}

	appName := pipeline[0].App.AppName
	builtChartPath, err := impl.deploymentTemplateService.BuildChartAndGetPath(appName, envConfigOverride, ctx)
	if err != nil {
		impl.logger.Errorw("error in parsing reference chart", "err", err)
		return manifestByteArray, err
	}

	// create values file in built chart path
	valuesFilePath := path.Join(builtChartPath, "valuesOverride.yaml")
	err = ioutil.WriteFile(valuesFilePath, []byte(pipelineOverride.PipelineMergedValues), 0600)
	if err != nil {
		return manifestByteArray, nil
	}

	manifestByteArray, err = impl.chartTemplateService.LoadChartInBytes(builtChartPath, true)
	if err != nil {
		impl.logger.Errorw("error in converting chart to bytes", "err", err)
		return manifestByteArray, err
	}

	return manifestByteArray, nil

}

func (impl *AppServiceImpl) CreateGitOpsRepo(app *app.App, userId int32) (gitopsRepoName string, chartGitAttr *commonBean.ChartGitAttribute, err error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(app.Id)
	if err != nil && pg.ErrNoRows != err {
		return "", nil, err
	}
	gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(app.AppName)
	chartGitAttr, err = impl.gitOperationService.CreateGitRepositoryForDevtronApp(context.Background(), gitOpsRepoName, userId)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git ", "gitOpsRepoName", gitOpsRepoName, "err", err)
		return "", nil, err
	}
	chartGitAttr.ChartLocation = filepath.Join(chart.ReferenceTemplate, chart.ChartVersion)
	return gitOpsRepoName, chartGitAttr, nil
}

// FIXME tmp workaround
func (impl *AppServiceImpl) GetCmSecretNew(appId int, envId int, isJob bool, scope resourceQualifiers.Scope) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error) {
	var configMapJson string
	var secretDataJson string
	var configMapJsonApp string
	var secretDataJsonApp string
	var configMapJsonEnv string
	var secretDataJsonEnv string
	// var configMapJsonPipeline string
	// var secretDataJsonPipeline string

	configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		return nil, nil, err
	}
	if configMapA != nil && configMapA.Id > 0 {
		configMapJsonApp = configMapA.ConfigMapData
		secretDataJsonApp = configMapA.SecretData
	}

	configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		return nil, nil, err
	}
	if configMapE != nil && configMapE.Id > 0 {
		configMapJsonEnv = configMapE.ConfigMapData
		secretDataJsonEnv = configMapE.SecretData
	}

	configMapJson, err = impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
	if err != nil {
		return nil, nil, err
	}
	var chartMajorVersion int
	var chartMinorVersion int
	if !isJob {
		chart, err := impl.commonService.FetchLatestChart(appId, envId)
		if err != nil {
			return nil, nil, err
		}

		chartVersion := chart.ChartVersion
		chartMajorVersion, chartMinorVersion, err = util2.ExtractChartVersion(chartVersion)
		if err != nil {
			impl.logger.Errorw("chart version parsing", "err", err)
			return nil, nil, err
		}
	}
	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion, isJob)
	if err != nil {
		return nil, nil, err
	}
	configResponse := bean.ConfigMapJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(configMapJson), &configResponse)
		if err != nil {
			return nil, nil, err
		}
	}
	secretResponse := bean.ConfigSecretJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
		if err != nil {
			return nil, nil, err
		}
	}

	resolvedConfigResponse, resolvedSecretResponse, err := impl.scopedVariableManager.ResolveForPrePostStageTrigger(scope, configResponse, secretResponse, configMapA.Id, configMapE.Id)
	if err != nil {
		return nil, nil, err
	}

	return resolvedConfigResponse, resolvedSecretResponse, nil
}

// depricated
// TODO remove this method
func (impl *AppServiceImpl) GetConfigMapAndSecretJson(appId int, envId int, pipelineId int) ([]byte, error) {
	var configMapJson string
	var secretDataJson string
	merged := []byte("{}")
	configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		return []byte("{}"), err
	}
	if configMapA != nil && configMapA.Id > 0 {
		configMapJson = configMapA.ConfigMapData
		secretDataJson = configMapA.SecretData
		if configMapJson == "" {
			configMapJson = "{}"
		}
		if secretDataJson == "" {
			secretDataJson = "{}"
		}
		config, err := impl.mergeUtil.JsonPatch([]byte(configMapJson), []byte(secretDataJson))
		if err != nil {
			return []byte("{}"), err
		}
		merged, err = impl.mergeUtil.JsonPatch(merged, config)
		if err != nil {
			return []byte("{}"), err
		}
	}

	configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		return []byte("{}"), err
	}
	if configMapE != nil && configMapE.Id > 0 {
		configMapJson = configMapE.ConfigMapData
		secretDataJson = configMapE.SecretData
		if configMapJson == "" {
			configMapJson = "{}"
		}
		if secretDataJson == "" {
			secretDataJson = "{}"
		}
		config, err := impl.mergeUtil.JsonPatch([]byte(configMapJson), []byte(secretDataJson))
		if err != nil {
			return []byte("{}"), err
		}
		merged, err = impl.mergeUtil.JsonPatch(merged, config)
		if err != nil {
			return []byte("{}"), err
		}
	}

	return merged, nil
}

func (impl *AppServiceImpl) synchCD(pipeline *pipelineConfig.Pipeline, ctx context.Context,
	overrideRequest *bean.ValuesOverrideRequest, envOverride *chartConfig.EnvConfigOverride) {
	req := new(application2.ApplicationSyncRequest)
	pipelineName := pipeline.App.AppName + "-" + envOverride.Environment.Name
	req.Name = &pipelineName
	prune := true
	req.Prune = &prune
	if ctx == nil {
		impl.logger.Errorw("err in syncing ACD, ctx is NULL", "pipelineId", overrideRequest.PipelineId)
		return
	}
	if _, err := impl.acdClient.Sync(ctx, req); err != nil {
		impl.logger.Errorw("err in syncing ACD", "pipelineId", overrideRequest.PipelineId, "err", err)
	}
}

type DeploymentEvent struct {
	ApplicationId      int
	EnvironmentId      int
	ReleaseId          int
	PipelineOverrideId int
	TriggerTime        time.Time
	PipelineMaterials  []*PipelineMaterialInfo
	CiArtifactId       int
}
type PipelineMaterialInfo struct {
	PipelineMaterialId int
	CommitHash         string
}

func buildCDTriggerEvent(impl *AppServiceImpl, overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline,
	envOverride *chartConfig.EnvConfigOverride, materialInfo map[string]string, artifact *repository.CiArtifact) client.Event {
	event := impl.eventFactory.Build(util.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util.CD)
	return event
}

func (impl *AppServiceImpl) BuildPayload(overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline,
	envOverride *chartConfig.EnvConfigOverride, materialInfo map[string]string, artifact *repository.CiArtifact) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = pipeline.App.AppName
	payload.PipelineName = pipeline.Name
	payload.EnvName = envOverride.Environment.Name

	var revision string
	for _, v := range materialInfo {
		revision = v
		break
	}
	payload.Source = url.PathEscape(revision)
	payload.DockerImageUrl = artifact.Image
	return payload
}

type ReleaseAttributes struct {
	Name           string
	Tag            string
	PipelineName   string
	ReleaseVersion string
	DeploymentType string
	App            string // App here corresponds to appId
	Env            string // Env here corresponds to envId
	AppMetrics     *bool
}

func NewReleaseAttributes(image, imageTag, pipelineName, deploymentStrategy string, appId, envId, pipelineReleaseCounter int, appMetricsEnabled *bool) *ReleaseAttributes {
	releaseAttribute := &ReleaseAttributes{
		Name:           image,
		Tag:            imageTag,
		PipelineName:   pipelineName,
		ReleaseVersion: strconv.Itoa(pipelineReleaseCounter),
		DeploymentType: deploymentStrategy,
		App:            strconv.Itoa(appId),
		Env:            strconv.Itoa(envId),
		AppMetrics:     appMetricsEnabled,
	}
	return releaseAttribute
}

func (releaseAttributes *ReleaseAttributes) RenderJson(jsonTemplate string) (string, error) {
	override, err := util2.Tprintf(jsonTemplate, releaseAttributes)
	return override, err
}

func (impl *AppServiceImpl) UpdateInstalledAppVersionHistoryByACDObject(app *v1alpha1.Application, installedAppVersionHistoryId int, updateTimedOutStatus bool) error {
	installedAppVersionHistory, err := impl.installedAppVersionHistoryRepository.GetInstalledAppVersionHistory(installedAppVersionHistoryId)
	if err != nil {
		impl.logger.Errorw("error on update installedAppVersionHistory, fetch failed for runner type", "installedAppVersionHistory", installedAppVersionHistoryId, "app", app, "err", err)
		return err
	}
	if updateTimedOutStatus {
		installedAppVersionHistory.Status = pipelineConfig.WorkflowTimedOut
	} else {
		if string(app.Status.Health.Status) == string(health.HealthStatusHealthy) {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
			installedAppVersionHistory.SetFinishedOn()
		} else {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowInProgress
		}
	}
	installedAppVersionHistory.UpdatedBy = 1
	installedAppVersionHistory.UpdatedOn = time.Now()
	_, err = impl.installedAppVersionHistoryRepository.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
	if err != nil {
		impl.logger.Errorw("error on update installedAppVersionHistory", "installedAppVersionHistoryId", installedAppVersionHistoryId, "app", app, "err", err)
		return err
	}
	return nil
}

func (impl *AppServiceImpl) UpdateCdWorkflowRunnerByACDObject(app *v1alpha1.Application, cdWfrId int, updateTimedOutStatus bool) error {
	wfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdWfrId)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner, fetch failed for runner type", "wfr", wfr, "app", app, "err", err)
		return err
	}
	if updateTimedOutStatus {
		wfr.Status = pipelineConfig.WorkflowTimedOut
	} else {
		if string(app.Status.Health.Status) == string(health.HealthStatusHealthy) {
			wfr.Status = pipelineConfig.WorkflowSucceeded
			wfr.FinishedOn = time.Now()
		} else {
			wfr.Status = pipelineConfig.WorkflowInProgress
		}
	}
	wfr.UpdatedBy = 1
	wfr.UpdatedOn = time.Now()
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(wfr)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner", "wfr", wfr, "app", app, "err", err)
		return err
	}
	util2.TriggerCDMetrics(cdWorkflow.GetTriggerMetricsFromRunnerObj(wfr), impl.appStatusConfig.ExposeCDMetrics)
	return nil
}

func (impl *AppServiceImpl) GetActiveCiCdAppsCount() (int, error) {
	return impl.appRepository.GetActiveCiCdAppsCount()
}
