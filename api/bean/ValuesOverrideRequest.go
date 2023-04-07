/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package bean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/models"
)

type WorkflowType string
type DeploymentConfigurationType string

const (
	CD_WORKFLOW_TYPE_PRE              WorkflowType                = "PRE"
	CD_WORKFLOW_TYPE_POST             WorkflowType                = "POST"
	CD_WORKFLOW_TYPE_DEPLOY           WorkflowType                = "DEPLOY"
	CI_WORKFLOW_TYPE                  WorkflowType                = "CI"
	WEBHOOK_WORKFLOW_TYPE             WorkflowType                = "WEBHOOK"
	DEPLOYMENT_CONFIG_TYPE_LAST_SAVED DeploymentConfigurationType = "LAST_SAVED_CONFIG"
	//latest trigger is not being used because this is being handled at FE and we anyhow identify latest trigger as
	//last deployed wfr which is also a specific trigger
	DEPLOYMENT_CONFIG_TYPE_LATEST_TRIGGER   DeploymentConfigurationType = "LATEST_TRIGGER_CONFIG"
	DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER DeploymentConfigurationType = "SPECIFIC_TRIGGER_CONFIG"
)

type ValuesOverrideRequest struct {
	PipelineId                            int                         `json:"pipelineId" validate:"required"`
	AppId                                 int                         `json:"appId" validate:"required"`
	CiArtifactId                          int                         `json:"ciArtifactId" validate:"required"`
	AdditionalOverride                    json.RawMessage             `json:"additionalOverride,omitempty"`
	TargetDbVersion                       int                         `json:"targetDbVersion"`
	ForceTrigger                          bool                        `json:"forceTrigger,notnull"`
	DeploymentTemplate                    string                      `json:"strategy,omitempty"` // validate:"oneof=BLUE-GREEN ROLLING"`
	DeploymentWithConfig                  DeploymentConfigurationType `json:"deploymentWithConfig"`
	WfrIdForDeploymentWithSpecificTrigger int                         `json:"wfrIdForDeploymentWithSpecificTrigger"`
	CdWorkflowType                        WorkflowType                `json:"cdWorkflowType,notnull"`
	CdWorkflowId                          int                         `json:"cdWorkflowId"`
	UserId                                int32                       `json:"-"`
	DeploymentType                        models.DeploymentType       `json:"-"`
}

type BulkCdDeployEvent struct {
	ValuesOverrideRequest *ValuesOverrideRequest `json:"valuesOverrideRequest"`
	UserId                int32                  `json:"userId"`
}

type ReleaseStatusUpdateRequest struct {
	RequestId string             `json:"requestId"`
	NewStatus models.ChartStatus `json:"newStatus"`
}
