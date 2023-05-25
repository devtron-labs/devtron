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
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type AppGroupMapping struct {
	tableName  struct{} `sql:"app_group_mapping" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	AppGroupId int      `sql:"app_group_id,notnull"`
	AppId      int      `sql:"app_id,notnull"`
	AppGroup   *AppGroup
	App        *app.App
	sql.AuditLog
}

type AppGroupMappingRepository interface {
	Save(model *AppGroupMapping, tx *pg.Tx) (*AppGroupMapping, error)
	Update(model *AppGroupMapping, tx *pg.Tx) error
	Delete(model *AppGroupMapping, tx *pg.Tx) error
	FindById(id int) (*AppGroupMapping, error)
	FindByAppGroupId(appGroupId int) ([]*AppGroupMapping, error)
	FindAll() ([]*AppGroupMapping, error)
	FindByAppGroupIds(appGroupIds []int) ([]*AppGroupMapping, error)
	GetConnection() (dbConnection *pg.DB)
}

type AppGroupMappingRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewAppGroupMappingRepositoryImpl(dbConnection *pg.DB) *AppGroupMappingRepositoryImpl {
	return &AppGroupMappingRepositoryImpl{dbConnection: dbConnection}
}

func (repo *AppGroupMappingRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return repo.dbConnection
}

func (repo AppGroupMappingRepositoryImpl) Save(model *AppGroupMapping, tx *pg.Tx) (*AppGroupMapping, error) {
	err := tx.Insert(model)
	return model, err
}

func (repo AppGroupMappingRepositoryImpl) Update(model *AppGroupMapping, tx *pg.Tx) error {
	err := tx.Update(model)
	return err
}

func (repo AppGroupMappingRepositoryImpl) Delete(model *AppGroupMapping, tx *pg.Tx) error {
	err := tx.Delete(model)
	return err
}

func (repo AppGroupMappingRepositoryImpl) FindById(id int) (*AppGroupMapping, error) {
	model := &AppGroupMapping{}
	err := repo.dbConnection.Model(model).Where("id = ?", id).
		Select()
	return model, err
}

func (repo AppGroupMappingRepositoryImpl) FindByAppGroupId(appGroupId int) ([]*AppGroupMapping, error) {
	var models []*AppGroupMapping
	err := repo.dbConnection.Model(&models).
		Column("app_group_mapping.*", "AppGroup", "App").
		Where("app_group_mapping.app_group_id = ?", appGroupId).
		Select()
	return models, err
}

func (repo AppGroupMappingRepositoryImpl) FindAll() ([]*AppGroupMapping, error) {
	var models []*AppGroupMapping
	err := repo.dbConnection.Model(&models).Select()
	return models, err
}

func (repo AppGroupMappingRepositoryImpl) FindByAppGroupIds(appGroupIds []int) ([]*AppGroupMapping, error) {
	var models []*AppGroupMapping
	err := repo.dbConnection.Model(&models).
		Where("app_group_mapping.app_group_id in (?)", pg.In(appGroupIds)).
		Select()
	return models, err
}
