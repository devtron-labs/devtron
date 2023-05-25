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
