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
