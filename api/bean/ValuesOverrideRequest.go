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
	"github.com/devtron-labs/devtron/internal/sql/models"
	"encoding/json"
)

type CdWorkflowType string

const CD_WORKFLOW_TYPE_PRE CdWorkflowType = "PRE"
const CD_WORKFLOW_TYPE_POST CdWorkflowType = "POST"
const CD_WORKFLOW_TYPE_DEPLOY CdWorkflowType = "DEPLOY"

type ValuesOverrideRequest struct {
	PipelineId         int                   `json:"pipelineId" validate:"required"`
	AppId              int                   `json:"appId" validate:"required"`
	CiArtifactId       int                   `json:"ciArtifactId" validate:"required"`
	AdditionalOverride json.RawMessage       `json:"additionalOverride"`
	TargetDbVersion    int                   `json:"targetDbVersion"`
	ForceTrigger       bool                  `json:"forceTrigger,notnull"`
	DeploymentTemplate string                `json:"strategy,omitempty"` // validate:"oneof=BLUE-GREEN ROLLING"`
	CdWorkflowType     CdWorkflowType        `json:"cdWorkflowType,notnull"`
	CdWorkflowId       int                   `json:"cdWorkflowId"`
	UserId             int32                 `json:"-"`
	DeploymentType     models.DeploymentType `json:"-"`
}

type ReleaseStatusUpdateRequest struct {
	RequestId string             `json:"requestId"`
	NewStatus models.ChartStatus `json:"newStatus"`
}
