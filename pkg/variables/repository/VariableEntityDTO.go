package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
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
