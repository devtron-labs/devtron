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

type ExternalLinksMonitoringTools struct {
	tableName struct{} `sql:"external_links_monitoring_Tools"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
	Icon      string   `sql:"icon,notnull"`
	Active    bool     `sql:"active,notnull"`
	sql.AuditLog
}
type ExternalLinkoutToolsRepository interface {
	Save(externalLinkoutTools *ExternalLinksMonitoringTools) error
	FindAllActive() ([]ExternalLinksMonitoringTools, error)
}
type ExternalLinkoutToolsRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewExternalLinkoutToolsRepositoryImpl(dbConnection *pg.DB) *ExternalLinkoutToolsRepositoryImpl {
	return &ExternalLinkoutToolsRepositoryImpl{dbConnection: dbConnection}
}
func (impl ExternalLinkoutToolsRepositoryImpl) FindAllActive() ([]ExternalLinksMonitoringTools, error) {
	var tools []ExternalLinksMonitoringTools
	err := impl.dbConnection.Model(&tools).Where("active = ?", true).Select()
	return tools, err
}
func (impl ExternalLinkoutToolsRepositoryImpl) Save(externalLinkoutTools *ExternalLinksMonitoringTools) error {
	err := impl.dbConnection.Insert(externalLinkoutTools)
	return err
}
