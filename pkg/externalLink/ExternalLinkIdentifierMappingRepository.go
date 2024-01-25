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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"time"
)

type ExternalLinkIdentifierMapping struct {
	tableName      struct{}      `sql:"external_link_identifier_mapping" pg:",discard_unknown_columns"`
	Id             int           `sql:"id,pk"`
	ExternalLinkId int           `sql:"external_link_id,notnull"`
	Type           AppIdentifier `sql:"type,notnull"`
	Identifier     string        `sql:"identifier,notnull"`
	EnvId          int           `sql:"env_id,notnull"`
	AppId          int           `sql:"app_id,notnull"`
	ClusterId      int           `sql:"cluster_id,notnull"`
	Active         bool          `sql:"active, notnull"`
	//ExternalLink   ExternalLink
	sql.AuditLog
}

type ExternalLinkIdentifierMappingData struct {
	Id                           int           `sql:"id"`
	ExternalLinkMonitoringToolId int           `sql:"external_link_monitoring_tool_id, notnull"`
	Name                         string        `sql:"name,notnull"`
	Url                          string        `sql:"url,notnull"`
	IsEditable                   bool          `sql:"is_editable,notnull"`
	Description                  string        `sql:"description"`
	MappingId                    int           `sql:"mapping_id"`
	Active                       bool          `sql:"active"`
	Type                         AppIdentifier `sql:"type,notnull"`
	Identifier                   string        `sql:"identifier,notnull"`
	EnvId                        int           `sql:"env_id,notnull"`
	AppId                        int           `sql:"app_id,notnull"`
	ClusterId                    int           `sql:"cluster_id,notnull"`
	UpdatedOn                    time.Time     `sql:"updated_on"`
}

type ExternalLinkIdentifierMappingRepository interface {
	Save(externalLinksClusters *ExternalLinkIdentifierMapping, tx *pg.Tx) error

	FindAllActiveByLinkIdentifier(identifier *LinkIdentifier, clusterId int) ([]ExternalLinkIdentifierMappingData, error)

	Update(link *ExternalLinkIdentifierMapping, tx *pg.Tx) error
	UpdateAllActiveToInActive(Id int, tx *pg.Tx) error
	FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinkIdentifierMapping, error)

	FindAllActiveLinkIdentifierData() ([]ExternalLinkIdentifierMappingData, error)
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

func (impl ExternalLinkIdentifierMappingRepositoryImpl) UpdateAllActiveToInActive(Id int, tx *pg.Tx) error {
	model := ExternalLinkIdentifierMapping{}
	_, err := tx.Model(&model).Set("active = false").
		Where("external_link_id = ?", Id).
		Where("active = true").
		Update()
	return err
}
func (impl ExternalLinkIdentifierMappingRepositoryImpl) Update(link *ExternalLinkIdentifierMapping, tx *pg.Tx) error {
	err := tx.Update(link)
	return err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) FindAllActiveByLinkIdentifier(linkIdentifier *LinkIdentifier, clusterId int) ([]ExternalLinkIdentifierMappingData, error) {
	var links []ExternalLinkIdentifierMappingData
	var query string

	if linkIdentifier.Type == getType(DEVTRON_APP) || linkIdentifier.Type == getType(DEVTRON_INSTALLED_APP) {
		query = fmt.Sprintf("select el.id,el.external_link_monitoring_tool_id,el.name,el.url,el.is_editable,el.description,el.updated_on,"+
			"elim.id as mapping_id,elim.active,elim.type,elim.identifier,elim.env_id,elim.app_id,elim.cluster_id"+
			" FROM external_link el"+
			" LEFT JOIN external_link_identifier_mapping elim ON el.id = elim.external_link_id"+
			" WHERE el.active = true and elim.active = true and ( (elim.type = %d and elim.app_id = %d and elim.cluster_id = 0) or (elim.type = 0 and elim.app_id = 0 and elim.cluster_id = %d) "+
			" or (elim.type = -1) );", TypeMappings[linkIdentifier.Type], linkIdentifier.AppId, clusterId)
	} else {
		query = fmt.Sprintf("select el.id,el.external_link_monitoring_tool_id,el.name,el.url,el.is_editable,el.description,el.updated_on,"+
			"elim.id as mapping_id,elim.active,elim.type,elim.identifier,elim.env_id,elim.app_id,elim.cluster_id"+
			" FROM external_link el"+
			" LEFT JOIN external_link_identifier_mapping elim ON el.id = elim.external_link_id"+
			" WHERE el.active = true and elim.active = true and ( (elim.type = %d and elim.identifier = '%s' and elim.cluster_id = 0) or (elim.type = 0 and elim.app_id = 0 and elim.cluster_id = %d) "+
			" or (elim.type = -1) );", TypeMappings[linkIdentifier.Type], linkIdentifier.Identifier, clusterId)
	}
	_, err := impl.dbConnection.Query(&links, query)
	return links, err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) FindAllActiveLinkIdentifierData() ([]ExternalLinkIdentifierMappingData, error) {
	var links []ExternalLinkIdentifierMappingData
	query := "select el.id,el.external_link_monitoring_tool_id,el.name,el.url,el.is_editable,el.description,el.updated_on," +
		"elim.id as mapping_id,elim.active,elim.type,elim.identifier,elim.env_id,elim.app_id,elim.cluster_id" +
		" FROM external_link el" +
		" LEFT JOIN external_link_identifier_mapping elim ON el.id = elim.external_link_id Where el.active=true and elim.active = true;"
	_, err := impl.dbConnection.Query(&links, query)
	return links, err
}

func (impl ExternalLinkIdentifierMappingRepositoryImpl) FindAllActiveByExternalLinkId(linkId int) ([]*ExternalLinkIdentifierMapping, error) {
	var links []*ExternalLinkIdentifierMapping
	err := impl.dbConnection.Model(&links).
		Column("external_link_identifier_mapping.*").
		Where("external_link_identifier_mapping.active = ?", true).
		Where("external_link_identifier_mapping.external_link_id = ?", linkId).
		Select()
	return links, err
}
