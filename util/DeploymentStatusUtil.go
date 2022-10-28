package util

import (
	"github.com/argoproj/gitops-engine/pkg/health"
)

const WorkflowAborted = "Aborted"
const WorkflowFailed = "Failed"

func IsTerminalStatus(status string) bool {
	switch status {
	case
		string(health.HealthStatusHealthy),
		string(health.HealthStatusDegraded),
		WorkflowAborted,
		WorkflowFailed:
		return true
	}
	return false
}
