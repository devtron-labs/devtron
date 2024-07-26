/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	GitHostName        string `json:"gitHostName"`
	EventType          string `json:"eventType"`
	RequestPayloadJson string `json:"requestPayloadJson"`
}
