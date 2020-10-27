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

type AppLevelMetrics struct {
	tableName    struct{} `sql:"app_level_metrics" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	AppId        int      `sql:"app_id,notnull"`
	AppMetrics   bool     `sql:"app_metrics,notnull"`
	InfraMetrics bool     `sql:"infra_metrics,notnull"`
	models.AuditLog
}

type AppLevelMetricsRepository interface {
	Save(metrics *AppLevelMetrics) error
	FindByAppId(id int) (*AppLevelMetrics, error)
	Update(metrics *AppLevelMetrics) error
}

type AppLevelMetricsRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewAppLevelMetricsRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *AppLevelMetricsRepositoryImpl {
	return &AppLevelMetricsRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *AppLevelMetricsRepositoryImpl) Save(metrics *AppLevelMetrics) error {
	return impl.dbConnection.Insert(metrics)
}

func (impl *AppLevelMetricsRepositoryImpl) FindByAppId(appId int) (*AppLevelMetrics, error) {
	appLevelMetrics := &AppLevelMetrics{}
	err := impl.dbConnection.Model(appLevelMetrics).Where("app_level_metrics.app_id = ? ", appId).Select()
	return appLevelMetrics, err
}

func (impl *AppLevelMetricsRepositoryImpl) Update(metrics *AppLevelMetrics) error {
	err := impl.dbConnection.Update(metrics)
	return err
}
