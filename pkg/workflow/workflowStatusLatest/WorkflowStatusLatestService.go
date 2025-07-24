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
	"fmt"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowStatusLatestService interface {
	// CI Workflow Status Latest methods
	SaveCiWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, ciWorkflowId int, userId int32) error
	GetCiWorkflowStatusLatestByPipelineIds(pipelineIds []int) ([]*pipelineConfig.CiWorkflowStatusLatest, error)

	// CD Workflow Status Latest methods
	SaveCdWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error
	GetCdWorkflowLatestByPipelineIds(pipelineIds []int) ([]*CdWorkflowStatusLatest, error)
}

type WorkflowStatusLatestServiceImpl struct {
	logger                         *zap.SugaredLogger
	workflowStatusLatestRepository pipelineConfig.WorkflowStatusLatestRepository
	ciWorkflowRepository           pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository           pipelineConfig.CdWorkflowRepository
	ciPipelineRepository           pipelineConfig.CiPipelineRepository
}

func NewWorkflowStatusLatestServiceImpl(
	logger *zap.SugaredLogger,
	workflowStatusLatestRepository pipelineConfig.WorkflowStatusLatestRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
) *WorkflowStatusLatestServiceImpl {
	return &WorkflowStatusLatestServiceImpl{
		logger:                         logger,
		workflowStatusLatestRepository: workflowStatusLatestRepository,
		ciWorkflowRepository:           ciWorkflowRepository,
		cdWorkflowRepository:           cdWorkflowRepository,
		ciPipelineRepository:           ciPipelineRepository,
	}
}

type CiWorkflowStatusLatest struct {
	PipelineId        int    `json:"pipelineId"`
	AppId             int    `json:"appId"`
	CiWorkflowId      int    `json:"ciWorkflowId"`
	Status            string `json:"status"` // Derived from ci_workflow table
	StorageConfigured bool   `json:"storageConfigured"`
}

type CdWorkflowStatusLatest struct {
	PipelineId       int    `json:"pipelineId"`
	AppId            int    `json:"appId"`
	EnvironmentId    int    `json:"environmentId"`
	WorkflowType     string `json:"workflowType"`
	WorkflowRunnerId int    `json:"workflowRunnerId"`
	Status           string `json:"status"` // Derived from cd_workflow_runner table
}

func (impl *WorkflowStatusLatestServiceImpl) SaveCiWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, ciWorkflowId int, userId int32) error {
	// Validate required parameters
	if ciWorkflowId <= 0 {
		impl.logger.Errorw("invalid ciWorkflowId provided", "ciWorkflowId", ciWorkflowId)
		return fmt.Errorf("invalid ciWorkflowId: %d", ciWorkflowId)
	}

	now := time.Now()
	model := &pipelineConfig.CiWorkflowStatusLatest{
		PipelineId:   pipelineId,
		AppId:        appId,
		CiWorkflowId: ciWorkflowId,
	}
	model.CreatedBy = userId
	model.CreatedOn = now
	model.UpdatedBy = userId
	model.UpdatedOn = now

	return impl.workflowStatusLatestRepository.SaveCiWorkflowStatusLatest(tx, model)
}

func (impl *WorkflowStatusLatestServiceImpl) GetCiWorkflowStatusLatestByPipelineIds(pipelineIds []int) ([]*pipelineConfig.CiWorkflowStatusLatest, error) {
	return impl.workflowStatusLatestRepository.GetCiWorkflowStatusLatestByPipelineIds(pipelineIds)
}

// CD Workflow Status Latest methods implementation
func (impl *WorkflowStatusLatestServiceImpl) SaveCdWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error {
	// Validate required parameters
	if workflowRunnerId <= 0 {
		impl.logger.Errorw("invalid workflowRunnerId provided", "workflowRunnerId", workflowRunnerId)
		return fmt.Errorf("invalid workflowRunnerId: %d", workflowRunnerId)
	}

	// Create new entry (always save, don't update)
	now := time.Now()
	model := &pipelineConfig.CdWorkflowStatusLatest{
		PipelineId:       pipelineId,
		AppId:            appId,
		EnvironmentId:    environmentId,
		WorkflowType:     workflowType,
		WorkflowRunnerId: workflowRunnerId,
	}
	model.CreatedBy = userId
	model.CreatedOn = now
	model.UpdatedBy = userId
	model.UpdatedOn = now

	return impl.workflowStatusLatestRepository.SaveCdWorkflowStatusLatest(tx, model)
}

func (impl *WorkflowStatusLatestServiceImpl) GetCdWorkflowLatestByPipelineIds(pipelineIds []int) ([]*CdWorkflowStatusLatest, error) {
	cdWorkflowStatusLatest, err := impl.workflowStatusLatestRepository.GetCdWorkflowStatusLatestByPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by pipeline ids", "pipelineIds", pipelineIds, "err", err)
		return nil, err
	}
	var result []*CdWorkflowStatusLatest
	for _, model := range cdWorkflowStatusLatest {
		result = append(result, &CdWorkflowStatusLatest{
			PipelineId:       model.PipelineId,
			AppId:            model.AppId,
			EnvironmentId:    model.EnvironmentId,
			WorkflowType:     model.WorkflowType,
			WorkflowRunnerId: model.WorkflowRunnerId,
		})
	}
	return result, nil
}
