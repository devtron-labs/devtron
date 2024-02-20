package bean

import (
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	bean4 "github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
)

type BulkTriggerRequest struct {
	CiArtifactId int `sql:"ci_artifact_id"`
	PipelineId   int `sql:"pipeline_id"`
}

type StopDeploymentGroupRequest struct {
	DeploymentGroupId int               `json:"deploymentGroupId" validate:"required"`
	UserId            int32             `json:"userId"`
	RequestType       bean4.RequestType `json:"requestType" validate:"oneof=START STOP"`
}

type DeploymentGroupAppWithEnv struct {
	EnvironmentId     int               `json:"environmentId"`
	DeploymentGroupId int               `json:"deploymentGroupId"`
	AppId             int               `json:"appId"`
	Active            bool              `json:"active"`
	UserId            int32             `json:"userId"`
	RequestType       bean4.RequestType `json:"requestType" validate:"oneof=START STOP"`
}

type DeployStageSuccessEventReq struct {
	DeployStageType            bean2.WorkflowType            `json:"deployStageType"`
	PipelineOverride           *chartConfig.PipelineOverride `json:"pipelineOverride"`
	CdWorkflowId               int                           `json:"cdWorkflowId"`
	PipelineId                 int                           `json:"pipelineId"`
	PluginRegistryImageDetails map[string][]string           `json:"pluginRegistryImageDetails"`
}
