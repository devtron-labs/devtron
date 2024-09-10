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
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type SelfRegistrationRolesRepository interface {
	GetAll() ([]SelfRegistrationRoles, error)
}

type SelfRegistrationRoles struct {
	TableName struct{} `sql:"self_registration_roles" pg:",discard_unknown_columns"`
	Role      string   `sql:"role,notnull"`
	sql.AuditLog
}

type SelfRegistrationRolesRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewSelfRegistrationRolesRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *SelfRegistrationRolesRepositoryImpl {
	return &SelfRegistrationRolesRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *SelfRegistrationRolesRepositoryImpl) GetAll() ([]SelfRegistrationRoles, error) {
	var models []SelfRegistrationRoles
	err := impl.dbConnection.Model(&models).Select()
	if err != nil {
		impl.logger.Error(err)
		return models, err
	}
	return models, nil
}
