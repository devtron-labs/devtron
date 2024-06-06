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
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"go.uber.org/zap"
)

type CdWorkflowService interface {
	CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error)
	UpdateWorkFlow(dto *bean.CdWorkflowDto) error
}

type CdWorkflowServiceImpl struct {
	logger               *zap.SugaredLogger
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewCdWorkflowServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *CdWorkflowServiceImpl {
	return &CdWorkflowServiceImpl{
		logger:               logger,
		cdWorkflowRepository: cdWorkflowRepository,
	}
}

func (impl *CdWorkflowServiceImpl) CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error) {
	latest, err = impl.cdWorkflowRepository.IsLatestWf(pipelineId, cdWfId)
	if err != nil {
		impl.logger.Errorw("error in checking if wf is latest", "pipelineId", pipelineId, "cdWfId", cdWfId, "err", err)
		return false, err
	}
	return latest, nil
}

func (impl *CdWorkflowServiceImpl) UpdateWorkFlow(dto *bean.CdWorkflowDto) error {
	dbObj := adapter.ConvertCdWorkflowDtoToDbObj(dto)
	err := impl.cdWorkflowRepository.UpdateWorkFlow(dbObj)
	if err != nil {
		impl.logger.Errorw("error in updating workflow", "err", err, "req", dto)
		return err
	}
	return nil
}
