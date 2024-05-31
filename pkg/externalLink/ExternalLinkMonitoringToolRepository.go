/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package externalLink

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ExternalLinkMonitoringTool struct {
	tableName struct{} `sql:"external_link_monitoring_tool" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
	Icon      string   `sql:"icon,notnull"`
	Category  int      `sql:"category"`
	Active    bool     `sql:"active,notnull"`
	sql.AuditLog
}
type ExternalLinkMonitoringToolRepository interface {
	FindAllActive() ([]ExternalLinkMonitoringTool, error)
}
type ExternalLinkMonitoringToolRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewExternalLinkMonitoringToolRepositoryImpl(dbConnection *pg.DB) *ExternalLinkMonitoringToolRepositoryImpl {
	return &ExternalLinkMonitoringToolRepositoryImpl{dbConnection: dbConnection}
}
func (impl ExternalLinkMonitoringToolRepositoryImpl) FindAllActive() ([]ExternalLinkMonitoringTool, error) {
	var tools []ExternalLinkMonitoringTool
	err := impl.dbConnection.Model(&tools).Where("active = ?", true).Select()
	return tools, err
}
