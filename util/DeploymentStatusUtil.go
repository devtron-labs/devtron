package util

import (
	"github.com/argoproj/gitops-engine/pkg/health"
)

const (
	WorkflowAborted   = "Aborted"
	WorkflowFailed    = "Failed"
	WorkflowSucceeded = "Succeeded"
)

func IsTerminalStatus(status string) bool {
	switch status {
	case
		string(health.HealthStatusHealthy),
		WorkflowAborted,
		WorkflowFailed,
		WorkflowSucceeded:
		return true
	}
	return false
}
