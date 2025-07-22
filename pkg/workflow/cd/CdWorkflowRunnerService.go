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
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	bean4 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/status/workflowStatusLatest"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CdWorkflowRunnerService interface {
	UpdateWfr(dto *bean.CdWorkflowRunnerDto, updatedBy int) error
	SaveWfr(tx *pg.Tx, wfr *pipelineConfig.CdWorkflowRunner) error
	UpdateIsArtifactUploaded(wfrId int, isArtifactUploaded bool) error
	SaveCDWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) (*pipelineConfig.CdWorkflowRunner, error)
	UpdateCdWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) error
	GetPrePostWorkflowStagesByWorkflowRunnerIdsList(wfIdWfTypeMap map[int]bean4.CdWorkflowWithArtifact) (map[int]map[string][]*bean3.WorkflowStageDto, error)
}

type CdWorkflowRunnerServiceImpl struct {
	logger                      *zap.SugaredLogger
	cdWorkflowRepository        pipelineConfig.CdWorkflowRepository
	workflowStageService        workflowStatus.WorkFlowStageStatusService
	transactionManager          sql.TransactionWrapper
	config                      *types.CiConfig
	workflowStatusUpdateService workflowStatusLatest.WorkflowStatusUpdateService
}

func NewCdWorkflowRunnerServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	workflowStageService workflowStatus.WorkFlowStageStatusService,
	transactionManager sql.TransactionWrapper,
	workflowStatusUpdateService workflowStatusLatest.WorkflowStatusUpdateService) *CdWorkflowRunnerServiceImpl {
	impl := &CdWorkflowRunnerServiceImpl{
		logger:                      logger,
		cdWorkflowRepository:        cdWorkflowRepository,
		workflowStageService:        workflowStageService,
		transactionManager:          transactionManager,
		workflowStatusUpdateService: workflowStatusUpdateService,
	}
	ciConfig, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	impl.config = ciConfig
	return impl
}

func (impl *CdWorkflowRunnerServiceImpl) UpdateWfr(dto *bean.CdWorkflowRunnerDto, updatedBy int) error {
	runnerDbObj := adapter.ConvertCdWorkflowRunnerDtoToDbObj(dto)
	runnerDbObj.UpdateAuditLog(int32(updatedBy))
	err := impl.UpdateCdWorkflowRunnerWithStage(runnerDbObj)
	if err != nil {
		impl.logger.Errorw("error in updating runner status in db", "runnerId", runnerDbObj.Id, "err", err)
		return err
	}

	// Update latest status table for CD workflow
	err = impl.workflowStatusUpdateService.UpdateCdWorkflowStatusLatest(runnerDbObj.CdWorkflow.PipelineId, runnerDbObj.CdWorkflow.Pipeline.AppId, runnerDbObj.CdWorkflow.Pipeline.EnvironmentId, runnerDbObj.Id, runnerDbObj.WorkflowType.String(), int32(updatedBy))
	if err != nil {
		impl.logger.Errorw("error in updating cd workflow status latest", "err", err, "pipelineId", runnerDbObj.CdWorkflow.PipelineId, "workflowRunnerId", runnerDbObj.Id)
		// Don't return error here as the main workflow update was successful
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

func (impl *CdWorkflowRunnerServiceImpl) SaveCDWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) (*pipelineConfig.CdWorkflowRunner, error) {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "workflowName", wfr.Name, "error", err)
		return wfr, err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil && dbErr.Error() != util.SqlAlreadyCommitedErrMsg {
			impl.logger.Errorw("error in rolling back transaction", "workflowName", wfr.Name, "error", dbErr)
		}
	}()
	if impl.config.EnableWorkflowExecutionStage {
		wfr.Status = cdWorkflow.WorkflowWaitingToStart
	}
	err = impl.cdWorkflowRepository.SaveWorkFlowRunnerWithTx(wfr, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wfr, "error", err)
		return wfr, err
	}

	err = impl.workflowStageService.SaveWorkflowStages(wfr.Id, wfr.WorkflowType.String(), wfr.Name, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow stages", "workflowName", wfr.Name, "error", err)
		return wfr, err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "workflowName", wfr.Name, "error", err)
		return wfr, err
	}

	// Update latest status table for CD workflow
	err = impl.workflowStatusUpdateService.UpdateCdWorkflowStatusLatest(wfr.CdWorkflow.PipelineId, wfr.CdWorkflow.Pipeline.AppId, wfr.CdWorkflow.Pipeline.EnvironmentId, wfr.Id, wfr.WorkflowType.String(), wfr.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in updating cd workflow status latest", "err", err, "pipelineId", wfr.CdWorkflow.PipelineId, "workflowRunnerId", wfr.Id)
		// Don't return error here as the main workflow save was successful
	}

	return wfr, nil
}

func (impl *CdWorkflowRunnerServiceImpl) UpdateCdWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) error {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "workflowName", wfr.Name, "error", err)
		return err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil && dbErr.Error() != util.SqlAlreadyCommitedErrMsg {
			impl.logger.Errorw("error in rolling back transaction", "workflowName", wfr.Name, "error", dbErr)
		}
	}()
	if wfr.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE || wfr.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST {
		wfr.Status, wfr.PodStatus, err = impl.workflowStageService.UpdateWorkflowStages(wfr.Id, wfr.WorkflowType.String(), wfr.Name, wfr.Status, wfr.PodStatus, wfr.Message, wfr.PodName, tx)
		if err != nil {
			impl.logger.Errorw("error in updating workflow stages", "workflowName", wfr.Name, "error", err)
			return err
		}
	}

	//update workflow runner now with updatedWfStatus if applicable
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunnerWithTx(wfr, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wfr, "error", err)
		return err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "workflowName", wfr.Name, "error", err)
		return err
	}

	// Update latest status table for CD workflow
	err = impl.workflowStatusUpdateService.UpdateCdWorkflowStatusLatest(wfr.CdWorkflow.PipelineId, wfr.CdWorkflow.Pipeline.AppId, wfr.CdWorkflow.Pipeline.EnvironmentId, wfr.Id, wfr.WorkflowType.String(), wfr.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in updating cd workflow status latest", "err", err, "pipelineId", wfr.CdWorkflow.PipelineId, "workflowRunnerId", wfr.Id)
		// Don't return error here as the main workflow update was successful
	}

	return nil

}

func (impl *CdWorkflowRunnerServiceImpl) GetPrePostWorkflowStagesByWorkflowRunnerIdsList(wfIdWfTypeMap map[int]bean4.CdWorkflowWithArtifact) (map[int]map[string][]*bean3.WorkflowStageDto, error) {
	// implementation
	resp := map[int]map[string][]*bean3.WorkflowStageDto{}
	if len(wfIdWfTypeMap) == 0 {
		return resp, nil
	}
	//first create a map of pre-runner ids and post-runner ids
	prePostRunnerIds := map[string][]int{}
	for wfId, wf := range wfIdWfTypeMap {
		if wf.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE.String() {
			prePostRunnerIds[bean2.CD_WORKFLOW_TYPE_PRE.String()] = append(prePostRunnerIds[bean2.CD_WORKFLOW_TYPE_PRE.String()], wfId)
		} else if wf.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST.String() {
			prePostRunnerIds[bean2.CD_WORKFLOW_TYPE_POST.String()] = append(prePostRunnerIds[bean2.CD_WORKFLOW_TYPE_POST.String()], wfId)
		}
	}

	preCdDbData, err := impl.workflowStageService.GetWorkflowStagesByWorkflowIdsAndWfType(prePostRunnerIds[bean2.CD_WORKFLOW_TYPE_PRE.String()], bean2.CD_WORKFLOW_TYPE_PRE.String())
	if err != nil {
		impl.logger.Errorw("error in getting pre-ci workflow stages", "error", err)
		return resp, err
	}
	//do the above for post cd
	postCdDbData, err := impl.workflowStageService.GetWorkflowStagesByWorkflowIdsAndWfType(prePostRunnerIds[bean2.CD_WORKFLOW_TYPE_POST.String()], bean2.CD_WORKFLOW_TYPE_POST.String())
	if err != nil {
		impl.logger.Errorw("error in getting post-ci workflow stages", "error", err)
		return resp, err
	}
	//iterate over prePostRunnerIds and create response structure using ConvertDBWorkflowStageToMap function
	for wfId, wf := range wfIdWfTypeMap {
		if wf.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE.String() {
			resp[wfId] = impl.workflowStageService.ConvertDBWorkflowStageToMap(preCdDbData, wfId, wf.Status, wf.PodStatus, wf.Message, wf.WorkflowType, wf.StartedOn, wf.FinishedOn)
		} else if wf.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST.String() {
			resp[wfId] = impl.workflowStageService.ConvertDBWorkflowStageToMap(postCdDbData, wfId, wf.Status, wf.PodStatus, wf.Message, wf.WorkflowType, wf.StartedOn, wf.FinishedOn)
		}
	}
	return resp, nil
}

func (impl *CdWorkflowRunnerServiceImpl) SaveWfr(tx *pg.Tx, wfr *pipelineConfig.CdWorkflowRunner) error {
	return impl.cdWorkflowRepository.SaveWorkFlowRunnerWithTx(wfr, tx)
}
