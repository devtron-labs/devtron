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

package adapter

import (
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"log"
	"strings"
	"time"
)

func ConvertDBWorkflowStageToDto(stage *repository.WorkflowExecutionStage) *bean.WorkflowStageDto {
	if stage == nil {
		return &bean.WorkflowStageDto{}
	}
	return &bean.WorkflowStageDto{
		Id:           stage.Id,
		StageName:    stage.StageName,
		Status:       stage.Status,
		Message:      stage.Message,
		Metadata:     getMetadataJson(stage.Metadata),
		WorkflowId:   stage.WorkflowId,
		WorkflowType: stage.WorkflowType,
		StartTime:    stage.StartTime,
		EndTime:      stage.EndTime,
	}
}

func getMetadataJson(metadata string) map[string]interface{} {
	var response map[string]interface{}
	//todo handle error
	json.Unmarshal([]byte(metadata), &response)
	//if err != nil {
	//	return nil, err
	//}
	return response
}

// for workflow there can be other status map than for pod status like in aborted case
func ConvertStatusToDevtronStatus(wfStatus string, wfMessage string) bean.WorkflowStageStatus {
	// implementation
	switch strings.ToLower(wfStatus) {
	case strings.ToLower(string(v1alpha1.NodePending)), strings.ToLower(cdWorkflow.WorkflowWaitingToStart):
		return bean.WORKFLOW_STAGE_STATUS_NOT_STARTED
	case strings.ToLower(cdWorkflow.WorkflowStarting), strings.ToLower(string(v1alpha1.NodeRunning)):
		return bean.WORKFLOW_STAGE_STATUS_RUNNING
	case strings.ToLower(cdWorkflow.WorkflowSucceeded):
		return bean.WORKFLOW_STAGE_STATUS_SUCCEEDED
	case strings.ToLower(cdWorkflow.WorkflowFailed), strings.ToLower(string(v1alpha1.NodeError)), "errored":
		if strings.ToLower(wfMessage) == strings.ToLower(constants.POD_TIMEOUT_MESSAGE) {
			return bean.WORKFLOW_STAGE_STATUS_TIMEOUT
		} else {
			return bean.WORKFLOW_STAGE_STATUS_FAILED
		}
	case strings.ToLower(cdWorkflow.WorkflowAborted), strings.ToLower(cdWorkflow.WorkflowCancel):
		return bean.WORKFLOW_STAGE_STATUS_ABORTED
	default:
		log.Println("unknown wf status", "wf", wfStatus)
		return bean.WORKFLOW_STAGE_STATUS_UNKNOWN
	}
}

func GetDefaultPipelineStatusForWorkflow(wfId int, wfType string) []*repository.WorkflowExecutionStage {
	// implementation
	resp := []*repository.WorkflowExecutionStage{}
	resp = append(resp, GetDefaultWorkflowPreparationStage(wfId, wfType))
	resp = append(resp, GetDefaultWorkflowExecutionStage(wfId, wfType))
	resp = append(resp, GetDefaultPodExecutionStage(wfId, wfType))
	return resp
}

func GetDefaultWorkflowPreparationStage(workflowId int, workflowType string) *repository.WorkflowExecutionStage {
	// implementation
	return &repository.WorkflowExecutionStage{
		StageName:    bean.WORKFLOW_PREPARATION,
		Status:       bean.WORKFLOW_STAGE_STATUS_RUNNING,
		StatusFor:    bean.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW,
		StartTime:    time.Now().Format(bean3.LayoutRFC3339),
		WorkflowId:   workflowId,
		WorkflowType: workflowType,
		Message:      "",
		Metadata:     "{}",
		EndTime:      "",
		//todo do we need audit log since ci-workflow also doesn't have it ??
		AuditLog: sql.NewDefaultAuditLog(1),
	}
}

func GetDefaultWorkflowExecutionStage(workflowId int, workflowType string) *repository.WorkflowExecutionStage {
	// implementation
	return &repository.WorkflowExecutionStage{
		StageName:    bean.WORKFLOW_EXECUTION,
		Status:       bean.WORKFLOW_STAGE_STATUS_NOT_STARTED,
		StatusFor:    bean.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW,
		StartTime:    "",
		WorkflowId:   workflowId,
		WorkflowType: workflowType,
		Message:      "",
		Metadata:     "{}",
		EndTime:      "",
		//todo do we need audit log since ci-workflow also doesn't have it ??
		AuditLog: sql.NewDefaultAuditLog(1),
	}
}

func GetDefaultPodExecutionStage(workflowId int, workflowType string) *repository.WorkflowExecutionStage {
	// implementation
	return &repository.WorkflowExecutionStage{
		StageName:    bean.POD_EXECUTION,
		Status:       bean.WORKFLOW_STAGE_STATUS_NOT_STARTED,
		StatusFor:    bean.WORKFLOW_STAGE_STATUS_TYPE_POD,
		StartTime:    "",
		WorkflowId:   workflowId,
		WorkflowType: workflowType,
		Message:      "",
		Metadata:     "{}",
		EndTime:      "",
		//todo do we need audit log since ci-workflow also doesn't have it ??
		AuditLog: sql.NewDefaultAuditLog(1),
	}
}
