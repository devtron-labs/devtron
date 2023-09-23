package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/models"
)

type VariableDefinition struct {
	tableName        struct{}            `sql:"variable_definition" pg:",discard_unknown_columns"`
	Id               int                 `sql:"id,pk"`
	Name             string              `sql:"name"`
	DataType         models.DataType     `sql:"data_type"`
	VarType          models.VariableType `sql:"var_type"`
	Active           bool                `sql:"active"`
	Description      string              `sql:"description"`
	ShortDescription string              `json:"short_description"`
	//VariableScope    []*VariableScope
	sql.AuditLog
}

//
//type VariableScope struct {
//	tableName             struct{} `sql:"variable_scope" pg:",discard_unknown_columns"`
//	Id                    int      `sql:"id,pk"`
//	VariableDefinitionId  int      `sql:"variable_definition_id"`
//	QualifierId           int      `sql:"qualifier_id"`
//	IdentifierKey         int      `sql:"identifier_key"`
//	IdentifierValueInt    int      `sql:"identifier_value_int"`
//	Active                bool     `sql:"active"`
//	IdentifierValueString string   `sql:"identifier_value_string"`
//	ParentIdentifier      int      `sql:"parent_identifier"`
//	CompositeKey          string   `sql:"-"`
//	Data                  string   `sql:"-"`
//	VariableData          *VariableData
//	sql.AuditLog
//}

type VariableData struct {
	tableName       struct{} `sql:"variable_data" pg:",discard_unknown_columns"`
	Id              int      `sql:"id,pk"`
	VariableScopeId int      `sql:"variable_scope_id"`
	Data            string   `sql:"data"`
	sql.AuditLog
}

type Qualifier int

const (
	GLOBAL_QUALIFIER Qualifier = 5
)

var CompoundQualifiers []Qualifier

func GetNumOfChildQualifiers(qualifier Qualifier) int {
	return 0
}

func CreateFromDefinition(definition models.Definition, auditLog sql.AuditLog) *VariableDefinition {
	varDefinition := &VariableDefinition{}
	varDefinition.Name = definition.VarName
	varDefinition.DataType = definition.DataType
	varDefinition.VarType = definition.VarType
	varDefinition.Description = definition.Description
	varDefinition.ShortDescription = definition.ShortDescription
	varDefinition.Active = true
	varDefinition.AuditLog = auditLog
	return varDefinition
}
