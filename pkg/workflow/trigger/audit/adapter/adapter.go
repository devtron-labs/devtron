package adapter

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/repository"
)

func GetWorkflowConfigSnapshot(workflowId int, workflowType types.WorkflowType, pipelineId int, compressedWorkflowJson string, triggerAuditSchemaVersion string, triggeredBy int32) *repository.WorkflowConfigSnapshot {
	return &repository.WorkflowConfigSnapshot{
		WorkflowId:                   workflowId,
		WorkflowType:                 workflowType,
		PipelineId:                   pipelineId,
		WorkflowRequestJson:          compressedWorkflowJson,
		WorkflowRequestSchemaVersion: triggerAuditSchemaVersion,
		AuditLog:                     sql.NewDefaultAuditLog(triggeredBy),
	}
}
