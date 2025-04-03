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

package pipeline

import (
	"errors"
	"fmt"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronAppStrategyService interface {
	//FetchCDPipelineStrategy : Retrieve CDPipelineStrategy for given appId
	FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error)

	// FetchDefaultCDPipelineStrategy :
	// TODO: uncomment this after code has been refactored as this function doesnt contain any logic to fetch strategy
	FetchDefaultCDPipelineStrategy(appId int, envId int) (bean.PipelineStrategy, error)
}

type DevtronAppStrategyServiceImpl struct {
	logger                                          *zap.SugaredLogger
	chartRepository                                 chartRepoRepository.ChartRepository
	globalStrategyMetadataChartRefMappingRepository chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepository
	ciCdPipelineOrchestrator                        CiCdPipelineOrchestrator
	cdPipelineConfigService                         CdPipelineConfigService
	chartRefService                                 chartRef.ChartRefService
}

func NewDevtronAppStrategyServiceImpl(
	logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	globalStrategyMetadataChartRefMappingRepository chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepository,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator,
	cdPipelineConfigService CdPipelineConfigService,
	chartRefService chartRef.ChartRefService,
) *DevtronAppStrategyServiceImpl {
	return &DevtronAppStrategyServiceImpl{
		logger:          logger,
		chartRepository: chartRepository,
		globalStrategyMetadataChartRefMappingRepository: globalStrategyMetadataChartRefMappingRepository,
		ciCdPipelineOrchestrator:                        ciCdPipelineOrchestrator,
		cdPipelineConfigService:                         cdPipelineConfigService,
		chartRefService:                                 chartRefService,
	}
}

func (impl *DevtronAppStrategyServiceImpl) FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error) {
	pipelineStrategiesResponse := PipelineStrategiesResponse{}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching chart for app", "appId", appId, "err", err)
		return pipelineStrategiesResponse, err
	}
	if chart.Id == 0 {
		return pipelineStrategiesResponse, fmt.Errorf("no chart configured")
	}
	pipelineStrategies, err := impl.chartRefService.GetDeploymentStrategiesForChartRef(chart.ChartRefId, chart.PipelineOverride)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment strategies for chart ref", "chartRefId", chart.ChartRefId, "err", err)
		return pipelineStrategiesResponse, err
	}
	pipelineStrategiesResponse.PipelineStrategy = pipelineStrategies
	return pipelineStrategiesResponse, nil
}

func (impl *DevtronAppStrategyServiceImpl) FetchDefaultCDPipelineStrategy(appId int, envId int) (bean.PipelineStrategy, error) {
	pipelineStrategy := bean.PipelineStrategy{}
	cdPipelines, err := impl.ciCdPipelineOrchestrator.GetCdPipelinesForAppAndEnv(appId, envId)
	if err != nil || (cdPipelines.Pipelines) == nil || len(cdPipelines.Pipelines) == 0 {
		return pipelineStrategy, err
	}
	cdPipelineId := cdPipelines.Pipelines[0].Id

	cdPipeline, err := impl.cdPipelineConfigService.GetCdPipelineById(cdPipelineId)
	if err != nil {
		return pipelineStrategy, nil
	}
	pipelineStrategy.DeploymentTemplate = cdPipeline.DeploymentTemplate
	pipelineStrategy.Default = true
	return pipelineStrategy, nil
}
