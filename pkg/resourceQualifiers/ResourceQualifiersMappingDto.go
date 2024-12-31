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

package resourceQualifiers

import "github.com/devtron-labs/devtron/pkg/sql"

type ResourceType int

const (
	Variable              ResourceType = 0
	Filter                             = 1
	ImageDigest                        = 2
	ImageDigestResourceId              = -1 // for ImageDigest resource id will is constant unlike filter and variables
	InfraProfile                       = 3
	ImagePromotionPolicy  ResourceType = 4
	DeploymentWindow      ResourceType = 5
)

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
	// Data                  string   `sql:"-"`
	// VariableData          *VariableData
	sql.AuditLog
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
