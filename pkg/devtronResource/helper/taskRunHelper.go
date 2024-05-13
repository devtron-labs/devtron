package helper

import (
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	slices2 "golang.org/x/exp/slices"
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

var StatusVsRolloutStatusMap = map[string]bean.RolloutStatus{
	bean.StartingStatus:      bean.Ongoing,
	bean.RunningStatus:       bean.Ongoing,
	bean.InitiatingStatus:    bean.Ongoing,
	bean.ProgressingStatus:   bean.Ongoing,
	bean.QueuedStatus:        bean.Ongoing,
	bean.AbortedStatus:       bean.Failed,
	bean.FailedStatus:        bean.Failed,
	bean.TimedOutStatus:      bean.Failed,
	bean.UnableToFetchStatus: bean.Failed,
	bean.NotTriggeredStatus:  bean.YetToTrigger,
	bean.SucceededStatus:     bean.Completed,
}

func CalculateRolloutStatus(releaseInfo *bean.CdPipelineReleaseInfo) bean.RolloutStatus {
	finalRolloutStatus := make([]bean.RolloutStatus, 0, 3)
	// appending deployment status first
	finalRolloutStatus = append(finalRolloutStatus, StatusVsRolloutStatusMap[releaseInfo.DeployStatus])
	if releaseInfo.ExistingStages.Pre && releaseInfo.ExistingStages.Post {
		finalRolloutStatus = append(finalRolloutStatus, StatusVsRolloutStatusMap[releaseInfo.PreStatus], StatusVsRolloutStatusMap[releaseInfo.PostStatus])
	} else if releaseInfo.ExistingStages.Pre {
		finalRolloutStatus = append(finalRolloutStatus, StatusVsRolloutStatusMap[releaseInfo.PreStatus])
	} else if releaseInfo.ExistingStages.Post {
		finalRolloutStatus = append(finalRolloutStatus, StatusVsRolloutStatusMap[releaseInfo.PostStatus])
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

func checkIfEveryElementIsGivenValue(slice []bean.RolloutStatus, value bean.RolloutStatus) bool {
	for _, v := range slice {
		if v != value {
			return false
		}
	}
	return true
}
