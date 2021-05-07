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
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ExternalAppsRepository interface {
	Create(model *ExternalApps) (*ExternalApps, error)
	Update(model *ExternalApps) (*ExternalApps, error)
	Delete(model *ExternalApps) (bool, error)
	FindById(id int) (*ExternalApps, error)
	FindAll() ([]*ExternalApps, error)
	FindByAppName(appName string) (*ExternalApps, error)
	SearchByFilter(appName string, clusterIds []int, namespaces []string, offset int, limit int) ([]*ExternalApps, error)
}

type ExternalAppsRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewExternalAppsRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *ExternalAppsRepositoryImpl {
	return &ExternalAppsRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type ExternalApps struct {
	TableName      struct{}  `sql:"external_apps" pg:",discard_unknown_columns"`
	Id             int       `sql:"id,pk"`
	AppName        string    `sql:"app_name"`
	Label          string    `sql:"label"`
	ChartName      string    `sql:"chart_name"`
	Namespace      string    `sql:"namespace"`
	ClusterId      int       `sql:"cluster_id"`
	LastDeployedOn time.Time `sql:"last_deployed_on"`
	Active         bool      `sql:"active,notnull"`
	Status         string    `sql:"status"`
	ChartVersion   string    `sql:"chart_version"`
	Deprecated     bool      `sql:"deprecated,notnull"`
	models.AuditLog
}

func (impl ExternalAppsRepositoryImpl) Create(model *ExternalApps) (*ExternalApps, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return nil, err
	}
	return model, nil
}

func (impl ExternalAppsRepositoryImpl) Update(model *ExternalApps) (*ExternalApps, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl ExternalAppsRepositoryImpl) Delete(model *ExternalApps) (bool, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return false, err
	}
	return true, nil
}

func (impl ExternalAppsRepositoryImpl) FindById(id int) (*ExternalApps, error) {
	externalApp := &ExternalApps{}
	err := impl.dbConnection.
		Model(externalApp).
		Where("external_apps.id = ?", id).
		Where("external_apps.active =?", true).
		Select()
	return externalApp, err
}

func (impl ExternalAppsRepositoryImpl) FindAll() ([]*ExternalApps, error) {
	var externalApps []*ExternalApps
	err := impl.dbConnection.
		Model(&externalApps).
		Where("external_apps.active =?", true).
		Select()
	return externalApps, err
}
func (impl ExternalAppsRepositoryImpl) FindByAppName(appName string) (*ExternalApps, error) {
	externalApp := &ExternalApps{}
	err := impl.dbConnection.
		Model(externalApp).
		Where("external_apps.app_name = ?", appName).
		Where("external_apps.active =?", true).
		Select()
	return externalApp, err
}

func (impl ExternalAppsRepositoryImpl) SearchByFilter(appName string, clusterIds []int, namespaces []string, offset int, limit int) ([]*ExternalApps, error) {
	var externalApps []*ExternalApps
	var err error
	if len(clusterIds) == 0 && len(namespaces) == 0 {
		err = impl.dbConnection.
			Model(&externalApps).
			Where("external_apps.app_name like (?)", "%"+appName+"%").
			Where("external_apps.active =?", true).Offset(offset).Limit(limit).
			Select()
	} else if len(clusterIds) > 0 && len(namespaces) == 0 {
		err = impl.dbConnection.
			Model(&externalApps).
			Where("external_apps.app_name like (?)", "%"+appName+"%").
			Where("external_apps.cluster_id in (?)", pg.In(clusterIds)).
			Where("external_apps.active =?", true).Offset(offset).Limit(limit).
			Select()
	} else if len(clusterIds) == 0 && len(namespaces) > 0 {
		err = impl.dbConnection.
			Model(&externalApps).
			Where("external_apps.app_name like (?)", "%"+appName+"%").
			Where("external_apps.namespace in (?)", pg.In(namespaces)).
			Where("external_apps.active =?", true).Offset(offset).Limit(limit).
			Select()
	} else if len(clusterIds) > 0 && len(namespaces) > 0 {
		err = impl.dbConnection.
			Model(&externalApps).
			Where("external_apps.app_name like (?)", "%"+appName+"%").
			Where("external_apps.cluster_id in (?)", pg.In(clusterIds)).
			Where("external_apps.namespace in (?)", pg.In(namespaces)).
			Where("external_apps.active =?", true).Offset(offset).Limit(limit).
			Select()
	}

	return externalApps, err
}
