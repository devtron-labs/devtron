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
	"fmt"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	userrepository "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
	"github.com/devtron-labs/devtron/util/argo"
	errors2 "github.com/juju/errors"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slices"

	"github.com/devtron-labs/devtron/api/bean"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AppListingService interface {
	FetchJobs(fetchJobListingRequest FetchAppListingRequest) ([]*bean.JobContainer, error)
	FetchOverviewCiPipelines(jobId int) ([]*bean.JobListingContainer, error)
	BuildAppListingResponseV2(fetchAppListingRequest FetchAppListingRequest, envContainers []*bean.AppEnvironmentContainer) ([]*bean.AppContainer, error)
	FetchAllDevtronManagedApps() ([]AppNameTypeIdContainer, error)
	FetchAppDetails(ctx context.Context, appId int, envId int) (bean.AppDetailContainer, error)

	//------------------

	FetchAppTriggerView(appId int) ([]bean.TriggerView, error)
	FetchAppStageStatus(appId int, appType int) ([]bean.AppStageStatus, error)

	FetchOtherEnvironment(ctx context.Context, appId int) ([]*bean.Environment, error)
	FetchMinDetailOtherEnvironment(appId int) ([]*bean.Environment, error)
	RedirectToLinkouts(Id int, appId int, envId int, podName string, containerName string) (string, error)
	ISLastReleaseStopType(appId, envId int) (bool, error)
	ISLastReleaseStopTypeV2(pipelineIds []int) (map[int]bool, error)
	GetReleaseCount(appId, envId int) (int, error)

	FetchAppsByEnvironmentV2(fetchAppListingRequest FetchAppListingRequest, w http.ResponseWriter, r *http.Request, token string) ([]*bean.AppEnvironmentContainer, int, error)
	FetchOverviewAppsByEnvironment(envId, limit, offset int) (*OverviewAppsByEnvironmentBean, error)
	FetchCiArtifactToGitTriggersMap(artifacts []*CiArtifactWithParentArtifact) (CiArtifactAndGitCommitsMap map[int][]string, err error)
}

const (
	APIVersionV1 string = "v1"
	APIVersionV2 string = "v2"
)

type FetchAppListingRequest struct {
	Environments      []int            `json:"environments"`
	Statuses          []string         `json:"statuses"`
	Teams             []int            `json:"teams"`
	AppNameSearch     string           `json:"appNameSearch"`
	SortOrder         helper.SortOrder `json:"sortOrder"`
	SortBy            helper.SortBy    `json:"sortBy"`
	Offset            int              `json:"offset"`
	Size              int              `json:"size"`
	DeploymentGroupId int              `json:"deploymentGroupId"`
	Namespaces        []string         `json:"namespaces"` // {clusterId}_{namespace}
	AppStatuses       []string         `json:"appStatuses"`
	AppIds            []int            `json:"-"` // internal use only
	// IsClusterOrNamespaceSelected bool             `json:"isClusterOrNamespaceSelected"`
}
type AppNameTypeIdContainer struct {
	AppName string `json:"appName"`
	Type    string `json:"type"`
	AppId   int    `json:"appId"`
}
type CiArtifactWithParentArtifact struct {
	ParentCiArtifact int `json:"parent_ci_artifact"`
	CiArtifactId     int `json:"ci_artifact_id"`
}

func (req FetchAppListingRequest) GetNamespaceClusterMapping() (namespaceClusterPair []*repository2.ClusterNamespacePair, clusterIds []int, err error) {
	for _, ns := range req.Namespaces {
		items := strings.Split(ns, "_")
		// TODO refactoring: invalid condition; always false
		if len(items) < 1 && len(items) > 2 {
			return nil, nil, fmt.Errorf("invalid namespaceds")
		}
		clusterId, err := strconv.Atoi(items[0])
		if err != nil {
			return nil, nil, fmt.Errorf("invalid clustrer id")
		}
		if len(items) == 2 {
			pair := &repository2.ClusterNamespacePair{
				ClusterId:     clusterId,
				NamespaceName: items[1],
			}
			namespaceClusterPair = append(namespaceClusterPair, pair)

		} else {
			clusterIds = append(clusterIds, clusterId)
		}
	}
	return namespaceClusterPair, clusterIds, nil
}

type AppListingServiceImpl struct {
	Logger                         *zap.SugaredLogger
	application                    application2.ServiceClient
	appRepository                  app.AppRepository
	appListingRepository           repository.AppListingRepository
	appListingViewBuilder          AppListingViewBuilder
	pipelineRepository             pipelineConfig.PipelineRepository
	cdWorkflowRepository           pipelineConfig.CdWorkflowRepository
	linkoutsRepository             repository.LinkoutsRepository
	pipelineOverrideRepository     chartConfig.PipelineOverrideRepository
	environmentRepository          repository2.EnvironmentRepository
	argoUserService                argo.ArgoUserService
	envOverrideRepository          chartConfig.EnvConfigOverrideRepository
	chartRepository                chartRepoRepository.ChartRepository
	ciPipelineRepository           pipelineConfig.CiPipelineRepository
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService
	userRepository                 userrepository.UserRepository
	deployedAppMetricsService      deployedAppMetrics.DeployedAppMetricsService
	ciWorkflowRepository           pipelineConfig.CiWorkflowRepository
}

func NewAppListingServiceImpl(Logger *zap.SugaredLogger, appListingRepository repository.AppListingRepository,
	application application2.ServiceClient, appRepository app.AppRepository,
	appListingViewBuilder AppListingViewBuilder, pipelineRepository pipelineConfig.PipelineRepository,
	linkoutsRepository repository.LinkoutsRepository, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, environmentRepository repository2.EnvironmentRepository,
	argoUserService argo.ArgoUserService, envOverrideRepository chartConfig.EnvConfigOverrideRepository,
	chartRepository chartRepoRepository.ChartRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService, userRepository userrepository.UserRepository,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService, ciWorkflowRepository pipelineConfig.CiWorkflowRepository) *AppListingServiceImpl {
	serviceImpl := &AppListingServiceImpl{
		Logger:                         Logger,
		appListingRepository:           appListingRepository,
		application:                    application,
		appRepository:                  appRepository,
		appListingViewBuilder:          appListingViewBuilder,
		pipelineRepository:             pipelineRepository,
		linkoutsRepository:             linkoutsRepository,
		cdWorkflowRepository:           cdWorkflowRepository,
		pipelineOverrideRepository:     pipelineOverrideRepository,
		environmentRepository:          environmentRepository,
		argoUserService:                argoUserService,
		envOverrideRepository:          envOverrideRepository,
		chartRepository:                chartRepository,
		ciPipelineRepository:           ciPipelineRepository,
		dockerRegistryIpsConfigService: dockerRegistryIpsConfigService,
		userRepository:                 userRepository,
		deployedAppMetricsService:      deployedAppMetricsService,
		ciWorkflowRepository:           ciWorkflowRepository,
	}
	return serviceImpl
}

const AcdInvalidAppErr = "invalid acd app name and env"
const NotDeployed = "Not Deployed"

type OverviewAppsByEnvironmentBean struct {
	EnvironmentId   int                             `json:"environmentId"`
	EnvironmentName string                          `json:"environmentName"`
	Namespace       string                          `json:"namespace"`
	ClusterName     string                          `json:"clusterName"`
	ClusterId       int                             `json:"clusterId"`
	Type            string                          `json:"environmentType"`
	Description     string                          `json:"description"`
	AppCount        int                             `json:"appCount"`
	Apps            []*bean.AppEnvironmentContainer `json:"apps"`
	CreatedOn       string                          `json:"createdOn"`
	CreatedBy       string                          `json:"createdBy"`
}

const (
	Production    = "Production"
	NonProduction = "Non-Production"
)

func (impl AppListingServiceImpl) FetchOverviewAppsByEnvironment(envId, limit, offset int) (*OverviewAppsByEnvironmentBean, error) {
	resp := &OverviewAppsByEnvironmentBean{}
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.Logger.Errorw("failed to fetch env", "err", err, "envId", envId)
		return resp, err
	}
	resp.EnvironmentId = envId
	resp.EnvironmentName = env.Name
	resp.ClusterName = env.Cluster.ClusterName
	resp.ClusterId = env.ClusterId
	resp.Namespace = env.Namespace
	resp.CreatedOn = env.CreatedOn.String()
	if env.Default {
		resp.Type = Production
	} else {
		resp.Type = NonProduction
	}
	resp.Description = env.Description
	createdBy, err := impl.userRepository.GetByIdIncludeDeleted(env.CreatedBy)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching user for app meta info", "error", err, "env.CreatedBy", env.CreatedBy)
		return nil, err
	}
	if createdBy != nil && createdBy.Id > 0 {
		if createdBy.Active {
			resp.CreatedBy = fmt.Sprintf(createdBy.EmailId)
		} else {
			resp.CreatedBy = fmt.Sprintf("%s (inactive)", createdBy.EmailId)
		}
	}
	envContainers, err := impl.appListingRepository.FetchOverviewAppsByEnvironment(envId, limit, offset)
	if err != nil {
		impl.Logger.Errorw("failed to fetch environment containers", "err", err, "envId", envId)
		return resp, err
	}

	//artifactStore :=make()

	artifactDetails := make([]*CiArtifactWithParentArtifact, 0)
	for _, envContainer := range envContainers {
		lastDeployed, err := impl.appListingRepository.FetchLastDeployedImage(envContainer.AppId, envId)
		if err != nil {
			impl.Logger.Errorw("failed to fetch last deployed image", "err", err, "appId", envContainer.AppId, "envId", envId)
			return resp, err
		}

		if lastDeployed != nil {
			envContainer.LastDeployedImage = lastDeployed.LastDeployedImage
			envContainer.LastDeployedBy = lastDeployed.LastDeployedBy
			artifactDetails = append(artifactDetails, &CiArtifactWithParentArtifact{
				lastDeployed.ParentCiArtifact, lastDeployed.CiArtifactId,
			})
			//envContainer.Commits = gitCommits
			envContainer.CiArtifactId = lastDeployed.CiArtifactId
		}
	}

	artifactWithGitCommit, err := impl.FetchCiArtifactToGitTriggersMap(artifactDetails)
	if err != nil {
		impl.Logger.Errorw("failed to fetch Artifacts to git Triggers ", "err", err, "envId", envId)
		return resp, err
	}
	for _, envContainer := range envContainers {
		if envContainer.CiArtifactId > 0 {
			envContainer.Commits = artifactWithGitCommit[envContainer.CiArtifactId]
		}
	}

	resp.Apps = envContainers
	return resp, err
}

func (impl AppListingServiceImpl) FetchAllDevtronManagedApps() ([]AppNameTypeIdContainer, error) {
	impl.Logger.Debug("reached at FetchAllDevtronManagedApps:")
	apps := make([]AppNameTypeIdContainer, 0)
	res, err := impl.appRepository.FetchAllActiveDevtronAppsWithAppIdAndName()
	if err != nil {
		impl.Logger.Errorw("failed to fetch devtron apps", "err", err)
		return nil, err
	}
	for _, r := range res {
		appContainer := AppNameTypeIdContainer{
			AppId:   r.Id,
			AppName: r.AppName,
			Type:    "devtron-app",
		}
		apps = append(apps, appContainer)
	}
	res, err = impl.appRepository.FetchAllActiveInstalledAppsWithAppIdAndName()
	if err != nil {
		impl.Logger.Errorw("failed to fetch devtron installed apps", "err", err)
		return nil, err
	}
	for _, r := range res {
		appContainer := AppNameTypeIdContainer{
			AppId:   r.Id,
			AppName: r.AppName,
			Type:    "devtron-installed-app",
		}
		apps = append(apps, appContainer)
	}
	return apps, nil
}
func (impl AppListingServiceImpl) FetchJobs(fetchJobListingRequest FetchAppListingRequest) ([]*bean.JobContainer, error) {

	jobListingFilter := helper.AppListingFilter{
		Teams:         fetchJobListingRequest.Teams,
		AppNameSearch: fetchJobListingRequest.AppNameSearch,
		SortOrder:     fetchJobListingRequest.SortOrder,
		SortBy:        fetchJobListingRequest.SortBy,
		Offset:        fetchJobListingRequest.Offset,
		Size:          fetchJobListingRequest.Size,
		AppStatuses:   fetchJobListingRequest.AppStatuses,
		Environments:  fetchJobListingRequest.Environments,
		AppIds:        fetchJobListingRequest.AppIds,
	}
	appIds, err := impl.appRepository.FetchAppIdsWithFilter(jobListingFilter)
	if err != nil {
		impl.Logger.Errorw("error in fetching app ids list", "error", err, jobListingFilter)
		return []*bean.JobContainer{}, err
	}
	jobListingContainers, err := impl.appListingRepository.FetchJobs(appIds, jobListingFilter.AppStatuses, jobListingFilter.Environments, string(jobListingFilter.SortOrder))
	if err != nil {
		impl.Logger.Errorw("error in fetching job list", "error", err, jobListingFilter)
		return []*bean.JobContainer{}, err
	}
	CiPipelineIDs := GetCIPipelineIDs(jobListingContainers)
	JobsLastSucceededOnTime, err := impl.appListingRepository.FetchJobsLastSucceededOn(CiPipelineIDs)
	jobContainers := BuildJobListingResponse(jobListingContainers, JobsLastSucceededOnTime)
	return jobContainers, nil
}

func (impl AppListingServiceImpl) FetchOverviewCiPipelines(jobId int) ([]*bean.JobListingContainer, error) {
	jobCiContainers, err := impl.appListingRepository.FetchOverviewCiPipelines(jobId)
	if err != nil {
		impl.Logger.Errorw("error in fetching job container", "error", err, jobId)
		return []*bean.JobListingContainer{}, err
	}
	return jobCiContainers, nil
}

func (impl AppListingServiceImpl) FetchAppsByEnvironmentV2(fetchAppListingRequest FetchAppListingRequest, w http.ResponseWriter, r *http.Request, token string) ([]*bean.AppEnvironmentContainer, int, error) {
	impl.Logger.Debug("reached at FetchAppsByEnvironment:")
	if len(fetchAppListingRequest.Namespaces) != 0 && len(fetchAppListingRequest.Environments) == 0 {
		return []*bean.AppEnvironmentContainer{}, 0, nil
	}
	appListingFilter := helper.AppListingFilter{
		Environments:      fetchAppListingRequest.Environments,
		Statuses:          fetchAppListingRequest.Statuses,
		Teams:             fetchAppListingRequest.Teams,
		AppNameSearch:     fetchAppListingRequest.AppNameSearch,
		SortOrder:         fetchAppListingRequest.SortOrder,
		SortBy:            fetchAppListingRequest.SortBy,
		Offset:            fetchAppListingRequest.Offset,
		Size:              fetchAppListingRequest.Size,
		DeploymentGroupId: fetchAppListingRequest.DeploymentGroupId,
		AppStatuses:       fetchAppListingRequest.AppStatuses,
		AppIds:            fetchAppListingRequest.AppIds,
	}
	_, span := otel.Tracer("appListingRepository").Start(r.Context(), "FetchAppsByEnvironment")
	envContainers, appSize, err := impl.appListingRepository.FetchAppsByEnvironmentV2(appListingFilter)
	span.End()
	if err != nil {
		impl.Logger.Errorw("error in fetching app list", "error", err, "filter", appListingFilter)
		return []*bean.AppEnvironmentContainer{}, appSize, err
	}

	envContainersMap := make(map[int][]*bean.AppEnvironmentContainer)
	envIds := make([]int, 0)
	envsSet := make(map[int]bool)

	for _, container := range envContainers {
		if container.EnvironmentId != 0 {
			if _, ok := envContainersMap[container.EnvironmentId]; !ok {
				envContainersMap[container.EnvironmentId] = make([]*bean.AppEnvironmentContainer, 0)
			}
			envContainersMap[container.EnvironmentId] = append(envContainersMap[container.EnvironmentId], container)
			if _, ok := envsSet[container.EnvironmentId]; !ok {
				envIds = append(envIds, container.EnvironmentId)
				envsSet[container.EnvironmentId] = true
			}
		}
	}
	envClusterInfos, err := impl.environmentRepository.FindEnvClusterInfosByIds(envIds)
	if err != nil {
		impl.Logger.Errorw("error in envClusterInfos list", "error", err, "envIds", envIds)
		return []*bean.AppEnvironmentContainer{}, appSize, err
	}
	for _, info := range envClusterInfos {
		for _, container := range envContainersMap[info.Id] {
			container.Namespace = info.Namespace
			container.ClusterName = info.ClusterName
			container.EnvironmentName = info.Name
		}
	}
	return envContainers, appSize, nil
}

func (impl AppListingServiceImpl) ISLastReleaseStopType(appId, envId int) (bool, error) {
	override, err := impl.pipelineOverrideRepository.GetLatestRelease(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting last release")
		return false, err
	} else if util.IsErrNoRows(err) {
		return false, nil
	} else {
		cdWfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), override.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil {
			impl.Logger.Errorw("error in getting latest wfr by pipelineId", "err", err, "cdWorkflowId", override.CdWorkflowId)
			return false, err
		}
		if slices.Contains([]string{pipelineConfig.WorkflowInitiated, pipelineConfig.WorkflowInQueue}, cdWfr.Status) {
			return false, nil
		}
		return models.DEPLOYMENTTYPE_STOP == override.DeploymentType, nil
	}
}

func (impl AppListingServiceImpl) ISLastReleaseStopTypeV2(pipelineIds []int) (map[int]bool, error) {
	releaseMap := make(map[int]bool)
	if len(pipelineIds) == 0 {
		return releaseMap, nil
	}
	overrides, err := impl.pipelineOverrideRepository.GetLatestReleaseDeploymentType(pipelineIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting last release")
		return releaseMap, err
	} else if util.IsErrNoRows(err) {
		return releaseMap, nil
	}
	for _, override := range overrides {
		if _, ok := releaseMap[override.PipelineId]; !ok {
			cdWfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), override.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
			if err != nil {
				impl.Logger.Errorw("error in getting latest wfr by pipelineId", "err", err, "cdWorkflowId", override.CdWorkflowId)
				releaseMap[override.PipelineId] = false
				continue
			}
			if slices.Contains([]string{pipelineConfig.WorkflowInitiated, pipelineConfig.WorkflowInQueue}, cdWfr.Status) {
				releaseMap[override.PipelineId] = false
				continue
			}
			isStopType := models.DEPLOYMENTTYPE_STOP == override.DeploymentType
			releaseMap[override.PipelineId] = isStopType
		}
	}
	return releaseMap, nil
}

func (impl AppListingServiceImpl) GetReleaseCount(appId, envId int) (int, error) {
	override, err := impl.pipelineOverrideRepository.GetAllRelease(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting releases")
		return 0, err
	} else if util.IsErrNoRows(err) {
		return 0, nil
	} else {
		return len(override), nil
	}
}

func (impl AppListingServiceImpl) BuildAppListingResponseV2(fetchAppListingRequest FetchAppListingRequest, envContainers []*bean.AppEnvironmentContainer) ([]*bean.AppContainer, error) {
	start := time.Now()
	appEnvMapping, err := impl.fetchACDAppStatusV2(fetchAppListingRequest, envContainers)
	middleware.AppListingDuration.WithLabelValues("fetchACDAppStatus", "devtron").Observe(time.Since(start).Seconds())
	if err != nil {
		impl.Logger.Errorw("error in fetching app statuses", "error", err)
		return []*bean.AppContainer{}, err
	}
	start = time.Now()
	appContainerResponses, err := impl.appListingViewBuilder.BuildView(fetchAppListingRequest, appEnvMapping)
	middleware.AppListingDuration.WithLabelValues("buildView", "devtron").Observe(time.Since(start).Seconds())
	return appContainerResponses, err
}
func GetCIPipelineIDs(jobContainers []*bean.JobListingContainer) []int {

	var ciPipelineIDs []int
	for _, jobContainer := range jobContainers {
		ciPipelineIDs = append(ciPipelineIDs, jobContainer.CiPipelineID)
	}
	return ciPipelineIDs
}
func BuildJobListingResponse(jobContainers []*bean.JobListingContainer, JobsLastSucceededOnTime []*bean.CiPipelineLastSucceededTime) []*bean.JobContainer {
	jobContainersMapping := make(map[int]bean.JobContainer)
	var appIds []int

	lastSucceededTimeMapping := make(map[int]time.Time)
	for _, lastSuccessTime := range JobsLastSucceededOnTime {
		lastSucceededTimeMapping[lastSuccessTime.CiPipelineID] = lastSuccessTime.LastSucceededOn
	}

	// Storing the sequence in appIds array
	for _, jobContainer := range jobContainers {
		val, ok := jobContainersMapping[jobContainer.JobId]
		if !ok {
			appIds = append(appIds, jobContainer.JobId)
			val = bean.JobContainer{}
			val.JobId = jobContainer.JobId
			val.JobName = jobContainer.JobName
			val.JobActualName = jobContainer.JobActualName
			val.ProjectId = jobContainer.ProjectId
		}

		if len(val.JobCiPipelines) == 0 {
			val.JobCiPipelines = make([]bean.JobCIPipeline, 0)
		}

		if jobContainer.CiPipelineID != 0 {
			ciPipelineObj := bean.JobCIPipeline{
				CiPipelineId:                 jobContainer.CiPipelineID,
				CiPipelineName:               jobContainer.CiPipelineName,
				Status:                       jobContainer.Status,
				LastRunAt:                    jobContainer.StartedOn,
				EnvironmentName:              jobContainer.EnvironmentName,
				EnvironmentId:                jobContainer.EnvironmentId,
				LastTriggeredEnvironmentName: jobContainer.LastTriggeredEnvironmentName,
				// LastSuccessAt: jobContainer.LastSuccessAt,
			}
			if lastSuccessAt, ok := lastSucceededTimeMapping[jobContainer.CiPipelineID]; ok {
				ciPipelineObj.LastSuccessAt = lastSuccessAt
			}

			val.JobCiPipelines = append(val.JobCiPipelines, ciPipelineObj)
		}
		jobContainersMapping[jobContainer.JobId] = val

	}

	result := make([]*bean.JobContainer, 0)
	for _, appId := range appIds {
		val := jobContainersMapping[appId]
		result = append(result, &val)
	}

	return result
}

func (impl AppListingServiceImpl) fetchACDAppStatus(fetchAppListingRequest FetchAppListingRequest, existingAppEnvContainers []*bean.AppEnvironmentContainer) (map[string][]*bean.AppEnvironmentContainer, error) {
	appEnvMapping := make(map[string][]*bean.AppEnvironmentContainer)
	var appNames []string
	var appIds []int
	var pipelineIds []int
	for _, env := range existingAppEnvContainers {
		appIds = append(appIds, env.AppId)
		if env.EnvironmentName == "" {
			continue
		}
		appName := fmt.Sprintf("%s-%s", env.AppName, env.EnvironmentName)
		appNames = append(appNames, appName)
		pipelineIds = append(pipelineIds, env.PipelineId)
	}

	appEnvPipelinesMap := make(map[string][]*pipelineConfig.Pipeline)
	appEnvCdWorkflowMap := make(map[string]*pipelineConfig.CdWorkflow)
	appEnvCdWorkflowRunnerMap := make(map[int][]*pipelineConfig.CdWorkflowRunner)

	// get all the active cd pipelines
	if len(pipelineIds) > 0 {
		pipelinesAll, err := impl.pipelineRepository.FindByIdsIn(pipelineIds) // TODO - OPTIMIZE 1
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("err", err)
			return nil, err
		}
		// here to build a map of pipelines list for each (appId and envId)
		for _, p := range pipelinesAll {
			key := fmt.Sprintf("%d-%d", p.AppId, p.EnvironmentId)
			if _, ok := appEnvPipelinesMap[key]; !ok {
				var appEnvPipelines []*pipelineConfig.Pipeline
				appEnvPipelines = append(appEnvPipelines, p)
				appEnvPipelinesMap[key] = appEnvPipelines
			} else {
				appEnvPipelines := appEnvPipelinesMap[key]
				appEnvPipelines = append(appEnvPipelines, p)
				appEnvPipelinesMap[key] = appEnvPipelines
			}
		}

		// from all the active pipeline, get all the cd workflow
		cdWorkflowAll, err := impl.cdWorkflowRepository.FindLatestCdWorkflowByPipelineIdV2(pipelineIds) // TODO - OPTIMIZE 2
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Error(err)
			return nil, err
		}
		// find and build a map of latest cd workflow for each (appId and envId), single latest CDWF for any of the cd pipelines.
		var wfIds []int
		for key, v := range appEnvPipelinesMap {
			if _, ok := appEnvCdWorkflowMap[key]; !ok {
				for _, itemW := range cdWorkflowAll {
					for _, itemP := range v {
						if itemW.PipelineId == itemP.Id {
							// GOT LATEST CD WF, AND PUT INTO MAP
							appEnvCdWorkflowMap[key] = itemW
							wfIds = append(wfIds, itemW.Id)
						}
					}
				}
				// if no cd wf found for appid-envid, add it into map with nil
				if _, ok := appEnvCdWorkflowMap[key]; !ok {
					appEnvCdWorkflowMap[key] = nil
				}
			}
		}

		// fetch all the cd workflow runner from cdWF ids,
		cdWorkflowRunnersAll, err := impl.cdWorkflowRepository.FindWorkflowRunnerByCdWorkflowId(wfIds) // TODO - OPTIMIZE 3
		if err != nil {
			impl.Logger.Errorw("error in getting wf", "err", err)
		}
		// build a map with key cdWF containing cdWFRunner List, which are later put in map for further requirement
		for _, item := range cdWorkflowRunnersAll {
			if _, ok := appEnvCdWorkflowRunnerMap[item.CdWorkflowId]; !ok {
				var cdWorkflowRunners []*pipelineConfig.CdWorkflowRunner
				cdWorkflowRunners = append(cdWorkflowRunners, item)
				appEnvCdWorkflowRunnerMap[item.CdWorkflowId] = cdWorkflowRunners
			} else {
				appEnvCdWorkflowRunnerMap[item.CdWorkflowId] = append(appEnvCdWorkflowRunnerMap[item.CdWorkflowId], item)
			}
		}
	}
	releaseMap, _ := impl.ISLastReleaseStopTypeV2(pipelineIds)

	for _, env := range existingAppEnvContainers {
		appKey := strconv.Itoa(env.AppId) + "_" + env.AppName
		if _, ok := appEnvMapping[appKey]; !ok {
			var appEnvContainers []*bean.AppEnvironmentContainer
			appEnvMapping[appKey] = appEnvContainers
		}

		key := fmt.Sprintf("%d-%d", env.AppId, env.EnvironmentId)
		pipelines := appEnvPipelinesMap[key]
		if len(pipelines) == 0 {
			impl.Logger.Debugw("no pipeline found")
			appEnvMapping[appKey] = append(appEnvMapping[appKey], env)
			continue
		}

		latestTriggeredWf := appEnvCdWorkflowMap[key]
		if latestTriggeredWf == nil || latestTriggeredWf.Id == 0 {
			appEnvMapping[appKey] = append(appEnvMapping[appKey], env)
			continue
		}
		var pipeline *pipelineConfig.Pipeline
		for _, p := range pipelines {
			if p.Id == latestTriggeredWf.PipelineId {
				pipeline = p
				break
			}
		}
		var preCdStageRunner, postCdStageRunner, cdStageRunner *pipelineConfig.CdWorkflowRunner
		cdStageRunners := appEnvCdWorkflowRunnerMap[latestTriggeredWf.Id]
		for _, runner := range cdStageRunners {
			if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
				preCdStageRunner = runner
			} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_DEPLOY {
				cdStageRunner = runner
			} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
				postCdStageRunner = runner
			}
		}

		if latestTriggeredWf.WorkflowStatus == pipelineConfig.WF_STARTED || latestTriggeredWf.WorkflowStatus == pipelineConfig.WF_UNKNOWN {
			if pipeline.PreStageConfig != "" {
				if preCdStageRunner != nil && preCdStageRunner.Id != 0 {
					env.PreStageStatus = &preCdStageRunner.Status
				} else {
					status := ""
					env.PreStageStatus = &status
				}
			}
			if pipeline.PostStageConfig != "" {
				if postCdStageRunner != nil && postCdStageRunner.Id != 0 {
					env.PostStageStatus = &postCdStageRunner.Status
				} else {
					status := ""
					env.PostStageStatus = &status
				}
			}
			if cdStageRunner != nil {
				status := cdStageRunner.Status
				if status == string(health.HealthStatusHealthy) {
					stopType := releaseMap[pipeline.Id]
					if stopType {
						status = argoApplication.HIBERNATING
						env.Status = status
					}
				}
				env.CdStageStatus = &status

			} else {
				status := ""
				env.CdStageStatus = &status
			}
		} else {
			if pipeline.PreStageConfig != "" {
				if preCdStageRunner != nil && preCdStageRunner.Id != 0 {
					var status string = latestTriggeredWf.WorkflowStatus.String()
					env.PreStageStatus = &status
				} else {
					status := ""
					env.PreStageStatus = &status
				}
			}
			if pipeline.PostStageConfig != "" {
				if postCdStageRunner != nil && postCdStageRunner.Id != 0 {
					var status string = latestTriggeredWf.WorkflowStatus.String()
					env.PostStageStatus = &status
				} else {
					status := ""
					env.PostStageStatus = &status
				}
			}
			var status string = latestTriggeredWf.WorkflowStatus.String()

			env.CdStageStatus = &status
		}

		appEnvMapping[appKey] = append(appEnvMapping[appKey], env)
	}
	return appEnvMapping, nil
}

func (impl AppListingServiceImpl) fetchACDAppStatusV2(fetchAppListingRequest FetchAppListingRequest, existingAppEnvContainers []*bean.AppEnvironmentContainer) (map[string][]*bean.AppEnvironmentContainer, error) {
	appEnvMapping := make(map[string][]*bean.AppEnvironmentContainer)
	for _, env := range existingAppEnvContainers {
		appKey := strconv.Itoa(env.AppId) + "_" + env.AppName
		appEnvMapping[appKey] = append(appEnvMapping[appKey], env)
	}
	return appEnvMapping, nil
}

func (impl AppListingServiceImpl) FetchAppDetails(ctx context.Context, appId int, envId int) (bean.AppDetailContainer, error) {
	appDetailContainer, err := impl.appListingRepository.FetchAppDetail(ctx, appId, envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching app detail", "error", err)
		return bean.AppDetailContainer{}, err
	}
	appDetailContainer.AppId = appId

	// set ifIpsAccess provided and relevant data
	appDetailContainer.IsExternalCi = true
	environment, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching env details, FetchAppDetails service", "error", err)
		return bean.AppDetailContainer{}, err
	}
	appDetailContainer, err = impl.setIpAccessProvidedData(ctx, appDetailContainer, appDetailContainer.ClusterId, environment.IsVirtualEnvironment)
	if err != nil {
		return appDetailContainer, err
	}

	return appDetailContainer, nil
}

func (impl AppListingServiceImpl) setIpAccessProvidedData(ctx context.Context, appDetailContainer bean.AppDetailContainer, clusterId int, isVirtualEnv bool) (bean.AppDetailContainer, error) {
	ciPipelineId := appDetailContainer.CiPipelineId
	if ciPipelineId > 0 {
		_, span := otel.Tracer("orchestrator").Start(ctx, "ciPipelineRepository.FindWithMinDataByCiPipelineId")
		ciPipeline, err := impl.ciPipelineRepository.FindWithMinDataByCiPipelineId(ciPipelineId)
		span.End()
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in fetching ciPipeline", "ciPipelineId", ciPipelineId, "error", err)
			return bean.AppDetailContainer{}, err
		}

		if ciPipeline != nil && ciPipeline.CiTemplate != nil && len(*ciPipeline.CiTemplate.DockerRegistryId) > 0 {
			dockerRegistryId := ciPipeline.CiTemplate.DockerRegistryId
			appDetailContainer.DockerRegistryId = *dockerRegistryId
			if !ciPipeline.IsExternal || ciPipeline.ParentCiPipeline != 0 {
				appDetailContainer.IsExternalCi = false
			}
			_, span = otel.Tracer("orchestrator").Start(ctx, "dockerRegistryIpsConfigService.IsImagePullSecretAccessProvided")
			// check ips access provided to this docker registry for that cluster
			ipsAccessProvided, err := impl.dockerRegistryIpsConfigService.IsImagePullSecretAccessProvided(*dockerRegistryId, clusterId, isVirtualEnv)
			span.End()
			if err != nil {
				impl.Logger.Errorw("error in checking if docker registry ips access provided", "dockerRegistryId", dockerRegistryId, "clusterId", clusterId, "error", err)
				return bean.AppDetailContainer{}, err
			}
			appDetailContainer.IpsAccessProvided = ipsAccessProvided
		}
	}
	return appDetailContainer, nil
}

func (impl AppListingServiceImpl) FetchAppTriggerView(appId int) ([]bean.TriggerView, error) {
	return impl.appListingRepository.FetchAppTriggerView(appId)
}

func (impl AppListingServiceImpl) FetchAppStageStatus(appId int, appType int) ([]bean.AppStageStatus, error) {
	appStageStatuses, err := impl.appListingRepository.FetchAppStageStatus(appId, appType)
	return appStageStatuses, err
}
func (impl AppListingServiceImpl) FetchCiArtifactToGitTriggersMap(artifacts []*CiArtifactWithParentArtifact) (CiArtifactAndGitCommitsMap map[int][]string, err error) {

	// Declare variables to store CI artifact IDs and child-parent relationships.
	var ciArtifactIds []int
	ciChildParentIdMap := make(map[int]int)

	// Iterate through the artifacts to build the child-parent relationship map and gather unique artifact IDs.
	for _, artifact := range artifacts {
		// Mapping the current artifact to its parent, or to itself if it has no parent.
		if artifact.ParentCiArtifact > 0 {
			ciChildParentIdMap[artifact.CiArtifactId] = artifact.ParentCiArtifact
		} else {
			ciChildParentIdMap[artifact.CiArtifactId] = artifact.CiArtifactId
		}

		// Ensure uniqueness of artifact IDs in the slice.
		if ciArtifactIds == nil {
			ciArtifactIds = make([]int, 0)
		}
		if !slices.Contains(ciArtifactIds, ciChildParentIdMap[artifact.CiArtifactId]) {
			ciArtifactIds = append(ciArtifactIds, ciChildParentIdMap[artifact.CiArtifactId])
		}
	}

	// Retrieve workflows associated with the artifact IDs, handling any potential error.
	artifactsWithGitTriggers, err := impl.ciWorkflowRepository.FindAllLastGitTriggeredWorkflowByArtifactIds(ciArtifactIds)
	if err != nil {
		// Log the error along with the artifact IDs that caused it.
		impl.Logger.Errorw("error retrieving GitTriggers of the  CiWorkflows", "ciArtifactIds", ciArtifactIds, "err", err)
		return nil, err // Return nil map and the encountered error.
	}

	// Map to hold the last Git-triggered workflow for each artifact.
	gitTriggersToParentArtifactMap := make(map[int]map[int]pipelineConfig.GitCommit)

	// Populate the gitTriggersToParentArtifactMap with Ci Workflow git triggers, ensuring no duplicates based on ArtifactId.
	for _, artifactWithGitTriggers := range artifactsWithGitTriggers {
		if artifactWithGitTriggers == nil {
			// If artifactWithGitTriggers is nil, skip processing it.
			continue
		}

		if _, ok := gitTriggersToParentArtifactMap[artifactWithGitTriggers.ArtifactId]; !ok {
			gitTriggersToParentArtifactMap[artifactWithGitTriggers.ArtifactId] = artifactWithGitTriggers.GitTriggers
		}
	}

	// Map to hold Git commits associated with each CI artifact.
	gitCommitsWithArtifactMap := make(map[int][]string)

	// Iterate through the child-parent relationship map to populate gitCommitsWithArtifactMap.
	for child, parent := range ciChildParentIdMap {
		// Retrieve the Git triggers associated with the parent artifact.
		ciWorkflowTemp, exists := gitTriggersToParentArtifactMap[parent]
		if !exists {
			// If there's no workflow for this parent, continue to the next.
			continue
		}

		// Declare a slice to hold git commits. If there are no GitTriggers, this remains empty.
		var gitCommits []string

		// Iterate through the Git triggers and extract the commit information.
		for _, git := range ciWorkflowTemp {
			gitCommits = append(gitCommits, git.Commit)
		}

		// Append the extracted commits to the corresponding child artifact in the map.
		gitCommitsWithArtifactMap[child] = append(gitCommitsWithArtifactMap[child], gitCommits...)

		// Ensure that gitCommitsWithArtifactMap[child] is initialized if it doesn't exist.
		if gitCommitsWithArtifactMap[child] == nil {
			gitCommitsWithArtifactMap[child] = make([]string, 0)
		}
	}

	return gitCommitsWithArtifactMap, nil
}

func (impl AppListingServiceImpl) FetchOtherEnvironment(ctx context.Context, appId int) ([]*bean.Environment, error) {
	newCtx, span := otel.Tracer("appListingRepository").Start(ctx, "FetchOtherEnvironment")
	envs, err := impl.appListingRepository.FetchOtherEnvironment(appId)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return envs, err
	}
	appLevelInfraMetrics := true //default val, not being derived from DB. TODO: remove this from FE since this is derived from prometheus config at cluster level and this logic is already present at FE
	newCtx, span = otel.Tracer("deployedAppMetricsService").Start(newCtx, "GetMetricsFlagByAppId")
	appLevelAppMetrics, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	span.End()
	if err != nil {
		impl.Logger.Errorw("error, GetMetricsFlagByAppId", "err", err, "appId", appId)
		return envs, err
	}
	newCtx, span = otel.Tracer("chartRepository").Start(newCtx, "FindLatestChartForAppByAppId")
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	span.End()
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching latest chart", "err", err)
		return envs, err
	}
	ciArtifactsWithParent := make([]*CiArtifactWithParentArtifact, 0)
	for _, env := range envs {

		if ok := env.CiArtifactId; ok > 0 {
			ciArtifactsWithParent = append(ciArtifactsWithParent, &CiArtifactWithParentArtifact{
				env.ParentCiArtifactId, env.CiArtifactId,
			})
		}

	}

	gitCommitsWithArtifacts, err := impl.FetchCiArtifactToGitTriggersMap(ciArtifactsWithParent)
	if err != nil {
		impl.Logger.Errorw("Error in fetching the git commits of the ciArtifacts")
		return envs, err
	}

	for _, env := range envs {
		newCtx, span = otel.Tracer("envOverrideRepository").Start(newCtx, "FindLatestChartForAppByAppIdAndEnvId")
		envOverride, err := impl.envOverrideRepository.FindLatestChartForAppByAppIdAndEnvId(appId, env.EnvironmentId)
		span.End()
		if err != nil && !errors2.IsNotFound(err) {
			impl.Logger.Errorw("error in fetching latest chart by appId and envId", "err", err, "appId", appId, "envId", env.EnvironmentId)
			return envs, err
		}
		if envOverride != nil && envOverride.Chart != nil {
			env.ChartRefId = envOverride.Chart.ChartRefId
		} else {
			env.ChartRefId = chart.ChartRefId
		}

		gitCommits, exists := gitCommitsWithArtifacts[env.CiArtifactId]
		if exists {
			env.Commits = gitCommits
		} else {
			gitCommits = make([]string, 0)
			env.Commits = gitCommits
		}

		if env.AppMetrics == nil {
			env.AppMetrics = &appLevelAppMetrics
		}
		env.InfraMetrics = &appLevelInfraMetrics //using default value, discarding value got from query
	}
	return envs, nil
}

func (impl AppListingServiceImpl) FetchMinDetailOtherEnvironment(appId int) ([]*bean.Environment, error) {
	envs, err := impl.appListingRepository.FetchMinDetailOtherEnvironment(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return envs, err
	}
	appLevelInfraMetrics := true //default val, not being derived from DB. TODO: remove this from FE since this is derived from prometheus config at cluster level and this logic is already present at FE
	appLevelAppMetrics, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error, GetMetricsFlagByAppId", "err", err, "appId", appId)
		return nil, err
	}

	chartRefId, err := impl.chartRepository.FindChartRefIdForLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching latest chartRefId", "err", err)
		return envs, err
	}
	var envIds []int
	for _, env := range envs {
		envIds = append(envIds, env.EnvironmentId)
	}
	if len(envIds) == 0 {
		impl.Logger.Infow("No environments found for appId", "appId", appId)
		return envs, nil
	}
	overrideChartRefIds, err := impl.envOverrideRepository.FindChartRefIdsForLatestChartForAppByAppIdAndEnvIds(appId, envIds)
	if err != nil && !errors2.IsNotFound(err) {
		impl.Logger.Errorw("error in fetching latest chartRefIds id by appId and envIds", "err", err, "appId", appId, "envId", envIds)
		return envs, err
	}
	for _, env := range envs {
		if len(overrideChartRefIds) != 0 && overrideChartRefIds[env.EnvironmentId] != 0 {
			env.ChartRefId = overrideChartRefIds[env.EnvironmentId]
		} else {
			env.ChartRefId = chartRefId
		}
		if env.AppMetrics == nil {
			env.AppMetrics = &appLevelAppMetrics
		}
		env.InfraMetrics = &appLevelInfraMetrics //using default value, discarding value got from query
	}
	return envs, nil
}

func arrContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (impl AppListingServiceImpl) RedirectToLinkouts(Id int, appId int, envId int, podName string, containerName string) (string, error) {
	linkout, err := impl.linkoutsRepository.FetchLinkoutById(Id)
	if err != nil {
		impl.Logger.Errorw("Exception", err)
		return "", err
	}
	link := linkout.Link
	if len(podName) > 0 && len(containerName) > 0 {
		link = strings.ReplaceAll(link, "{appName}", linkout.AppName)
		link = strings.ReplaceAll(link, "{envName}", linkout.EnvName)
		link = strings.ReplaceAll(link, "{podName}", podName)
		link = strings.ReplaceAll(link, "{containerName}", containerName)
	} else if len(podName) > 0 {
		link = strings.ReplaceAll(link, "{appName}", linkout.AppName)
		link = strings.ReplaceAll(link, "{envName}", linkout.EnvName)
		link = strings.ReplaceAll(link, "{podName}", podName)
	} else if len(containerName) > 0 {
		link = strings.ReplaceAll(link, "{appName}", linkout.AppName)
		link = strings.ReplaceAll(link, "{envName}", linkout.EnvName)
		link = strings.ReplaceAll(link, "{containerName}", containerName)
	} else {
		link = strings.ReplaceAll(link, "{appName}", linkout.AppName)
		link = strings.ReplaceAll(link, "{envName}", linkout.EnvName)
	}

	return link, nil
}
