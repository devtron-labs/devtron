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

package telemetry

import (
	"context"
	"encoding/json"
	cloudProviderIdentifier "github.com/devtron-labs/common-lib/cloud-provider-identifier"
	util2 "github.com/devtron-labs/common-lib/utils/k8s"
	client "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	posthogClient "github.com/devtron-labs/devtron/client/telemetry/posthog"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	"github.com/devtron-labs/devtron/pkg/auth/sso"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/read"
	repository3 "github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	ciConfig "github.com/devtron-labs/devtron/pkg/build/pipeline/read"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	ucidService "github.com/devtron-labs/devtron/pkg/ucid"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	cron3 "github.com/devtron-labs/devtron/util/cron"
	"net/http"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type TelemetryEventClientImplExtended struct {
	environmentService            environment.EnvironmentService
	appListingRepository          repository.AppListingRepository
	ciPipelineConfigReadService   ciConfig.CiPipelineConfigReadService
	pipelineRepository            pipelineConfig.PipelineRepository
	gitProviderRepository         repository3.GitProviderRepository
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository
	appRepository                 app.AppRepository
	ciWorkflowRepository          pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	gitMaterialReadService        read.GitMaterialReadService
	ciTemplateRepository          pipelineConfig.CiTemplateRepository
	chartRepository               chartRepoRepository.ChartRepository
	ciBuildConfigService          pipeline.CiBuildConfigService
	gitOpsConfigReadService       config.GitOpsConfigReadService
	*TelemetryEventClientImpl
}

func NewTelemetryEventClientImplExtended(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *util2.K8sServiceImpl, aCDAuthConfig *util3.ACDAuthConfig,
	environmentService environment.EnvironmentService, userService user2.UserService,
	appListingRepository repository.AppListingRepository, posthog *posthogClient.PosthogClient, ucid ucidService.Service,
	ciPipelineConfigReadService ciConfig.CiPipelineConfigReadService, pipelineRepository pipelineConfig.PipelineRepository,
	gitProviderRepository repository3.GitProviderRepository, attributeRepo repository.AttributesRepository,
	ssoLoginService sso.SSOLoginService, appRepository app.AppRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository,
	gitMaterialReadService read.GitMaterialReadService, ciTemplateRepository pipelineConfig.CiTemplateRepository,
	chartRepository chartRepoRepository.ChartRepository, userAuditService user2.UserAuditService,
	ciBuildConfigService pipeline.CiBuildConfigService, moduleRepository moduleRepo.ModuleRepository, serverDataStore *serverDataStore.ServerDataStore,
	helmAppClient client.HelmAppClient, installedAppReadService installedAppReader.InstalledAppReadService, userAttributesRepository repository.UserAttributesRepository,
	cloudProviderIdentifierService cloudProviderIdentifier.ProviderIdentifierService, cronLogger *cron3.CronLoggerImpl,
	gitOpsConfigReadService config.GitOpsConfigReadService) (*TelemetryEventClientImplExtended, error) {

	cron := cron.New(
		cron.WithChain(cron.Recover(cronLogger)))
	cron.Start()
	watcher := &TelemetryEventClientImplExtended{
		environmentService:            environmentService,
		appListingRepository:          appListingRepository,
		ciPipelineConfigReadService:   ciPipelineConfigReadService,
		pipelineRepository:            pipelineRepository,
		gitProviderRepository:         gitProviderRepository,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
		appRepository:                 appRepository,
		cdWorkflowRepository:          cdWorkflowRepository,
		ciWorkflowRepository:          ciWorkflowRepository,
		gitMaterialReadService:        gitMaterialReadService,
		ciTemplateRepository:          ciTemplateRepository,
		chartRepository:               chartRepository,
		ciBuildConfigService:          ciBuildConfigService,
		gitOpsConfigReadService:       gitOpsConfigReadService,
		TelemetryEventClientImpl: &TelemetryEventClientImpl{
			cron:                           cron,
			logger:                         logger,
			client:                         client,
			clusterService:                 clusterService,
			K8sUtil:                        K8sUtil,
			aCDAuthConfig:                  aCDAuthConfig,
			userService:                    userService,
			attributeRepo:                  attributeRepo,
			ssoLoginService:                ssoLoginService,
			posthogClient:                  posthog,
			ucid:                           ucid,
			moduleRepository:               moduleRepository,
			serverDataStore:                serverDataStore,
			userAuditService:               userAuditService,
			helmAppClient:                  helmAppClient,
			installedAppReadService:        installedAppReadService,
			userAttributesRepository:       userAttributesRepository,
			cloudProviderIdentifierService: cloudProviderIdentifierService,
			telemetryConfig:                TelemetryConfig{},
		},
	}

	watcher.HeartbeatEventForTelemetry()
	_, err := cron.AddFunc(posthogClient.SummaryCronExpr, watcher.SummaryEventForTelemetry)
	if err != nil {
		logger.Errorw("error in starting summery event", "err", err)
		return nil, err
	}

	_, err = cron.AddFunc(posthogClient.HeartbeatCronExpr, watcher.HeartbeatEventForTelemetry)
	if err != nil {
		logger.Errorw("error in starting heartbeat event", "err", err)
		return nil, err
	}
	return watcher, err
}

func (impl *TelemetryEventClientImplExtended) SummaryEventForTelemetry() {
	err := impl.SendSummaryEvent(string(Summary))
	if err != nil {
		impl.logger.Errorw("error occurred in SummaryEventForTelemetry", "err", err)
	}
}

func (impl *TelemetryEventClientImplExtended) SendSummaryEvent(eventType string) error {
	impl.logger.Infow("sending summary event", "eventType", eventType)
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event while retrieving ucid", "err", err)
		return err
	}

	if posthogClient.IsOptOut {
		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
		return err
	}

	clusters, users, k8sServerVersion, hostURL, ssoSetup, HelmAppAccessCount, ChartStoreVisitCount, SkippedOnboarding, HelmAppUpdateCounter, HelmChartSuccessfulDeploymentCount, ExternalHelmAppClusterCount := impl.SummaryDetailsForTelemetry()
	payload := &TelemetryEventDto{UCID: ucid, Timestamp: time.Now(), EventType: TelemetryEventType(eventType), DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()

	environmentCount, err := impl.environmentService.GetAllActiveEnvironmentCount()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event while retrieving environmentCount, setting its value to -1", "err", err)
		environmentCount = -1
	}

	prodApps, err := impl.appListingRepository.FindAppCount(true)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving prodApps, setting its value to -1", "err", err)
		prodApps = -1
	}

	nonProdApps, err := impl.appListingRepository.FindAppCount(false)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving nonProdApps, setting its value to -1", "err", err)
		nonProdApps = -1
	}

	ciPipelineCount, err := impl.ciPipelineConfigReadService.FindAllPipelineCreatedCountInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving ciPipelineCount, setting its value to -1", "err", err)
		ciPipelineCount = -1
	}
	ciPipelineDeletedCount, err := impl.ciPipelineConfigReadService.FindAllDeletedPipelineCountInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving ciPipelineDeletedCount, setting its value to -1", "err", err)
		ciPipelineDeletedCount = -1
	}
	ciPipelineTriggeredCount, err := impl.ciWorkflowRepository.FindAllTriggeredWorkflowCountInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving ciPipelineTriggeredCount, setting its value to -1", "err", err)
		ciPipelineTriggeredCount = -1
	}

	cdPipelineCount, err := impl.pipelineRepository.FindAllPipelineCreatedCountInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving cdPipelineCount, setting its value to -1", "err", err)
		cdPipelineCount = -1
	}
	cdPipelineDeletedCount, err := impl.pipelineRepository.FindAllDeletedPipelineCountInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving cdPipelineDeletedCount, setting its value to -1", "err", err)
		cdPipelineDeletedCount = -1
	}
	cdPipelineTriggeredCount, err := impl.cdWorkflowRepository.FindAllTriggeredWorkflowCountInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving cdPipelineTriggeredCount, setting its value to -1", "err", err)
		cdPipelineTriggeredCount = -1
	}
	gitAccounts, err := impl.gitProviderRepository.FindAllGitProviderCount()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event, while retrieving gitAccounts, setting its value to -1", "err", err)
		gitAccounts = -1
	}

	gitOpsCount, err := impl.gitOpsConfigReadService.GetConfiguredGitOpsCount()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving gitOpsCount, setting its value to -1", "err", err)
		gitOpsCount = -1
	}

	containerRegistryCount, err := impl.dockerArtifactStoreRepository.FindAllDockerArtifactCount()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving containerRegistryCount, setting its value to -1", "err", err)
		containerRegistryCount = -1
	}

	//appSetup := false
	apps, err := impl.appRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving apps", "err", err)
		return err
	}

	var appIds []int
	for _, appInfo := range apps {
		appIds = append(appIds, appInfo.Id)
	}

	payload.AppCount = len(appIds)
	if len(appIds) < AppsCount {
		payload.AppsWithGitRepoConfigured, err = impl.gitMaterialReadService.FindNumberOfAppsWithGitRepo(appIds)
		if err != nil {
			impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving AppsWithGitRepoConfigured", "err", err)
		}
		payload.AppsWithDockerConfigured, err = impl.ciTemplateRepository.FindNumberOfAppsWithDockerConfigured(appIds)
		if err != nil {
			impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving AppsWithDockerConfigured", "err", err)
		}
		payload.AppsWithDeploymentTemplateConfigured, err = impl.chartRepository.FindNumberOfAppsWithDeploymentTemplate(appIds)
		if err != nil {
			impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving AppsWithDeploymentTemplateConfigured", "err", err)
		}
		payload.AppsWithCiPipelineConfigured, err = impl.ciPipelineConfigReadService.FindNumberOfAppsWithCiPipeline(appIds)
		if err != nil {
			impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving AppsWithCiPipelineConfigured", "err", err)
		}

		payload.AppsWithCdPipelineConfigured, err = impl.pipelineRepository.FindNumberOfAppsWithCdPipeline(appIds)
		if err != nil {
			impl.logger.Errorw("exception caught inside telemetry summary event,while retrieving AppsWithCdPipelineConfigured", "err", err)
		}
	}

	build, err := impl.ciWorkflowRepository.ExistsByStatus("Succeeded")

	deployment, err := impl.cdWorkflowRepository.ExistsByStatus("Healthy")

	// build integrations data
	installedIntegrations, installFailedIntegrations, installTimedOutIntegrations, installingIntegrations, err := impl.buildIntegrationsList()
	if err != nil {
		return err
	}

	selfDockerfileCount, managedDockerfileCount, buildpackCount := impl.getCiBuildTypeData()

	successCount, failureCount := impl.getCiBuildTypeVsStatusVsCount()

	devtronVersion := util.GetDevtronVersion()
	payload.ProdAppCount = prodApps
	payload.NonProdAppCount = nonProdApps
	payload.RegistryCount = containerRegistryCount
	payload.SSOLogin = ssoSetup
	payload.UserCount = len(users)
	payload.EnvironmentCount = environmentCount
	payload.ClusterCount = len(clusters)
	payload.CiCreatedPerDay = ciPipelineCount
	payload.CiDeletedPerDay = ciPipelineDeletedCount
	payload.CiTriggeredPerDay = ciPipelineTriggeredCount
	payload.CdCreatedPerDay = cdPipelineCount
	payload.CdDeletedPerDay = cdPipelineDeletedCount
	payload.CdTriggeredPerDay = cdPipelineTriggeredCount
	payload.GitAccountsCount = gitAccounts
	payload.GitOpsCount = gitOpsCount
	payload.HostURL = hostURL
	payload.DevtronGitVersion = devtronVersion.GitCommit
	payload.Build = build
	payload.Deployment = deployment
	payload.DevtronMode = devtronVersion.ServerMode
	payload.InstalledIntegrations = installedIntegrations
	payload.InstallFailedIntegrations = installFailedIntegrations
	payload.InstallTimedOutIntegrations = installTimedOutIntegrations
	payload.InstallingIntegrations = installingIntegrations
	payload.DevtronReleaseVersion = impl.serverDataStore.CurrentVersion
	payload.HelmAppAccessCounter = HelmAppAccessCount
	payload.ChartStoreVisitCount = ChartStoreVisitCount
	payload.SkippedOnboarding = SkippedOnboarding
	payload.HelmAppUpdateCounter = HelmAppUpdateCounter
	payload.HelmChartSuccessfulDeploymentCount = HelmChartSuccessfulDeploymentCount
	payload.ExternalHelmAppClusterCount = ExternalHelmAppClusterCount

	payload.ClusterProvider, err = impl.GetCloudProvider()
	if err != nil {
		impl.logger.Errorw("error while getting cluster provider", "error", err)
		return err
	}

	latestUser, err := impl.userAuditService.GetLatestUser()
	if err == nil {
		loginTime := latestUser.UpdatedOn
		if loginTime.IsZero() {
			loginTime = latestUser.CreatedOn
		}
		payload.LastLoginTime = loginTime
	}

	payload.SelfDockerfileCount = selfDockerfileCount
	payload.SelfDockerfileSuccessCount = successCount[bean.SELF_DOCKERFILE_BUILD_TYPE]
	payload.SelfDockerfileFailureCount = failureCount[bean.SELF_DOCKERFILE_BUILD_TYPE]

	payload.ManagedDockerfileCount = managedDockerfileCount
	payload.ManagedDockerfileSuccessCount = successCount[bean.MANAGED_DOCKERFILE_BUILD_TYPE]
	payload.ManagedDockerfileFailureCount = failureCount[bean.MANAGED_DOCKERFILE_BUILD_TYPE]

	payload.BuildPackCount = buildpackCount
	payload.BuildPackSuccessCount = successCount[bean.BUILDPACK_BUILD_TYPE]
	payload.BuildPackFailureCount = failureCount[bean.BUILDPACK_BUILD_TYPE]

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, payload marshal error", "error", err)
		return err
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, payload unmarshal error", "error", err)
		return err
	}

	err = impl.EnqueuePostHog(ucid, TelemetryEventType(eventType), prop)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, failed to push event", "ucid", ucid, "error", err)
		return err
	}
	return nil
}

func (impl *TelemetryEventClientImplExtended) getCiBuildTypeData() (int, int, int) {
	countByBuildType := impl.ciBuildConfigService.GetCountByBuildType()
	return countByBuildType[bean.SELF_DOCKERFILE_BUILD_TYPE], countByBuildType[bean.MANAGED_DOCKERFILE_BUILD_TYPE], countByBuildType[bean.BUILDPACK_BUILD_TYPE]
}

func (impl *TelemetryEventClientImplExtended) getCiBuildTypeVsStatusVsCount() (successCount map[bean.CiBuildType]int, failureCount map[bean.CiBuildType]int) {
	successCount = make(map[bean.CiBuildType]int)
	failureCount = make(map[bean.CiBuildType]int)
	buildTypeAndStatusVsCount := impl.ciWorkflowRepository.FindBuildTypeAndStatusDataOfLast1Day()
	for _, buildTypeCount := range buildTypeAndStatusVsCount {
		if buildTypeCount == nil {
			continue
		}
		if buildTypeCount.Type == "" {
			buildTypeCount.Type = string(bean.SELF_DOCKERFILE_BUILD_TYPE)
		}
		if buildTypeCount.Status == "Succeeded" {
			successCount[bean.CiBuildType(buildTypeCount.Type)] = buildTypeCount.Count
		} else {
			failureCount[bean.CiBuildType(buildTypeCount.Type)] = buildTypeCount.Count
		}
	}
	return successCount, failureCount
}
