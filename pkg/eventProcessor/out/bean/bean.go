package bean

import (
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

type CdPipelineDeleteEvent struct {
	PipelineId  int   `json:"pipelineId"`
	TriggeredBy int32 `json:"triggeredBy"`
}

type CIPipelineGitWebhookEvent struct {
	GitHostId          int    `json:"gitHostId"`
	EventType          string `json:"eventType"`
	RequestPayloadJson string `json:"requestPayloadJson"`
}
