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

package generateManifest

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/gammazero/workerpool"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type DeploymentTemplateService interface {
	FetchDeploymentsWithChartRefs(appId int, envId int) ([]*repository.DeploymentTemplateComparisonMetadata, error)
	GetDeploymentTemplate(ctx context.Context, request DeploymentTemplateRequest) (DeploymentTemplateResponse, error)
	GenerateManifest(ctx context.Context, chartRefId int, valuesYaml string) (*openapi2.TemplateChartResponse, error)
	GetRestartWorkloadData(ctx context.Context, appIds []int, envId int) (*RestartPodResponse, error)
}
type DeploymentTemplateServiceImpl struct {
	Logger                           *zap.SugaredLogger
	chartService                     chart.ChartService
	appListingService                app.AppListingService
	deploymentTemplateRepository     repository.DeploymentTemplateRepository
	helmAppService                   client.HelmAppService
	chartTemplateServiceImpl         util.ChartTemplateService
	K8sUtil                          *k8s.K8sServiceImpl
	helmAppClient                    gRPC.HelmAppClient
	propertiesConfigService          pipeline.PropertiesConfigService
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
	environmentRepository            repository3.EnvironmentRepository
	appRepository                    appRepository.AppRepository
	scopedVariableManager            variables.ScopedVariableManager
	chartRefService                  chartRef.ChartRefService
	pipelineOverrideRepository       chartConfig.PipelineOverrideRepository
	chartRepository                  chartRepoRepository.ChartRepository
	restartWorkloadConfig            *RestartWorkloadConfig
}

func GetRestartWorkloadConfig() (*RestartWorkloadConfig, error) {
	cfg := &RestartWorkloadConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewDeploymentTemplateServiceImpl(Logger *zap.SugaredLogger, chartService chart.ChartService,
	appListingService app.AppListingService,
	deploymentTemplateRepository repository.DeploymentTemplateRepository,
	helmAppService client.HelmAppService,
	chartTemplateServiceImpl util.ChartTemplateService,
	helmAppClient gRPC.HelmAppClient,
	K8sUtil *k8s.K8sServiceImpl,
	propertiesConfigService pipeline.PropertiesConfigService,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService,
	environmentRepository repository3.EnvironmentRepository,
	appRepository appRepository.AppRepository,
	scopedVariableManager variables.ScopedVariableManager,
	chartRefService chartRef.ChartRefService,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	chartRepository chartRepoRepository.ChartRepository,
) (*DeploymentTemplateServiceImpl, error) {
	deploymentTemplateServiceImpl := &DeploymentTemplateServiceImpl{
		Logger:                           Logger,
		chartService:                     chartService,
		appListingService:                appListingService,
		deploymentTemplateRepository:     deploymentTemplateRepository,
		helmAppService:                   helmAppService,
		chartTemplateServiceImpl:         chartTemplateServiceImpl,
		K8sUtil:                          K8sUtil,
		helmAppClient:                    helmAppClient,
		propertiesConfigService:          propertiesConfigService,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		environmentRepository:            environmentRepository,
		appRepository:                    appRepository,
		scopedVariableManager:            scopedVariableManager,
		chartRefService:                  chartRefService,
		pipelineOverrideRepository:       pipelineOverrideRepository,
		chartRepository:                  chartRepository,
	}
	cfg, err := GetRestartWorkloadConfig()
	if err != nil {
		return nil, err
	}
	deploymentTemplateServiceImpl.restartWorkloadConfig = cfg
	return deploymentTemplateServiceImpl, nil
}

func (impl DeploymentTemplateServiceImpl) FetchDeploymentsWithChartRefs(appId int, envId int) ([]*repository.DeploymentTemplateComparisonMetadata, error) {

	var responseList []*repository.DeploymentTemplateComparisonMetadata

	defaultVersions, err := impl.chartService.ChartRefAutocompleteForAppOrEnv(appId, 0)
	if err != nil {
		impl.Logger.Errorw("error in getting defaultVersions", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, item := range defaultVersions.ChartRefs {
		res := &repository.DeploymentTemplateComparisonMetadata{
			ChartRefId:   item.Id,
			ChartVersion: item.Version,
			ChartType:    item.Name,
			Type:         repository.DefaultVersions,
		}
		responseList = append(responseList, res)
	}

	publishedOnEnvs, err := impl.appListingService.FetchMinDetailOtherEnvironment(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting publishedOnEnvs", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, env := range publishedOnEnvs {
		item := &repository.DeploymentTemplateComparisonMetadata{
			ChartRefId:      env.ChartRefId,
			EnvironmentId:   env.EnvironmentId,
			EnvironmentName: env.EnvironmentName,
			Type:            repository.PublishedOnEnvironments,
		}
		responseList = append(responseList, item)
	}

	deployedOnEnv, err := impl.deploymentTemplateRepository.FetchDeploymentHistoryWithChartRefs(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting deployedOnEnv", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, deployedItem := range deployedOnEnv {
		deployedItem.Type = repository.DeployedOnSelfEnvironment
		deployedItem.EnvironmentId = envId
		responseList = append(responseList, deployedItem)
	}

	deployedOnOtherEnvs, err := impl.deploymentTemplateRepository.FetchLatestDeploymentWithChartRefs(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting deployedOnOtherEnvs", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, deployedItem := range deployedOnOtherEnvs {
		deployedItem.Type = repository.DeployedOnOtherEnvironment
		responseList = append(responseList, deployedItem)
	}

	return responseList, nil
}

func (impl DeploymentTemplateServiceImpl) GetDeploymentTemplate(ctx context.Context, request DeploymentTemplateRequest) (DeploymentTemplateResponse, error) {
	var result DeploymentTemplateResponse
	var values, resolvedValue string
	var err error
	var variableSnapshot map[string]string

	if request.Values != "" {
		values = request.Values
		resolvedValue, variableSnapshot, err = impl.resolveTemplateVariables(ctx, request.Values, request)
		if err != nil {
			return result, err
		}
	} else {
		switch request.Type {
		case repository.DefaultVersions:
			_, values, err = impl.chartRefService.GetAppOverrideForDefaultTemplate(request.ChartRefId)
			resolvedValue = values
		case repository.PublishedOnEnvironments:
			values, resolvedValue, variableSnapshot, err = impl.fetchResolvedTemplateForPublishedEnvs(ctx, request)
		case repository.DeployedOnSelfEnvironment, repository.DeployedOnOtherEnvironment:
			values, resolvedValue, variableSnapshot, err = impl.fetchTemplateForDeployedEnv(ctx, request)
		}
		if err != nil {
			impl.Logger.Errorw("error in getting values", "err", err)
			return result, err
		}
	}

	if request.RequestDataMode == Values {
		result.Data = values
		result.ResolvedData = resolvedValue
		result.VariableSnapshot = variableSnapshot
		return result, nil
	}

	manifest, err := impl.GenerateManifest(ctx, request.ChartRefId, resolvedValue)
	if err != nil {
		return result, err
	}
	result.Data = *manifest.Manifest
	return result, nil
}

func (impl DeploymentTemplateServiceImpl) fetchResolvedTemplateForPublishedEnvs(ctx context.Context, request DeploymentTemplateRequest) (string, string, map[string]string, error) {
	var values string
	override, err := impl.propertiesConfigService.GetEnvironmentProperties(request.AppId, request.EnvId, request.ChartRefId)
	if err == nil && override.GlobalConfig != nil {
		if override.EnvironmentConfig.EnvOverrideValues != nil {
			values = string(override.EnvironmentConfig.EnvOverrideValues)
		} else {
			values = string(override.GlobalConfig)
		}
	} else {
		impl.Logger.Errorw("error in getting overridden values", "err", err)
		return "", "", nil, err
	}
	resolvedTemplate, variableSnapshot, err := impl.resolveTemplateVariables(ctx, values, request)
	if err != nil {
		return values, values, variableSnapshot, err
	}
	return values, resolvedTemplate, variableSnapshot, nil
}

func (impl DeploymentTemplateServiceImpl) fetchTemplateForDeployedEnv(ctx context.Context, request DeploymentTemplateRequest) (string, string, map[string]string, error) {
	historyObject, err := impl.deploymentTemplateHistoryService.GetHistoryForDeployedTemplateById(ctx, request.DeploymentTemplateHistoryId, request.PipelineId)
	if err != nil {
		impl.Logger.Errorw("error in getting deployment template history", "err", err, "id", request.DeploymentTemplateHistoryId, "pipelineId", request.PipelineId)
		return "", "", nil, err
	}

	//todo Subhashish solve variable leak
	return historyObject.CodeEditorValue.Value, historyObject.CodeEditorValue.ResolvedValue, historyObject.CodeEditorValue.VariableSnapshot, nil
}

func (impl DeploymentTemplateServiceImpl) resolveTemplateVariables(ctx context.Context, values string, request DeploymentTemplateRequest) (string, map[string]string, error) {

	isSuperAdmin, err := util2.GetIsSuperAdminFromContext(ctx)
	if err != nil {
		return values, nil, err
	}
	scope, err := impl.extractScopeData(request)
	if err != nil {
		return values, nil, err
	}
	maskUnknownVariableForHelmGenerate := request.RequestDataMode == Manifest
	resolvedTemplate, variableSnapshot, err := impl.scopedVariableManager.ExtractVariablesAndResolveTemplate(scope, values, parsers.StringVariableTemplate, isSuperAdmin, maskUnknownVariableForHelmGenerate)
	if err != nil {
		return values, variableSnapshot, err
	}
	return resolvedTemplate, variableSnapshot, nil
}

func (impl DeploymentTemplateServiceImpl) extractScopeData(request DeploymentTemplateRequest) (resourceQualifiers.Scope, error) {
	app, err := impl.appRepository.FindById(request.AppId)
	scope := resourceQualifiers.Scope{}
	if err != nil {
		return scope, err
	}
	scope.AppId = request.AppId
	scope.EnvId = request.EnvId
	scope.SystemMetadata = &resourceQualifiers.SystemMetadata{AppName: app.AppName}

	if request.EnvId != 0 {
		environment, err := impl.environmentRepository.FindById(request.EnvId)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in getting system metadata", "err", err)
			return scope, err
		}
		if environment != nil {
			scope.ClusterId = environment.ClusterId
			scope.SystemMetadata.EnvironmentName = environment.Name
			scope.SystemMetadata.ClusterName = environment.Cluster.ClusterName
			scope.SystemMetadata.Namespace = environment.Namespace
		}
	}
	return scope, nil
}

func (impl DeploymentTemplateServiceImpl) GenerateManifest(ctx context.Context, chartRefId int, valuesYaml string) (*openapi2.TemplateChartResponse, error) {
	refChart, template, version, _, err := impl.chartRefService.GetRefChart(chartRefId)
	if err != nil {
		impl.Logger.Errorw("error in getting refChart", "err", err, "chartRefId", chartRefId)
		return nil, err
	}

	outputChartPathDir := fmt.Sprintf("%s-%v", refChart, strconv.FormatInt(time.Now().UnixNano(), 16))
	if _, err := os.Stat(outputChartPathDir); os.IsNotExist(err) {
		err = os.Mkdir(outputChartPathDir, 0755)
		if err != nil {
			impl.Logger.Errorw("error in creating temp outputChartPathDir", "err", err, "outputChartPathDir", outputChartPathDir, "chartRefId", chartRefId)
			return nil, err
		}
	}
	//load chart from given refChart
	chart, err := impl.chartTemplateServiceImpl.LoadChartFromDir(refChart)
	if err != nil {
		impl.Logger.Errorw("error in LoadChartFromDir", "err", err, "chartRefId", chartRefId)
		return nil, err
	}

	//create the .tgz file in temp location
	chartBytes, err := impl.chartTemplateServiceImpl.CreateZipFileForChart(chart, outputChartPathDir)
	if err != nil {
		impl.Logger.Errorw("error in CreateZipFileForChart", "err", err, "chartRefId", chartRefId)
		return nil, err
	}

	//deleted the .tgz temp file after reading chart bytes
	defer impl.chartTemplateServiceImpl.CleanDir(outputChartPathDir)

	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.Logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}

	sanitizedK8sVersion := k8sServerVersion.String()
	//handle specific case for all cronjob charts from cronjob-chart_1-2-0 to cronjob-chart_1-5-0 where semverCompare
	//comparison func has wrong api version mentioned, so for already installed charts via these charts that comparison
	//is always false, handles the gh issue:- https://github.com/devtron-labs/devtron/issues/4860
	cronJobChartRegex := regexp.MustCompile(bean2.CronJobChartRegexExpression)
	if cronJobChartRegex.MatchString(template) {
		sanitizedK8sVersion = k8s2.StripPrereleaseFromK8sVersion(sanitizedK8sVersion)
	}

	installReleaseRequest := &gRPC.InstallReleaseRequest{
		ChartName:         template,
		ChartVersion:      version,
		ValuesYaml:        valuesYaml,
		K8SVersion:        sanitizedK8sVersion,
		ChartRepository:   ChartRepository,
		ReleaseIdentifier: ReleaseIdentifier,
		ChartContent: &gRPC.ChartContent{
			Content: chartBytes,
		},
	}
	config, err := impl.helmAppService.GetClusterConf(bean.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}

	installReleaseRequest.ReleaseIdentifier.ClusterConfig = config

	templateChartResponse, err := impl.helmAppClient.TemplateChart(ctx, installReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in templating chart", "err", err)
		clientErrCode, errMsg := util.GetClientDetailedError(err)
		if clientErrCode.IsFailedPreconditionCode() {
			return nil, &util.ApiError{HttpStatusCode: http.StatusUnprocessableEntity, Code: strconv.Itoa(http.StatusUnprocessableEntity), InternalMessage: errMsg, UserMessage: errMsg}
		} else if clientErrCode.IsInvalidArgumentCode() {
			return nil, &util.ApiError{HttpStatusCode: http.StatusConflict, Code: strconv.Itoa(http.StatusConflict), InternalMessage: errMsg, UserMessage: errMsg}
		}
		return nil, err
	}
	response := &openapi2.TemplateChartResponse{
		Manifest: &templateChartResponse.GeneratedManifest,
	}

	return response, nil
}
func (impl DeploymentTemplateServiceImpl) GetRestartWorkloadData(ctx context.Context, appIds []int, envId int) (*RestartPodResponse, error) {
	wp := workerpool.New(impl.restartWorkloadConfig.WorkerPoolSize)
	var templateChartResponse []*gRPC.TemplateChartResponse
	templateChartResponseLock := &sync.Mutex{}
	podResp := &RestartPodResponse{}
	appNameToId := make(map[string]int)
	if len(appIds) == 0 {
		return podResp, nil
	}
	apps, err := impl.appRepository.FindByIds(util2.GetReferencedArray(appIds))
	if err != nil {
		impl.Logger.Errorw("error in fetching app", "appIds", appIds, "err", err)
		return nil, err
	}
	environment, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching environment", "envId", envId, "err", err)
		return nil, err
	}
	installReleaseRequests, err := impl.constructInstallReleaseBulkReq(apps, environment)
	if err != nil {
		impl.Logger.Errorw("error in fetching installReleaseRequests", "appIds", appIds, "envId", envId, "err", err)
		return nil, err
	}
	for _, app := range apps {
		appNameToId[app.AppName] = app.Id
	}
	partitionedRequests := utils.PartitionSlice(installReleaseRequests, impl.restartWorkloadConfig.RequestBatchSize)
	var finalError error
	for i, _ := range partitionedRequests {
		req := partitionedRequests[i]
		err = impl.setChartContent(ctx, req, appNameToId)
		if err != nil {
			impl.Logger.Errorw("error in setting chart content for apps", "appNames", maps.Keys(appNameToId), "err", err)
			// continue processing next batch
			continue
		}
		wp.Submit(func() {
			resp, err := impl.helmAppClient.TemplateChartBulk(ctx, &gRPC.BulkInstallReleaseRequest{BulkInstallReleaseRequest: req})
			if err != nil {
				impl.Logger.Errorw("error in getting data from template chart", "err", err)
				finalError = err
				return
			}
			templateChartResponseLock.Lock()
			templateChartResponse = append(templateChartResponse, resp.BulkTemplateChartResponse...)
			templateChartResponseLock.Unlock()

		})
	}
	wp.StopWait()
	if finalError != nil {
		impl.Logger.Errorw("error in fetching response", "installReleaseRequests", installReleaseRequests, "templateChartResponse", templateChartResponse)
		return nil, finalError
	}
	impl.Logger.Infow("fetching template chart resp", "templateChartResponse", templateChartResponse, "err", err)

	podResp, err = impl.constructRotatePodResponse(templateChartResponse, appNameToId, environment)
	if err != nil {
		impl.Logger.Errorw("error in constructing pod resp", "templateChartResponse", templateChartResponse, "appNameToId", appNameToId, "environment", environment, "err", err)
		return nil, err
	}
	return podResp, nil
}
