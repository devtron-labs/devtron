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

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
)

type VariableEntityMapping struct {
	tableName    struct{} `sql:"variable_entity_mapping" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	VariableName string   `sql:"variable_name"`
	IsDeleted    bool     `sql:"is_deleted,notnull"`
	Entity
	sql.AuditLog
}

type Entity struct {
	EntityType EntityType `sql:"entity_type"`
	EntityId   int        `sql:"entity_id"`
}

type EntityType int

const (
	EntityTypeDeploymentTemplateAppLevel EntityType = 1
	EntityTypeDeploymentTemplateEnvLevel EntityType = 2
	EntityTypePipelineStage              EntityType = 3
	EntityTypeConfigMapAppLevel          EntityType = 4
	EntityTypeConfigMapEnvLevel          EntityType = 5
	EntityTypeSecretAppLevel             EntityType = 6
	EntityTypeSecretEnvLevel             EntityType = 7
)

func GetEntity(entityId int, entityType EntityType) Entity {

	return Entity{
		EntityType: entityType,
		EntityId:   entityId,
	}
}

func CollectVariables(entityToVariables map[Entity][]string, entities []Entity) []string {
	varNames := make([]string, 0)

	for _, entity := range entities {
		if names, ok := entityToVariables[entity]; ok {
			varNames = append(varNames, names...)
		}
	}
	return utils.FilterDuplicatesInStringArray(varNames)
}
