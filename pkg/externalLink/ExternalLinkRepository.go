/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package externalLink

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ExternalLink struct {
	tableName                    struct{} `sql:"external_link" pg:",discard_unknown_columns"`
	Id                           int      `sql:"id,pk"`
	ExternalLinkMonitoringToolId int      `sql:"external_link_monitoring_tool_id, notnull"`
	Name                         string   `sql:"name,notnull"`
	Url                          string   `sql:"url,notnull"`
	IsEditable                   bool     `sql:"is_editable,notnull"`
	Description                  string   `sql:"description"`
	Active                       bool     `sql:"active,notnull"`
	sql.AuditLog
}

type ExternalLinkRepository interface {
	Save(externalLinks *ExternalLink, tx *pg.Tx) error

	FindOne(id int) (ExternalLink, error)
	Update(link *ExternalLink, tx *pg.Tx) error

	GetConnection() *pg.DB
	FindAllClusterLinks() ([]ExternalLink, error)
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

func (impl ExternalLinkRepositoryImpl) FindAllClusterLinks() ([]ExternalLink, error) {
	var res []ExternalLink
	query := " select * " +
		"from external_link el" +
		"  where el.id not in" +
		" (select distinct elim.external_link_id from external_link_identifier_mapping elim where elim.active = true)" +
		" and el.active = true;"
	_, err := impl.dbConnection.Query(&res, query)
	return res, err
}
