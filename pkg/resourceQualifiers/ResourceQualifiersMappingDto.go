package resourceQualifiers

import (
	"github.com/devtron-labs/devtron/pkg/sql"
)

type ResourceType int

const (
	Variable                ResourceType = 0
	Filter                  ResourceType = 1
	ImageDigest             ResourceType = 2
	ImageDigestResourceId                = -1 // for ImageDigest resource id will is constant unlike filter and variables
	InfraProfile            ResourceType = 3
	DeploymentWindowProfile ResourceType = 4
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
	ResourceId          int
	ResourceType        ResourceType
	SelectionIdentifier *SelectionIdentifier
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
	sql.AuditLog
}

type QualifierMappingWithExtraColumns struct {
	QualifierMapping
	TotalCount int
}

type ResourceMappingSelection struct {
	ResourceType        ResourceType
	ResourceId          int
	QualifierSelector   QualifierSelector
	SelectionIdentifier *SelectionIdentifier
	Id                  int
}

type SelectionIdentifier struct {
	AppId                   int                      `json:"appId"`
	EnvId                   int                      `json:"envId"`
	ClusterId               int                      `json:"clusterId"`
	SelectionIdentifierName *SelectionIdentifierName `json:"-"`
}

type SelectionIdentifierName struct {
	AppName         string
	EnvironmentName string
	ClusterName     string
}

func (mapping *QualifierMapping) GetIdValueAndName() (int, string) {
	return mapping.IdentifierValueInt, mapping.IdentifierValueString
}
