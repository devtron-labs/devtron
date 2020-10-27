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

package cluster

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
)

type ClusterHelmConfig struct {
	tableName  struct{} `sql:"cluster_helm_config" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	ClusterId  int      `sql:"cluster_id"`
	Cluster    Cluster
	TillerUrl  string `sql:"tiller_url"`
	TillerCert string `sql:"tiller_cert"`
	TillerKey  string `sql:"tiller_key"`
	Active     bool   `sql:"active"`
	models.AuditLog
}

type ClusterHelmConfigRepository interface {
	Save(clusterHelmConfig *ClusterHelmConfig) error
	FindOneByEnvironment(environment string) (*ClusterHelmConfig, error)
}

type ClusterHelmConfigRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewClusterHelmConfigRepositoryImpl(dbConnection *pg.DB) *ClusterHelmConfigRepositoryImpl {
	return &ClusterHelmConfigRepositoryImpl{
		dbConnection: dbConnection,
	}
}

func (impl ClusterHelmConfigRepositoryImpl) Save(clusterHelmConfig *ClusterHelmConfig) error {
	return impl.dbConnection.Insert(clusterHelmConfig)
}

func (impl ClusterHelmConfigRepositoryImpl) FindOneByEnvironment(environment string) (*ClusterHelmConfig, error) {
	clusterHelmConfig := &ClusterHelmConfig{}
	err := impl.dbConnection.
		Model(clusterHelmConfig).
		Column("cluster_helm_config.*").
		Join("inner join environment ecm on cluster_helm_config.cluster_id = ecm.cluster_id").
		Where("ecm.environment_name = ?", environment).
		Where("ecm.active = ?", true).
		Where("cluster_helm_config.active = ?", true).
		Limit(1).
		Select()
	return clusterHelmConfig, err
}
