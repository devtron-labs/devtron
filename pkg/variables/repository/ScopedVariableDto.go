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
	sql.AuditLog
}

type VariableData struct {
	tableName       struct{} `sql:"variable_data" pg:",discard_unknown_columns"`
	Id              int      `sql:"id,pk"`
	VariableScopeId int      `sql:"variable_scope_id"`
	Data            string   `sql:"data"`
	sql.AuditLog
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
