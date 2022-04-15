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
	tableName       struct{} `sql:"external_links_clusters"`
	Id              int      `sql:"id,pk"`
	ExternalLinksId int      `sql:"external_links_id,notnull"`
	ClusterId       int      `sql:"cluster_id,notnull"`
	Active          bool     `sql:"active, notnull"`
	ExternalLinks   ExternalLinks
	sql.AuditLog
}

type ExternalLinksClustersRepository interface {
	Save(externalLinksClusters *ExternalLinksClusters) error
	FindAllActiveByClusterId(clusterId int) ([]ExternalLinksClusters, error)
	FindAllActive() ([]ExternalLinksClusters, error)
	Update(link *ExternalLinksClusters) error
	FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinksClusters, error)
	FindAllActiveByLinkIdAndNotMatchedByClusterId(linkId int, clusterId int) ([]ExternalLinksClusters, error)
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

func (impl ExternalLinksClustersRepositoryImpl) FindAllActiveByClusterId(clusterId int) ([]ExternalLinksClusters, error) {
	var links []ExternalLinksClusters
	err := impl.dbConnection.Model(&links).
		Column("external_links_clusters.*", "ExternalLinks").
		Where("external_links_clusters.active = ?", true).
		Where("external_links_clusters.cluster_id = ?", clusterId).
		Select()
	return links, err
}
func (impl ExternalLinksClustersRepositoryImpl) FindAllActiveByLinkIdAndNotMatchedByClusterId(linkId int, clusterId int) ([]ExternalLinksClusters, error) {
	var links []ExternalLinksClusters
	err := impl.dbConnection.Model(&links).
		Column("external_links_clusters.*", "ExternalLinks").
		Where("external_links_clusters.active = ?", true).
		Where("external_links_clusters.external_links_id = ?", linkId).
		Where("external_links_clusters.cluster_id = ?", clusterId).
		Select()
	return links, err
}

func (impl ExternalLinksClustersRepositoryImpl) FindAllActive() ([]ExternalLinksClusters, error) {
	var links []ExternalLinksClusters
	err := impl.dbConnection.Model(&links).
		Column("external_links_clusters.*", "ExternalLinks").
		Where("external_links_clusters.active = ?", true).
		Select()

	return links, err
}

func (impl ExternalLinksClustersRepositoryImpl) FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinksClusters, error) {
	var links []*ExternalLinksClusters
	err := impl.dbConnection.Model(&links).
		Column("external_links_clusters.*", "ExternalLinks").
		Where("external_links_clusters.active = ?", true).
		Where("external_links_clusters.external_links_id = ?", linkId).
		Select()
	return links, err
}
