package adapter

import (
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
