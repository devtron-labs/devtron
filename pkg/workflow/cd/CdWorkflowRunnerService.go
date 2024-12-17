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

package cd

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"go.uber.org/zap"
)

type CdWorkflowRunnerService interface {
	UpdateWfr(dto *bean.CdWorkflowRunnerDto, updatedBy int) error
	UpdateIsArtifactUploaded(wfrId int, isArtifactUploaded bool) error
}

type CdWorkflowRunnerServiceImpl struct {
	logger               *zap.SugaredLogger
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewCdWorkflowRunnerServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *CdWorkflowRunnerServiceImpl {
	return &CdWorkflowRunnerServiceImpl{
		logger:               logger,
		cdWorkflowRepository: cdWorkflowRepository,
	}
}

func (impl *CdWorkflowRunnerServiceImpl) UpdateWfr(dto *bean.CdWorkflowRunnerDto, updatedBy int) error {
	runnerDbObj := adapter.ConvertCdWorkflowRunnerDtoToDbObj(dto)
	runnerDbObj.UpdateAuditLog(int32(updatedBy))
	err := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runnerDbObj)
	if err != nil {
		impl.logger.Errorw("error in updating runner status in db", "runnerId", runnerDbObj.Id, "err", err)
		return err
	}
	return nil
}

func (impl *CdWorkflowRunnerServiceImpl) UpdateIsArtifactUploaded(wfrId int, isArtifactUploaded bool) error {
	err := impl.cdWorkflowRepository.UpdateIsArtifactUploaded(wfrId, workflow.GetArtifactUploadedType(isArtifactUploaded))
	if err != nil {
		impl.logger.Errorw("error in updating isArtifactUploaded in db", "wfrId", wfrId, "err", err)
		return err
	}
	return nil
}
