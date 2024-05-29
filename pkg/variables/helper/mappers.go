/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func GetSelectorForAttributeType(attributeType models.AttributeType) resourceQualifiers.QualifierSelector {
	switch attributeType {
	case models.ApplicationEnv:
		return resourceQualifiers.ApplicationEnvironmentSelector
	case models.Application:
		return resourceQualifiers.ApplicationSelector
	case models.Env:
		return resourceQualifiers.EnvironmentSelector
	case models.Cluster:
		return resourceQualifiers.ClusterSelector
	case models.Global:
		return resourceQualifiers.GlobalSelector
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
