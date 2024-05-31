/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
