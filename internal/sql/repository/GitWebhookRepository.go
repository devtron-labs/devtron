/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"github.com/go-pg/pg"
	"time"
)

type GitWebhook struct {
	tableName     struct{}  `sql:"git_web_hook" pg:",discard_unknown_columns"`
	Id            int       `sql:"id,pk"`
	CiMaterialId  int       `sql:"ci_material_id"`
	GitMaterialId int       `sql:"git_material_id"`
	Type          string    `sql:"type"`
	Value         string    `sql:"value"`
	Active        bool      `sql:"active"`
	LastSeenHash  string    `sql:"last_seen_hash"`
	CreatedOn     time.Time `sql:"created_on"`
}

type GitWebhookRepository interface {
	Save(gitWebhook *GitWebhook) error
}

type GitWebhookRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewGitWebhookRepositoryImpl(dbConnection *pg.DB) *GitWebhookRepositoryImpl {
	return &GitWebhookRepositoryImpl{dbConnection: dbConnection}
}

func (impl *GitWebhookRepositoryImpl) Save(gitWebhook *GitWebhook) error {
	return impl.dbConnection.Insert(gitWebhook)
}
