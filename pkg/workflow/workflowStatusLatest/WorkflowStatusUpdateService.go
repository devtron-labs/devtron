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

package workflowStatusLatest

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowStatusUpdateService interface {
	// Methods to update latest status tables when workflow status changes
	SaveCiWorkflowStatusLatest(tx *pg.Tx, pipelineId, ciWorkflowId int, userId int32) error
	UpdateCdWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error
}

type WorkflowStatusUpdateServiceImpl struct {
	logger                      *zap.SugaredLogger
	workflowStatusLatestService WorkflowStatusLatestService
	ciWorkflowRepository        pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository        pipelineConfig.CdWorkflowRepository
	ciPipelineRepository        pipelineConfig.CiPipelineRepository
	pipelineRepository          pipelineConfig.PipelineRepository
}

func NewWorkflowStatusUpdateServiceImpl(
	logger *zap.SugaredLogger,
	workflowStatusLatestService WorkflowStatusLatestService,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
) *WorkflowStatusUpdateServiceImpl {
	return &WorkflowStatusUpdateServiceImpl{
		logger:                      logger,
		workflowStatusLatestService: workflowStatusLatestService,
		ciWorkflowRepository:        ciWorkflowRepository,
		cdWorkflowRepository:        cdWorkflowRepository,
		ciPipelineRepository:        ciPipelineRepository,
		pipelineRepository:          pipelineRepository,
	}
}

func (impl *WorkflowStatusUpdateServiceImpl) SaveCiWorkflowStatusLatest(tx *pg.Tx, pipelineId, ciWorkflowId int, userId int32) error {
	return impl.workflowStatusLatestService.SaveCiWorkflowStatusLatest(tx, pipelineId, ciWorkflowId, userId)
}

func (impl *WorkflowStatusUpdateServiceImpl) UpdateCdWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error {
	return impl.workflowStatusLatestService.SaveCdWorkflowStatusLatest(tx, pipelineId, appId, environmentId, workflowRunnerId, workflowType, userId)
}
