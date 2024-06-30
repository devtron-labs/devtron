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
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"time"
)

type CdWorkflowDto struct {
	Id             int                           `json:"id"`
	CiArtifactId   int                           `json:"ciArtifactId"`
	PipelineId     int                           `json:"pipelineId"`
	WorkflowStatus pipelineConfig.WorkflowStatus `json:"workflowStatus"`
	UserId         int32                         `json:"-"`
}

type CdWorkflowRunnerDto struct {
	Id           int               `json:"id"`
	Name         string            `json:"name"`
	WorkflowType bean.WorkflowType `json:"workflowType"` // pre,post,deploy
	//TODO: extract from repo service layer
	ExecutorType            pipelineConfig.WorkflowExecutorType `json:"executorType"` // awf, system
	Status                  string                              `json:"status"`
	PodStatus               string                              `json:"podStatus"`
	Message                 string                              `json:"message"`
	StartedOn               time.Time                           `json:"startedOn"`
	FinishedOn              time.Time                           `json:"finishedOn"`
	Namespace               string                              `json:"namespace"`
	LogLocation             string                              `json:"logFilePath"`
	TriggeredBy             int32                               `json:"triggeredBy"`
	CdWorkflowId            int                                 `json:"cdWorkflowId"`
	PodName                 string                              `json:"podName"`
	BlobStorageEnabled      bool                                `json:"blobStorageEnabled"`
	RefCdWorkflowRunnerId   int                                 `json:"refCdWorkflowRunnerId"`
	ImagePathReservationIds []int                               `json:"imagePathReservationIds"`
	ReferenceId             *string                             `json:"referenceId"`
}
