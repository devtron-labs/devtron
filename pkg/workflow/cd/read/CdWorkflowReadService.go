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
	bean4 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"go.uber.org/zap"
)

type CdWorkflowReadService interface {
	CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error)
	FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId int, environmentId int, runnerType bean4.WorkflowType) (pipelineConfig.CdWorkflowRunner, error)
	FindLatestCdWorkflowRunnerArtifactMetadataForAppAndEnvIds(appVsEnvIdMap map[int][]int, runnerType bean4.WorkflowType) ([]*cdWorkflow.CdWorkflowRunnerArtifactMetadata, error)
}

type CdWorkflowReadServiceImpl struct {
	logger               *zap.SugaredLogger
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewCdWorkflowReadServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *CdWorkflowReadServiceImpl {
	return &CdWorkflowReadServiceImpl{
		logger:               logger,
		cdWorkflowRepository: cdWorkflowRepository,
	}
}

func (impl *CdWorkflowReadServiceImpl) CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error) {
	latest, err = impl.cdWorkflowRepository.IsLatestWf(pipelineId, cdWfId)
	if err != nil {
		impl.logger.Errorw("error in checking if wf is latest", "pipelineId", pipelineId, "cdWfId", cdWfId, "err", err)
		return false, err
	}
	return latest, nil
}

func (impl *CdWorkflowReadServiceImpl) FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId int, environmentId int, runnerType bean4.WorkflowType) (pipelineConfig.CdWorkflowRunner, error) {
	return impl.cdWorkflowRepository.FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId, environmentId, runnerType)
}

func (impl *CdWorkflowReadServiceImpl) FindLatestCdWorkflowRunnerArtifactMetadataForAppAndEnvIds(appVsEnvIdMap map[int][]int, runnerType bean4.WorkflowType) ([]*cdWorkflow.CdWorkflowRunnerArtifactMetadata, error) {
	return impl.cdWorkflowRepository.FindLatestCdWorkflowRunnerArtifactMetadataForAppAndEnvIds(appVsEnvIdMap, runnerType)
}
