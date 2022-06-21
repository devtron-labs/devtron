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
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/util/argo"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/devtron-labs/devtron/api/bean"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/prometheus"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"go.uber.org/zap"
)

type AppListingService interface {
	FetchAppsByEnvironment(fetchAppListingRequest FetchAppListingRequest, w http.ResponseWriter, r *http.Request, token string) ([]*bean.AppEnvironmentContainer, error)
	BuildAppListingResponse(fetchAppListingRequest FetchAppListingRequest, envContainers []*bean.AppEnvironmentContainer) ([]*bean.AppContainer, error)

	FetchAppDetails(appId int, envId int) (bean.AppDetailContainer, error)

	PodCountByAppLabel(appLabel string, namespace string, env string, proEndpoint string) int
	PodListByAppLabel(appLabel string, namespace string, env string, proEndpoint string) map[string]string

	// below 4 functions used for pod level cpu and memory usage
	CpuUsageGroupByPod(namespace string, env string, proEndpoint string) map[string]string
	CpuRequestGroupByPod(namespace string, env string, proEndpoint string) map[string]string
	MemoryUsageGroupByPod(namespace string, env string, proEndpoint string) map[string]string
	MemoryRequestGroupByPod(namespace string, env string, proEndpoint string) map[string]string

	//Currently not in use
	CpuUsageGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string
	CpuRequestGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string
	MemoryUsageGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string
	MemoryRequestGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string

	//Currently not in use (intent to fetch graph data from prometheus)
	CpuUsageGroupByPodGraph(podName string, namespace string, env string, proEndpoint string, r v1.Range) map[string][]interface{}
	MemoryUsageGroupByPodGraph(podName string, namespace string, env string, proEndpoint string, r v1.Range) map[string][]interface{}
	GraphAPI(appId int, envId int) error

	FetchAppTriggerView(appId int) ([]bean.TriggerView, error)
	FetchAppStageStatus(appId int) ([]bean.AppStageStatus, error)

	FetchOtherEnvironment(appId int) ([]*bean.Environment, error)
	RedirectToLinkouts(Id int, appId int, envId int, podName string, containerName string) (string, error)
	GetLastDeploymentStatusesByAppNames(appNames []string) ([]repository.DeploymentStatus, error)
	GetLastDeploymentStatuses() (map[string]repository.DeploymentStatus, error)
	ISLastReleaseStopType(appId, envId int) (bool, error)
	ISLastReleaseStopTypeV2(pipelineIds []int) (map[int]bool, error)
	GetReleaseCount(appId, envId int) (int, error)
}

const (
	Initiate              string = "Initiate"
	ScalingReplicaSetDown string = "ScalingReplicaSetDown"
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
	Namespaces        []string         `json:"namespaces"` //{clusterId}_{namespace}

}

func (req FetchAppListingRequest) GetNamespaceClusterMapping() (namespaceClusterPair []*repository2.ClusterNamespacePair, clusterIds []int, err error) {
	for _, ns := range req.Namespaces {
		items := strings.Split(ns, "_")
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
	Logger                     *zap.SugaredLogger
	application                application2.ServiceClient
	appRepository              app.AppRepository
	appListingRepository       repository.AppListingRepository
	appListingViewBuilder      AppListingViewBuilder
	pipelineRepository         pipelineConfig.PipelineRepository
	cdWorkflowRepository       pipelineConfig.CdWorkflowRepository
	linkoutsRepository         repository.LinkoutsRepository
	appLevelMetricsRepository  repository.AppLevelMetricsRepository
	envLevelMetricsRepository  repository.EnvLevelAppMetricsRepository
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository
	environmentRepository      repository2.EnvironmentRepository
	argoUserService            argo.ArgoUserService
}

func NewAppListingServiceImpl(Logger *zap.SugaredLogger, appListingRepository repository.AppListingRepository,
	application application2.ServiceClient, appRepository app.AppRepository,
	appListingViewBuilder AppListingViewBuilder, pipelineRepository pipelineConfig.PipelineRepository,
	linkoutsRepository repository.LinkoutsRepository, appLevelMetricsRepository repository.AppLevelMetricsRepository,
	envLevelMetricsRepository repository.EnvLevelAppMetricsRepository, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, environmentRepository repository2.EnvironmentRepository,
	argoUserService argo.ArgoUserService) *AppListingServiceImpl {
	serviceImpl := &AppListingServiceImpl{
		Logger:                     Logger,
		appListingRepository:       appListingRepository,
		application:                application,
		appRepository:              appRepository,
		appListingViewBuilder:      appListingViewBuilder,
		pipelineRepository:         pipelineRepository,
		linkoutsRepository:         linkoutsRepository,
		appLevelMetricsRepository:  appLevelMetricsRepository,
		envLevelMetricsRepository:  envLevelMetricsRepository,
		cdWorkflowRepository:       cdWorkflowRepository,
		pipelineOverrideRepository: pipelineOverrideRepository,
		environmentRepository:      environmentRepository,
		argoUserService:            argoUserService,
	}
	return serviceImpl
}

const AcdInvalidAppErr = "invalid acd app name and env"
const NotDeployed = "Not Deployed"

func (impl AppListingServiceImpl) FetchAppsByEnvironment(fetchAppListingRequest FetchAppListingRequest, w http.ResponseWriter, r *http.Request, token string) ([]*bean.AppEnvironmentContainer, error) {
	impl.Logger.Debug("reached at FetchAppsByEnvironment:")
	// TODO: check statuses
	mappings, clusterIds, err := fetchAppListingRequest.GetNamespaceClusterMapping()
	if err != nil {
		impl.Logger.Errorw("error in fetching app list", "error", err)
		return []*bean.AppEnvironmentContainer{}, err
	}
	if len(mappings) > 0 {
		envs, err := impl.environmentRepository.FindByClusterIdAndNamespace(mappings)
		if err != nil {
			impl.Logger.Errorw("error in cluster ns mapping")
			return []*bean.AppEnvironmentContainer{}, err
		}
		for _, env := range envs {
			fetchAppListingRequest.Environments = append(fetchAppListingRequest.Environments, env.Id)
		}
	}
	if len(clusterIds) > 0 {
		envs, err := impl.environmentRepository.FindByClusterIds(clusterIds)
		if err != nil {
			impl.Logger.Errorw("error in cluster ns mapping")
			return []*bean.AppEnvironmentContainer{}, err
		}
		for _, env := range envs {
			fetchAppListingRequest.Environments = append(fetchAppListingRequest.Environments, env.Id)
		}

	}
	if (len(clusterIds) > 0 || len(mappings) > 0) && len(fetchAppListingRequest.Environments) == 0 {
		// no result when no matching cluster and env
		return []*bean.AppEnvironmentContainer{}, nil
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
	}
	envContainers, err := impl.appListingRepository.FetchAppsByEnvironment(appListingFilter)
	if err != nil {
		impl.Logger.Errorw("error in fetching app list", "error", err)
		return []*bean.AppEnvironmentContainer{}, err
	}
	return envContainers, err
}

func (impl AppListingServiceImpl) ISLastReleaseStopType(appId, envId int) (bool, error) {
	override, err := impl.pipelineOverrideRepository.GetLatestRelease(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting last release")
		return false, err
	} else if util.IsErrNoRows(err) {
		return false, nil
	} else {
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

func (impl AppListingServiceImpl) BuildAppListingResponse(fetchAppListingRequest FetchAppListingRequest, envContainers []*bean.AppEnvironmentContainer) ([]*bean.AppContainer, error) {
	appEnvMapping, err := impl.fetchACDAppStatus(fetchAppListingRequest, envContainers)
	if err != nil {
		impl.Logger.Errorw("error in fetching app statuses", "error", err)
		return []*bean.AppContainer{}, err
	}
	appContainerResponses, err := impl.appListingViewBuilder.BuildView(fetchAppListingRequest, appEnvMapping)
	return appContainerResponses, err
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
	deploymentStatuses, err := impl.GetLastDeploymentStatusesByAppNames(appNames)
	if err != nil {
		impl.Logger.Error(err)
		return map[string][]*bean.AppEnvironmentContainer{}, err
	}
	existingAppEnvStatusMapping := make(map[string]string)
	for _, ds := range deploymentStatuses {
		if _, ok := existingAppEnvStatusMapping[ds.AppName]; ok {
			continue
		}
		existingAppEnvStatusMapping[ds.AppName] = ds.Status
	}

	appEnvPipelinesMap := make(map[string][]*pipelineConfig.Pipeline)
	appEnvCdWorkflowMap := make(map[string]*pipelineConfig.CdWorkflow)
	appEnvCdWorkflowRunnerMap := make(map[int][]*pipelineConfig.CdWorkflowRunner)

	//get all the active cd pipelines
	if len(pipelineIds) > 0 {
		pipelinesAll, err := impl.pipelineRepository.FindByIdsIn(pipelineIds) //TODO - OPTIMIZE 1
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("err", err)
			return nil, err
		}
		//here to build a map of pipelines list for each (appId and envId)
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
		cdWorkflowAll, err := impl.cdWorkflowRepository.FindLatestCdWorkflowByPipelineIdV2(pipelineIds) //TODO - OPTIMIZE 2
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
				//if no cd wf found for appid-envid, add it into map with nil
				if _, ok := appEnvCdWorkflowMap[key]; !ok {
					appEnvCdWorkflowMap[key] = nil
				}
			}
		}

		//fetch all the cd workflow runner from cdWF ids,
		cdWorkflowRunnersAll, err := impl.cdWorkflowRepository.FindWorkflowRunnerByCdWorkflowId(wfIds) //TODO - OPTIMIZE 3
		if err != nil {
			impl.Logger.Errorw("error in getting wf", "err", err)
		}
		//build a map with key cdWF containing cdWFRunner List, which are later put in map for further requirement
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
		appEnvKey := env.AppName + "-" + env.EnvironmentName
		status, ok := existingAppEnvStatusMapping[appEnvKey]
		if !ok || env.DataSource == "" {
			status = NotDeployed
		}
		env.Status = status

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
						status = application2.HIBERNATING
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

func (impl AppListingServiceImpl) GetLastDeploymentStatusesByAppNames(appNames []string) ([]repository.DeploymentStatus, error) {
	deploymentStatuses, err := impl.appListingRepository.FindLastDeployedStatuses(appNames)
	if err != nil {
		return []repository.DeploymentStatus{}, err
	}
	return deploymentStatuses, nil
}

func (impl AppListingServiceImpl) GetLastDeploymentStatuses() (map[string]repository.DeploymentStatus, error) {
	deploymentStatuses, err := impl.appListingRepository.FindLastDeployedStatusesForAllApps()
	if err != nil {
		return map[string]repository.DeploymentStatus{}, err
	}
	existingAppEnvStatusMapping := make(map[string]repository.DeploymentStatus)
	for _, ds := range deploymentStatuses {
		if _, ok := existingAppEnvStatusMapping[ds.AppName]; ok {
			continue
		}
		existingAppEnvStatusMapping[ds.AppName] = ds
	}
	return existingAppEnvStatusMapping, nil
}

func (impl AppListingServiceImpl) getAppACDStatus(env bean.AppEnvironmentContainer, w http.ResponseWriter, r *http.Request, token string) (string, error) {
	if len(env.AppName) > 0 && len(env.EnvironmentName) > 0 {
		acdAppName := env.AppName + "-" + env.EnvironmentName
		query := &application.ResourcesQuery{
			ApplicationName: &acdAppName,
		}
		ctx, cancel := context.WithCancel(r.Context())
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}
		defer cancel()
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.Logger.Errorw("error in getting acd token", "err", err)
			return "", err
		}
		ctx = context.WithValue(ctx, "token", acdToken)
		impl.Logger.Debugf("Getting status for app %s in env %s", env.AppId, env.EnvironmentId)
		start := time.Now()
		resp, err := impl.application.ResourceTree(ctx, query)
		elapsed := time.Since(start)
		impl.Logger.Debugf("Time elapsed %s in fetching application %s for environment %s", elapsed, env.AppId, env.EnvironmentId)
		if err != nil {
			impl.Logger.Errorw("error fetching resource tree", "error", err)
			err = &util.ApiError{
				Code:            constants.AppDetailResourceTreeNotFound,
				InternalMessage: "app detail fetched, failed to get resource tree from acd",
				UserMessage:     "app detail fetched, failed to get resource tree from acd",
			}
			return "", err
		}
		return resp.Status, nil
	}
	impl.Logger.Error("invalid acd app name and env ", env.AppName, " - ", env.EnvironmentName)
	return "", errors.New(AcdInvalidAppErr)
}

// TODO: Status mapping
func (impl AppListingServiceImpl) adaptStatusForView(status string) string {
	return status
}

func (impl AppListingServiceImpl) FetchAppDetails(appId int, envId int) (bean.AppDetailContainer, error) {
	appDetailContainer, err := impl.appListingRepository.FetchAppDetail(appId, envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching app detail", "error", err)
		return bean.AppDetailContainer{}, err
	}

	var appMetrics bool
	var infraMetrics bool
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in app metrics app level flag", "error", err)
		return bean.AppDetailContainer{}, err
	} else if appLevelMetrics != nil {
		appMetrics = appLevelMetrics.AppMetrics
		infraMetrics = appLevelMetrics.InfraMetrics
	}

	i := 0
	for _, env := range appDetailContainer.Environments {
		var envLevelMetrics *bool
		var envLevelInfraMetrics *bool
		envLevelAppMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(appId, env.EnvironmentId)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in app metrics env level flag", "error", err)
			return bean.AppDetailContainer{}, err
		}

		if envLevelAppMetrics != nil && envLevelAppMetrics.Id != 0 && envLevelAppMetrics.AppMetrics != nil {
			envLevelMetrics = envLevelAppMetrics.AppMetrics
		} else {
			envLevelMetrics = &appMetrics
		}
		if envLevelAppMetrics != nil && envLevelAppMetrics.Id != 0 && envLevelAppMetrics.InfraMetrics != nil {
			envLevelInfraMetrics = envLevelAppMetrics.InfraMetrics
		} else {
			envLevelInfraMetrics = &infraMetrics
		}
		appDetailContainer.Environments[i].AppMetrics = envLevelMetrics
		appDetailContainer.Environments[i].InfraMetrics = envLevelInfraMetrics
		i++
	}

	linkoutsModel, err := impl.linkoutsRepository.FetchLinkoutsByAppIdAndEnvId(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching linkouts", "error", err)
		return bean.AppDetailContainer{}, err
	}
	var linkouts []bean.LinkOuts
	for _, linkout := range linkoutsModel {
		linkouts = append(linkouts, bean.LinkOuts{Id: linkout.Id, Name: linkout.Name})
	}

	appDetailContainer.LinkOuts = linkouts
	appDetailContainer.AppId = appId

	envModel, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching environment", "error", err)
		return bean.AppDetailContainer{}, err
	}
	appDetailContainer.K8sVersion = envModel.Cluster.K8sVersion
	appDetailContainer.ClusterId = envModel.ClusterId
	return appDetailContainer, nil
}

//Return only a integer value pod count, aggregated of all the pod inside a app
//(includes all the pods running different cd pipeline for same app)
func (impl AppListingServiceImpl) PodCountByAppLabel(appLabel string, namespace string, env string, proEndpoint string) int {
	if appLabel == "" || namespace == "" || proEndpoint == "" || env == "" {
		impl.Logger.Warnw("not a complete data found for prometheus call", "missing", "AppName or namespace or prometheus url or env")
		return 0
	}

	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return 0
	}

	podCountQuery := "count(kube_pod_labels{label_app='" + appLabel + "', namespace='" + namespace + "'})"
	out, _, err := prometheusAPI.Query(context.Background(), podCountQuery, time.Now())
	if err != nil {
		impl.Logger.Errorw("pod count query failed in prometheus:", "error", err)
		return 0
	}
	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("pod count data marshal failed:", "error", err)
		return 0
	}

	podCount := 0
	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("pod count data unmarshal failed: ", "error", err)
		return 0
	}
	for _, value := range resultMap {
		data := value.([]interface{})

		for _, item := range data {

			ito := item
			for k, v := range ito.(map[string]interface{}) {
				if k == "value" {
					vArr := v.([]interface{})
					//t := (vArr[1].(string))
					feetInt, err := strconv.Atoi(vArr[1].(string))
					if err != nil {
						feetInt = 0
						impl.Logger.Errorw("casting error", "err", err)
					}
					podCount = feetInt
				}
			}
		}
	}
	return podCount
}

//Returns map of running pod names
func (impl AppListingServiceImpl) PodListByAppLabel(appLabel string, namespace string, env string, proEndpoint string) map[string]string {
	response := make(map[string]interface{})
	podList := make(map[string]string)
	resultMap := make(map[string]interface{})
	if appLabel == "" || namespace == "" || proEndpoint == "" || env == "" {
		impl.Logger.Warnw("not a complete data found for prometheus call", "missing", "AppName or namespace or prometheus url or env")
		return podList
	}

	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return podList
	}

	podCountQuery := "kube_pod_labels{label_app='" + appLabel + "', namespace='" + namespace + "'}"
	out, _, err := prometheusAPI.Query(context.Background(), podCountQuery, time.Now())
	if err != nil {
		impl.Logger.Errorw("pod list query failed in prometheus:", "error", err)
		return podList
	}

	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("pod count data unmarshal failed:", "error", err)
		return podList
	}

	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("pod count data unmarshal failed:", "error", err)
		return podList
	}
	for _, value := range resultMap {
		if value != nil {
			data := value.([]interface{})

			for _, item := range data {

				ito := item
				for k, v := range ito.(map[string]interface{}) {
					if k == "metric" {
						vMap := v.(map[string]interface{})
						key := vMap["pod"].(string)
						podList[key] = "1"
					}
					if k == "value" {
					}
				}
			}
		}
	}
	return podList
}

func (impl AppListingServiceImpl) CpuUsageGroupByPod(namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing cpuUsageGroupByPod:")
	cpuUsageMetric := make(map[string]string)

	if namespace == "" || proEndpoint == "" || env == "" {
		impl.Logger.Warnw("not a complete data found for prometheus call", "missing", "AppName or namespace or prometheus url or env")
		return cpuUsageMetric
	}

	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return cpuUsageMetric
	}

	query := "sum(rate (container_cpu_usage_seconds_total{image!='',pod_name!='',container_name!='POD',namespace='" + namespace + "'}[1m])) by (pod_name)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in getting CpuUsageGroupByPod:", "error", err)
		return cpuUsageMetric
	}

	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal CpuUsageGroupByPod:", "error", err)
		return cpuUsageMetric
	}

	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal CpuUsageGroupByPod:", "error", err)
		return cpuUsageMetric
	}

	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod_name"].(string)
					cpuUsageMetric[key] = "1.0"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := cpuUsageMetric[temp]; ok {
						cpuUsageMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}
	return cpuUsageMetric
}

func (impl AppListingServiceImpl) CpuRequestGroupByPod(namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing cpuUsageGroupByPod:")
	cpuRequestMetric := make(map[string]string)

	if namespace == "" || proEndpoint == "" || env == "" {
		impl.Logger.Warnw("not a complete data found for prometheus call", "missing", "AppName or namespace or prometheus url or env")
		return cpuRequestMetric
	}

	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return cpuRequestMetric
	}

	query := "sum(kube_pod_container_resource_requests_cpu_cores{namespace='" + namespace + "'}) by (pod)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return cpuRequestMetric
	}

	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return cpuRequestMetric
	}

	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return cpuRequestMetric
	}
	for _, value := range resultMap {
		data := value.([]interface{})

		for _, item := range data {

			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod"].(string)
					cpuRequestMetric[key] = "1"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := cpuRequestMetric[temp]; ok {
						cpuRequestMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}
	return cpuRequestMetric
}

func (impl AppListingServiceImpl) MemoryUsageGroupByPod(namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing memoryUsageGroupByPod")
	memoryUsageMetric := make(map[string]string)

	if namespace == "" || proEndpoint == "" || env == "" {
		impl.Logger.Warnw("not a complete data found for prometheus call", "missing", "AppName or namespace or prometheus url or env")
		return memoryUsageMetric
	}

	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return memoryUsageMetric
	}

	query := "sum(container_memory_usage_bytes{container_name!='POD', container_name!='', namespace='" + namespace + "'}) by (pod_name)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return memoryUsageMetric
	}
	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return memoryUsageMetric
	}
	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return memoryUsageMetric
	}
	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {

			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod_name"].(string)
					memoryUsageMetric[key] = "1"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := memoryUsageMetric[temp]; ok {
						memoryUsageMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}
	return memoryUsageMetric
}

func (impl AppListingServiceImpl) MemoryRequestGroupByPod(namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing memoryRequestGroupByPod")
	memoryRequestMetric := make(map[string]string)
	if namespace == "" || proEndpoint == "" || env == "" {
		impl.Logger.Warnw("not a complete data found for prometheus call", "missing", "AppName or namespace or prometheus url or env")
		return memoryRequestMetric
	}

	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return memoryRequestMetric
	}

	query := "sum(kube_pod_container_resource_requests_memory_bytes{container!='POD', container!='', namespace='" + namespace + "'}) by (pod)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return memoryRequestMetric
	}

	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return memoryRequestMetric
	}

	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return memoryRequestMetric
	}

	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod"].(string)
					memoryRequestMetric[key] = "1"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := memoryRequestMetric[temp]; ok {
						memoryRequestMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}
	return memoryRequestMetric
}

//Deprecated: Currently not in use
func (impl AppListingServiceImpl) CpuUsageGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing cpuUsageGroupByPod:")
	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	cpuUsageMetric := make(map[string]string)

	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return cpuUsageMetric
	}

	query := "sum(rate(container_cpu_usage_seconds_total{image!='', pod_name='" + podName + "',container_name!='POD', namespace='" + podName + "'}[1m])) by (container_name)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return cpuUsageMetric
	}
	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return cpuUsageMetric
	}

	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return cpuUsageMetric
	}

	for _, value := range resultMap {
		data := value.([]interface{})

		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod_name"].(string)
					cpuUsageMetric[key] = "1"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := cpuUsageMetric[temp]; ok {
						cpuUsageMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}

	return cpuUsageMetric
}

//Deprecated: Currently not in use
func (impl AppListingServiceImpl) CpuRequestGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing cpuUsageGroupByPod:")
	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	cpuRequestMetric := make(map[string]string)

	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return cpuRequestMetric
	}

	query := "sum(kube_pod_container_resource_requests_cpu_cores{namespace='" + namespace + "',pod='" + podName + "'}) by (container)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return cpuRequestMetric
	}

	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return cpuRequestMetric
	}

	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return cpuRequestMetric
	}

	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod"].(string)
					cpuRequestMetric[key] = "1"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := cpuRequestMetric[temp]; ok {
						cpuRequestMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}
	return cpuRequestMetric
}

//Deprecated: Currently not in use
func (impl AppListingServiceImpl) MemoryUsageGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing memoryUsageGroupByPod")
	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	memoryUsageMetric := make(map[string]string)

	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return memoryUsageMetric
	}

	query := "sum(container_memory_usage_bytes{container_name!='POD', container_name!='',pod_name='" + podName + "', namespace='" + namespace + "'}) by (container_name)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return memoryUsageMetric
	}
	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return memoryUsageMetric
	}

	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return memoryUsageMetric
	}
	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod_name"].(string)
					memoryUsageMetric[key] = "1"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := memoryUsageMetric[temp]; ok {
						memoryUsageMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}
	return memoryUsageMetric
}

//Deprecated: Currently not in use
func (impl AppListingServiceImpl) MemoryRequestGroupByContainer(podName string, namespace string, env string, proEndpoint string) map[string]string {
	impl.Logger.Debug("executing memoryRequestGroupByPod")
	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	memoryRequestMetric := make(map[string]string)
	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return memoryRequestMetric
	}

	query := "sum(kube_pod_container_resource_requests_memory_bytes{container!='POD', container!='',pod='" + podName + "', namespace='" + namespace + "'}) by (container)"
	out, _, err := prometheusAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return memoryRequestMetric
	}

	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return memoryRequestMetric
	}

	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return memoryRequestMetric
	}
	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod"].(string)
					memoryRequestMetric[key] = "1"
					temp = key
				}
				if k == "value" {
					vArr := v.([]interface{})
					if _, ok := memoryRequestMetric[temp]; ok {
						memoryRequestMetric[temp] = vArr[1].(string)
					}
				}
			}
		}
	}
	return memoryRequestMetric
}

//Deprecated: Currently not in use (intent to fetch graph data from prometheus)
func (impl AppListingServiceImpl) CpuUsageGroupByPodGraph(podName string, namespace string, env string, proEndpoint string, r v1.Range) map[string][]interface{} {
	impl.Logger.Debug("executing CpuUsageGroupByPodGraph:")
	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	cpuUsageMetric := make(map[string][]interface{})

	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return cpuUsageMetric
	}

	query := "sum(rate(container_cpu_usage_seconds_total{namespace='" + namespace + "', container_name!='POD'}[1m])) by (pod_name)"
	time1 := time.Now()
	r1 := v1.Range{
		Start: time1.Add(-time.Hour),
		End:   time1,
		Step:  time.Minute,
	}
	out, _, err := prometheusAPI.QueryRange(context.Background(), query, r1)
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return cpuUsageMetric
	}

	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return cpuUsageMetric
	}
	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return cpuUsageMetric
	}
	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod_name"].(string)
					cpuUsageMetric[key] = nil
					temp = key
				}
				if k == "values" {
					vArr := v.([]interface{})
					if _, ok := cpuUsageMetric[temp]; ok {
						cpuUsageMetric[temp] = vArr
					}
				}
			}
		}
	}
	return cpuUsageMetric
}

//Deprecated: Currently not in use (intent to fetch graph data from prometheus)
func (impl AppListingServiceImpl) MemoryUsageGroupByPodGraph(podName string, namespace string, env string, proEndpoint string, r v1.Range) map[string][]interface{} {
	impl.Logger.Debug("executing MemoryUsageGroupByPodGraph")
	prometheusAPI, err := prometheus.ContextByEnv(env, proEndpoint)
	memoryUsageMetric := make(map[string][]interface{})

	if err != nil {
		impl.Logger.Errorw("error in getting prometheus api client:", "error", err)
		return memoryUsageMetric
	}

	query := "sum(container_memory_usage_bytes{namespace='" + namespace + "', container_name!='POD', container_name!=''}) by (pod_name)"
	time1 := time.Now()
	r1 := v1.Range{
		Start: time1.Add(-time.Hour),
		End:   time1,
		Step:  time.Minute,
	}
	out, _, err := prometheusAPI.QueryRange(context.Background(), query, r1)
	if err != nil {
		impl.Logger.Errorw("error in prometheus query:", "error", err)
		return memoryUsageMetric
	}
	response := make(map[string]interface{})
	response["data"] = out
	resJson, err := json.Marshal(response)
	if err != nil {
		impl.Logger.Errorw("error in marshal:", "error", err)
		return memoryUsageMetric
	}
	resultMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resJson), &resultMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal:", "error", err)
		return memoryUsageMetric
	}
	for _, value := range resultMap {
		data := value.([]interface{})
		for _, item := range data {
			ito := item
			temp := ""
			for k, v := range ito.(map[string]interface{}) {
				if k == "metric" {
					vMap := v.(map[string]interface{})
					key := vMap["pod_name"].(string)
					memoryUsageMetric[key] = nil
					temp = key
				}
				if k == "values" {
					vArr := v.([]interface{})
					if _, ok := memoryUsageMetric[temp]; ok {
						memoryUsageMetric[temp] = vArr
					}
				}
			}
		}
	}
	return memoryUsageMetric
}

//Deprecated: Currently not in use (intent to fetch graph data from prometheus)
func (impl AppListingServiceImpl) GraphAPI(appId int, envId int) error {
	impl.Logger.Debug("reached at GraphAPI:")
	/*
		appDetailView, err := impl.appListingRepository.FetchAppDetail(appId, envId)
		if err != nil {
			impl.Logger.Errorw("Exception", err)
			return err
		}

		//calculating cpu and memory usage percent
		appLabel := appDetailView.AppName
		namespace := appDetailView.Namespace
		proEndpoint := appDetailView.PrometheusEndpoint
		env := appDetailView.EnvironmentName
		podList := impl.PodListByAppLabel(appLabel, namespace, env, proEndpoint)

		//TODO - Pod List By Label- Release

		time1 := time.Time{}
		r1 := v1.Range{
			Start: time1.Add(-time.Minute),
			End:   time1,
			Step:  time.Minute,
		}
		podName := "prometheus-monitoring-prometheus-oper-prometheus-0"
		impl.CpuUsageGroupByPodGraph(podName, namespace, env, proEndpoint, r1)
		//data := impl.MemoryUsageGroupByPodGraph(podName, "monitoring", env, proEndpoint, r1)

		for fKey, _ := range podList {
			fmt.Println(fKey)
		}
	*/
	return nil
}

func (impl AppListingServiceImpl) FetchAppTriggerView(appId int) ([]bean.TriggerView, error) {
	return impl.appListingRepository.FetchAppTriggerView(appId)
}

func (impl AppListingServiceImpl) FetchAppStageStatus(appId int) ([]bean.AppStageStatus, error) {
	return impl.appListingRepository.FetchAppStageStatus(appId)
}

func (impl AppListingServiceImpl) FetchOtherEnvironment(appId int) ([]*bean.Environment, error) {
	envs, err := impl.appListingRepository.FetchOtherEnvironment(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return envs, err
	}
	appLevelAppMetrics := false  //default value
	appLevelInfraMetrics := true //default val
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in fetching app metrics", "err", err)
		return envs, err
	} else if util.IsErrNoRows(err) {
		//populate default val
		appLevelAppMetrics = false  //default value
		appLevelInfraMetrics = true //default val
	} else {
		appLevelAppMetrics = appLevelMetrics.AppMetrics
		appLevelInfraMetrics = appLevelMetrics.InfraMetrics
	}

	for _, env := range envs {
		if env.AppMetrics == nil {
			env.AppMetrics = &appLevelAppMetrics
		}
		if env.InfraMetrics == nil {
			env.InfraMetrics = &appLevelInfraMetrics
		}
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
