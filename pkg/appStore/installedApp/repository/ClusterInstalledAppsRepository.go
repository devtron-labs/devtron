/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ClusterInstalledApps struct {
	tableName      struct{} `sql:"cluster_installed_apps" pg:",discard_unknown_columns"`
	Id             int      `sql:"id,pk"`
	ClusterId      int      `sql:"cluster_id,notnull"`
	InstalledAppId int      `sql:"installed_app_id,notnull"`
	InstalledApp   InstalledApps
	sql.AuditLog
}

type ClusterInstalledAppsRepository interface {
	Save(model *ClusterInstalledApps, tx *pg.Tx) error
	FindByClusterId(clusterId int) ([]*ClusterInstalledApps, error)
	FindByClusterIds(clusterIds []int) ([]*ClusterInstalledApps, error)
	FindAll() ([]ClusterInstalledApps, error)
	Update(model *ClusterInstalledApps) error
	Delete(model *ClusterInstalledApps) error
}

func NewClusterInstalledAppsRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ClusterInstalledAppsRepositoryImpl {
	return &ClusterInstalledAppsRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type ClusterInstalledAppsRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (impl ClusterInstalledAppsRepositoryImpl) Save(model *ClusterInstalledApps, tx *pg.Tx) error {
	return tx.Insert(model)
}

func (impl ClusterInstalledAppsRepositoryImpl) FindByClusterId(clusterId int) ([]*ClusterInstalledApps, error) {
	var clusters []*ClusterInstalledApps
	err := impl.dbConnection.
		Model(clusters).
		Where("cluster_id = ?", clusterId).
		Select()
	return clusters, err
}

func (impl ClusterInstalledAppsRepositoryImpl) FindByClusterIds(clusterIds []int) ([]*ClusterInstalledApps, error) {
	var clusters []*ClusterInstalledApps
	err := impl.dbConnection.
		Model(clusters).
		Where("cluster_id in (?)", pg.In(clusterIds)).
		Select()
	return clusters, err
}

func (impl ClusterInstalledAppsRepositoryImpl) FindAll() ([]ClusterInstalledApps, error) {
	var clusters []ClusterInstalledApps
	err := impl.dbConnection.
		Model(&clusters).
		Select()
	return clusters, err
}

func (impl ClusterInstalledAppsRepositoryImpl) Update(model *ClusterInstalledApps) error {
	return impl.dbConnection.Update(model)
}

func (impl ClusterInstalledAppsRepositoryImpl) Delete(model *ClusterInstalledApps) error {
	return impl.dbConnection.Delete(model)
}
