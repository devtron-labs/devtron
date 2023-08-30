package utils

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
)

func GetPriority(qualifier repository.Qualifier) int {
	switch qualifier {
	case repository.APP_AND_ENV_QUALIFIER:
		return 1
	case repository.APP_QUALIFIER:
		return 2
	case repository.ENV_QUALIFIER:
		return 3
	case repository.CLUSTER_QUALIFIER:
		return 4
	case repository.GLOBAL_QUALIFIER:
		return 5
	default:
		return 0
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

func GetQualifierId(attributeType models.AttributeType) repository.Qualifier {
	switch attributeType {
	case models.ApplicationEnv:
		return repository.APP_AND_ENV_QUALIFIER
	case models.Application:
		return repository.APP_QUALIFIER
	case models.Env:
		return repository.ENV_QUALIFIER
	case models.Cluster:
		return repository.CLUSTER_QUALIFIER
	case models.Global:
		return repository.GLOBAL_QUALIFIER
	default:
		return 0
	}
}

func GetAttributeType(qualifier repository.Qualifier) models.AttributeType {
	switch qualifier {
	case repository.APP_AND_ENV_QUALIFIER:
		return models.ApplicationEnv
	case repository.APP_QUALIFIER:
		return models.Application
	case repository.ENV_QUALIFIER:
		return models.Env
	case repository.CLUSTER_QUALIFIER:
		return models.Cluster
	case repository.GLOBAL_QUALIFIER:
		return models.Global
	default:
		return ""
	}
}
