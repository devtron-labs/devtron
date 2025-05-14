/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package read

import (
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/pkg/chart/adaptor"
	"github.com/devtron-labs/devtron/pkg/chart/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	chartRefRead "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/read"
	"go.uber.org/zap"
)

type ChartReadService interface {
	GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *bean.TemplateRequest, err error)
	IsGitOpsRepoConfiguredForDevtronApps(appIds []int) (map[int]bool, error)
	FindLatestChartForAppByAppId(appId int) (chartTemplate *bean.TemplateRequest, err error)
	GetChartRefConfiguredForApp(appId int) (*bean3.ChartRefDto, error)
}

type ChartReadServiceImpl struct {
	logger                    *zap.SugaredLogger
	chartRepository           chartRepoRepository.ChartRepository
	deploymentConfigService   common.DeploymentConfigService
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService
	gitOpsConfigReadService   config.GitOpsConfigReadService
	ChartRefReadService       chartRefRead.ChartRefReadService
}

func NewChartReadServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	deploymentConfigService common.DeploymentConfigService,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	ChartRefReadService chartRefRead.ChartRefReadService) *ChartReadServiceImpl {
	return &ChartReadServiceImpl{
		logger:                    logger,
		chartRepository:           chartRepository,
		deploymentConfigService:   deploymentConfigService,
		deployedAppMetricsService: deployedAppMetricsService,
		gitOpsConfigReadService:   gitOpsConfigReadService,
		ChartRefReadService:       ChartRefReadService,
	}

}

func (impl *ChartReadServiceImpl) GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *bean.TemplateRequest, err error) {
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
	deploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(appId, 0)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment config by appId", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = adaptor.ChartAdaptor(chart, isAppMetricsEnabled, deploymentConfig)
	return chartTemplate, err
}

func (impl *ChartReadServiceImpl) IsGitOpsRepoConfiguredForDevtronApps(appIds []int) (map[int]bool, error) {
	gitOpsConfigStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId")
		return nil, err
	}
	appIdRepoConfiguredMap := make(map[int]bool, len(appIds))
	for _, appId := range appIds {
		if !gitOpsConfigStatus.IsGitOpsConfiguredAndArgoCdInstalled() {
			appIdRepoConfiguredMap[appId] = false
		} else if !gitOpsConfigStatus.AllowCustomRepository {
			appIdRepoConfiguredMap[appId] = true
		} else {
			latestChartConfiguredInApp, err := impl.FindLatestChartForAppByAppId(appId)
			if err != nil {
				impl.logger.Errorw("error in fetching latest chart for app by appId")
				return nil, err
			}
			appIdRepoConfiguredMap[appId] = !apiGitOpsBean.IsGitOpsRepoNotConfigured(latestChartConfiguredInApp.GitRepoUrl)
		}
	}
	return appIdRepoConfiguredMap, nil
}

func (impl *ChartReadServiceImpl) FindLatestChartForAppByAppId(appId int) (chartTemplate *bean.TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}

	deploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(appId, 0)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment config by appId", "appId", appId, "err", err)
		return nil, err
	}

	isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app-metrics", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = adaptor.ChartAdaptor(chart, isAppMetricsEnabled, deploymentConfig)
	return chartTemplate, err
}

func (impl *ChartReadServiceImpl) GetChartRefConfiguredForApp(appId int) (*bean3.ChartRefDto, error) {
	latestChart, err := impl.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in finding latest chart by appId", "appId", appId, "err", err)
		return nil, nil
	}
	chartRef, err := impl.ChartRefReadService.FindById(latestChart.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding latest chart by appId", "appId", appId, "err", err)
		return nil, nil
	}
	return chartRef, nil
}
