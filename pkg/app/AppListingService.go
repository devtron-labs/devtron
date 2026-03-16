/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package app

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean/AppView"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	read2 "github.com/devtron-labs/devtron/api/helm-app/service/read"
	bean7 "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/globalTag"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/util/helm"
	read4 "github.com/devtron-labs/devtron/pkg/app/appDetails/read"
	bean2 "github.com/devtron-labs/devtron/pkg/app/bean"
	userrepository "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	buildCommonBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	ciConfig "github.com/devtron-labs/devtron/pkg/build/pipeline/read"
	read3 "github.com/devtron-labs/devtron/pkg/chart/read"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	cdConfigRead "github.com/devtron-labs/devtron/pkg/deployment/common/read"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	constants2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/constants"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	bean3 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	util4 "github.com/devtron-labs/devtron/pkg/devtronResource/util"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	util3 "github.com/devtron-labs/devtron/util"

	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
	errors2 "github.com/juju/errors"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
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
	FetchJobs(fetchJobListingRequest FetchAppListingRequest) ([]*AppView.JobContainer, error)
	FetchOverviewCiPipelines(jobId int) ([]*AppView.JobListingContainer, error)
	FetchJobCiPipelines() ([]*AppView.JobListingContainer, error)
	BuildAppListingResponseV2(fetchAppListingRequest FetchAppListingRequest, envContainers []*AppView.AppEnvironmentContainer) ([]*AppView.AppContainer, error)
	FetchAllDevtronManagedApps() ([]AppNameTypeIdContainer, error)
	FetchAppDetails(ctx context.Context, appId int, envId int) (AppView.AppDetailContainer, error)
	FetchAllDevtronAppsWithAtleastOneWorkflow(fetchAppListingRequest FetchAppListingRequest) ([]*AppView.AppResponseDto, error)
	FetchAppPolicyConsequences(appId int, envId int) (appBlockedState AppView.AppPolicyConsequence, err error)

	// ------------------

	FetchAppTriggerView(appId int) ([]AppView.TriggerView, error)
	FetchAppStageStatus(appId int, appType int) ([]AppView.AppStageStatus, error)

	FetchOtherEnvironment(ctx context.Context, appId int) ([]*AppView.Environment, error)
	FetchMinDetailOtherEnvironment(appId int) ([]*AppView.Environment, error)

	FetchEnvMinData(appIds []int) ([]*bean2.AppEnvNames, error)
	RedirectToLinkouts(Id int, appId int, envId int, podName string, containerName string) (string, error)
	ISLastReleaseStopType(appId, envId int) (bool, error)
	ISLastReleaseStopTypeV2(pipelineIds []int) (map[int]bool, error)
	GetReleaseCount(appId, envId int) (int, error)
	GetAllAppEnvsFromResourceNames(filterCriteria []string) ([]*bean2.AppEnvResponse, error)

	FetchAppsByEnvironmentV2(fetchAppListingRequest FetchAppListingRequest, w http.ResponseWriter, r *http.Request, token string) ([]*AppView.AppEnvironmentContainer, int, error)
	FetchOverviewAppsByEnvironment(envId, limit, offset int) (*OverviewAppsByEnvironmentBean, error)
	FetchAppsEnvContainers(envId int, appIds []int, limit, offset int) ([]*AppView.AppEnvironmentContainer, error)
}

const (
	APIVersionV1 string = "v1"
	APIVersionV2 string = "v2"
)

type FetchAppListingRequest struct {
	Environments      []int              `json:"environments"`
	Statuses          []string           `json:"statuses"`
	Teams             []int              `json:"teams"`
	TagFilters        []helper.TagFilter `json:"tagFilters"`
	AppNameSearch     string             `json:"appNameSearch"`
	SortOrder         helper.SortOrder   `json:"sortOrder"`
	SortBy            helper.SortBy      `json:"sortBy"`
	Offset            int                `json:"offset"`
	Size              int                `json:"size"`
	DeploymentGroupId int                `json:"deploymentGroupId"`
	Namespaces        []string           `json:"namespaces"` // {clusterId}_{namespace}
	AppStatuses       []string           `json:"appStatuses"`
	AppIds            []int              `json:"-"` // internal use only
	// IsClusterOrNamespaceSelected bool             `json:"isClusterOrNamespaceSelected"`
}
type AppNameTypeIdContainer struct {
	AppName string `json:"appName"`
	Type    string `json:"type"`
	AppId   int    `json:"appId"`
}
type CiArtifactWithParentArtifact struct {
	ParentCiArtifact int
	CiArtifactId     int
}

// NormalizeAndValidateTagFilters validates each tag filter row and normalizes value handling.
// Rules:
// 1) key must be present.
// 2) operator must be one of the supported backend operators.
// 3) for EXISTS/DOES_NOT_EXIST, value must not be sent.
// 4) for EQUALS/DOES_NOT_EQUAL/CONTAINS/DOES_NOT_CONTAIN, value must be non-empty.
func NormalizeAndValidateTagFilters(tagFilters []helper.TagFilter) ([]helper.TagFilter, error) {
	if len(tagFilters) == 0 {
		return tagFilters, nil
	}
	normalizedFilters := make([]helper.TagFilter, 0, len(tagFilters))
	for index, tagFilter := range tagFilters {
		tagFilter.Key = strings.TrimSpace(tagFilter.Key)
		if len(tagFilter.Key) == 0 {
			return nil, fmt.Errorf("tagFilters[%d].key is required", index)
		}
		if !tagFilter.Operator.IsValid() {
			return nil, fmt.Errorf("tagFilters[%d].operator is invalid: %s", index, tagFilter.Operator)
		}
		switch tagFilter.Operator {
		case helper.TagFilterOperatorExists, helper.TagFilterOperatorDoesNotExist:
			if tagFilter.Value != nil {
				return nil, fmt.Errorf("tagFilters[%d].value must be empty for operator %s", index, tagFilter.Operator)
			}
		default:
			if tagFilter.Value == nil || len(*tagFilter.Value) == 0 {
				return nil, fmt.Errorf("tagFilters[%d].value is required for operator %s", index, tagFilter.Operator)
			}
		}
		normalizedFilters = append(normalizedFilters, tagFilter)
	}
	return normalizedFilters, nil
}

func GetNamespaceClusterMapping(underscoreSeperatedClusterIdNamespaces []string) (namespaceClusterPair []*repository3.ClusterNamespacePair, clusterIds []int, err error) {
	for _, ns := range underscoreSeperatedClusterIdNamespaces {
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
			pair := &repository3.ClusterNamespacePair{
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
	appRepository                  app.AppRepository
	appListingRepository           repository.AppListingRepository
	appDetailsReadService          read4.AppDetailsReadService
	appListingViewBuilder          AppListingViewBuilder
	pipelineRepository             pipelineConfig.PipelineRepository
	cdWorkflowRepository           pipelineConfig.CdWorkflowRepository
	linkoutsRepository             repository.LinkoutsRepository
	pipelineOverrideRepository     chartConfig.PipelineOverrideRepository
	environmentRepository          repository3.EnvironmentRepository
	chartRepository                chartRepoRepository.ChartRepository
	ciPipelineRepository           pipelineConfig.CiPipelineRepository
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService
	userRepository                 userrepository.UserRepository
	deployedAppMetricsService      deployedAppMetrics.DeployedAppMetricsService
	ciArtifactRepository           repository.CiArtifactRepository
	appLabelRepository             pipelineConfig.AppLabelRepository
	globalTagService               globalTag.GlobalTagService
	envConfigOverrideReadService   read.EnvConfigOverrideService
	ciPipelineConfigReadService    ciConfig.CiPipelineConfigReadService
	deploymentConfigService        cdConfigRead.DeploymentConfigReadService

	//ent only
	helmAppReadService read2.HelmAppReadService
	helmAppService     service.HelmAppService
	globalEnvVariables *util3.GlobalEnvVariables
	chartReadService   read3.ChartReadService
}

func NewAppListingServiceImpl(Logger *zap.SugaredLogger,
	appListingRepository repository.AppListingRepository,
	appDetailsReadService read4.AppDetailsReadService,
	appRepository app.AppRepository,
	appListingViewBuilder AppListingViewBuilder, pipelineRepository pipelineConfig.PipelineRepository,
	linkoutsRepository repository.LinkoutsRepository, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, environmentRepository repository3.EnvironmentRepository,
	chartRepository chartRepoRepository.ChartRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService, userRepository userrepository.UserRepository,
	appLabelRepository pipelineConfig.AppLabelRepository,
	globalTagService globalTag.GlobalTagService,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService, ciArtifactRepository repository.CiArtifactRepository,
	envConfigOverrideReadService read.EnvConfigOverrideService,
	ciPipelineConfigReadService ciConfig.CiPipelineConfigReadService,
	deploymentConfigService cdConfigRead.DeploymentConfigReadService,
	helmAppReadService read2.HelmAppReadService,
	helmAppService service.HelmAppService,
	envVariables *util3.EnvironmentVariables,
	chartReadService read3.ChartReadService) *AppListingServiceImpl {
	return &AppListingServiceImpl{
		Logger:                         Logger,
		appListingRepository:           appListingRepository,
		appDetailsReadService:          appDetailsReadService,
		appRepository:                  appRepository,
		appListingViewBuilder:          appListingViewBuilder,
		pipelineRepository:             pipelineRepository,
		linkoutsRepository:             linkoutsRepository,
		cdWorkflowRepository:           cdWorkflowRepository,
		pipelineOverrideRepository:     pipelineOverrideRepository,
		environmentRepository:          environmentRepository,
		chartRepository:                chartRepository,
		ciPipelineRepository:           ciPipelineRepository,
		dockerRegistryIpsConfigService: dockerRegistryIpsConfigService,
		userRepository:                 userRepository,
		deployedAppMetricsService:      deployedAppMetricsService,
		ciArtifactRepository:           ciArtifactRepository,
		envConfigOverrideReadService:   envConfigOverrideReadService,
		appLabelRepository:             appLabelRepository,
		globalTagService:               globalTagService,
		ciPipelineConfigReadService:    ciPipelineConfigReadService,
		deploymentConfigService:        deploymentConfigService,
		helmAppReadService:             helmAppReadService,
		helmAppService:                 helmAppService,
		globalEnvVariables:             envVariables.GlobalEnvVariables,
		chartReadService:               chartReadService,
	}
}

const AcdInvalidAppErr = "invalid acd app name and env"
const NotDeployed = "Not Deployed"

type OverviewAppsByEnvironmentBean struct {
	EnvironmentId   int                                `json:"environmentId"`
	EnvironmentName string                             `json:"environmentName"`
	Namespace       string                             `json:"namespace"`
	ClusterName     string                             `json:"clusterName"`
	ClusterId       int                                `json:"clusterId"`
	Type            string                             `json:"environmentType"`
	Description     string                             `json:"description"`
	AppCount        int                                `json:"appCount"`
	Apps            []*AppView.AppEnvironmentContainer `json:"apps"`
	CreatedOn       string                             `json:"createdOn"`
	CreatedBy       string                             `json:"createdBy"`
}

const (
	Production    = "Production"
	NonProduction = "Non-Production"
)

func (impl AppListingServiceImpl) FetchEnvMinData(appIds []int) ([]*bean2.AppEnvNames, error) {
	if len(appIds) == 0 {
		return []*bean2.AppEnvNames{}, nil
	}
	apps, err := impl.appRepository.FindAppAndProjectByIdsIn(appIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching apps by ids", "err", err, "appIds", appIds)
		return nil, err
	}

	appNames := make([]string, 0, len(apps))
	for _, app := range apps {
		appNames = append(appNames, app.AppName)
	}
	return impl.appListingRepository.GetAllAppEnvsFromResourceNames(bean2.AppEnvFilterRequest{AppNames: appNames})
}

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
	var appIds []int
	envContainers, err := impl.FetchAppsEnvContainers(envId, appIds, limit, offset)
	if err != nil {
		impl.Logger.Errorw("failed to fetch env containers", "err", err, "envId", envId)
		return resp, err
	}
	artifactIds := make([]int, 0)
	for _, envContainer := range envContainers {
		lastDeployed, err := impl.appListingRepository.FetchLastDeployedImage(envContainer.AppId, envId)
		if err != nil {
			impl.Logger.Errorw("failed to fetch last deployed image", "err", err, "appId", envContainer.AppId, "envId", envId)
			return resp, err
		}

		if lastDeployed != nil {
			envContainer.LastDeployedImage = lastDeployed.LastDeployedImage
			envContainer.LastDeployedBy = lastDeployed.LastDeployedBy
			envContainer.CiArtifactId = lastDeployed.CiArtifactId
			artifactIds = append(artifactIds, lastDeployed.CiArtifactId)
		}
	}
	uniqueArtifacts := getUniqueArtifacts(artifactIds)

	artifactWithGitCommit, err := impl.generateArtifactIDCommitMap(uniqueArtifacts)
	if err != nil {
		impl.Logger.Errorw("failed to fetch Artifacts to git Triggers ", "envId", envId, "err", err)
		return resp, err
	}
	for _, envContainer := range envContainers {
		envContainer.Commits = []string{}
		if envContainer.CiArtifactId > 0 {
			if commits, ok := artifactWithGitCommit[envContainer.CiArtifactId]; ok && commits != nil {
				envContainer.Commits = commits
			}
		}
	}
	resp.Apps = envContainers
	return resp, err
}

func getUniqueArtifacts(artifactIds []int) (uniqueArtifactIds []int) {
	uniqueArtifactIds = make([]int, 0)

	uniqueArtifactMap := make(map[int]bool)

	for _, artifactId := range artifactIds {
		if ok := uniqueArtifactMap[artifactId]; !ok {
			uniqueArtifactIds = append(uniqueArtifactIds, artifactId)
			uniqueArtifactMap[artifactId] = true
		}
	}

	return uniqueArtifactIds
}

func (impl AppListingServiceImpl) FetchAppsEnvContainers(envId int, appIds []int, limit, offset int) ([]*AppView.AppEnvironmentContainer, error) {
	envContainers, err := impl.appListingRepository.FetchAppsEnvContainers(envId, appIds, limit, offset)
	if err != nil {
		impl.Logger.Errorw("failed to fetch environment containers", "err", err, "envId", envId)
		return nil, err
	}

	err = impl.updateAppStatusForHelmTypePipelines(envContainers)
	if err != nil {
		impl.Logger.Errorw("err, updateAppStatusForHelmTypePipelines", "envId", envId, "err", err)
		return nil, err
	}
	return envContainers, nil
}

func (impl AppListingServiceImpl) FetchAllDevtronAppsWithAtleastOneWorkflow(fetchAppListingRequest FetchAppListingRequest) ([]*AppView.AppResponseDto, error) {
	// Basic validation of pagination parameters.

	if fetchAppListingRequest.Offset < 0 {
		impl.Logger.Errorw("invalid offset provided", "offset", fetchAppListingRequest.Offset)
		return nil, fmt.Errorf("invalid offset provided")
	}
	appListingFilter := helper.AppListingFilter{
		AppNameSearch: fetchAppListingRequest.AppNameSearch,
		SortOrder:     fetchAppListingRequest.SortOrder,
		Offset:        fetchAppListingRequest.Offset,
		Size:          fetchAppListingRequest.Size,
	}
	impl.Logger.Debugw("Fetching all Devtron apps with at least one workflow", "filter", appListingFilter)
	res, err := impl.appListingRepository.FetchAllDevtronAppsWithAtleastOneWorkflow(appListingFilter)
	if err != nil {
		impl.Logger.Errorw("failed to fetch devtron apps", "err", err)
		return nil, err
	}
	resp := adapter.ConvertDbObjToRespDto(res)
	return resp, nil
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

func (impl AppListingServiceImpl) FetchJobs(fetchJobListingRequest FetchAppListingRequest) ([]*AppView.JobContainer, error) {
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
		return []*AppView.JobContainer{}, err
	}
	jobListingContainers, err := impl.appListingRepository.FetchJobs(appIds, jobListingFilter.AppStatuses, jobListingFilter.Environments, string(jobListingFilter.SortOrder))
	if err != nil {
		impl.Logger.Errorw("error in fetching job list", "error", err, jobListingFilter)
		return []*AppView.JobContainer{}, err
	}
	userEmailMap, err := impl.extractEmailIdFromUserId(jobListingContainers)
	if err != nil {
		impl.Logger.Errorw("Error in extractEmailIdFromUserId", "jobContainers", jobListingContainers, "err", err)
		return nil, err
	}

	CiPipelineIDs := GetCIPipelineIDs(jobListingContainers)
	JobsLastSucceededOnTime, err := impl.appListingRepository.FetchJobsLastSucceededOn(CiPipelineIDs)
	jobContainers := BuildJobListingResponse(jobListingContainers, JobsLastSucceededOnTime, userEmailMap)
	return jobContainers, nil
}

func (impl AppListingServiceImpl) FetchOverviewCiPipelines(jobId int) ([]*AppView.JobListingContainer, error) {
	jobCiContainers, err := impl.appListingRepository.FetchOverviewCiPipelines(jobId)
	if err != nil {
		impl.Logger.Errorw("error in fetching job container", "error", err, jobId)
		return []*AppView.JobListingContainer{}, err
	}
	return jobCiContainers, nil
}

func (impl AppListingServiceImpl) FetchJobCiPipelines() ([]*AppView.JobListingContainer, error) {
	jobCiContainers, err := impl.appListingRepository.FetchJobCiPipelines()
	if err != nil {
		impl.Logger.Errorw("error in fetching job ci pipelines", "error", err)
		return jobCiContainers, err
	}
	return jobCiContainers, nil
}

func (impl AppListingServiceImpl) FetchAppsByEnvironmentV2(fetchAppListingRequest FetchAppListingRequest, w http.ResponseWriter, r *http.Request, token string) ([]*AppView.AppEnvironmentContainer, int, error) {
	impl.Logger.Debug("reached at FetchAppsByEnvironment:")
	if len(fetchAppListingRequest.Namespaces) != 0 && len(fetchAppListingRequest.Environments) == 0 {
		return []*AppView.AppEnvironmentContainer{}, 0, nil
	}
	normalizedTagFilters, err := NormalizeAndValidateTagFilters(fetchAppListingRequest.TagFilters)
	if err != nil {
		return []*AppView.AppEnvironmentContainer{}, 0, err
	}
	fetchAppListingRequest.TagFilters = normalizedTagFilters

	// Currently AppStatus is available in Db for only ArgoApps
	// We fetch AppStatus on the fly for Helm Apps from scoop, So AppStatus filter will be applied in last
	// fun to check if "HIBERNATING" exists in fetchAppListingRequest.AppStatuses
	isFilteredOnHibernatingStatus := impl.isFilteredOnHibernatingStatus(fetchAppListingRequest)
	// remove ""HIBERNATING" from fetchAppListingRequest.AppStatuses
	appStatusesFilter := make([]string, 0)
	if isFilteredOnHibernatingStatus {
		appStatusesFilter = fetchAppListingRequest.AppStatuses
		fetchAppListingRequest.AppStatuses = []string{}
		// log updated filter
		impl.Logger.Debugw("HIBERNATING status found, updated filter", "filter", fetchAppListingRequest)
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
		TagFilters:        fetchAppListingRequest.TagFilters,
		AppIds:            fetchAppListingRequest.AppIds,
	}
	_, span := otel.Tracer("appListingRepository").Start(r.Context(), "FetchAppsByEnvironment")
	envContainers, appSize, err := impl.appListingRepository.FetchAppsByEnvironmentV2(appListingFilter)
	span.End()
	if err != nil {
		impl.Logger.Errorw("error in fetching app list", "error", err, "filter", appListingFilter)
		return []*AppView.AppEnvironmentContainer{}, appSize, err
	}

	envContainersMap := make(map[int][]*AppView.AppEnvironmentContainer)
	envIds := make([]int, 0)
	envsSet := make(map[int]bool)

	for _, container := range envContainers {
		if container.EnvironmentId != 0 {
			if _, ok := envContainersMap[container.EnvironmentId]; !ok {
				envContainersMap[container.EnvironmentId] = make([]*AppView.AppEnvironmentContainer, 0)
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
		return []*AppView.AppEnvironmentContainer{}, appSize, err
	}
	for _, info := range envClusterInfos {
		for _, container := range envContainersMap[info.Id] {
			container.Namespace = info.Namespace
			container.ClusterName = info.ClusterName
			container.EnvironmentName = info.Name
			container.IsVirtualEnvironment = info.IsVirtualEnvironment
		}
	}
	//Removed the below logic call due to slowness at zupee, to implement faster alternative
	//err = impl.updateAppStatusForHelmTypePipelines(envContainers)
	//if err != nil {
	//	impl.Logger.Errorw("error, UpdateAppStatusForHelmTypePipelines", "envIds", envIds, "err", err)
	//}

	// apply filter for "HIBERNATING" status
	if isFilteredOnHibernatingStatus {
		filteredContainers := make([]*AppView.AppEnvironmentContainer, 0)
		for _, container := range envContainers {
			if slices.Contains(appStatusesFilter, container.AppStatus) {
				filteredContainers = append(filteredContainers, container)
			}
		}
		envContainers = filteredContainers
		appSize = len(filteredContainers)
	}
	return envContainers, appSize, nil
}

func (impl AppListingServiceImpl) isFilteredOnHibernatingStatus(fetchAppListingRequest FetchAppListingRequest) bool {
	if fetchAppListingRequest.AppStatuses != nil && len(fetchAppListingRequest.AppStatuses) > 0 && slices.Contains(fetchAppListingRequest.AppStatuses, bean7.HIBERNATING) {
		return true
	}
	return false
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
		if slices.Contains([]string{cdWorkflow.WorkflowInitiated, cdWorkflow.WorkflowInQueue, cdWorkflow.WorkflowFailed}, cdWfr.Status) {
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
			if slices.Contains([]string{cdWorkflow.WorkflowInitiated, cdWorkflow.WorkflowInQueue}, cdWfr.Status) {
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

func (impl AppListingServiceImpl) BuildAppListingResponseV2(fetchAppListingRequest FetchAppListingRequest, envContainers []*AppView.AppEnvironmentContainer) ([]*AppView.AppContainer, error) {
	start := time.Now()
	appEnvMapping, err := impl.fetchACDAppStatusV2(fetchAppListingRequest, envContainers)
	middleware.AppListingDuration.WithLabelValues("fetchACDAppStatus", "devtron").Observe(time.Since(start).Seconds())
	if err != nil {
		impl.Logger.Errorw("error in fetching app statuses", "error", err)
		return []*AppView.AppContainer{}, err
	}
	start = time.Now()
	appContainerResponses, err := impl.appListingViewBuilder.BuildView(fetchAppListingRequest, appEnvMapping)
	middleware.AppListingDuration.WithLabelValues("buildView", "devtron").Observe(time.Since(start).Seconds())
	return appContainerResponses, err
}
func GetCIPipelineIDs(jobContainers []*AppView.JobListingContainer) []int {

	var ciPipelineIDs []int
	for _, jobContainer := range jobContainers {
		ciPipelineIDs = append(ciPipelineIDs, jobContainer.CiPipelineID)
	}
	return ciPipelineIDs
}
func BuildJobListingResponse(jobContainers []*AppView.JobListingContainer, JobsLastSucceededOnTime []*AppView.CiPipelineLastSucceededTime, userEmailMap map[int32]string) []*AppView.JobContainer {
	jobContainersMapping := make(map[int]AppView.JobContainer)
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
			val = AppView.JobContainer{}
			val.JobId = jobContainer.JobId
			val.JobName = jobContainer.JobName
			val.JobActualName = jobContainer.JobActualName
			val.ProjectId = jobContainer.ProjectId
			val.Description = AppView.GenericNoteResponseBean{Description: jobContainer.Description, CreatedBy: userEmailMap[jobContainer.CreatedBy]}
		}

		if len(val.JobCiPipelines) == 0 {
			val.JobCiPipelines = make([]AppView.JobCIPipeline, 0)
		}

		if jobContainer.CiPipelineID != 0 {
			ciPipelineObj := AppView.JobCIPipeline{
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

		if len(jobContainer.Description) >= 0 {
		}

		jobContainersMapping[jobContainer.JobId] = val

	}

	result := make([]*AppView.JobContainer, 0)
	for _, appId := range appIds {
		val := jobContainersMapping[appId]
		result = append(result, &val)
	}

	return result
}

func (impl AppListingServiceImpl) fetchACDAppStatusV2(fetchAppListingRequest FetchAppListingRequest, existingAppEnvContainers []*AppView.AppEnvironmentContainer) (map[string][]*AppView.AppEnvironmentContainer, error) {
	appEnvMapping := make(map[string][]*AppView.AppEnvironmentContainer)
	for _, env := range existingAppEnvContainers {
		appKey := strconv.Itoa(env.AppId) + "_" + env.AppName
		appEnvMapping[appKey] = append(appEnvMapping[appKey], env)
	}
	return appEnvMapping, nil
}

func (impl AppListingServiceImpl) FetchAppDetails(ctx context.Context, appId int, envId int) (AppView.AppDetailContainer, error) {
	pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
	if err != nil {
		impl.Logger.Errorw("error encountered in FetchAppDetail", "appId", appId, "envId", envId, "err", err)
		return AppView.AppDetailContainer{}, err
	}
	var appDetailContainer AppView.AppDetailContainer
	if len(pipelines) > 0 && pipelines[0].Environment.IsVirtualEnvironment {
		appDetailContainer, err = impl.appDetailsReadService.FetchAppDetailForVirtualEnvironment(ctx, appId, envId, pipelines[0].Id)
		if err != nil {
			impl.Logger.Errorw("error in fetching app detail", "error", err)
			return AppView.AppDetailContainer{}, err
		}
		imagePath := appDetailContainer.DeploymentDetailContainer.Image
		if len(imagePath) > 0 {
			imageMetadata, err := util3.ExtractImageRepoAndTag(imagePath)
			if err != nil {
				impl.Logger.Errorw("error in extracting image tag and data", "imagePath", imagePath, "err", err)
			}
			appDetailContainer.DeploymentDetailContainer.ImageTag = imageMetadata.Tag
			var pipeline *pipelineConfig.Pipeline
			for _, p := range pipelines {
				pipeline = p
				break
			}
			appDetailContainer.HelmPackageName = helm.BuildHelmPackageNameForDevtronApps(pipeline.App.AppName, pipeline.Environment.Name, imageMetadata.Tag, bean.CD_WORKFLOW_TYPE_DEPLOY)
		}
	} else {
		appDetailContainer, err = impl.appDetailsReadService.FetchAppDetail(ctx, appId, envId)
		if err != nil {
			impl.Logger.Errorw("error in fetching app detail", "error", err)
			return AppView.AppDetailContainer{}, err
		}
	}
	appDetailContainer.AppId = appId

	// set ifIpsAccess provided and relevant data
	appDetailContainer.IsExternalCi = true
	environment, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching env details, FetchAppDetails service", "error", err)
		return AppView.AppDetailContainer{}, err
	}
	appDetailContainer, err = impl.setIpAccessProvidedData(ctx, appDetailContainer, appDetailContainer.ClusterId, environment.IsVirtualEnvironment)
	if err != nil {
		return appDetailContainer, err
	}

	return appDetailContainer, nil
}

func (impl AppListingServiceImpl) FetchAppPolicyConsequences(appId int, envId int) (AppView.AppPolicyConsequence, error) {
	var appPolicyConsequence AppView.AppPolicyConsequence

	// check if mandatory tags are added as labels or not
	appLabels, err := impl.appLabelRepository.FindAllByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting app labels", "appId", appId)
		return appPolicyConsequence, err
	}

	teamId, err := impl.appRepository.GetTeamIdById(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting teamId", "appId", appId)
		return appPolicyConsequence, err
	}

	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.Logger.Errorw("error in getting env", "envId", envId)
		return appPolicyConsequence, err
	}
	// convert appLabels to map
	appLabelsMap := make(map[string]string)
	for _, appLabel := range appLabels {
		appLabelsMap[appLabel.Key] = appLabel.Value
	}

	err = impl.globalTagService.ValidateTagDeploymentPolicy(teamId, env.Default, appLabelsMap)
	if err != nil {
		impl.Logger.Errorw("error in validating mandatory labels for project", "appId", appId, "err", err)
		blockedState := AppView.BlockedState{
			IsBlocked: true,
			BlockedBy: AppView.MandatoryTags,
			Reason:    constants2.MANDATORY_TAG_DOES_NOT_EXIST_MESSAGE,
		}
		appPolicyConsequence.Cd = AppView.BlockedStateStage{
			Pre:  blockedState,
			Post: blockedState,
			Node: blockedState,
		}
		return appPolicyConsequence, nil
	}
	return appPolicyConsequence, nil
}

func (impl AppListingServiceImpl) setIpAccessProvidedData(ctx context.Context, appDetailContainer AppView.AppDetailContainer, clusterId int, isVirtualEnv bool) (AppView.AppDetailContainer, error) {
	ciPipelineId := appDetailContainer.CiPipelineId
	if ciPipelineId > 0 {
		_, span := otel.Tracer("orchestrator").Start(ctx, "ciPipelineRepository.FindWithMinDataByCiPipelineId")
		ciPipeline, err := impl.ciPipelineRepository.FindWithMinDataByCiPipelineId(ciPipelineId)
		span.End()
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in fetching ciPipeline", "ciPipelineId", ciPipelineId, "error", err)
			return AppView.AppDetailContainer{}, err
		}

		if ciPipeline != nil && ciPipeline.CiTemplate != nil && len(*ciPipeline.CiTemplate.DockerRegistryId) > 0 {
			if (!ciPipeline.IsExternal || ciPipeline.ParentCiPipeline != 0) && ciPipeline.PipelineType != string(buildCommonBean.LINKED_CD) {
				appDetailContainer.IsExternalCi = false
			}

			// get dockerRegistryId starts
			artifact, err := impl.ciArtifactRepository.Get(appDetailContainer.CiArtifactId)
			// artifact can be nil which is a valid case, so we are not returning the error
			dockerRegistryId, err := impl.ciPipelineConfigReadService.GetDockerRegistryIdForCiPipeline(ciPipelineId, artifact)
			if err != nil {
				impl.Logger.Errorw("error in fetching docker registry id", "ciPipelineId", ciPipelineId, "error", err)
				return AppView.AppDetailContainer{}, err
			}

			if dockerRegistryId == nil {
				impl.Logger.Errorw("docker registry id not found", "ciPipelineId", ciPipelineId)
				return appDetailContainer, nil
			}
			// get dockerRegistryId ends

			appDetailContainer.DockerRegistryId = *dockerRegistryId

			_, span = otel.Tracer("orchestrator").Start(ctx, "dockerRegistryIpsConfigService.IsImagePullSecretAccessProvided")
			// check ips access provided to this docker registry for that cluster
			ipsAccessProvided, err := impl.dockerRegistryIpsConfigService.IsImagePullSecretAccessProvided(*dockerRegistryId, clusterId, isVirtualEnv)
			span.End()
			if err != nil {
				impl.Logger.Errorw("error in checking if docker registry ips access provided", "dockerRegistryId", dockerRegistryId, "clusterId", clusterId, "error", err)
				return AppView.AppDetailContainer{}, err
			}
			appDetailContainer.IpsAccessProvided = ipsAccessProvided
		}
	}
	return appDetailContainer, nil
}

func (impl AppListingServiceImpl) FetchAppTriggerView(appId int) ([]AppView.TriggerView, error) {
	return impl.appListingRepository.FetchAppTriggerView(appId)
}

func (impl AppListingServiceImpl) FetchAppStageStatus(appId int, appType int) ([]AppView.AppStageStatus, error) {
	appStageStatuses, err := impl.appDetailsReadService.FetchAppStageStatus(appId, appType)
	return appStageStatuses, err
}

func (impl AppListingServiceImpl) generateArtifactIDCommitMap(artifactIds []int) (ciArtifactAndGitCommitsMap map[int][]string, err error) {

	if len(artifactIds) == 0 {
		impl.Logger.Errorw("error in getting the ArtifactIds", "ArtifactIds", artifactIds, "err", err)
		return make(map[int][]string), err
	}

	artifacts, err := impl.ciArtifactRepository.GetByIds(artifactIds)
	if err != nil {
		return make(map[int][]string), err
	}

	ciArtifactAndGitCommitsMap = make(map[int][]string)
	ciArtifactWithModificationMap := make(map[int][]repository.Modification)

	for _, artifact := range artifacts {
		materialInfo, err := repository.GetCiMaterialInfo(artifact.MaterialInfo, artifact.DataSource)
		if err != nil {
			impl.Logger.Errorw("error in getting the MaterialInfo", "ArtifactId", artifact.Id, "err", err)
			return make(map[int][]string), err
		}
		if len(materialInfo) == 0 {
			continue
		}
		for _, material := range materialInfo {
			ciArtifactWithModificationMap[artifact.Id] = append(ciArtifactWithModificationMap[artifact.Id], material.Modifications...)
		}
	}

	for artifactId, modifications := range ciArtifactWithModificationMap {

		gitCommits := make([]string, 0)

		for _, modification := range modifications {
			gitCommits = append(gitCommits, modification.Revision)
		}

		ciArtifactAndGitCommitsMap[artifactId] = gitCommits
	}

	return ciArtifactAndGitCommitsMap, nil
}

func (impl AppListingServiceImpl) FetchOtherEnvironment(ctx context.Context, appId int) ([]*AppView.Environment, error) {
	newCtx, span := otel.Tracer("appListingRepository").Start(ctx, "FetchOtherEnvironment")
	envs, err := impl.appListingRepository.FetchOtherEnvironment(appId)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return envs, err
	}
	appLevelInfraMetrics := true // default val, not being derived from DB. TODO: remove this from FE since this is derived from prometheus config at cluster level and this logic is already present at FE
	newCtx, span = otel.Tracer("deployedAppMetricsService").Start(newCtx, "GetMetricsFlagByAppId")
	appLevelAppMetrics, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	span.End()
	if err != nil {
		impl.Logger.Errorw("error, GetMetricsFlagByAppId", "err", err, "appId", appId)
		return envs, err
	}
	newCtx, span = otel.Tracer("chartRepository").Start(newCtx, "FindLatestChartForAppByAppId")
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(nil, appId)
	span.End()
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching latest chart", "err", err)
		return envs, err
	}

	ciArtifacts := make([]int, 0)
	for _, env := range envs {
		ciArtifacts = append(ciArtifacts, env.CiArtifactId)
	}

	uniqueArtifacts := getUniqueArtifacts(ciArtifacts)

	gitCommitsWithArtifacts, err := impl.generateArtifactIDCommitMap(uniqueArtifacts)
	if err != nil {
		impl.Logger.Errorw("Error in fetching the git commits of the ciArtifacts", "err", err, "ciArtifacts", ciArtifacts)
		return envs, err
	}
	for _, env := range envs {
		newCtx, span = otel.Tracer("envOverrideRepository").Start(newCtx, "FindLatestChartForAppByAppIdAndEnvId")
		envOverride, err := impl.envConfigOverrideReadService.FindLatestChartForAppByAppIdAndEnvId(nil, appId, env.EnvironmentId)
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
		if env.AppMetrics == nil {
			env.AppMetrics = &appLevelAppMetrics
		}

		if _, ok := gitCommitsWithArtifacts[env.CiArtifactId]; ok {
			env.Commits = gitCommitsWithArtifacts[env.CiArtifactId]
		} else {
			env.Commits = make([]string, 0)
		}
		env.InfraMetrics = &appLevelInfraMetrics // using default value, discarding value got from query
	}
	return envs, nil
}

func (impl AppListingServiceImpl) FetchMinDetailOtherEnvironment(appId int) ([]*AppView.Environment, error) {
	envs, err := impl.appListingRepository.FetchMinDetailOtherEnvironment(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return envs, err
	}
	appLevelInfraMetrics := true // default val, not being derived from DB. TODO: remove this from FE since this is derived from prometheus config at cluster level and this logic is already present at FE
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
	overrideChartRefIds, err := impl.envConfigOverrideReadService.FindChartRefIdsForLatestChartForAppByAppIdAndEnvIds(appId, envIds)
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
		env.InfraMetrics = &appLevelInfraMetrics // using default value, discarding value got from query
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

func (impl AppListingServiceImpl) GetAllAppEnvsFromResourceNames(filterCriteria []string) ([]*bean2.AppEnvResponse, error) {
	filter := bean2.AppEnvFilterRequest{}
	resp := []*bean2.AppEnvResponse{}
	for _, criteria := range filterCriteria {
		criteriaDecoder, err := util4.DecodeFilterCriteriaStringMulti(criteria)
		if err != nil {
			impl.Logger.Errorw("error encountered in applyFilterCriteriaOnResourceObjects", "filterCriteria", filterCriteria, "err", bean3.InvalidFilterCriteria)
			return nil, err
		}

		fullKindEnum := adapter.GetStringForKindAndSubKind(criteriaDecoder.Kind, criteriaDecoder.SubKind)
		switch fullKindEnum {
		case string(bean3.DevtronResourceDevtronApplicationFull):
			for _, val := range criteriaDecoder.Values {
				if strings.TrimSpace(val) != "" {
					filter.AppNames = append(filter.AppNames, val)
				}
			}
		case string(bean3.DevtronResourceEnvironment):
			for _, val := range criteriaDecoder.Values {
				isVirtual, virtualEnvId, _ := resourceQualifiers.IsVirtualResource(val)
				if isVirtual {
					if virtualEnvId == resourceQualifiers.AllExistingAndFutureNonProdEnvsInt {
						nonProdEnvs, err := impl.environmentRepository.FindAllActiveProdOrNonProd(false)
						if err != nil {
							impl.Logger.Errorw("Error in getting non prod envs", "err", err)
							return nil, err
						}
						for _, env := range nonProdEnvs {
							filter.EnvNames = append(filter.EnvNames, env.Name)
						}
					} else if virtualEnvId == resourceQualifiers.AllExistingAndFutureProdEnvsInt {
						prodEnvs, err := impl.environmentRepository.FindAllActiveProdOrNonProd(true)
						if err != nil {
							impl.Logger.Errorw("Error in getting non prod envs", "err", err)
							return nil, err
						}
						for _, env := range prodEnvs {
							filter.EnvNames = append(filter.EnvNames, env.Name)
						}
					}
				}
				if strings.TrimSpace(val) != "" && !slices.Contains(filter.EnvNames, val) {
					filter.EnvNames = append(filter.EnvNames, val)
				}
			}

		case string(bean3.DevtronResourceCluster):
			for _, val := range criteriaDecoder.Values {
				if strings.TrimSpace(val) != "" {
					filter.ClusterNames = append(filter.ClusterNames, val)
				}
			}
		case string(bean3.DevtronResourceProject):
			for _, val := range criteriaDecoder.Values {
				if strings.TrimSpace(val) != "" {
					filter.TeamNames = append(filter.TeamNames, val)
				}
			}
		}
	}

	appEnvs, err := impl.appListingRepository.GetAllAppEnvsFromResourceNames(filter)
	if err != nil {
		impl.Logger.Errorw("error in getting app-envs", "err", err)
		return nil, err
	}
	appIdEnvMap := make(map[int][]*bean2.AppEnvResponse, 0)
	appIdAppMap := make(map[int]*bean2.AppEnvResponse, 0)
	processedEnvs := map[int]map[int]bool{}
	for _, appEnv := range appEnvs {
		if processedEnvs[appEnv.AppId] == nil {
			processedEnvs[appEnv.AppId] = map[int]bool{}
		}

		if appIdAppMap[appEnv.AppId] == nil {
			appResKind := &bean2.AppEnvResponse{}
			appResKind.Kind = string(bean3.DevtronResourceDevtronApplicationFull)
			appResKind.Id = appEnv.AppId
			appResKind.Identifier = appEnv.AppName
			appIdAppMap[appEnv.AppId] = appResKind
		}

		if !processedEnvs[appEnv.AppId][appEnv.EnvironmentId] && appEnv.EnvironmentId != 0 {
			processedEnvs[appEnv.AppId][appEnv.EnvironmentId] = true
			envResKind := &bean2.AppEnvResponse{}
			envResKind.Kind = string(bean3.DevtronResourceEnvironment)
			envResKind.Id = appEnv.EnvironmentId
			envResKind.Identifier = appEnv.EnvironmentName

			if appIdEnvMap[appEnv.AppId] == nil {
				appIdEnvMap[appEnv.AppId] = make([]*bean2.AppEnvResponse, 0)
			}
			appIdEnvMap[appEnv.AppId] = append(appIdEnvMap[appEnv.AppId], envResKind)
		}

	}

	for appId, appDetail := range appIdAppMap {
		if appIdEnvMap[appId] == nil || len(appIdEnvMap[appId]) == 0 {
			appDetail.RelatedObjects = []*bean2.AppEnvResponse{}
		} else {
			appDetail.RelatedObjects = appIdEnvMap[appId]
		}
		resp = append(resp, appDetail)
	}
	return resp, nil
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

func (impl AppListingServiceImpl) extractEmailIdFromUserId(jobContainers []*AppView.JobListingContainer) (map[int32]string, error) {
	var userIds []int32
	userEmailMap := make(map[int32]string)
	for _, job := range jobContainers {
		if job.CreatedBy != 0 {
			userIds = append(userIds, job.CreatedBy)
		}
	}
	if len(userIds) > 0 {
		users, err := impl.userRepository.GetByIds(userIds)
		if err != nil {
			impl.Logger.Errorw("Error in getting users", "userIds", userIds, "err", err)
			return userEmailMap, err
		}
		for _, user := range users {
			userEmailMap[user.Id] = user.EmailId
		}
	}
	return userEmailMap, nil
}
