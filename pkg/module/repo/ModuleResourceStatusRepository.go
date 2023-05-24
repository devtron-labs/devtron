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

package moduleRepo

import (
	"github.com/go-pg/pg"
	"time"
)

type ModuleResourceStatus struct {
	tableName     struct{}  `sql:"module_resource_status"`
	Id            int       `sql:"id,pk"`
	ModuleId      int       `sql:"module_id,notnull"`
	Group         string    `sql:"group, notnull"`
	Version       string    `sql:"version, notnull"`
	Kind          string    `sql:"kind, notnull"`
	Name          string    `sql:"name, notnull"`
	HealthStatus  string    `sql:"health_status"`
	HealthMessage string    `sql:"health_message"`
	Active        bool      `sql:"active"`
	CreatedOn     time.Time `sql:"created_on, notnull"`
	UpdatedOn     time.Time `sql:"updated_on"`
}

type ModuleResourceStatusRepository interface {
	GetConnection() *pg.DB
	FindAllActiveByModuleId(moduleId int) ([]*ModuleResourceStatus, error)
	Update(status *ModuleResourceStatus, tx *pg.Tx) error
	Save(statuses []*ModuleResourceStatus, tx *pg.Tx) error
}

type ModuleResourceStatusRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewModuleResourceStatusRepositoryImpl(dbConnection *pg.DB) *ModuleResourceStatusRepositoryImpl {
	return &ModuleResourceStatusRepositoryImpl{dbConnection: dbConnection}
}

func (impl ModuleResourceStatusRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl ModuleResourceStatusRepositoryImpl) FindAllActiveByModuleId(moduleId int) ([]*ModuleResourceStatus, error) {
	var moduleResourcesStatus []*ModuleResourceStatus
	err := impl.dbConnection.Model(&moduleResourcesStatus).
		Where("module_id = ?", moduleId).
		Where("active = ?", true).
		Select()
	return moduleResourcesStatus, err
}

func (impl ModuleResourceStatusRepositoryImpl) Update(status *ModuleResourceStatus, tx *pg.Tx) error {
	return tx.Update(status)
}

func (impl ModuleResourceStatusRepositoryImpl) Save(statuses []*ModuleResourceStatus, tx *pg.Tx) error {
	return tx.Insert(&statuses)
}
