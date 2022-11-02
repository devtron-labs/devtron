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

type ExternalLinkIdentifierMapping struct {
	tableName      struct{} `sql:"external_link_identifier_mapping"`
	Id             int      `sql:"id,pk"`
	ExternalLinkId int      `sql:"external_link_id,notnull"`
	Type           string   `sql:"type,notnull"`
	Identifier     string   `sql:"identifier,notnull"`
	ClusterId      int      `sql:"cluster_id,notnull"`
	Active         bool     `sql:"active, notnull"`
	//ExternalLink   ExternalLink
	sql.AuditLog
}

type ExternalLinkIdentifierMappingRepository interface {
	Save(externalLinksClusters *ExternalLinkIdentifierMapping, tx *pg.Tx) error
	//FindAllActiveByClusterId(clusterId int) ([]ExternalLinkIdentifierMapping, error)
	FindAllActiveByLinkIdentifier(identifier LinkIdentifier) ([]ExternalLinkIdentifierMapping, error)
	FindAllActive() ([]ExternalLinkIdentifierMapping, error)
	Update(link *ExternalLinkIdentifierMapping, tx *pg.Tx) error
	FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinkIdentifierMapping, error)
	FindAllByExternalLinkId(linkId int) ([]*ExternalLinkIdentifierMapping, error)
}
type ExternalLinkIdentifierMappingRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewExternalLinkIdentifierMappingRepositoryImpl(dbConnection *pg.DB) *ExternalLinkIdentifierMappingRepositoryImpl {
	return &ExternalLinkIdentifierMappingRepositoryImpl{dbConnection: dbConnection}
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) Save(externalLinksClusters *ExternalLinkIdentifierMapping, tx *pg.Tx) error {
	err := tx.Insert(externalLinksClusters)
	return err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) Update(link *ExternalLinkIdentifierMapping, tx *pg.Tx) error {
	err := tx.Update(link)
	return err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) FindAllActiveByLinkIdentifier(linkIdentifier LinkIdentifier) ([]ExternalLinkIdentifierMapping, error) {
	var links []ExternalLinkIdentifierMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_identifier_mapping.*").
		Where("external_link_identifier_mapping.active = ?", true).
		Where("external_link_identifier_mapping.cluster_id = ?", linkIdentifier.ClusterId).
		Where("external_link_identifier_mapping.type = ?", linkIdentifier.Type).
		Where("external_link_identifier_mapping.identifier = ?", linkIdentifier.Identifier).
		Select()
	return links, err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) FindAllActive() ([]ExternalLinkIdentifierMapping, error) {
	var links []ExternalLinkIdentifierMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_identifier_mapping.*").
		Where("external_link_identifier_mapping.active = ?", true).
		Select()
	return links, err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinkIdentifierMapping, error) {
	var links []*ExternalLinkIdentifierMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_cluster_mapping.*").
		Where("external_link_cluster_mapping.active = ?", true).
		Where("external_link_cluster_mapping.external_link_id = ?", linkId).
		Select()
	return links, err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) FindAllByExternalLinkId(linkId int) ([]*ExternalLinkIdentifierMapping, error) {
	var links []*ExternalLinkIdentifierMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_cluster_mapping.*").
		Where("external_link_cluster_mapping.external_link_id = ?", linkId).
		Select()
	return links, err
}
