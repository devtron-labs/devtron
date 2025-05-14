/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package adapter

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
)

func NewCDPipelineMinConfigFromModel(pipeline *pipelineConfig.Pipeline) *bean3.CDPipelineMinConfig {
	deploymentConfigMin := &bean3.CDPipelineMinConfig{
		Id:                         pipeline.Id,
		Name:                       pipeline.Name,
		CiPipelineId:               pipeline.CiPipelineId,
		EnvironmentId:              pipeline.EnvironmentId,
		AppId:                      pipeline.AppId,
		DeploymentAppDeleteRequest: pipeline.DeploymentAppDeleteRequest,
		DeploymentAppCreated:       pipeline.DeploymentAppCreated,
		DeploymentAppType:          pipeline.DeploymentAppType,

		// pipeline.App is not of pointer type
		AppName: pipeline.App.AppName,
		TeamId:  pipeline.App.TeamId,

		// pipeline.Environment is not of pointer type
		EnvironmentName:       pipeline.Environment.Name,
		EnvironmentIdentifier: pipeline.Environment.EnvironmentIdentifier,
		Namespace:             pipeline.Environment.Namespace,
		IsProdEnv:             pipeline.Environment.Default,
	}
	return deploymentConfigMin
}
