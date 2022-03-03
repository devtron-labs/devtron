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
	"github.com/go-pg/pg/orm"
)

type Environment struct {
	tableName             struct{} `sql:"environment" pg:",discard_unknown_columns"`
	Id                    int      `sql:"id,pk"`
	Name                  string   `sql:"environment_name"`
	ClusterId             int      `sql:"cluster_id"`
	Cluster               *Cluster
	Active                bool   `sql:"active,notnull"`
	Default               bool   `sql:"default,notnull"`
	GrafanaDatasourceId   int    `sql:"grafana_datasource_id"`
	Namespace             string `sql:"namespace"`
	EnvironmentIdentifier string `sql:"environment_identifier"`
	sql.AuditLog
}

type EnvironmentRepository interface {
	FindOne(environment string) (*Environment, error)
	Create(mappings *Environment) error
	FindAll() ([]Environment, error)
	FindAllActive() ([]Environment, error)
	MarkEnvironmentDeleted(mappings *Environment, tx *pg.Tx) error
	GetConnection() (dbConnection *pg.DB)

	FindById(id int) (*Environment, error)
	Update(mappings *Environment) error
	FindByName(name string) (*Environment, error)
	FindByIdentifier(identifier string) (*Environment, error)
	FindByNameOrIdentifier(name string, identifier string) (*Environment, error)
	FindByClusterId(clusterId int) ([]*Environment, error)
	FindByIds(ids []*int) ([]*Environment, error)
	FindByNamespaceAndClusterName(namespaces string, clusterName string) (*Environment, error)
	FindOneByNamespaceAndClusterId(namespace string, clusterId int) (*Environment, error)
	FindByClusterIdAndNamespace(namespaceClusterPair []*ClusterNamespacePair) ([]*Environment, error)
	FindByClusterIds(clusterIds []int) ([]*Environment, error)
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

func (repositoryImpl EnvironmentRepositoryImpl) FindOneByNamespaceAndClusterId(namespace string, clusterId int) (*Environment, error) {
	environmentCluster := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environmentCluster).
		Column("environment.*", "Cluster").
		Join("inner join cluster c on environment.cluster_id = c.id").
		Where("namespace = ?", namespace).
		Where("environment.active = ?", true).
		Where("c.active = ?", true).
		Where("c.id =?", clusterId).
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

func (repositoryImpl EnvironmentRepositoryImpl) FindByIdentifier(identifier string) (*Environment, error) {
	environment := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environment).
		Where("environment_identifier = ?", identifier).
		Where("active = ?", true).
		Limit(1).
		Select()
	return environment, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByNameOrIdentifier(name string, identifier string) (*Environment, error) {
	environment := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environment).
		Where("active = ?", true).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.Where("environment_identifier = ?", identifier).WhereOr("environment_name = ?", name)
			return query, nil
		}).
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

type ClusterNamespacePair struct {
	ClusterId     int
	NamespaceName string
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByClusterIdAndNamespace(namespaceClusterPair []*ClusterNamespacePair) ([]*Environment, error) {
	var mappings []*Environment
	var clusterNsPair []interface{}
	for _, _pair := range namespaceClusterPair {
		if len(_pair.NamespaceName) > 0 {
			clusterNsPair = append(clusterNsPair, []interface{}{_pair.ClusterId, _pair.NamespaceName})
		}
	}
	err := repositoryImpl.dbConnection.
		Model(&mappings).
		Where("active = true").
		Where(" (cluster_id, namespace) in (?)", pg.InMulti(clusterNsPair...)).
		Select()
	return mappings, err
}
func (repositoryImpl EnvironmentRepositoryImpl) FindByClusterIds(clusterIds []int) ([]*Environment, error) {
	var mappings []*Environment
	err := repositoryImpl.dbConnection.
		Model(&mappings).
		Column("environment.*", "Cluster").
		Where("environment.active = true").
		Where("environment.cluster_id in (?)", pg.In(clusterIds)).
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

func (repositoryImpl EnvironmentRepositoryImpl) FindByIds(ids []*int) ([]*Environment, error) {
	var apps []*Environment
	err := repositoryImpl.dbConnection.Model(&apps).Where("active = ?", true).Where("id in (?)", pg.In(ids)).Select()
	return apps, err
}

func (repo EnvironmentRepositoryImpl) MarkEnvironmentDeleted(deleteReq *Environment, tx *pg.Tx) error {
	deleteReq.Active = false
	return tx.Update(deleteReq)
}
func (repo EnvironmentRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return repo.dbConnection
}
