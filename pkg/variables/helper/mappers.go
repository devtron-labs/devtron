package helper

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables/models"
)

func GetIdentifierTypeFromAttributeType(attribute models.AttributeType) []models.IdentifierType {
	switch attribute {
	case models.ApplicationEnv:
		return []models.IdentifierType{models.ApplicationName, models.EnvName}
	case models.Application:
		return []models.IdentifierType{models.ApplicationName}
	case models.Env:
		return []models.IdentifierType{models.EnvName}
	case models.Cluster:
		return []models.IdentifierType{models.ClusterName}
	default:
		return nil
	}
}

func GetIdentifierTypeFromResourceKey(searchableKeyId int, resourceKeyMap map[int]bean.DevtronResourceSearchableKeyName) models.IdentifierType {
	switch resourceKeyMap[searchableKeyId] {
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID:
		return models.ApplicationName
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID:
		return models.EnvName
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID:
		return models.ClusterName
	default:
		return ""
	}
}

func GetIdentifierKey(identifierType models.IdentifierType, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) int {
	switch identifierType {
	case models.ApplicationName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID]
	case models.ClusterName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	case models.EnvName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	default:
		return 0
	}
}

func GetQualifierId(attributeType models.AttributeType) resourceQualifiers.Qualifier {
	switch attributeType {
	case models.ApplicationEnv:
		return resourceQualifiers.APP_AND_ENV_QUALIFIER
	case models.Application:
		return resourceQualifiers.APP_QUALIFIER
	case models.Env:
		return resourceQualifiers.ENV_QUALIFIER
	case models.Cluster:
		return resourceQualifiers.CLUSTER_QUALIFIER
	case models.Global:
		return resourceQualifiers.GLOBAL_QUALIFIER
	default:
		return 0
	}
}

func GetAttributeType(qualifier resourceQualifiers.Qualifier) models.AttributeType {
	switch qualifier {
	case resourceQualifiers.APP_AND_ENV_QUALIFIER:
		return models.ApplicationEnv
	case resourceQualifiers.APP_QUALIFIER:
		return models.Application
	case resourceQualifiers.ENV_QUALIFIER:
		return models.Env
	case resourceQualifiers.CLUSTER_QUALIFIER:
		return models.Cluster
	case resourceQualifiers.GLOBAL_QUALIFIER:
		return models.Global
	default:
		return ""
	}
}
