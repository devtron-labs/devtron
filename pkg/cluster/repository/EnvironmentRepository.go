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
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type EnvCluserInfo struct {
	Id          int    `sql:"id"`
	ClusterName string `sql:"cluster_name"`
	Namespace   string `sql:"namespace"`
	Name        string `sql:"name"`
}
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
	Description           string `sql:"description"`
	IsVirtualEnvironment  bool   `sql:"is_virtual_environment"`
	sql.AuditLog
}

type EnvironmentRepository interface {
	FindOne(environment string) (*Environment, error)
	Create(mappings *Environment) error
	FindAll() ([]Environment, error)
	FindAllActive() ([]*Environment, error)
	MarkEnvironmentDeleted(mappings *Environment, tx *pg.Tx) error
	GetConnection() (dbConnection *pg.DB)
	FindAllActiveEnvOnlyDetails() ([]*Environment, error)

	FindById(id int) (*Environment, error)
	Update(mappings *Environment) error
	FindByName(name string) (*Environment, error)
	FindByIdentifier(identifier string) (*Environment, error)
	FindByNameOrIdentifier(name string, identifier string) (*Environment, error)
	FindByEnvNameOrIdentifierOrNamespace(clusterId int, envName string, identifier string, namespace string) (*Environment, error)
	FindByClusterId(clusterId int) ([]*Environment, error)
	FindByIds(ids []*int) ([]*Environment, error)
	FindByNamespaceAndClusterName(namespaces string, clusterName string) (*Environment, error)
	FindOneByNamespaceAndClusterId(namespace string, clusterId int) (*Environment, error)
	FindByClusterIdAndNamespace(namespaceClusterPair []*ClusterNamespacePair) ([]*Environment, error)
	FindByClusterIds(clusterIds []int) ([]*Environment, error)
	FindIdsByNames(envNames []string) ([]int, error)
	FindByNames(envNames []string) ([]*Environment, error)

	FindByEnvName(envName string) ([]*Environment, error)
	FindByEnvNameAndClusterIds(envName string, clusterIds []int) ([]*Environment, error)
	FindByClusterIdsWithFilter(clusterIds []int) ([]*Environment, error)
	FindAllActiveWithFilter() ([]*Environment, error)
	FindEnvClusterInfosByIds([]int) ([]*EnvCluserInfo, error)
	FindEnvLinkedWithCiPipelines(externalCi bool, ciPipelineIds []int) ([]*Environment, error)
}

func NewEnvironmentRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger, appStatusRepository appStatus.AppStatusRepository) *EnvironmentRepositoryImpl {
	return &EnvironmentRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		appStatusRepository: appStatusRepository,
	}
}

type EnvironmentRepositoryImpl struct {
	dbConnection        *pg.DB
	appStatusRepository appStatus.AppStatusRepository
	logger              *zap.SugaredLogger
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

func (repositoryImpl EnvironmentRepositoryImpl) FindEnvClusterInfosByIds(envIds []int) ([]*EnvCluserInfo, error) {
	query := "SELECT env.id as id,cluster.cluster_name,env.environment_name as name,env.namespace " +
		" FROM environment env INNER JOIN  cluster ON env.cluster_id = cluster.id "
	if len(envIds) > 0 {
		query += fmt.Sprintf(" WHERE env.id IN (%s)", helper.GetCommaSepratedString(envIds))
	}
	res := make([]*EnvCluserInfo, 0)
	_, err := repositoryImpl.dbConnection.Query(&res, query)
	return res, err
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

func (repositoryImpl EnvironmentRepositoryImpl) FindByEnvNameOrIdentifierOrNamespace(clusterId int, envName string, identifier string, namespace string) (*Environment, error) {
	environment := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environment).
		Where("active = ?", true).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.Where("environment_identifier = ?", identifier).
				WhereOr("environment_name = ?", envName).
				WhereOr("cluster_id = ? AND namespace = ?", clusterId, namespace)
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

func (repositoryImpl EnvironmentRepositoryImpl) FindAllActive() ([]*Environment, error) {
	var mappings []*Environment
	err := repositoryImpl.
		dbConnection.Model(&mappings).
		Where("environment.active = ?", true).
		Column("environment.*", "Cluster").
		Join("inner join cluster c on environment.cluster_id = c.id").
		Select()
	return mappings, err
}
func (repositoryImpl EnvironmentRepositoryImpl) FindAllActiveEnvOnlyDetails() ([]*Environment, error) {
	var mappings []*Environment
	err := repositoryImpl.
		dbConnection.Model(&mappings).
		Where("environment.active = ?", true).
		Select()
	return mappings, err
}
func (repositoryImpl EnvironmentRepositoryImpl) FindById(id int) (*Environment, error) {
	environmentCluster := &Environment{}
	err := repositoryImpl.dbConnection.
		Model(environmentCluster).
		Column("environment.*", "Cluster").
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
	//TODO : delete entries in app_status repo
	err := repo.appStatusRepository.DeleteWithEnvId(tx, deleteReq.Id)
	if err != nil {
		repo.logger.Errorw("error in deleting from app_status table with appId", "appId", deleteReq.Id, "err", err)
		return err
	}
	deleteReq.Active = false
	return tx.Update(deleteReq)
}
func (repo EnvironmentRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return repo.dbConnection
}

func (repo EnvironmentRepositoryImpl) FindIdsByNames(envNames []string) ([]int, error) {
	var ids []int
	query := "select id from environment where environment_name in (?) and active=?;"
	_, err := repo.dbConnection.Query(&ids, query, pg.In(envNames), true)
	return ids, err
}

func (repo EnvironmentRepositoryImpl) FindByNames(envNames []string) ([]*Environment, error) {
	var environment []*Environment
	err := repo.dbConnection.
		Model(&environment).
		Where("active = ?", true).
		Where("environment_name in (?)", pg.In(envNames)).
		Select()
	return environment, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByEnvName(envName string) ([]*Environment, error) {
	var environmentCluster []*Environment
	err := repositoryImpl.dbConnection.
		Model(&environmentCluster).
		Column("environment.*", "Cluster").
		Where("environment_name like ?", "%"+envName+"%").
		Where("environment.active = ?", true).
		Order("environment.environment_name ASC").Select()
	return environmentCluster, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByEnvNameAndClusterIds(envName string, clusterIds []int) ([]*Environment, error) {
	var mappings []*Environment
	err := repositoryImpl.dbConnection.
		Model(&mappings).
		Column("environment.*", "Cluster").
		Where("environment_name like ?", "%"+envName+"%").
		Where("environment.active = true").
		Where("environment.cluster_id in (?)", pg.In(clusterIds)).
		Order("environment.environment_name ASC").Select()
	return mappings, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindByClusterIdsWithFilter(clusterIds []int) ([]*Environment, error) {
	var mappings []*Environment
	err := repositoryImpl.dbConnection.Model(&mappings).
		Column("environment.*", "Cluster").
		Where("environment.active = true").
		Where("environment.cluster_id in (?)", pg.In(clusterIds)).
		Order("environment.environment_name ASC").Select()
	return mappings, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindAllActiveWithFilter() ([]*Environment, error) {
	var mappings []*Environment
	err := repositoryImpl.dbConnection.Model(&mappings).
		Column("environment.*", "Cluster").
		Where("environment.active = ?", true).
		Order("environment.environment_name ASC").Select()
	return mappings, err
}

func (repositoryImpl EnvironmentRepositoryImpl) FindEnvLinkedWithCiPipelines(externalCi bool, ciPipelineIds []int) ([]*Environment, error) {
	var mappings []*Environment
	componentType := "CI_PIPELINE"
	if externalCi {
		componentType = "WEBHOOK"
	}
	query := "SELECT env.* " +
		" FROM environment env " +
		" INNER JOIN pipeline ON pipeline.environment_id=env.id and env.active = true " +
		" INNER JOIN app_workflow_mapping apf ON component_id=pipeline.id AND type='CD_PIPELINE' AND apf.active=true " +
		" WHERE apf.app_workflow_id IN (SELECT apf2.app_workflow_id FROM app_workflow_mapping apf2 WHERE component_id IN (?) AND type='%s');"
	query = fmt.Sprintf(query, componentType)
	_, err := repositoryImpl.dbConnection.Query(&mappings, query, pg.In(ciPipelineIds))
	return mappings, err
}

//query := "SELECT env.* " +
//" FROM environment env " +
//" INNER JOIN pipeline ON pipeline.environment_id=env.id and env.active = true " +
//" INNER JOIN app_workflow_mapping apf ON component_id=pipeline.id AND type='CD_PIPELINE' AND apf.active=true " +
//" INNER JOIN " +
//" (SELECT apf2.app_workflow_id FROM app_workflow_mapping apf2 WHERE component_id IN (?) AND type='CI_PIPELINE') sqt " +
//" ON apf.app_workflow_id = sqt.app_workflow_id;"
