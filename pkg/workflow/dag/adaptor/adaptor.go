package adaptor

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	bean2 "github.com/devtron-labs/devtron/pkg/workflow/dag/bean"
	"time"
)

func GetBuildArtifact(request *bean2.CiArtifactWebhookRequest, ciPipelineId int, materialJson []byte, createdOn, updatedOn time.Time) *repository.CiArtifact {
	return &repository.CiArtifact{
		Image:              request.Image,
		ImageDigest:        request.ImageDigest,
		MaterialInfo:       string(materialJson),
		DataSource:         request.DataSource,
		PipelineId:         ciPipelineId,
		WorkflowId:         request.WorkflowId,
		ScanEnabled:        request.IsScanEnabled,
		IsArtifactUploaded: request.IsArtifactUploaded, // for backward compatibility
		Scanned:            false,
		TargetPlatforms:    utils.ConvertTargetPlatformListToString(request.TargetPlatforms),
		AuditLog:           sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: createdOn, UpdatedOn: updatedOn},
	}
}
