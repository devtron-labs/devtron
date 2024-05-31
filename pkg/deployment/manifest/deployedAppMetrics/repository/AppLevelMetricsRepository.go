/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AppLevelMetrics struct {
	tableName  struct{} `sql:"app_level_metrics" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	AppId      int      `sql:"app_id,notnull"`
	AppMetrics bool     `sql:"app_metrics,notnull"`
	//InfraMetrics bool     `sql:"infra_metrics,notnull"` not being used
	sql.AuditLog
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
