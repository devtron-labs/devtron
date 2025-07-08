package types

// CiTriggerAuditRequest represents the request for CI trigger audit
type CiTriggerAuditRequest struct {
	WorkflowId int
	*CommonAuditRequest
}

// CdTriggerAuditRequest represents the request for CD trigger audit (Pre-CD, Post-CD)
type CdTriggerAuditRequest struct {
	WorkflowRunnerId int
	WorkflowType     WorkflowType
	*CommonAuditRequest
}

type CommonAuditRequest struct {
	WorkflowRequest *WorkflowRequest
	TriggeredBy     int32
	PipelineId      int //ciPipelineId in case auditing ci trigger, else cdPipelineId
}

type WorkflowType string

const (
	CI_WORKFLOW_TYPE      WorkflowType = "CI"
	PRE_CD_WORKFLOW_TYPE  WorkflowType = "PRE_CD"
	POST_CD_WORKFLOW_TYPE WorkflowType = "POST_CD"
)

const (
	TriggerAuditSchemaVersionV1 = "V1"
)
