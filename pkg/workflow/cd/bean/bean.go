package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"time"
)

type CdWorkflowDto struct {
	Id             int                           `json:"id"`
	CiArtifactId   int                           `json:"ciArtifactId"`
	PipelineId     int                           `json:"pipelineId"`
	WorkflowStatus pipelineConfig.WorkflowStatus `json:"workflowStatus"`
	UserId         int32                         `json:"-"`
}

type CdWorkflowRunnerDto struct {
	Id           int               `json:"id"`
	Name         string            `json:"name"`
	WorkflowType bean.WorkflowType `json:"workflowType"` // pre,post,deploy
	//TODO: extract from repo service layer
	ExecutorType            pipelineConfig.WorkflowExecutorType `json:"executorType"` // awf, system
	Status                  string                              `json:"status"`
	PodStatus               string                              `json:"podStatus"`
	Message                 string                              `json:"message"`
	StartedOn               time.Time                           `json:"startedOn"`
	FinishedOn              time.Time                           `json:"finishedOn"`
	Namespace               string                              `json:"namespace"`
	LogLocation             string                              `json:"logFilePath"`
	TriggeredBy             int32                               `json:"triggeredBy"`
	CdWorkflowId            int                                 `json:"cdWorkflowId"`
	PodName                 string                              `json:"podName"`
	BlobStorageEnabled      bool                                `json:"blobStorageEnabled"`
	RefCdWorkflowRunnerId   int                                 `json:"refCdWorkflowRunnerId"`
	ImagePathReservationIds []int                               `json:"imagePathReservationIds"`
	ReferenceId             *string                             `json:"referenceId"`
}
