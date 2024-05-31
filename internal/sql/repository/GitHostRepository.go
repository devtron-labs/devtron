/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type GitHost struct {
	tableName       struct{} `sql:"git_host" pg:",discard_unknown_columns"`
	Id              int      `sql:"id,pk"`
	Name            string   `sql:"name,notnull"`
	Active          bool     `sql:"active,notnull"`
	WebhookUrl      string   `sql:"webhook_url"`
	WebhookSecret   string   `sql:"webhook_secret"`
	EventTypeHeader string   `sql:"event_type_header"`
	SecretHeader    string   `sql:"secret_header"`
	SecretValidator string   `sql:"secret_validator"`
	sql.AuditLog
}

type GitHostRepository interface {
	FindAll() ([]GitHost, error)
	FindOneById(Id int) (GitHost, error)
	FindOneByName(name string) (GitHost, error)
	Exists(name string) (bool, error)
	Save(gitHost *GitHost) error
}

type GitHostRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewGitHostRepositoryImpl(dbConnection *pg.DB) *GitHostRepositoryImpl {
	return &GitHostRepositoryImpl{dbConnection: dbConnection}
}

func (impl GitHostRepositoryImpl) FindAll() ([]GitHost, error) {
	var hosts []GitHost
	err := impl.dbConnection.Model(&hosts).Select()
	return hosts, err
}

func (impl GitHostRepositoryImpl) FindOneById(id int) (GitHost, error) {
	var host GitHost
	err := impl.dbConnection.Model(&host).
		Where("id = ?", id).Select()
	return host, err
}

func (impl GitHostRepositoryImpl) FindOneByName(name string) (GitHost, error) {
	var host GitHost
	err := impl.dbConnection.Model(&host).
		Where("name = ?", name).Select()
	return host, err
}

func (impl GitHostRepositoryImpl) Exists(name string) (bool, error) {
	gitHost := &GitHost{}
	exists, err := impl.dbConnection.
		Model(gitHost).
		Where("name = ?", name).
		Exists()
	return exists, err
}

func (impl GitHostRepositoryImpl) Save(gitHost *GitHost) error {
	err := impl.dbConnection.Insert(gitHost)
	return err
}
