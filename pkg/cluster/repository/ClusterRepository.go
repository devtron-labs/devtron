/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type Cluster struct {
	tableName              struct{}          `sql:"cluster" pg:",discard_unknown_columns"`
	Id                     int               `sql:"id,pk"`
	ClusterName            string            `sql:"cluster_name"`
	Description            string            `sql:"description"`
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
	ErrorInConnecting      string            `sql:"error_in_connecting"`
	IsVirtualCluster       bool              `sql:"is_virtual_cluster"`
	InsecureSkipTlsVerify  bool              `sql:"insecure_skip_tls_verify"`
	sql.AuditLog
}

type ClusterRepository interface {
	Save(model *Cluster) error
	FindOne(clusterName string) (*Cluster, error)
	FindOneActive(clusterName string) (*Cluster, error)
	FindAll() ([]Cluster, error)
	FindAllActive() ([]Cluster, error)
	FindAllActiveExceptVirtual() ([]Cluster, error)
	FindById(id int) (*Cluster, error)
	FindByIds(id []int) ([]Cluster, error)
	Update(model *Cluster) error
	SetDescription(id int, description string, userId int32) error
	Delete(model *Cluster) error
	MarkClusterDeleted(model *Cluster) error
	UpdateClusterConnectionStatus(clusterId int, errorInConnecting string) error
	FindActiveClusters() ([]Cluster, error)
	SaveAll(models []*Cluster) error
	FindByNames(clusterNames []string) ([]*Cluster, error)
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
func (impl ClusterRepositoryImpl) SaveAll(models []*Cluster) error {
	return impl.dbConnection.Insert(models)
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

func (impl ClusterRepositoryImpl) FindActiveClusters() ([]Cluster, error) {
	activeClusters := make([]Cluster, 0)
	query := "select id, cluster_name, active from cluster where active = true"
	_, err := impl.dbConnection.Query(&activeClusters, query)
	return activeClusters, err
}

func (impl ClusterRepositoryImpl) FindAllActive() ([]Cluster, error) {
	var clusters []Cluster
	err := impl.dbConnection.
		Model(&clusters).
		Where("active=?", true).
		Select()
	return clusters, err
}

func (impl ClusterRepositoryImpl) FindAllActiveExceptVirtual() ([]Cluster, error) {
	var clusters []Cluster
	err := impl.dbConnection.
		Model(&clusters).
		Where("active=?", true).
		Where("is_virtual_cluster=? OR is_virtual_cluster IS NULL", false).
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

func (impl ClusterRepositoryImpl) FindByNames(clusterNames []string) ([]*Cluster, error) {
	var cluster []*Cluster
	err := impl.dbConnection.
		Model(&cluster).
		Where("cluster_name in (?)", pg.In(clusterNames)).
		Where("active = ?", true).
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

func (impl ClusterRepositoryImpl) SetDescription(id int, description string, userId int32) error {
	_, err := impl.dbConnection.Model((*Cluster)(nil)).
		Set("description = ?", description).Set("updated_by = ?", userId).Set("updated_on = ?", time.Now()).
		Where("id = ?", id).Update()
	return err
}

func (impl ClusterRepositoryImpl) Delete(model *Cluster) error {
	return impl.dbConnection.Delete(model)
}

func (impl ClusterRepositoryImpl) MarkClusterDeleted(model *Cluster) error {
	model.Active = false
	return impl.dbConnection.Update(model)
}

func (impl ClusterRepositoryImpl) UpdateClusterConnectionStatus(clusterId int, errorInConnecting string) error {
	cluster := &Cluster{}
	_, err := impl.dbConnection.Model(cluster).
		Set("error_in_connecting = ?", errorInConnecting).Where("id = ?", clusterId).
		Update()
	return err
}
