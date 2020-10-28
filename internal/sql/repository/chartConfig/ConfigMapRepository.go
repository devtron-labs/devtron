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

package chartConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ConfigMapRepository interface {
	CreateAppLevel(model *ConfigMapAppModel) (*ConfigMapAppModel, error)
	GetByIdAppLevel(id int) (*ConfigMapAppModel, error)
	GetAllAppLevel() ([]ConfigMapAppModel, error)
	UpdateAppLevel(model *ConfigMapAppModel) (*ConfigMapAppModel, error)

	CreateEnvLevel(model *ConfigMapEnvModel) (*ConfigMapEnvModel, error)
	GetByIdEnvLevel(id int) (*ConfigMapEnvModel, error)
	GetAllEnvLevel() ([]ConfigMapEnvModel, error)
	UpdateEnvLevel(model *ConfigMapEnvModel) (*ConfigMapEnvModel, error)

	GetByAppIdAppLevel(appId int) (*ConfigMapAppModel, error)
	GetByAppIdAndEnvIdEnvLevel(appId int, envId int) (*ConfigMapEnvModel, error)
}

type ConfigMapRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewConfigMapRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *ConfigMapRepositoryImpl {
	return &ConfigMapRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type ConfigMapAppModel struct {
	TableName     struct{} `sql:"config_map_app_level" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	AppId         int      `sql:"app_id,notnull"`
	ConfigMapData string   `sql:"config_map_data"`
	SecretData    string   `sql:"secret_data"`
	models.AuditLog
}

func (impl ConfigMapRepositoryImpl) CreateAppLevel(model *ConfigMapAppModel) (*ConfigMapAppModel, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Errorw("err on config map ", "err;", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapRepositoryImpl) GetByIdAppLevel(id int) (*ConfigMapAppModel, error) {
	var model ConfigMapAppModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	return &model, err
}
func (impl ConfigMapRepositoryImpl) GetAllAppLevel() ([]ConfigMapAppModel, error) {
	var models []ConfigMapAppModel
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl ConfigMapRepositoryImpl) UpdateAppLevel(model *ConfigMapAppModel) (*ConfigMapAppModel, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Errorw("err on config map ", "err;", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapRepositoryImpl) GetByAppIdAppLevel(appId int) (*ConfigMapAppModel, error) {
	var model ConfigMapAppModel
	err := impl.dbConnection.Model(&model).Where("app_id = ?", appId).Select()
	return &model, err
}

// ---------------------------------------------------------------------------------------------

type ConfigMapEnvModel struct {
	TableName     struct{} `sql:"config_map_env_level" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	AppId         int      `sql:"app_id,notnull"`
	EnvironmentId int      `sql:"environment_id,notnull"`
	ConfigMapData string   `sql:"config_map_data"`
	SecretData    string   `sql:"secret_data"`
	models.AuditLog
}

func (impl ConfigMapRepositoryImpl) CreateEnvLevel(model *ConfigMapEnvModel) (*ConfigMapEnvModel, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Errorw("err on config map ", "err;", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapRepositoryImpl) GetByIdEnvLevel(id int) (*ConfigMapEnvModel, error) {
	var model ConfigMapEnvModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	return &model, err
}
func (impl ConfigMapRepositoryImpl) GetAllEnvLevel() ([]ConfigMapEnvModel, error) {
	var models []ConfigMapEnvModel
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl ConfigMapRepositoryImpl) UpdateEnvLevel(model *ConfigMapEnvModel) (*ConfigMapEnvModel, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Errorw("err on config map ", "err;", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapRepositoryImpl) GetByAppIdAndEnvIdEnvLevel(appId int, envId int) (*ConfigMapEnvModel, error) {
	var model ConfigMapEnvModel
	err := impl.dbConnection.Model(&model).Where("app_id = ?", appId).Where("environment_id = ?", envId).Select()
	return &model, err
}
