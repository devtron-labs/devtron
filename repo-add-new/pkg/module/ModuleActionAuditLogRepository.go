/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package module

import (
	"github.com/go-pg/pg"
	"time"
)

type ModuleActionAuditLog struct {
	tableName  struct{}  `sql:"module_action_audit_log"`
	Id         int       `sql:"id,pk"`
	ModuleName string    `sql:"module_name, notnull"`
	Action     string    `sql:"action,notnull"`
	Version    string    `sql:"version,notnull"`
	CreatedOn  time.Time `sql:"created_on,notnull"`
	CreatedBy  int32     `sql:"created_by,notnull"`
}

type ModuleActionAuditLogRepository interface {
	Save(moduleActionAuditLog *ModuleActionAuditLog) error
}

type ModuleActionAuditLogRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewModuleActionAuditLogRepositoryImpl(dbConnection *pg.DB) *ModuleActionAuditLogRepositoryImpl {
	return &ModuleActionAuditLogRepositoryImpl{dbConnection: dbConnection}
}

func (impl ModuleActionAuditLogRepositoryImpl) Save(moduleActionAuditLog *ModuleActionAuditLog) error {
	err := impl.dbConnection.Insert(moduleActionAuditLog)
	return err
}
