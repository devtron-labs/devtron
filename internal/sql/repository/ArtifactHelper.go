/*
 * Copyright (c) 2024. Devtron Inc.
 */

package repository

func (artifact *CiArtifact) CopyArtifactMetadata(pipelineId int, userId int32) *CiArtifact {
	if artifact == nil {
		return nil
	}
	copiedArtifact := &CiArtifact{
		Image:                 artifact.Image,
		ImageDigest:           artifact.ImageDigest,
		MaterialInfo:          artifact.MaterialInfo,
		DataSource:            artifact.DataSource,
		ScanEnabled:           artifact.ScanEnabled,
		Scanned:               artifact.Scanned,
		IsArtifactUploaded:    artifact.IsArtifactUploaded,
		ParentCiArtifact:      artifact.Id,
		CredentialsSourceType: artifact.CredentialsSourceType,
		CredentialSourceValue: artifact.CredentialSourceValue,
		PipelineId:            pipelineId,
	}
	copiedArtifact.CreateAuditLog(userId)
	if artifact.ParentCiArtifact > 0 {
		copiedArtifact.ParentCiArtifact = artifact.ParentCiArtifact
	}
	if artifact.ExternalCiPipelineId > 0 {
		copiedArtifact.ExternalCiPipelineId = artifact.ExternalCiPipelineId
	}
	return copiedArtifact
}
