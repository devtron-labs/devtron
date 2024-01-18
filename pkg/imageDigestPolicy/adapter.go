package imageDigestPolicy

import (
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func QualifierMappingDao(qualifierId, identifierKey, IdentifierValueInt int, userId int32) *resourceQualifiers.QualifierMapping {
	return &resourceQualifiers.QualifierMapping{
		ResourceId:         resourceQualifiers.ImageDigestResourceId,
		ResourceType:       resourceQualifiers.ImageDigest,
		QualifierId:        qualifierId,
		IdentifierKey:      identifierKey,
		IdentifierValueInt: IdentifierValueInt,
		Active:             true,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
}
