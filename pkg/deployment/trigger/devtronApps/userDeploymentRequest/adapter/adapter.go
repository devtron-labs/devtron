package adapter

import (
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/repository"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
)

func NewValuesOverrideRequest(userDeploymentRequest *repository.UserDeploymentRequest) *apiBean.ValuesOverrideRequest {
	return &apiBean.ValuesOverrideRequest{
		PipelineId:                            userDeploymentRequest.PipelineId,
		CiArtifactId:                          userDeploymentRequest.CiArtifactId,
		AdditionalOverride:                    userDeploymentRequest.AdditionalOverride,
		ForceTrigger:                          userDeploymentRequest.ForceTrigger,
		DeploymentTemplate:                    userDeploymentRequest.Strategy,
		DeploymentWithConfig:                  userDeploymentRequest.DeploymentWithConfig,
		WfrIdForDeploymentWithSpecificTrigger: userDeploymentRequest.SpecificTriggerWfrId,
		CdWorkflowType:                        apiBean.CD_WORKFLOW_TYPE_DEPLOY,
		CdWorkflowId:                          userDeploymentRequest.CdWorkflowId,
		DeploymentType:                        userDeploymentRequest.DeploymentType,
		ForceSyncDeployment:                   userDeploymentRequest.ForceSyncDeployment,
	}
}

func NewAsyncCdDeployRequest(userDeploymentRequest *repository.UserDeploymentRequest) *bean.UserDeploymentRequest {
	return &bean.UserDeploymentRequest{
		Id:                    userDeploymentRequest.Id,
		ValuesOverrideRequest: NewValuesOverrideRequest(userDeploymentRequest),
		TriggeredAt:           userDeploymentRequest.TriggeredAt,
		TriggeredBy:           userDeploymentRequest.TriggeredBy,
	}
}

func NewUserDeploymentRequest(asyncCdDeployRequest *bean.UserDeploymentRequest) *repository.UserDeploymentRequest {
	valuesOverrideRequest := asyncCdDeployRequest.ValuesOverrideRequest
	return &repository.UserDeploymentRequest{
		PipelineId:           valuesOverrideRequest.PipelineId,
		CiArtifactId:         valuesOverrideRequest.CiArtifactId,
		AdditionalOverride:   valuesOverrideRequest.AdditionalOverride,
		ForceTrigger:         valuesOverrideRequest.ForceTrigger,
		Strategy:             valuesOverrideRequest.DeploymentTemplate,
		DeploymentWithConfig: valuesOverrideRequest.DeploymentWithConfig,
		SpecificTriggerWfrId: valuesOverrideRequest.WfrIdForDeploymentWithSpecificTrigger,
		CdWorkflowId:         valuesOverrideRequest.CdWorkflowId,
		DeploymentType:       valuesOverrideRequest.DeploymentType,
		ForceSyncDeployment:  valuesOverrideRequest.ForceSyncDeployment,
		TriggeredAt:          asyncCdDeployRequest.TriggeredAt,
		TriggeredBy:          asyncCdDeployRequest.TriggeredBy,
	}
}
