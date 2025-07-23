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

	util2 "github.com/devtron-labs/devtron/internal/util"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowStatusLatestService interface {
	// CI Workflow Status Latest methods
	SaveOrUpdateCiWorkflowStatusLatest(pipelineId, appId, ciWorkflowId int, userId int32) error
	GetCiWorkflowStatusLatestByPipelineId(pipelineId int) (*CiWorkflowStatusLatest, error)
	GetCiWorkflowStatusLatestByAppId(appId int) ([]*CiWorkflowStatusLatest, error)

	// CD Workflow Status Latest methods
	SaveOrUpdateCdWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error
	GetCdWorkflowStatusLatestByPipelineIdAndWorkflowType(pipelineId int, workflowType string) (*CdWorkflowStatusLatest, error)
	GetCdWorkflowStatusLatestByAppId(appId int) ([]*CdWorkflowStatusLatest, error)
	GetCdWorkflowStatusLatestByPipelineId(pipelineId int) ([]*CdWorkflowStatusLatest, error)
	GetCdWorkflowLatestByPipelineIds(pipelineIds []int) ([]*CdWorkflowStatusLatest, error)
}

type WorkflowStatusLatestServiceImpl struct {
	logger                         *zap.SugaredLogger
	workflowStatusLatestRepository pipelineConfig.WorkflowStatusLatestRepository
	ciWorkflowRepository           pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository           pipelineConfig.CdWorkflowRepository
}

func NewWorkflowStatusLatestServiceImpl(
	logger *zap.SugaredLogger,
	workflowStatusLatestRepository pipelineConfig.WorkflowStatusLatestRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
) *WorkflowStatusLatestServiceImpl {
	return &WorkflowStatusLatestServiceImpl{
		logger:                         logger,
		workflowStatusLatestRepository: workflowStatusLatestRepository,
		ciWorkflowRepository:           ciWorkflowRepository,
		cdWorkflowRepository:           cdWorkflowRepository,
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

// CI Workflow Status Latest methods implementation
func (impl *WorkflowStatusLatestServiceImpl) SaveOrUpdateCiWorkflowStatusLatest(pipelineId, appId, ciWorkflowId int, userId int32) error {
	// Validate required parameters
	if pipelineId <= 0 {
		impl.logger.Errorw("invalid pipelineId provided", "pipelineId", pipelineId)
		return fmt.Errorf("invalid pipelineId: %d", pipelineId)
	}

	if ciWorkflowId <= 0 {
		impl.logger.Errorw("invalid ciWorkflowId provided", "ciWorkflowId", ciWorkflowId)
		return fmt.Errorf("invalid ciWorkflowId: %d", ciWorkflowId)
	}

	// If appId is not provided (0), fetch it from the CiPipeline
	if appId <= 0 {
		ciPipeline, err := impl.ciWorkflowRepository.FindById(ciWorkflowId)
		if err != nil {
			impl.logger.Errorw("error in fetching ci workflow to get appId", "err", err, "ciWorkflowId", ciWorkflowId)
			return err
		}

		if ciPipeline == nil {
			impl.logger.Errorw("ci workflow not found", "ciWorkflowId", ciWorkflowId)
			return fmt.Errorf("ci workflow not found with id: %d", ciWorkflowId)
		}

		// Check if CiPipeline is loaded
		if ciPipeline.CiPipeline == nil {
			impl.logger.Errorw("ci pipeline not loaded in ci workflow", "ciWorkflowId", ciWorkflowId, "ciPipelineId", ciPipeline.CiPipelineId)
			return fmt.Errorf("ci pipeline not loaded for workflow id: %d", ciWorkflowId)
		}

		appId = ciPipeline.CiPipeline.AppId
		if appId <= 0 {
			impl.logger.Errorw("invalid appId in ci pipeline", "ciWorkflowId", ciWorkflowId, "ciPipelineId", ciPipeline.CiPipelineId, "appId", appId)
			return fmt.Errorf("invalid appId in ci pipeline: %d", appId)
		}

		impl.logger.Debugw("fetched appId from ci workflow", "ciWorkflowId", ciWorkflowId, "appId", appId)
	}

	// Check if entry exists
	existingEntry, err := impl.workflowStatusLatestRepository.GetCiWorkflowStatusLatestByPipelineId(pipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ci workflow status latest", "err", err, "pipelineId", pipelineId)
		return err
	}

	now := time.Now()
	if err == pg.ErrNoRows {
		// Create new entry
		model := &pipelineConfig.CiWorkflowStatusLatest{
			PipelineId:   pipelineId,
			AppId:        appId,
			CiWorkflowId: ciWorkflowId,
		}
		model.CreatedBy = userId
		model.CreatedOn = now
		model.UpdatedBy = userId
		model.UpdatedOn = now

		return impl.workflowStatusLatestRepository.SaveCiWorkflowStatusLatest(model)
	} else {
		// Update existing entry
		existingEntry.CiWorkflowId = ciWorkflowId
		existingEntry.UpdatedBy = userId
		existingEntry.UpdatedOn = now

		return impl.workflowStatusLatestRepository.UpdateCiWorkflowStatusLatest(existingEntry)
	}
}

func (impl *WorkflowStatusLatestServiceImpl) GetCiWorkflowStatusLatestByPipelineId(pipelineId int) (*CiWorkflowStatusLatest, error) {
	model, err := impl.workflowStatusLatestRepository.GetCiWorkflowStatusLatestByPipelineId(pipelineId)
	if err != nil {
		if err == pg.ErrNoRows {
			// Fallback to old method
			return impl.getCiWorkflowStatusFromOldMethod(pipelineId)
		}
		impl.logger.Errorw("error in getting ci workflow status latest", "err", err, "pipelineId", pipelineId)
		return nil, err
	}

	// Get status from ci_workflow table
	ciWorkflow, err := impl.ciWorkflowRepository.FindById(model.CiWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow", "err", err, "ciWorkflowId", model.CiWorkflowId)
		return nil, err
	}

	return &CiWorkflowStatusLatest{
		PipelineId:   model.PipelineId,
		AppId:        model.AppId,
		CiWorkflowId: model.CiWorkflowId,
		Status:       ciWorkflow.Status,
	}, nil
}

func (impl *WorkflowStatusLatestServiceImpl) GetCiWorkflowStatusLatestByAppId(appId int) ([]*CiWorkflowStatusLatest, error) {
	models, err := impl.workflowStatusLatestRepository.GetCiWorkflowStatusLatestByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow status latest by app id", "err", err, "appId", appId)
		return nil, err
	}

	var result []*CiWorkflowStatusLatest
	for _, model := range models {
		// Get status from ci_workflow table
		ciWorkflow, err := impl.ciWorkflowRepository.FindById(model.CiWorkflowId)
		if err != nil {
			impl.logger.Errorw("error in getting ci workflow", "err", err, "ciWorkflowId", model.CiWorkflowId)
			continue // Skip this entry if we can't get the workflow
		}

		result = append(result, &CiWorkflowStatusLatest{
			PipelineId:        model.PipelineId,
			AppId:             model.AppId,
			CiWorkflowId:      model.CiWorkflowId,
			Status:            ciWorkflow.Status,
			StorageConfigured: ciWorkflow.BlobStorageEnabled,
		})
	}

	return result, nil
}

// CD Workflow Status Latest methods implementation
func (impl *WorkflowStatusLatestServiceImpl) SaveOrUpdateCdWorkflowStatusLatest(tx *pg.Tx, pipelineId, appId, environmentId, workflowRunnerId int, workflowType string, userId int32) error {
	// Validate required parameters
	if pipelineId <= 0 {
		impl.logger.Errorw("invalid pipelineId provided", "pipelineId", pipelineId)
		return fmt.Errorf("invalid pipelineId: %d", pipelineId)
	}

	if workflowRunnerId <= 0 {
		impl.logger.Errorw("invalid workflowRunnerId provided", "workflowRunnerId", workflowRunnerId)
		return fmt.Errorf("invalid workflowRunnerId: %d", workflowRunnerId)
	}

	// If appId or environmentId is not provided (0), fetch them from the CdWorkflow
	if appId <= 0 || environmentId <= 0 {
		cdWorkflowRunner, err := impl.cdWorkflowRepository.FindBasicWorkflowRunnerById(workflowRunnerId)
		if err != nil {
			impl.logger.Errorw("error in fetching cd workflow runner to get appId/environmentId", "err", err, "workflowRunnerId", workflowRunnerId)
			return err
		}

		if cdWorkflowRunner == nil {
			impl.logger.Errorw("cd workflow runner not found", "workflowRunnerId", workflowRunnerId)
			return fmt.Errorf("cd workflow runner not found with id: %d", workflowRunnerId)
		}

		// Check if CdWorkflow is loaded
		if cdWorkflowRunner.CdWorkflow == nil {
			impl.logger.Errorw("cd workflow not loaded in cd workflow runner", "workflowRunnerId", workflowRunnerId, "cdWorkflowId", cdWorkflowRunner.CdWorkflowId)
			return fmt.Errorf("cd workflow not loaded for workflow runner id: %d", workflowRunnerId)
		}

		// Check if Pipeline is loaded
		if cdWorkflowRunner.CdWorkflow.Pipeline == nil {
			impl.logger.Errorw("pipeline not loaded in cd workflow", "workflowRunnerId", workflowRunnerId, "cdWorkflowId", cdWorkflowRunner.CdWorkflowId, "pipelineId", cdWorkflowRunner.CdWorkflow.PipelineId)
			return fmt.Errorf("pipeline not loaded for workflow runner id: %d", workflowRunnerId)
		}

		if appId <= 0 {
			appId = cdWorkflowRunner.CdWorkflow.Pipeline.AppId
			if appId <= 0 {
				impl.logger.Errorw("invalid appId in pipeline", "workflowRunnerId", workflowRunnerId, "pipelineId", cdWorkflowRunner.CdWorkflow.PipelineId, "appId", appId)
				return fmt.Errorf("invalid appId in pipeline: %d", appId)
			}
			impl.logger.Debugw("fetched appId from cd workflow runner", "workflowRunnerId", workflowRunnerId, "appId", appId)
		}

		if environmentId <= 0 {
			environmentId = cdWorkflowRunner.CdWorkflow.Pipeline.EnvironmentId
			if environmentId <= 0 {
				impl.logger.Errorw("invalid environmentId in pipeline", "workflowRunnerId", workflowRunnerId, "pipelineId", cdWorkflowRunner.CdWorkflow.PipelineId, "environmentId", environmentId)
				return fmt.Errorf("invalid environmentId in pipeline: %d", environmentId)
			}
			impl.logger.Debugw("fetched environmentId from cd workflow runner", "workflowRunnerId", workflowRunnerId, "environmentId", environmentId)
		}
	}

	// Check if entry exists
	existingEntry, err := impl.workflowStatusLatestRepository.GetCdWorkflowStatusLatestByPipelineIdAndWorkflowType(tx, pipelineId, workflowType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting cd workflow status latest", "err", err, "pipelineId", pipelineId, "workflowType", workflowType)
		return err
	}

	now := time.Now()
	if err == pg.ErrNoRows {
		// Create new entry
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
	} else {
		// Update existing entry
		existingEntry.WorkflowRunnerId = workflowRunnerId
		existingEntry.UpdatedBy = userId
		existingEntry.UpdatedOn = now

		return impl.workflowStatusLatestRepository.UpdateCdWorkflowStatusLatest(tx, existingEntry)
	}
}

func (impl *WorkflowStatusLatestServiceImpl) GetCdWorkflowStatusLatestByPipelineIdAndWorkflowType(pipelineId int, workflowType string) (*CdWorkflowStatusLatest, error) {
	model, err := impl.workflowStatusLatestRepository.GetCdWorkflowStatusLatestByPipelineIdAndWorkflowType(nil, pipelineId, workflowType)
	if err != nil {
		if err == pg.ErrNoRows {
			// Fallback to old method
			return impl.getCdWorkflowStatusFromOldMethod(pipelineId, workflowType)
		}
		impl.logger.Errorw("error in getting cd workflow status latest", "err", err, "pipelineId", pipelineId, "workflowType", workflowType)
		return nil, err
	}

	// Get status from cd_workflow_runner table
	cdWorkflowRunner, err := impl.cdWorkflowRepository.FindBasicWorkflowRunnerById(model.WorkflowRunnerId)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow runner", "err", err, "workflowRunnerId", model.WorkflowRunnerId)
		return nil, err
	}

	return &CdWorkflowStatusLatest{
		PipelineId:       model.PipelineId,
		AppId:            model.AppId,
		EnvironmentId:    model.EnvironmentId,
		WorkflowType:     model.WorkflowType,
		WorkflowRunnerId: model.WorkflowRunnerId,
		Status:           cdWorkflowRunner.Status,
	}, nil
}

func (impl *WorkflowStatusLatestServiceImpl) GetCdWorkflowStatusLatestByAppId(appId int) ([]*CdWorkflowStatusLatest, error) {
	models, err := impl.workflowStatusLatestRepository.GetCdWorkflowStatusLatestByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by app id", "err", err, "appId", appId)
		return nil, err
	}

	var result []*CdWorkflowStatusLatest
	for _, model := range models {
		// Get status from cd_workflow_runner table
		cdWorkflowRunner, err := impl.cdWorkflowRepository.FindBasicWorkflowRunnerById(model.WorkflowRunnerId)
		if err != nil {
			impl.logger.Errorw("error in getting cd workflow runner", "err", err, "workflowRunnerId", model.WorkflowRunnerId)
			continue // Skip this entry if we can't get the workflow runner
		}

		result = append(result, &CdWorkflowStatusLatest{
			PipelineId:       model.PipelineId,
			AppId:            model.AppId,
			EnvironmentId:    model.EnvironmentId,
			WorkflowType:     model.WorkflowType,
			WorkflowRunnerId: model.WorkflowRunnerId,
			Status:           cdWorkflowRunner.Status,
		})
	}

	return result, nil
}

func (impl *WorkflowStatusLatestServiceImpl) GetCdWorkflowStatusLatestByPipelineId(pipelineId int) ([]*CdWorkflowStatusLatest, error) {
	models, err := impl.workflowStatusLatestRepository.GetCdWorkflowStatusLatestByPipelineId(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by pipeline id", "err", err, "pipelineId", pipelineId)
		return nil, err
	}

	var result []*CdWorkflowStatusLatest
	for _, model := range models {
		// Get status from cd_workflow_runner table
		cdWorkflowRunner, err := impl.cdWorkflowRepository.FindBasicWorkflowRunnerById(model.WorkflowRunnerId)
		if err != nil {
			impl.logger.Errorw("error in getting cd workflow runner", "err", err, "workflowRunnerId", model.WorkflowRunnerId)
			continue // Skip this entry if we can't get the workflow runner
		}

		result = append(result, &CdWorkflowStatusLatest{
			PipelineId:       model.PipelineId,
			AppId:            model.AppId,
			EnvironmentId:    model.EnvironmentId,
			WorkflowType:     model.WorkflowType,
			WorkflowRunnerId: model.WorkflowRunnerId,
			Status:           cdWorkflowRunner.Status,
		})
	}

	return result, nil
}

// Fallback methods to old implementation when no entry found in latest status tables
func (impl *WorkflowStatusLatestServiceImpl) getCiWorkflowStatusFromOldMethod(pipelineId int) (*CiWorkflowStatusLatest, error) {
	// Get the latest CI workflow for this pipeline using the old method
	workflow, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflow(pipelineId)
	if err != nil {
		if util2.IsErrNoRows(err) {
			return &CiWorkflowStatusLatest{
				PipelineId:   pipelineId,
				AppId:        0, // Will need to be populated from pipeline info
				CiWorkflowId: 0,
				Status:       "Not Triggered",
			}, nil
		}
		impl.logger.Errorw("error in getting last triggered workflow", "err", err, "pipelineId", pipelineId)
		return nil, err
	}

	return &CiWorkflowStatusLatest{
		PipelineId:   pipelineId,
		AppId:        workflow.CiPipeline.AppId,
		CiWorkflowId: workflow.Id,
		Status:       workflow.Status,
	}, nil
}

func (impl *WorkflowStatusLatestServiceImpl) getCdWorkflowStatusFromOldMethod(pipelineId int, workflowType string) (*CdWorkflowStatusLatest, error) {
	// Convert workflowType to the appropriate enum
	var runnerType bean.WorkflowType
	switch workflowType {
	case "PRE":
		runnerType = bean.CD_WORKFLOW_TYPE_PRE
	case "DEPLOY":
		runnerType = bean.CD_WORKFLOW_TYPE_DEPLOY
	case "POST":
		runnerType = bean.CD_WORKFLOW_TYPE_POST
	default:
		runnerType = bean.WorkflowType(workflowType)
	}

	// Get the latest CD workflow runner for this pipeline and type using the old method
	wfr, err := impl.cdWorkflowRepository.FindLatestByPipelineIdAndRunnerType(pipelineId, runnerType)
	if err != nil {
		if err == pg.ErrNoRows {
			return &CdWorkflowStatusLatest{
				PipelineId:       pipelineId,
				AppId:            0, // Will need to be populated from pipeline info
				EnvironmentId:    0,
				WorkflowType:     workflowType,
				WorkflowRunnerId: 0,
				Status:           "Not Triggered",
			}, nil
		}
		impl.logger.Errorw("error in getting latest cd workflow runner", "err", err, "pipelineId", pipelineId, "runnerType", runnerType)
		return nil, err
	}

	return &CdWorkflowStatusLatest{
		PipelineId:       pipelineId,
		AppId:            wfr.CdWorkflow.Pipeline.AppId,
		EnvironmentId:    wfr.CdWorkflow.Pipeline.EnvironmentId,
		WorkflowType:     workflowType,
		WorkflowRunnerId: wfr.Id,
		Status:           wfr.Status,
	}, nil
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
