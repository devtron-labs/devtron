package adapter

import (
	"context"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
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

func GetTriggerRequirementRequest(artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, runner *pipelineConfig.CdWorkflowRunner, triggeredBy int32, context context.Context) *bean2.TriggerRequirementRequestDto {
	return &bean2.TriggerRequirementRequestDto{
		Artifact:    artifact,
		Pipeline:    pipeline,
		Runner:      runner,
		TriggeredBy: triggeredBy,
		Context:     context,
	}
}
