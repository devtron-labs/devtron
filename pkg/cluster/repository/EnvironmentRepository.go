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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type Environment struct {
	tableName           struct{} `sql:"environment" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	Name                string   `sql:"environment_name"`
	ClusterId           int      `sql:"cluster_id"`
	Cluster             *Cluster
	Active              bool   `sql:"active,notnull"`
	Default             bool   `sql:"default,notnull"`
	GrafanaDatasourceId int    `sql:"grafana_datasource_id"`
	Namespace           string `sql:"namespace"`
	sql.AuditLog
}

type EnvironmentRepository interface {
	FindOne(environment string) (*Environment, error)
	Create(mappings *Environment) error
	FindAll() ([]Environment, error)
	FindAllActive() ([]Environment, error)

	FindById(id int) (*Environment, error)
	Update(mappings *Environment) error
	FindByName(name string) (*Environment, error)
	FindByClusterId(clusterId int) ([]*Environment, error)
	FindByIds(ids []*int) ([]*Environment, error)
	FindByNamespaceAndClusterName(namespaces string, clusterName string) (*Environment, error)
}

func NewEnvironmentRepositoryImpl(dbConnection *pg.DB) *EnvironmentRepositoryImpl {
	return &EnvironmentRepositoryImpl{dbConnection: dbConnection}
}

type EnvironmentRepositoryImpl struct {
	dbConnection *pg.DB
}

func (repositoryImpl EnvironmentRepositoryImpl) FindOne(environment string) (*Environment, error) {
	environmentCluster := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environmentCluster).
		Column("environment.*", "Cluster").
		Join("inner join cluster c on environment.cluster_id = c.id").
		Where("environment_name = ?", environment).
		Where("environment.active = ?", true).
		Where("c.active = ?", true).
		Limit(1).
		Select()
	return environmentCluster, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByNamespaceAndClusterName(namespaces string, clusterName string) (*Environment, error) {
	environmentCluster := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environmentCluster).
		Column("environment.*", "Cluster").
		Join("inner join cluster c on environment.cluster_id = c.id").
		Where("namespace = ?", namespaces).
		Where("environment.active = ?", true).
		Where("c.active = ?", true).
		Where("c.cluster_name =?", clusterName).
		Select()
	return environmentCluster, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByName(name string) (*Environment, error) {
	environment := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environment).
		Where("environment_name = ?", name).
		Where("active = ?", true).
		Limit(1).
		Select()
	return environment, err
}

func (repositoryImpl EnvironmentRepositoryImpl) Create(mappings *Environment) error {
	return repositoryImpl.dbConnection.Insert(mappings)
}

func (repositoryImpl EnvironmentRepositoryImpl) FindAll() ([]Environment, error) {
	var mappings []Environment
	err := repositoryImpl.
		dbConnection.Model(&mappings).
		Column("environment.*", "Cluster").
		Join("inner join cluster c on environment.cluster_id = c.id").
		Where("environment.active = ?", true).
		Select()
	return mappings, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindAllActive() ([]Environment, error) {
	var mappings []Environment
	err := repositoryImpl.
		dbConnection.Model(&mappings).
		Where("environment.active = ?", true).
		Column("environment.*", "Cluster").
		Join("inner join cluster c on environment.cluster_id = c.id").
		Select()
	return mappings, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindById(id int) (*Environment, error) {
	environmentCluster := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environmentCluster).
		Column("environment.*", "Cluster").
		Join("inner join cluster c on environment.cluster_id = c.id").
		Where("environment.id = ?", id).
		Where("environment.active = ?", true).
		Limit(1).
		Select()
	return environmentCluster, err
}

func (repositoryImpl EnvironmentRepositoryImpl) Update(mappings *Environment) error {
	return repositoryImpl.dbConnection.Update(mappings)
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByClusterId(clusterId int) ([]*Environment, error) {
	var mappings []*Environment
	err := repositoryImpl.
		dbConnection.Model(&mappings).
		Where("environment.active = ?", true).
		Column("environment.*").
		Where("environment.cluster_id = ?", clusterId).
		Select()
	return mappings, err
}

func (repo EnvironmentRepositoryImpl) FindByIds(ids []*int) ([]*Environment, error) {
	var apps []*Environment
	err := repo.dbConnection.Model(&apps).Where("active = ?", true).Where("id in (?)", pg.In(ids)).Select()
	return apps, err
}
