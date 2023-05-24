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

package appStoreValuesRepository

import (
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AppStoreVersionValuesRepository interface {
	CreateAppStoreVersionValues(model *AppStoreVersionValues) (*AppStoreVersionValues, error)
	UpdateAppStoreVersionValues(model *AppStoreVersionValues) (*AppStoreVersionValues, error)
	DeleteAppStoreVersionValues(model *AppStoreVersionValues) (bool, error)
	/*	FindAll() ([]*AppStoreVersionValues, error)*/
	FindById(id int) (*AppStoreVersionValues, error)
	FindValuesByAppStoreId(appStoreVersionId int) ([]*AppStoreVersionValues, error)
	FindValuesByAppStoreIdAndReferenceType(appStoreVersionId int, referenceType string) ([]*AppStoreVersionValues, error)
}

type AppStoreVersionValuesRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewAppStoreVersionValuesRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *AppStoreVersionValuesRepositoryImpl {
	return &AppStoreVersionValuesRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type AppStoreVersionValues struct {
	TableName                    struct{} `sql:"app_store_version_values" pg:",discard_unknown_columns"`
	Id                           int      `sql:"id,pk"`
	Name                         string   `sql:"name"`
	ValuesYaml                   string   `sql:"values_yaml"`
	AppStoreApplicationVersionId int      `sql:"app_store_application_version_id"`
	ReferenceType                string   `sql:"reference_type"`
	Description                  string   `sql:"description"`
	Deleted                      bool     `sql:"deleted,notnull"`
	sql.AuditLog
	AppStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion
}

func (impl AppStoreVersionValuesRepositoryImpl) CreateAppStoreVersionValues(model *AppStoreVersionValues) (*AppStoreVersionValues, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return nil, err
	}
	return model, nil
}

func (impl AppStoreVersionValuesRepositoryImpl) UpdateAppStoreVersionValues(model *AppStoreVersionValues) (*AppStoreVersionValues, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl AppStoreVersionValuesRepositoryImpl) DeleteAppStoreVersionValues(model *AppStoreVersionValues) (bool, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return false, err
	}
	return true, nil
}

/*func (impl *AppStoreVersionValuesRepositoryImpl) FindAll() ([]*AppStoreVersionValues, error) {
	var appStoreWithVersion []*AppStoreVersionValues
	err := impl.dbConnection.
		Model(appStoreWithVersion).
		Column("app_store_version_values.*", "AppStoreApplicationVersion").
		Select()
	return appStoreWithVersion, err
}*/

func (impl AppStoreVersionValuesRepositoryImpl) FindById(id int) (*AppStoreVersionValues, error) {
	appStoreWithVersion := &AppStoreVersionValues{}
	err := impl.dbConnection.
		Model(appStoreWithVersion).
		Column("app_store_version_values.*", "AppStoreApplicationVersion").
		Where("app_store_version_values.id = ?", id).
		Where("app_store_version_values.deleted =?", false).
		Limit(1).
		Select()
	return appStoreWithVersion, err
}

func (impl AppStoreVersionValuesRepositoryImpl) FindValuesByAppStoreId(appStoreId int) ([]*AppStoreVersionValues, error) {
	var appStoreVersionValues []*AppStoreVersionValues
	err := impl.dbConnection.
		Model(&appStoreVersionValues).
		Column("app_store_version_values.id", "app_store_version_values.name", "AppStoreApplicationVersion.version", "AppStoreApplicationVersion.id").
		Join("inner join app_store_application_version apv on apv.id = app_store_version_values.app_store_application_version_id").
		Where("apv.app_store_id = ?", appStoreId).
		Where("app_store_version_values.deleted =?", false).
		Select()
	return appStoreVersionValues, err
}

func (impl AppStoreVersionValuesRepositoryImpl) FindValuesByAppStoreIdAndReferenceType(appStoreId int, referenceType string) ([]*AppStoreVersionValues, error) {
	var appStoreVersionValues []*AppStoreVersionValues
	err := impl.dbConnection.
		Model(&appStoreVersionValues).
		Column("app_store_version_values.id", "app_store_version_values.name", "app_store_version_values.description", "app_store_version_values.updated_on", "app_store_version_values.updated_by", "AppStoreApplicationVersion.version").
		Join("inner join app_store_application_version apv on apv.id = app_store_version_values.app_store_application_version_id").
		Where("apv.app_store_id = ?", appStoreId).Where("app_store_version_values.reference_type = ?", referenceType).
		Where("app_store_version_values.deleted =?", false).
		Select()
	return appStoreVersionValues, err
}
