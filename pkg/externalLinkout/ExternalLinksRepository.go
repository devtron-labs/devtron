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

type ExternalLinks struct {
	tableName                     struct{} `sql:"external_links"`
	Id                            int      `sql:"id,pk"`
	ExternalLinksMonitoringToolId int      `sql:"external_links_monitoring_tool_id, notnull"`
	Name                          string   `sql:"name,notnull"`
	Url                           string   `sql:"url,notnull"`
	Active                        bool     `sql:"active,notnull"`
	sql.AuditLog
}

type ExternalLinksRepository interface {
	Save(externalLinks *ExternalLinks) error
	FindAllActive() ([]ExternalLinks, error)
	FindOne(id int) (ExternalLinks, error)
	Update(link *ExternalLinks) error
	FindAllNonMapped(mappedExternalLinksIds []int) ([]ExternalLinks, error)
}
type ExternalLinksRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewExternalLinksRepositoryImpl(dbConnection *pg.DB) *ExternalLinksRepositoryImpl {
	return &ExternalLinksRepositoryImpl{dbConnection: dbConnection}
}
func (impl ExternalLinksRepositoryImpl) Save(externalLinks *ExternalLinks) error {
	err := impl.dbConnection.Insert(externalLinks)
	return err
}
func (impl ExternalLinksRepositoryImpl) FindAllActive() ([]ExternalLinks, error) {
	var links []ExternalLinks
	err := impl.dbConnection.Model(&links).Where("active = ?", true).Select()
	return links, err
}
func (impl ExternalLinksRepositoryImpl) Update(link *ExternalLinks) error {
	err := impl.dbConnection.Update(link)
	return err
}
func (impl ExternalLinksRepositoryImpl) FindOne(id int) (ExternalLinks, error) {
	var link ExternalLinks
	err := impl.dbConnection.Model(&link).
		Where("id = ?", id).
		Where("active = ?", true).Select()
	return link, err
}
func (impl ExternalLinksRepositoryImpl) FindAllNonMapped(mappedExternalLinksIds []int) ([]ExternalLinks, error) {
	var links []ExternalLinks
	err := impl.dbConnection.Model(&links).
		Where("active = ?", true).
		Where("id not in (?)", pg.In(mappedExternalLinksIds)).
		Select()
	return links, err
}
