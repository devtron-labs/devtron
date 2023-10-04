package helper

import (
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables/models"
)

func GetQualifierId(attributeType models.AttributeType) resourceQualifiers.Qualifier {
	switch attributeType {
	case models.Global:
		return resourceQualifiers.GLOBAL_QUALIFIER
	default:
		return 0
	}
}

func GetAttributeType(qualifier resourceQualifiers.Qualifier) models.AttributeType {
	switch qualifier {
	case resourceQualifiers.GLOBAL_QUALIFIER:
		return models.Global
	default:
		return ""
	}
}

func GetIdentifierTypeFromAttributeType(attribute models.AttributeType) []models.IdentifierType {
	switch attribute {
	default:
		return nil
	}
}
