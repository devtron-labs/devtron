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

package adapter

import (
	apiBean "github.com/devtron-labs/devtron/api/bean"
	helmBean "github.com/devtron-labs/devtron/api/helm-app/service/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"time"
)

func SetPipelineFieldsInOverrideRequest(overrideRequest *apiBean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline, deploymentConfig *bean2.DeploymentConfig) {
	overrideRequest.PipelineId = pipeline.Id
	overrideRequest.PipelineName = pipeline.Name
	overrideRequest.EnvId = pipeline.EnvironmentId
	overrideRequest.EnvName = pipeline.Environment.Name
	overrideRequest.ClusterId = pipeline.Environment.ClusterId
	overrideRequest.AppId = pipeline.AppId
	overrideRequest.AppName = pipeline.App.AppName
	overrideRequest.DeploymentAppType = deploymentConfig.DeploymentAppType
	overrideRequest.Namespace = pipeline.Environment.Namespace
	overrideRequest.ReleaseName = pipeline.DeploymentAppName
}

func GetVulnerabilityCheckRequest(cdPipeline *pipelineConfig.Pipeline, imageDigest string) *bean.VulnerabilityCheckRequest {
	return &bean.VulnerabilityCheckRequest{
		CdPipeline:  cdPipeline,
		ImageDigest: imageDigest,
	}
}

func NewUserDeploymentRequest(overrideRequest *apiBean.ValuesOverrideRequest, triggeredAt time.Time, triggeredBy int32) *eventProcessorBean.UserDeploymentRequest {
	return &eventProcessorBean.UserDeploymentRequest{
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
