package helper

import (
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
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
