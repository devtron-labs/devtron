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

package appGroup

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type AppGroup struct {
	tableName     struct{} `sql:"app_group" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	Name          string   `sql:"name,notnull"`
	Description   string   `sql:"description,notnull"`
	Active        bool     `sql:"active, notnull"`
	EnvironmentId int      `sql:"environment_id,notnull"`
	sql.AuditLog
}

type AppGroupRepository interface {
	Save(model *AppGroup, tx *pg.Tx) (*AppGroup, error)
	Update(model *AppGroup, tx *pg.Tx) error
	FindById(id int) (*AppGroup, error)
	FindByNameAndEnvId(name string, envId int) (*AppGroup, error)
	FindActiveListByEnvId(envId int) ([]*AppGroup, error)
	GetConnection() (dbConnection *pg.DB)
}

type AppGroupRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewAppGroupRepositoryImpl(dbConnection *pg.DB) *AppGroupRepositoryImpl {
	return &AppGroupRepositoryImpl{dbConnection: dbConnection}
}

func (repo *AppGroupRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return repo.dbConnection
}

func (repo AppGroupRepositoryImpl) Save(model *AppGroup, tx *pg.Tx) (*AppGroup, error) {
	err := tx.Insert(model)
	return model, err
}

func (repo AppGroupRepositoryImpl) Update(model *AppGroup, tx *pg.Tx) error {
	err := tx.Update(model)
	return err
}

func (repo AppGroupRepositoryImpl) FindById(id int) (*AppGroup, error) {
	model := &AppGroup{}
	err := repo.dbConnection.Model(model).Where("id = ?", id).Where("active = ?", true).
		Select()
	return model, err
}

func (repo AppGroupRepositoryImpl) FindByNameAndEnvId(name string, envId int) (*AppGroup, error) {
	model := &AppGroup{}
	err := repo.dbConnection.Model(model).
		Where("name = ?", name).Where("environment_id=?", envId).
		Where("active = ?", true).
		Select()
	return model, err
}

func (repo AppGroupRepositoryImpl) FindActiveListByEnvId(envId int) ([]*AppGroup, error) {
	var models []*AppGroup
	err := repo.dbConnection.Model(&models).Where("active = ?", true).
		Where("environment_id=?", envId).
		Select()
	return models, err
}
