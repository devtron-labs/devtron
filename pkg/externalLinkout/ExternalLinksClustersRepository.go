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

package externalLinkout

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ExternalLinksClusters struct {
	tableName      struct{} `sql:"external_links_clusters"`
	Id             int      `sql:"id,pk"`
	ExternalLinkId int      `sql:"external_link_id,notnull"`
	ClusterId      int      `sql:"cluster_id,notnull"`
	IsActive       bool     `sql:"is_active,default true"`
	ExternalLinks  ExternalLinks
	sql.AuditLog
}

type ExternalLinksClustersRepository interface {
	Save(externalLinksClusters *ExternalLinksClusters) error
	FindAllActive(clusterId int) ([]ExternalLinksClusters, error)
	FindAll() ([]ExternalLinksClusters, error)
	Update(link *ExternalLinksClusters) error
	FindAllClusters(linkId int) ([]int, error)
}
type ExternalLinksClustersRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewExternalLinksClustersRepositoryImpl(dbConnection *pg.DB) *ExternalLinksClustersRepositoryImpl {
	return &ExternalLinksClustersRepositoryImpl{dbConnection: dbConnection}
}

func (impl ExternalLinksClustersRepositoryImpl) Save(externalLinksClusters *ExternalLinksClusters) error {
	err := impl.dbConnection.Insert(externalLinksClusters)
	return err
}

func (impl ExternalLinksClustersRepositoryImpl) Update(link *ExternalLinksClusters) error {
	err := impl.dbConnection.Update(link)
	return err
}

func (impl ExternalLinksClustersRepositoryImpl) FindAllActive(clusterId int) ([]ExternalLinksClusters, error) {
	var links []ExternalLinksClusters
	err := impl.dbConnection.Model(&links).
		Column("external_links_clusters.*", "ExternalLinks").
		Where("external_links_clusters.is_active = ?", true).
		Where("external_links_clusters.cluster_id = ?", clusterId).
		Select()
	return links, err
}

func (impl ExternalLinksClustersRepositoryImpl) FindAll() ([]ExternalLinksClusters, error) {
	var links []ExternalLinksClusters
	err := impl.dbConnection.Model(&links).
		Column("external_links_clusters.*", "ExternalLinks").
		Where("external_links_clusters.is_active = ?", true).
		Select()

	return links, err
}

func (impl ExternalLinksClustersRepositoryImpl) FindAllClusters(linkId int) ([]int, error) {
	var links []int
	err := impl.dbConnection.Model(&links).
		Where("is_active = ?", true).
		Where("external_link_id = ?", linkId).
		Select("cluster_id")

	return links, err
}
