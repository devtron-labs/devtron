/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"context"
	"fmt"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"go.uber.org/zap"
)

type InsightsService interface {
	GetInsights(ctx context.Context, request *bean.InsightsRequest) (*bean.InsightsResponse, error)
}

type InsightsServiceImpl struct {
	logger                *zap.SugaredLogger
	appRepository         app.AppRepository
	pipelineRepository    pipelineConfig.PipelineRepository
	ciPipelineRepository  pipelineConfig.CiPipelineRepository
	ciWorkflowRepository  pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository  pipelineConfig.CdWorkflowRepository
	environmentRepository repository.EnvironmentRepository
}

func NewInsightsServiceImpl(
	logger *zap.SugaredLogger,
	appRepository app.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	environmentRepository repository.EnvironmentRepository,
) *InsightsServiceImpl {
	return &InsightsServiceImpl{
		logger:                logger,
		appRepository:         appRepository,
		pipelineRepository:    pipelineRepository,
		ciPipelineRepository:  ciPipelineRepository,
		ciWorkflowRepository:  ciWorkflowRepository,
		cdWorkflowRepository:  cdWorkflowRepository,
		environmentRepository: environmentRepository,
	}
}

func (impl *InsightsServiceImpl) GetInsights(ctx context.Context, request *bean.InsightsRequest) (*bean.InsightsResponse, error) {
	var pipelines []bean.PipelineUsageItem
	var totalCount int
	var err error

	switch request.PipelineType {
	case bean.BuildPipelines:
		pipelines, totalCount, err = impl.getTriggeredBuildPipelines(ctx, request)
		if err != nil {
			impl.logger.Errorw("error getting triggered build pipelines", "err", err)
			return nil, err
		}
	case bean.DeploymentPipelines:
		pipelines, totalCount, err = impl.getTriggeredDeploymentPipelines(ctx, request)
		if err != nil {
			impl.logger.Errorw("error getting triggered deployment pipelines", "err", err)
			return nil, err
		}
	default:
		impl.logger.Errorw("invalid pipeline type", "pipelineType", request.PipelineType)
		return nil, fmt.Errorf("invalid pipeline type: %s", request.PipelineType)
	}

	response := &bean.InsightsResponse{
		Pipelines:  pipelines,
		TotalCount: totalCount,
	}

	return response, nil
}

func (impl *InsightsServiceImpl) getTriggeredBuildPipelines(ctx context.Context, request *bean.InsightsRequest) ([]bean.PipelineUsageItem, int, error) {
	pipelineData, totalCount, err := impl.ciWorkflowRepository.GetTriggeredCIPipelines(request.TimeRangeRequest.From, request.TimeRangeRequest.To, request.SortOrder, request.Limit, request.Offset)
	if err != nil {
		impl.logger.Errorw("error getting triggered CI pipelines", "err", err)
		return nil, 0, err
	}

	var pipelineUsage []bean.PipelineUsageItem
	for _, data := range pipelineData {
		pipelineUsage = append(pipelineUsage, bean.PipelineUsageItem{
			AppID:        data.AppID,
			PipelineID:   data.PipelineID,
			PipelineName: data.PipelineName,
			AppName:      data.AppName,
			TriggerCount: data.TriggerCount,
		})
	}

	return pipelineUsage, totalCount, nil
}

func (impl *InsightsServiceImpl) getTriggeredDeploymentPipelines(ctx context.Context, request *bean.InsightsRequest) ([]bean.PipelineUsageItem, int, error) {
	pipelineData, totalCount, err := impl.cdWorkflowRepository.GetTriggeredCDPipelines(request.TimeRangeRequest.From, request.TimeRangeRequest.To, request.SortOrder, request.Limit, request.Offset)
	if err != nil {
		impl.logger.Errorw("error getting triggered CD pipelines", "err", err)
		return nil, 0, err
	}

	var pipelineUsage []bean.PipelineUsageItem
	for _, data := range pipelineData {
		pipelineUsage = append(pipelineUsage, bean.PipelineUsageItem{
			AppID:        data.AppID,
			EnvID:        data.EnvID,
			PipelineID:   data.PipelineID,
			PipelineName: data.PipelineName,
			AppName:      data.AppName,
			EnvName:      data.EnvName,
			TriggerCount: data.TriggerCount,
		})
	}

	return pipelineUsage, totalCount, nil
}

// Approval policy coverage methods moved to ApprovalPolicyService
