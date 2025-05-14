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

package workflowStatus

import (
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"slices"
	"strings"
	"time"
)

type WorkFlowStageStatusService interface {
	GetWorkflowStagesByWorkflowIdsAndWfType(wfIds []int, wfType string) ([]*repository.WorkflowExecutionStage, error)
	GetWorkflowStagesByWorkflowIdAndType(wfId int, wfType string) ([]*repository.WorkflowExecutionStage, error)

	SaveWorkflowStages(wfId int, wfType, wfName string, tx *pg.Tx) error
	UpdateWorkflowStages(wfId int, wfType, wfName, wfStatus, podStatus, message, podName string, tx *pg.Tx) (string, string, error)
	ConvertDBWorkflowStageToMap(workflowStages []*repository.WorkflowExecutionStage, wfId int, status, podStatus, message, wfType string, startTime, endTime time.Time) map[string][]*bean2.WorkflowStageDto
}

type WorkFlowStageStatusServiceImpl struct {
	logger                   *zap.SugaredLogger
	workflowStatusRepository repository.WorkflowStageRepository
	ciWorkflowRepository     pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository     pipelineConfig.CdWorkflowRepository
	transactionManager       sql.TransactionWrapper
	config                   *types.CiConfig
}

func NewWorkflowStageFlowStatusServiceImpl(logger *zap.SugaredLogger,
	workflowStatusRepository repository.WorkflowStageRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	transactionManager sql.TransactionWrapper,
) *WorkFlowStageStatusServiceImpl {
	wfStageServiceImpl := &WorkFlowStageStatusServiceImpl{
		logger:                   logger,
		workflowStatusRepository: workflowStatusRepository,
		ciWorkflowRepository:     ciWorkflowRepository,
		cdWorkflowRepository:     cdWorkflowRepository,
		transactionManager:       transactionManager,
	}
	ciConfig, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	wfStageServiceImpl.config = ciConfig
	return wfStageServiceImpl
}

func (impl *WorkFlowStageStatusServiceImpl) getUpdatedPipelineStagesForWorkflow(wfId int, wfType string, wfStatus string, podStatus string, message string, podName string) ([]*repository.WorkflowExecutionStage, string, string) {
	// implementation
	currentWorkflowStages, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdAndType(wfId, wfType)
	if err != nil {
		impl.logger.Errorw("error in getting workflow stages", "workflowId", wfId, "error", err)
		return nil, wfStatus, podStatus
	}
	if len(currentWorkflowStages) == 0 {
		return []*repository.WorkflowExecutionStage{}, wfStatus, podStatus
	}

	var currentWfDBstatus, currentPodStatus string

	if wfType == bean.CI_WORKFLOW_TYPE.String() {
		//get current status from db
		dbWf, err := impl.ciWorkflowRepository.FindById(wfId)
		if err != nil {
			impl.logger.Errorw("error in getting workflow", "wfId", wfId, "error", err)
			return nil, wfStatus, podStatus
		}
		currentWfDBstatus = dbWf.Status
		currentPodStatus = dbWf.PodStatus
	} else {
		dbWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfId)
		if err != nil {
			impl.logger.Errorw("error in getting workflow runner", "wfId", wfId, "error", err)
			return nil, wfStatus, podStatus
		}
		currentWfDBstatus = dbWfr.Status
		currentPodStatus = dbWfr.PodStatus
	}

	impl.logger.Infow("step-1", "wfId", wfId, "wfType", wfType, "wfStatus", wfStatus, "currentWfDBstatus", currentWfDBstatus, "podStatus", podStatus, "currentPodStatus", currentPodStatus, "message", message)
	currentWorkflowStages, updatedPodStatus := impl.updatePodStages(currentWorkflowStages, podStatus, currentPodStatus, message, podName, wfStatus)
	impl.logger.Infow("step-2", "updatedPodStatus", updatedPodStatus, "updated pod stages", currentWorkflowStages)
	currentWorkflowStages, updatedWfStatus := impl.updateWorkflowStagesToDevtronStatus(currentWorkflowStages, wfStatus, currentWfDBstatus, message, podStatus)
	impl.logger.Infow("step-3", "updatedWfStatus", updatedWfStatus, "updatedPodStatus", updatedPodStatus, "updated workflow stages", currentWorkflowStages)

	return currentWorkflowStages, updatedWfStatus, updatedPodStatus
}

func (impl *WorkFlowStageStatusServiceImpl) updatePodStages(currentWorkflowStages []*repository.WorkflowExecutionStage, podStatus string, currentPodStatus string, message string, podName string, wfStatus string) ([]*repository.WorkflowExecutionStage, string) {
	updatedPodStatus := currentPodStatus
	if !slices.Contains(cdWorkflow.WfrTerminalStatusList, currentPodStatus) {
		updatedPodStatus = podStatus
	}
	//update pod stage status by using convertPodStatusToDevtronStatus
	for _, stage := range currentWorkflowStages {
		if stage.StatusFor == bean2.WORKFLOW_STAGE_STATUS_TYPE_POD {
			// add pod name in stage metadata if not empty
			if len(podName) > 0 {
				marshalledMetadata, _ := json.Marshal(map[string]string{"podName": podName})
				stage.Metadata = string(marshalledMetadata)
			}
			switch podStatus {
			case string(v1alpha1.NodePending):
				if !stage.Status.IsTerminal() {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED
				}
			case string(v1alpha1.NodeRunning):
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_UNKNOWN {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_RUNNING
					stage.StartTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			case string(v1alpha1.NodeSucceeded):
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_UNKNOWN {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_SUCCEEDED
					stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			case string(v1alpha1.NodeFailed), string(v1alpha1.NodeError), string(v1alpha1.NodeSkipped):
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_UNKNOWN {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_FAILED
					stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
						stage.StartTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}
			default:
				impl.logger.Errorw("unknown pod status", "podStatus", podStatus, "message", message)
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
					extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, "")
					if extractedStatus.IsTerminal() {
						stage.Status = extractedStatus
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
						updatedPodStatus = wfStatus
					}
				} else {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN
					stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			}
		}
	}
	return currentWorkflowStages, updatedPodStatus
}

// Each case has 2 steps to do
// step-1: update latest status field if its not terminal already
// step-2: accordingly update stage status
func (impl *WorkFlowStageStatusServiceImpl) updateWorkflowStagesToDevtronStatus(currentWorkflowStages []*repository.WorkflowExecutionStage, wfStatus string, currentWfDBstatus, wfMessage string, podStatus string) ([]*repository.WorkflowExecutionStage, string) {
	// implementation
	updatedWfStatus := currentWfDBstatus
	//todo for switch case use enums
	switch strings.ToLower(podStatus) {
	case strings.ToLower(string(v1alpha1.NodePending)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, cdWorkflow.WorkflowWaitingToStart)

		// update workflow preparation stage and pod status if terminal
		for _, stage := range currentWorkflowStages {
			if stage.StageName == bean2.WORKFLOW_PREPARATION && !stage.Status.IsTerminal() {
				extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
				if extractedStatus != bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
					stage.Status = extractedStatus
				}
			}

			//also mark pod status as terminal if wfstatus is terminal
			if stage.StageName == bean2.POD_EXECUTION && slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
				stage.Status = adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
				stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
			}
		}
	case strings.ToLower(string(v1alpha1.NodeRunning)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, constants.Running)

		//if pod is running, update preparation and execution stages
		for _, stage := range currentWorkflowStages {
			if stage.StatusFor == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark preparation stage as completed
				if stage.StageName == bean2.WORKFLOW_PREPARATION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_SUCCEEDED
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}

				//mark execution stage as started
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_RUNNING
						stage.StartTime = time.Now().Format(bean3.LayoutRFC3339)
					} else if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
						if extractedStatus.IsTerminal() {
							stage.Status = extractedStatus
							stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
						}
					}
				}
			}
		}
	case strings.ToLower(string(v1alpha1.NodeSucceeded)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, cdWorkflow.WorkflowSucceeded)

		//if pod is succeeded, update execution stage
		for _, stage := range currentWorkflowStages {
			if stage.StatusFor == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_SUCCEEDED
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}
			}
		}
	case strings.ToLower(string(v1alpha1.NodeFailed)), strings.ToLower(string(v1alpha1.NodeError)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, cdWorkflow.WorkflowFailed)

		//if pod is failed, update execution stage
		for _, stage := range currentWorkflowStages {
			if stage.StatusFor == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
						if extractedStatus.IsTerminal() {
							stage.Status = extractedStatus
							stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
							if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_TIMEOUT {
								updatedWfStatus = cdWorkflow.WorkflowTimedOut
							}
							if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_ABORTED {
								updatedWfStatus = cdWorkflow.WorkflowCancel
							}
						}
					}
				} else if stage.StageName == bean2.WORKFLOW_PREPARATION && !stage.Status.IsTerminal() {
					extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
					if extractedStatus.IsTerminal() {
						stage.Status = extractedStatus
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
						if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_TIMEOUT {
							updatedWfStatus = cdWorkflow.WorkflowTimedOut
						}
						if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_ABORTED {
							updatedWfStatus = cdWorkflow.WorkflowCancel
						}
					}
				}
			}
		}
	default:
		impl.logger.Errorw("unknown pod status", "podStatus", podStatus)
		//mark workflow stage status as unknown
		for _, stage := range currentWorkflowStages {
			if stage.StatusFor == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN
						updatedWfStatus = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN.ToString()
					}
				}
				if stage.StageName == bean2.WORKFLOW_PREPARATION && !stage.Status.IsTerminal() {
					//assumption: once pod is running we don't internally do any extra operation which would call this function and simply update status accrording to kubewatch events
					//that's why we are getting pod status as unknown because we don't explicity set pod status
					//this is the case when our internal code has called to update status before actually scheduling pod
					//update wf status as given in request, don't change that
					updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, "")
					extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
					if extractedStatus.IsTerminal() {
						stage.Status = extractedStatus
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
						if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_TIMEOUT {
							updatedWfStatus = cdWorkflow.WorkflowTimedOut
						}
						if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_ABORTED {
							updatedWfStatus = cdWorkflow.WorkflowCancel
						}
					}
				}
			}

		}
	}

	return currentWorkflowStages, updatedWfStatus
}

func (impl *WorkFlowStageStatusServiceImpl) GetWorkflowStagesByWorkflowIdsAndWfType(wfIds []int, wfType string) ([]*repository.WorkflowExecutionStage, error) {
	// implementation

	dbData, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdsAndWtype(wfIds, wfType)
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow stages", "error", err)
		return nil, err
	}
	if len(dbData) == 0 {
		return []*repository.WorkflowExecutionStage{}, nil
	} else {
		return dbData, nil
	}
}

func (impl *WorkFlowStageStatusServiceImpl) GetWorkflowStagesByWorkflowIdAndType(wfId int, wfType string) ([]*repository.WorkflowExecutionStage, error) {
	// implementation

	dbData, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdAndWtype(wfId, wfType)
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow stages", "error", err)
		return nil, err
	}
	if len(dbData) == 0 {
		return []*repository.WorkflowExecutionStage{}, nil
	} else {
		return dbData, nil
	}
}

func (impl *WorkFlowStageStatusServiceImpl) ConvertDBWorkflowStageToMap(workflowStages []*repository.WorkflowExecutionStage, wfId int, status, podStatus, message, wfType string, startTime, endTime time.Time) map[string][]*bean2.WorkflowStageDto {
	wfMap := make(map[string][]*bean2.WorkflowStageDto)
	foundInDb := false
	if !impl.config.EnableWorkflowExecutionStage {
		// if flag is not enabled then return empty map
		return map[string][]*bean2.WorkflowStageDto{}
	}
	for _, wfStage := range workflowStages {
		if wfStage.WorkflowId == wfId {
			wfMap[wfStage.StatusFor.ToString()] = append(wfMap[wfStage.StatusFor.ToString()], adapter.ConvertDBWorkflowStageToDto(wfStage))
			foundInDb = true
		}
	}

	if !foundInDb {
		//for old data where stages are not saved in db return empty map
		return map[string][]*bean2.WorkflowStageDto{}
	}

	return wfMap

}

func (impl *WorkFlowStageStatusServiceImpl) SaveWorkflowStages(wfId int, wfType, wfName string, tx *pg.Tx) error {
	if impl.config.EnableWorkflowExecutionStage {
		pipelineStageStatus := adapter.GetDefaultPipelineStatusForWorkflow(wfId, wfType)
		pipelineStageStatus, err := impl.workflowStatusRepository.SaveWorkflowStages(pipelineStageStatus, tx)
		if err != nil {
			impl.logger.Errorw("error in saving workflow stages", "workflowName", wfName, "error", err)
			return err
		}
	} else {
		impl.logger.Debugw("workflow execution stage is disabled", "workflowName", wfName)
	}
	return nil
}

func (impl *WorkFlowStageStatusServiceImpl) UpdateWorkflowStages(wfId int, wfType, wfName, wfStatus, podStatus, message, podName string, tx *pg.Tx) (string, string, error) {
	if impl.config.EnableWorkflowExecutionStage {
		pipelineStageStatus, updatedWfStatus, updatedPodStatus := impl.getUpdatedPipelineStagesForWorkflow(wfId, wfType, wfStatus, podStatus, message, podName)
		pipelineStageStatus, err := impl.workflowStatusRepository.UpdateWorkflowStages(pipelineStageStatus, tx)
		if err != nil {
			impl.logger.Errorw("error in saving workflow stages", "workflowName", wfName, "error", err)
			return wfStatus, podStatus, err
		}

		return updatedWfStatus, updatedPodStatus, nil
	} else {
		impl.logger.Debugw("workflow execution stage is disabled", "workflowName", wfName)
	}
	return wfStatus, podStatus, nil
}
