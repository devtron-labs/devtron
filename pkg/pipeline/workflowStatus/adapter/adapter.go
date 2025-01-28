package adapter

import (
	"encoding/json"
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

func ConvertWfStatusToDevtronStatus(wfStatus string, wfMessage string) bean.WorkflowStageStatus {
	// implementation
	switch strings.ToLower(wfStatus) {
	case "pending", strings.ToLower(cdWorkflow.WorkflowWaitingToStart):
		return bean.WORKFLOW_STAGE_STATUS_NOT_STARTED
	case "starting", "running":
		return bean.WORKFLOW_STAGE_STATUS_RUNNING
	case "succeeded":
		return bean.WORKFLOW_STAGE_STATUS_SUCCEEDED
	case "failed", "error", "errored":
		if strings.ToLower(wfMessage) == strings.ToLower(constants.POD_TIMEOUT_MESSAGE) {
			return bean.WORKFLOW_STAGE_STATUS_TIMEOUT
		} else {
			return bean.WORKFLOW_STAGE_STATUS_FAILED
		}
	case "aborted", "cancelled":
		return bean.WORKFLOW_STAGE_STATUS_ABORTED
	default:
		log.Println("unknown wf status", "wf", wfStatus)
		return bean.WORKFLOW_STAGE_STATUS_UNKNOWN
	}
}

func GetWfStageDataForOldRunners(wfId int, wfStatus, wfMessage, podStatus, wfType string, startTime, endTime time.Time) map[string][]*bean.WorkflowStageDto {
	// implementation
	resp := make(map[string][]*bean.WorkflowStageDto)

	executionStage := GetDefaultWorkflowExecutionStage(wfId, wfType)
	executionStage.Status = ConvertWfStatusToDevtronStatus(wfStatus, wfMessage)
	executionStage.StartTime = startTime.Format(bean3.LayoutRFC3339)
	executionStage.EndTime = endTime.Format(bean3.LayoutRFC3339)
	executionStage.Message = wfMessage

	resp[bean.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW.ToString()] = append(resp[bean.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW.ToString()], ConvertDBWorkflowStageToDto(executionStage))

	podStage := GetDefaultPodExecutionStage(wfId, wfType)
	podStage.Status = ConvertWfStatusToDevtronStatus(podStatus, wfMessage)
	podStage.StartTime = startTime.Format(bean3.LayoutRFC3339)
	podStage.EndTime = endTime.Format(bean3.LayoutRFC3339)
	podStage.Message = wfMessage
	resp[bean.WORKFLOW_STAGE_STATUS_TYPE_POD.ToString()] = append(resp[bean.WORKFLOW_STAGE_STATUS_TYPE_POD.ToString()], ConvertDBWorkflowStageToDto(executionStage))

	return resp
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
		StatusType:   bean.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW,
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
		StatusType:   bean.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW,
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
		StatusType:   bean.WORKFLOW_STAGE_STATUS_TYPE_POD,
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
