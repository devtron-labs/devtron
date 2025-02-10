package util

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"slices"
)

func ComputeWorkflowStatus(currentWfDBstatus, wfStatus, stageStatus string) string {
	updatedWfStatus := currentWfDBstatus
	if !slices.Contains(cdWorkflow.WfrTerminalStatusList, currentWfDBstatus) {
		if len(stageStatus) > 0 && !slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
			updatedWfStatus = stageStatus
		} else {
			updatedWfStatus = wfStatus
		}
	}
	return updatedWfStatus
}
