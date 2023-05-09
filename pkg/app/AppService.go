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

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/argoproj/gitops-engine/pkg/sync/common"
	"github.com/caarlos0/env"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	client2 "github.com/devtron-labs/devtron/api/helm-app"
	application3 "github.com/devtron-labs/devtron/client/k8s/application"
	status2 "github.com/devtron-labs/devtron/pkg/app/status"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	chart2 "k8s.io/helm/pkg/proto/hapi/chart"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util3 "github.com/devtron-labs/devtron/pkg/util"

	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	. "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/commonService"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/util"
	util "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	errors2 "github.com/juju/errors"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppServiceConfig struct {
	CdPipelineStatusCronTime            string `env:"CD_PIPELINE_STATUS_CRON_TIME" envDefault:"*/2 * * * *"`
	CdHelmPipelineStatusCronTime        string `env:"CD_HELM_PIPELINE_STATUS_CRON_TIME" envDefault:"*/2 * * * *"`
	CdPipelineStatusTimeoutDuration     string `env:"CD_PIPELINE_STATUS_TIMEOUT_DURATION" envDefault:"20"`                   //in minutes
	PipelineDegradedTime                string `env:"PIPELINE_DEGRADED_TIME" envDefault:"10"`                                //in minutes
	GetPipelineDeployedWithinHours      int    `env:"DEPLOY_STATUS_CRON_GET_PIPELINE_DEPLOYED_WITHIN_HOURS" envDefault:"12"` //in hours
	HelmPipelineStatusCheckEligibleTime string `env:"HELM_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME" envDefault:"120"`             //in seconds
	ExposeCDMetrics                     bool   `env:"EXPOSE_CD_METRICS" envDefault:"false"`
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
	ciArtifactRepository                   repository.CiArtifactRepository
	pipelineRepository                     pipelineConfig.PipelineRepository
	gitFactory                             *GitFactory
	dbMigrationConfigRepository            pipelineConfig.DbMigrationConfigRepository
	eventClient                            client.EventClient
	eventFactory                           client.EventFactory
	acdClient                              application.ServiceClient
	tokenCache                             *util3.TokenCache
	acdAuthConfig                          *util3.ACDAuthConfig
	enforcer                               casbin.Enforcer
	enforcerUtil                           rbac.EnforcerUtil
	user                                   user.UserService
	appListingRepository                   repository.AppListingRepository
	appRepository                          app.AppRepository
	envRepository                          repository2.EnvironmentRepository
	pipelineConfigRepository               chartConfig.PipelineConfigRepository
	configMapRepository                    chartConfig.ConfigMapRepository
	chartRepository                        chartRepoRepository.ChartRepository
	appRepo                                app.AppRepository
	appLevelMetricsRepository              repository.AppLevelMetricsRepository
	envLevelMetricsRepository              repository.EnvLevelAppMetricsRepository
	ciPipelineMaterialRepository           pipelineConfig.CiPipelineMaterialRepository
	cdWorkflowRepository                   pipelineConfig.CdWorkflowRepository
	commonService                          commonService.CommonService
	imageScanDeployInfoRepository          security.ImageScanDeployInfoRepository
	imageScanHistoryRepository             security.ImageScanHistoryRepository
	ArgoK8sClient                          argocdServer.ArgoK8sClient
	pipelineStrategyHistoryService         history2.PipelineStrategyHistoryService
	configMapHistoryService                history2.ConfigMapHistoryService
	deploymentTemplateHistoryService       history2.DeploymentTemplateHistoryService
	chartTemplateService                   ChartTemplateService
	refChartDir                            chartRepoRepository.RefChartDir
	helmAppClient                          client2.HelmAppClient
	helmAppService                         client2.HelmAppService
	chartRefRepository                     chartRepoRepository.ChartRefRepository
	chartService                           chart.ChartService
	argoUserService                        argo.ArgoUserService
	pipelineStatusTimelineRepository       pipelineConfig.PipelineStatusTimelineRepository
	appCrudOperationService                AppCrudOperationService
	configMapHistoryRepository             repository3.ConfigMapHistoryRepository
	strategyHistoryRepository              repository3.PipelineStrategyHistoryRepository
	deploymentTemplateHistoryRepository    repository3.DeploymentTemplateHistoryRepository
	dockerRegistryIpsConfigService         dockerRegistry.DockerRegistryIpsConfigService
	pipelineStatusTimelineResourcesService status2.PipelineStatusTimelineResourcesService
	pipelineStatusSyncDetailService        status2.PipelineStatusSyncDetailService
	pipelineStatusTimelineService          status2.PipelineStatusTimelineService
	appStatusConfig                        *AppServiceConfig
	gitOpsConfigRepository                 repository.GitOpsConfigRepository
	appStatusService                       appStatus.AppStatusService
	installedAppRepository                 repository4.InstalledAppRepository
	AppStoreDeploymentService              service.AppStoreDeploymentService
	k8sApplicationService                  k8s.K8sApplicationService
	installedAppVersionHistoryRepository   repository4.InstalledAppVersionHistoryRepository
	globalEnvVariables                     *util2.GlobalEnvVariables
}

type AppService interface {
	TriggerRelease(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time, triggeredBy int32, wfrId int) (id int, err error)
	UpdateReleaseStatus(request *bean.ReleaseStatusUpdateRequest) (bool, error)
	UpdateDeploymentStatusAndCheckIsSucceeded(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, error)
	TriggerCD(artifact *repository.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error
	GetConfigMapAndSecretJson(appId int, envId int, pipelineId int) ([]byte, error)
	UpdateCdWorkflowRunnerByACDObject(app *v1alpha1.Application, cdWfrId int, updateTimedOutStatus bool) error
	GetCmSecretNew(appId int, envId int) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error)
	MarkImageScanDeployed(appId int, envId int, imageDigest string, clusterId int) error
	GetChartRepoName(gitRepoUrl string) string
	UpdateDeploymentStatusForGitOpsPipelines(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, bool, error)
	WriteCDSuccessEvent(appId int, envId int, override *chartConfig.PipelineOverride)
	GetGitOpsRepoPrefix() string
}

func NewAppService(
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	mergeUtil *MergeUtil,
	logger *zap.SugaredLogger,
	ciArtifactRepository repository.CiArtifactRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	dbMigrationConfigRepository pipelineConfig.DbMigrationConfigRepository,
	eventClient client.EventClient,
	eventFactory client.EventFactory, acdClient application.ServiceClient,
	cache *util3.TokenCache, authConfig *util3.ACDAuthConfig,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, user user.UserService,
	appListingRepository repository.AppListingRepository,
	appRepository app.AppRepository,
	envRepository repository2.EnvironmentRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	appLevelMetricsRepository repository.AppLevelMetricsRepository,
	envLevelMetricsRepository repository.EnvLevelAppMetricsRepository,
	chartRepository chartRepoRepository.ChartRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	commonService commonService.CommonService,
	imageScanDeployInfoRepository security.ImageScanDeployInfoRepository,
	imageScanHistoryRepository security.ImageScanHistoryRepository,
	ArgoK8sClient argocdServer.ArgoK8sClient,
	gitFactory *GitFactory,
	pipelineStrategyHistoryService history2.PipelineStrategyHistoryService,
	configMapHistoryService history2.ConfigMapHistoryService,
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService,
	chartTemplateService ChartTemplateService,
	refChartDir chartRepoRepository.RefChartDir,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	chartService chart.ChartService,
	helmAppClient client2.HelmAppClient,
	argoUserService argo.ArgoUserService,
	cdPipelineStatusTimelineRepo pipelineConfig.PipelineStatusTimelineRepository,
	appCrudOperationService AppCrudOperationService,
	configMapHistoryRepository repository3.ConfigMapHistoryRepository,
	strategyHistoryRepository repository3.PipelineStrategyHistoryRepository,
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository,
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService,
	pipelineStatusTimelineResourcesService status2.PipelineStatusTimelineResourcesService,
	pipelineStatusSyncDetailService status2.PipelineStatusSyncDetailService,
	pipelineStatusTimelineService status2.PipelineStatusTimelineService,
	appStatusConfig *AppServiceConfig,
	gitOpsConfigRepository repository.GitOpsConfigRepository,
	appStatusService appStatus.AppStatusService,
	installedAppRepository repository4.InstalledAppRepository,
	AppStoreDeploymentService service.AppStoreDeploymentService,
	k8sApplicationService k8s.K8sApplicationService,
	installedAppVersionHistoryRepository repository4.InstalledAppVersionHistoryRepository,
	globalEnvVariables *util2.GlobalEnvVariables, helmAppService client2.HelmAppService) *AppServiceImpl {
	appServiceImpl := &AppServiceImpl{
		environmentConfigRepository:            environmentConfigRepository,
		mergeUtil:                              mergeUtil,
		pipelineOverrideRepository:             pipelineOverrideRepository,
		logger:                                 logger,
		ciArtifactRepository:                   ciArtifactRepository,
		pipelineRepository:                     pipelineRepository,
		dbMigrationConfigRepository:            dbMigrationConfigRepository,
		eventClient:                            eventClient,
		eventFactory:                           eventFactory,
		acdClient:                              acdClient,
		tokenCache:                             cache,
		acdAuthConfig:                          authConfig,
		enforcer:                               enforcer,
		enforcerUtil:                           enforcerUtil,
		user:                                   user,
		appListingRepository:                   appListingRepository,
		appRepository:                          appRepository,
		envRepository:                          envRepository,
		pipelineConfigRepository:               pipelineConfigRepository,
		configMapRepository:                    configMapRepository,
		chartRepository:                        chartRepository,
		appLevelMetricsRepository:              appLevelMetricsRepository,
		envLevelMetricsRepository:              envLevelMetricsRepository,
		ciPipelineMaterialRepository:           ciPipelineMaterialRepository,
		cdWorkflowRepository:                   cdWorkflowRepository,
		commonService:                          commonService,
		imageScanDeployInfoRepository:          imageScanDeployInfoRepository,
		imageScanHistoryRepository:             imageScanHistoryRepository,
		ArgoK8sClient:                          ArgoK8sClient,
		gitFactory:                             gitFactory,
		pipelineStrategyHistoryService:         pipelineStrategyHistoryService,
		configMapHistoryService:                configMapHistoryService,
		deploymentTemplateHistoryService:       deploymentTemplateHistoryService,
		chartTemplateService:                   chartTemplateService,
		refChartDir:                            refChartDir,
		chartRefRepository:                     chartRefRepository,
		chartService:                           chartService,
		helmAppClient:                          helmAppClient,
		argoUserService:                        argoUserService,
		pipelineStatusTimelineRepository:       cdPipelineStatusTimelineRepo,
		appCrudOperationService:                appCrudOperationService,
		configMapHistoryRepository:             configMapHistoryRepository,
		strategyHistoryRepository:              strategyHistoryRepository,
		deploymentTemplateHistoryRepository:    deploymentTemplateHistoryRepository,
		dockerRegistryIpsConfigService:         dockerRegistryIpsConfigService,
		pipelineStatusTimelineResourcesService: pipelineStatusTimelineResourcesService,
		pipelineStatusSyncDetailService:        pipelineStatusSyncDetailService,
		pipelineStatusTimelineService:          pipelineStatusTimelineService,
		appStatusConfig:                        appStatusConfig,
		gitOpsConfigRepository:                 gitOpsConfigRepository,
		appStatusService:                       appStatusService,
		installedAppRepository:                 installedAppRepository,
		AppStoreDeploymentService:              AppStoreDeploymentService,
		k8sApplicationService:                  k8sApplicationService,
		installedAppVersionHistoryRepository:   installedAppVersionHistoryRepository,
		globalEnvVariables:                     globalEnvVariables,
		helmAppService:                         helmAppService,
	}
	return appServiceImpl
}

const (
	Success = "SUCCESS"
	Failure = "FAILURE"
)

func (impl *AppServiceImpl) getValuesFileForEnv(environmentId int) string {
	return fmt.Sprintf("_%d-values.yaml", environmentId) //-{envId}-values.yaml
}
func (impl *AppServiceImpl) createArgoApplicationIfRequired(appId int, appName string, envConfigOverride *chartConfig.EnvConfigOverride, pipeline *pipelineConfig.Pipeline, userId int32) (string, error) {
	//repo has been registered while helm create
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("no chart found ", "app", appId)
		return "", err
	}
	envModel, err := impl.envRepository.FindById(envConfigOverride.TargetEnvironment)
	if err != nil {
		return "", err
	}
	argoAppName := pipeline.DeploymentAppName
	if pipeline.DeploymentAppCreated {
		return argoAppName, nil
	} else {
		//create
		appNamespace := envConfigOverride.Namespace
		if appNamespace == "" {
			appNamespace = "default"
		}
		namespace := argocdServer.DevtronInstalationNs
		appRequest := &argocdServer.AppTemplate{
			ApplicationName: argoAppName,
			Namespace:       namespace,
			TargetNamespace: appNamespace,
			TargetServer:    envModel.Cluster.ServerUrl,
			Project:         "default",
			ValuesFile:      impl.getValuesFileForEnv(envModel.Id),
			RepoPath:        chart.ChartLocation,
			RepoUrl:         chart.GitRepoUrl,
		}

		argoAppName, err := impl.ArgoK8sClient.CreateAcdApp(appRequest, envModel.Cluster)
		if err != nil {
			return "", err
		}
		//update cd pipeline to mark deployment app created
		_, err = impl.updatePipeline(pipeline, userId)
		if err != nil {
			impl.logger.Errorw("error in update cd pipeline for deployment app created or not", "err", err)
			return "", err
		}
		return argoAppName, nil
	}
}

func (impl *AppServiceImpl) UpdateReleaseStatus(updateStatusRequest *bean.ReleaseStatusUpdateRequest) (bool, error) {
	count, err := impl.pipelineOverrideRepository.UpdateStatusByRequestIdentifier(updateStatusRequest.RequestId, updateStatusRequest.NewStatus)
	if err != nil {
		impl.logger.Errorw("error in updating release status", "request", updateStatusRequest, "error", err)
		return false, err
	}
	return count == 1, nil
}

func (impl *AppServiceImpl) UpdateDeploymentStatusAndCheckIsSucceeded(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, error) {
	isSucceeded := false
	var err error
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
			return isSucceeded, err
		}
		if installAppDeleteRequest.EnvironmentId > 0 {
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(installAppDeleteRequest.AppId, installAppDeleteRequest.EnvironmentId, string(app.Status.Health.Status))
			if err != nil {
				impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", installAppDeleteRequest.AppId, "envId", installAppDeleteRequest.EnvironmentId)
			}
			impl.logger.Debugw("skipping application status update as this app is chart", "appId", installAppDeleteRequest.AppId, "envId", installAppDeleteRequest.EnvironmentId)
			return isSucceeded, nil
		}
	} else {
		repoUrl := app.Spec.Source.RepoURL
		// backward compatibility for updating application status - if unable to find app check it in charts
		chart, err := impl.chartRepository.FindChartByGitRepoUrl(repoUrl)
		if err != nil {
			impl.logger.Errorw("error in fetching chart", "repoUrl", repoUrl, "err", err)
			return isSucceeded, err
		}
		if chart == nil {
			impl.logger.Errorw("no git repo found for url", "repoUrl", repoUrl)
			return isSucceeded, fmt.Errorf("no git repo found for url %s", repoUrl)
		}
		envId, err := impl.appRepository.FindEnvironmentIdForInstalledApp(chart.AppId)
		if err != nil {
			impl.logger.Errorw("error in fetching app", "err", err, "app", chart.AppId)
			return isSucceeded, err
		}
		if envId > 0 {
			err = impl.appStatusService.UpdateStatusWithAppIdEnvId(chart.AppId, envId, string(app.Status.Health.Status))
			if err != nil {
				impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", chart.AppId, "envId", envId)
			}
			impl.logger.Debugw("skipping application status update as this app is chart", "appId", chart.AppId, "envId", envId)
			return isSucceeded, nil
		}
	}

	isSucceeded, _, err = impl.UpdateDeploymentStatusForGitOpsPipelines(app, statusTime, isAppStore)
	if err != nil {
		impl.logger.Errorw("error in updating deployment status", "argoAppName", app.Name)
		return isSucceeded, err
	}
	return isSucceeded, nil
}

func (impl *AppServiceImpl) UpdateDeploymentStatusForGitOpsPipelines(app *v1alpha1.Application, statusTime time.Time, isAppStore bool) (bool, bool, error) {
	isSucceeded := false
	isTimelineUpdated := false
	isTimelineTimedOut := false
	gitHash := ""
	if app != nil {
		gitHash = app.Status.Sync.Revision
	}
	if !isAppStore {
		isValid, cdPipeline, cdWfr, pipelineOverride, err := impl.CheckIfPipelineUpdateEventIsValid(app.Name, gitHash)
		if err != nil {
			impl.logger.Errorw("service err, CheckIfPipelineUpdateEventIsValid", "err", err)
			return isSucceeded, isTimelineUpdated, err
		}
		if !isValid {
			impl.logger.Infow("deployment status event invalid, skipping", "appName", app.Name)
			return isSucceeded, isTimelineUpdated, nil
		}
		timeoutDuration, err := strconv.Atoi(impl.appStatusConfig.CdPipelineStatusTimeoutDuration)
		if err != nil {
			impl.logger.Errorw("error in converting string to int", "err", err)
			return isSucceeded, isTimelineUpdated, err
		}
		latestTimelineBeforeThisEvent, err := impl.pipelineStatusTimelineRepository.FetchLatestTimelineByWfrId(cdWfr.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline before update", "err", err, "cdWfrId", cdWfr.Id)
			return isSucceeded, isTimelineUpdated, err
		}
		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(cdPipeline.AppId, cdPipeline.EnvironmentId, string(app.Status.Health.Status))
		if err != nil {
			impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", cdPipeline.AppId, "envId", cdPipeline.EnvironmentId)
		}
		reconciledAt := &metav1.Time{}
		if app != nil {
			reconciledAt = app.Status.ReconciledAt
		}
		var kubectlSyncedTimeline *pipelineConfig.PipelineStatusTimeline
		//updating cd pipeline status timeline
		isTimelineUpdated, isTimelineTimedOut, kubectlSyncedTimeline, err = impl.UpdatePipelineStatusTimelineForApplicationChanges(app, cdWfr.Id, statusTime, cdWfr.StartedOn, timeoutDuration, latestTimelineBeforeThisEvent, reconciledAt, false)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline status timeline", "err", err)
		}
		if isTimelineTimedOut {
			//not checking further and directly updating timedOutStatus
			err := impl.UpdateCdWorkflowRunnerByACDObject(app, cdWfr.Id, true)
			if err != nil {
				impl.logger.Errorw("error on update cd workflow runner", "CdWorkflowId", pipelineOverride.CdWorkflowId, "status", pipelineConfig.WorkflowTimedOut, "err", err)
				return isSucceeded, isTimelineUpdated, err
			}
			return isSucceeded, isTimelineUpdated, nil
		}
		if reconciledAt.IsZero() || (kubectlSyncedTimeline != nil && kubectlSyncedTimeline.Id > 0 && reconciledAt.After(kubectlSyncedTimeline.StatusTime)) {
			releaseCounter, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(pipelineOverride.PipelineId)
			if err != nil {
				impl.logger.Errorw("error on update application status", "releaseCounter", releaseCounter, "gitHash", gitHash, "pipelineOverride", pipelineOverride, "err", err)
				return isSucceeded, isTimelineUpdated, err
			}
			if pipelineOverride.PipelineReleaseCounter == releaseCounter {
				isSucceeded, err = impl.UpdateDeploymentStatusForPipeline(app, pipelineOverride, cdWfr.Id)
				if err != nil {
					impl.logger.Errorw("error in updating deployment status for pipeline", "err", err)
					return isSucceeded, isTimelineUpdated, err
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
		isValid, installedAppVersionHistory, appId, envId, err := impl.CheckIfPipelineUpdateEventIsValidForAppStore(app.ObjectMeta.Name)
		if err != nil {
			impl.logger.Errorw("service err, CheckIfPipelineUpdateEventIsValidForAppStore", "err", err)
			return isSucceeded, isTimelineUpdated, err
		}
		if !isValid {
			impl.logger.Infow("deployment status event invalid, skipping", "appName", app.Name)
			return isSucceeded, isTimelineUpdated, nil
		}
		timeoutDuration, err := strconv.Atoi(impl.appStatusConfig.CdPipelineStatusTimeoutDuration)
		if err != nil {
			impl.logger.Errorw("error in converting string to int", "err", err)
			return isSucceeded, isTimelineUpdated, err
		}
		latestTimelineBeforeThisEvent, err := impl.pipelineStatusTimelineRepository.FetchLatestTimelinesByInstalledAppVersionHistoryId(installedAppVersionHistory.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline before update", "err", err, "installedAppVersionHistoryId", installedAppVersionHistory.Id)
			return isSucceeded, isTimelineUpdated, err
		}

		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(appId, envId, string(app.Status.Health.Status))
		if err != nil {
			impl.logger.Errorw("error occurred while updating app status in app_status table", "error", err, "appId", appId, "envId", envId)
		}
		//reconcile time is how often your applications will sync from Argo CD to the Git repository
		reconciledAt := &metav1.Time{}
		if app != nil {
			reconciledAt = app.Status.ReconciledAt
		}
		var kubectlSyncedTimeline *pipelineConfig.PipelineStatusTimeline
		//updating versionHistory pipeline status timeline
		isTimelineUpdated, isTimelineTimedOut, kubectlSyncedTimeline, err = impl.UpdatePipelineStatusTimelineForApplicationChanges(app, installedAppVersionHistory.Id, statusTime, installedAppVersionHistory.StartedOn, timeoutDuration, latestTimelineBeforeThisEvent, reconciledAt, true)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline status timeline", "err", err)
		}
		if isTimelineTimedOut {
			//not checking further and directly updating timedOutStatus
			err := impl.UpdateInstalledAppVersionHistoryByACDObject(app, installedAppVersionHistory.Id, true)
			if err != nil {
				impl.logger.Errorw("error on update installedAppVersionHistory", "installedAppVersionHistory", installedAppVersionHistory.Id, "status", pipelineConfig.WorkflowTimedOut, "err", err)
				return isSucceeded, isTimelineUpdated, err
			}
			return isSucceeded, isTimelineUpdated, nil
		}

		if reconciledAt.IsZero() || (kubectlSyncedTimeline != nil && kubectlSyncedTimeline.Id > 0) {
			isSucceeded, err = impl.UpdateDeploymentStatusForAppStore(app, installedAppVersionHistory.Id)
			if err != nil {
				impl.logger.Errorw("error in updating deployment status for pipeline", "err", err)
				return isSucceeded, isTimelineUpdated, err
			}
			if isSucceeded {
				impl.logger.Infow("writing installed app success event", "gitHash", gitHash, "installedAppVersionHistory", installedAppVersionHistory)
			}
		} else {
			impl.logger.Debugw("event received for older triggered revision", "gitHash", gitHash)
		}
	}

	return isSucceeded, isTimelineUpdated, nil
}

func (impl *AppServiceImpl) CheckIfPipelineUpdateEventIsValidForAppStore(gitOpsAppName string) (bool, *repository4.InstalledAppVersionHistory, int, int, error) {
	isValid := false
	var err error
	installedAppVersionHistory := &repository4.InstalledAppVersionHistory{}
	installedAppId := 0
	gitOpsAppNameAndInstalledAppMapping := make(map[string]*int)
	//checking if the gitOpsAppName is present in installed_apps table, if yes the find installed_app_version_history else return
	gitOpsAppNameAndInstalledAppId, err := impl.installedAppRepository.GetAllGitOpsAppNameAndInstalledAppMapping()
	if err != nil {
		impl.logger.Errorw("error in getting all installed apps in GetAllGitOpsAppNameAndInstalledAppMapping", "err", err, "gitOpsAppName", gitOpsAppName)
		return isValid, installedAppVersionHistory, 0, 0, err
	}
	for _, item := range gitOpsAppNameAndInstalledAppId {
		gitOpsAppNameAndInstalledAppMapping[item.GitOpsAppName] = &item.InstalledAppId
	}
	var devtronAcdAppName string
	if len(impl.globalEnvVariables.GitOpsRepoPrefix) > 0 {
		devtronAcdAppName = fmt.Sprintf("%s-%s", impl.globalEnvVariables.GitOpsRepoPrefix, gitOpsAppName)
	} else {
		devtronAcdAppName = gitOpsAppName
	}

	if gitOpsAppNameAndInstalledAppMapping[devtronAcdAppName] != nil {
		installedAppId = *gitOpsAppNameAndInstalledAppMapping[devtronAcdAppName]
	}

	installedAppVersionHistory, err = impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistoryByInstalledAppId(installedAppId)
	if err != nil {
		impl.logger.Errorw("error in getting latest installedAppVersionHistory by installedAppId", "err", err, "installedAppId", installedAppId)
		return isValid, installedAppVersionHistory, 0, 0, err
	}
	appId, envId, err := impl.installedAppVersionHistoryRepository.GetAppIdAndEnvIdWithInstalledAppVersionId(installedAppVersionHistory.InstalledAppVersionId)
	if err != nil {
		impl.logger.Errorw("error in getting appId and environmentId using installedAppVersionId", "err", err, "installedAppVersionId", installedAppVersionHistory.InstalledAppVersionId)
		return isValid, installedAppVersionHistory, 0, 0, err
	}
	if util2.IsTerminalStatus(installedAppVersionHistory.Status) {
		//drop event
		return isValid, installedAppVersionHistory, appId, envId, nil
	}
	isValid = true
	return isValid, installedAppVersionHistory, appId, envId, err
}

func (impl *AppServiceImpl) CheckIfPipelineUpdateEventIsValid(argoAppName, gitHash string) (bool, pipelineConfig.Pipeline, pipelineConfig.CdWorkflowRunner, *chartConfig.PipelineOverride, error) {
	isValid := false
	var err error
	//var deploymentStatus repository.DeploymentStatus
	var pipeline pipelineConfig.Pipeline
	var pipelineOverride *chartConfig.PipelineOverride
	var cdWfr pipelineConfig.CdWorkflowRunner
	pipeline, err = impl.pipelineRepository.GetArgoPipelineByArgoAppName(argoAppName)
	if err != nil {
		impl.logger.Errorw("error in getting cd pipeline by argoAppName", "err", err, "argoAppName", argoAppName)
		return isValid, pipeline, cdWfr, pipelineOverride, err
	}
	//getting latest pipelineOverride for app (by appId and envId)
	pipelineOverride, err = impl.pipelineOverrideRepository.FindLatestByAppIdAndEnvId(pipeline.AppId, pipeline.EnvironmentId)
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
			//we have received trigger hash which is committed before this apps actual gitHash stored by us
			// this means that the hash stored by us will be synced later, so we will drop this event
			return isValid, pipeline, cdWfr, pipelineOverride, nil
		}
	}
	cdWfr, err = impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), pipelineOverride.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest wfr by pipelineId", "err", err, "pipelineId", pipeline.Id)
		return isValid, pipeline, cdWfr, pipelineOverride, err
	}
	if util2.IsTerminalStatus(cdWfr.Status) {
		//drop event
		return isValid, pipeline, cdWfr, pipelineOverride, nil
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
	if application.Healthy == app.Status.Health.Status {
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
	if application.Healthy == app.Status.Health.Status {
		isSucceeded = true
	}
	return isSucceeded, nil
}

func (impl *AppServiceImpl) UpdatePipelineStatusTimelineForApplicationChanges(app *v1alpha1.Application, pipelineId int,
	statusTime time.Time, triggeredAt time.Time, statusTimeoutDuration int,
	latestTimelineBeforeUpdate *pipelineConfig.PipelineStatusTimeline, reconciledAt *metav1.Time, isAppStore bool) (isTimelineUpdated bool,
	isTimelineTimedOut bool, kubectlApplySyncedTimeline *pipelineConfig.PipelineStatusTimeline, err error) {

	//pipelineId can be wfrId or installedAppVersionHistoryId
	impl.logger.Debugw("updating pipeline status timeline", "app", app, "pipelineOverride", pipelineId, "APP_TO_UPDATE", app.Name)
	isTimelineUpdated = false
	isTimelineTimedOut = false
	if !isAppStore {
		terminalStatusExists, err := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByWfrId(pipelineId)
		if err != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", err, "wfrId", pipelineId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		if terminalStatusExists {
			impl.logger.Infow("terminal status timeline exists for cdWfr, skipping more timeline changes", "wfrId", pipelineId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, nil
		}
		err = impl.pipelineStatusSyncDetailService.SaveOrUpdateSyncDetail(pipelineId, 1)
		if err != nil {
			impl.logger.Errorw("error in save/update pipeline status fetch detail", "err", err, "cdWfrId", pipelineId)
		}
		// creating cd pipeline status timeline
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: pipelineId,
			StatusTime:         statusTime,
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		timeline.Status = pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_STARTED
		if app != nil && app.Status.OperationState != nil {
			timeline.StatusDetail = app.Status.OperationState.Message
		}
		//checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
		_, err, isTimelineUpdated = impl.SavePipelineStatusTimelineIfNotAlreadyPresent(pipelineId, timeline.Status, timeline, false)
		if err != nil {
			impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		//saving timeline resource details
		err = impl.pipelineStatusTimelineResourcesService.SaveOrUpdatePipelineTimelineResources(pipelineId, app, nil, 1, false)
		if err != nil {
			impl.logger.Errorw("error in saving/updating timeline resources", "err", err, "cdWfrId", pipelineId)
		}
		var kubectlSyncTimelineFetchErr error
		kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr = impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatus(pipelineId, pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED)
		if kubectlSyncTimelineFetchErr != nil && kubectlSyncTimelineFetchErr != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline", "err", kubectlSyncTimelineFetchErr, "cdWfrId", pipelineId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr
		}
		if (kubectlApplySyncedTimeline == nil || kubectlApplySyncedTimeline.Id == 0) && app != nil && app.Status.OperationState != nil && app.Status.OperationState.Phase == common.OperationSucceeded {
			timeline.Id = 0
			timeline.Status = pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED
			timeline.StatusDetail = app.Status.OperationState.Message
			//checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
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
			if app.Status.Health.Status == health.HealthStatusHealthy {
				impl.logger.Infow("updating pipeline status timeline for healthy app", "app", app, "APP_TO_UPDATE", app.Name)
				haveNewTimeline = true
				timeline.Status = pipelineConfig.TIMELINE_STATUS_APP_HEALTHY
				timeline.StatusDetail = "App status is Healthy."
			}
			if haveNewTimeline {
				//not checking if this status is already present or not because already checked for terminal status existence earlier
				err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
				if err != nil {
					impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
					return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
				}
				isTimelineUpdated = true
				impl.logger.Debugw("APP_STATUS_UPDATE_REQ", "stage", "terminal_status", "app", app, "status", timeline.Status)
			}
		}

		if !isTimelineUpdated {
			//no timeline updated since before, in this case we will check for timeout cases
			var lastTimeToCheckForTimeout time.Time
			if latestTimelineBeforeUpdate == nil {
				lastTimeToCheckForTimeout = triggeredAt
			} else {
				lastTimeToCheckForTimeout = latestTimelineBeforeUpdate.StatusTime
			}
			if time.Since(lastTimeToCheckForTimeout) >= time.Duration(statusTimeoutDuration)*time.Minute {
				//mark as timed out if not already marked
				timeline.Status = pipelineConfig.TIMELINE_STATUS_FETCH_TIMED_OUT
				timeline.StatusDetail = "Deployment timed out."
				_, err, isTimelineUpdated = impl.SavePipelineStatusTimelineIfNotAlreadyPresent(pipelineId, timeline.Status, timeline, false)
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
		terminalStatusExists, err := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(pipelineId)
		if err != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", pipelineId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		if terminalStatusExists {
			impl.logger.Infow("terminal status timeline exists for installed App, skipping more timeline changes", "installedAppVersionHistoryId", pipelineId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, nil
		}
		err = impl.pipelineStatusSyncDetailService.SaveOrUpdateSyncDetailForAppStore(pipelineId, 1)
		if err != nil {
			impl.logger.Errorw("error in save/update pipeline status fetch detail", "err", err, "installedAppVersionHistoryId", pipelineId)
		}
		// creating installedAppVersionHistory status timeline
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: pipelineId,
			StatusTime:                   statusTime,
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		timeline.Status = pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_STARTED
		if app != nil && app.Status.OperationState != nil {
			timeline.StatusDetail = app.Status.OperationState.Message
		}
		//checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
		_, err, isTimelineUpdated = impl.SavePipelineStatusTimelineIfNotAlreadyPresent(pipelineId, timeline.Status, timeline, true)
		if err != nil {
			impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
		}
		//saving timeline resource details
		err = impl.pipelineStatusTimelineResourcesService.SaveOrUpdatePipelineTimelineResources(pipelineId, app, nil, 1, true)
		if err != nil {
			impl.logger.Errorw("error in saving/updating timeline resources", "err", err, "installedAppVersionId", pipelineId)
		}
		var kubectlSyncTimelineFetchErr error
		kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr = impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(pipelineId, pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED)
		if kubectlSyncTimelineFetchErr != nil && kubectlSyncTimelineFetchErr != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline", "err", kubectlSyncTimelineFetchErr, "installedAppVersionHistoryId", pipelineId)
			return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, kubectlSyncTimelineFetchErr
		}
		if (kubectlApplySyncedTimeline == nil || kubectlApplySyncedTimeline.Id == 0) && app != nil && app.Status.OperationState != nil && app.Status.OperationState.Phase == common.OperationSucceeded {
			timeline.Id = 0
			timeline.Status = pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED
			timeline.StatusDetail = app.Status.OperationState.Message
			//checking and saving if this timeline is present or not because kubewatch may stream same objects multiple times
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, true)
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
			if app.Status.Health.Status == health.HealthStatusHealthy {
				impl.logger.Infow("updating pipeline status timeline for healthy app", "app", app, "APP_TO_UPDATE", app.Name)
				haveNewTimeline = true
				timeline.Status = pipelineConfig.TIMELINE_STATUS_APP_HEALTHY
				timeline.StatusDetail = "App status is Healthy."
			}
			if haveNewTimeline {
				//not checking if this status is already present or not because already checked for terminal status existence earlier
				err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, true)
				if err != nil {
					impl.logger.Errorw("error in creating timeline status", "err", err, "timeline", timeline)
					return isTimelineUpdated, isTimelineTimedOut, kubectlApplySyncedTimeline, err
				}
				isTimelineUpdated = true
				impl.logger.Debugw("APP_STATUS_UPDATE_REQ", "stage", "terminal_status", "app", app, "status", timeline.Status)
			}
		}

		if !isTimelineUpdated {
			//no timeline updated since before, in this case we will check for timeout cases
			var lastTimeToCheckForTimeout time.Time
			if latestTimelineBeforeUpdate == nil {
				lastTimeToCheckForTimeout = triggeredAt
			} else {
				lastTimeToCheckForTimeout = latestTimelineBeforeUpdate.StatusTime
			}
			if time.Since(lastTimeToCheckForTimeout) >= time.Duration(statusTimeoutDuration)*time.Minute {
				//mark as timed out if not already marked
				timeline.Status = pipelineConfig.TIMELINE_STATUS_FETCH_TIMED_OUT
				timeline.StatusDetail = "Deployment timed out."
				_, err, isTimelineUpdated = impl.SavePipelineStatusTimelineIfNotAlreadyPresent(pipelineId, timeline.Status, timeline, true)
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

func (impl *AppServiceImpl) SavePipelineStatusTimelineIfNotAlreadyPresent(pipelineId int, timelineStatus pipelineConfig.TimelineStatus, timeline *pipelineConfig.PipelineStatusTimeline, isAppStore bool) (latestTimeline *pipelineConfig.PipelineStatusTimeline, err error, isTimelineUpdated bool) {
	isTimelineUpdated = false
	if isAppStore {
		latestTimeline, err = impl.pipelineStatusTimelineRepository.FetchTimelineByInstalledAppVersionHistoryIdAndStatus(pipelineId, timelineStatus)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting latest timeline", "err", err)
			return nil, err, isTimelineUpdated
		} else if err == pg.ErrNoRows {
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, true)
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
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
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

type EnvironmentOverride struct {
	Enabled   bool        `json:"enabled"`
	EnvValues []*KeyValue `json:"envValues"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (conf *EnvironmentOverride) appendEnvironmentVariable(key, value string) {
	item := &KeyValue{Key: key, Value: value}
	conf.EnvValues = append(conf.EnvValues, item)
}

func (impl *AppServiceImpl) TriggerCD(artifact *repository.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	impl.logger.Debugw("automatic pipeline trigger attempt async", "artifactId", artifact.Id)

	return impl.triggerReleaseAsync(artifact, cdWorkflowId, wfrId, pipeline, triggeredAt)
}

func (impl *AppServiceImpl) triggerReleaseAsync(artifact *repository.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	err := impl.validateAndTrigger(pipeline, artifact, cdWorkflowId, wfrId, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in trigger for pipeline", "pipelineId", strconv.Itoa(pipeline.Id))
	}
	impl.logger.Debugw("trigger attempted for all pipeline ", "artifactId", artifact.Id)
	return err
}

func (impl *AppServiceImpl) validateAndTrigger(p *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	object := impl.enforcerUtil.GetAppRBACNameByAppId(p.AppId)
	envApp := strings.Split(object, "/")
	if len(envApp) != 2 {
		impl.logger.Error("invalid req, app and env not found from rbac")
		return errors.New("invalid req, app and env not found from rbac")
	}
	err := impl.releasePipeline(p, artifact, cdWorkflowId, wfrId, triggeredAt)
	return err
}

func (impl *AppServiceImpl) releasePipeline(pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	impl.logger.Debugw("triggering release for ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id)
	//Iterate for each even if there is error in one
	request := &bean.ValuesOverrideRequest{
		PipelineId:           pipeline.Id,
		UserId:               artifact.CreatedBy,
		CiArtifactId:         artifact.Id,
		AppId:                pipeline.AppId,
		CdWorkflowId:         cdWorkflowId,
		ForceTrigger:         true,
		DeploymentWithConfig: bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
	}

	ctx, err := impl.buildACDContext()
	if err != nil {
		impl.logger.Errorw("error in creating acd synch context", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
		return err
	}
	//setting deployedBy as 1(system user) since case of auto trigger
	id, err := impl.TriggerRelease(request, ctx, triggeredAt, 1, wfrId)
	if err != nil {
		impl.logger.Errorw("error in auto  cd pipeline trigger", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
	} else {
		impl.logger.Infow("pipeline successfully triggered ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id, "releaseId", id)
	}
	return err
}

func (impl *AppServiceImpl) buildACDContext() (acdContext context.Context, err error) {
	//this method should only call in case of argo-integration and gitops configured
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return nil, err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	return ctx, nil
}

func (impl *AppServiceImpl) getDbMigrationOverride(overrideRequest *bean.ValuesOverrideRequest, artifact *repository.CiArtifact, isRollback bool) (overrideJson []byte, err error) {
	if isRollback {
		return nil, fmt.Errorf("rollback not supported ye")
	}
	notConfigured := false
	config, err := impl.dbMigrationConfigRepository.FindByPipelineId(overrideRequest.PipelineId)
	if err != nil && !IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching pipeline override config", "req", overrideRequest, "err", err)
		return nil, err
	} else if IsErrNoRows(err) {
		notConfigured = true
	}
	envVal := &EnvironmentOverride{}
	if notConfigured {
		impl.logger.Warnw("no active db migration found", "pipeline", overrideRequest.PipelineId)
		envVal.Enabled = false
	} else {
		materialInfos, err := artifact.ParseMaterialInfo()
		if err != nil {
			return nil, err
		}

		hash, ok := materialInfos[config.GitMaterial.Url]
		if !ok {
			impl.logger.Errorf("wrong url map ", "map", materialInfos, "url", config.GitMaterial.Url)
			return nil, fmt.Errorf("configured url not found in material %s", config.GitMaterial.Url)
		}

		envVal.Enabled = true
		if config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_USERNAME_PASSWORD &&
			config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_ACCESS_TOKEN &&
			config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_ANONYMOUS {
			return nil, fmt.Errorf("auth mode %s not supported for migration", config.GitMaterial.GitProvider.AuthMode)
		}
		envVal.appendEnvironmentVariable("GIT_REPO_URL", config.GitMaterial.Url)
		envVal.appendEnvironmentVariable("GIT_USER", config.GitMaterial.GitProvider.UserName)
		var password string
		if config.GitMaterial.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
			password = config.GitMaterial.GitProvider.Password
		} else {
			password = config.GitMaterial.GitProvider.AccessToken
		}
		envVal.appendEnvironmentVariable("GIT_AUTH_TOKEN", password)
		// parse git-tag not required
		//envVal.appendEnvironmentVariable("GIT_TAG", "")
		envVal.appendEnvironmentVariable("GIT_HASH", hash)
		envVal.appendEnvironmentVariable("SCRIPT_LOCATION", config.ScriptSource)
		envVal.appendEnvironmentVariable("DB_TYPE", string(config.DbConfig.Type))
		envVal.appendEnvironmentVariable("DB_USER_NAME", config.DbConfig.UserName)
		envVal.appendEnvironmentVariable("DB_PASSWORD", config.DbConfig.Password)
		envVal.appendEnvironmentVariable("DB_HOST", config.DbConfig.Host)
		envVal.appendEnvironmentVariable("DB_PORT", config.DbConfig.Port)
		envVal.appendEnvironmentVariable("DB_NAME", config.DbConfig.DbName)
		//Will be used for rollback don't delete it
		//envVal.appendEnvironmentVariable("MIGRATE_TO_VERSION", strconv.Itoa(overrideRequest.TargetDbVersion))
	}
	dbMigrationConfig := map[string]interface{}{"dbMigrationConfig": envVal}
	confByte, err := json.Marshal(dbMigrationConfig)
	if err != nil {
		return nil, err
	}
	return confByte, nil
}

func (impl *AppServiceImpl) TriggerRelease(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time, deployedBy int32, wfrId int) (id int, err error) {
	if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	if len(overrideRequest.DeploymentWithConfig) == 0 {
		overrideRequest.DeploymentWithConfig = bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineRepository.FindById")
	pipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	span.End()
	if err != nil {
		impl.logger.Errorw("invalid req", "err", err, "req", overrideRequest)
		return 0, err
	}
	envOverride := &chartConfig.EnvConfigOverride{}
	var appMetrics *bool
	strategy := &chartConfig.PipelineStrategy{}
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		_, span := otel.Tracer("orchestrator").Start(ctx, "deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId")
		deploymentTemplateHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting deployed deployment template history by pipelineId and wfrId", "err", err, "pipelineId", &overrideRequest, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return 0, err
		}
		templateName := deploymentTemplateHistory.TemplateName
		templateVersion := deploymentTemplateHistory.TemplateVersion
		if templateName == "Rollout Deployment" {
			templateName = ""
		}
		//getting chart_ref by id
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartRefRepository.FindByVersionAndName")
		chartRef, err := impl.chartRefRepository.FindByVersionAndName(templateName, templateVersion)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting chartRef by version and name", "err", err, "version", templateVersion, "name", templateName)
			return 0, err
		}
		//assuming that if a chartVersion is deployed then it's envConfigOverride will be available
		_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.GetByAppIdEnvIdAndChartRefId")
		envOverride, err = impl.environmentConfigRepository.GetByAppIdEnvIdAndChartRefId(pipeline.AppId, pipeline.EnvironmentId, chartRef.Id)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting envConfigOverride for pipeline for specific chartVersion", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId, "chartRefId", chartRef.Id)
			return 0, err
		}
		//updating historical data in envConfigOverride and appMetrics flag
		envOverride.IsOverride = true
		envOverride.EnvOverrideValues = deploymentTemplateHistory.Template
		appMetrics = &deploymentTemplateHistory.IsAppMetricsEnabled
		_, span = otel.Tracer("orchestrator").Start(ctx, "strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId")
		strategyHistory, err := impl.strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting deployed strategy history by pipleinId and wfrId", "err", err, "pipelineId", overrideRequest.PipelineId, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return 0, err
		}
		strategy.Strategy = strategyHistory.Strategy
		strategy.Config = strategyHistory.Config
		strategy.PipelineId = pipeline.Id
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.ActiveEnvConfigOverride")
		envOverride, err = impl.environmentConfigRepository.ActiveEnvConfigOverride(overrideRequest.AppId, pipeline.EnvironmentId)
		span.End()
		if err != nil {
			impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
			return 0, err
		}
		if envOverride.Id == 0 {
			_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
			chart, err := impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
			span.End()
			if err != nil {
				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
				return 0, err
			}
			_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId")
			envOverride, err = impl.environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId(overrideRequest.AppId, pipeline.EnvironmentId, chart.ChartRefId)
			span.End()
			if err != nil && !errors2.IsNotFound(err) {
				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
				return 0, err
			}

			//creating new env override config
			if errors2.IsNotFound(err) || envOverride == nil {
				_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
				environment, err := impl.envRepository.FindById(pipeline.EnvironmentId)
				span.End()
				if err != nil && !IsErrNoRows(err) {
					return 0, err
				}
				envOverride = &chartConfig.EnvConfigOverride{
					Active:            true,
					ManualReviewed:    true,
					Status:            models.CHARTSTATUS_SUCCESS,
					TargetEnvironment: pipeline.EnvironmentId,
					ChartId:           chart.Id,
					AuditLog:          sql.AuditLog{UpdatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId},
					Namespace:         environment.Namespace,
					IsOverride:        false,
					EnvOverrideValues: "{}",
					Latest:            false,
					IsBasicViewLocked: chart.IsBasicViewLocked,
					CurrentViewEditor: chart.CurrentViewEditor,
				}
				_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.Save")
				err = impl.environmentConfigRepository.Save(envOverride)
				span.End()
				if err != nil {
					impl.logger.Errorw("error in creating envconfig", "data", envOverride, "error", err)
					return 0, err
				}
			}
			envOverride.Chart = chart
		} else if envOverride.Id > 0 && !envOverride.IsOverride {
			_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
			chart, err := impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
			span.End()
			if err != nil {
				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
				return 0, err
			}
			envOverride.Chart = chart
		}

		_, span = otel.Tracer("orchestrator").Start(ctx, "appLevelMetricsRepository.FindByAppId")
		appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(pipeline.AppId)
		span.End()
		if err != nil && !IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return 0, &ApiError{InternalMessage: "unable to fetch app level metrics flag"}
		}
		appMetrics = &appLevelMetrics.AppMetrics

		_, span = otel.Tracer("orchestrator").Start(ctx, "envLevelMetricsRepository.FindByAppIdAndEnvId")
		envLevelMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(pipeline.AppId, pipeline.EnvironmentId)
		span.End()
		if err != nil && !IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return 0, &ApiError{InternalMessage: "unable to fetch env level metrics flag"}
		}
		if envLevelMetrics.Id != 0 && envLevelMetrics.AppMetrics != nil {
			appMetrics = envLevelMetrics.AppMetrics
		}
		//fetch pipeline config from strategy table, if pipeline is automatic fetch always default, else depends on request

		//forceTrigger true if CD triggered Auto, triggered occurred from CI
		if overrideRequest.ForceTrigger {
			_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
			strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
			span.End()
		} else {
			var deploymentTemplate chartRepoRepository.DeploymentStrategy
			if overrideRequest.DeploymentTemplate == "ROLLING" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_ROLLING
			} else if overrideRequest.DeploymentTemplate == "BLUE-GREEN" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_BLUE_GREEN
			} else if overrideRequest.DeploymentTemplate == "CANARY" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_CANARY
			} else if overrideRequest.DeploymentTemplate == "RECREATE" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_RECREATE
			}

			if len(deploymentTemplate) > 0 {
				_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.FindByStrategyAndPipelineId")
				strategy, err = impl.pipelineConfigRepository.FindByStrategyAndPipelineId(deploymentTemplate, overrideRequest.PipelineId)
				span.End()
			} else {
				_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
				strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
				span.End()
			}
		}
		if err != nil && errors2.IsNotFound(err) == false {
			impl.logger.Errorf("invalid state", "err", err, "req", strategy)
			return 0, err
		}
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "CreateHistoriesForDeploymentTrigger")
	err = impl.CreateHistoriesForDeploymentTrigger(pipeline, strategy, envOverride, envOverride.Chart.ImageDescriptorTemplate, triggeredAt, deployedBy)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in creating history entries for deployment trigger", "err", err)
		return 0, err
	}

	// auto-healing :  data corruption fix - if ChartLocation in chart is not correct, need correction
	if !strings.HasSuffix(envOverride.Chart.ChartLocation, fmt.Sprintf("%s%s", "/", envOverride.Chart.ChartVersion)) {
		_, span = otel.Tracer("orchestrator").Start(ctx, "autoHealChartLocationInChart")
		err = impl.autoHealChartLocationInChart(ctx, envOverride)
		span.End()
		if err != nil {
			return 0, err
		}
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
	env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
	span.End()
	if err != nil {
		impl.logger.Errorw("unable to find env", "err", err)
		return 0, err
	}
	envOverride.Environment = env

	// CHART COMMIT and PUSH STARTS HERE, it will push latest version, if found modified on deployment template and overrides
	chartMetaData := &chart2.Metadata{
		Name:    pipeline.App.AppName,
		Version: envOverride.Chart.ChartVersion,
	}
	referenceTemplatePath := path.Join(string(impl.refChartDir), envOverride.Chart.ReferenceTemplate)
	if IsAcdApp(pipeline.DeploymentAppType) {
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.GetGitOpsRepoName")
		// CHART COMMIT and PUSH STARTS HERE, it will push latest version, if found modified on deployment template and overrides
		gitOpsRepoName := impl.chartTemplateService.GetGitOpsRepoName(pipeline.App.AppName)
		span.End()
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartService.CheckChartExists")
		err = impl.chartService.CheckChartExists(envOverride.Chart.ChartRefId)
		span.End()
		if err != nil {
			impl.logger.Errorw("err in getting chart info", "err", err)
			return 0, err
		}
		var gitCommitStatus pipelineConfig.TimelineStatus
		var gitCommitStatusDetail string
		err = impl.buildChartAndPushToGitRepo(overrideRequest, ctx, chartMetaData, referenceTemplatePath, wfrId, gitOpsRepoName, envOverride)
		if err != nil {
			impl.saveTimelineForError(overrideRequest, ctx, err, wfrId)
			return 0, err
		} else {
			gitCommitStatus = pipelineConfig.TIMELINE_STATUS_GIT_COMMIT
			gitCommitStatusDetail = "Git commit done successfully."
			// creating cd pipeline status timeline for git commit
			timeline := &pipelineConfig.PipelineStatusTimeline{
				CdWorkflowRunnerId: wfrId,
				Status:             gitCommitStatus,
				StatusDetail:       gitCommitStatusDetail,
				StatusTime:         time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: overrideRequest.UserId,
					CreatedOn: time.Now(),
					UpdatedBy: overrideRequest.UserId,
					UpdatedOn: time.Now(),
				},
			}
			_, span = otel.Tracer("orchestrator").Start(ctx, "cdPipelineStatusTimelineRepo.SaveTimeline")
			err := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
			span.End()
			if err != nil {
				impl.logger.Errorw("error in creating timeline status for git commit", "err", err, "timeline", timeline)
			}
		}

		// ACD app creation STARTS HERE, it will use existing if already created
		impl.logger.Debugw("new pipeline found", "pipeline", pipeline)
		_, span = otel.Tracer("orchestrator").Start(ctx, "createArgoApplicationIfRequired")
		name, err := impl.createArgoApplicationIfRequired(overrideRequest.AppId, pipeline.App.AppName, envOverride, pipeline, overrideRequest.UserId)
		span.End()
		if err != nil {
			impl.logger.Errorw("acd application create error on cd trigger", "err", err, "req", overrideRequest)
			return 0, err
		}
		impl.logger.Debugw("argocd application created", "name", name)
		// ENDS HERE
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
	artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
	span.End()
	if err != nil {
		return 0, err
	}
	materialInfoMap, mErr := artifact.ParseMaterialInfo()
	if mErr != nil {
		impl.logger.Errorw("material info map error", mErr)
		return 0, err
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "getDbMigrationOverride")
	//FIXME: how to determine rollback
	//we can't depend on ciArtifact ID because CI pipeline can be manually triggered in any order regardless of sourcecode status
	dbMigrationOverride, err := impl.getDbMigrationOverride(overrideRequest, artifact, false)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching db migration config", "req", overrideRequest, "err", err)
		return 0, err
	}
	chartVersion := envOverride.Chart.ChartVersion
	_, span = otel.Tracer("orchestrator").Start(ctx, "getConfigMapAndSecretJsonV2")
	configMapJson, err := impl.getConfigMapAndSecretJsonV2(overrideRequest.AppId, envOverride.TargetEnvironment, overrideRequest.PipelineId, chartVersion, overrideRequest.DeploymentWithConfig, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching config map n secret ", "err", err)
		configMapJson = nil
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "appCrudOperationService.GetLabelsByAppIdForDeployment")
	appLabelJsonByte, err := impl.appCrudOperationService.GetLabelsByAppIdForDeployment(overrideRequest.AppId)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching app labels for gitOps commit", "err", err)
		appLabelJsonByte = nil
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "mergeAndSave")
	releaseId, pipelineOverrideId, mergeAndSave, saveErr := impl.mergeAndSave(envOverride, overrideRequest, dbMigrationOverride, artifact, pipeline, configMapJson, appLabelJsonByte, strategy, ctx, triggeredAt, deployedBy, appMetrics)
	span.End()
	if releaseId != 0 {
		//updating the acd app with updated values and sync operation
		if IsAcdApp(pipeline.DeploymentAppType) {
			_, span = otel.Tracer("orchestrator").Start(ctx, "updateArgoPipeline")
			updateAppInArgocd, err := impl.updateArgoPipeline(overrideRequest.AppId, pipeline.Name, envOverride, ctx)
			span.End()
			if err != nil {
				impl.logger.Errorw("error in updating argocd app ", "err", err)
				return 0, err
			}
			if updateAppInArgocd {
				impl.logger.Debug("argo-cd successfully updated")
			} else {
				impl.logger.Debug("argo-cd failed to update, ignoring it")
			}
			//	impl.synchCD(pipeline, ctx, overrideRequest, envOverride)
		}
		//for helm type cd pipeline, create install helm application, update deployment status, update workflow runner for app detail status.
		if IsHelmApp(pipeline.DeploymentAppType) {
			_, span = otel.Tracer("orchestrator").Start(ctx, "createHelmAppForCdPipeline")
			_, err = impl.createHelmAppForCdPipeline(overrideRequest, envOverride, referenceTemplatePath, chartMetaData, triggeredAt, pipeline, mergeAndSave, ctx)
			span.End()
			if err != nil {
				impl.logger.Errorw("error in creating or updating helm application for cd pipeline", "err", err)
				return 0, err
			}
		}

		go impl.WriteCDTriggerEvent(overrideRequest, pipeline, envOverride, materialInfoMap, artifact, releaseId, pipelineOverrideId)

		if artifact.ScanEnabled {
			_, span = otel.Tracer("orchestrator").Start(ctx, "MarkImageScanDeployed")
			_ = impl.MarkImageScanDeployed(overrideRequest.AppId, envOverride.TargetEnvironment, artifact.ImageDigest, pipeline.Environment.ClusterId)
			span.End()
		}
	}
	middleware.CdTriggerCounter.WithLabelValues(pipeline.App.AppName, pipeline.Environment.Name).Inc()
	return releaseId, saveErr
}

func (impl *AppServiceImpl) buildChartAndPushToGitRepo(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, chartMetaData *chart2.Metadata, referenceTemplatePath string, wfrId int, gitOpsRepoName string, envOverride *chartConfig.EnvConfigOverride) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.BuildChart")
	tempReferenceTemplateDir, err := impl.chartTemplateService.BuildChart(ctx, chartMetaData, referenceTemplatePath)
	span.End()
	defer impl.chartTemplateService.CleanDir(tempReferenceTemplateDir)
	if err != nil {
		return err
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.PushChartToGitRepo")
	err = impl.chartTemplateService.PushChartToGitRepo(gitOpsRepoName, envOverride.Chart.ReferenceTemplate, envOverride.Chart.ChartVersion, tempReferenceTemplateDir, envOverride.Chart.GitRepoUrl, overrideRequest.UserId)
	span.End()
	return err
}

func (impl *AppServiceImpl) saveTimelineForError(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, err error, wfrId int) {
	impl.logger.Errorw("Ref chart commit error on cd trigger", "err", err, "req", overrideRequest)
	gitCommitStatus := pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED
	gitCommitStatusDetail := fmt.Sprintf("Git commit failed - %v", err)
	// creating cd pipeline status timeline for git commit
	timeline := &pipelineConfig.PipelineStatusTimeline{
		CdWorkflowRunnerId: wfrId,
		Status:             gitCommitStatus,
		StatusDetail:       gitCommitStatusDetail,
		StatusTime:         time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: overrideRequest.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: overrideRequest.UserId,
			UpdatedOn: time.Now(),
		},
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "cdPipelineStatusTimelineRepo.SaveTimeline")
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
	span.End()
	if timelineErr != nil {
		impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
	}
}

func (impl *AppServiceImpl) autoHealChartLocationInChart(ctx context.Context, envOverride *chartConfig.EnvConfigOverride) error {
	chartId := envOverride.Chart.Id
	impl.logger.Infow("auto-healing: Chart location in chart not correct. modifying ", "chartId", chartId,
		"current chartLocation", envOverride.Chart.ChartLocation, "current chartVersion", envOverride.Chart.ChartVersion)

	// get chart from DB (getting it from DB because envOverride.Chart does not have full row of DB)
	_, span := otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindById")
	chart, err := impl.chartRepository.FindById(chartId)
	span.End()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching chart from DB", "chartId", chartId, "err", err)
		return err
	}

	// get chart ref from DB (to get location)
	chartRefId := chart.ChartRefId
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartRefRepository.FindById")
	chartRef, err := impl.chartRefRepository.FindById(chartRefId)
	span.End()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching chartRef from DB", "chartRefId", chartRefId, "err", err)
		return err
	}

	// build new chart location
	newChartLocation := filepath.Join(chartRef.Location, envOverride.Chart.ChartVersion)
	impl.logger.Infow("new chart location build", "chartId", chartId, "newChartLocation", newChartLocation)

	// update chart in DB
	chart.ChartLocation = newChartLocation
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.Update")
	err = impl.chartRepository.Update(chart)
	span.End()
	if err != nil {
		impl.logger.Errorw("error occurred while saving chart into DB", "chartId", chartId, "err", err)
		return err
	}

	// update newChartLocation in model
	envOverride.Chart.ChartLocation = newChartLocation
	return nil
}

func (impl *AppServiceImpl) MarkImageScanDeployed(appId int, envId int, imageDigest string, clusterId int) error {
	impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "imageDigest", imageDigest)
	executionHistory, err := impl.imageScanHistoryRepository.FindByImageDigest(imageDigest)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching execution history", "err", err)
		return err
	}
	if executionHistory == nil || executionHistory.Id == 0 {
		impl.logger.Errorw("no execution history found for digest", "digest", imageDigest)
		return fmt.Errorf("no execution history found for digest - %s", imageDigest)
	}
	impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "executionHistory", executionHistory)
	var ids []int
	ids = append(ids, executionHistory.Id)

	ot, err := impl.imageScanDeployInfoRepository.FindByTypeMetaAndTypeId(appId, security.ScanObjectType_APP) //todo insure this touple unique in db
	if err != nil && err != pg.ErrNoRows {
		return err
	} else if err == pg.ErrNoRows {
		imageScanDeployInfo := &security.ImageScanDeployInfo{
			ImageScanExecutionHistoryId: ids,
			ScanObjectMetaId:            appId,
			ObjectType:                  security.ScanObjectType_APP,
			EnvId:                       envId,
			ClusterId:                   clusterId,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: 1,
				UpdatedOn: time.Now(),
				UpdatedBy: 1,
			},
		}
		impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "imageScanDeployInfo", imageScanDeployInfo)
		err = impl.imageScanDeployInfoRepository.Save(imageScanDeployInfo)
		if err != nil {
			impl.logger.Errorw("error in creating deploy info", "err", err)
		}
	} else {
		impl.logger.Debugw("pt", "ot", ot)
	}
	return err
}

// FIXME tmp workaround
func (impl *AppServiceImpl) GetCmSecretNew(appId int, envId int) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error) {
	var configMapJson string
	var secretDataJson string
	var configMapJsonApp string
	var secretDataJsonApp string
	var configMapJsonEnv string
	var secretDataJsonEnv string
	//var configMapJsonPipeline string
	//var secretDataJsonPipeline string

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

	chart, err := impl.commonService.FetchLatestChart(appId, envId)
	if err != nil {
		return nil, nil, err
	}
	chartVersion := chart.ChartVersion
	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, nil, err
	}
	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion)
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
	return &configResponse, &secretResponse, nil
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

func (impl *AppServiceImpl) getConfigMapAndSecretJsonV2(appId int, envId int, pipelineId int, chartVersion string, deploymentWithConfig bean.DeploymentConfigurationType, wfrIdForDeploymentWithSpecificTrigger int) ([]byte, error) {

	var configMapJson string
	var secretDataJson string
	var configMapJsonApp string
	var secretDataJsonApp string
	var configMapJsonEnv string
	var secretDataJsonEnv string
	var err error
	//var configMapJsonPipeline string
	//var secretDataJsonPipeline string

	merged := []byte("{}")
	if deploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
		if err != nil && pg.ErrNoRows != err {
			return []byte("{}"), err
		}
		if configMapA != nil && configMapA.Id > 0 {
			configMapJsonApp = configMapA.ConfigMapData
			secretDataJsonApp = configMapA.SecretData
		}
		configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
		if err != nil && pg.ErrNoRows != err {
			return []byte("{}"), err
		}
		if configMapE != nil && configMapE.Id > 0 {
			configMapJsonEnv = configMapE.ConfigMapData
			secretDataJsonEnv = configMapE.SecretData
		}
	} else if deploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		//fetching history and setting envLevelConfig and not appLevelConfig because history already contains merged appLevel and envLevel configs
		configMapHistory, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrIdForDeploymentWithSpecificTrigger, repository3.CONFIGMAP_TYPE)
		if err != nil {
			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", pipelineId, "wfrid", wfrIdForDeploymentWithSpecificTrigger)
			return []byte("{}"), err
		}
		configMapJsonEnv = configMapHistory.Data
		secretHistory, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrIdForDeploymentWithSpecificTrigger, repository3.SECRET_TYPE)
		if err != nil {
			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", pipelineId, "wfrid", wfrIdForDeploymentWithSpecificTrigger)
			return []byte("{}"), err
		}
		secretDataJsonEnv = secretHistory.Data
	}
	configMapJson, err = impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
	if err != nil {
		return []byte("{}"), err
	}
	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return []byte("{}"), err
	}
	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion)
	if err != nil {
		return []byte("{}"), err
	}
	configResponseR := bean.ConfigMapRootJson{}
	configResponse := bean.ConfigMapJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(configMapJson), &configResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	configResponseR.ConfigMapJson = configResponse
	secretResponseR := bean.ConfigSecretRootJson{}
	secretResponse := bean.ConfigSecretJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	secretResponseR.ConfigSecretJson = secretResponse

	configMapByte, err := json.Marshal(configResponseR)
	if err != nil {
		return []byte("{}"), err
	}
	secretDataByte, err := json.Marshal(secretResponseR)
	if err != nil {
		return []byte("{}"), err
	}

	merged, err = impl.mergeUtil.JsonPatch(configMapByte, secretDataByte)
	if err != nil {
		return []byte("{}"), err
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

func (impl *AppServiceImpl) WriteCDTriggerEvent(overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline,
	envOverride *chartConfig.EnvConfigOverride, materialInfoMap map[string]string, artifact *repository.CiArtifact, releaseId, pipelineOverrideId int) {
	event := impl.eventFactory.Build(util.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util.CD)
	impl.logger.Debugw("event WriteCDTriggerEvent", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, nil, pipelineOverrideId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	deploymentEvent := DeploymentEvent{
		ApplicationId:      pipeline.AppId,
		EnvironmentId:      pipeline.EnvironmentId, //check for production Environment
		ReleaseId:          releaseId,
		PipelineOverrideId: pipelineOverrideId,
		TriggerTime:        time.Now(),
		CiArtifactId:       overrideRequest.CiArtifactId,
	}
	ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(artifact.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in ")
	}
	for _, ciPipelineMaterial := range ciPipelineMaterials {
		hash := materialInfoMap[ciPipelineMaterial.GitMaterial.Url]
		pipelineMaterialInfo := &PipelineMaterialInfo{PipelineMaterialId: ciPipelineMaterial.Id, CommitHash: hash}
		deploymentEvent.PipelineMaterials = append(deploymentEvent.PipelineMaterials, pipelineMaterialInfo)
	}
	impl.logger.Infow("triggering deployment event", "event", deploymentEvent)
	err = impl.eventClient.WriteNatsEvent(pubsub.CD_SUCCESS, deploymentEvent)
	if err != nil {
		impl.logger.Errorw("error in writing cd trigger event", "err", err)
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
	App            string
	Env            string
	AppMetrics     *bool
}

func (impl *AppServiceImpl) getReleaseOverride(envOverride *chartConfig.EnvConfigOverride,
	overrideRequest *bean.ValuesOverrideRequest,
	artifact *repository.CiArtifact,
	pipeline *pipelineConfig.Pipeline,
	pipelineOverride *chartConfig.PipelineOverride, strategy *chartConfig.PipelineStrategy, appMetrics *bool) (releaseOverride string, err error) {

	artifactImage := artifact.Image
	imageTag := strings.Split(artifactImage, ":")

	imageTagLen := len(imageTag)

	imageName := ""

	for i := 0; i < imageTagLen-1; i++ {
		if i != imageTagLen-2 {
			imageName = imageName + imageTag[i] + ":"
		} else {
			imageName = imageName + imageTag[i]
		}
	}

	appId := strconv.Itoa(pipeline.App.Id)
	envId := strconv.Itoa(pipeline.EnvironmentId)

	deploymentStrategy := ""
	if strategy != nil {
		deploymentStrategy = string(strategy.Strategy)
	}
	releaseAttribute := ReleaseAttributes{
		Name:           imageName,
		Tag:            imageTag[imageTagLen-1],
		PipelineName:   pipeline.Name,
		ReleaseVersion: strconv.Itoa(pipelineOverride.PipelineReleaseCounter),
		DeploymentType: deploymentStrategy,
		App:            appId,
		Env:            envId,
		AppMetrics:     appMetrics,
	}
	override, err := util2.Tprintf(envOverride.Chart.ImageDescriptorTemplate, releaseAttribute)
	if err != nil {
		return "", &ApiError{InternalMessage: "unable to render ImageDescriptorTemplate"}
	}
	if overrideRequest.AdditionalOverride != nil {
		userOverride, err := overrideRequest.AdditionalOverride.MarshalJSON()
		if err != nil {
			return "", err
		}
		data, err := impl.mergeUtil.JsonPatch(userOverride, []byte(override))
		if err != nil {
			return "", err
		}
		override = string(data)
	}
	return override, nil
}

func (impl *AppServiceImpl) GetChartRepoName(gitRepoUrl string) string {
	gitRepoUrl = gitRepoUrl[strings.LastIndex(gitRepoUrl, "/")+1:]
	chartRepoName := strings.ReplaceAll(gitRepoUrl, ".git", "")
	return chartRepoName
}

func (impl *AppServiceImpl) mergeAndSave(envOverride *chartConfig.EnvConfigOverride,
	overrideRequest *bean.ValuesOverrideRequest,
	dbMigrationOverride []byte,
	artifact *repository.CiArtifact,
	pipeline *pipelineConfig.Pipeline, configMapJson, appLabelJsonByte []byte, strategy *chartConfig.PipelineStrategy, ctx context.Context,
	triggeredAt time.Time, deployedBy int32, appMetrics *bool) (releaseId int, overrideId int, mergedValues string, err error) {

	//register release , obtain release id TODO: populate releaseId to template
	override, err := impl.savePipelineOverride(overrideRequest, envOverride.Id, triggeredAt)
	if err != nil {
		return 0, 0, "", err
	}
	//TODO: check status and apply lock
	overrideJson, err := impl.getReleaseOverride(envOverride, overrideRequest, artifact, pipeline, override, strategy, appMetrics)
	if err != nil {
		return 0, 0, "", err
	}

	//merge three values on the fly
	//ordering is important here
	//global < environment < db< release
	var merged []byte
	if !envOverride.IsOverride {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.Chart.GlobalOverride))
		if err != nil {
			return 0, 0, "", err
		}
	} else {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.EnvOverrideValues))
		if err != nil {
			return 0, 0, "", err
		}
	}

	//pipeline override here comes from pipeline strategy table
	if strategy != nil && len(strategy.Config) > 0 {
		merged, err = impl.mergeUtil.JsonPatch(merged, []byte(strategy.Config))
		if err != nil {
			return 0, 0, "", err
		}
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, dbMigrationOverride)
	if err != nil {
		return 0, 0, "", err
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, []byte(overrideJson))
	if err != nil {
		return 0, 0, "", err
	}

	if configMapJson != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, configMapJson)
		if err != nil {
			return 0, 0, "", err
		}
	}

	if appLabelJsonByte != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, appLabelJsonByte)
		if err != nil {
			return 0, 0, "", err
		}
	}

	appName := fmt.Sprintf("%s-%s", pipeline.App.AppName, envOverride.Environment.Name)
	merged = impl.autoscalingCheckBeforeTrigger(ctx, appName, envOverride.Namespace, merged, pipeline, overrideRequest)

	_, span := otel.Tracer("orchestrator").Start(ctx, "dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment")
	// handle image pull secret if access given
	merged, err = impl.dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment(envOverride.Environment, pipeline.CiPipelineId, merged)
	span.End()
	if err != nil {
		return 0, 0, "", err
	}

	commitHash := ""
	commitTime := time.Time{}
	if IsAcdApp(pipeline.DeploymentAppType) {
		chartRepoName := impl.GetChartRepoName(envOverride.Chart.GitRepoUrl)
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit")
		//getting username & emailId for commit author data
		userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(overrideRequest.UserId)
		span.End()
		chartGitAttr := &ChartConfig{
			FileName:       fmt.Sprintf("_%d-values.yaml", envOverride.TargetEnvironment),
			FileContent:    string(merged),
			ChartName:      envOverride.Chart.ChartName,
			ChartLocation:  envOverride.Chart.ChartLocation,
			ChartRepoName:  chartRepoName,
			ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", override.Id, envOverride.TargetEnvironment),
			UserName:       userName,
			UserEmailId:    userEmailId,
		}
		gitOpsConfigBitbucket, err := impl.gitOpsConfigRepository.GetGitOpsConfigByProvider(BITBUCKET_PROVIDER)
		if err != nil {
			if err == pg.ErrNoRows {
				gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
			} else {
				return 0, 0, "", err
			}
		}
		gitOpsConfig := &bean.GitOpsConfigDto{BitBucketWorkspaceId: gitOpsConfigBitbucket.BitBucketWorkspaceId}
		_, span = otel.Tracer("orchestrator").Start(ctx, "gitFactory.Client.CommitValues")
		commitHash, commitTime, err = impl.gitFactory.Client.CommitValues(chartGitAttr, gitOpsConfig)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in git commit", "err", err)
			return 0, 0, "", err
		}
	}
	if commitTime.IsZero() {
		commitTime = time.Now()
	}
	pipelineOverride := &chartConfig.PipelineOverride{
		Id:                     override.Id,
		GitHash:                commitHash,
		CommitTime:             commitTime,
		EnvConfigOverrideId:    envOverride.Id,
		PipelineOverrideValues: overrideJson,
		PipelineId:             overrideRequest.PipelineId,
		CiArtifactId:           overrideRequest.CiArtifactId,
		PipelineMergedValues:   string(merged),
		AuditLog:               sql.AuditLog{UpdatedOn: triggeredAt, UpdatedBy: deployedBy},
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineOverrideRepository.Update")
	err = impl.pipelineOverrideRepository.Update(pipelineOverride)
	span.End()
	if err != nil {
		return 0, 0, "", err
	}
	mergedValues = string(merged)
	return override.PipelineReleaseCounter, override.Id, mergedValues, nil
}

func (impl *AppServiceImpl) savePipelineOverride(overrideRequest *bean.ValuesOverrideRequest, envOverrideId int, triggeredAt time.Time) (override *chartConfig.PipelineOverride, err error) {
	currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(overrideRequest.PipelineId)
	if err != nil {
		return nil, err
	}
	po := &chartConfig.PipelineOverride{
		EnvConfigOverrideId:    envOverrideId,
		Status:                 models.CHARTSTATUS_NEW,
		PipelineId:             overrideRequest.PipelineId,
		CiArtifactId:           overrideRequest.CiArtifactId,
		PipelineReleaseCounter: currentReleaseNo + 1,
		CdWorkflowId:           overrideRequest.CdWorkflowId,
		AuditLog:               sql.AuditLog{CreatedBy: overrideRequest.UserId, CreatedOn: triggeredAt, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
		DeploymentType:         overrideRequest.DeploymentType,
	}

	err = impl.pipelineOverrideRepository.Save(po)
	if err != nil {
		return nil, err
	}
	err = impl.checkAndFixDuplicateReleaseNo(po)
	if err != nil {
		impl.logger.Errorw("error in checking release no duplicacy", "pipeline", po, "err", err)
		return nil, err
	}
	return po, nil
}

func (impl *AppServiceImpl) checkAndFixDuplicateReleaseNo(override *chartConfig.PipelineOverride) error {

	uniqueVerified := false
	retryCount := 0

	for !uniqueVerified && retryCount < 5 {
		retryCount = retryCount + 1
		overrides, err := impl.pipelineOverrideRepository.GetByPipelineIdAndReleaseNo(override.PipelineId, override.PipelineReleaseCounter)
		if err != nil {
			return err
		}
		if overrides[0].Id == override.Id {
			uniqueVerified = true
		} else {
			//duplicate might be due to concurrency, lets fix it
			currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(override.PipelineId)
			if err != nil {
				return err
			}
			override.PipelineReleaseCounter = currentReleaseNo + 1
			err = impl.pipelineOverrideRepository.Save(override)
			if err != nil {
				return err
			}
		}
	}
	if !uniqueVerified {
		return fmt.Errorf("duplicate verification retry count exide max overrideId: %d ,count: %d", override.Id, retryCount)
	}
	return nil
}

func (impl *AppServiceImpl) updateArgoPipeline(appId int, pipelineName string, envOverride *chartConfig.EnvConfigOverride, ctx context.Context) (bool, error) {
	//repo has been registered while helm create
	if ctx == nil {
		impl.logger.Errorw("err in syncing ACD, ctx is NULL", "pipelineName", pipelineName)
		return false, nil
	}
	app, err := impl.appRepository.FindById(appId)
	if err != nil {
		impl.logger.Errorw("no app found ", "err", err)
		return false, err
	}
	envModel, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
	if err != nil {
		return false, err
	}
	argoAppName := fmt.Sprintf("%s-%s", app.AppName, envModel.Name)
	impl.logger.Infow("received payload, updateArgoPipeline", "appId", appId, "pipelineName", pipelineName, "envId", envOverride.TargetEnvironment, "argoAppName", argoAppName, "context", ctx)
	application, err := impl.acdClient.Get(ctx, &application2.ApplicationQuery{Name: &argoAppName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "app", argoAppName, "pipeline", pipelineName)
		return false, err
	}
	//if status, ok:=status.FromError(err);ok{
	appStatus, _ := status.FromError(err)

	if appStatus.Code() == codes.OK {
		impl.logger.Debugw("argo app exists", "app", argoAppName, "pipeline", pipelineName)
		if application.Spec.Source.Path != envOverride.Chart.ChartLocation || application.Spec.Source.TargetRevision != "master" {
			patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: v1alpha1.ApplicationSource{Path: envOverride.Chart.ChartLocation, RepoURL: envOverride.Chart.GitRepoUrl, TargetRevision: "master"}}}
			reqbyte, err := json.Marshal(patchReq)
			if err != nil {
				impl.logger.Errorw("error in creating patch", "err", err)
			}
			reqString := string(reqbyte)
			patchType := "merge"
			_, err = impl.acdClient.Patch(ctx, &application2.ApplicationPatchRequest{Patch: &reqString, Name: &argoAppName, PatchType: &patchType})
			if err != nil {
				impl.logger.Errorw("error in creating argo pipeline ", "name", pipelineName, "patch", string(reqbyte), "err", err)
				return false, err
			}
			impl.logger.Debugw("pipeline update req ", "res", patchReq)
		} else {
			impl.logger.Debug("pipeline no need to update ")
		}
		return true, nil
	} else if appStatus.Code() == codes.NotFound {
		impl.logger.Errorw("argo app not found", "app", argoAppName, "pipeline", pipelineName)
		return false, nil
	} else {
		impl.logger.Errorw("err in checking application on gocd", "err", err, "pipeline", pipelineName)
		return false, err
	}
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
		if app.Status.Health.Status == health.HealthStatusHealthy {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
			installedAppVersionHistory.FinishedOn = time.Now()
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
		if app.Status.Health.Status == health.HealthStatusHealthy {
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
	cdMetrics := util2.CDMetrics{
		AppName:         wfr.CdWorkflow.Pipeline.DeploymentAppName,
		Status:          wfr.Status,
		DeploymentType:  wfr.CdWorkflow.Pipeline.DeploymentAppType,
		EnvironmentName: wfr.CdWorkflow.Pipeline.Environment.Name,
		Time:            time.Since(wfr.StartedOn).Seconds() - time.Since(wfr.FinishedOn).Seconds(),
	}
	util2.TriggerCDMetrics(cdMetrics, impl.appStatusConfig.ExposeCDMetrics)
	return nil
}

func (impl *AppServiceImpl) autoscalingCheckBeforeTrigger(ctx context.Context, appName string, namespace string, merged []byte, pipeline *pipelineConfig.Pipeline, overrideRequest *bean.ValuesOverrideRequest) []byte {
	var appId = pipeline.AppId
	pipelineId := pipeline.Id
	var appDeploymentType = pipeline.DeploymentAppType
	var clusterId = pipeline.Environment.ClusterId
	deploymentType := overrideRequest.DeploymentType
	templateMap := make(map[string]interface{})
	err := json.Unmarshal(merged, &templateMap)
	if err != nil {
		return merged
	}
	if _, ok := templateMap[autoscaling.ServiceName]; ok {
		as := templateMap[autoscaling.ServiceName]
		asd := as.(map[string]interface{})
		isEnable := false
		if _, ok := asd["enabled"]; ok {
			isEnable = asd["enabled"].(bool)
		}
		if isEnable {
			reqReplicaCount := templateMap["replicaCount"].(float64)
			reqMaxReplicas := asd["MaxReplicas"].(float64)
			reqMinReplicas := asd["MinReplicas"].(float64)
			version := ""
			group := autoscaling.ServiceName
			kind := "HorizontalPodAutoscaler"
			resourceName := fmt.Sprintf("%s-%s", appName, "hpa")
			resourceManifest := make(map[string]interface{})
			if IsAcdApp(appDeploymentType) {
				query := &application2.ApplicationResourceRequest{
					Name:         &appName,
					Version:      &version,
					Group:        &group,
					Kind:         &kind,
					ResourceName: &resourceName,
					Namespace:    &namespace,
				}
				recv, err := impl.acdClient.GetResource(ctx, query)
				impl.logger.Debugw("resource manifest get replica count", "response", recv)
				if err != nil {
					impl.logger.Errorw("ACD Get Resource API Failed", "err", err)
					middleware.AcdGetResourceCounter.WithLabelValues(strconv.Itoa(appId), namespace, appName).Inc()
					return merged
				}
				if recv != nil && len(*recv.Manifest) > 0 {
					err := json.Unmarshal([]byte(*recv.Manifest), &resourceManifest)
					if err != nil {
						impl.logger.Errorw("unmarshal failed for hpa check", "err", err)
						return merged
					}
				}
			} else {
				version = "v2beta2"
				k8sResource, err := impl.k8sApplicationService.GetResource(ctx, &k8s.ResourceRequestBean{ClusterId: clusterId,
					K8sRequest: &application3.K8sRequestBean{ResourceIdentifier: application3.ResourceIdentifier{Name: resourceName,
						Namespace: namespace, GroupVersionKind: schema.GroupVersionKind{Group: group, Kind: kind, Version: version}}}})
				if err != nil {
					impl.logger.Errorw("error occurred while fetching resource for app", "resourceName", resourceName, "err", err)
					return merged
				}
				resourceManifest = k8sResource.Manifest.Object
			}
			if len(resourceManifest) > 0 {
				statusMap := resourceManifest["status"].(map[string]interface{})
				currentReplicaVal := statusMap["currentReplicas"]
				currentReplicaCount, err := util2.ParseFloatNumber(currentReplicaVal)
				if err != nil {
					impl.logger.Errorw("error occurred while parsing replica count", "currentReplicas", currentReplicaVal, "err", err)
					return merged
				}

				reqReplicaCount = impl.fetchRequiredReplicaCount(currentReplicaCount, reqMaxReplicas, reqMinReplicas)
				templateMap["replicaCount"] = reqReplicaCount
				merged, err = json.Marshal(&templateMap)
				if err != nil {
					impl.logger.Errorw("marshaling failed for hpa check", "err", err)
					return merged
				}
			}
		} else {
			impl.logger.Errorw("autoscaling is not enabled", "pipelineId", pipelineId)
		}
	}
	//check for custom chart support
	if autoscalingEnabledPath, ok := templateMap[bean2.CustomAutoScalingEnabledPathKey]; ok {
		if deploymentType == models.DEPLOYMENTTYPE_STOP {
			merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoScalingEnabledPathKey, merged, false)
			if err != nil {
				return merged
			}
			merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, 0)
			if err != nil {
				return merged
			}
		} else {
			autoscalingEnabled := false
			autoscalingEnabledValue := gjson.Get(string(merged), autoscalingEnabledPath.(string)).Value()
			if val, ok := autoscalingEnabledValue.(bool); ok {
				autoscalingEnabled = val
			}
			if autoscalingEnabled {
				// extract replica count, min, max and check for required value
				replicaCount, err := impl.getReplicaCountFromCustomChart(templateMap, merged)
				if err != nil {
					return merged
				}
				merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, replicaCount)
				if err != nil {
					return merged
				}
			}
		}
	}

	return merged
}

func (impl *AppServiceImpl) getReplicaCountFromCustomChart(templateMap map[string]interface{}, merged []byte) (float64, error) {
	autoscalingMinVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingMinPathKey, merged)
	if err != nil {
		return 0, err
	}
	autoscalingMaxVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingMaxPathKey, merged)
	if err != nil {
		return 0, err
	}
	autoscalingReplicaCountVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged)
	if err != nil {
		return 0, err
	}
	return impl.fetchRequiredReplicaCount(autoscalingReplicaCountVal, autoscalingMaxVal, autoscalingMinVal), nil
}

func (impl *AppServiceImpl) extractParamValue(inputMap map[string]interface{}, key string, merged []byte) (float64, error) {
	if _, ok := inputMap[key]; !ok {
		return 0, errors.New("empty-val-err")
	}
	floatNumber, err := util2.ParseFloatNumber(gjson.Get(string(merged), inputMap[key].(string)).Value())
	if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", key, "err", err)
	}
	return floatNumber, err
}

func (impl *AppServiceImpl) setScalingValues(templateMap map[string]interface{}, customScalingKey string, merged []byte, value interface{}) ([]byte, error) {
	autoscalingJsonPath := templateMap[customScalingKey]
	autoscalingJsonPathKey := autoscalingJsonPath.(string)
	mergedRes, err := sjson.Set(string(merged), autoscalingJsonPathKey, value)
	if err != nil {
		impl.logger.Errorw("error occurred while setting autoscaling key", "JsonPathKey", autoscalingJsonPathKey, "err", err)
		return []byte{}, err
	}
	return []byte(mergedRes), nil
}

func (impl *AppServiceImpl) fetchRequiredReplicaCount(currentReplicaCount float64, reqMaxReplicas float64, reqMinReplicas float64) float64 {
	var reqReplicaCount float64
	if currentReplicaCount <= reqMaxReplicas && currentReplicaCount >= reqMinReplicas {
		reqReplicaCount = currentReplicaCount
	} else if currentReplicaCount > reqMaxReplicas {
		reqReplicaCount = reqMaxReplicas
	} else if currentReplicaCount < reqMinReplicas {
		reqReplicaCount = reqMinReplicas
	}
	return reqReplicaCount
}

func (impl *AppServiceImpl) CreateHistoriesForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, strategy *chartConfig.PipelineStrategy, envOverride *chartConfig.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) error {
	//creating history for deployment template
	err := impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryForDeploymentTrigger(pipeline, envOverride, renderedImageTemplate, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating deployment template history for deployment trigger", "err", err)
		return err
	}
	err = impl.configMapHistoryService.CreateCMCSHistoryForDeploymentTrigger(pipeline, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating CM/CS history for deployment trigger", "err", err)
		return err
	}
	if strategy != nil {
		err = impl.pipelineStrategyHistoryService.CreateStrategyHistoryForDeploymentTrigger(strategy, deployedOn, deployedBy, pipeline.TriggerType)
		if err != nil {
			impl.logger.Errorw("error in creating strategy history for deployment trigger", "err", err)
			return err
		}
	}
	return nil
}

func (impl *AppServiceImpl) updatePipeline(pipeline *pipelineConfig.Pipeline, userId int32) (bool, error) {
	pipeline.DeploymentAppCreated = true
	pipeline.UpdatedOn = time.Now()
	pipeline.UpdatedBy = userId
	err := impl.pipelineRepository.UpdateCdPipeline(pipeline)
	if err != nil {
		impl.logger.Errorw("error on updating cd pipeline for setting deployment app created", "err", err)
		return false, err
	}
	return true, nil
}

func (impl *AppServiceImpl) createHelmAppForCdPipeline(overrideRequest *bean.ValuesOverrideRequest,
	envOverride *chartConfig.EnvConfigOverride, referenceTemplatePath string, chartMetaData *chart2.Metadata,
	triggeredAt time.Time, pipeline *pipelineConfig.Pipeline, mergeAndSave string, ctx context.Context) (bool, error) {
	if IsHelmApp(pipeline.DeploymentAppType) {
		referenceChartByte := envOverride.Chart.ReferenceChart
		// here updating reference chart into database.
		if len(envOverride.Chart.ReferenceChart) == 0 {
			refChartByte, err := impl.chartTemplateService.GetByteArrayRefChart(chartMetaData, referenceTemplatePath)
			if err != nil {
				impl.logger.Errorw("ref chart commit error on cd trigger", "err", err, "req", overrideRequest)
				return false, err
			}
			ch := envOverride.Chart
			ch.ReferenceChart = refChartByte
			ch.UpdatedOn = time.Now()
			ch.UpdatedBy = overrideRequest.UserId
			err = impl.chartRepository.Update(ch)
			if err != nil {
				impl.logger.Errorw("chart update error", "err", err, "req", overrideRequest)
				return false, err
			}
			referenceChartByte = refChartByte
		}

		releaseName := pipeline.DeploymentAppName
		bearerToken := envOverride.Environment.Cluster.Config["bearer_token"]

		releaseIdentifier := &client2.ReleaseIdentifier{
			ReleaseName:      releaseName,
			ReleaseNamespace: envOverride.Namespace,
			ClusterConfig: &client2.ClusterConfig{
				ClusterName:  envOverride.Environment.Cluster.ClusterName,
				Token:        bearerToken,
				ApiServerUrl: envOverride.Environment.Cluster.ServerUrl,
			},
		}

		if pipeline.DeploymentAppCreated {
			req := &client2.UpgradeReleaseRequest{
				ReleaseIdentifier: releaseIdentifier,
				ValuesYaml:        mergeAndSave,
				HistoryMax:        impl.helmAppService.GetRevisionHistoryMaxValue(client2.SOURCE_DEVTRON_APP),
			}

			updateApplicationResponse, err := impl.helmAppClient.UpdateApplication(ctx, req)

			// For cases where helm release was not found but db flag for deployment app created was true
			if err != nil && strings.Contains(err.Error(), "release: not found") {

				// retry install
				_, err = impl.helmInstallReleaseWithCustomChart(ctx, releaseIdentifier, referenceChartByte, mergeAndSave)

				// if retry failed, return
				if err != nil {
					impl.logger.Errorw("release not found, failed to re-install helm application", "err", err)
					return false, err
				}
			} else if err != nil {
				impl.logger.Errorw("error in updating helm application for cd pipeline", "err", err)
				return false, err
			} else {
				impl.logger.Debugw("updated helm application", "response", updateApplicationResponse, "isSuccess", updateApplicationResponse.Success)
			}

		} else {

			helmResponse, err := impl.helmInstallReleaseWithCustomChart(ctx, releaseIdentifier, referenceChartByte, mergeAndSave)

			// For connection related errors, no need to update the db
			if err != nil && strings.Contains(err.Error(), "connection error") {
				impl.logger.Errorw("error in helm install custom chart", "err", err)
				return false, err
			}

			// IMP: update cd pipeline to mark deployment app created, even if helm install fails
			// If the helm install fails, it still creates the app in failed state, so trying to
			// re-create the app results in error from helm that cannot re-use name which is still in use
			_, pgErr := impl.updatePipeline(pipeline, overrideRequest.UserId)

			if err != nil {
				impl.logger.Errorw("error in helm install custom chart", "err", err)

				if pgErr != nil {
					impl.logger.Errorw("failed to update deployment app created flag in pipeline table", "err", err)
				}
				return false, err
			}

			if pgErr != nil {
				impl.logger.Errorw("failed to update deployment app created flag in pipeline table", "err", err)
				return false, err
			}

			impl.logger.Debugw("received helm release response", "helmResponse", helmResponse, "isSuccess", helmResponse.Success)
		}

		//update workflow runner status, used in app workflow view
		cdWf, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("err on fetching cd workflow", "err", err)
			return false, err
		}
		cdWorkflowId := cdWf.CdWorkflowId
		if cdWf.CdWorkflowId == 0 {
			cdWf := &pipelineConfig.CdWorkflow{
				CiArtifactId: overrideRequest.CiArtifactId,
				PipelineId:   overrideRequest.PipelineId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			err := impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
			if err != nil {
				impl.logger.Errorw("err on updating cd workflow for status update", "err", err)
				return false, err
			}
			cdWorkflowId = cdWf.Id
			runner := &pipelineConfig.CdWorkflowRunner{
				Id:           cdWf.Id,
				Name:         pipeline.Name,
				WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
				ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
				Status:       pipelineConfig.WorkflowInProgress,
				TriggeredBy:  overrideRequest.UserId,
				StartedOn:    triggeredAt,
				CdWorkflowId: cdWorkflowId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
			if err != nil {
				impl.logger.Errorw("err on updating cd workflow runner for status update", "err", err)
				return false, err
			}
		} else {
			cdWf.Status = pipelineConfig.WorkflowInProgress
			cdWf.FinishedOn = time.Now()
			cdWf.UpdatedBy = overrideRequest.UserId
			cdWf.UpdatedOn = time.Now()
			err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&cdWf)
			if err != nil {
				impl.logger.Errorw("error on update cd workflow runner", "cdWf", cdWf, "err", err)
				return false, err
			}
		}
	}
	return true, nil
}

// helmInstallReleaseWithCustomChart performs helm install with custom chart
func (impl *AppServiceImpl) helmInstallReleaseWithCustomChart(ctx context.Context, releaseIdentifier *client2.ReleaseIdentifier, referenceChartByte []byte, valuesYaml string) (*client2.HelmInstallCustomResponse, error) {

	helmInstallRequest := client2.HelmInstallCustomRequest{
		ValuesYaml:        valuesYaml,
		ChartContent:      &client2.ChartContent{Content: referenceChartByte},
		ReleaseIdentifier: releaseIdentifier,
	}

	// Request exec
	return impl.helmAppClient.InstallReleaseWithCustomChart(ctx, &helmInstallRequest)
}

func (impl *AppServiceImpl) GetGitOpsRepoPrefix() string {
	return impl.globalEnvVariables.GitOpsRepoPrefix
}
