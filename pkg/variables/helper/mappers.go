package helper

import (
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
)

func GetQualifierId(attributeType models.AttributeType) repository.Qualifier {
	switch attributeType {
	case models.Global:
		return repository.GLOBAL_QUALIFIER
	default:
		return 0
	}
}

func GetAttributeType(qualifier repository.Qualifier) models.AttributeType {
	switch qualifier {
	case repository.GLOBAL_QUALIFIER:
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
