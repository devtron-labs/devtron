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

package apiToken

import (
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type ApiToken struct {
	tableName    struct{} `sql:"api_token"`
	Id           int      `sql:"id,pk"`
	UserId       int32    `sql:"user_id, notnull"`
	Name         string   `sql:"name, notnull"`
	Version      int      `sql:"version, notnull"`
	Description  string   `sql:"description, notnull"`
	ExpireAtInMs int64    `sql:"expire_at_in_ms"`
	Token        string   `sql:"token, notnull"`
	User         *repository.UserModel
	sql.AuditLog
}

type ApiTokenRepository interface {
	GetConnection() *pg.DB
	Save(tx *pg.Tx, apiToken *ApiToken) error
	Update(tx *pg.Tx, apiToken *ApiToken) error
	FindAllActive() ([]*ApiToken, error)
	FindActiveById(id int) (*ApiToken, error)
	FindByName(name string) (*ApiToken, error)
	CheckIfTokenExistsByNameAndVersion(tx *pg.Tx, name string, version int) (bool, error)
}

type ApiTokenRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewApiTokenRepositoryImpl(dbConnection *pg.DB) *ApiTokenRepositoryImpl {
	return &ApiTokenRepositoryImpl{dbConnection: dbConnection}
}

func (impl ApiTokenRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl ApiTokenRepositoryImpl) Save(tx *pg.Tx, apiToken *ApiToken) error {
	return tx.Insert(apiToken)
}

func (impl ApiTokenRepositoryImpl) Update(tx *pg.Tx, apiToken *ApiToken) error {
	return tx.Update(apiToken)
}

func (impl ApiTokenRepositoryImpl) FindAllActive() ([]*ApiToken, error) {
	var apiTokens []*ApiToken
	err := impl.dbConnection.Model(&apiTokens).
		Column("api_token.*", "User").
		Relation("User", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("active IS TRUE"), nil
		}).
		Select()
	return apiTokens, err
}

func (impl ApiTokenRepositoryImpl) FindActiveById(id int) (*ApiToken, error) {
	apiToken := &ApiToken{}
	err := impl.dbConnection.Model(apiToken).
		Column("api_token.*", "User").
		Relation("User", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("active IS TRUE"), nil
		}).
		Where("api_token.id = ?", id).
		Select()
	return apiToken, err
}

func (impl ApiTokenRepositoryImpl) FindByName(name string) (*ApiToken, error) {
	apiToken := &ApiToken{}
	err := impl.dbConnection.Model(apiToken).
		Column("api_token.*", "User").
		Where("api_token.name = ?", name).
		Select()
	return apiToken, err
}

func (impl ApiTokenRepositoryImpl) CheckIfTokenExistsByNameAndVersion(tx *pg.Tx, name string, version int) (bool, error) {
	apiToken := &ApiToken{}
	exists, err := tx.Model(apiToken).
		Where("name = ?", name).
		Where("version = ?", version).
		Exists()
	return exists, err
}
