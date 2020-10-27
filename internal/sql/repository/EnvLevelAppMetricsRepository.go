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

type EnvLevelAppMetrics struct {
	tableName    struct{} `sql:"env_level_app_metrics" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	AppId        int      `sql:"app_id,notnull"`
	EnvId        int      `sql:"env_id,notnull"`
	AppMetrics   *bool    `sql:"app_metrics,notnull"`
	InfraMetrics *bool    `sql:"infra_metrics,notnull"`
	models.AuditLog
}

type EnvLevelAppMetricsRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

type EnvLevelAppMetricsRepository interface {
	Save(metrics *EnvLevelAppMetrics) error
	Update(metrics *EnvLevelAppMetrics) error
	FindByAppIdAndEnvId(appId int, envId int) (*EnvLevelAppMetrics, error)
	Delete(metrics *EnvLevelAppMetrics) error
	FindByAppId(appId int) ([]*EnvLevelAppMetrics, error)
}

func NewEnvLevelAppMetricsRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *EnvLevelAppMetricsRepositoryImpl {
	return &EnvLevelAppMetricsRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *EnvLevelAppMetricsRepositoryImpl) Save(metrics *EnvLevelAppMetrics) error {
	return impl.dbConnection.Insert(metrics)
}

func (impl *EnvLevelAppMetricsRepositoryImpl) FindByAppIdAndEnvId(appId int, envId int) (*EnvLevelAppMetrics, error) {
	envAppLevelMetrics := &EnvLevelAppMetrics{}
	err := impl.dbConnection.Model(envAppLevelMetrics).Where("env_level_app_metrics.app_id = ? ", appId).Where("env_level_app_metrics.env_id = ? ", envId).Select()
	return envAppLevelMetrics, err
}
func (impl *EnvLevelAppMetricsRepositoryImpl) FindByAppId(appId int) ([]*EnvLevelAppMetrics, error) {
	var envAppLevelMetrics []*EnvLevelAppMetrics
	err := impl.dbConnection.Model(&envAppLevelMetrics).
		Where("env_level_app_metrics.app_id = ? ", appId).
		Select()
	return envAppLevelMetrics, err
}

func (impl *EnvLevelAppMetricsRepositoryImpl) Update(metrics *EnvLevelAppMetrics) error {
	err := impl.dbConnection.Update(metrics)
	return err
}

func (impl *EnvLevelAppMetricsRepositoryImpl) Delete(metrics *EnvLevelAppMetrics) error {
	return impl.dbConnection.Delete(metrics)
}
