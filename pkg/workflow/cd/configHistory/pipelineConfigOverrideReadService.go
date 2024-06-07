/*
 * Copyright (c) 2024. Devtron Inc.
 */

package configHistory

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"go.uber.org/zap"
)

type PipelineConfigOverrideReadService interface {
	GetLastDeployedArtifactsInOrder(appId, envId, limit int) ([]int, error)
}

type PipelineConfigOverrideReadServiceImpl struct {
	logger                     *zap.SugaredLogger
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository
}

func NewPipelineConfigOverrideReadServiceImpl(logger *zap.SugaredLogger,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository) *PipelineConfigOverrideReadServiceImpl {
	return &PipelineConfigOverrideReadServiceImpl{
		logger:                     logger,
		pipelineOverrideRepository: pipelineOverrideRepository,
	}
}

func (impl *PipelineConfigOverrideReadServiceImpl) GetLastDeployedArtifactsInOrder(appId, envId, limit int) ([]int, error) {
	pipelineConfigOverrides, err := impl.pipelineOverrideRepository.FetchLatestNDeployedArtifacts(appId, envId, limit)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow runner by id", "appId", appId, "envId", envId, "limit", limit, "err", err)
		return nil, err
	}
	var ciArtifacts []int
	for _, pco := range pipelineConfigOverrides {
		ciArtifacts = append(ciArtifacts, pco.CiArtifactId)
	}
	return ciArtifacts, nil
}
