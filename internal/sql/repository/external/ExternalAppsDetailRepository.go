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

package external

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ExternalAppsDetailRepository interface {
	Create(model *ExternalAppsDetail) (*ExternalAppsDetail, error)
	Update(model *ExternalAppsDetail) (*ExternalAppsDetail, error)
	Delete(model *ExternalAppsDetail) (bool, error)
	FindById(id int) (*ExternalAppsDetail, error)
	FindAll() ([]*ExternalAppsDetail, error)
	FindByAppName(appName string) (*ExternalAppsDetail, error)
	FindByExternalAppsId(id int) (*ExternalAppsDetail, error)
}

type ExternalAppsDetailRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewExternalAppsDetailRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *ExternalAppsDetailRepositoryImpl {
	return &ExternalAppsDetailRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type ExternalAppsDetail struct {
	TableName        struct{} `sql:"external_apps_detail" pg:",discard_unknown_columns"`
	Id               int      `sql:"id,pk"`
	ExternalAppsId   int      `sql:"external_apps_id"`
	ResourceTree     string   `sql:"resource_tree"`
	ResourceManifest string   `sql:"resource_manifest"`
	//models.AuditLog
}

func (impl ExternalAppsDetailRepositoryImpl) Create(model *ExternalAppsDetail) (*ExternalAppsDetail, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return nil, err
	}
	return model, nil
}

func (impl ExternalAppsDetailRepositoryImpl) Update(model *ExternalAppsDetail) (*ExternalAppsDetail, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl ExternalAppsDetailRepositoryImpl) Delete(model *ExternalAppsDetail) (bool, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return false, err
	}
	return true, nil
}

func (impl ExternalAppsDetailRepositoryImpl) FindById(id int) (*ExternalAppsDetail, error) {
	externalApp := &ExternalAppsDetail{}
	err := impl.dbConnection.
		Model(externalApp).
		Where("external_apps_id = ?", id).
		//Where("external_apps.active =?", true).
		Select()
	return externalApp, err
}

func (impl ExternalAppsDetailRepositoryImpl) FindAll() ([]*ExternalAppsDetail, error) {
	var externalApps []*ExternalAppsDetail
	err := impl.dbConnection.
		Model(&externalApps).
		Where("external_apps.active =?", true).
		Select()
	return externalApps, err
}
func (impl ExternalAppsDetailRepositoryImpl) FindByAppName(appName string) (*ExternalAppsDetail, error) {
	externalApp := &ExternalAppsDetail{}
	err := impl.dbConnection.
		Model(externalApp).
		Where("external_apps.app_name = ?", appName).
		Select()
	return externalApp, err
}

func (impl ExternalAppsDetailRepositoryImpl) FindByExternalAppsId(id int) (*ExternalAppsDetail, error) {
	externalApp := &ExternalAppsDetail{}
	err := impl.dbConnection.
		Model(externalApp).
		Where("external_apps_id = ?", id).
		Select()
	return externalApp, err
}
