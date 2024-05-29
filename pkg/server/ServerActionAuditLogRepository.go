/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package server

import (
	"github.com/go-pg/pg"
	"time"
)

type ServerActionAuditLog struct {
	tableName struct{}  `sql:"server_action_audit_log"`
	Id        int       `sql:"id,pk"`
	Action    string    `sql:"action,notnull"`
	Version   string    `sql:"version"`
	CreatedOn time.Time `sql:"created_on,notnull"`
	CreatedBy int32     `sql:"created_by,notnull"`
}

type ServerActionAuditLogRepository interface {
	Save(serverActionAuditLog *ServerActionAuditLog) error
}

type ServerActionAuditLogRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewServerActionAuditLogRepositoryImpl(dbConnection *pg.DB) *ServerActionAuditLogRepositoryImpl {
	return &ServerActionAuditLogRepositoryImpl{dbConnection: dbConnection}
}

func (impl ServerActionAuditLogRepositoryImpl) Save(serverActionAuditLog *ServerActionAuditLog) error {
	err := impl.dbConnection.Insert(serverActionAuditLog)
	return err
}
