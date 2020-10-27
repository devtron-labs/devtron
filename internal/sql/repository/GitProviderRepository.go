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
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
)

type AuthMode string

const (
	AUTH_MODE_USERNAME_PASSWORD AuthMode = "USERNAME_PASSWORD"
	AUTH_MODE_SSH               AuthMode = "SSH"
	AUTH_MODE_ACCESS_TOKEN      AuthMode = "ACCESS_TOKEN"
	AUTH_MODE_ANONYMOUS         AuthMode = "ANONYMOUS"
)

type GitProvider struct {
	tableName   struct{} `sql:"git_provider" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Name        string   `sql:"name,notnull"`
	Url         string   `sql:"url,notnull"`
	UserName    string   `sql:"user_name"`
	Password    string   `sql:"password"`
	SshKey      string   `sql:"ssh_key"`
	AccessToken string   `sql:"access_token"`
	AuthMode    AuthMode `sql:"auth_mode,notnull"`
	Active      bool     `sql:"active,notnull"`
	models.AuditLog
}

type GitProviderRepository interface {
	Save(gitProvider *GitProvider) error
	ProviderExists(url string) (bool, error)
	FindAllActiveForAutocomplete() ([]GitProvider, error)
	FindAll() ([]GitProvider, error)
	FindOne(providerId string) (GitProvider, error)
	Update(gitProvider *GitProvider) error
}
type GitProviderRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewGitProviderRepositoryImpl(dbConnection *pg.DB) *GitProviderRepositoryImpl {
	return &GitProviderRepositoryImpl{dbConnection: dbConnection}
}

func (impl GitProviderRepositoryImpl) Save(gitProvider *GitProvider) error {
	err := impl.dbConnection.Insert(gitProvider)
	return err
}

func (impl GitProviderRepositoryImpl) ProviderExists(url string) (bool, error) {
	provider := &GitProvider{}
	exists, err := impl.dbConnection.
		Model(provider).
		Where("url = ?", url).
		Exists()
	return exists, err
}

func (impl GitProviderRepositoryImpl) FindAllActiveForAutocomplete() ([]GitProvider, error) {
	var providers []GitProvider
	err := impl.dbConnection.Model(&providers).
		Where("active = ?", true).Column("id", "name", "url").Select()
	return providers, err
}

func (impl GitProviderRepositoryImpl) FindAll() ([]GitProvider, error) {
	var providers []GitProvider
	err := impl.dbConnection.Model(&providers).Select()
	return providers, err
}

func (impl GitProviderRepositoryImpl) FindOne(providerId string) (GitProvider, error) {
	var provider GitProvider
	err := impl.dbConnection.Model(&provider).
		Where("id = ?", providerId).Select()
	return provider, err
}

func (impl GitProviderRepositoryImpl) Update(gitProvider *GitProvider) error {
	err := impl.dbConnection.Update(gitProvider)
	return err
}
