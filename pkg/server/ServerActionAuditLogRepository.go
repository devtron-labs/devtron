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
