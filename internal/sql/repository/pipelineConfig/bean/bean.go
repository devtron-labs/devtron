package bean

import (
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
)

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

type WorkflowStatus int

var WfrTerminalStatusList = []string{WorkflowAborted, WorkflowFailed, WorkflowSucceeded, application.HIBERNATING, string(health.HealthStatusHealthy), string(health.HealthStatusDegraded)}

func GetDeploymentStatus(isSuccess bool) string {
	switch isSuccess {
	case true:
		return WorkflowSucceeded
	default:
		return WorkflowFailed
	}
}

func GetDeploymentStartStatus(isAsync bool) string {
	switch isAsync {
	case true:
		return WorkflowInQueue
	default:
		return WorkflowInProgress
	}
}

func (a WorkflowStatus) String() string {
	return [...]string{"WF_UNKNOWN", "REQUEST_ACCEPTED", "ENQUEUED", "QUE_ERROR", "WF_STARTED", "DROPPED_STALE", "DEQUE_ERROR", "TRIGGER_ERROR"}[a]
}
