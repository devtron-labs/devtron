/*
 * Copyright (c) 2024. Devtron Inc.
 */

package helper

import (
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/health"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	pipelineConfig "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	stageBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	slices2 "golang.org/x/exp/slices"
	"k8s.io/utils/strings/slices"
)

func GetTaskTypeBasedOnWorkflowType(workflowType bean3.WorkflowType) bean.TaskType {
	switch workflowType {
	case bean3.CD_WORKFLOW_TYPE_PRE:
		return bean.TaskTypePreDeployment
	case bean3.CD_WORKFLOW_TYPE_POST:
		return bean.TaskTypePostDeployment
	case bean3.CD_WORKFLOW_TYPE_DEPLOY:
		return bean.TaskTypeDeployment
	}
	// default task type is empty
	return ""
}

func ValidateTasksPayload(tasks []*bean.Task) bool {
	if len(tasks) > 0 {
		return true
	}
	return false
}

func ConvertErrorAccordingToDeployment(err error) error {
	if err == nil {
		return util.NewApiError().WithCode("200")
	}
	return util.NewApiError().WithCode(constants.NotProcessed).WithUserMessage(err.Error()).WithInternalMessage(err.Error())
}

func ConvertErrorAccordingToFeasibility(err error, deploymentWindowByPassed bool) error {
	if deploymentWindowByPassed {
		return util.NewApiError().WithCode(constants.DeploymentWindowByPassed).WithUserMessage(bean.DeploymentByPassingMessage)
	}
	if err == nil {
		return util.NewApiError().WithCode("200")
	}
	return err
}

func GetTaskRunSourceIdentifier(id int, idType bean.IdType, resourceId, resourceSchemaId int) string {
	return fmt.Sprintf("%d|%s|%d|%d", id, idType, resourceId, resourceSchemaId)
}

func GetTaskRunSourceDependencyIdentifier(id int, idType bean.IdType, resourceId, resourceSchemaId int) string {
	return fmt.Sprintf("%d|%s|%d|%d", id, idType, resourceId, resourceSchemaId)
}

func GetTaskRunIdentifier(id int, idType bean.IdType, resourceId, resourceSchemaId int) string {
	return fmt.Sprintf("%d|%s|%d|%d", id, idType, resourceId, resourceSchemaId)
}

var DeploymentStatusVsRolloutStatusMap = map[string]bean.ReleaseDeploymentStatus{
	pipelineConfig.WorkflowStarting:           bean.Ongoing,
	bean.RunningStatus:                        bean.Ongoing,
	pipelineConfig.WorkflowInitiated:          bean.Ongoing,
	pipelineConfig.WorkflowInProgress:         bean.Ongoing,
	pipelineConfig.WorkflowInQueue:            bean.Ongoing,
	pipelineConfig.WorkflowAborted:            bean.Failed,
	pipelineConfig.WorkflowFailed:             bean.Failed,
	pipelineConfig.WorkflowTimedOut:           bean.Failed,
	pipelineConfig.WorkflowUnableToFetchState: bean.Failed,
	bean2.Degraded:                            bean.Failed,
	bean.Error:                                bean.Failed,
	executors.WorkflowCancel:                  bean.Failed,
	stageBean.NotTriggered:                    bean.YetToTrigger,
	pipelineConfig.WorkflowSucceeded:          bean.Completed,
	bean2.Healthy:                             bean.Completed,
}

func CalculateRolloutStatus(releaseInfo *bean.CdPipelineReleaseInfo) bean.ReleaseDeploymentStatus {
	finalRolloutStatus := make([]bean.ReleaseDeploymentStatus, 0, 3)
	// appending deployment status first
	finalRolloutStatus = append(finalRolloutStatus, DeploymentStatusVsRolloutStatusMap[releaseInfo.DeployStatus])
	if releaseInfo.ExistingStages.Pre {
		finalRolloutStatus = append(finalRolloutStatus, DeploymentStatusVsRolloutStatusMap[releaseInfo.PreStatus])
	}
	if releaseInfo.ExistingStages.Post {
		finalRolloutStatus = append(finalRolloutStatus, DeploymentStatusVsRolloutStatusMap[releaseInfo.PostStatus])
	}
	if slices2.Contains(finalRolloutStatus, bean.Failed) {
		return bean.Failed
	}
	if checkIfEveryElementIsGivenValue(finalRolloutStatus, bean.Completed) {
		return bean.Completed
	}
	if checkIfEveryElementIsGivenValue(finalRolloutStatus, bean.YetToTrigger) {
		return bean.YetToTrigger
	}
	return bean.Ongoing
}

func checkIfEveryElementIsGivenValue(slice []bean.ReleaseDeploymentStatus, value bean.ReleaseDeploymentStatus) bool {
	for _, v := range slice {
		if v != value {
			return false
		}
	}
	return true
}

func IsStatusSucceeded(status string) bool {
	return slices.Contains([]string{pipelineConfig.WorkflowSucceeded, string(health.HealthStatusHealthy)}, status)

}
