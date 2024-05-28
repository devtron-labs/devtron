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
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"time"
)

func ConvertCdWorkflowRunnerDbObjToDto(dbObj *pipelineConfig.CdWorkflowRunner) *bean.CdWorkflowRunnerDto {
	newReferenceId := ""
	if dbObj.ReferenceId != nil {
		newReferenceId = *dbObj.ReferenceId
	}
	return &bean.CdWorkflowRunnerDto{
		Id:                      dbObj.Id,
		Name:                    dbObj.Name,
		WorkflowType:            dbObj.WorkflowType,
		ExecutorType:            dbObj.ExecutorType,
		Status:                  dbObj.Status,
		PodStatus:               dbObj.PodStatus,
		Message:                 dbObj.Message,
		StartedOn:               dbObj.StartedOn,
		FinishedOn:              dbObj.FinishedOn,
		Namespace:               dbObj.Namespace,
		LogLocation:             dbObj.LogLocation,
		TriggeredBy:             dbObj.TriggeredBy,
		CdWorkflowId:            dbObj.CdWorkflowId,
		PodName:                 dbObj.PodName,
		BlobStorageEnabled:      dbObj.BlobStorageEnabled,
		RefCdWorkflowRunnerId:   dbObj.RefCdWorkflowRunnerId,
		ImagePathReservationIds: dbObj.ImagePathReservationIds,
		ReferenceId:             &newReferenceId,
	}
}

func ConvertCdWorkflowRunnerDtoToDbObj(dto *bean.CdWorkflowRunnerDto) *pipelineConfig.CdWorkflowRunner {
	return &pipelineConfig.CdWorkflowRunner{
		Id:                      dto.Id,
		Name:                    dto.Name,
		WorkflowType:            dto.WorkflowType,
		ExecutorType:            dto.ExecutorType,
		Status:                  dto.Status,
		PodStatus:               dto.PodStatus,
		Message:                 dto.Message,
		StartedOn:               dto.StartedOn,
		Namespace:               dto.Namespace,
		LogLocation:             dto.LogLocation,
		TriggeredBy:             dto.TriggeredBy,
		CdWorkflowId:            dto.CdWorkflowId,
		PodName:                 dto.PodName,
		BlobStorageEnabled:      dto.BlobStorageEnabled,
		RefCdWorkflowRunnerId:   dto.RefCdWorkflowRunnerId,
		ImagePathReservationIds: dto.ImagePathReservationIds,
		ReferenceId:             dto.ReferenceId,
		AuditLog: sql.AuditLog{
			CreatedOn: dto.StartedOn,
			CreatedBy: dto.TriggeredBy,
			UpdatedOn: dto.StartedOn,
			UpdatedBy: dto.TriggeredBy,
		},
	}
}

func ConvertCdWorkflowDtoToDbObj(dto *bean.CdWorkflowDto) *pipelineConfig.CdWorkflow {
	return &pipelineConfig.CdWorkflow{
		Id:             dto.Id,
		CiArtifactId:   dto.CiArtifactId,
		PipelineId:     dto.PipelineId,
		WorkflowStatus: dto.WorkflowStatus,
		AuditLog: sql.AuditLog{ //not handling for creation auditLog currently
			UpdatedOn: time.Now(),
			UpdatedBy: dto.UserId,
		},
	}
}

func ConvertCdWorkflowDtoToDbObjForCreation(dto *bean.CdWorkflowDto, timeCreated time.Time) *pipelineConfig.CdWorkflow {
	return &pipelineConfig.CdWorkflow{
		Id:             dto.Id,
		CiArtifactId:   dto.CiArtifactId,
		PipelineId:     dto.PipelineId,
		WorkflowStatus: dto.WorkflowStatus,
		AuditLog: sql.AuditLog{
			CreatedOn: timeCreated,
			CreatedBy: dto.UserId,
			UpdatedOn: timeCreated,
			UpdatedBy: dto.UserId,
		},
	}
}

func BuildCdWorkflowDto(pipelineId, ciArtifactId int, userId int32) *bean.CdWorkflowDto {
	return &bean.CdWorkflowDto{PipelineId: pipelineId, CiArtifactId: ciArtifactId, UserId: userId}
}

func BuildCdWorkflowRunnerDto(name string, workflowType bean2.WorkflowType, executorType pipelineConfig.WorkflowExecutorType, status string, triggeredBy int32, startedOn time.Time, namespace string, cdWorkflowId int, blobStorageEnabled bool, logLocation string) *bean.CdWorkflowRunnerDto {
	return &bean.CdWorkflowRunnerDto{
		Name:               name,
		WorkflowType:       workflowType,
		ExecutorType:       executorType,
		Status:             status,
		TriggeredBy:        triggeredBy,
		StartedOn:          startedOn,
		Namespace:          namespace,
		CdWorkflowId:       cdWorkflowId,
		BlobStorageEnabled: blobStorageEnabled,
		LogLocation:        logLocation,
	}
}
