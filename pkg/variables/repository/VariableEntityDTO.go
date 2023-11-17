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
