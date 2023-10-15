package pipeline

import (
	"encoding/json"
	"fmt"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronAppStrategyService interface {
	//FetchCDPipelineStrategy : Retrieve CDPipelineStrategy for given appId
	FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error)

	// FetchDefaultCDPipelineStrategy :
	// TODO: uncomment this after code has been refactored as this function doesnt contain any logic to fetch strategy
	FetchDefaultCDPipelineStrategy(appId int, envId int) (PipelineStrategy, error)
}

type DevtronAppStrategyServiceImpl struct {
	logger                                          *zap.SugaredLogger
	chartRepository                                 chartRepoRepository.ChartRepository
	globalStrategyMetadataChartRefMappingRepository chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepository
	ciCdPipelineOrchestrator                        CiCdPipelineOrchestrator

	cdPipelineConfigService CdPipelineConfigService
}

func NewDevtronAppStrategyServiceImpl(
	logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	globalStrategyMetadataChartRefMappingRepository chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepository,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator,
	cdPipelineConfigService CdPipelineConfigService) *DevtronAppStrategyServiceImpl {
	return &DevtronAppStrategyServiceImpl{
		logger:          logger,
		chartRepository: chartRepository,
		globalStrategyMetadataChartRefMappingRepository: globalStrategyMetadataChartRefMappingRepository,
		ciCdPipelineOrchestrator:                        ciCdPipelineOrchestrator,
		cdPipelineConfigService:                         cdPipelineConfigService,
	}
}

func (impl *DevtronAppStrategyServiceImpl) FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error) {
	pipelineStrategiesResponse := PipelineStrategiesResponse{}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorf("invalid state", "err", err, "appId", appId)
		return pipelineStrategiesResponse, err
	}
	if chart.Id == 0 {
		return pipelineStrategiesResponse, fmt.Errorf("no chart configured")
	}

	//get global strategy for this chart
	globalStrategies, err := impl.globalStrategyMetadataChartRefMappingRepository.GetByChartRefId(chart.ChartRefId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global strategies", "err", err)
		return pipelineStrategiesResponse, err
	} else if err == pg.ErrNoRows {
		impl.logger.Infow("no strategies configured for chart", "chartRefId", chart.ChartRefId)
		return pipelineStrategiesResponse, nil
	}
	pipelineOverride := chart.PipelineOverride
	for _, globalStrategy := range globalStrategies {
		pipelineStrategyJson, err := impl.filterDeploymentTemplate(globalStrategy.GlobalStrategyMetadata.Key, pipelineOverride)
		if err != nil {
			return pipelineStrategiesResponse, err
		}
		pipelineStrategy := PipelineStrategy{
			DeploymentTemplate: globalStrategy.GlobalStrategyMetadata.Name,
			Config:             []byte(pipelineStrategyJson),
		}
		pipelineStrategy.Default = globalStrategy.Default
		pipelineStrategiesResponse.PipelineStrategy = append(pipelineStrategiesResponse.PipelineStrategy, pipelineStrategy)
	}
	return pipelineStrategiesResponse, nil
}

func (impl *DevtronAppStrategyServiceImpl) FetchDefaultCDPipelineStrategy(appId int, envId int) (PipelineStrategy, error) {
	pipelineStrategy := PipelineStrategy{}
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

func (impl *DevtronAppStrategyServiceImpl) filterDeploymentTemplate(strategyKey string, pipelineStrategiesJson string) (string, error) {
	var pipelineStrategies DeploymentType
	err := json.Unmarshal([]byte(pipelineStrategiesJson), &pipelineStrategies)
	if err != nil {
		impl.logger.Errorw("error while unmarshal strategies", "err", err)
		return "", err
	}
	if pipelineStrategies.Deployment.Strategy[strategyKey] == nil {
		return "", fmt.Errorf("no deployment strategy found for %s", strategyKey)
	}
	strategy := make(map[string]interface{})
	strategy[strategyKey] = pipelineStrategies.Deployment.Strategy[strategyKey].(map[string]interface{})
	pipelineStrategy := DeploymentType{
		Deployment: Deployment{
			Strategy: strategy,
		},
	}
	pipelineOverrideBytes, err := json.Marshal(pipelineStrategy)
	if err != nil {
		impl.logger.Errorw("error while marshal strategies", "err", err)
		return "", err
	}
	pipelineStrategyJson := string(pipelineOverrideBytes)
	return pipelineStrategyJson, nil
}
