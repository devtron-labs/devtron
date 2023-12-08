package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func (artifact *CiArtifact) CopyArtifactMetadata(userId int32) *CiArtifact {
	copiedArtifact := &CiArtifact{
		Image:              artifact.Image,
		ImageDigest:        artifact.ImageDigest,
		MaterialInfo:       artifact.MaterialInfo,
		DataSource:         artifact.DataSource,
		ScanEnabled:        artifact.ScanEnabled,
		Scanned:            artifact.Scanned,
		IsArtifactUploaded: artifact.IsArtifactUploaded,
		ParentCiArtifact:   artifact.Id,
		AuditLog:           sql.AuditLog{CreatedBy: userId, UpdatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
	}
	if artifact.ParentCiArtifact > 0 {
		copiedArtifact.ParentCiArtifact = artifact.ParentCiArtifact
	}
	if artifact.ExternalCiPipelineId > 0 {
		copiedArtifact.ExternalCiPipelineId = artifact.ExternalCiPipelineId
	}
	return copiedArtifact
}
