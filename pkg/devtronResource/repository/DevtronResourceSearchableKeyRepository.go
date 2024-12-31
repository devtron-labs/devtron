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
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronResourceSearchableKeyRepository interface {
	GetAll() ([]*DevtronResourceSearchableKey, error)
}

type DevtronResourceSearchableKeyRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDevtronResourceSearchableKeyRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DevtronResourceSearchableKeyRepositoryImpl {
	return &DevtronResourceSearchableKeyRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DevtronResourceSearchableKey struct {
	tableName struct{}                              `sql:"devtron_resource_searchable_key" pg:",discard_unknown_columns"`
	Id        int                                   `sql:"id,pk"`
	Name      bean.DevtronResourceSearchableKeyName `sql:"name"`
	IsRemoved bool                                  `sql:"is_removed,notnull"`
	sql.AuditLog
}

func (repo *DevtronResourceSearchableKeyRepositoryImpl) GetAll() ([]*DevtronResourceSearchableKey, error) {
	var models []*DevtronResourceSearchableKey
	err := repo.dbConnection.Model(&models).Where("is_removed = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting all devtron resources searchable key", "err", err)
		return nil, err
	}
	return models, nil
}
