/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package pipeline

import (
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	bean5 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	buildBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/workflowStatusLatest"
	util3 "github.com/devtron-labs/devtron/util"
	util2 "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
)

type CiService interface {
	WriteCITriggerEvent(trigger *types.CiTriggerRequest, workflowRequest *types.WorkflowRequest)
	WriteCIFailEvent(ciWorkflow *pipelineConfig.CiWorkflow)
	SaveCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error
	UpdateCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error
}

type CiServiceImpl struct {
	Logger                      *zap.SugaredLogger
	workflowStageStatusService  workflowStatus.WorkFlowStageStatusService
	eventClient                 client.EventClient
	eventFactory                client.EventFactory
	config                      *types.CiConfig
	ciWorkflowRepository        pipelineConfig.CiWorkflowRepository
	transactionManager          sql.TransactionWrapper
	workflowStatusUpdateService workflowStatusLatest.WorkflowStatusUpdateService
}

func NewCiServiceImpl(Logger *zap.SugaredLogger,
	workflowStageStatusService workflowStatus.WorkFlowStageStatusService, eventClient client.EventClient,
	eventFactory client.EventFactory,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	transactionManager sql.TransactionWrapper,
	workflowStatusUpdateService workflowStatusLatest.WorkflowStatusUpdateService,
) *CiServiceImpl {
	cis := &CiServiceImpl{
		Logger:                      Logger,
		workflowStageStatusService:  workflowStageStatusService,
		eventClient:                 eventClient,
		eventFactory:                eventFactory,
		ciWorkflowRepository:        ciWorkflowRepository,
		transactionManager:          transactionManager,
		workflowStatusUpdateService: workflowStatusUpdateService,
	}
	config, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	cis.config = config
	return cis
}

func (impl *CiServiceImpl) WriteCITriggerEvent(trigger *types.CiTriggerRequest, workflowRequest *types.WorkflowRequest) {
	impl.Logger.Debugw("triggering notification for ci trigger event", "pipelineId", workflowRequest.PipelineId, "ciWorkflowId", workflowRequest.WorkflowId)
	event, _ := impl.eventFactory.Build(util2.Trigger, &workflowRequest.PipelineId, workflowRequest.AppId, nil, util2.CI)
	material := &buildBean.MaterialTriggerInfo{}

	material.GitTriggers = trigger.CommitHashes

	event.UserId = int(trigger.TriggeredBy)
	event.CiWorkflowRunnerId = workflowRequest.WorkflowId
	event = impl.eventFactory.BuildExtraCIData(event, material)

	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.Logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *CiServiceImpl) WriteCIFailEvent(ciWorkflow *pipelineConfig.CiWorkflow) {
	event, _ := impl.eventFactory.Build(util2.Fail, &ciWorkflow.CiPipelineId, ciWorkflow.CiPipeline.AppId, nil, util2.CI)
	material := &buildBean.MaterialTriggerInfo{}
	material.GitTriggers = ciWorkflow.GitTriggers
	event.CiWorkflowRunnerId = ciWorkflow.Id
	event.UserId = int(ciWorkflow.TriggeredBy)
	event = impl.eventFactory.BuildExtraCIData(event, material)
	event.CiArtifactId = 0
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.Logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *CiServiceImpl) SaveCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.Logger.Errorw("error in starting transaction to save default configurations", "workflowName", wf.Name, "error", err)
		return err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil && dbErr.Error() != util3.SqlAlreadyCommitedErrMsg {
			impl.Logger.Errorw("error in rolling back transaction", "workflowName", wf.Name, "error", dbErr)
		}
	}()
	if impl.config.EnableWorkflowExecutionStage {
		wf.Status = cdWorkflow.WorkflowWaitingToStart
		wf.PodStatus = string(v1alpha1.NodePending)
	}
	err = impl.ciWorkflowRepository.SaveWorkFlowWithTx(wf, tx)
	if err != nil {
		impl.Logger.Errorw("error in saving workflow", "payload", wf, "error", err)
		return err
	}

	err = impl.workflowStageStatusService.SaveWorkflowStages(wf.Id, bean5.CI_WORKFLOW_TYPE.String(), wf.Name, tx)
	if err != nil {
		impl.Logger.Errorw("error in saving workflow stages", "workflowName", wf.Name, "error", err)
		return err
	}

	// Update latest status table for CI workflow within the transaction
	// Check if CiPipeline is loaded, if not pass 0 as appId to let the function fetch it
	appId := 0
	if wf.CiPipeline != nil {
		appId = wf.CiPipeline.AppId
	}
	err = impl.workflowStatusUpdateService.UpdateCiWorkflowStatusLatest(tx, wf.CiPipelineId, appId, wf.Id, wf.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("error in updating ci workflow status latest", "err", err, "pipelineId", wf.CiPipelineId, "workflowId", wf.Id)
		return err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.Logger.Errorw("error in committing transaction", "workflowName", wf.Name, "error", err)
		return err
	}

	return nil

}

func (impl *CiServiceImpl) UpdateCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.Logger.Errorw("error in starting transaction to save default configurations", "workflowName", wf.Name, "error", err)
		return err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil && dbErr.Error() != util3.SqlAlreadyCommitedErrMsg {
			impl.Logger.Errorw("error in rolling back transaction", "workflowName", wf.Name, "error", dbErr)
		}
	}()

	wf.Status, wf.PodStatus, err = impl.workflowStageStatusService.UpdateWorkflowStages(wf.Id, bean5.CI_WORKFLOW_TYPE.String(), wf.Name, wf.Status, wf.PodStatus, wf.Message, wf.PodName, tx)
	if err != nil {
		impl.Logger.Errorw("error in updating workflow stages", "workflowName", wf.Name, "error", err)
		return err
	}

	err = impl.ciWorkflowRepository.UpdateWorkFlowWithTx(wf, tx)
	if err != nil {
		impl.Logger.Errorw("error in saving workflow", "payload", wf, "error", err)
		return err
	}

	// Update latest status table for CI workflow within the transaction
	// Check if CiPipeline is loaded, if not pass 0 as appId to let the function fetch it
	appId := 0
	if wf.CiPipeline != nil {
		appId = wf.CiPipeline.AppId
	}
	err = impl.workflowStatusUpdateService.UpdateCiWorkflowStatusLatest(tx, wf.CiPipelineId, appId, wf.Id, wf.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("error in updating ci workflow status latest", "err", err, "pipelineId", wf.CiPipelineId, "workflowId", wf.Id)
		return err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.Logger.Errorw("error in committing transaction", "workflowName", wf.Name, "error", err)
		return err
	}

	return nil

}
