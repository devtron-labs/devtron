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
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DbType string

const (
	Db_TYPE_POSTGRESS DbType = "postgres"
	Db_TYPE_MYSQL     DbType = "mysql"
	DB_TYPE_MARIADB   DbType = "mariadb"
)

func (t DbType) IsValid() bool {
	types := map[string]DbType{"postgres": Db_TYPE_POSTGRESS, "mysql": Db_TYPE_MYSQL, "mariadb": DB_TYPE_MARIADB}
	_, ok := types[string(t)]
	return ok
}

type DbConfig struct {
	tableName struct{} `sql:"db_config" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name"` //name by which user identifies this db
	Type      DbType   `sql:"type"` //type of db, PG, MYsql, MariaDb
	Host      string   `sql:"host"`
	Port      string   `sql:"port"`
	DbName    string   `sql:"db_name"` //name of database inside PG
	UserName  string   `sql:"user_name"`
	Password  string   `sql:"password"`
	Active    bool     `sql:"active"`
	models.AuditLog
}

type DbConfigRepository interface {
	Save(config *DbConfig) error
	GetAll() (configs []*DbConfig, err error)
	GetById(id int) (*DbConfig, error)
	Update(config *DbConfig) (*DbConfig, error)
	GetActiveForAutocomplete() (configs []*DbConfig, err error)
}
type DbConfigRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewDbConfigRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DbConfigRepositoryImpl {
	return &DbConfigRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl DbConfigRepositoryImpl) Save(config *DbConfig) error {
	return impl.dbConnection.Insert(config)
}

func (impl DbConfigRepositoryImpl) GetAll() (configs []*DbConfig, err error) {
	err = impl.dbConnection.Model(&configs).Select()
	return configs, err
}

func (impl DbConfigRepositoryImpl) GetById(id int) (*DbConfig, error) {
	cfg := &DbConfig{Id: id}
	err := impl.dbConnection.Model(cfg).WherePK().Select()
	return cfg, err
}

func (impl DbConfigRepositoryImpl) Update(config *DbConfig) (*DbConfig, error) {
	_, err := impl.dbConnection.Model(config).WherePK().UpdateNotNull()
	return config, err
}

func (impl DbConfigRepositoryImpl) GetActiveForAutocomplete() (configs []*DbConfig, err error) {
	err = impl.dbConnection.Model(&configs).
		Where("active = ?", true).
		Column("id", "name").
		Select()

	return configs, err
}
