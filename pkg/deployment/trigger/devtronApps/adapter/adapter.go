package adapter

import (
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
)

func SetPipelineFieldsInOverrideRequest(overrideRequest *bean3.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline) {
	overrideRequest.PipelineId = pipeline.Id
	overrideRequest.PipelineName = pipeline.Name
	overrideRequest.EnvId = pipeline.EnvironmentId
	environment := pipeline.Environment
	overrideRequest.EnvName = environment.Name
	overrideRequest.ClusterId = environment.ClusterId
	overrideRequest.IsProdEnv = environment.Default
	overrideRequest.AppId = pipeline.AppId
	overrideRequest.ProjectId = pipeline.App.TeamId
	overrideRequest.AppName = pipeline.App.AppName
	overrideRequest.DeploymentAppType = pipeline.DeploymentAppType
}

func GetTriggerRequirementRequest(scope resourceQualifiers.Scope, triggerRequest bean2.TriggerRequest, stage resourceFilter.ReferenceType) *bean2.TriggerRequirementRequestDto {
	return &bean2.TriggerRequirementRequestDto{
		TriggerRequest: triggerRequest,
		Scope:          scope,
		Stage:          stage,
	}
}

func GetTriggerFeasibilityResponse(approvalRequestId int, triggerRequest bean2.TriggerRequest, filterIdVsState map[int]resourceFilter.FilterState, filters []*resourceFilter.FilterMetaDataBean) *bean2.TriggerFeasibilityResponse {
	return &bean2.TriggerFeasibilityResponse{
		ApprovalRequestId: approvalRequestId,
		TriggerRequest:    triggerRequest,
		FilterIdVsState:   filterIdVsState,
		Filters:           filters,
	}
}
