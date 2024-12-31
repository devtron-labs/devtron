/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package chartConfig

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"time"
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
	GetEnvLevelByAppId(appId int) ([]*ConfigMapEnvModel, error)
	GetConfigNamesForAppAndEnvLevel(appId int, envId int) ([]bean.ConfigNameAndType, error)
}

type ConfigMapRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewConfigMapRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *ConfigMapRepositoryImpl {
	return &ConfigMapRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

const (
	ConfigMapAppLevel string = "config_map_app_level"
	ConfigMapEnvLevel string = "config_map_env_level"
)

type ConfigMapAppModel struct {
	TableName     struct{} `sql:"config_map_app_level" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	AppId         int      `sql:"app_id,notnull"`
	ConfigMapData string   `sql:"config_map_data"`
	SecretData    string   `sql:"secret_data"`
	sql.AuditLog
}
type cMCSNames struct {
	Id     int    `json:"id"`
	CMName string `json:"cm_name"`
	CSName string `json:"cs_name"`
}

func (impl ConfigMapRepositoryImpl) GetConfigNamesForAppAndEnvLevel(appId int, envId int) ([]bean.ConfigNameAndType, error) {
	var cMCSNames []cMCSNames
	tableName := ConfigMapEnvLevel
	if envId == -1 {
		tableName = ConfigMapAppLevel
	}
	//below query iterates over the cm, cs stored as json element, and fetches cmName and csName, id for a particular appId or envId if provided
	query := impl.dbConnection.
		Model().
		Table(tableName).
		Column("id").
		ColumnExpr("json_array_elements(CASE WHEN (config_map_data::json->'maps')::TEXT != 'null' THEN (config_map_data::json->'maps') ELSE '[]' END )->>'name' AS cm_name").
		ColumnExpr("json_array_elements(CASE WHEN (secret_data::json->'secrets')::TEXT != 'null' THEN (secret_data::json->'secrets') ELSE '[]' END )->>'name' AS cs_name").
		Where("app_id = ?", appId)

	if envId > 0 {
		query = query.Where("environment_id=?", envId)
	}
	if err := query.Select(&cMCSNames); err != nil {
		if err != pg.ErrNoRows {
			impl.Logger.Errorw("error occurred while fetching CM/CS names", "appId", appId, "err", err)
			return nil, err
		}
	}
	var configNames []bean.ConfigNameAndType
	for _, name := range cMCSNames {
		if name.CMName != "" {
			configNames = append(configNames, bean.ConfigNameAndType{
				Id:   name.Id,
				Name: name.CMName,
				Type: bean.CM,
			})
		}
		if name.CSName != "" {
			configNames = append(configNames, bean.ConfigNameAndType{
				Id:   name.Id,
				Name: name.CSName,
				Type: bean.CS,
			})
		}
	}
	return configNames, nil
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
	Deleted       bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (impl ConfigMapRepositoryImpl) CreateEnvLevel(model *ConfigMapEnvModel) (*ConfigMapEnvModel, error) {
	currentTime := time.Now()
	model.CreatedOn = currentTime
	model.UpdatedOn = currentTime
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
	model.UpdatedOn = time.Now()
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

func (impl ConfigMapRepositoryImpl) GetEnvLevelByAppId(appId int) ([]*ConfigMapEnvModel, error) {
	var models []*ConfigMapEnvModel
	err := impl.dbConnection.Model(&models).
		Where("app_id = ?", appId).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.WhereOr("deleted = ? ", false).WhereOr("deleted IS NULL")
			return query, nil
		}).Select()
	if err != nil {
		impl.Logger.Errorw("err in getting cm/cs env level", "err", err)
		return models, err
	}
	return models, err
}
