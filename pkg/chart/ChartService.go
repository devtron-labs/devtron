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

package chart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chart/adaptor"
	bean3 "github.com/devtron-labs/devtron/pkg/chart/bean"
	read2 "github.com/devtron-labs/devtron/pkg/chart/read"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	adapter2 "github.com/devtron-labs/devtron/pkg/deployment/common/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	chartRefBean "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	variablesRepository "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"
)

type ChartService interface {
	Create(templateRequest bean3.TemplateRequest, ctx context.Context) (chart *bean3.TemplateRequest, err error)
	CreateChartFromEnvOverride(ctx context.Context, templateRequest bean3.TemplateRequest) (chart *bean3.TemplateRequest, err error)
	UpdateAppOverride(ctx context.Context, templateRequest *bean3.TemplateRequest) (*bean3.TemplateRequest, error)
	IsReadyToTrigger(appId int, envId int, pipelineId int) (IsReady, error)
	FindPreviousChartByAppId(appId int) (chartTemplate *bean3.TemplateRequest, err error)
	UpgradeForApp(appId int, chartRefId int, newAppOverride map[string]interface{}, userId int32, ctx context.Context) (bool, error)
	CheckIfChartRefUserUploadedByAppId(id int) (bool, error)

	ChartRefAutocompleteGlobalData() (*chartRefBean.ChartRefAutocompleteResponse, error)
	ChartRefAutocompleteForAppOrEnv(appId int, envId int) (*chartRefBean.ChartRefAutocompleteResponse, error)

	ConfigureGitOpsRepoUrlForApp(appId int, repoUrl, chartLocation string, isCustomRepo bool, userId int32) (*bean2.DeploymentConfig, error)

	IsGitOpsRepoConfiguredForDevtronApp(appId int) (bool, error)
	IsGitOpsRepoAlreadyRegistered(gitOpsRepoUrl string) (bool, error)

	GetDeploymentTemplateDataByAppIdAndCharRefId(appId, chartRefId int) (map[string]interface{}, error)

	ChartServiceEnt
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
	environmentRepository            repository.EnvironmentRepository
	deploymentTemplateHistoryService deploymentTemplate.DeploymentTemplateHistoryService
	scopedVariableManager            variables.ScopedVariableManager
	deployedAppMetricsService        deployedAppMetrics.DeployedAppMetricsService
	chartRefService                  chartRef.ChartRefService
	gitOpsConfigReadService          config.GitOpsConfigReadService
	deploymentConfigService          common.DeploymentConfigService
	envConfigOverrideReadService     read.EnvConfigOverrideService
	chartReadService                 read2.ChartReadService
}

func NewChartServiceImpl(chartRepository chartRepoRepository.ChartRepository,
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	repoRepository chartRepoRepository.ChartRepoRepository,
	pipelineGroupRepository app.AppRepository,
	mergeUtil util.MergeUtil,
	envOverrideRepository chartConfig.EnvConfigOverrideRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	environmentRepository repository.EnvironmentRepository,
	deploymentTemplateHistoryService deploymentTemplate.DeploymentTemplateHistoryService,
	scopedVariableManager variables.ScopedVariableManager,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	chartRefService chartRef.ChartRefService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	deploymentConfigService common.DeploymentConfigService,
	envConfigOverrideReadService read.EnvConfigOverrideService,
	chartReadService read2.ChartReadService) *ChartServiceImpl {
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
		deploymentConfigService:          deploymentConfigService,
		envConfigOverrideReadService:     envConfigOverrideReadService,
		chartReadService:                 chartReadService,
	}
}

func (impl *ChartServiceImpl) Create(templateRequest bean3.TemplateRequest, ctx context.Context) (*bean3.TemplateRequest, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ChartServiceImpl.Create")
	defer span.End()
	err := impl.chartRefService.CheckChartExists(templateRequest.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting missing chart for chartRefId", "err", err, "chartRefId")
		return nil, err
	}
	chartMeta, err := impl.getChartMetaData(templateRequest)
	if err != nil {
		return nil, err
	}

	existingChart, _ := impl.chartRepository.FindChartByAppIdAndRefId(templateRequest.AppId, templateRequest.ChartRefId)
	if existingChart != nil && existingChart.Id > 0 {
		return nil, fmt.Errorf("this reference chart already has added to appId %d refId %d", templateRequest.AppId, templateRequest.Id)
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

	// STARTS
	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	if err != nil && pg.ErrNoRows != err {
		return nil, err
	}
	gitRepoUrl := apiGitOpsBean.GIT_REPO_NOT_CONFIGURED
	if currentLatestChart.GitRepoUrl != "" {
		gitRepoUrl = currentLatestChart.GitRepoUrl
	}

	tx, err := impl.chartRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update charts", "error", err)
		return nil, err
	}
	defer impl.chartRepository.RollbackTx(tx)

	impl.logger.Debugw("current latest chart in db", "chartId", currentLatestChart.Id)
	if currentLatestChart.Id > 0 {
		err = impl.UpdateExistingChartsToLatestFalse(tx, &templateRequest, currentLatestChart)
		if err != nil {
			impl.logger.Errorw("error in updating charts", "chartId", currentLatestChart.Id, "error", err)
			return nil, err
		}
	}

	// ENDS

	impl.logger.Debug("now finally create new chart and make it latest entry in db and previous flag = true")

	version, err := impl.getNewVersion(charRepository.Name, chartMeta.Name, refChart)
	chartMeta.Version = version
	if err != nil {
		return nil, err
	}
	chartValues, err := impl.chartTemplateService.FetchValuesFromReferenceChart(chartMeta, refChart, pipelineStrategyPath)
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
		GitRepoUrl:              gitRepoUrl,
		ChartLocation:           chartLocation,
		Status:                  models.CHARTSTATUS_NEW,
		Active:                  true,
		ReferenceTemplate:       templateName,
		ChartRefId:              templateRequest.ChartRefId,
		Latest:                  true,
		Previous:                false,
		IsBasicViewLocked:       templateRequest.IsBasicViewLocked,
		CurrentViewEditor:       templateRequest.CurrentViewEditor,
		AuditLog:                sql.AuditLog{CreatedBy: templateRequest.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: templateRequest.UserId},
	}

	err = impl.chartRepository.Save(tx, chart)
	if err != nil {
		impl.logger.Errorw("error in saving chart ", "chart", chart, "error", err)
		//If found any error, rollback chart museum
		return nil, err
	}

	deploymentConfig := &bean2.DeploymentConfig{
		AppId:      templateRequest.AppId,
		ConfigType: adapter2.GetDeploymentConfigType(templateRequest.IsCustomGitRepository),
		RepoURL:    gitRepoUrl,
		ReleaseConfiguration: &bean2.ReleaseConfiguration{
			Version: bean2.Version,
			ArgoCDSpec: bean2.ArgoCDSpec{
				Spec: bean2.ApplicationSpec{
					Source: &bean2.ApplicationSource{
						RepoURL: gitRepoUrl,
						Path:    chartLocation,
					},
				},
			},
		},
		Active: true,
	}
	deploymentConfig, err = impl.deploymentConfigService.CreateOrUpdateConfig(tx, deploymentConfig, templateRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config", "appId", templateRequest.AppId, "err", err)
		return nil, err
	}

	err = impl.updateChartLocationForEnvironmentConfigs(newCtx, tx, templateRequest.AppId, chart.ChartRefId, templateRequest.UserId, version)
	if err != nil {
		impl.logger.Errorw("error in updating chart location in env overrides", "appId", templateRequest.AppId, "err", err)
		return nil, err
	}

	//creating history entry for deployment template
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(chart, tx, templateRequest.IsAppMetricsEnabled)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", chart)
		return nil, err
	}

	//VARIABLE_MAPPING_UPDATE
	err = impl.scopedVariableManager.ExtractAndMapVariables(chart.GlobalOverride, chart.Id, variablesRepository.EntityTypeDeploymentTemplateAppLevel, chart.CreatedBy, nil)
	if err != nil {
		return nil, err
	}

	err = impl.chartRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update charts", "error", err)
		return nil, err
	}

	appLevelMetricsUpdateReq := &bean.DeployedAppMetricsRequest{
		EnableMetrics: templateRequest.IsAppMetricsEnabled,
		AppId:         templateRequest.AppId,
		ChartRefId:    templateRequest.ChartRefId,
		UserId:        templateRequest.UserId,
	}
	err = impl.deployedAppMetricsService.CreateOrUpdateAppOrEnvLevelMetrics(newCtx, appLevelMetricsUpdateReq)
	if err != nil {
		impl.logger.Errorw("error, CheckAndUpdateAppOrEnvLevelMetrics", "err", err, "req", appLevelMetricsUpdateReq)
		return nil, err
	}

	chartVal, err := adaptor.ChartAdaptor(chart, appLevelMetricsUpdateReq.EnableMetrics, deploymentConfig)
	return chartVal, err
}

func (impl *ChartServiceImpl) UpdateExistingChartsToLatestFalse(tx *pg.Tx, templateRequest *bean3.TemplateRequest, currentLatestChart *chartRepoRepository.Chart) error {

	impl.logger.Debugw("updating env and pipeline config which are currently latest in db", "chartId", currentLatestChart.Id)

	impl.logger.Debug("updating all other charts which are not latest but may be set previous true, setting previous=false")
	//step 2
	noLatestCharts, dbErr := impl.chartRepository.FindNoLatestChartForAppByAppId(templateRequest.AppId)
	if dbErr != nil && !util.IsErrNoRows(dbErr) {
		impl.logger.Errorw("error in getting non-latest charts", "appId", templateRequest.AppId, "err", dbErr)
		return dbErr
	}
	var updatedCharts []*chartRepoRepository.Chart
	for _, noLatestChart := range noLatestCharts {
		if noLatestChart.Id != templateRequest.Id {
			noLatestChart.Latest = false // these are already false by d way
			noLatestChart.Previous = false
			updatedCharts = append(updatedCharts, noLatestChart)
		}
	}
	err := impl.chartRepository.UpdateAllInTx(tx, updatedCharts)
	if err != nil {
		return err
	}
	impl.logger.Debug("now going to update latest entry in db to false and previous flag = true")
	// now finally update latest entry in db to false and previous true
	currentLatestChart.Latest = false // these are already false by d way
	currentLatestChart.Previous = true
	err = impl.chartRepository.Update(tx, currentLatestChart)
	if err != nil {
		return err
	}
	return nil
}

func (impl *ChartServiceImpl) updateChartLocationForEnvironmentConfigs(ctx context.Context, tx *pg.Tx, appId, chartRefId int, userId int32, version string) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ChartServiceImpl.updateChartLocationForEnvironmentConfigs")
	defer span.End()
	envOverrides, err := impl.envConfigOverrideReadService.GetAllOverridesForApp(newCtx, appId)
	if err != nil {
		impl.logger.Errorw("error in getting all overrides for app", "appId", appId, "err", err)
		return err
	}
	uniqueEnvMap := make(map[int]bool)
	for _, override := range envOverrides {
		if _, ok := uniqueEnvMap[override.TargetEnvironment]; !ok && !override.IsOverride {
			uniqueEnvMap[override.TargetEnvironment] = true
			err := impl.deploymentConfigService.UpdateChartLocationInDeploymentConfig(tx, appId, override.TargetEnvironment, chartRefId, userId, version)
			if err != nil {
				impl.logger.Errorw("error in updating chart location for env level deployment configs", "appId", appId, "envId", override.TargetEnvironment, "err", err)
				return err
			}
		}
	}
	return nil
}

func (impl *ChartServiceImpl) CreateChartFromEnvOverride(ctx context.Context, templateRequest bean3.TemplateRequest) (*bean3.TemplateRequest, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "ChartServiceImpl.CreateChartFromEnvOverride")
	defer span.End()
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

	impl.logger.Debug("now finally create new chart and make it latest entry in db and previous flag = true")
	version, err := impl.getNewVersion(chartRepository.Name, chartMeta.Name, refChart)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	chartMeta.Version = version
	if err != nil {
		return nil, err
	}
	chartValues, err := impl.chartTemplateService.FetchValuesFromReferenceChart(chartMeta, refChart, pipelineStrategyPath)
	if err != nil {
		return nil, err
	}

	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	if err != nil && pg.ErrNoRows != err {
		return nil, err
	}

	tx, err := impl.chartRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update charts", "error", err)
		return nil, err
	}
	defer impl.chartRepository.RollbackTx(tx)

	deploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(tx, templateRequest.AppId, 0)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config by appId", "appId", templateRequest.AppId, "err", err)
		return nil, err
	}

	chartLocation := filepath.Join(templateName, version)
	gitRepoUrl := apiGitOpsBean.GIT_REPO_NOT_CONFIGURED
	if currentLatestChart.Id > 0 && deploymentConfig.GetRepoURL() != "" {
		gitRepoUrl = currentLatestChart.GitRepoUrl
	}

	// maintained for backward compatibility;
	// adding git repo url to both deprecated and new state
	deploymentConfig = deploymentConfig.SetRepoURL(gitRepoUrl)
	deploymentConfig.SetChartLocation(chartLocation)

	deploymentConfig, err = impl.deploymentConfigService.CreateOrUpdateConfig(tx, deploymentConfig, templateRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config", "appId", templateRequest.AppId, "err", err)
		return nil, err
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

	err = impl.chartRepository.Save(tx, chart)
	if err != nil {
		impl.logger.Errorw("error in saving chart ", "chart", chart, "error", err)
		return nil, err
	}

	//creating history entry for deployment template
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(chart, tx, appMetrics)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", chart)
		return nil, err
	}
	//VARIABLE_MAPPING_UPDATE
	err = impl.scopedVariableManager.ExtractAndMapVariables(chart.GlobalOverride, chart.Id, variablesRepository.EntityTypeDeploymentTemplateAppLevel, chart.CreatedBy, nil)
	if err != nil {
		return nil, err
	}

	err = impl.chartRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update charts", "error", err)
		return nil, err
	}

	chartVal, err := adaptor.ChartAdaptor(chart, false, deploymentConfig)
	return chartVal, err
}

func (impl *ChartServiceImpl) getChartMetaData(templateRequest bean3.TemplateRequest) (*chart.Metadata, error) {
	pg, err := impl.pipelineGroupRepository.FindById(templateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching pg", "id", templateRequest.AppId, "err", err)
	}
	metadata := &chart.Metadata{
		Name: pg.AppName,
	}
	return metadata, err
}

func (impl *ChartServiceImpl) getChartRepo(templateRequest bean3.TemplateRequest) (*chartRepoRepository.ChartRepo, error) {
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

func (impl *ChartServiceImpl) IsGitOpsRepoConfiguredForDevtronApp(appId int) (bool, error) {
	gitOpsConfigStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId")
		return false, err
	} else if !gitOpsConfigStatus.IsGitOpsConfiguredAndArgoCdInstalled() {
		return false, nil
	} else if !gitOpsConfigStatus.AllowCustomRepository {
		return true, nil
	}
	latestChartConfiguredInApp, err := impl.chartReadService.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId")
		return false, err
	}
	return !apiGitOpsBean.IsGitOpsRepoNotConfigured(latestChartConfiguredInApp.GitRepoUrl), nil
}

func (impl *ChartServiceImpl) UpdateAppOverride(ctx context.Context, templateRequest *bean3.TemplateRequest) (*bean3.TemplateRequest, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ChartServiceImpl.UpdateAppOverride")
	defer span.End()
	_, span = otel.Tracer("orchestrator").Start(newCtx, "chartRepository.FindById")
	template, err := impl.chartRepository.FindById(templateRequest.Id)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching chart config", "id", templateRequest.Id, "err", err)
		return nil, err
	}

	// STARTS
	_, span = otel.Tracer("orchestrator").Start(newCtx, "chartRepository.FindLatestChartForAppByAppId")
	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	span.End()
	if err != nil {
		return nil, err
	}

	chartRef, err := impl.chartRefService.FindById(template.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding chart ref by id", "chartRefId", template.ChartRefId, "err", err)
		return nil, err
	}

	tx, err := impl.chartRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update charts", "error", err)
		return nil, err
	}
	defer impl.chartRepository.RollbackTx(tx)

	if currentLatestChart.Id > 0 && currentLatestChart.Id == templateRequest.Id {

	} else if currentLatestChart.Id != templateRequest.Id {
		err = impl.UpdateExistingChartsToLatestFalse(tx, templateRequest, currentLatestChart)
		if err != nil {
			impl.logger.Errorw("error in updating charts", "chartId", currentLatestChart.Id, "error", err)
			return nil, err
		}
	}
	// ENDS

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
	_, span = otel.Tracer("orchestrator").Start(newCtx, "chartRepository.Update.requestTemplate")
	err = impl.chartRepository.Update(tx, template)
	span.End()
	if err != nil {
		return nil, err
	}

	config, err := impl.deploymentConfigService.GetConfigForDevtronApps(tx, template.AppId, 0)
	if err != nil {
		impl.logger.Errorw("error in fetching config", "appId", template.AppId, "err", err)
		return nil, err
	}

	chartGitLocation := filepath.Join(chartRef.Location, template.ChartVersion)
	deploymentConfig := &bean2.DeploymentConfig{
		AppId:      template.AppId,
		ConfigType: adapter2.GetDeploymentConfigType(template.IsCustomGitRepository),
		RepoURL:    config.GetRepoURL(),
		ReleaseConfiguration: &bean2.ReleaseConfiguration{
			Version: bean2.Version,
			ArgoCDSpec: bean2.ArgoCDSpec{
				Spec: bean2.ApplicationSpec{
					Source: &bean2.ApplicationSource{
						RepoURL: config.GetRepoURL(),
						Path:    chartGitLocation,
					},
				},
			},
		},
		Active: true,
	}

	deploymentConfig, err = impl.deploymentConfigService.CreateOrUpdateConfig(tx, deploymentConfig, templateRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in creating or updating deploymentConfig", "appId", templateRequest.AppId, "err", err)
		return nil, err
	}

	if currentLatestChart.Id != 0 && currentLatestChart.Id != templateRequest.Id {
		err = impl.updateChartLocationForEnvironmentConfigs(newCtx, tx, templateRequest.AppId, templateRequest.ChartRefId, templateRequest.UserId, template.ChartVersion)
		if err != nil {
			impl.logger.Errorw("error in updating chart location in env overrides", "appId", templateRequest.AppId, "err", err)
			return nil, err
		}
	}

	appLevelMetricsUpdateReq := &bean.DeployedAppMetricsRequest{
		EnableMetrics: templateRequest.IsAppMetricsEnabled,
		AppId:         templateRequest.AppId,
		ChartRefId:    templateRequest.ChartRefId,
		UserId:        templateRequest.UserId,
	}
	err = impl.deployedAppMetricsService.CreateOrUpdateAppOrEnvLevelMetrics(newCtx, appLevelMetricsUpdateReq)
	if err != nil {
		impl.logger.Errorw("error, CheckAndUpdateAppOrEnvLevelMetrics", "err", err, "req", appLevelMetricsUpdateReq)
		return nil, err
	}
	_, span = otel.Tracer("orchestrator").Start(newCtx, "CreateDeploymentTemplateHistoryFromGlobalTemplate")
	//creating history entry for deployment template
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(template, tx, appLevelMetricsUpdateReq.EnableMetrics)
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
	envOverride, err := impl.envConfigOverrideReadService.ActiveEnvConfigOverride(appId, envId)
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
	chartRefResponse, err := impl.ChartRefAutocompleteGlobalData()
	if err != nil {
		impl.logger.Errorw("error, ChartRefAutocompleteGlobalData", "err", err)
		return nil, err
	}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching latest chart", "err", err)
		return chartRefResponse, err
	}
	chartRefResponse.LatestAppChartRef = chart.ChartRefId
	if envId > 0 {
		envOverride, err := impl.envConfigOverrideReadService.FindLatestChartForAppByAppIdAndEnvId(appId, envId)
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
	return chartRefResponse, nil
}

func (impl *ChartServiceImpl) ChartRefAutocompleteGlobalData() (*chartRefBean.ChartRefAutocompleteResponse, error) {
	results, err := impl.chartRefService.GetAll()
	if err != nil {
		impl.logger.Errorw("error in fetching chart ref", "err", err)
		return nil, err
	}
	resultsMetadataMap, err := impl.chartRefService.GetAllChartMetadata()
	if err != nil {
		impl.logger.Errorw("error in fetching chart metadata", "err", err)
		return nil, err
	}
	var latestChartRef int
	chartRefs := make([]chartRefBean.ChartRefAutocompleteDto, 0, len(results))
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
			latestChartRef = result.Id
		}
	}
	return &chartRefBean.ChartRefAutocompleteResponse{
		ChartsMetadata: resultsMetadataMap,
		ChartRefs:      chartRefs,
		LatestChartRef: latestChartRef,
	}, nil
}

func (impl *ChartServiceImpl) FindPreviousChartByAppId(appId int) (chartTemplate *bean3.TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindPreviousChartByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}
	deploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(nil, appId, 0)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment config by appId", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = adaptor.ChartAdaptor(chart, false, deploymentConfig)
	return chartTemplate, err
}

func (impl *ChartServiceImpl) UpgradeForApp(appId int, chartRefId int, newAppOverride map[string]interface{}, userId int32, ctx context.Context) (bool, error) {

	currentChart, err := impl.chartReadService.FindLatestChartForAppByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Error(err)
		return false, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Errorw("no chart configured for this app", "appId", appId)
		return false, fmt.Errorf("no chart configured for this app, skip it for upgrade")
	}

	templateRequest := bean3.TemplateRequest{}
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
	envOverrides, err := impl.envConfigOverrideReadService.GetEnvConfigByChartId(currentChart.Id)
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
		envOverrideNewDTO := adapter.EnvOverrideDBToDTO(envOverrideNew)
		err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverrideNewDTO, nil, isAppMetricsEnabled, 0)
		if err != nil {
			impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", envOverrideNewDTO)
			return false, err
		}
		//VARIABLE_MAPPING_UPDATE
		//TODO ayush, check if this is needed
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

func (impl *ChartServiceImpl) ConfigureGitOpsRepoUrlForApp(appId int, repoUrl, chartLocation string, isCustomRepo bool, userId int32) (*bean2.DeploymentConfig, error) {

	////update in both charts and deployment config
	charts, err := impl.chartRepository.FindActiveChartsByAppId(appId)
	if err != nil {
		return nil, err
	}
	tx, err := impl.chartRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update charts", "error", err)
		return nil, err
	}
	defer impl.chartRepository.RollbackTx(tx)
	var updatedCharts []*chartRepoRepository.Chart
	for _, ch := range charts {
		if !ch.IsCustomGitRepository {
			ch.GitRepoUrl = repoUrl
			ch.UpdateAuditLog(userId)
			updatedCharts = append(updatedCharts, ch)
		}
	}
	err = impl.chartRepository.UpdateAllInTx(tx, updatedCharts)
	if err != nil {
		return nil, err
	}
	err = impl.chartRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update charts", "error", err)
		return nil, err
	}

	deploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(tx, appId, 0)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config", "appId", appId, "error", err)
		return nil, err
	}
	deploymentConfig = deploymentConfig.SetRepoURL(repoUrl)
	deploymentConfig.SetChartLocation(chartLocation)

	deploymentConfig, err = impl.deploymentConfigService.CreateOrUpdateConfig(tx, deploymentConfig, userId)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config for app", "appId", appId, "err", err)
		return nil, err
	}

	return deploymentConfig, nil
}

//func (impl *ChartServiceImpl) OverrideGitOpsRepoUrl(appId int, repoUrl string, userId int32) error {
//	charts, err := impl.chartRepository.FindActiveChartsByAppId(appId)
//	if err != nil {
//		return err
//	}
//	tx, err := impl.chartRepository.StartTx()
//	if err != nil {
//		impl.logger.Errorw("error in starting transaction to update charts", "error", err)
//		return err
//	}
//	defer impl.chartRepository.RollbackTx(tx)
//	var updatedCharts []*chartRepoRepository.Chart
//	for _, ch := range charts {
//		if !ch.IsCustomGitRepository {
//			ch.GitRepoUrl = repoUrl
//			ch.UpdateAuditLog(userId)
//			updatedCharts = append(updatedCharts, ch)
//		}
//	}
//	err = impl.chartRepository.UpdateAllInTx(tx, updatedCharts)
//	if err != nil {
//		return err
//	}
//	err = impl.chartRepository.CommitTx(tx)
//	if err != nil {
//		impl.logger.Errorw("error in committing transaction to update charts", "error", err)
//		return err
//	}
//	return nil
//}

func (impl *ChartServiceImpl) IsGitOpsRepoAlreadyRegistered(gitOpsRepoUrl string) (bool, error) {

	isURLPresent, err := impl.deploymentConfigService.CheckIfURLAlreadyPresent(gitOpsRepoUrl)
	if err != nil {
		impl.logger.Errorw("error in checking if gitOps repo url is already present", "error", err)
		return false, err
	}
	if isURLPresent {
		return true, nil
	}

	chartModel, err := impl.chartRepository.FindChartByGitRepoUrl(gitOpsRepoUrl)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching chartModel", "repoUrl", gitOpsRepoUrl, "err", err)
		return true, err
	} else if util.IsErrNoRows(err) {
		return false, nil
	}
	impl.logger.Errorw("repository is already in use for devtron app", "repoUrl", gitOpsRepoUrl, "appId", chartModel.AppId)
	return true, nil
}

func (impl *ChartServiceImpl) GetDeploymentTemplateDataByAppIdAndCharRefId(appId, chartRefId int) (map[string]interface{}, error) {
	appConfigResponse := make(map[string]interface{})
	appConfigResponse["globalConfig"] = nil

	err := impl.chartRefService.CheckChartExists(chartRefId)
	if err != nil {
		impl.logger.Errorw("refChartDir Not Found err, JsonSchemaExtractFromFile", err)
		return nil, err
	}

	schema, readme, err := impl.chartRefService.GetSchemaAndReadmeForTemplateByChartRefId(chartRefId)
	if err != nil {
		impl.logger.Errorw("err in getting schema and readme, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
	}

	template, err := impl.chartReadService.FindLatestChartForAppByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
		return nil, err
	}

	if pg.ErrNoRows == err {
		appOverride, _, err := impl.chartRefService.GetAppOverrideForDefaultTemplate(chartRefId)
		if err != nil {
			impl.logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			return nil, err
		}
		appOverride["schema"] = json.RawMessage(schema)
		appOverride["readme"] = string(readme)
		mapB, err := json.Marshal(appOverride)
		if err != nil {
			impl.logger.Errorw("marshal err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			return nil, err
		}
		appConfigResponse["globalConfig"] = json.RawMessage(mapB)
	} else {
		if template.ChartRefId != chartRefId {
			templateRequested, err := impl.chartReadService.GetByAppIdAndChartRefId(appId, chartRefId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
				return nil, err
			}

			if pg.ErrNoRows == err {
				template.ChartRefId = chartRefId
				template.Id = 0
				template.Latest = false
			} else {
				template.ChartRefId = templateRequested.ChartRefId
				template.Id = templateRequested.Id
				template.ChartRepositoryId = templateRequested.ChartRepositoryId
				template.RefChartTemplate = templateRequested.RefChartTemplate
				template.RefChartTemplateVersion = templateRequested.RefChartTemplateVersion
				template.Latest = templateRequested.Latest
			}
		}
		template.Schema = schema
		template.Readme = string(readme)
		bytes, err := json.Marshal(template)
		if err != nil {
			impl.logger.Errorw("marshal err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			return nil, err
		}
		appOverride := json.RawMessage(bytes)
		appConfigResponse["globalConfig"] = appOverride
	}
	return appConfigResponse, nil
}
