/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository4 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/repository"
	"time"
)

// CiTriggerAuditRequest represents the request for CI trigger audit
type CiTriggerAuditRequest struct {
	WorkflowId int                        `json:"workflowId"`
	Pipeline   *pipelineConfig.CiPipeline `json:"pipeline"`
	*CommonAuditRequest
}

// CdTriggerAuditRequest represents the request for CD trigger audit (Pre-CD, Post-CD)
type CdTriggerAuditRequest struct {
	WorkflowRunnerId int                      `json:"workflowRunnerId"`
	Pipeline         *pipelineConfig.Pipeline `json:"pipeline"`
	Environment      *repository4.Environment `json:"environment"`
	*CommonAuditRequest
}
type CommonAuditRequest struct {
	WorkflowRequest             *types.WorkflowRequest `json:"workflowRequest"`
	ArtifactId                  int                    `json:"artifactId"`
	TriggerType                 string                 `json:"triggerType"` // MANUAL, AUTO, WEBHOOK
	TriggeredBy                 int32                  `json:"triggeredBy"`
	InfraConfigTriggerHistoryId int                    `json:"infraConfigTriggerHistoryId"`
}

// WorkflowTriggerAuditResponse represents the response for workflow trigger audit
type WorkflowTriggerAuditResponse struct {
	Id              int                                `json:"id"`
	WorkflowId      int                                `json:"workflowId"`
	WorkflowType    string                             `json:"workflowType"`
	PipelineId      int                                `json:"pipelineId"`
	AppId           int                                `json:"appId"`
	EnvironmentId   int                                `json:"environmentId"`
	ArtifactId      int                                `json:"artifactId"`
	TriggerType     string                             `json:"triggerType"`
	TriggeredBy     int32                              `json:"triggeredBy"`
	TriggerMetadata string                             `json:"triggerMetadata"`
	Status          string                             `json:"status"`
	CreatedOn       time.Time                          `json:"createdOn"`
	ConfigSnapshot  *repository.WorkflowConfigSnapshot `json:"configSnapshot"`
}

//
//// RetriggerWorkflowConfig represents the configuration needed for retrigger
//type RetriggerWorkflowConfig struct {
//	AuditId             int                                `json:"auditId"`
//	WorkflowType        string                             `json:"workflowType"`
//	PipelineId          int                                `json:"pipelineId"`
//	AppId               int                                `json:"appId"`
//	EnvironmentId       *int                               `json:"environmentId"`
//	ArtifactId          *int                               `json:"artifactId"`
//	WorkflowRequest     *types.WorkflowRequest             `json:"workflowRequest"`
//	ConfigSnapshot      *repository.WorkflowConfigSnapshot `json:"configSnapshot"`
//	OriginalTriggeredBy int32                              `json:"originalTriggeredBy"`
//	OriginalTriggerTime time.Time                          `json:"originalTriggerTime"`
//}
