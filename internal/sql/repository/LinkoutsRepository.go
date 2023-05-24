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
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type LinkoutsRepository interface {
	Create(model *LinkoutsModel) (*LinkoutsModel, error)
	GetById(id int) (*LinkoutsModel, error)
	GetAll() ([]LinkoutsModel, error)
	Update(model *LinkoutsModel) (*LinkoutsModel, error)
	FetchLinkoutsByAppIdAndEnvId(appId int, envId int) ([]LinkoutsModel, error)
	FetchLinkoutById(Id int) (bean.LinkOuts, error)
}

type LinkoutsRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewLinkoutsRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *LinkoutsRepositoryImpl {
	return &LinkoutsRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type LinkoutsModel struct {
	TableName     struct{} `sql:"app_env_linkouts" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	AppId         int      `sql:"app_id,notnull"`
	EnvironmentId int      `sql:"environment_id,notnull"`
	Link          string   `sql:"link"`
	Description   string   `sql:"description"`
	Name          string   `sql:"name"`
	sql.AuditLog
}

func (impl LinkoutsRepositoryImpl) Create(model *LinkoutsModel) (*LinkoutsModel, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl LinkoutsRepositoryImpl) GetById(id int) (*LinkoutsModel, error) {
	var model LinkoutsModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	return &model, err
}
func (impl LinkoutsRepositoryImpl) GetAll() ([]LinkoutsModel, error) {
	var models []LinkoutsModel
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl LinkoutsRepositoryImpl) Update(model *LinkoutsModel) (*LinkoutsModel, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl LinkoutsRepositoryImpl) FetchLinkoutsByAppIdAndEnvId(appId int, envId int) ([]LinkoutsModel, error) {
	var models []LinkoutsModel
	err := impl.dbConnection.Model(&models).Where("app_id = ?", appId).
		Where("environment_id = ?", envId).
		Select()
	return models, err
}

func (impl LinkoutsRepositoryImpl) FetchLinkoutById(Id int) (bean.LinkOuts, error) {
	var linkout bean.LinkOuts
	query := "" +
		" SELECT l.id, l.name , l.link, a.app_name, e.environment_name as env_name" +
		" from app_env_linkouts l" +
		" INNER JOIN app a on a.id=l.app_id" +
		" INNER JOIN environment e on e.id=l.environment_id" +
		" WHERE l.id = ?"

	_, err := impl.dbConnection.Query(&linkout, query, Id)
	if err != nil {
		impl.Logger.Errorw("Exception caught:", "err", err)
		return linkout, err
	}
	return linkout, err
}
