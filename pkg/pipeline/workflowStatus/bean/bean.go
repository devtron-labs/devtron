package bean

type WorkflowStageName string

const (
	WORKFLOW_PREPARATION WorkflowStageName = "Preparation"
	WORKFLOW_EXECUTION   WorkflowStageName = "Execution"
	POD_EXECUTION        WorkflowStageName = "Execution"
)

func (n WorkflowStageName) ToString() string {
	return string(n)
}

type WorkflowStageStatusType string

const (
	WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW WorkflowStageStatusType = "workflow"
	WORKFLOW_STAGE_STATUS_TYPE_POD      WorkflowStageStatusType = "pod"
)

func (n WorkflowStageStatusType) ToString() string {
	return string(n)
}

type WorkflowStageStatus string

const (
	WORKFLOW_STAGE_STATUS_NOT_STARTED WorkflowStageStatus = "NOT_STARTED"
	WORKFLOW_STAGE_STATUS_UNKNOWN     WorkflowStageStatus = "UNKNOWN"
	WORKFLOW_STAGE_STATUS_RUNNING     WorkflowStageStatus = "RUNNING"
	WORKFLOW_STAGE_STATUS_SUCCEEDED   WorkflowStageStatus = "SUCCEEDED"
	WORKFLOW_STAGE_STATUS_FAILED      WorkflowStageStatus = "FAILED"
	WORKFLOW_STAGE_STATUS_ABORTED     WorkflowStageStatus = "ABORTED"
	WORKFLOW_STAGE_STATUS_CANCELLED   WorkflowStageStatus = "CANCELLED"
	WORKFLOW_STAGE_STATUS_TIMEOUT     WorkflowStageStatus = "TIMEOUT"
	//don't forget to add new status in IsTerminal() method if it is terminal status
)

func (n WorkflowStageStatus) ToString() string {
	return string(n)
}

func (n WorkflowStageStatus) IsTerminal() bool {
	switch n {
	case WORKFLOW_STAGE_STATUS_SUCCEEDED, WORKFLOW_STAGE_STATUS_FAILED, WORKFLOW_STAGE_STATUS_ABORTED, WORKFLOW_STAGE_STATUS_TIMEOUT, WORKFLOW_STAGE_STATUS_CANCELLED:
		return true
	default:
		return false
	}
}

type WorkflowStageDto struct {
	Id           int                    `json:"id"`
	StageName    WorkflowStageName      `json:"stageName"`
	Status       WorkflowStageStatus    `json:"status"`
	Message      string                 `json:"message"`
	Metadata     map[string]interface{} `json:"metadata"`
	WorkflowId   int                    `json:"workflowId"`
	WorkflowType string                 `json:"workflowType"`
	StartTime    string                 `json:"startTime"`
	EndTime      string                 `json:"endTime"`
}
