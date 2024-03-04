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

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type Attributes struct {
	tableName struct{} `sql:"attributes" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Key       string   `sql:"key,notnull"`
	Value     string   `sql:"value,notnull"`
	Active    bool     `sql:"active, notnull"`
	sql.AuditLog
}

type AttributesRepository interface {
	Save(model *Attributes, tx *pg.Tx) (*Attributes, error)
	Update(model *Attributes, tx *pg.Tx) error
	FindByKey(key string) (*Attributes, error)
	FindById(id int) (*Attributes, error)
	FindActiveList() ([]*Attributes, error)
	GetConnection() (dbConnection *pg.DB)
}

type AttributesRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewAttributesRepositoryImpl(dbConnection *pg.DB) *AttributesRepositoryImpl {
	return &AttributesRepositoryImpl{dbConnection: dbConnection}
}

func (impl *AttributesRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (repo AttributesRepositoryImpl) Save(model *Attributes, tx *pg.Tx) (*Attributes, error) {
	err := tx.Insert(model)
	return model, err
}

func (repo AttributesRepositoryImpl) Update(model *Attributes, tx *pg.Tx) error {
	err := tx.Update(model)
	return err
}

func (repo AttributesRepositoryImpl) FindByKey(key string) (*Attributes, error) {
	model := &Attributes{}
	err := repo.dbConnection.Model(model).Where("key = ?", key).Where("active = ?", true).
		Select()
	return model, err
}

func (repo AttributesRepositoryImpl) FindById(id int) (*Attributes, error) {
	model := &Attributes{}
	err := repo.dbConnection.Model(model).Where("id = ?", id).Where("active = ?", true).
		Select()
	return model, err
}

func (repo AttributesRepositoryImpl) FindActiveList() ([]*Attributes, error) {
	var models []*Attributes
	err := repo.dbConnection.Model(&models).Where("active = ?", true).
		Select()
	return models, err
}
