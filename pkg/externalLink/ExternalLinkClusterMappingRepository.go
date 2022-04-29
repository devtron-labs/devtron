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

package externalLink

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ExternalLinkClusterMapping struct {
	tableName      struct{} `sql:"external_link_cluster_mapping"`
	Id             int      `sql:"id,pk"`
	ExternalLinkId int      `sql:"external_link_id,notnull"`
	ClusterId      int      `sql:"cluster_id,notnull"`
	Active         bool     `sql:"active, notnull"`
	ExternalLink   ExternalLink
	sql.AuditLog
}

type ExternalLinkClusterMappingRepository interface {
	Save(externalLinksClusters *ExternalLinkClusterMapping, tx *pg.Tx) error
	FindAllActiveByClusterId(clusterId int) ([]ExternalLinkClusterMapping, error)
	FindAllActive() ([]ExternalLinkClusterMapping, error)
	Update(link *ExternalLinkClusterMapping, tx *pg.Tx) error
	FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinkClusterMapping, error)
	FindAllByExternalLinkId(linkId int) ([]*ExternalLinkClusterMapping, error)
}
type ExternalLinkClusterMappingRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewExternalLinkClusterMappingRepositoryImpl(dbConnection *pg.DB) *ExternalLinkClusterMappingRepositoryImpl {
	return &ExternalLinkClusterMappingRepositoryImpl{dbConnection: dbConnection}
}

func (impl ExternalLinkClusterMappingRepositoryImpl) Save(externalLinksClusters *ExternalLinkClusterMapping, tx *pg.Tx) error {
	err := tx.Insert(externalLinksClusters)
	return err
}

func (impl ExternalLinkClusterMappingRepositoryImpl) Update(link *ExternalLinkClusterMapping, tx *pg.Tx) error {
	err := tx.Update(link)
	return err
}

func (impl ExternalLinkClusterMappingRepositoryImpl) FindAllActiveByClusterId(clusterId int) ([]ExternalLinkClusterMapping, error) {
	var links []ExternalLinkClusterMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_cluster_mapping.*", "ExternalLink").
		Where("external_link_cluster_mapping.active = ?", true).
		Where("external_link_cluster_mapping.cluster_id = ?", clusterId).
		Select()
	return links, err
}

func (impl ExternalLinkClusterMappingRepositoryImpl) FindAllActive() ([]ExternalLinkClusterMapping, error) {
	var links []ExternalLinkClusterMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_cluster_mapping.*", "ExternalLink").
		Where("external_link_cluster_mapping.active = ?", true).
		Select()
	return links, err
}

func (impl ExternalLinkClusterMappingRepositoryImpl) FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinkClusterMapping, error) {
	var links []*ExternalLinkClusterMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_cluster_mapping.*", "ExternalLink").
		Where("external_link_cluster_mapping.active = ?", true).
		Where("external_link_cluster_mapping.external_link_id = ?", linkId).
		Select()
	return links, err
}

func (impl ExternalLinkClusterMappingRepositoryImpl) FindAllByExternalLinkId(linkId int) ([]*ExternalLinkClusterMapping, error) {
	var links []*ExternalLinkClusterMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_cluster_mapping.*", "ExternalLink").
		Where("external_link_cluster_mapping.external_link_id = ?", linkId).
		Select()
	return links, err
}
