package cdWorkflow

import (
	"errors"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
)

var WfrTerminalStatusList = []string{WorkflowAborted, WorkflowFailed, WorkflowSucceeded, bean.HIBERNATING, string(health.HealthStatusHealthy), string(health.HealthStatusDegraded)}

type WorkflowStatus int

const (
	WF_UNKNOWN WorkflowStatus = iota
	REQUEST_ACCEPTED
	ENQUEUED
	QUE_ERROR
	WF_STARTED
	DROPPED_STALE
	DEQUE_ERROR
	TRIGGER_ERROR
)

const (
	WorkflowStarting           = "Starting"
	WorkflowInQueue            = "Queued"
	WorkflowInitiated          = "Initiating"
	WorkflowInProgress         = "Progressing"
	WorkflowAborted            = "Aborted"
	WorkflowFailed             = "Failed"
	WorkflowSucceeded          = "Succeeded"
	WorkflowTimedOut           = "TimedOut"
	WorkflowUnableToFetchState = "UnableToFetch"
	WorkflowTypeDeploy         = "DEPLOY"
	WorkflowTypePre            = "PRE"
	WorkflowTypePost           = "POST"
)

func (a WorkflowStatus) String() string {
	return [...]string{"WF_UNKNOWN", "REQUEST_ACCEPTED", "ENQUEUED", "QUE_ERROR", "WF_STARTED", "DROPPED_STALE", "DEQUE_ERROR", "TRIGGER_ERROR"}[a]
}

var ErrorDeploymentSuperseded = errors.New(NEW_DEPLOYMENT_INITIATED)

const (
	WORKFLOW_EXECUTOR_TYPE_AWF    = "AWF"
	WORKFLOW_EXECUTOR_TYPE_SYSTEM = "SYSTEM"
	NEW_DEPLOYMENT_INITIATED      = "A new deployment was initiated before this deployment completed!"
	PIPELINE_DELETED              = "The pipeline has been deleted!"
	FOUND_VULNERABILITY           = "Found vulnerability on image"
	GITOPS_REPO_NOT_CONFIGURED    = "GitOps repository is not configured for the app"
)

type WorkflowExecutorType string

type CdWorkflowRunnerArtifactMetadata struct {
	AppId            int  `pg:"app_id"`
	EnvId            int  `pg:"env_id"`
	CiArtifactId     int  `pg:"ci_artifact_id"`
	ParentCiArtifact int  `pg:"parent_ci_artifact"`
	Scanned          bool `pg:"scanned"`
}
