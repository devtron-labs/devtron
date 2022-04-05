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
	"go.uber.org/zap"
)

type Cluster struct {
	tableName              struct{}          `sql:"cluster" pg:",discard_unknown_columns"`
	Id                     int               `sql:"id,pk"`
	ClusterName            string            `sql:"cluster_name"`
	ServerUrl              string            `sql:"server_url"`
	PrometheusEndpoint     string            `sql:"prometheus_endpoint"`
	Active                 bool              `sql:"active,notnull"`
	CdArgoSetup            bool              `sql:"cd_argo_setup,notnull"`
	Config                 map[string]string `sql:"config"`
	PUserName              string            `sql:"p_username"`
	PPassword              string            `sql:"p_password"`
	PTlsClientCert         string            `sql:"p_tls_client_cert"`
	PTlsClientKey          string            `sql:"p_tls_client_key"`
	AgentInstallationStage int               `sql:"agent_installation_stage"`
	K8sVersion             string            `sql:"k8s_version"`
	sql.AuditLog
}

type ClusterRepository interface {
	Save(model *Cluster) error
	FindOne(clusterName string) (*Cluster, error)
	FindOneActive(clusterName string) (*Cluster, error)
	FindAll() ([]Cluster, error)
	FindAllActive() ([]Cluster, error)

	FindById(id int) (*Cluster, error)
	FindByIds(id []int) ([]Cluster, error)
	Update(model *Cluster) error
	Delete(model *Cluster) error
	MarkClusterDeleted(model *Cluster) error
}

func NewClusterRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ClusterRepositoryImpl {
	return &ClusterRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type ClusterRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (impl ClusterRepositoryImpl) Save(model *Cluster) error {
	return impl.dbConnection.Insert(model)
}

func (impl ClusterRepositoryImpl) FindOne(clusterName string) (*Cluster, error) {
	cluster := &Cluster{}
	err := impl.dbConnection.
		Model(cluster).
		Where("cluster_name =?", clusterName).
		Where("active =?", true).
		Limit(1).
		Select()
	return cluster, err
}

func (impl ClusterRepositoryImpl) FindOneActive(clusterName string) (*Cluster, error) {
	cluster := &Cluster{}
	err := impl.dbConnection.
		Model(cluster).
		Where("cluster_name =?", clusterName).
		Where("active=?", true).
		Limit(1).
		Select()
	return cluster, err
}

func (impl ClusterRepositoryImpl) FindAll() ([]Cluster, error) {
	var clusters []Cluster
	err := impl.dbConnection.
		Model(&clusters).
		Where("active =?", true).
		Select()
	return clusters, err
}

func (impl ClusterRepositoryImpl) FindAllActive() ([]Cluster, error) {
	var clusters []Cluster
	err := impl.dbConnection.
		Model(&clusters).
		Where("active=?", true).
		Select()
	return clusters, err
}

func (impl ClusterRepositoryImpl) FindById(id int) (*Cluster, error) {
	cluster := &Cluster{}
	err := impl.dbConnection.
		Model(cluster).
		Where("id =?", id).
		Where("active =?", true).
		Limit(1).
		Select()
	return cluster, err
}

func (impl ClusterRepositoryImpl) FindByIds(id []int) ([]Cluster, error) {
	var cluster []Cluster
	err := impl.dbConnection.
		Model(&cluster).
		Where("id in(?)", pg.In(id)).
		Where("active =?", true).
		Select()
	return cluster, err
}

func (impl ClusterRepositoryImpl) Update(model *Cluster) error {
	return impl.dbConnection.Update(model)
}

func (impl ClusterRepositoryImpl) Delete(model *Cluster) error {
	return impl.dbConnection.Delete(model)
}

func (impl ClusterRepositoryImpl) MarkClusterDeleted(model *Cluster) error {
	model.Active = false
	return impl.dbConnection.Update(model)
}
