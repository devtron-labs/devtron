package resourceQualifiers

import (
	"github.com/devtron-labs/devtron/pkg/sql"
)

type ResourceType int

const (
	Variable              ResourceType = 0
	Filter                ResourceType = 1
	ImageDigest           ResourceType = 2
	ImageDigestResourceId              = -1 // for ImageDigest resource id will is constant unlike filter and variables
	InfraProfile          ResourceType = 3
	ImagePromotionPolicy  ResourceType = 4
)

const (
	AllProjectsValue                     = "-1"
	AllProjectsInt                       = -1
	AllExistingAndFutureProdEnvsValue    = "-2"
	AllExistingAndFutureProdEnvsInt      = -2
	AllExistingAndFutureNonProdEnvsValue = "-1"
	AllExistingAndFutureNonProdEnvsInt   = -1
	AllExistingAndFutureEnvsString       = "-3"
	AllExistingAndFutureEnvsInt          = -3
)

func GetEnvIdentifierValue(scope Scope) int {
	if scope.IsProdEnv {
		return AllExistingAndFutureProdEnvsInt
	}
	return AllExistingAndFutureNonProdEnvsInt
}

type ResourceQualifierMappings struct {
	ResourceId   int
	ResourceType ResourceType
	Scope        *Scope
	// qualifierSelector QualifierSelector
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
	// Data                  string   `sql:"-"`
	// VariableData          *VariableData
	sql.AuditLog
}

type QualifierMappingWithExtraColumns struct {
	QualifierMapping
	TotalCount int
}

type ResourceMappingSelection struct {
	ResourceType      ResourceType
	ResourceId        int
	QualifierSelector QualifierSelector
	Scope             *Scope
	Id                int
}
