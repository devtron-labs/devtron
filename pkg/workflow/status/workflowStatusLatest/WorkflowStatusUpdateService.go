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
	"go.uber.org/zap"
)

type WorkflowStatusUpdateService interface {
	// Methods to update latest status tables when workflow status changes
	UpdateCiWorkflowStatusLatest(pipelineId, appId, ciWorkflowId int, userId int32) error
	UpdateCdWorkflowStatusLatest(pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error

	// Methods to fetch optimized status for trigger view
	FetchCiStatusForTriggerViewOptimized(appId int) ([]*pipelineConfig.CiWorkflowStatus, error)
	FetchCdStatusForTriggerViewOptimized(appId int) ([]*pipelineConfig.CdWorkflowStatus, error)
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

func (impl *WorkflowStatusUpdateServiceImpl) UpdateCiWorkflowStatusLatest(pipelineId, appId, ciWorkflowId int, userId int32) error {
	return impl.workflowStatusLatestService.SaveOrUpdateCiWorkflowStatusLatest(pipelineId, appId, ciWorkflowId, userId)
}

func (impl *WorkflowStatusUpdateServiceImpl) UpdateCdWorkflowStatusLatest(pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error {
	return impl.workflowStatusLatestService.SaveOrUpdateCdWorkflowStatusLatest(pipelineId, appId, environmentId, workflowRunnerId, workflowType, userId)
}

func (impl *WorkflowStatusUpdateServiceImpl) FetchCiStatusForTriggerViewOptimized(appId int) ([]*pipelineConfig.CiWorkflowStatus, error) {
	latestStatuses, err := impl.workflowStatusLatestService.GetCiWorkflowStatusLatestByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow status latest by app id", "err", err, "appId", appId)
		// Fallback to old method
		return impl.ciWorkflowRepository.FIndCiWorkflowStatusesByAppId(appId)
	}

	// Convert to the expected format
	var ciWorkflowStatuses []*pipelineConfig.CiWorkflowStatus
	for _, latestStatus := range latestStatuses {
		ciPipeline, err := impl.ciPipelineRepository.FindById(latestStatus.PipelineId)
		if err != nil {
			impl.logger.Errorw("error in getting ci pipeline", "err", err, "pipelineId", latestStatus.PipelineId)
			continue
		}

		ciWorkflowStatus := &pipelineConfig.CiWorkflowStatus{
			CiPipelineId:      latestStatus.PipelineId,
			CiPipelineName:    ciPipeline.Name,
			CiStatus:          latestStatus.Status,
			CiWorkflowId:      latestStatus.CiWorkflowId,
			StorageConfigured: latestStatus.StorageConfigured,
		}
		ciWorkflowStatuses = append(ciWorkflowStatuses, ciWorkflowStatus)
	}

	// If no entries found in latest status table, fallback to old method
	if len(ciWorkflowStatuses) == 0 {
		impl.logger.Infow("no entries found in ci workflow status latest table, falling back to old method", "appId", appId)
		return impl.ciWorkflowRepository.FIndCiWorkflowStatusesByAppId(appId)
	}

	return ciWorkflowStatuses, nil
}

func (impl *WorkflowStatusUpdateServiceImpl) FetchCdStatusForTriggerViewOptimized(appId int) ([]*pipelineConfig.CdWorkflowStatus, error) {
	// First try to get from the optimized latest status table
	latestStatuses, err := impl.workflowStatusLatestService.GetCdWorkflowStatusLatestByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by app id", "err", err, "appId", appId)
		// Fallback to old method - would need to implement this based on existing CD status fetching logic
		return nil, err
	}

	// Convert to the expected format
	var cdWorkflowStatuses []*pipelineConfig.CdWorkflowStatus
	for _, latestStatus := range latestStatuses {
		// Get pipeline info
		pipeline, err := impl.pipelineRepository.FindById(latestStatus.PipelineId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline", "err", err, "pipelineId", latestStatus.PipelineId)
			continue
		}

		var status string
		switch latestStatus.WorkflowType {
		case "PRE":
			status = latestStatus.Status
		case "DEPLOY":
			status = latestStatus.Status
		case "POST":
			status = latestStatus.Status
		default:
			status = latestStatus.Status
		}

		cdWorkflowStatus := &pipelineConfig.CdWorkflowStatus{
			CiPipelineId: pipeline.CiPipelineId,
			PipelineId:   latestStatus.PipelineId,
			PipelineName: pipeline.Name,
			WorkflowType: latestStatus.WorkflowType,
			WfrId:        latestStatus.WorkflowRunnerId,
		}

		// Set the appropriate status field based on workflow type
		switch latestStatus.WorkflowType {
		case "PRE":
			cdWorkflowStatus.PreStatus = status
		case "DEPLOY":
			cdWorkflowStatus.DeployStatus = status
		case "POST":
			cdWorkflowStatus.PostStatus = status
		}

		cdWorkflowStatuses = append(cdWorkflowStatuses, cdWorkflowStatus)
	}

	return cdWorkflowStatuses, nil
}
