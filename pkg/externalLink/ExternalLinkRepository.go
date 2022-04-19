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

type ExternalLink struct {
	tableName                    struct{} `sql:"external_link"`
	Id                           int      `sql:"id,pk"`
	ExternalLinkMonitoringToolId int      `sql:"external_link_monitoring_tool_id, notnull"`
	Name                         string   `sql:"name,notnull"`
	Url                          string   `sql:"url,notnull"`
	Active                       bool     `sql:"active,notnull"`
	sql.AuditLog
}

type ExternalLinkRepository interface {
	Save(externalLinks *ExternalLink, tx *pg.Tx) error
	FindAllActive() ([]ExternalLink, error)
	FindOne(id int) (ExternalLink, error)
	Update(link *ExternalLink, tx *pg.Tx) error
	FindAllFilterOutByIds(ids []int) ([]ExternalLink, error)
	GetConnection() *pg.DB
}
type ExternalLinkRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewExternalLinkRepositoryImpl(dbConnection *pg.DB) *ExternalLinkRepositoryImpl {
	return &ExternalLinkRepositoryImpl{dbConnection: dbConnection}
}
func (repo ExternalLinkRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}
func (impl ExternalLinkRepositoryImpl) Save(externalLinks *ExternalLink, tx *pg.Tx) error {
	err := tx.Insert(externalLinks)
	return err
}
func (impl ExternalLinkRepositoryImpl) FindAllActive() ([]ExternalLink, error) {
	var links []ExternalLink
	err := impl.dbConnection.Model(&links).Where("active = ?", true).Select()
	return links, err
}
func (impl ExternalLinkRepositoryImpl) Update(link *ExternalLink, tx *pg.Tx) error {
	err := tx.Update(link)
	return err
}
func (impl ExternalLinkRepositoryImpl) FindOne(id int) (ExternalLink, error) {
	var link ExternalLink
	err := impl.dbConnection.Model(&link).
		Where("id = ?", id).
		Where("active = ?", true).Select()
	return link, err
}
func (impl ExternalLinkRepositoryImpl) FindAllFilterOutByIds(ids []int) ([]ExternalLink, error) {
	if ids == nil || len(ids) == 0 {
		return impl.FindAllActive()
	}
	var links []ExternalLink
	err := impl.dbConnection.Model(&links).
		Where("active = ?", true).
		Where("id not in (?)", pg.In(ids)).
		Select()
	return links, err
}
