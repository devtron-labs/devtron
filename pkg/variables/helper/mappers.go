package helper

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
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
