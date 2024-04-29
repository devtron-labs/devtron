package adapter

import (
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
)

func SetPipelineFieldsInOverrideRequest(overrideRequest *bean3.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline) {
	overrideRequest.PipelineId = pipeline.Id
	overrideRequest.PipelineName = pipeline.Name
	overrideRequest.EnvId = pipeline.EnvironmentId
	overrideRequest.EnvName = pipeline.Environment.Name
	overrideRequest.ClusterId = pipeline.Environment.ClusterId
	overrideRequest.AppId = pipeline.AppId
	overrideRequest.AppName = pipeline.App.AppName
	overrideRequest.DeploymentAppType = pipeline.DeploymentAppType
}

func GetVulnerabilityCheckRequest(cdPipeline *pipelineConfig.Pipeline, imageDigest string) *bean.VulnerabilityCheckRequest {
	return &bean.VulnerabilityCheckRequest{
		CdPipeline:  cdPipeline,
		ImageDigest: imageDigest,
	}
}
