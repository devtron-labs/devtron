package adaptor

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"time"
)

func GetCiWorkflowFromRefCiWorkflow(refCiWorkflow *pipelineConfig.CiWorkflow, workflowStatus string, triggeredBy int32) *pipelineConfig.CiWorkflow {
	return &pipelineConfig.CiWorkflow{
		Name:                  refCiWorkflow.Name,
		Status:                workflowStatus, // starting CIStage
		StartedOn:             time.Now(),
		CiPipelineId:          refCiWorkflow.CiPipelineId,
		Namespace:             refCiWorkflow.Namespace,
		BlobStorageEnabled:    refCiWorkflow.BlobStorageEnabled,
		GitTriggers:           refCiWorkflow.GitTriggers,
		CiBuildType:           refCiWorkflow.CiBuildType,
		TriggeredBy:           triggeredBy,
		ReferenceCiWorkflowId: refCiWorkflow.Id, // Reference to original workflow
		ExecutorType:          refCiWorkflow.ExecutorType,
		EnvironmentId:         refCiWorkflow.EnvironmentId,
	}
}
