package imageDigestPolicy

import (
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func QualifierMappingDao(qualifierId, identifierKey, IdentifierValueInt int, identifierValueName string, userId int32) *resourceQualifiers.QualifierMapping {
	return &resourceQualifiers.QualifierMapping{
		ResourceId:            resourceQualifiers.ImageDigestResourceId,
		ResourceType:          resourceQualifiers.ImageDigest,
		QualifierId:           qualifierId,
		IdentifierKey:         identifierKey,
		IdentifierValueInt:    IdentifierValueInt,
		IdentifierValueString: identifierValueName,
		Active:                true,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
}

type DigestPolicyConfigurationRequest struct {
	PipelineId    int
	ClusterId     int
	EnvironmentId int
}

func (request DigestPolicyConfigurationRequest) getQualifierMappingScope() *resourceQualifiers.Scope {
	return &resourceQualifiers.Scope{
		EnvId:      request.EnvironmentId,
		ClusterId:  request.ClusterId,
		PipelineId: request.PipelineId,
	}
}

type DigestPolicyConfigurationResponse struct {
	DigestConfiguredForPipeline     bool
	DigestConfiguredForEnvOrCluster bool
}

func (config DigestPolicyConfigurationResponse) UseDigestForTrigger() bool {
	return config.DigestConfiguredForPipeline
}
