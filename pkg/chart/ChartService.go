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

package chart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository4 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	chartRefBean "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	variablesRepository "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util/ChartsUtil"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"
)

type ChartService interface {
	Create(templateRequest TemplateRequest, ctx context.Context) (chart *TemplateRequest, err error)
	CreateChartFromEnvOverride(templateRequest TemplateRequest, ctx context.Context) (chart *TemplateRequest, err error)
	FindLatestChartForAppByAppId(appId int) (chartTemplate *TemplateRequest, err error)
	GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *TemplateRequest, err error)
	UpdateAppOverride(ctx context.Context, templateRequest *TemplateRequest) (*TemplateRequest, error)
	IsReadyToTrigger(appId int, envId int, pipelineId int) (IsReady, error)
	FindPreviousChartByAppId(appId int) (chartTemplate *TemplateRequest, err error)
	UpgradeForApp(appId int, chartRefId int, newAppOverride map[string]interface{}, userId int32, ctx context.Context) (bool, error)
	CheckIfChartRefUserUploadedByAppId(id int) (bool, error)
	PatchEnvOverrides(values json.RawMessage, oldChartType string, newChartType string) (json.RawMessage, error)

	ChartRefAutocompleteForAppOrEnv(appId int, envId int) (*chartRefBean.ChartRefAutocompleteResponse, error)

	UpdateGitRepoUrlInCharts(appId int, repoUrl, chartLocation string, userId int32) error

	IsGitOpsRepoConfiguredForDevtronApps(appId int) (bool, error)
	IsGitOpsRepoAlreadyRegistered(gitOpsRepoUrl string) (bool, error)
}

type ChartServiceImpl struct {
	chartRepository                  chartRepoRepository.ChartRepository
	logger                           *zap.SugaredLogger
	repoRepository                   chartRepoRepository.ChartRepoRepository
	chartTemplateService             util.ChartTemplateService
	pipelineGroupRepository          app.AppRepository
	mergeUtil                        util.MergeUtil
	envOverrideRepository            chartConfig.EnvConfigOverrideRepository
	pipelineConfigRepository         chartConfig.PipelineConfigRepository
	environmentRepository            repository4.EnvironmentRepository
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
	scopedVariableManager            variables.ScopedVariableManager
	deployedAppMetricsService        deployedAppMetrics.DeployedAppMetricsService
	chartRefService                  chartRef.ChartRefService
	gitOpsConfigReadService          config.GitOpsConfigReadService
}

func NewChartServiceImpl(chartRepository chartRepoRepository.ChartRepository,
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	repoRepository chartRepoRepository.ChartRepoRepository,
	pipelineGroupRepository app.AppRepository,
	mergeUtil util.MergeUtil,
	envOverrideRepository chartConfig.EnvConfigOverrideRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	environmentRepository repository4.EnvironmentRepository,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService,
	scopedVariableManager variables.ScopedVariableManager,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	chartRefService chartRef.ChartRefService,
	gitOpsConfigReadService config.GitOpsConfigReadService) *ChartServiceImpl {
	return &ChartServiceImpl{
		chartRepository:                  chartRepository,
		logger:                           logger,
		chartTemplateService:             chartTemplateService,
		repoRepository:                   repoRepository,
		pipelineGroupRepository:          pipelineGroupRepository,
		mergeUtil:                        mergeUtil,
		envOverrideRepository:            envOverrideRepository,
		pipelineConfigRepository:         pipelineConfigRepository,
		environmentRepository:            environmentRepository,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		scopedVariableManager:            scopedVariableManager,
		deployedAppMetricsService:        deployedAppMetricsService,
		chartRefService:                  chartRefService,
		gitOpsConfigReadService:          gitOpsConfigReadService,
	}
}

func (impl *ChartServiceImpl) PatchEnvOverrides(values json.RawMessage, oldChartType string, newChartType string) (json.RawMessage, error) {
	return PatchWinterSoldierConfig(values, newChartType)
}

func (impl *ChartServiceImpl) Create(templateRequest TemplateRequest, ctx context.Context) (*TemplateRequest, error) {
	err := impl.chartRefService.CheckChartExists(templateRequest.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting missing chart for chartRefId", "err", err, "chartRefId")
		return nil, err
	}
	chartMeta, err := impl.getChartMetaData(templateRequest)
	if err != nil {
		return nil, err
	}

	//save chart
	// 1. create chart, 2. push in repo, 3. add value of chart variable 4. save chart
	charRepository, err := impl.getChartRepo(templateRequest)
	if err != nil {
		impl.logger.Errorw("error in fetching chart repo detail", "req", templateRequest)
		return nil, err
	}

	refChart, templateName, _, pipelineStrategyPath, err := impl.chartRefService.GetRefChart(templateRequest.ChartRefId)
	if err != nil {
		return nil, err
	}

	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	existingChart, _ := impl.chartRepository.FindChartByAppIdAndRefId(templateRequest.AppId, templateRequest.ChartRefId)
	if existingChart != nil && existingChart.Id > 0 {
		return nil, fmt.Errorf("this reference chart already has added to appId %d refId %d", templateRequest.AppId, templateRequest.Id)
	}

	// STARTS
	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	if err != nil && pg.ErrNoRows != err {
		return nil, err
	}
	gitRepoUrl := apiBean.GIT_REPO_NOT_CONFIGURED
	impl.logger.Debugw("current latest chart in db", "chartId", currentLatestChart.Id)
	if currentLatestChart.Id > 0 {
		impl.logger.Debugw("updating env and pipeline config which are currently latest in db", "chartId", currentLatestChart.Id)

		impl.logger.Debug("updating all other charts which are not latest but may be set previous true, setting previous=false")
		//step 2
		noLatestCharts, err := impl.chartRepository.FindNoLatestChartForAppByAppId(templateRequest.AppId)
		for _, noLatestChart := range noLatestCharts {
			if noLatestChart.Id != templateRequest.Id {

				noLatestChart.Latest = false // these are already false by d way
				noLatestChart.Previous = false
				err = impl.chartRepository.Update(noLatestChart)
				if err != nil {
					return nil, err
				}
			}
		}

		impl.logger.Debug("now going to update latest entry in db to false and previous flag = true")
		// now finally update latest entry in db to false and previous true
		currentLatestChart.Latest = false // these are already false by d way
		currentLatestChart.Previous = true
		err = impl.chartRepository.Update(currentLatestChart)
		if err != nil {
			return nil, err
		}
		if currentLatestChart.GitRepoUrl != "" {
			gitRepoUrl = currentLatestChart.GitRepoUrl
		}
	}
	// ENDS

	impl.logger.Debug("now finally create new chart and make it latest entry in db and previous flag = true")

	version, err := impl.getNewVersion(charRepository.Name, chartMeta.Name, refChart)
	chartMeta.Version = version
	if err != nil {
		return nil, err
	}
	chartValues, err := impl.chartTemplateService.FetchValuesFromReferenceChart(chartMeta, refChart, templateName, templateRequest.UserId, pipelineStrategyPath)
	if err != nil {
		return nil, err
	}
	chartLocation := filepath.Join(templateName, version)
	override, err := templateRequest.ValuesOverride.MarshalJSON()
	if err != nil {
		return nil, err
	}
	valuesJson, err := yaml.YAMLToJSON([]byte(chartValues.Values))
	if err != nil {
		return nil, err
	}
	merged, err := impl.mergeUtil.JsonPatch(valuesJson, []byte(templateRequest.ValuesOverride))
	if err != nil {
		return nil, err
	}

	dst := new(bytes.Buffer)
	err = json.Compact(dst, override)
	if err != nil {
		return nil, err
	}
	override = dst.Bytes()
	chart := &chartRepoRepository.Chart{
		AppId:                   templateRequest.AppId,
		ChartRepoId:             charRepository.Id,
		Values:                  string(merged),
		GlobalOverride:          string(override),
		ReleaseOverride:         chartValues.ReleaseOverrides, //image descriptor template
		PipelineOverride:        chartValues.PipelineOverrides,
		ImageDescriptorTemplate: chartValues.ImageDescriptorTemplate,
		ChartName:               chartMeta.Name,
		ChartRepo:               charRepository.Name,
		ChartRepoUrl:            charRepository.Url,
		ChartVersion:            chartMeta.Version,
		Status:                  models.CHARTSTATUS_NEW,
		Active:                  true,
		ChartLocation:           chartLocation,
		GitRepoUrl:              gitRepoUrl,
		ReferenceTemplate:       templateName,
		ChartRefId:              templateRequest.ChartRefId,
		Latest:                  true,
		Previous:                false,
		IsBasicViewLocked:       templateRequest.IsBasicViewLocked,
		CurrentViewEditor:       templateRequest.CurrentViewEditor,
		AuditLog:                sql.AuditLog{CreatedBy: templateRequest.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: templateRequest.UserId},
	}

	err = impl.chartRepository.Save(chart)
	if err != nil {
		impl.logger.Errorw("error in saving chart ", "chart", chart, "error", err)
		//If found any error, rollback chart museum
		return nil, err
	}

	//creating history entry for deployment template
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(chart, nil, templateRequest.IsAppMetricsEnabled)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", chart)
		return nil, err
	}

	//VARIABLE_MAPPING_UPDATE
	err = impl.scopedVariableManager.ExtractAndMapVariables(chart.GlobalOverride, chart.Id, variablesRepository.EntityTypeDeploymentTemplateAppLevel, chart.CreatedBy, nil)
	if err != nil {
		return nil, err
	}

	appLevelMetricsUpdateReq := &bean.DeployedAppMetricsRequest{
		EnableMetrics: templateRequest.IsAppMetricsEnabled,
		AppId:         templateRequest.AppId,
		ChartRefId:    templateRequest.ChartRefId,
		UserId:        templateRequest.UserId,
	}
	err = impl.deployedAppMetricsService.CreateOrUpdateAppOrEnvLevelMetrics(ctx, appLevelMetricsUpdateReq)
	if err != nil {
		impl.logger.Errorw("error, CheckAndUpdateAppOrEnvLevelMetrics", "err", err, "req", appLevelMetricsUpdateReq)
		return nil, err
	}

	chartVal, err := impl.chartAdaptor(chart, appLevelMetricsUpdateReq.EnableMetrics)
	return chartVal, err
}

func (impl *ChartServiceImpl) CreateChartFromEnvOverride(templateRequest TemplateRequest, ctx context.Context) (*TemplateRequest, error) {
	err := impl.chartRefService.CheckChartExists(templateRequest.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting missing chart for chartRefId", "err", err, "chartRefId")
		return nil, err
	}

	chartMeta, err := impl.getChartMetaData(templateRequest)
	if err != nil {
		return nil, err
	}

	appMetrics := templateRequest.IsAppMetricsEnabled

	//save chart
	// 1. create chart, 2. push in repo, 3. add value of chart variable 4. save chart
	chartRepository, err := impl.getChartRepo(templateRequest)
	if err != nil {
		impl.logger.Errorw("error in fetching chart repo detail", "req", templateRequest, "err", err)
		return nil, err
	}

	refChart, templateName, _, pipelineStrategyPath, err := impl.chartRefService.GetRefChart(templateRequest.ChartRefId)
	if err != nil {
		return nil, err
	}

	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	impl.logger.Debug("now finally create new chart and make it latest entry in db and previous flag = true")
	version, err := impl.getNewVersion(chartRepository.Name, chartMeta.Name, refChart)
	chartMeta.Version = version
	if err != nil {
		return nil, err
	}
	chartValues, err := impl.chartTemplateService.FetchValuesFromReferenceChart(chartMeta, refChart, templateName, templateRequest.UserId, pipelineStrategyPath)
	if err != nil {
		return nil, err
	}
	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	if err != nil && pg.ErrNoRows != err {
		return nil, err
	}
	chartLocation := filepath.Join(templateName, version)
	gitRepoUrl := apiBean.GIT_REPO_NOT_CONFIGURED
	if currentLatestChart.Id > 0 && currentLatestChart.GitRepoUrl != "" {
		gitRepoUrl = currentLatestChart.GitRepoUrl
	}
	override, err := templateRequest.ValuesOverride.MarshalJSON()
	if err != nil {
		return nil, err
	}
	valuesJson, err := yaml.YAMLToJSON([]byte(chartValues.Values))
	if err != nil {
		return nil, err
	}
	merged, err := impl.mergeUtil.JsonPatch(valuesJson, []byte(templateRequest.ValuesOverride))
	if err != nil {
		return nil, err
	}

	dst := new(bytes.Buffer)
	err = json.Compact(dst, override)
	if err != nil {
		return nil, err
	}
	override = dst.Bytes()
	chart := &chartRepoRepository.Chart{
		AppId:                   templateRequest.AppId,
		ChartRepoId:             chartRepository.Id,
		Values:                  string(merged),
		GlobalOverride:          string(override),
		ReleaseOverride:         chartValues.ReleaseOverrides,
		PipelineOverride:        chartValues.PipelineOverrides,
		ImageDescriptorTemplate: chartValues.ImageDescriptorTemplate,
		ChartName:               chartMeta.Name,
		ChartRepo:               chartRepository.Name,
		ChartRepoUrl:            chartRepository.Url,
		ChartVersion:            chartMeta.Version,
		Status:                  models.CHARTSTATUS_NEW,
		Active:                  true,
		ChartLocation:           chartLocation,
		GitRepoUrl:              gitRepoUrl,
		ReferenceTemplate:       templateName,
		ChartRefId:              templateRequest.ChartRefId,
		Latest:                  false,
		Previous:                false,
		IsBasicViewLocked:       templateRequest.IsBasicViewLocked,
		CurrentViewEditor:       templateRequest.CurrentViewEditor,
		AuditLog:                sql.AuditLog{CreatedBy: templateRequest.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: templateRequest.UserId},
	}

	err = impl.chartRepository.Save(chart)
	if err != nil {
		impl.logger.Errorw("error in saving chart ", "chart", chart, "error", err)
		return nil, err
	}
	//creating history entry for deployment template
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(chart, nil, appMetrics)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", chart)
		return nil, err
	}
	//VARIABLE_MAPPING_UPDATE
	err = impl.scopedVariableManager.ExtractAndMapVariables(chart.GlobalOverride, chart.Id, variablesRepository.EntityTypeDeploymentTemplateAppLevel, chart.CreatedBy, nil)
	if err != nil {
		return nil, err
	}

	chartVal, err := impl.chartAdaptor(chart, false)
	return chartVal, err
}

// converts db object to bean
func (impl *ChartServiceImpl) chartAdaptor(chart *chartRepoRepository.Chart, isAppMetricsEnabled bool) (*TemplateRequest, error) {
	if chart == nil || chart.Id == 0 {
		return &TemplateRequest{}, &util.ApiError{UserMessage: "no chart found"}
	}
	gitRepoUrl := ""
	if !ChartsUtil.IsGitOpsRepoNotConfigured(chart.GitRepoUrl) {
		gitRepoUrl = chart.GitRepoUrl
	}
	return &TemplateRequest{
		RefChartTemplate:        chart.ReferenceTemplate,
		Id:                      chart.Id,
		AppId:                   chart.AppId,
		ChartRepositoryId:       chart.ChartRepoId,
		DefaultAppOverride:      json.RawMessage(chart.GlobalOverride),
		RefChartTemplateVersion: impl.getParentChartVersion(chart.ChartVersion),
		Latest:                  chart.Latest,
		ChartRefId:              chart.ChartRefId,
		IsAppMetricsEnabled:     isAppMetricsEnabled,
		IsBasicViewLocked:       chart.IsBasicViewLocked,
		CurrentViewEditor:       chart.CurrentViewEditor,
		GitRepoUrl:              gitRepoUrl,
		IsCustomGitRepository:   chart.IsCustomGitRepository,
	}, nil
}

func (impl *ChartServiceImpl) getChartMetaData(templateRequest TemplateRequest) (*chart.Metadata, error) {
	pg, err := impl.pipelineGroupRepository.FindById(templateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching pg", "id", templateRequest.AppId, "err", err)
	}
	metadata := &chart.Metadata{
		Name: pg.AppName,
	}
	return metadata, err
}

func (impl *ChartServiceImpl) getChartRepo(templateRequest TemplateRequest) (*chartRepoRepository.ChartRepo, error) {
	if templateRequest.ChartRepositoryId == 0 {
		chartRepo, err := impl.repoRepository.GetDefault()
		if err != nil {
			impl.logger.Errorw("error in fetching default repo", "err", err)
			return nil, err
		}
		return chartRepo, err
	} else {
		chartRepo, err := impl.repoRepository.FindById(templateRequest.ChartRepositoryId)
		if err != nil {
			impl.logger.Errorw("error in fetching chart repo", "err", err, "id", templateRequest.ChartRepositoryId)
			return nil, err
		}
		return chartRepo, err
	}
}

func (impl *ChartServiceImpl) getParentChartVersion(childVersion string) string {
	placeholders := strings.Split(childVersion, ".")
	return fmt.Sprintf("%s.%s.0", placeholders[0], placeholders[1])
}

// this method is not thread safe
func (impl *ChartServiceImpl) getNewVersion(chartRepo, chartName, refChartLocation string) (string, error) {
	parentVersion, err := impl.chartTemplateService.GetChartVersion(refChartLocation)
	if err != nil {
		return "", err
	}
	placeholders := strings.Split(parentVersion, ".")
	if len(placeholders) != 3 {
		return "", fmt.Errorf("invalid parent chart version %s", parentVersion)
	}

	currentVersion, err := impl.chartRepository.FindCurrentChartVersion(chartRepo, chartName, placeholders[0]+"."+placeholders[1])
	if err != nil {
		return placeholders[0] + "." + placeholders[1] + ".1", nil
	}
	patch := strings.Split(currentVersion, ".")[2]
	count, err := strconv.ParseInt(patch, 10, 32)
	if err != nil {
		return "", err
	}
	count += 1

	return placeholders[0] + "." + placeholders[1] + "." + strconv.FormatInt(count, 10), nil
}

func (impl *ChartServiceImpl) IsGitOpsRepoConfiguredForDevtronApps(appId int) (bool, error) {
	activeGitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsConfigActive()
	if util.IsErrNoRows(err) {
		return false, nil
	}
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId")
		return false, err
	}
	if !activeGitOpsConfig.AllowCustomRepository {
		return true, nil
	}
	latestChartConfiguredInApp, err := impl.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId")
		return false, err
	}
	if ChartsUtil.IsGitOpsRepoNotConfigured(latestChartConfiguredInApp.GitRepoUrl) {
		return false, nil
	}
	return true, nil
}

func (impl *ChartServiceImpl) FindLatestChartForAppByAppId(appId int) (chartTemplate *TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}
	isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app-metrics", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = impl.chartAdaptor(chart, isAppMetricsEnabled)
	return chartTemplate, err
}

func (impl *ChartServiceImpl) GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindChartByAppIdAndRefId(appId, chartRefId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}
	isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app-metrics", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = impl.chartAdaptor(chart, isAppMetricsEnabled)
	return chartTemplate, err
}

func (impl *ChartServiceImpl) UpdateAppOverride(ctx context.Context, templateRequest *TemplateRequest) (*TemplateRequest, error) {

	_, span := otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindById")
	template, err := impl.chartRepository.FindById(templateRequest.Id)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching chart config", "id", templateRequest.Id, "err", err)
		return nil, err
	}

	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	//STARTS
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	span.End()
	if err != nil {
		return nil, err
	}
	if currentLatestChart.Id > 0 && currentLatestChart.Id == templateRequest.Id {

	} else if currentLatestChart.Id != templateRequest.Id {
		impl.logger.Debug("updating env and pipeline config which are currently latest in db", "chartId", currentLatestChart.Id)

		impl.logger.Debugw("updating request chart env config and pipeline config - making configs latest", "chartId", templateRequest.Id)

		impl.logger.Debug("updating all other charts which are not latest but may be set previous true, setting previous=false")
		//step 3
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindNoLatestChartForAppByAppId")
		noLatestCharts, err := impl.chartRepository.FindNoLatestChartForAppByAppId(templateRequest.AppId)
		span.End()
		for _, noLatestChart := range noLatestCharts {
			if noLatestChart.Id != templateRequest.Id {

				noLatestChart.Latest = false // these are already false by d way
				noLatestChart.Previous = false
				_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.Update")
				err = impl.chartRepository.Update(noLatestChart)
				span.End()
				if err != nil {
					return nil, err
				}
			}
		}

		impl.logger.Debug("now going to update latest entry in db to false and previous flag = true")
		// now finally update latest entry in db to false and previous true
		currentLatestChart.Latest = false // these are already false by d way
		currentLatestChart.Previous = true
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.Update.LatestChart")
		err = impl.chartRepository.Update(currentLatestChart)
		span.End()
		if err != nil {
			return nil, err
		}

	} else {
		return nil, nil
	}
	//ENDS

	impl.logger.Debug("now finally update request chart in db to latest and previous flag = false")
	values, err := impl.mergeUtil.JsonPatch([]byte(template.Values), templateRequest.ValuesOverride)
	if err != nil {
		return nil, err
	}
	template.Values = string(values)
	template.UpdatedOn = time.Now()
	template.UpdatedBy = templateRequest.UserId
	template.GlobalOverride = string(templateRequest.ValuesOverride)
	template.Latest = true
	template.Previous = false
	template.IsBasicViewLocked = templateRequest.IsBasicViewLocked
	template.CurrentViewEditor = templateRequest.CurrentViewEditor
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.Update.requestTemplate")
	err = impl.chartRepository.Update(template)
	span.End()
	if err != nil {
		return nil, err
	}

	appLevelMetricsUpdateReq := &bean.DeployedAppMetricsRequest{
		EnableMetrics: templateRequest.IsAppMetricsEnabled,
		AppId:         templateRequest.AppId,
		ChartRefId:    templateRequest.ChartRefId,
		UserId:        templateRequest.UserId,
	}
	err = impl.deployedAppMetricsService.CreateOrUpdateAppOrEnvLevelMetrics(ctx, appLevelMetricsUpdateReq)
	if err != nil {
		impl.logger.Errorw("error, CheckAndUpdateAppOrEnvLevelMetrics", "err", err, "req", appLevelMetricsUpdateReq)
		return nil, err
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "CreateDeploymentTemplateHistoryFromGlobalTemplate")
	//creating history entry for deployment template
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(template, nil, appLevelMetricsUpdateReq.EnableMetrics)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", template)
		return nil, err
	}

	//VARIABLE_MAPPING_UPDATE
	err = impl.scopedVariableManager.ExtractAndMapVariables(template.GlobalOverride, template.Id, variablesRepository.EntityTypeDeploymentTemplateAppLevel, template.CreatedBy, nil)
	if err != nil {
		return nil, err
	}
	return templateRequest, nil
}

type IsReady struct {
	Flag    bool   `json:"flag"`
	Message string `json:"message"`
}

func (impl *ChartServiceImpl) IsReadyToTrigger(appId int, envId int, pipelineId int) (IsReady, error) {
	isReady := IsReady{Flag: false}
	envOverride, err := impl.envOverrideRepository.ActiveEnvConfigOverride(appId, envId)
	if err != nil {
		impl.logger.Errorf("invalid state", "err", err, "envId", envId)
		isReady.Message = "Something went wrong"
		return isReady, err
	}

	if envOverride.Latest == false {
		impl.logger.Error("chart is updated for this app, may be environment or pipeline config is older")
		isReady.Message = "chart is updated for this app, may be environment or pipeline config is older"
		return isReady, nil
	}

	strategy, err := impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(pipelineId)
	if err != nil {
		impl.logger.Errorw("invalid state", "err", err, "req", strategy)
		if errors.IsNotFound(err) {
			isReady.Message = "no strategy found for request pipeline in this environment"
			return isReady, fmt.Errorf("no pipeline config found for request pipeline in this environment")
		}
		isReady.Message = "Something went wrong"
		return isReady, err
	}

	isReady.Flag = true
	isReady.Message = "Pipeline is well enough configured for trigger"
	return isReady, nil
}

func (impl *ChartServiceImpl) ChartRefAutocompleteForAppOrEnv(appId int, envId int) (*chartRefBean.ChartRefAutocompleteResponse, error) {
	chartRefResponse := &chartRefBean.ChartRefAutocompleteResponse{}
	var chartRefs []chartRefBean.ChartRefAutocompleteDto

	results, err := impl.chartRefService.GetAll()
	if err != nil {
		impl.logger.Errorw("error in fetching chart ref", "err", err)
		return chartRefResponse, err
	}

	resultsMetadataMap, err := impl.chartRefService.GetAllChartMetadata()
	if err != nil {
		impl.logger.Errorw("error in fetching chart metadata", "err", err)
		return chartRefResponse, err
	}
	chartRefResponse.ChartsMetadata = resultsMetadataMap
	var LatestAppChartRef int
	for _, result := range results {
		chartRefs = append(chartRefs, chartRefBean.ChartRefAutocompleteDto{
			Id:                    result.Id,
			Version:               result.Version,
			Name:                  result.Name,
			Description:           result.ChartDescription,
			UserUploaded:          result.UserUploaded,
			IsAppMetricsSupported: result.IsAppMetricsSupported,
		})
		if result.Default == true {
			LatestAppChartRef = result.Id
		}
	}

	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching latest chart", "err", err)
		return chartRefResponse, err
	}

	if envId > 0 {
		envOverride, err := impl.envOverrideRepository.FindLatestChartForAppByAppIdAndEnvId(appId, envId)
		if err != nil && !errors.IsNotFound(err) {
			impl.logger.Errorw("error in fetching latest chart", "err", err)
			return chartRefResponse, err
		}
		if envOverride != nil && envOverride.Chart != nil {
			chartRefResponse.LatestEnvChartRef = envOverride.Chart.ChartRefId
		} else {
			chartRefResponse.LatestEnvChartRef = chart.ChartRefId
		}
	}
	chartRefResponse.LatestAppChartRef = chart.ChartRefId
	chartRefResponse.ChartRefs = chartRefs
	chartRefResponse.LatestChartRef = LatestAppChartRef
	return chartRefResponse, nil
}

func (impl *ChartServiceImpl) FindPreviousChartByAppId(appId int) (chartTemplate *TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindPreviousChartByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = impl.chartAdaptor(chart, false)
	return chartTemplate, err
}

func (impl *ChartServiceImpl) UpgradeForApp(appId int, chartRefId int, newAppOverride map[string]interface{}, userId int32, ctx context.Context) (bool, error) {

	currentChart, err := impl.FindLatestChartForAppByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Error(err)
		return false, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Errorw("no chart configured for this app", "appId", appId)
		return false, fmt.Errorf("no chart configured for this app, skip it for upgrade")
	}

	templateRequest := TemplateRequest{}
	templateRequest.ChartRefId = chartRefId
	templateRequest.AppId = appId
	templateRequest.ChartRepositoryId = currentChart.ChartRepositoryId
	templateRequest.DefaultAppOverride = newAppOverride["defaultAppOverride"].(json.RawMessage)
	templateRequest.ValuesOverride = currentChart.DefaultAppOverride
	templateRequest.UserId = userId
	templateRequest.IsBasicViewLocked = currentChart.IsBasicViewLocked
	templateRequest.CurrentViewEditor = currentChart.CurrentViewEditor
	upgradedChartReq, err := impl.Create(templateRequest, ctx)
	if err != nil {
		return false, err
	}
	if upgradedChartReq == nil || upgradedChartReq.Id == 0 {
		impl.logger.Infow("unable to upgrade app", "appId", appId)
		return false, fmt.Errorf("unable to upgrade app, got no error on creating chart but unable to complete")
	}
	updatedChart, err := impl.chartRepository.FindById(upgradedChartReq.Id)
	if err != nil {
		return false, err
	}

	//STEP 2 - env upgrade
	impl.logger.Debugw("creating env and pipeline config for app", "appId", appId)
	//step 1
	envOverrides, err := impl.envOverrideRepository.GetEnvConfigByChartId(currentChart.Id)
	if err != nil && envOverrides == nil {
		return false, err
	}
	for _, envOverride := range envOverrides {

		//STEP 4 = create environment config
		env, err := impl.environmentRepository.FindById(envOverride.TargetEnvironment)
		if err != nil {
			return false, err
		}
		envOverrideNew := &chartConfig.EnvConfigOverride{
			Active:            true,
			ManualReviewed:    true,
			Status:            models.CHARTSTATUS_SUCCESS,
			EnvOverrideValues: string(envOverride.EnvOverrideValues),
			TargetEnvironment: envOverride.TargetEnvironment,
			ChartId:           updatedChart.Id,
			AuditLog:          sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now(), CreatedBy: userId},
			Namespace:         env.Namespace,
			Latest:            true,
			Previous:          false,
			IsBasicViewLocked: envOverride.IsBasicViewLocked,
			CurrentViewEditor: envOverride.CurrentViewEditor,
		}
		err = impl.envOverrideRepository.Save(envOverrideNew)
		if err != nil {
			impl.logger.Errorw("error in creating env config", "data", envOverride, "error", err)
			return false, err
		}
		//creating history entry for deployment template
		isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagForAPipelineByAppIdAndEnvId(appId, envOverrideNew.TargetEnvironment)
		if err != nil {
			impl.logger.Errorw("error, GetMetricsFlagForAPipelineByAppIdAndEnvId", "err", err, "appId", appId, "envId", envOverrideNew.TargetEnvironment)
			return false, err
		}
		err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverrideNew, nil, isAppMetricsEnabled, 0)
		if err != nil {
			impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", envOverrideNew)
			return false, err
		}
		//VARIABLE_MAPPING_UPDATE
		err = impl.scopedVariableManager.ExtractAndMapVariables(envOverrideNew.EnvOverrideValues, envOverrideNew.Id, variablesRepository.EntityTypeDeploymentTemplateEnvLevel, envOverrideNew.CreatedBy, nil)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (impl *ChartServiceImpl) CheckIfChartRefUserUploadedByAppId(id int) (bool, error) {
	chartInfo, err := impl.chartRepository.FindLatestChartForAppByAppId(id)
	if err != nil {
		return false, err
	}
	chartData, err := impl.chartRefService.FindById(chartInfo.ChartRefId)
	if err != nil {
		return false, err
	}
	return chartData.UserUploaded, err
}

func (impl *ChartServiceImpl) UpdateGitRepoUrlInCharts(appId int, repoUrl, chartLocation string, userId int32) error {
	charts, err := impl.chartRepository.FindActiveChartsByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		return err
	}
	for _, ch := range charts {
		if ChartsUtil.IsGitOpsRepoNotConfigured(ch.GitRepoUrl) {
			ch.GitRepoUrl = repoUrl
			ch.ChartLocation = chartLocation
			ch.UpdatedOn = time.Now()
			ch.UpdatedBy = userId
			err = impl.chartRepository.Update(ch)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl *ChartServiceImpl) IsGitOpsRepoAlreadyRegistered(gitOpsRepoUrl string) (bool, error) {
	chartModel, err := impl.chartRepository.FindChartByGitRepoUrl(gitOpsRepoUrl)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching chartModel", "repoUrl", gitOpsRepoUrl, "err", err)
		return true, err
	}
	if util.IsErrNoRows(err) {
		return false, nil
	}
	impl.logger.Errorw("repository is already in use for devtron app", "repoUrl", gitOpsRepoUrl, "appId", chartModel.AppId)
	return true, nil
}

func (impl *ChartServiceImpl) isGitRepoUrlPresent(appId int) bool {
	fetchedChart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error fetching git repo url from the latest chart")
		return false
	}
	if ChartsUtil.IsGitOpsRepoNotConfigured(fetchedChart.GitRepoUrl) {
		return false
	}
	return true
}
