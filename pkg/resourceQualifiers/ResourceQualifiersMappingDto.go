package resourceQualifiers

import "github.com/devtron-labs/devtron/pkg/sql"

type ResourceType int

const (
	Variable ResourceType = 0
	Filter                = 1
)

type QualifierSelector int

const (
	ApplicationSelector                  QualifierSelector = 0
	EnvironmentSelectorQualifierSelector                   = 1
)

const (
	AllProjectsValue                     = "-1"
	AllProjectsInt                       = -1
	AllExistingAndFutureProdEnvsValue    = "-2"
	AllExistingAndFutureProdEnvsInt      = -2
	AllExistingAndFutureNonProdEnvsValue = "-1"
	AllExistingAndFutureNonProdEnvsInt   = -1
)

func GetEnvIdentifierValue(scope Scope) int {
	if scope.IsProdEnv {
		return AllExistingAndFutureProdEnvsInt
	}
	return AllExistingAndFutureNonProdEnvsInt
}

type QualifierMapping struct {
	tableName             struct{}     `sql:"resource_qualifier_mapping" pg:",discard_unknown_columns"`
	Id                    int          `sql:"id,pk"`
	ResourceId            int          `sql:"resource_id"`
	ResourceType          ResourceType `sql:"resource_type"`
	QualifierId           int          `sql:"qualifier_id"`
	IdentifierKey         int          `sql:"identifier_key"`
	IdentifierValueInt    int          `sql:"identifier_value_int"`
	Active                bool         `sql:"active"`
	IdentifierValueString string       `sql:"identifier_value_string"`
	ParentIdentifier      int          `sql:"parent_identifier"`
	CompositeKey          string       `sql:"-"`
	//Data                  string   `sql:"-"`
	//VariableData          *VariableData
	sql.AuditLog
}
