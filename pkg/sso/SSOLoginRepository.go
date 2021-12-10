package sso

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

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

/*
	@description: user crud
*/

type SSOLoginRepository interface {
	Create(userModel *SSOLoginModel, tx *pg.Tx) (*SSOLoginModel, error)
	Update(userModel *SSOLoginModel, tx *pg.Tx) (*SSOLoginModel, error)
	GetById(id int32) (*SSOLoginModel, error)
	GetAll() ([]SSOLoginModel, error)
	GetActive() (*SSOLoginModel, error)
	Delete(userModel *SSOLoginModel, tx *pg.Tx) (bool, error)
	GetByName(name string) (*SSOLoginModel, error)

	GetConnection() (dbConnection *pg.DB)
}

type SSOLoginRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewSSOLoginRepositoryImpl(dbConnection *pg.DB) *SSOLoginRepositoryImpl {
	return &SSOLoginRepositoryImpl{dbConnection: dbConnection}
}

type SSOLoginModel struct {
	TableName struct{} `sql:"sso_login_config"`
	Id        int32    `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
	Label     string   `sql:"label"`
	Url       string   `sql:"url"`
	Config    string   `sql:"config"`
	Active    bool     `sql:"active,notnull"`
	sql.AuditLog
}

func (impl SSOLoginRepositoryImpl) Create(userModel *SSOLoginModel, tx *pg.Tx) (*SSOLoginModel, error) {
	err := tx.Insert(userModel)
	if err != nil {
		return userModel, err
	}
	return userModel, nil
}
func (impl SSOLoginRepositoryImpl) Update(userModel *SSOLoginModel, tx *pg.Tx) (*SSOLoginModel, error) {
	err := tx.Update(userModel)
	if err != nil {
		return userModel, err
	}

	return userModel, nil
}
func (impl SSOLoginRepositoryImpl) GetById(id int32) (*SSOLoginModel, error) {
	var model SSOLoginModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Order("id desc").Limit(1).Select()
	return &model, err
}
func (impl SSOLoginRepositoryImpl) GetAll() ([]SSOLoginModel, error) {
	var userModel []SSOLoginModel
	err := impl.dbConnection.Model(&userModel).Order("updated_on desc").Select()
	return userModel, err
}

func (impl SSOLoginRepositoryImpl) GetActive() (*SSOLoginModel, error) {
	var model SSOLoginModel
	err := impl.dbConnection.Model(&model).Where("active = ?", true).Limit(1).Select()
	return &model, err
}

func (impl SSOLoginRepositoryImpl) Delete(userModel *SSOLoginModel, tx *pg.Tx) (bool, error) {
	err := tx.Delete(userModel)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl SSOLoginRepositoryImpl) GetByName(name string) (*SSOLoginModel, error) {
	var model SSOLoginModel
	err := impl.dbConnection.Model(&model).Where("name = ?", name).Order("id desc").Limit(1).Select()
	return &model, err
}

func (impl *SSOLoginRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}
