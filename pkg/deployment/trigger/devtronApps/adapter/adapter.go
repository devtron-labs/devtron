package adapter

import (
	apiBean "github.com/devtron-labs/devtron/api/bean"
	helmBean "github.com/devtron-labs/devtron/api/helm-app/service/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"time"
)

func SetPipelineFieldsInOverrideRequest(overrideRequest *apiBean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline) {
	overrideRequest.PipelineId = pipeline.Id
	overrideRequest.PipelineName = pipeline.Name
	overrideRequest.EnvId = pipeline.EnvironmentId
	overrideRequest.EnvName = pipeline.Environment.Name
	overrideRequest.ClusterId = pipeline.Environment.ClusterId
	overrideRequest.AppId = pipeline.AppId
	overrideRequest.AppName = pipeline.App.AppName
	overrideRequest.DeploymentAppType = pipeline.DeploymentAppType
	overrideRequest.Namespace = pipeline.Environment.Namespace
	overrideRequest.ReleaseName = pipeline.DeploymentAppName
}

func GetVulnerabilityCheckRequest(cdPipeline *pipelineConfig.Pipeline, imageDigest string) *bean.VulnerabilityCheckRequest {
	return &bean.VulnerabilityCheckRequest{
		CdPipeline:  cdPipeline,
		ImageDigest: imageDigest,
	}
}

func NewAsyncCdDeployRequest(overrideRequest *apiBean.ValuesOverrideRequest, triggeredAt time.Time, triggeredBy int32) *eventProcessorBean.AsyncCdDeployRequest {
	return &eventProcessorBean.AsyncCdDeployRequest{
		ValuesOverrideRequest: overrideRequest,
		TriggeredAt:           triggeredAt,
		TriggeredBy:           triggeredBy,
	}
}

func NewAppIdentifierFromOverrideRequest(overrideRequest *apiBean.ValuesOverrideRequest) *helmBean.AppIdentifier {
	return &helmBean.AppIdentifier{
		ClusterId:   overrideRequest.ClusterId,
		Namespace:   overrideRequest.Namespace,
		ReleaseName: overrideRequest.ReleaseName,
	}
}
